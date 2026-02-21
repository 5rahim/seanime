package autodownloader

import (
	"context"
	"encoding/json"
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/database/db_bridge"
	"seanime/internal/database/models"
	debrid_client "seanime/internal/debrid/client"
	"seanime/internal/debrid/debrid"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/hook"
	"seanime/internal/library/anime"
	"seanime/internal/notifier"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/util"
	"seanime/internal/util/comparison"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/5rahim/habari"
	"github.com/adrg/strutil/metrics"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
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
		metadataProviderRef     *util.Ref[metadata_provider.Provider]
		settingsUpdatedCh       chan struct{}
		stopCh                  chan struct{}
		startCh                 chan bool
		debugTrace              bool
		mu                      sync.Mutex
		isOfflineRef            *util.Ref[bool]
		simulationResults       []*SimulationResult // Stores results when running in simulation mode
	}

	// SimulationResult represents a torrent that would be downloaded in simulation mode
	SimulationResult struct {
		RuleID      uint   `json:"ruleId"`
		MediaID     int    `json:"mediaId"`
		Episode     int    `json:"episode"`
		Link        string `json:"link"`
		Hash        string `json:"hash"`
		TorrentName string `json:"torrentName"`
		Score       int    `json:"score"`
		ExtensionID string `json:"extensionId"`
		IsDelayed   bool   `json:"isDelayed"`
	}

	NewAutoDownloaderOptions struct {
		Logger                  *zerolog.Logger
		TorrentClientRepository *torrent_client.Repository
		TorrentRepository       *torrent.Repository
		WSEventManager          events.WSEventManagerInterface
		Database                *db.Database
		MetadataProviderRef     *util.Ref[metadata_provider.Provider]
		DebridClientRepository  *debrid_client.Repository
		IsOfflineRef            *util.Ref[bool]
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
		metadataProviderRef:     opts.MetadataProviderRef,
		debridClientRepository:  opts.DebridClientRepository,
		settings: &models.AutoDownloaderSettings{
			Provider:              "", // Default provider, will be updated after the settings are fetched
			Interval:              20,
			Enabled:               false,
			DownloadAutomatically: false,
			EnableEnhancedQueries: false,
		},
		settingsUpdatedCh: make(chan struct{}, 1),
		stopCh:            make(chan struct{}, 1),
		startCh:           make(chan bool, 1),
		debugTrace:        true,
		mu:                sync.Mutex{},
		isOfflineRef:      opts.IsOfflineRef,
		simulationResults: make([]*SimulationResult, 0),
	}
}

