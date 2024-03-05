package autodownloader

import (
	"github.com/adrg/strutil/metrics"
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/db"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/torrent_client"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	NyaaProvider        = "nyaa"
	AnimeToshoProvider  = "animetosho"
	ComparisonThreshold = 0.8
)

type (
	anilistListEntry = anilist.AnimeCollection_MediaListCollection_Lists_Entries
	AutoDownloader   struct {
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		Database                *db.Database
		AnilistCollection       *anilist.AnimeCollection
		WSEventManager          events.IWSEventManager
		Settings                *models.AutoDownloaderSettings
		AniZipCache             *anizip.Cache
		settingsUpdatedCh       chan struct{}
		stopCh                  chan struct{}
		startCh                 chan struct{}
		debugTrace              bool
		mu                      sync.Mutex
	}

	NewAutoDownloaderOptions struct {
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		WSEventManager          events.IWSEventManager
		Database                *db.Database
		AnilistCollection       *anilist.AnimeCollection
		AniZipCache             *anizip.Cache
	}

	tmpTorrentToDownload struct {
		torrent *NormalizedTorrent
		episode int
	}
)

func NewAutoDownloader(opts *NewAutoDownloaderOptions) *AutoDownloader {
	return &AutoDownloader{
		Logger:                  opts.Logger,
		TorrentClientRepository: opts.TorrentClientRepository,
		Database:                opts.Database,
		WSEventManager:          opts.WSEventManager,
		AnilistCollection:       opts.AnilistCollection,
		AniZipCache:             opts.AniZipCache,
		Settings: &models.AutoDownloaderSettings{
			Provider:              NyaaProvider, // Default provider, will be updated after the settings are fetched
			Interval:              10,
			Enabled:               false,
			DownloadAutomatically: false,
		},
		settingsUpdatedCh: make(chan struct{}, 1),
		stopCh:            make(chan struct{}, 1),
		startCh:           make(chan struct{}, 1),
		debugTrace:        true,
		mu:                sync.Mutex{},
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
// If the AutoDownloader is not active, it will start it if the settings are enabled.
// If the AutoDownloader is active, it will stop it if the settings are disabled.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings, provider string) {
	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		defer ad.mu.Unlock()
		ad.Settings = settings
		// Update the provider if it's provided
		if provider != "" {
			ad.Settings.Provider = provider
		}
		ad.settingsUpdatedCh <- struct{}{} // Notify that the settings have been updated
		if ad.Settings.Enabled {
			ad.startCh <- struct{}{} // Start the auto downloader
		} else if !ad.Settings.Enabled {
			ad.stopCh <- struct{}{} // Stop the auto downloader
		}
	}()
}

func (ad *AutoDownloader) SetAnilistCollection(collection *anilist.AnimeCollection) {
	if ad == nil {
		return
	}
	ad.AnilistCollection = collection
}

// Start will start the auto downloader in a goroutine
func (ad *AutoDownloader) Start() {
	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		if ad.Settings.Enabled {
			started := ad.TorrentClientRepository.Start() // Start torrent client if it's not running
			if !started {
				ad.Logger.Warn().Msg("autodownloader: Failed to start torrent client. Make sure it's running for the Auto Downloader to work.")
				return
			}
		}
		ad.mu.Unlock()

		// Start the auto downloader
		ad.start()
	}()
}

func (ad *AutoDownloader) Run() {
	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		defer ad.mu.Unlock()
		ad.startCh <- struct{}{}
	}()
}

// CleanUpDownloadedItems will clean up downloaded items from the database.
// This should be run after a scan is completed.
func (ad *AutoDownloader) CleanUpDownloadedItems() {
	if ad == nil {
		return
	}
	ad.mu.Lock()
	defer ad.mu.Unlock()
	err := ad.Database.DeleteDownloadedAutoDownloaderItems()
	if err != nil {
		return
	}
}

func (ad *AutoDownloader) start() {
	if ad.Settings.Enabled {
		ad.Logger.Info().Msg("autodownloader: Module started")
	}

	for {
		interval := 10
		if ad.Settings != nil && ad.Settings.Interval > 0 {
			interval = ad.Settings.Interval
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Minute)
		select {
		case <-ad.settingsUpdatedCh:
			break // Restart the loop
		case <-ad.stopCh:
			ad.Logger.Info().Msg("autodownloader: Auto Downloader stopped")
		case <-ad.startCh:
			if ad.Settings.Enabled {
				ad.Logger.Info().Msg("autodownloader: Auto Downloader started")
				ad.checkForNewEpisodes()
			}
		case <-ticker.C:
			if ad.Settings.Enabled {
				ad.checkForNewEpisodes()
			}
		}
		ticker.Stop()
	}

}

