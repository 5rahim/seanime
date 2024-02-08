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
	ad.Logger.Info().Msg("autodownloader: Starting auto downloader module")

	// Start the auto downloader
	ad.start()
}

func (ad *AutoDownloader) start() {

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

		listEntry, found := ad.getRuleListEntry(rule)
		// If the media is not found, skip the rule
		if !found {
			continue // Skip rule
		}
		// If the media is not releasing AND has more than one episode, skip the rule
		// e.g. Movies might be set to "Finished" but we still want to download them
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
			if ad.torrentFollowsRule(t, rule, listEntry, localEntry) {
				ad.downloadTorrent(t, rule)
			}
		}
	}

}

func (ad *AutoDownloader) torrentFollowsRule(
	t *NormalizedTorrent,
	rule *entities.AutoDownloaderRule,
	listEntry *anilistListEntry,
	localEntry *entities.LocalFileWrapperEntry,
) bool {

	if ok := ad.isReleaseGroupMatch(t.ParsedData.ReleaseGroup, rule); !ok {
		return false
	}

	if ok := ad.isTitleMatch(t.ParsedData.Title, rule, listEntry); !ok {
		return false
	}

	if ok := ad.isEpisodeMatch(t.ParsedData.EpisodeNumber, rule, listEntry, localEntry); !ok {
		return false
	}

	return false
}

func (ad *AutoDownloader) downloadTorrent(t *NormalizedTorrent, rule *entities.AutoDownloaderRule) {

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
) bool {
	if listEntry == nil || localEntry == nil {
		return false
	}

	// Skip if we parsed more than one episode number (e.g. "01-02")
	// We can't handle this case since it might be a batch release
	if len(episodes) > 0 {
		return false
	}

	episode, ok := util.StringToInt(episodes[0])

	// +---------------------+
	// |  No episode number  |
	// +---------------------+

	// We can't parse the episode number
	if !ok {
		// Return true if the media (has only one episode or is a movie) AND (is not in the library)
		if listEntry.GetMedia().GetCurrentEpisodeCount() == 1 || *listEntry.GetMedia().GetFormat() == anilist.MediaFormatMovie {
			if _, found := localEntry.FindLocalFileWithEpisodeNumber(1); !found {
				return true
			}
		}
		return false
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

	// Return false if the episode is already in the library
	if _, found := localEntry.FindLocalFileWithEpisodeNumber(episode); found {
		return false
	}

	switch rule.EpisodeType {

	case entities.AutoDownloaderRuleEpisodeRecent:
		// +---------------------+
		// |  Episode "Recent"   |
		// +---------------------+
		// Return false if the user has already watched the episode
		if listEntry.Progress != nil && *listEntry.GetProgress() > episode {
			return false
		}
		return true // Good to go
	case entities.AutoDownloaderRuleEpisodeSelected:
		// +---------------------+
		// | Episode "Selected"  |
		// +---------------------+
		// Return true if the episode is in the list of selected episodes
		for _, ep := range rule.EpisodeNumbers {
			if ep == episode {
				return true // Good to go
			}
		}
		return false
	}
	return false
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