// SetSettings should be called after the settings are fetched and updated from the database.
// If the AutoDownloader is not active, it will start it if the settings are enabled.
// If the AutoDownloader is active, it will stop it if the settings are disabled.
func (ad *AutoDownloader) SetSettings(settings *models.AutoDownloaderSettings) {
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
		ad.settingsUpdatedCh <- struct{}{} // Notify that the settings have been updated
		if ad.settings.Enabled {
			ad.startCh <- false // Start the auto downloader
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

// GetSimulationResults returns the simulation results from the last run
func (ad *AutoDownloader) GetSimulationResults() []*SimulationResult {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	return ad.simulationResults
}

// ClearSimulationResults clears the simulation results
func (ad *AutoDownloader) ClearSimulationResults() {
	ad.mu.Lock()
	defer ad.mu.Unlock()
	ad.simulationResults = make([]*SimulationResult, 0)
}

// RunCheck runs the auto downloader synchronously for testing purposes
// This directly calls checkForNewEpisodes without using goroutines
func (ad *AutoDownloader) RunCheck(ctx context.Context, isSimulation bool, ruleIDs ...uint) {
	defer util.HandlePanicInModuleThen("autodownloader/RunCheck", func() {})

	if ad == nil {
		return
	}
	ad.checkForNewEpisodes(ctx, isSimulation, ruleIDs...)
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

func (ad *AutoDownloader) Run(isSimulation bool) {
	defer util.HandlePanicInModuleThen("autodownloader/Run", func() {})

	if ad == nil {
		return
	}
	go func() {
		ad.mu.Lock()
		defer ad.mu.Unlock()
		ad.startCh <- isSimulation
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

		case isSumulation := <-ad.startCh:
			if ad.settings.Enabled {
				ad.logger.Debug().Msg("autodownloader: Auto Downloader started")
				ad.checkForNewEpisodes(context.Background(), isSumulation)
			}
		case <-ticker.C:
			if ad.settings.Enabled {
				ad.checkForNewEpisodes(context.Background(), false)
			}
		}
		ticker.Stop()
	}

}

// checkForNewEpisodes will check the RSS feeds for new episodes.
func (ad *AutoDownloader) checkForNewEpisodes(ctx context.Context, isSimulation bool, ruleIDs ...uint) {
	defer util.HandlePanicInModuleThen("autodownloader/checkForNewEpisodes", func() {})

	if ad.isOfflineRef.Get() {
		ad.logger.Debug().Msg("autodownloader: Skipping check for new episodes. AutoDownloader is in offline mode.")
		return
	}

	ad.mu.Lock()
	if ad.torrentRepository == nil || !ad.settings.Enabled || ad.settings.Provider == torrent.ProviderNone {
		ad.logger.Warn().Msg("autodownloader: Could not check for new episodes. AutoDownloader is not enabled or provider is not set.")
		ad.mu.Unlock()
		return
	}
	ad.mu.Unlock()

	// Fetch all necessary data
	data, err := ad.fetchRunData(ctx, ruleIDs...)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to fetch check data")
		return
	}

	// Event
	event := &AutoDownloaderRunStartedEvent{
		Rules:    data.rules,
		Profiles: data.profiles,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderRunStarted().Trigger(event)
	data.rules = event.Rules

	// Default prevented, return
	if event.DefaultPrevented {
		return
	}

	// If there are no rules, return
	if len(data.rules) == 0 {
		ad.logger.Debug().Msg("autodownloader: No rules found")
		return
	}

	// Group matched torrents by rule and episode
	groupedCandidates := ad.groupTorrentCandidates(data)

	// Select best candidates and download
	downloaded := ad.selectAndDownloadBestCandidates(isSimulation, groupedCandidates, data.rules, data.profiles)

	// Download delayed items that can be downloaded
	delayedDownloaded := ad.downloadDelayedItems(isSimulation)
	downloaded += delayedDownloaded

	// Notify user
	ad.notifyDownloadResults(downloaded)
}

// runData holds all data needed for checking new episodes
type runData struct {
	rules            []*anime.AutoDownloaderRule
	profiles         []*anime.AutoDownloaderProfile
	localFileWrapper *anime.LocalFileWrapper
	torrents         []*NormalizedTorrent
	existingTorrents []*torrent_client.Torrent
}

// fetchRunData fetches all data needed for checking new episodes
func (ad *AutoDownloader) fetchRunData(ctx context.Context, ruleIDs ...uint) (*runData, error) {
	// Get rules from the database
	rules, err := db_bridge.GetAutoDownloaderRules(ad.database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rules: %w", err)
	}

	// Filter out disabled rules
	rules = lo.Filter(rules, func(r *anime.AutoDownloaderRule, _ int) bool {
		if len(ruleIDs) > 0 && !slices.Contains(ruleIDs, r.DbID) {
			return false
		}
		return r.Enabled
	})

	// Get profiles from the database
	profiles, err := db_bridge.GetAutoDownloaderProfiles(ad.database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profiles: %w", err)
	}

	// Filter out extensions that don't exist
	if ad.torrentRepository != nil {
		for _, rule := range rules {
			if len(rule.Providers) > 0 {
				rule.Providers = lo.Filter(rule.Providers, func(id string, _ int) bool {
					_, found := ad.torrentRepository.GetAnimeProviderExtension(id)
					return found
				})
			}
		}
		for _, profile := range profiles {
			if len(profile.Providers) > 0 {
				profile.Providers = lo.Filter(profile.Providers, func(id string, _ int) bool {
					_, found := ad.torrentRepository.GetAnimeProviderExtension(id)
					return found
				})
			}
		}
	}

	// Get local files from the database
	lfs, _, err := db_bridge.GetLocalFiles(ad.database)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch local files: %w", err)
	}
	lfWrapper := anime.NewLocalFileWrapper(lfs)

	// Identify distinct providers from rules and profiles
	// Returns the default provider + any other provider used by rules or profiles
	providerExtensions := ad.getProvidersForRules(rules, profiles)

	// Fetch torrents from all identified providers
	torrents, err := ad.fetchTorrentsFromProviders(ctx, providerExtensions, rules, profiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest torrents: %w", err)
	}

	// Event
	fetchedEvent := &AutoDownloaderTorrentsFetchedEvent{
		Torrents: torrents,
	}
	_ = hook.GlobalHookManager.OnAutoDownloaderTorrentsFetched().Trigger(fetchedEvent)
	torrents = fetchedEvent.Torrents

	// Get existing torrents
	var existingTorrents []*torrent_client.Torrent
	if ad.torrentClientRepository != nil {
		existingTorrents, _ = ad.torrentClientRepository.GetList(&torrent_client.GetListOptions{})
	}

	return &runData{
		rules:            rules,
		profiles:         profiles,
		localFileWrapper: lfWrapper,
		torrents:         torrents,
		existingTorrents: existingTorrents,
	}, nil
}

// Candidate represents a potential torrent to download with its score
type Candidate struct {
	Torrent *NormalizedTorrent
	Score   int
}

// groupTorrentCandidates groups torrents by rule ID and episode number
// Returns: Map: Rule ID -> Episode Number -> List of candidates
func (ad *AutoDownloader) groupTorrentCandidates(data *runData) map[uint]map[int][]*Candidate {
	groupedCandidates := make(map[uint]map[int][]*Candidate)

	for _, rule := range data.rules {
		if rule.MediaId == 0 {
			continue
		}

		// Get the rule's profiles (global + specific)
		ruleProfiles := ad.getRuleProfiles(rule, data.profiles)

		listEntry, ok := ad.getRuleListEntry(rule)
		if !ok {
			continue
		}

		// Get all queued items from this media
		ruleQueuedItems, _ := ad.database.GetAutoDownloaderItemByMediaId(listEntry.GetMedia().GetID())

		// Initialize map for this rule
		groupedCandidates[rule.DbID] = make(map[int][]*Candidate)

		// Process each torrent
		for _, t := range data.torrents {
			// Skip if already exists
			if ad.isTorrentAlreadyDownloaded(t, data.existingTorrents) {
				continue
			}

			// Check if torrent matches rule
			episode, follows := ad.torrentFollowsRule(t, rule, listEntry, ruleProfiles)
			if !follows || episode == -1 {
				continue
			}

			// Skip if already in library or queue (not delayed)
			if ad.isEpisodeAlreadyHandled(episode, rule.CustomEpisodeNumberAbsoluteOffset, rule.DbID, rule.MediaId, data.localFileWrapper, ruleQueuedItems) {
				continue
			}

			// Calculate score
			score, requiredMinScore := ad.calculateCandidateScore(t, ruleProfiles)

			// Skip if score doesn't meet minimum
			if score < requiredMinScore {
				continue
			}

			// Add to candidates
			if groupedCandidates[rule.DbID][episode] == nil {
				groupedCandidates[rule.DbID][episode] = make([]*Candidate, 0)
			}
			groupedCandidates[rule.DbID][episode] = append(groupedCandidates[rule.DbID][episode], &Candidate{
				Torrent: t,
				Score:   score,
			})
		}
	}

	return groupedCandidates
}

// getRuleProfiles returns all profiles that apply to a rule (global + specific)
func (ad *AutoDownloader) getRuleProfiles(rule *anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) []*anime.AutoDownloaderProfile {
	ruleProfiles := make([]*anime.AutoDownloaderProfile, 0)
	for _, profile := range profiles {
		if profile.Global {
			ruleProfiles = append(ruleProfiles, profile)
		} else if rule.ProfileID != nil && profile.DbID == *rule.ProfileID {
			// Add profile assigned to the rule
			ruleProfiles = append(ruleProfiles, profile)
		}
	}
	return ruleProfiles
}

// isTorrentAlreadyDownloaded checks if a torrent already exists in the client
func (ad *AutoDownloader) isTorrentAlreadyDownloaded(t *NormalizedTorrent, existingTorrents []*torrent_client.Torrent) bool {
	for _, et := range existingTorrents {
		if et.Hash == t.InfoHash {
			return true
		}
	}
	return false
}

// isEpisodeAlreadyHandled checks if an episode is already in the library or queue but not delayed
func (ad *AutoDownloader) isEpisodeAlreadyHandled(episode int, absoluteOffset int, ruleId uint, mediaId int, lfWrapper *anime.LocalFileWrapper, queuedItems []*models.AutoDownloaderItem) bool {
	// Check if already in the library
	le, found := lfWrapper.GetLocalEntryById(mediaId)
	if found {
		if _, ok := le.FindLocalFileWithEpisodeNumber(episode); ok {
			return true
		}
		// Check for the episode number by taking the custom offset into account
		if absoluteOffset != 0 {
			if _, ok := le.FindLocalFileWithEpisodeNumber(episode - absoluteOffset); ok {
				return true
			}
		}
	}

	ac, ok := ad.animeCollection.Get()
	if ok {
		// Episode has already been watched
		listEntry, found := ac.GetListEntryFromAnimeId(mediaId)
		if found && listEntry.GetProgressSafe() >= episode {
			return true
		}
	}

	// Check if already in the queue for this specific rule
	for _, item := range queuedItems {
		if item.IsDelayed {
			continue
		}
		// Check rule id again
		if item.Episode == episode && item.RuleID == ruleId {
			return true
		}
		// Check for the episode number by taking the custom offset into account
		if absoluteOffset != 0 {
			if item.Episode == episode-absoluteOffset && item.RuleID == ruleId {
				return true
			}
		}
	}

	return false
}

// calculateCandidateScore calculates the score for a torrent candidate based on profiles
// Returns the total score and the required minimum score
func (ad *AutoDownloader) calculateCandidateScore(t *NormalizedTorrent, ruleProfiles []*anime.AutoDownloaderProfile) (score int, requiredMinScore int) {
	score = 0
	requiredMinScore = 0

	for _, p := range ruleProfiles {
		score += ad.calculateTorrentScore(t, p)

		// If multiple profiles are active, enforce the strictest threshold
		if p.MinimumScore > requiredMinScore {
			requiredMinScore = p.MinimumScore
		}
	}

	return score, requiredMinScore
}

// delaySettings holds the delay configuration for a rule
type delaySettings struct {
	hasDelay       bool
	delayMinutes   int
	skipDelayScore int
}

// getDelaySettings extracts delay configuration from profiles
// Uses the highest delay and skip delay score.
func (ad *AutoDownloader) getDelaySettings(rule *anime.AutoDownloaderRule, ruleProfiles []*anime.AutoDownloaderProfile) delaySettings {
	settings := delaySettings{
		hasDelay:       false,
		delayMinutes:   0,
		skipDelayScore: 0,
	}

	for _, p := range ruleProfiles {
		if p.DelayMinutes > 0 {
			settings.hasDelay = true
			if p.DelayMinutes > settings.delayMinutes {
				settings.delayMinutes = p.DelayMinutes
			}
			if p.SkipDelayScore > settings.skipDelayScore {
				settings.skipDelayScore = p.SkipDelayScore
			}
		}
	}

	return settings
}

// findStoredItemForEpisode finds an existing item for a specific episode and rule
func (ad *AutoDownloader) findStoredItemForEpisode(episode int, ruleID uint, existingItems []*models.AutoDownloaderItem) *models.AutoDownloaderItem {
	for _, item := range existingItems {
		if item.Episode == episode && item.RuleID == ruleID {
			return item
		}
	}
	return nil
}

// handleDelayedItem processes a delayed item (upgrade check, threshold check, time check)
// Returns true if the item was downloaded
func (ad *AutoDownloader) handleDelayedItem(
	isSimulation bool,
	storedItem *models.AutoDownloaderItem,
	bestCandidate *Candidate,
	rule *anime.AutoDownloaderRule,
	episode int,
	settings delaySettings,
) bool {
	// 1. Upgrade if this torrent is better than the one delayed
	if bestCandidate.Score > storedItem.Score {
		// Serialize the updated torrent data
		torrentData, err := json.Marshal(bestCandidate.Torrent)
		if err != nil {
			ad.logger.Error().Err(err).Msg("autodownloader: Failed to serialize upgraded torrent")
		} else {
			storedItem.TorrentData = torrentData
		}

		storedItem.Link = bestCandidate.Torrent.Link
		storedItem.Hash = bestCandidate.Torrent.InfoHash
		storedItem.TorrentName = bestCandidate.Torrent.Name
		storedItem.Score = bestCandidate.Score
		// Do NOT reset DelayUntil, keep the original timer
		_ = ad.database.UpdateAutoDownloaderItem(storedItem.ID, storedItem)

		ad.logger.Debug().Str("title", rule.ComparisonTitle).Int("episode", episode).Msg("autodownloader: Upgraded delayed item to better match")
	}

	// 2. Does the torrent now exceed the SkipDelayScore?
	if storedItem.Score >= settings.skipDelayScore {
		ad.logger.Debug().Str("title", rule.ComparisonTitle).Int("episode", episode).Msg("autodownloader: Skip delay threshold met")
		return ad.downloadTorrent(isSimulation, bestCandidate.Torrent, rule, episode, bestCandidate.Score, storedItem)
	}

	// 3. Has the delay passed?
	if time.Now().After(storedItem.DelayUntil) {
		ad.logger.Debug().Str("title", rule.ComparisonTitle).Int("episode", episode).Msg("autodownloader: Delay timer expired, downloading")
		return ad.downloadTorrent(isSimulation, bestCandidate.Torrent, rule, episode, bestCandidate.Score, storedItem)
	}

	return false
}

// handleNewEpisode processes a new episode (threshold check, delay queue, immediate download)
// Returns true if the item was downloaded or queued
func (ad *AutoDownloader) handleNewEpisode(
	isSimulation bool,
	bestCandidate *Candidate,
	rule *anime.AutoDownloaderRule,
	episode int,
	settings delaySettings,
) bool {
	// 1. Delay the torrent
	if settings.hasDelay && bestCandidate.Score < settings.skipDelayScore {
		ad.logger.Debug().Int("episode", episode).Int("minutes", settings.delayMinutes).Msg("autodownloader: Queueing item for delay")
		ad.queueTorrentForDelay(isSimulation, rule, episode, bestCandidate, settings.delayMinutes)
		return false // not downloaded
	}

	// 2. Download or queue
	return ad.downloadTorrent(isSimulation, bestCandidate.Torrent, rule, episode, bestCandidate.Score, nil)
}

// processEpisodeCandidate processes a single episode's candidates
// Returns true if an item was downloaded
func (ad *AutoDownloader) processEpisodeCandidate(
	isSimulation bool,
	episode int,
	candidates []*Candidate,
	rule *anime.AutoDownloaderRule,
	existingItems []*models.AutoDownloaderItem,
	settings delaySettings,
) bool {
	if len(candidates) == 0 {
		return false
	}

	// 1. Identify best candidate
	bestCandidate := ad.selectBestCandidate(candidates)
	if bestCandidate == nil {
		return false
	}

	ad.logger.Debug().
		Str("releaseGroup", bestCandidate.Torrent.ParsedData.ReleaseGroup).
		Str("resolution", bestCandidate.Torrent.ParsedData.VideoResolution).
		Int("seeders", bestCandidate.Torrent.Seeders).
		Int("score", bestCandidate.Score).
		Str("rule", rule.ComparisonTitle).
		Int("episode", episode).
		Msg("autodownloader: Found best torrent")

	// 2. Check existing state
	storedItem := ad.findStoredItemForEpisode(episode, rule.DbID, existingItems)

	// 3. Decision

	// CASE A: Item already confirmed (not delayed)
	if storedItem != nil && !storedItem.IsDelayed {
		return false
	}

	// CASE B: Item is currently delayed
	if storedItem != nil && storedItem.IsDelayed {
		return ad.handleDelayedItem(isSimulation, storedItem, bestCandidate, rule, episode, settings)
	}

	// CASE C: This is a new episode
	if storedItem == nil {
		return ad.handleNewEpisode(isSimulation, bestCandidate, rule, episode, settings)
	}

	return false
}

// selectAndDownloadBestCandidates selects the best candidate for each episode and downloads it
// Returns the number of successfully downloaded episodes
func (ad *AutoDownloader) selectAndDownloadBestCandidates(isSimulation bool, groupedCandidates map[uint]map[int][]*Candidate, rules []*anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) int {
	downloaded := 0
	mu := sync.Mutex{}

	for ruleID, episodes := range groupedCandidates {
		rule, found := lo.Find(rules, func(r *anime.AutoDownloaderRule) bool {
			return r.DbID == ruleID
		})
		if !found {
			continue
		}

		// Resolve profiles and delay settings for this rule
		ruleProfiles := ad.getRuleProfiles(rule, profiles)
		settings := ad.getDelaySettings(rule, ruleProfiles)

		// Get all existing items for this media to check state
		existingItems, _ := ad.database.GetAutoDownloaderItemByMediaId(rule.MediaId)

		for episode, candidates := range episodes {
			if ad.processEpisodeCandidate(isSimulation, episode, candidates, rule, existingItems, settings) {
				mu.Lock()
				downloaded++
				mu.Unlock()
			}
		}
	}

	return downloaded
}

// downloadDelayedItems checks all delayed items and downloads those whose delay has expired
// Returns the number of successfully downloaded items
func (ad *AutoDownloader) downloadDelayedItems(isSimulation bool) int {
	// Get all delayed items across all media
	allItems, err := ad.database.GetDelayedAutoDownloaderItems()
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to get delayed items")
		return 0
	}

	// Filter to only delayed items with expired delay
	now := time.Now()
	expiredItems := make([]*models.AutoDownloaderItem, 0)
	for _, item := range allItems {
		if item.IsDelayed && now.After(item.DelayUntil) {
			expiredItems = append(expiredItems, item)
		}
	}

	if len(expiredItems) == 0 {
		return 0
	}

	ad.logger.Debug().Int("count", len(expiredItems)).Msg("autodownloader: Processing expired delayed items")

	// Get all rules to find the rule for each item
	rules, err := db_bridge.GetAutoDownloaderRules(ad.database)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to get rules for delayed items")
		return 0
	}

	downloaded := 0

	for _, item := range expiredItems {
		// Find the rule for this item
		rule, found := lo.Find(rules, func(r *anime.AutoDownloaderRule) bool {
			return r.DbID == item.RuleID
		})
		if !found {
			ad.logger.Warn().Uint("ruleId", item.RuleID).Msg("autodownloader: Rule not found for delayed item, skipping")
			continue
		}

		// Check if rule is enabled
		if !rule.Enabled {
			ad.logger.Debug().Uint("ruleId", item.RuleID).Msg("autodownloader: Rule disabled for delayed item, skipping")
			continue
		}

		// Deserialize the stored torrent data
		var t NormalizedTorrent
		if len(item.TorrentData) > 0 {
			if err := json.Unmarshal(item.TorrentData, &t); err != nil {
				ad.logger.Error().Err(err).Str("name", item.TorrentName).Msg("autodownloader: Failed to deserialize torrent data, skipping")
				continue
			}
		} else {
			ad.logger.Warn().Uint("ruleId", item.RuleID).Str("name", item.TorrentName).Msg("autodownloader: No torrent data found for delayed item, skipping")
			continue
		}

		// Download the stored torrent
		if ad.downloadTorrent(isSimulation, &t, rule, item.Episode, item.Score, item) {
			downloaded++
		}
	}

	if downloaded > 0 {
		ad.logger.Info().Int("count", downloaded).Msg("autodownloader: Downloaded expired delayed items")
	}

	return downloaded
}

