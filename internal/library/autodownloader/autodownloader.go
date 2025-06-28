package autodownloader

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/notifier"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/habari"
	"github.com/adrg/strutil/metrics"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/sourcegraph/conc/pool"
)

const (
	ComparisonThreshold = 0.8
)

type (
	AutoDownloader struct {
		logger                  *zerolog.Logger
		torrentClientRepository *torrent_client.Repository
		torrentRepository       *torrent.Repository
		debridClientRepository  *debrid_client.Repository
		database                *db.Database
		animeCollection         mo.Option[*anilist.AnimeCollection]
		wsEventManager          events.WSEventManagerInterface
		settings                *models.AutoDownloaderSettings
		metadataProvider        metadata.Provider
		settingsUpdatedCh       chan struct{}
		stopCh                  chan struct{}
		startCh                 chan struct{}
		debugTrace              bool
		mu                      sync.Mutex
		isOffline               *bool
	}

	NewAutoDownloaderOptions struct {
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		TorrentRepository       *torrent.Repository
		WSEventManager          events.WSEventManagerInterface
		Database                *db.Database
		MetadataProvider        metadata.Provider
		DebridClientRepository  *debrid_client.Repository
		IsOffline               *bool
	}

	tmpTorrentToDownload struct {
		torrent *NormalizedTorrent
		episode int
	}
)

func New(opts *NewAutoDownloaderOptions) *AutoDownloader {
	return &AutoDownloader{
		logger:                  opts.Logger,
		torrentClientRepository: opts.TorrentClientRepository,
		torrentRepository:       opts.TorrentRepository,
		database:                opts.Database,
		wsEventManager:          opts.WSEventManager,
		animeCollection:         mo.None[*anilist.AnimeCollection](),
		metadataProvider:        opts.MetadataProvider,
		debridClientRepository:  opts.DebridClientRepository,
		settings: &models.AutoDownloaderSettings{
			Provider:              torrent.ProviderAnimeTosho, // Default provider, will be updated after the settings are fetched
			Interval:              20,
			Enabled:               false,
			DownloadAutomatically: false,
			EnableEnhancedQueries: false,
		},
		settingsUpdatedCh: make(chan struct{}, 1),
		stopCh:            make(chan struct{}, 1),
		startCh:           make(chan struct{}, 1),
		debugTrace:        true,
		mu:                sync.Mutex{},
		isOffline:         opts.IsOffline,
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
// If the AutoDownloader is not active, it will start it if the settings are enabled.
// If the AutoDownloader is active, it will stop it if the settings are disabled.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings, provider string) {
	defer util.HandlePanicInModuleThen("autodownloader/SetSettings", func() {})

	event := &AutoDownloaderSettingsUpdatedEvent{
		Settings: settings,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderSettingsUpdated().Trigger(event)
	settings = event.Settings

	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		defer ad.mu.Unlock()
		ad.settings = settings
		// Update the provider if it's provided
		if provider != "" {
			ad.settings.Provider = provider
		}
		ad.settingsUpdatedCh <- struct{}{} // Notify that the settings have been updated
		if ad.settings.Enabled {
			ad.startCh <- struct{}{} // Start the auto downloader
		} else if !ad.settings.Enabled {
			ad.stopCh <- struct{}{} // Stop the auto downloader
		}
	}()
}

func (ad *AutoDownloader) SetAnimeCollection(ac *anilist.AnimeCollection) {
	ad.animeCollection = mo.Some(ac)
}

func (ad *AutoDownloader) SetTorrentClientRepository(repo *torrent_client.Repository) {
	defer util.HandlePanicInModuleThen("autodownloader/SetTorrentClientRepository", func() {})

	if ad == nil {
		return
	}
	ad.torrentClientRepository = repo
}

// Start will start the auto downloader in a goroutine
func (ad *AutoDownloader) Start() {
	defer util.HandlePanicInModuleThen("autodownloader/Start", func() {})

	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		if ad.settings.Enabled {
			started := ad.torrentClientRepository.Start() // Start torrent client if it's not running
			if !started {
				ad.logger.Warn().Msg("autodownloader: Failed to start torrent client. Make sure it's running for the Auto Downloader to work.")
				ad.mu.Unlock()
				return
			}
		}
		ad.mu.Unlock()

		// Start the auto downloader
		ad.start()
	}()
}