func (ad *AutoDownloader) checkForNewEpisodes() {
	ad.mu.Lock()
	if ad == nil || !ad.Settings.Enabled {
		return
	}
	ad.mu.Unlock()

	torrents := make([]*NormalizedTorrent, 0)

	// Get rules from the database
	rules, err := ad.Database.GetAutoDownloaderRules()
	if err != nil {
		ad.Logger.Error().Err(err).Msg("autodownloader: Failed to fetch rules from the database")
		return
	}

	// Get local files from the database
	lfs, _, err := ad.Database.GetLocalFiles()
	if err != nil {
		ad.Logger.Error().Err(err).Msg("autodownloader: Failed to fetch local files from the database")
		return
	}

	if ad.Settings.Provider == NyaaProvider {
		nyaaTorrents, err := ad.getCurrentTorrentsFromNyaa()
		if err != nil {
			ad.Logger.Error().Err(err).Msg("autodownloader: Failed to fetch torrents from Nyaa")
		} else {
			torrents = nyaaTorrents
		}
	} else if ad.Settings.Provider == AnimeToshoProvider {
		toshoTorrents, err := ad.getCurrentTorrentsFromAnimeTosho()
		if err != nil {
			ad.Logger.Error().Err(err).Msg("autodownloader: Failed to fetch torrents from AnimeTosho")
		} else {
			torrents = toshoTorrents
		}
	}

	// Get existing torrents
	existingTorrents := make([]*torrent_client.Torrent, 0)
	if ad.TorrentClientRepository != nil {
		existingTorrents, err = ad.TorrentClientRepository.GetList()
		if err != nil {
			existingTorrents = make([]*torrent_client.Torrent, 0)
		}
	}

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

			// If the media is not releasing AND has more than one episode, skip the rule
			// This is to avoid skipping movies and single-episode OVAs
			if *listEntry.GetMedia().GetStatus() != anilist.MediaStatusReleasing && listEntry.GetMedia().GetCurrentEpisodeCount() > 1 {
				return // Skip rule
			}

			// Create a LocalFileWrapper
			lfWrapper := entities.NewLocalFileWrapper(lfs)
			localEntry, _ := lfWrapper.GetLocalEntryById(listEntry.GetMedia().GetID())

			// Get all torrents that follow the rule
			torrentsToDownload := make([]*tmpTorrentToDownload, 0)
		outer:
			for _, t := range torrents {
				// If the torrent is already added, skip it
				for _, et := range existingTorrents {
					if et.Hash == t.Hash {
						continue outer // Skip the torrent
					}
				}

				episode, ok := ad.torrentFollowsRule(t, rule, listEntry, localEntry)
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
				ad.downloadTorrent(t.torrent, rule, t.episode)
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
					ad.downloadTorrent(torrents[0].torrent, rule, ep)
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

				ad.downloadTorrent(torrents[0].torrent, rule, ep)
			}
		})
	}
	p.Wait()

}

func (ad *AutoDownloader) torrentFollowsRule(
	t *NormalizedTorrent,
	rule *entities.AutoDownloaderRule,
	listEntry *anilistListEntry,
	localEntry *entities.LocalFileWrapperEntry,
) (int, bool) {

	if ok := ad.isReleaseGroupMatch(t.ParsedData.ReleaseGroup, rule); !ok {
		return -1, false
	}

	if ok := ad.isResolutionMatch(t.ParsedData.VideoResolution, rule); !ok {
		return -1, false
	}

	if ok := ad.isTitleMatch(t.ParsedData.Title, rule, listEntry); !ok {
		return -1, false
	}

	episode, ok := ad.isEpisodeMatch(t.ParsedData.EpisodeNumber, rule, listEntry, localEntry)
	if !ok {
		return -1, false
	}

	return episode, true
}