// queueTorrentForDelay inserts an item with IsDelayed=true
func (ad *AutoDownloader) queueTorrentForDelay(isSimulation bool, rule *anime.AutoDownloaderRule, episode int, candidate *Candidate, delayMinutes int) {
	if isSimulation {
		// Store in memory for simulation mode
		ad.simulationResults = append(ad.simulationResults, &SimulationResult{
			RuleID:      rule.DbID,
			MediaID:     rule.MediaId,
			Episode:     episode,
			Link:        candidate.Torrent.Link,
			Hash:        candidate.Torrent.InfoHash,
			TorrentName: candidate.Torrent.Name,
			Score:       candidate.Score,
			ExtensionID: candidate.Torrent.ExtensionID,
			IsDelayed:   true,
		})
		return
	}

	// Serialize the torrent data
	torrentData, err := json.Marshal(candidate.Torrent)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to serialize torrent for delay")
		return
	}

	item := &models.AutoDownloaderItem{
		RuleID:      rule.DbID,
		MediaID:     rule.MediaId,
		Episode:     episode,
		Link:        candidate.Torrent.Link,
		Hash:        candidate.Torrent.InfoHash,
		TorrentName: candidate.Torrent.Name,
		Downloaded:  false,
		IsDelayed:   true,
		DelayUntil:  time.Now().Add(time.Duration(delayMinutes) * time.Minute),
		Score:       candidate.Score,
		TorrentData: torrentData,
	}
	_ = ad.database.InsertAutoDownloaderItem(item)

	ad.wsEventManager.SendEvent(events.AutoDownloaderItemAdded, candidate.Torrent.Name)
}