func (ad *AutoDownloader) Run() {
	defer util.HandlePanicInModuleThen("autodownloader/Run", func() {})

	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		defer ad.mu.Unlock()
		ad.startCh <- struct{}{}
		ad.logger.Trace().Msg("autodownloader: Received start signal")
	}()
}

// CleanUpDownloadedItems will clean up downloaded items from the database.
// This should be run after a scan is completed.
func (ad *AutoDownloader) CleanUpDownloadedItems() {
	defer util.HandlePanicInModuleThen("autodownloader/CleanUpDownloadedItems", func() {})

	if ad == nil {
		return
	}
	ad.mu.Lock()
	defer ad.mu.Unlock()
	err := ad.database.DeleteDownloadedAutoDownloaderItems()
	if err != nil {
		return
	}
}

func (ad *AutoDownloader) start() {
	defer util.HandlePanicInModuleThen("autodownloader/start", func() {})

	if ad.settings.Enabled {
		ad.logger.Info().Msg("autodownloader: Module started")
	}

	for {
		interval := 20
		// Use the user-defined interval if it's greater or equal to 15
		if ad.settings != nil && ad.settings.Interval > 0 && ad.settings.Interval >= 15 {
			interval = ad.settings.Interval
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		select {
		case <-ad.settingsUpdatedCh:
			break // Restart the loop
		case <-ad.stopCh:

		case <-ad.startCh:
			if ad.settings.Enabled {
				ad.logger.Debug().Msg("autodownloader: Auto Downloader started")
				ad.checkForNewEpisodes()
			}
		case <-ticker.C:
			if ad.settings.Enabled {
				ad.checkForNewEpisodes()
			}
		}
		ticker.Stop()
	}

}

func (ad *AutoDownloader) checkForNewEpisodes() {
	defer util.HandlePanicInModuleThen("autodownloader/checkForNewEpisodes", func() {})

	if ad.isOffline != nil && *ad.isOffline {
		ad.logger.Debug().Msg("autodownloader: Skipping check for new episodes. AutoDownloader is in offline mode.")
		return
	}

	ad.mu.Lock()
	if ad == nil || ad.torrentRepository == nil || !ad.settings.Enabled || ad.settings.Provider == "" || ad.settings.Provider == torrent.ProviderNone {
		ad.logger.Warn().Msg("autodownloader: Could not check for new episodes. AutoDownloader is not enabled or provider is not set.")
		ad.mu.Unlock()
		return
	}

	// DEVNOTE: [checkForNewEpisodes] is called on startup, when the default anime provider extension has not yet been loaded.
	providerExt, found := ad.torrentRepository.GetDefaultAnimeProviderExtension()
	if !found {
		//ad.logger.Warn().Msg("autodownloader: Could not check for new episodes. Default provider not found.")
		ad.mu.Unlock()
		return
	}
	if providerExt.GetProvider().GetSettings().Type != hibiketorrent.AnimeProviderTypeMain {
		ad.logger.Warn().Msgf("autodownloader: Could not check for new episodes. Provider '%s' cannot be used for auto downloading.", providerExt.GetName())
		ad.mu.Unlock()
		return
	}
	ad.mu.Unlock()

	torrents := make([]*NormalizedTorrent, 0)

	// Get rules from the database
	rules, err := db_bridge.GetAutoDownloaderRules(ad.database)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to fetch rules from the database")
		return
	}

	// Filter out disabled rules
	_filteredRules := make([]*anime.AutoDownloaderRule, 0)
	for _, rule := range rules {
		if rule.Enabled {
			_filteredRules = append(_filteredRules, rule)
		}
	}
	rules = _filteredRules

	// Event
	event := &AutoDownloaderRunStartedEvent{
		Rules: rules,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderRunStarted().Trigger(event)
	rules = event.Rules

	// Default prevented, return
	if event.DefaultPrevented {
		return
	}

	// If there are no rules, return
	if len(rules) == 0 {
		ad.logger.Debug().Msg("autodownloader: No rules found")
		return
	}

	// Get local files from the database
	lfs, _, err := db_bridge.GetLocalFiles(ad.database)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to fetch local files from the database")
		return
	}
	// Create a LocalFileWrapper
	lfWrapper := anime.NewLocalFileWrapper(lfs)

	// Get the latest torrents
	torrents, err = ad.getLatestTorrents(rules)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to get latest torrents")
		return
	}

	// Event
	fetchedEvent := &AutoDownloaderTorrentsFetchedEvent{
		Torrents: torrents,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderTorrentsFetched().Trigger(fetchedEvent)
	torrents = fetchedEvent.Torrents

	// // Try to start the torrent client if it's not running
	// if ad.torrentClientRepository != nil {
	// 	started := ad.torrentClientRepository.Start() // Start torrent client if it's not running
	// 	if !started {
	// 		ad.logger.Warn().Msg("autodownloader: Failed to start torrent client. Make sure it's running.")
	// 	}
	// }

	// Get existing torrents
	existingTorrents := make([]*torrent_client.Torrent, 0)
	if ad.torrentClientRepository != nil {
		existingTorrents, err = ad.torrentClientRepository.GetList()
		if err != nil {
			existingTorrents = make([]*torrent_client.Torrent, 0)
		}
	}

	downloaded := 0
	mu := sync.Mutex{}

	// Going through each rule
	p := pool.New()
	for _, rule := range rules {
		rule := rule
		p.Go(func() {
			if !rule.Enabled {
				return // Skip rule
			}
			listEntry, found := ad.getRuleListEntry(rule)
			// If the media is not found, skip the rule
			if !found {
				return // Skip rule
			}

			// DEVNOTE: This is bad, do not skip anime that are not releasing because dubs are delayed
			// If the media is not releasing AND has more than one episode, skip the rule
			// This is to avoid skipping movies and single-episode OVAs
			//if *listEntry.GetMedia().GetStatus() != anilist.MediaStatusReleasing && listEntry.GetMedia().GetCurrentEpisodeCount() > 1 {
			//	return // Skip rule
			//}

			localEntry, _ := lfWrapper.GetLocalEntryById(listEntry.GetMedia().GetID())

			// +---------------------+
			// |    Existing Item    |
			// +---------------------+
			items, err := ad.database.GetAutoDownloaderItemByMediaId(listEntry.GetMedia().GetID())
			if err != nil {
				items = make([]*models.AutoDownloaderItem, 0)
			}

			// Get all torrents that follow the rule
			torrentsToDownload := make([]*tmpTorrentToDownload, 0)
		outer:
			for _, t := range torrents {
				// If the torrent is already added, skip it
				for _, et := range existingTorrents {
					if et.Hash == t.InfoHash {
						continue outer // Skip the torrent
					}
				}

				episode, ok := ad.torrentFollowsRule(t, rule, listEntry, localEntry, items)
				event := &AutoDownloaderMatchVerifiedEvent{
					Torrent:    t,
					Rule:       rule,
					ListEntry:  listEntry,
					LocalEntry: localEntry,
					Episode:    episode,
					MatchFound: ok,
				}
				_ = hook.GlobalHookManager.OnAutoDownloaderMatchVerified().Trigger(event)
				t = event.Torrent
				rule = event.Rule
				listEntry = event.ListEntry
				localEntry = event.LocalEntry
				episode = event.Episode
				ok = event.MatchFound

				// Default prevented, skip the torrent
				if event.DefaultPrevented {
					continue outer // Skip the torrent
				}

				if ok {
					torrentsToDownload = append(torrentsToDownload, &tmpTorrentToDownload{
						torrent: t,
						episode: episode,
					})
				}
			}

			// Download the torrent if there's only one
			if len(torrentsToDownload) == 1 {
				t := torrentsToDownload[0]
				ok := ad.downloadTorrent(t.torrent, rule, t.episode)
				if ok {
					downloaded++
				}
				return
			}

			// If there's more than one, we will group them by episode and sort them
			// Make a map [episode]torrents
			epMap := make(map[int][]*tmpTorrentToDownload)
			for _, t := range torrentsToDownload {
				if _, ok := epMap[t.episode]; !ok {
					epMap[t.episode] = make([]*tmpTorrentToDownload, 0)
					epMap[t.episode] = append(epMap[t.episode], t)
				} else {
					epMap[t.episode] = append(epMap[t.episode], t)
				}
			}

			// Go through each episode group and download the best torrent (by resolution and seeders)
			for ep, torrents := range epMap {

				// If there's only one torrent for the episode, download it
				if len(torrents) == 1 {
					ok := ad.downloadTorrent(torrents[0].torrent, rule, ep)
					if ok {
						mu.Lock()
						downloaded++
						mu.Unlock()
					}
					continue
				}

				// If there are more than one
				// Sort by resolution
				sort.Slice(torrents, func(i, j int) bool {
					qI := comparison.ExtractResolutionInt(torrents[i].torrent.ParsedData.VideoResolution)
					qJ := comparison.ExtractResolutionInt(torrents[j].torrent.ParsedData.VideoResolution)
					return qI > qJ
				})
				// Sort by seeds
				sort.Slice(torrents, func(i, j int) bool {
					return torrents[i].torrent.Seeders > torrents[j].torrent.Seeders
				})

				ok := ad.downloadTorrent(torrents[0].torrent, rule, ep)
				if ok {
					mu.Lock()
					downloaded++
					mu.Unlock()
				}
			}
		})
	}
	p.Wait()

	if downloaded > 0 {
		if ad.settings.DownloadAutomatically {
			notifier.GlobalNotifier.Notify(
				notifier.AutoDownloader,
				fmt.Sprintf("%d %s %s been downloaded.", downloaded, util.Pluralize(downloaded, "episode", "episodes"), util.Pluralize(downloaded, "has", "have")),
			)
		} else {
			notifier.GlobalNotifier.Notify(
				notifier.AutoDownloader,
				fmt.Sprintf("%d %s %s been added to the queue.", downloaded, util.Pluralize(downloaded, "episode", "episodes"), util.Pluralize(downloaded, "has", "have")),
			)
		}
	}

}