func (ad *AutoDownloader) downloadTorrent(t *NormalizedTorrent, rule *entities.AutoDownloaderRule, episode int) {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	if ad.TorrentClientRepository == nil {
		ad.Logger.Error().Msg("autodownloader: torrent client not found")
		return
	}

	started := ad.TorrentClientRepository.Start() // Start torrent client if it's not running
	if !started {
		ad.Logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to download torrent. torrent client is not running.")
		return
	}

	// Return if the torrent is already added
	torrentExists := ad.TorrentClientRepository.TorrentExists(t.Hash)
	if torrentExists {
		//ad.Logger.Debug().Str("name", t.Name).Msg("autodownloader: Torrent already added")
		return
	}

	magnet, found := t.GetMagnet()
	if !found {
		ad.Logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to get magnet link for torrent")
		return
	}

	downloaded := false

	// Pause the torrent when it's added
	if ad.Settings.DownloadAutomatically {

		ad.Logger.Debug().Msgf("autodownloader: Downloading torrent: %s", t.Name)

		// Add the torrent to torrent client
		err := ad.TorrentClientRepository.AddMagnets([]string{magnet}, rule.Destination)
		if err != nil {
			ad.Logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to torrent client")
			return
		}

		downloaded = true
	}

	ad.Logger.Info().Str("name", t.Name).Msg("autodownloader: Added torrent")
	ad.WSEventManager.SendEvent(events.AutoDownloaderItemAdded, t.Name)

	// Add the torrent to the database
	item := &models.AutoDownloaderItem{
		RuleID:      rule.DbID,
		MediaID:     rule.MediaId,
		Episode:     episode,
		Link:        t.Link,
		Hash:        t.Hash,
		TorrentName: t.Name,
		Magnet:      magnet,
		Downloaded:  downloaded,
	}
	_ = ad.Database.InsertAutoDownloaderItem(item)

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ad *AutoDownloader) isReleaseGroupMatch(releaseGroup string, rule *entities.AutoDownloaderRule) bool {
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
func (ad *AutoDownloader) isResolutionMatch(quality string, rule *entities.AutoDownloaderRule) bool {
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

func (ad *AutoDownloader) isTitleMatch(torrentTitle string, rule *entities.AutoDownloaderRule, listEntry *anilistListEntry) bool {
	switch rule.TitleComparisonType {
	case entities.AutoDownloaderRuleTitleComparisonContains:
		// +---------------------+
		// |   Title "Contains"  |
		// +---------------------+

		if strings.Contains(strings.ToLower(torrentTitle), strings.ToLower(rule.ComparisonTitle)) {
			// Make sure the distance is not too great
			lev := metrics.NewLevenshtein()
			lev.CaseSensitive = false
			res := lev.Distance(torrentTitle, rule.ComparisonTitle)
			if res < 30 {
				return true
			}
			return false
		}
	case entities.AutoDownloaderRuleTitleComparisonLikely:
		// +---------------------+
		// |   Title "Likely"    |
		// +---------------------+

		// First, use comparison title
		ok := strings.Contains(strings.ToLower(torrentTitle), strings.ToLower(rule.ComparisonTitle))
		if ok {
			// Make sure the distance is not too great
			lev := metrics.NewLevenshtein()
			lev.CaseSensitive = false
			res := lev.Distance(torrentTitle, rule.ComparisonTitle)
			if res < 4 {
				return true
			}
		}

		titles := listEntry.GetMedia().GetAllTitles()
		res, found := comparison.FindBestMatchWithSorensenDice(&torrentTitle, titles)

		// If the best match is not found
		if !found {
			// Compare using rule comparison title
			lev := metrics.NewSorensenDice()
			lev.CaseSensitive = false
			res := lev.Compare(torrentTitle, rule.ComparisonTitle)

			if res > ComparisonThreshold {
				return true
			}
			return false
		}

		// If the best match is found
		if res.Rating > ComparisonThreshold {
			return true
		}

		return false
	}
	return false
}

func (ad *AutoDownloader) isEpisodeMatch(
	episodes []string,
	rule *entities.AutoDownloaderRule,
	listEntry *anilistListEntry,
	localEntry *entities.LocalFileWrapperEntry,
) (int, bool) {
	if listEntry == nil {
		return -1, false
	}

	// +---------------------+
	// |    Existing Item    |
	// +---------------------+
	items, err := ad.Database.GetAutoDownloaderItemByMediaId(listEntry.GetMedia().GetID())
	if err != nil {
		items = make([]*models.AutoDownloaderItem, 0)
	}

	// Skip if we parsed more than one episode number (e.g. "01-02")
	// We can't handle this case since it might be a batch release
	if len(episodes) > 1 {
		return -1, false
	}

	episode, ok := util.StringToInt(episodes[0])

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
					return -1, false // Skip, file already downloaded
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

	// Handle ABSOLUTE episode numbers
	if listEntry.GetMedia().GetCurrentEpisodeCount() != -1 && episode > listEntry.GetMedia().GetCurrentEpisodeCount() {
		// Fetch the AniZip media in order to normalize the episode number
		ad.mu.Lock()
		anizipMedia, err := anizip.FetchAniZipMediaC("anilist", listEntry.GetMedia().GetID(), ad.AniZipCache)
		// If the media is found and the offset is greater than 0
		if err == nil && anizipMedia.GetOffset() > 0 {
			episode = episode - anizipMedia.GetOffset()
		}
		ad.mu.Unlock()
	}

	// Return false if the episode is already downloaded
	for _, item := range items {
		if item.Episode == episode {
			return -1, false // Skip, file already downloaded
		}
	}

	// Return false if the episode is already in the library
	if localEntry != nil {
		if _, found := localEntry.FindLocalFileWithEpisodeNumber(episode); found {
			return -1, false
		}
	}

	switch rule.EpisodeType {

	case entities.AutoDownloaderRuleEpisodeRecent:
		// +---------------------+
		// |  Episode "Recent"   |
		// +---------------------+
		// Return false if the user has already watched the episode
		if listEntry.Progress != nil && *listEntry.GetProgress() > episode {
			return -1, false
		}
		return episode, true // Good to go
	case entities.AutoDownloaderRuleEpisodeSelected:
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

func (ad *AutoDownloader) getRuleListEntry(rule *entities.AutoDownloaderRule) (*anilistListEntry, bool) {
	if rule == nil || rule.MediaId == 0 || ad.AnilistCollection == nil {
		return nil, false
	}

	listEntry, found := ad.AnilistCollection.GetListEntryFromMediaId(rule.MediaId)
	if !found {
		return nil, false
	}

	return listEntry, true
}
