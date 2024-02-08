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
	"github.com/seanime-app/seanime/internal/qbittorrent"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/util"
	"strings"
	"time"
)

const (
	NyaaProvider        = "nyaa"
	ComparisonThreshold = 0.8
)

type (
	anilistListEntry = anilist.AnimeCollection_MediaListCollection_Lists_Entries
	AutoDownloader   struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		Database          *db.Database
		AnilistCollection *anilist.AnimeCollection
		WSEventManager    events.IWSEventManager
		Rules             []*entities.AutoDownloaderRule
		Settings          *models.AutoDownloaderSettings
		AniZipCache       *anizip.Cache
		settingsUpdatedCh chan struct{}
		stopCh            chan struct{}
		startCh           chan struct{}
		active            bool
	}

	NewAutoDownloaderOptions struct {
		Logger            *zerolog.Logger
		QbittorrentClient *qbittorrent.Client
		WSEventManager    events.IWSEventManager
		Rules             []*entities.AutoDownloaderRule
		Database          *db.Database
		AnilistCollection *anilist.AnimeCollection
		AniZipCache       *anizip.Cache
	}
)

func NewAutoDownloader(opts *NewAutoDownloaderOptions) *AutoDownloader {
	return &AutoDownloader{
		Logger:            opts.Logger,
		QbittorrentClient: opts.QbittorrentClient,
		Database:          opts.Database,
		WSEventManager:    opts.WSEventManager,
		Rules:             opts.Rules,
		AnilistCollection: opts.AnilistCollection,
		Settings:          &models.AutoDownloaderSettings{},
		settingsUpdatedCh: make(chan struct{}, 1),
		stopCh:            make(chan struct{}, 1),
		startCh:           make(chan struct{}, 1),
		active:            false,
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings) {
	ad.Settings = settings
	ad.settingsUpdatedCh <- struct{}{} // Notify that the settings have been updated
	if ad.Settings.Enabled && !ad.active {
		ad.startCh <- struct{}{} // Start the auto downloader
	} else if !ad.Settings.Enabled && ad.active {
		ad.stopCh <- struct{}{} // Stop the auto downloader
	}
}

// Start will start the auto downloader.
// This should be run in a goroutine.
func (ad *AutoDownloader) Start() {
	ad.Logger.Info().Msg("autodownloader: Starting module")

	started := ad.QbittorrentClient.CheckStart() // Start qBittorrent if it's not running
	if !started {
		ad.Logger.Error().Msg("autodownloader: Failed to start qBittorrent. Make sure it's running for the Auto Downloader to work.")
		return
	}

	// Start the auto downloader
	ad.start()
}

func (ad *AutoDownloader) start() {
	ad.Logger.Info().Msg("autodownloader: Started")

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
			ad.active = false
			ad.Logger.Debug().Msg("autodownloader: Auto Downloader stopped")
		case <-ad.startCh:
			ad.active = true
			ad.Logger.Debug().Msg("autodownloader: Auto Downloader started")
			ad.checkForNewEpisodes()
		case <-ticker.C:
			if ad.active {
				ad.checkForNewEpisodes()
			}
		}
		ticker.Stop()
	}

}

func (ad *AutoDownloader) checkForNewEpisodes() {
	torrents := make([]*NormalizedTorrent, 0)

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
	}

	// Going through each rule
	for _, rule := range ad.Rules {
		if !rule.Enabled {
			continue // Skip rule
		}
		listEntry, found := ad.getRuleListEntry(rule)
		// If the media is not found, skip the rule
		if !found {
			continue // Skip rule
		}
		// If the media is not releasing AND has more than one episode, skip the rule
		// This is to avoid skipping movies and single-episode OVAs
		if *listEntry.GetMedia().GetStatus() != anilist.MediaStatusReleasing && listEntry.GetMedia().GetCurrentEpisodeCount() > 1 {
			continue // Skip rule
		}

		// Create a LocalFileWrapper
		lfWrapper := entities.NewLocalFileWrapper(lfs)
		localEntry, found := lfWrapper.GetLocalEntryById(listEntry.GetMedia().GetID())
		if !found {
			continue // Skip rule
		}

		for _, t := range torrents {
			if episode, ok := ad.torrentFollowsRule(t, rule, listEntry, localEntry); ok {
				ad.downloadTorrent(t, rule, episode)
			}
		}
	}

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

	if ok := ad.isResolutionMatch(t.ParsedData.ReleaseGroup, rule); !ok {
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

	started := ad.QbittorrentClient.CheckStart() // Start qBittorrent if it's not running
	if !started {
		ad.Logger.Error().Str("link", t.Link).Msg("autodownloader: Failed to download torrent. qBittorrent is not running.")
		return
	}

	ad.Logger.Debug().Msgf("autodownloader: Downloading torrent: %s", t.Name)

	magnet, found := t.GetMagnet()
	if !found {
		ad.Logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to get magnet link for torrent")
		return
	}

	downloaded := false

	// Pause the torrent when it's added
	if ad.Settings.DownloadAutomatically {

		// Add the torrent to qBittorrent
		err := ad.QbittorrentClient.Torrent.AddURLs([]string{magnet}, &qbittorrent_model.AddTorrentsOptions{
			Savepath: rule.Destination,
		})
		if err != nil {
			ad.Logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to qBittorrent")
			return
		}

		downloaded = true
	}

	ad.Logger.Debug().Str("name", t.Name).Msg("autodownloader: Added torrent")
	ad.WSEventManager.SendEvent(events.AutoDownloaderTorrentAdded, t.Name)

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

func (ad *AutoDownloader) isResolutionMatch(quality string, rule *entities.AutoDownloaderRule) bool {
	if len(rule.Resolutions) == 0 {
		return true
	}
	for _, q := range rule.Resolutions {
		qualityWithoutP, _ := strings.CutSuffix(q, "p")
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

		return strings.Contains(strings.ToLower(torrentTitle), strings.ToLower(rule.ComparisonTitle))
	case entities.AutoDownloaderRuleTitleComparisonLikely:
		// +---------------------+
		// |   Title "Likely"    |
		// +---------------------+

		// First, use comparison title
		ok := strings.Contains(strings.ToLower(torrentTitle), strings.ToLower(rule.ComparisonTitle))
		if ok {
			return true
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
	if listEntry == nil || localEntry == nil {
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
	if len(episodes) > 0 {
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
			if _, found := localEntry.FindLocalFileWithEpisodeNumber(1); found {
				return -1, false // Skip, file already exists
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
		anizipMedia, err := anizip.FetchAniZipMediaC("anilist", listEntry.GetMedia().GetID(), ad.AniZipCache)
		// If the media is found and the offset is greater than 0
		if err == nil && anizipMedia.GetOffset() > 0 {
			episode = episode - anizipMedia.GetOffset()
		}
	}

	// Return false if the episode is already downloaded
	for _, item := range items {
		if item.Episode == episode {
			return -1, false // Skip, file already downloaded
		}
	}

	// Return false if the episode is already in the library
	if _, found := localEntry.FindLocalFileWithEpisodeNumber(episode); found {
		return -1, false
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
	if rule == nil || rule.MediaId == 0 {
		return nil, false
	}

	listEntry, found := ad.AnilistCollection.GetListEntryFromMediaId(rule.MediaId)
	if !found {
		return nil, false
	}

	return listEntry, true
}