func (ad *AutoDownloader) torrentFollowsRule(
	t *NormalizedTorrent,
	rule *anime.AutoDownloaderRule,
	listEntry *anilist.AnimeListEntry,
	localEntry *anime.LocalFileWrapperEntry,
	items []*models.AutoDownloaderItem,
) (int, bool) {
	defer util.HandlePanicInModuleThen("autodownloader/torrentFollowsRule", func() {})

	if ok := ad.isReleaseGroupMatch(t.ParsedData.ReleaseGroup, rule); !ok {
		return -1, false
	}

	if ok := ad.isResolutionMatch(t.ParsedData.VideoResolution, rule); !ok {
		return -1, false
	}

	if ok := ad.isTitleMatch(t.ParsedData, t.Name, rule, listEntry); !ok {
		return -1, false
	}

	if ok := ad.isAdditionalTermsMatch(t.Name, rule); !ok {
		return -1, false
	}

	episode, ok := ad.isSeasonAndEpisodeMatch(t.ParsedData, rule, listEntry, localEntry, items)
	if !ok {
		return -1, false
	}

	return episode, true
}

func (ad *AutoDownloader) downloadTorrent(t *NormalizedTorrent, rule *anime.AutoDownloaderRule, episode int) bool {
	defer util.HandlePanicInModuleThen("autodownloader/downloadTorrent", func() {})

	ad.mu.Lock()
	defer ad.mu.Unlock()

	// Double check that the episode hasn't been added while we have the lock
	items, err := ad.database.GetAutoDownloaderItemByMediaId(rule.MediaId)
	if err == nil {
		for _, item := range items {
			if item.Episode == episode {
				return false // Skip, episode was added by another goroutine
			}
		}
	}

	// Event
	beforeEvent := &AutoDownloaderBeforeDownloadTorrentEvent{
		Torrent: t,
		Rule:    rule,
		Items:   items,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderBeforeDownloadTorrent().Trigger(beforeEvent)
	t = beforeEvent.Torrent
	rule = beforeEvent.Rule
	_ = beforeEvent.Items

	// Default prevented, return
	if beforeEvent.DefaultPrevented {
		return false
	}

	providerExtension, found := ad.torrentRepository.GetDefaultAnimeProviderExtension()
	if !found {
		ad.logger.Warn().Msg("autodownloader: Could not download torrent. Default provider not found")
		return false
	}

	if ad.torrentClientRepository == nil {
		ad.logger.Error().Msg("autodownloader: torrent client not found")
		return false
	}

	useDebrid := false

	if ad.settings.UseDebrid {
		// Check if the debrid provider is enabled
		if !ad.debridClientRepository.HasProvider() || !ad.debridClientRepository.GetSettings().Enabled {
			ad.logger.Error().Msg("autodownloader: Debrid provider not found or not enabled")
			// We return instead of falling back to torrent client
			return false
		}
		useDebrid = true
	}

	// Get torrent magnet
	magnet, err := t.GetMagnet(providerExtension.GetProvider())
	if err != nil {
		ad.logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to get magnet link for torrent")
		return false
	}

	downloaded := false

	if useDebrid {
		//
		// Debrid
		//

		if ad.settings.DownloadAutomatically {
			// Add the torrent to the debrid provider and queue it
			_, err := ad.debridClientRepository.AddAndQueueTorrent(debrid.AddTorrentOptions{
				MagnetLink:   magnet,
				SelectFileId: "all", // RD-only, select all files
			}, rule.Destination, rule.MediaId)
			if err != nil {
				ad.logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to debrid")
				return false
			}
		} else {
			debridProvider, err := ad.debridClientRepository.GetProvider()
			if err != nil {
				ad.logger.Error().Err(err).Msg("autodownloader: Failed to get debrid provider")
				return false
			}

			// Add the torrent to the debrid provider
			_, err = debridProvider.AddTorrent(debrid.AddTorrentOptions{
				MagnetLink:   magnet,
				SelectFileId: "all", // RD-only, select all files
			})
			if err != nil {
				ad.logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to debrid")
				return false
			}
		}

	} else {
		// Pause the torrent when it's added
		if ad.settings.DownloadAutomatically {

			//
			// Torrent client
			//
			started := ad.torrentClientRepository.Start() // Start torrent client if it's not running
			if !started {
				ad.logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to download torrent. torrent client is not running.")
				return false
			}

			// Return if the torrent is already added
			torrentExists := ad.torrentClientRepository.TorrentExists(t.InfoHash)
			if torrentExists {
				//ad.Logger.Debug().Str("name", t.Name).Msg("autodownloader: Torrent already added")
				return false
			}

			ad.logger.Debug().Msgf("autodownloader: Downloading torrent: %s", t.Name)

			// Add the torrent to torrent client
			err := ad.torrentClientRepository.AddMagnets([]string{magnet}, rule.Destination)
			if err != nil {
				ad.logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to torrent client")
				return false
			}

			downloaded = true
		}
	}

	ad.logger.Info().Str("name", t.Name).Msg("autodownloader: Added torrent")
	ad.wsEventManager.SendEvent(events.AutoDownloaderItemAdded, t.Name)

	// Add the torrent to the database
	item := &models.AutoDownloaderItem{
		RuleID:      rule.DbID,
		MediaID:     rule.MediaId,
		Episode:     episode,
		Link:        t.Link,
		Hash:        t.InfoHash,
		Magnet:      magnet,
		TorrentName: t.Name,
		Downloaded:  downloaded,
	}
	_ = ad.database.InsertAutoDownloaderItem(item)

	// Event
	afterEvent := &AutoDownloaderAfterDownloadTorrentEvent{
		Torrent: t,
		Rule:    rule,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderAfterDownloadTorrent().Trigger(afterEvent)

	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ad *AutoDownloader) isAdditionalTermsMatch(torrentName string, rule *anime.AutoDownloaderRule) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isAdditionalTermsMatch", func() {
		ok = false
	})

	if len(rule.AdditionalTerms) == 0 {
		return true
	}

	// Go through each additional term
	for _, optionsText := range rule.AdditionalTerms {
		// Split the options by comma
		options := strings.Split(strings.TrimSpace(optionsText), ",")
		// Check if the torrent name contains at least one of the options
		foundOption := false
		for _, option := range options {
			option := strings.TrimSpace(option)
			if strings.Contains(strings.ToLower(torrentName), strings.ToLower(option)) {
				foundOption = true
			}
		}
		// If the torrent name doesn't contain any of the options, return false
		if !foundOption {
			return false
		}
	}

	// If all options are found, return true
	return true
}
func (ad *AutoDownloader) isReleaseGroupMatch(releaseGroup string, rule *anime.AutoDownloaderRule) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isReleaseGroupMatch", func() {
		ok = false
	})

	if len(rule.ReleaseGroups) == 0 {
		return true
	}
	for _, rg := range rule.ReleaseGroups {
		if strings.ToLower(rg) == strings.ToLower(releaseGroup) {
			return true
		}
	}
	return false
}