// selectBestCandidate selects the best candidate from a list based on score and seeders
func (ad *AutoDownloader) selectBestCandidate(candidates []*Candidate) *Candidate {
	// Sort candidates by score (desc) -> seeders (desc)
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Torrent.Seeders > candidates[j].Torrent.Seeders
	})

	return candidates[0]
}

// notifyDownloadResults sends a notification about the download results
func (ad *AutoDownloader) notifyDownloadResults(downloaded int) {
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
	profiles []*anime.AutoDownloaderProfile,
) (int, bool) {
	defer util.HandlePanicInModuleThen("autodownloader/torrentFollowsRule", func() {})

	if ok := ad.isProviderMatch(t, rule); !ok {
		return -1, false
	}

	// Inherit release groups from profiles if rule has none
	releaseGroups := ad.inheritReleaseGroupsFromProfiles(rule, profiles)

	if ok := ad.isReleaseGroupMatch(t.ParsedData.ReleaseGroup, releaseGroups); !ok {
		return -1, false
	}

	// If rule has no resolutions, inherit from profiles
	resolutions := ad.inheritResolutionsFromProfiles(rule, profiles)

	if ok := ad.isResolutionMatch(t.ParsedData.VideoResolution, resolutions); !ok {
		return -1, false
	}

	if ok := ad.isTitleMatch(t.ParsedData, t.Name, rule, listEntry); !ok {
		return -1, false
	}

	if ok := ad.isAdditionalTermsMatch(t.Name, rule); !ok {
		return -1, false
	}

	if ok := ad.isExcludedTermsMatch(t.Name, rule); !ok {
		return -1, false
	}

	if ok := ad.isConstraintsMatch(t, rule); !ok {
		return -1, false
	}

	// Check if the torrent matches all profiles (global & specific)
	for _, p := range profiles {
		if !ad.isProfileValidChecks(t, p) {
			return -1, false
		}
	}

	episode, ok := ad.isSeasonAndEpisodeMatch(t.ParsedData, rule, listEntry)
	if !ok {
		return -1, false
	}

	return episode, true
}