// isResolutionMatch
// DEVOTE: Improve this
func (ad *AutoDownloader) isResolutionMatch(quality string, rule *anime.AutoDownloaderRule) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isResolutionMatch", func() {
		ok = false
	})

	if len(rule.Resolutions) == 0 {
		return true
	}
	if quality == "" {
		return false
	}
	for _, q := range rule.Resolutions {
		qualityWithoutP := strings.TrimSuffix(quality, "p")
		qWithoutP := strings.TrimSuffix(q, "p")
		if quality == q || qualityWithoutP == qWithoutP {
			return true
		}
		if strings.Contains(quality, qWithoutP) { // e.g. 1080 in 1920x1080
			return true
		}
	}
	return false
}

func (ad *AutoDownloader) isTitleMatch(torrentParsedData *habari.Metadata, torrentName string, rule *anime.AutoDownloaderRule, listEntry *anilist.AnimeListEntry) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isTitleMatch", func() {
		ok = false
	})

	switch rule.TitleComparisonType {
	case anime.AutoDownloaderRuleTitleComparisonContains:
		// +---------------------+
		// |   Title "Contains"  |
		// +---------------------+

		// Check if the torrent name contains the comparison title exactly
		// This will fail for torrent titles that don't contain a season number if the comparison title has a season number
		if strings.Contains(strings.ToLower(torrentParsedData.Title), strings.ToLower(rule.ComparisonTitle)) {
			return true
		}
		if strings.Contains(strings.ToLower(torrentName), strings.ToLower(rule.ComparisonTitle)) {
			return true
		}

	case anime.AutoDownloaderRuleTitleComparisonLikely:
		// +---------------------+
		// |   Title "Likely"    |
		// +---------------------+

		torrentTitle := torrentParsedData.Title
		comparisonTitle := strings.ReplaceAll(strings.ReplaceAll(rule.ComparisonTitle, "[", ""), "]", "")

		// 1. Use comparison title (without season number - if it exists)
		// Remove season number from the torrent title if it exists
		parsedComparisonTitle := habari.Parse(comparisonTitle)
		if parsedComparisonTitle.Title != "" && len(parsedComparisonTitle.SeasonNumber) > 0 {
			_comparisonTitle := parsedComparisonTitle.Title
			if len(parsedComparisonTitle.ReleaseGroup) > 0 {
				_comparisonTitle = fmt.Sprintf("%s %s", parsedComparisonTitle.ReleaseGroup, _comparisonTitle)
				_comparisonTitle = strings.TrimSpace(_comparisonTitle)
			}

			// First, use comparison title, compare without season number
			// e.g. Torrent: "[Seanime] Jujutsu Kaisen 2nd Season - 20 [...].mkv" -> "Jujutsu Kaisen"
			// e.g. Comparison Title: "Jujutsu Kaisen 2nd Season" -> "Jujutsu Kaisen"

			// DEVNOTE: isSeasonAndEpisodeMatch will handle the case where the torrent has a season number

			// Make sure the distance is not too great
			lev := metrics.NewLevenshtein()
			lev.CaseSensitive = false
			res := lev.Distance(torrentTitle, _comparisonTitle)
			if res < 4 {
				return true
			}
		}

		// 2. Use media titles
		// If we're here, it means that either
		// - the comparison title doesn't have a season number
		// - the comparison title (w/o season number) is not similar to the torrent title

		torrentTitleVariations := []*string{&torrentTitle}

		if len(torrentParsedData.SeasonNumber) > 0 {
			season := util.StringToIntMust(torrentParsedData.SeasonNumber[0])
			if season > 1 {
				// If the torrent has a season number, add it to the variations
				torrentTitleVariations = []*string{
					lo.ToPtr(fmt.Sprintf("%s Season %s", torrentParsedData.Title, torrentParsedData.SeasonNumber[0])),
					lo.ToPtr(fmt.Sprintf("%s S%s", torrentParsedData.Title, torrentParsedData.SeasonNumber[0])),
					lo.ToPtr(fmt.Sprintf("%s %s Season", torrentParsedData.Title, util.IntegerToOrdinal(util.StringToIntMust(torrentParsedData.SeasonNumber[0])))),
				}
			}
		}

		// If the parsed comparison title doesn't match, compare the torrent title with media titles
		mediaTitles := listEntry.GetMedia().GetAllTitles()
		var compRes *comparison.SorensenDiceResult
		for _, title := range torrentTitleVariations {
			res, found := comparison.FindBestMatchWithSorensenDice(title, mediaTitles)
			if found {
				if compRes == nil || res.Rating > compRes.Rating {
					compRes = res
				}
			}
		}

		// If the best match is not found
		// /!\ This shouldn't happen since the media titles are always present
		if compRes == nil {
			// Compare using rule comparison title
			sd := metrics.NewSorensenDice()
			sd.CaseSensitive = false
			res := sd.Compare(torrentTitle, comparisonTitle)

			if res > ComparisonThreshold {
				return true
			}
			return false
		}

		// If the best match is found
		if compRes.Rating > ComparisonThreshold {
			return true
		}

		return false
	}
	return false
}

func (ad *AutoDownloader) isSeasonAndEpisodeMatch(
	parsedData *habari.Metadata,
	rule *anime.AutoDownloaderRule,
	listEntry *anilist.AnimeListEntry,
	localEntry *anime.LocalFileWrapperEntry,
	items []*models.AutoDownloaderItem,
) (a int, b bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isSeasonAndEpisodeMatch", func() {
		b = false
	})

	if listEntry == nil {
		return -1, false
	}

	episodes := parsedData.EpisodeNumber

	// Skip if we parsed more than one episode number (e.g. "01-02")
	// We can't handle this case since it might be a batch release
	if len(episodes) > 1 {
		return -1, false
	}

	var ok bool
	episode := 1
	if len(episodes) == 1 {
		_episode, _ok := util.StringToInt(episodes[0])
		if _ok {
			episode = _episode
			ok = true
		}
	}

	// +---------------------+
	// |  No episode number  |
	// +---------------------+

	// We can't parse the episode number
	if !ok {
		// Return true if the media (has only one episode or is a movie) AND (is not in the library)
		if listEntry.GetMedia().GetCurrentEpisodeCount() == 1 || *listEntry.GetMedia().GetFormat() == anilist.MediaFormatMovie {
			// Make sure it wasn't already added
			for _, item := range items {
				if item.Episode == 1 {
					return -1, false // Skip, file already queued or downloaded
				}
			}
			// Make sure it doesn't exist in the library
			if localEntry != nil {
				if _, found := localEntry.FindLocalFileWithEpisodeNumber(1); found {
					return -1, false // Skip, file already exists
				}
			}
			return 1, true // Good to go
		}
		return -1, false
	}

	// +---------------------+
	// |   Episode number    |
	// +---------------------+

	hasAbsoluteEpisode := false

	// Handle ABSOLUTE episode numbers
	if listEntry.GetMedia().GetCurrentEpisodeCount() != -1 && episode > listEntry.GetMedia().GetCurrentEpisodeCount() {
		// Fetch the AniZip media in order to normalize the episode number
		ad.mu.Lock()
		animeMetadata, err := ad.metadataProvider.GetAnimeMetadata(metadata.AnilistPlatform, listEntry.GetMedia().GetID())
		// If the media is found and the offset is greater than 0
		if err == nil && animeMetadata.GetOffset() > 0 {
			hasAbsoluteEpisode = true
			episode = episode - animeMetadata.GetOffset()
		}
		ad.mu.Unlock()
	}

	// Return false if the episode is already downloaded
	for _, item := range items {
		if item.Episode == episode {
			return -1, false // Skip, file already queued or downloaded
		}
	}

	// Return false if the episode is already in the library
	if localEntry != nil {
		if _, found := localEntry.FindLocalFileWithEpisodeNumber(episode); found {
			return -1, false
		}
	}

	// If there's no absolute episode number, check that the episode number is not greater than the current episode count
	if !hasAbsoluteEpisode && episode > listEntry.GetMedia().GetCurrentEpisodeCount() {
		return -1, false
	}

	// As a last check, make sure the seasons match ONLY if the episode number is not absolute
	// We do this check only for "likely" title comparison type since the season numbers are not compared
	if ad.settings.EnableSeasonCheck {
		if !hasAbsoluteEpisode {
			switch rule.TitleComparisonType {
			case anime.AutoDownloaderRuleTitleComparisonLikely:
				// If the title comparison type is "Likely", we will compare the season numbers
				if len(parsedData.SeasonNumber) > 0 {
					season, ok := util.StringToInt(parsedData.SeasonNumber[0])
					if ok && season > 1 {
						parsedComparisonTitle := habari.Parse(rule.ComparisonTitle)
						if len(parsedComparisonTitle.SeasonNumber) == 0 {
							return -1, false
						}
						if season != util.StringToIntMust(parsedComparisonTitle.SeasonNumber[0]) {
							return -1, false
						}
					}
				}
			}
		}
	}

	switch rule.EpisodeType {
	case anime.AutoDownloaderRuleEpisodeRecent:
		// +---------------------+
		// |  Episode "Recent"   |
		// +---------------------+
		// Return false if the user has already watched the episode
		if listEntry.Progress != nil && *listEntry.GetProgress() > episode {
			return -1, false
		}
		return episode, true // Good to go
	case anime.AutoDownloaderRuleEpisodeSelected:
		// +---------------------+
		// | Episode "Selected"  |
		// +---------------------+
		// Return true if the episode is in the list of selected episodes
		for _, ep := range rule.EpisodeNumbers {
			if ep == episode {
				return episode, true // Good to go
			}
		}
		return -1, false
	}
	return -1, false
}

func (ad *AutoDownloader) getRuleListEntry(rule *anime.AutoDownloaderRule) (*anilist.AnimeListEntry, bool) {
	if rule == nil || rule.MediaId == 0 || ad.animeCollection.IsAbsent() {
		return nil, false
	}

	listEntry, found := ad.animeCollection.MustGet().GetListEntryFromAnimeId(rule.MediaId)
	if !found {
		return nil, false
	}

	return listEntry, true
}