func (ad *AutoDownloader) inheritReleaseGroupsFromProfiles(rule *anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) []string {
	res := rule.ReleaseGroups
	if len(res) == 0 {
		for _, p := range profiles {
			res = append(res, p.ReleaseGroups...)
		}
	}
	res = lo.Uniq(res)
	return res
}

func (ad *AutoDownloader) inheritResolutionsFromProfiles(rule *anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) []string {
	res := rule.Resolutions
	if len(res) == 0 {
		for _, p := range profiles {
			res = append(res, p.Resolutions...)
		}
	}
	res = lo.Uniq(res)
	return res
}

func (ad *AutoDownloader) downloadTorrent(isSimulation bool, t *NormalizedTorrent, rule *anime.AutoDownloaderRule, episode int, score int, existingItem *models.AutoDownloaderItem) bool {
	defer util.HandlePanicInModuleThen("autodownloader/downloadTorrent", func() {})

	ad.logger.Debug().Str("name", t.Name).Msg("autodownloader: Downloading torrent")

	if isSimulation {
		// Store in memory for simulation mode
		ad.simulationResults = append(ad.simulationResults, &SimulationResult{
			RuleID:      rule.DbID,
			MediaID:     rule.MediaId,
			Episode:     episode,
			Link:        t.Link,
			Hash:        t.InfoHash,
			TorrentName: t.Name,
			Score:       score,
			ExtensionID: t.ExtensionID,
		})
		return true
	}

	// Double check that the episode hasn't been added while we have the lock
	// Skip this check if we're updating an existing delayed item
	var items []*models.AutoDownloaderItem
	if existingItem == nil {
		var err error
		items, err = ad.database.GetAutoDownloaderItemByMediaId(rule.MediaId)
		if err == nil {
			for _, item := range items {
				if item.Episode == episode && !item.IsDelayed {
					return false // Skip, episode was added by another goroutine
				}
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

	// Use the provider that found the torrent
	providerExtension, found := ad.torrentRepository.GetAnimeProviderExtension(t.ExtensionID)
	if !found {
		// This shouldn't happen
		ad.logger.Error().Str("extensionId", t.ExtensionID).Msg("autodownloader: Provider extension not found and no default provider available")
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

	// Get torrent magnet (use stored magnet if this is a delayed item)
	var magnet string
	var err error
	if existingItem != nil && existingItem.Magnet != "" {
		// Use stored magnet for delayed items
		magnet = existingItem.Magnet
	} else {
		// Fetch magnet from provider
		magnet, err = t.GetMagnet(providerExtension.GetProvider())
		if err != nil {
			// Try to construct from hash as fallback
			if t.InfoHash != "" {
				magnet = fmt.Sprintf("magnet:?xt=urn:btih:%s", t.InfoHash)
			} else {
				ad.logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to get magnet link for torrent")
				return false
			}
		}
	}

	downloaded := false

	downloadImmediately := ad.settings.DownloadAutomatically

downloadScope:
	if useDebrid {
		//
		// Debrid
		//

		if ad.debridClientRepository == nil {
			ad.logger.Error().Msg("autodownloader: debrid client not found")
			return false
		}

		if downloadImmediately {
			// Add the torrent to the debrid provider and queue it
			_, err := ad.debridClientRepository.AddAndQueueTorrent(debrid.AddTorrentOptions{
				MagnetLink:   magnet,
				SelectFileId: "all", // RD-only, select all files
			}, rule.Destination, rule.MediaId)
			if err != nil {
				ad.logger.Error().Err(err).Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to add torrent to debrid")
				downloadImmediately = false
				ad.logger.Warn().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Torrent will be queued.")
				goto downloadScope
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
		if downloadImmediately {

			if ad.torrentClientRepository == nil {
				ad.logger.Error().Msg("autodownloader: torrent client not found")
				downloadImmediately = false
				ad.logger.Warn().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Torrent will be queued.")
				goto downloadScope
			}

			//
			// Torrent client
			//
			started := ad.torrentClientRepository.Start() // Start torrent client if it's not running
			if !started {
				ad.logger.Error().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Failed to download torrent. torrent client is not running.")
				downloadImmediately = false
				ad.logger.Warn().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Torrent will be queued.")
				goto downloadScope
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
				downloadImmediately = false
				ad.logger.Warn().Str("link", t.Link).Str("name", t.Name).Msg("autodownloader: Torrent will be queued.")
				goto downloadScope
			}

			downloaded = true
		}
	}

	ad.wsEventManager.SendEvent(events.AutoDownloaderItemAdded, t.Name)

	// Serialize torrent data for storage
	torrentData, err := json.Marshal(t)
	if err != nil {
		ad.logger.Error().Err(err).Msg("autodownloader: Failed to serialize torrent data")
		torrentData = nil
	}

	// Update or insert the torrent in the database
	if existingItem != nil {
		// Update existing delayed item
		existingItem.Link = t.Link
		existingItem.Hash = t.InfoHash
		existingItem.Magnet = magnet
		existingItem.TorrentName = t.Name
		existingItem.Downloaded = downloaded
		existingItem.IsDelayed = false
		existingItem.Score = score
		existingItem.TorrentData = torrentData
		_ = ad.database.UpdateAutoDownloaderItem(existingItem.ID, existingItem)
		ad.logger.Info().Str("name", t.Name).Bool("downloaded", downloaded).Msg("autodownloader: Updated queued item")
	} else {
		// Insert new item
		item := &models.AutoDownloaderItem{
			RuleID:      rule.DbID,
			MediaID:     rule.MediaId,
			Episode:     episode,
			Link:        t.Link,
			Hash:        t.InfoHash,
			Magnet:      magnet,
			TorrentName: t.Name,
			Downloaded:  downloaded,
			IsDelayed:   false,
			Score:       score,
			TorrentData: torrentData,
		}
		_ = ad.database.InsertAutoDownloaderItem(item)
		ad.logger.Info().Str("name", t.Name).Bool("downloaded", downloaded).Msg("autodownloader: Added item to queue")
	}

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
		// If the torrent name doesn't contain any of the options
		if !foundOption {
			return false
		}
	}

	// If all options are found, return true
	return true
}
func (ad *AutoDownloader) isReleaseGroupMatch(releaseGroup string, releaseGroups []string) (ok bool) {
	defer util.HandlePanicInModuleThen("autodownloader/isReleaseGroupMatch", func() {
		ok = false
	})

	if len(releaseGroups) == 0 {
		return true
	}
	for _, rg := range releaseGroups {
		if strings.ToLower(rg) == strings.ToLower(releaseGroup) {
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
					new(fmt.Sprintf("%s Season %s", torrentParsedData.Title, torrentParsedData.SeasonNumber[0])),
					new(fmt.Sprintf("%s S%s", torrentParsedData.Title, torrentParsedData.SeasonNumber[0])),
					new(fmt.Sprintf("%s %s Season", torrentParsedData.Title, util.IntegerToOrdinal(util.StringToIntMust(torrentParsedData.SeasonNumber[0])))),
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
		// Return true if the media (has only one episode or is a movie)
		if listEntry.GetMedia().GetCurrentEpisodeCount() == 1 || *listEntry.GetMedia().GetFormat() == anilist.MediaFormatMovie {
			// Note: We used to check if items/locals exist here.
			// But now moved to the main loop to group first.
			return 1, true // Good to go
		}
		return -1, false
	}

	// +---------------------+
	// |   Episode number    |
	// +---------------------+

	hasAbsoluteEpisode := false

	// Add the absolute offset
	if rule.CustomEpisodeNumberAbsoluteOffset > 0 {
		episode = episode - rule.CustomEpisodeNumberAbsoluteOffset
		hasAbsoluteEpisode = true
		if episode < 1 {
			episode = 1
			hasAbsoluteEpisode = false
		}
	} else {
		// Handle absolute episode numbers from metadata
		if listEntry.GetMedia().GetCurrentEpisodeCount() != -1 && episode > listEntry.GetMedia().GetCurrentEpisodeCount() {
			// Fetch the Animap media in order to normalize the episode number
			ad.mu.Lock()
			animeMetadata, err := ad.metadataProviderRef.Get().GetAnimeMetadata(metadata.AnilistPlatform, listEntry.GetMedia().GetID())
			// If the media is found and the offset is greater than 0
			if err == nil && animeMetadata.GetOffset() > 0 {
				hasAbsoluteEpisode = true
				episode = episode - animeMetadata.GetOffset()
			}
			ad.mu.Unlock()
		}
	}

	// Note: We used to check if items/locals exist here.

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

// getProvidersForRules returns all providers that will be used
// Returns the default provider and any other provider used by rules or profiles
func (ad *AutoDownloader) getProvidersForRules(rules []*anime.AutoDownloaderRule, profiles []*anime.AutoDownloaderProfile) []extension.AnimeTorrentProviderExtension {
	providerIDs := make(map[string]struct{})

	// Add default provider
	defaultProv, found := ad.torrentRepository.GetAnimeProviderExtensionOrDefault(ad.settings.Provider)
	if found {
		providerIDs[defaultProv.GetID()] = struct{}{}
	}

	for _, rule := range rules {
		for _, p := range rule.Providers {
			providerIDs[p] = struct{}{}
		}
		if rule.ProfileID != nil {
			profile, found := lo.Find(profiles, func(p *anime.AutoDownloaderProfile) bool {
				return p.DbID == *rule.ProfileID
			})
			if found {
				for _, p := range profile.Providers {
					providerIDs[p] = struct{}{}
				}
			}
		}
	}

	for _, profile := range profiles {
		if profile.Global {
			for _, p := range profile.Providers {
				providerIDs[p] = struct{}{}
			}
		}
	}

	ret := make([]extension.AnimeTorrentProviderExtension, 0)
	for id := range providerIDs {
		ext, found := ad.torrentRepository.GetAnimeProviderExtension(id)
		if found {
			ret = append(ret, ext)
		}
	}

	return ret
}

func (ad *AutoDownloader) isProviderMatch(t *NormalizedTorrent, rule *anime.AutoDownloaderRule) bool {
	if len(rule.Providers) == 0 {
		return true
	}
	return lo.Contains(rule.Providers, t.ExtensionID)
}
