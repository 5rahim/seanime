package handlers

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/tvdb"
	"seanime/internal/continuity"
	"seanime/internal/database/models"
	"seanime/internal/debrid/client"
	"seanime/internal/debrid/debrid"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/extension_playground"
	"seanime/internal/extension_repo"
	"seanime/internal/hook_resolver"
	"seanime/internal/library/anime"
	"seanime/internal/library/summary"
	"seanime/internal/manga"
	"seanime/internal/manga/downloader"
	"seanime/internal/mediastream"
	"seanime/internal/onlinestream"
	"seanime/internal/report"
	"seanime/internal/sync"
	"seanime/internal/torrent_clients/torrent_client"
	"seanime/internal/torrents/torrent"
	"seanime/internal/torrentstream"
	"seanime/internal/updater"
)

// HandleGetAnimeCollectionRequestedEvent is triggered when GetAnimeCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnimeCollectionRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleGetAnimeCollectionEvent is triggered after processing GetAnimeCollection.
type HandleGetAnimeCollectionEvent struct {
	hook_resolver.Event
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleGetRawAnimeCollectionRequestedEvent is triggered when GetRawAnimeCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetRawAnimeCollectionRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleGetRawAnimeCollectionEvent is triggered after processing GetRawAnimeCollection.
type HandleGetRawAnimeCollectionEvent struct {
	hook_resolver.Event
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleEditAnilistListEntryRequestedEvent is triggered when EditAnilistListEntry is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleEditAnilistListEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId   int                      `json:"mediaId"`
	Status    *anilist.MediaListStatus `json:"status"`
	Score     int                      `json:"score"`
	Progress  int                      `json:"progress"`
	StartDate *anilist.FuzzyDateInput  `json:"startedAt"`
	EndDate   *anilist.FuzzyDateInput  `json:"completedAt"`
	Type      string                   `json:"type"`
}

// HandleGetAnilistAnimeDetailsRequestedEvent is triggered when GetAnilistAnimeDetails is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnilistAnimeDetailsRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.AnimeDetailsById_Media `json:"data"`
}

// HandleGetAnilistAnimeDetailsEvent is triggered after processing GetAnilistAnimeDetails.
type HandleGetAnilistAnimeDetailsEvent struct {
	hook_resolver.Event
	Data *anilist.AnimeDetailsById_Media `json:"data"`
}

// HandleGetAnilistStudioDetailsRequestedEvent is triggered when GetAnilistStudioDetails is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnilistStudioDetailsRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.StudioDetails `json:"data"`
}

// HandleGetAnilistStudioDetailsEvent is triggered after processing GetAnilistStudioDetails.
type HandleGetAnilistStudioDetailsEvent struct {
	hook_resolver.Event
	Data *anilist.StudioDetails `json:"data"`
}

// HandleDeleteAnilistListEntryRequestedEvent is triggered when DeleteAnilistListEntry is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteAnilistListEntryRequestedEvent struct {
	hook_resolver.Event
	MediaId int    `json:"mediaId"`
	Type    string `json:"type"`
}

// HandleAnilistListAnimeRequestedEvent is triggered when AnilistListAnime is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnilistListAnimeRequestedEvent struct {
	hook_resolver.Event
	Page                int                    `json:"page"`
	Search              string                 `json:"search"`
	PerPage             int                    `json:"perPage"`
	Sort                *[]anilist.MediaSort   `json:"sort"`
	Status              *[]anilist.MediaStatus `json:"status"`
	Genres              *[]string              `json:"genres"`
	AverageScoreGreater int                    `json:"averageScore_greater"`
	Season              *anilist.MediaSeason   `json:"season"`
	SeasonYear          int                    `json:"seasonYear"`
	Format              *anilist.MediaFormat   `json:"format"`
	IsAdult             bool                   `json:"isAdult"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.ListAnime `json:"data"`
}

// HandleAnilistListAnimeEvent is triggered after processing AnilistListAnime.
type HandleAnilistListAnimeEvent struct {
	hook_resolver.Event
	Data *anilist.ListAnime `json:"data"`
}

// HandleAnilistListRecentAiringAnimeRequestedEvent is triggered when AnilistListRecentAiringAnime is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnilistListRecentAiringAnimeRequestedEvent struct {
	hook_resolver.Event
	Page            int                   `json:"page"`
	Search          string                `json:"search"`
	PerPage         int                   `json:"perPage"`
	AiringAtGreater int                   `json:"airingAt_greater"`
	AiringAtLesser  int                   `json:"airingAt_lesser"`
	NotYetAired     bool                  `json:"notYetAired"`
	Sort            *[]anilist.AiringSort `json:"sort"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.ListRecentAnime `json:"data"`
}

// HandleAnilistListRecentAiringAnimeEvent is triggered after processing AnilistListRecentAiringAnime.
type HandleAnilistListRecentAiringAnimeEvent struct {
	hook_resolver.Event
	Data *anilist.ListRecentAnime `json:"data"`
}

// HandleAnilistListMissedSequelsRequestedEvent is triggered when AnilistListMissedSequels is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnilistListMissedSequelsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.BaseAnime `json:"data"`
}

// HandleAnilistListMissedSequelsEvent is triggered after processing AnilistListMissedSequels.
type HandleAnilistListMissedSequelsEvent struct {
	hook_resolver.Event
	Data *anilist.BaseAnime `json:"data"`
}

// HandleGetAniListStatsRequestedEvent is triggered when GetAniListStats is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAniListStatsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.Stats `json:"data"`
}

// HandleGetAniListStatsEvent is triggered after processing GetAniListStats.
type HandleGetAniListStatsEvent struct {
	hook_resolver.Event
	Data *anilist.Stats `json:"data"`
}

// HandleGetLibraryCollectionRequestedEvent is triggered when GetLibraryCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetLibraryCollectionRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LibraryCollection `json:"data"`
}

// HandleGetLibraryCollectionEvent is triggered after processing GetLibraryCollection.
type HandleGetLibraryCollectionEvent struct {
	hook_resolver.Event
	Data *anime.LibraryCollection `json:"data"`
}

// HandleAddUnknownMediaRequestedEvent is triggered when AddUnknownMedia is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAddUnknownMediaRequestedEvent struct {
	hook_resolver.Event
	MediaIds *[]int `json:"mediaIds"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleAddUnknownMediaEvent is triggered after processing AddUnknownMedia.
type HandleAddUnknownMediaEvent struct {
	hook_resolver.Event
	Data *anilist.AnimeCollection `json:"data"`
}

// HandleGetAnimeEntryRequestedEvent is triggered when GetAnimeEntry is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnimeEntryRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.Entry `json:"data"`
}

// HandleGetAnimeEntryEvent is triggered after processing GetAnimeEntry.
type HandleGetAnimeEntryEvent struct {
	hook_resolver.Event
	Data *anime.Entry `json:"data"`
}

// HandleAnimeEntryBulkActionRequestedEvent is triggered when AnimeEntryBulkAction is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnimeEntryBulkActionRequestedEvent struct {
	hook_resolver.Event
	MediaId int    `json:"mediaId"`
	Action  string `json:"action"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleAnimeEntryBulkActionEvent is triggered after processing AnimeEntryBulkAction.
type HandleAnimeEntryBulkActionEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleOpenAnimeEntryInExplorerRequestedEvent is triggered when OpenAnimeEntryInExplorer is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleOpenAnimeEntryInExplorerRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleFetchAnimeEntrySuggestionsRequestedEvent is triggered when FetchAnimeEntrySuggestions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleFetchAnimeEntrySuggestionsRequestedEvent struct {
	hook_resolver.Event
	Dir string `json:"dir"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.BaseAnime `json:"data"`
}

// HandleFetchAnimeEntrySuggestionsEvent is triggered after processing FetchAnimeEntrySuggestions.
type HandleFetchAnimeEntrySuggestionsEvent struct {
	hook_resolver.Event
	Data *anilist.BaseAnime `json:"data"`
}

// HandleAnimeEntryManualMatchRequestedEvent is triggered when AnimeEntryManualMatch is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnimeEntryManualMatchRequestedEvent struct {
	hook_resolver.Event
	Paths   *[]string `json:"paths"`
	MediaId int       `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleAnimeEntryManualMatchEvent is triggered after processing AnimeEntryManualMatch.
type HandleAnimeEntryManualMatchEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleGetMissingEpisodesRequestedEvent is triggered when GetMissingEpisodes is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMissingEpisodesRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.MissingEpisodes `json:"data"`
}

// HandleGetMissingEpisodesEvent is triggered after processing GetMissingEpisodes.
type HandleGetMissingEpisodesEvent struct {
	hook_resolver.Event
	Data *anime.MissingEpisodes `json:"data"`
}

// HandleGetAnimeEntrySilenceStatusRequestedEvent is triggered when GetAnimeEntrySilenceStatus is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnimeEntrySilenceStatusRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.SilencedMediaEntry `json:"data"`
}

// HandleGetAnimeEntrySilenceStatusEvent is triggered after processing GetAnimeEntrySilenceStatus.
type HandleGetAnimeEntrySilenceStatusEvent struct {
	hook_resolver.Event
	Data *models.SilencedMediaEntry `json:"data"`
}

// HandleToggleAnimeEntrySilenceStatusRequestedEvent is triggered when ToggleAnimeEntrySilenceStatus is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleToggleAnimeEntrySilenceStatusRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleUpdateAnimeEntryProgressRequestedEvent is triggered when UpdateAnimeEntryProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateAnimeEntryProgressRequestedEvent struct {
	hook_resolver.Event
	MediaId       int `json:"mediaId"`
	MalId         int `json:"malId"`
	EpisodeNumber int `json:"episodeNumber"`
	TotalEpisodes int `json:"totalEpisodes"`
}

// HandleUpdateAnimeEntryRepeatRequestedEvent is triggered when UpdateAnimeEntryRepeat is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateAnimeEntryRepeatRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	Repeat  int `json:"repeat"`
}

// HandleLoginRequestedEvent is triggered when Login is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleLoginRequestedEvent struct {
	hook_resolver.Event
	Token string `json:"token"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleLoginEvent is triggered after processing Login.
type HandleLoginEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleLogoutRequestedEvent is triggered when Logout is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleLogoutRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleLogoutEvent is triggered after processing Logout.
type HandleLogoutEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleRunAutoDownloaderRequestedEvent is triggered when RunAutoDownloader is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRunAutoDownloaderRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetAutoDownloaderRuleRequestedEvent is triggered when GetAutoDownloaderRule is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAutoDownloaderRuleRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleGetAutoDownloaderRuleEvent is triggered after processing GetAutoDownloaderRule.
type HandleGetAutoDownloaderRuleEvent struct {
	hook_resolver.Event
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleGetAutoDownloaderRulesByAnimeRequestedEvent is triggered when GetAutoDownloaderRulesByAnime is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAutoDownloaderRulesByAnimeRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleGetAutoDownloaderRulesByAnimeEvent is triggered after processing GetAutoDownloaderRulesByAnime.
type HandleGetAutoDownloaderRulesByAnimeEvent struct {
	hook_resolver.Event
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleGetAutoDownloaderRulesRequestedEvent is triggered when GetAutoDownloaderRules is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAutoDownloaderRulesRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleGetAutoDownloaderRulesEvent is triggered after processing GetAutoDownloaderRules.
type HandleGetAutoDownloaderRulesEvent struct {
	hook_resolver.Event
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleCreateAutoDownloaderRuleRequestedEvent is triggered when CreateAutoDownloaderRule is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleCreateAutoDownloaderRuleRequestedEvent struct {
	hook_resolver.Event
	Enabled             bool                                         `json:"enabled"`
	MediaId             int                                          `json:"mediaId"`
	ReleaseGroups       *[]string                                    `json:"releaseGroups"`
	Resolutions         *[]string                                    `json:"resolutions"`
	AdditionalTerms     *[]string                                    `json:"additionalTerms"`
	ComparisonTitle     string                                       `json:"comparisonTitle"`
	TitleComparisonType *anime.AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType"`
	EpisodeType         *anime.AutoDownloaderRuleEpisodeType         `json:"episodeType"`
	EpisodeNumbers      *[]int                                       `json:"episodeNumbers"`
	Destination         string                                       `json:"destination"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleCreateAutoDownloaderRuleEvent is triggered after processing CreateAutoDownloaderRule.
type HandleCreateAutoDownloaderRuleEvent struct {
	hook_resolver.Event
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleUpdateAutoDownloaderRuleRequestedEvent is triggered when UpdateAutoDownloaderRule is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateAutoDownloaderRuleRequestedEvent struct {
	hook_resolver.Event
	Rule *anime.AutoDownloaderRule `json:"rule"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleUpdateAutoDownloaderRuleEvent is triggered after processing UpdateAutoDownloaderRule.
type HandleUpdateAutoDownloaderRuleEvent struct {
	hook_resolver.Event
	Data *anime.AutoDownloaderRule `json:"data"`
}

// HandleDeleteAutoDownloaderRuleRequestedEvent is triggered when DeleteAutoDownloaderRule is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteAutoDownloaderRuleRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
}

// HandleGetAutoDownloaderItemsRequestedEvent is triggered when GetAutoDownloaderItems is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAutoDownloaderItemsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.AutoDownloaderItem `json:"data"`
}

// HandleGetAutoDownloaderItemsEvent is triggered after processing GetAutoDownloaderItems.
type HandleGetAutoDownloaderItemsEvent struct {
	hook_resolver.Event
	Data *models.AutoDownloaderItem `json:"data"`
}

// HandleDeleteAutoDownloaderItemRequestedEvent is triggered when DeleteAutoDownloaderItem is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteAutoDownloaderItemRequestedEvent struct {
	hook_resolver.Event
	Id int  `json:"id"`
	ID uint `json:"id"`
}

// HandleUpdateContinuityWatchHistoryItemRequestedEvent is triggered when UpdateContinuityWatchHistoryItem is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateContinuityWatchHistoryItemRequestedEvent struct {
	hook_resolver.Event
	Options *continuity.UpdateWatchHistoryItemOptions `json:"options"`
}

// HandleGetContinuityWatchHistoryItemRequestedEvent is triggered when GetContinuityWatchHistoryItem is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetContinuityWatchHistoryItemRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *continuity.WatchHistoryItemResponse `json:"data"`
}

// HandleGetContinuityWatchHistoryItemEvent is triggered after processing GetContinuityWatchHistoryItem.
type HandleGetContinuityWatchHistoryItemEvent struct {
	hook_resolver.Event
	Data *continuity.WatchHistoryItemResponse `json:"data"`
}

// HandleGetContinuityWatchHistoryRequestedEvent is triggered when GetContinuityWatchHistory is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetContinuityWatchHistoryRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *continuity.WatchHistory `json:"data"`
}

// HandleGetContinuityWatchHistoryEvent is triggered after processing GetContinuityWatchHistory.
type HandleGetContinuityWatchHistoryEvent struct {
	hook_resolver.Event
	Data *continuity.WatchHistory `json:"data"`
}

// HandleGetDebridSettingsRequestedEvent is triggered when GetDebridSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetDebridSettingsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.DebridSettings `json:"data"`
}

// HandleGetDebridSettingsEvent is triggered after processing GetDebridSettings.
type HandleGetDebridSettingsEvent struct {
	hook_resolver.Event
	Data *models.DebridSettings `json:"data"`
}

// HandleSaveDebridSettingsRequestedEvent is triggered when SaveDebridSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveDebridSettingsRequestedEvent struct {
	hook_resolver.Event
	Settings *models.DebridSettings `json:"settings"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.DebridSettings `json:"data"`
}

// HandleSaveDebridSettingsEvent is triggered after processing SaveDebridSettings.
type HandleSaveDebridSettingsEvent struct {
	hook_resolver.Event
	Data *models.DebridSettings `json:"data"`
}

// HandleDebridAddTorrentsRequestedEvent is triggered when DebridAddTorrents is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridAddTorrentsRequestedEvent struct {
	hook_resolver.Event
	Torrents    *[]hibiketorrent.AnimeTorrent `json:"torrents"`
	Media       *anilist.BaseAnime            `json:"media"`
	Destination string                        `json:"destination"`
}

// HandleDebridDownloadTorrentRequestedEvent is triggered when DebridDownloadTorrent is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridDownloadTorrentRequestedEvent struct {
	hook_resolver.Event
	TorrentItem *debrid.TorrentItem `json:"torrentItem"`
	Destination string              `json:"destination"`
}

// HandleDebridCancelDownloadRequestedEvent is triggered when DebridCancelDownload is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridCancelDownloadRequestedEvent struct {
	hook_resolver.Event
	ItemID string `json:"itemID"`
}

// HandleDebridDeleteTorrentRequestedEvent is triggered when DebridDeleteTorrent is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridDeleteTorrentRequestedEvent struct {
	hook_resolver.Event
	TorrentItem *debrid.TorrentItem `json:"torrentItem"`
}

// HandleDebridGetTorrentsRequestedEvent is triggered when DebridGetTorrents is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridGetTorrentsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *debrid.TorrentItem `json:"data"`
}

// HandleDebridGetTorrentsEvent is triggered after processing DebridGetTorrents.
type HandleDebridGetTorrentsEvent struct {
	hook_resolver.Event
	Data *debrid.TorrentItem `json:"data"`
}

// HandleDebridGetTorrentInfoRequestedEvent is triggered when DebridGetTorrentInfo is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridGetTorrentInfoRequestedEvent struct {
	hook_resolver.Event
	Torrent *hibiketorrent.AnimeTorrent `json:"torrent"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *debrid.TorrentInfo `json:"data"`
}

// HandleDebridGetTorrentInfoEvent is triggered after processing DebridGetTorrentInfo.
type HandleDebridGetTorrentInfoEvent struct {
	hook_resolver.Event
	Data *debrid.TorrentInfo `json:"data"`
}

// HandleDebridGetTorrentFilePreviewsRequestedEvent is triggered when DebridGetTorrentFilePreviews is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridGetTorrentFilePreviewsRequestedEvent struct {
	hook_resolver.Event
	Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
	EpisodeNumber int                         `json:"episodeNumber"`
	Media         *anilist.BaseAnime          `json:"media"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *debrid_client.FilePreview `json:"data"`
}

// HandleDebridGetTorrentFilePreviewsEvent is triggered after processing DebridGetTorrentFilePreviews.
type HandleDebridGetTorrentFilePreviewsEvent struct {
	hook_resolver.Event
	Data *debrid_client.FilePreview `json:"data"`
}

// HandleDebridStartStreamRequestedEvent is triggered when DebridStartStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridStartStreamRequestedEvent struct {
	hook_resolver.Event
	MediaId       int                               `json:"mediaId"`
	EpisodeNumber int                               `json:"episodeNumber"`
	AniDBEpisode  string                            `json:"aniDBEpisode"`
	AutoSelect    bool                              `json:"autoSelect"`
	Torrent       *hibiketorrent.AnimeTorrent       `json:"torrent"`
	FileId        string                            `json:"fileId"`
	FileIndex     int                               `json:"fileIndex"`
	PlaybackType  *debrid_client.StreamPlaybackType `json:"playbackType"`
	ClientId      string                            `json:"clientId"`
}

// HandleDebridCancelStreamRequestedEvent is triggered when DebridCancelStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDebridCancelStreamRequestedEvent struct {
	hook_resolver.Event
	Options *debrid_client.CancelStreamOptions `json:"options"`
}

// HandleDirectorySelectorRequestedEvent is triggered when DirectorySelector is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDirectorySelectorRequestedEvent struct {
	hook_resolver.Event
	Input string `json:"input"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *DirectorySelectorResponse `json:"data"`
}

// HandleDirectorySelectorEvent is triggered after processing DirectorySelector.
type HandleDirectorySelectorEvent struct {
	hook_resolver.Event
	Data *DirectorySelectorResponse `json:"data"`
}

// HandleSetDiscordMangaActivityRequestedEvent is triggered when SetDiscordMangaActivity is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSetDiscordMangaActivityRequestedEvent struct {
	hook_resolver.Event
	MediaId int    `json:"mediaId"`
	Title   string `json:"title"`
	Image   string `json:"image"`
	Chapter string `json:"chapter"`
}

// HandleSetDiscordLegacyAnimeActivityRequestedEvent is triggered when SetDiscordLegacyAnimeActivity is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSetDiscordLegacyAnimeActivityRequestedEvent struct {
	hook_resolver.Event
	MediaId       int    `json:"mediaId"`
	Title         string `json:"title"`
	Image         string `json:"image"`
	IsMovie       bool   `json:"isMovie"`
	EpisodeNumber int    `json:"episodeNumber"`
}

// HandleSetDiscordAnimeActivityWithProgressRequestedEvent is triggered when SetDiscordAnimeActivityWithProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSetDiscordAnimeActivityWithProgressRequestedEvent struct {
	hook_resolver.Event
	MediaId             int    `json:"mediaId"`
	Title               string `json:"title"`
	Image               string `json:"image"`
	IsMovie             bool   `json:"isMovie"`
	EpisodeNumber       int    `json:"episodeNumber"`
	Progress            int    `json:"progress"`
	Duration            int    `json:"duration"`
	TotalEpisodes       int    `json:"totalEpisodes"`
	CurrentEpisodeCount int    `json:"currentEpisodeCount"`
	EpisodeTitle        string `json:"episodeTitle"`
}

// HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent is triggered when UpdateDiscordAnimeActivityWithProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateDiscordAnimeActivityWithProgressRequestedEvent struct {
	hook_resolver.Event
	Progress int  `json:"progress"`
	Duration int  `json:"duration"`
	Paused   bool `json:"paused"`
}

// HandleCancelDiscordActivityRequestedEvent is triggered when CancelDiscordActivity is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleCancelDiscordActivityRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetDocsRequestedEvent is triggered when GetDocs is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetDocsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *ApiDocsGroup `json:"data"`
}

// HandleGetDocsEvent is triggered after processing GetDocs.
type HandleGetDocsEvent struct {
	hook_resolver.Event
	Data *ApiDocsGroup `json:"data"`
}

// HandleDownloadTorrentFileRequestedEvent is triggered when DownloadTorrentFile is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDownloadTorrentFileRequestedEvent struct {
	hook_resolver.Event
	DownloadUrls *[]string          `json:"download_urls"`
	Destination  string             `json:"destination"`
	Media        *anilist.BaseAnime `json:"media"`
}

// HandleDownloadReleaseRequestedEvent is triggered when DownloadRelease is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDownloadReleaseRequestedEvent struct {
	hook_resolver.Event
	DownloadUrl string `json:"download_url"`
	Destination string `json:"destination"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *DownloadReleaseResponse `json:"data"`
}

// HandleDownloadReleaseEvent is triggered after processing DownloadRelease.
type HandleDownloadReleaseEvent struct {
	hook_resolver.Event
	Data *DownloadReleaseResponse `json:"data"`
}

// HandleOpenInExplorerRequestedEvent is triggered when OpenInExplorer is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleOpenInExplorerRequestedEvent struct {
	hook_resolver.Event
	Path string `json:"path"`
}

// HandleFetchExternalExtensionDataRequestedEvent is triggered when FetchExternalExtensionData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleFetchExternalExtensionDataRequestedEvent struct {
	hook_resolver.Event
	ManifestURI string `json:"manifestUri"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension.Extension `json:"data"`
}

// HandleFetchExternalExtensionDataEvent is triggered after processing FetchExternalExtensionData.
type HandleFetchExternalExtensionDataEvent struct {
	hook_resolver.Event
	Data *extension.Extension `json:"data"`
}

// HandleInstallExternalExtensionRequestedEvent is triggered when InstallExternalExtension is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleInstallExternalExtensionRequestedEvent struct {
	hook_resolver.Event
	ManifestURI string `json:"manifestUri"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.ExtensionInstallResponse `json:"data"`
}

// HandleInstallExternalExtensionEvent is triggered after processing InstallExternalExtension.
type HandleInstallExternalExtensionEvent struct {
	hook_resolver.Event
	Data *extension_repo.ExtensionInstallResponse `json:"data"`
}

// HandleUninstallExternalExtensionRequestedEvent is triggered when UninstallExternalExtension is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUninstallExternalExtensionRequestedEvent struct {
	hook_resolver.Event
	ID string `json:"id"`
}

// HandleUpdateExtensionCodeRequestedEvent is triggered when UpdateExtensionCode is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateExtensionCodeRequestedEvent struct {
	hook_resolver.Event
	ID      string `json:"id"`
	Payload string `json:"payload"`
}

// HandleReloadExternalExtensionsRequestedEvent is triggered when ReloadExternalExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleReloadExternalExtensionsRequestedEvent struct {
	hook_resolver.Event
}

// HandleReloadExternalExtensionRequestedEvent is triggered when ReloadExternalExtension is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleReloadExternalExtensionRequestedEvent struct {
	hook_resolver.Event
	ID string `json:"id"`
}

// HandleListExtensionDataRequestedEvent is triggered when ListExtensionData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleListExtensionDataRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension.Extension `json:"data"`
}

// HandleListExtensionDataEvent is triggered after processing ListExtensionData.
type HandleListExtensionDataEvent struct {
	hook_resolver.Event
	Data *extension.Extension `json:"data"`
}

// HandleGetExtensionPayloadRequestedEvent is triggered when GetExtensionPayload is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetExtensionPayloadRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetExtensionPayloadEvent is triggered after processing GetExtensionPayload.
type HandleGetExtensionPayloadEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleListDevelopmentModeExtensionsRequestedEvent is triggered when ListDevelopmentModeExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleListDevelopmentModeExtensionsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension.Extension `json:"data"`
}

// HandleListDevelopmentModeExtensionsEvent is triggered after processing ListDevelopmentModeExtensions.
type HandleListDevelopmentModeExtensionsEvent struct {
	hook_resolver.Event
	Data *extension.Extension `json:"data"`
}

// HandleGetAllExtensionsRequestedEvent is triggered when GetAllExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAllExtensionsRequestedEvent struct {
	hook_resolver.Event
	WithUpdates bool `json:"withUpdates"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.AllExtensions `json:"data"`
}

// HandleGetAllExtensionsEvent is triggered after processing GetAllExtensions.
type HandleGetAllExtensionsEvent struct {
	hook_resolver.Event
	Data *extension_repo.AllExtensions `json:"data"`
}

// HandleGetExtensionUpdateDataRequestedEvent is triggered when GetExtensionUpdateData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetExtensionUpdateDataRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.UpdateData `json:"data"`
}

// HandleGetExtensionUpdateDataEvent is triggered after processing GetExtensionUpdateData.
type HandleGetExtensionUpdateDataEvent struct {
	hook_resolver.Event
	Data *extension_repo.UpdateData `json:"data"`
}

// HandleListMangaProviderExtensionsRequestedEvent is triggered when ListMangaProviderExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleListMangaProviderExtensionsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.MangaProviderExtensionItem `json:"data"`
}

// HandleListMangaProviderExtensionsEvent is triggered after processing ListMangaProviderExtensions.
type HandleListMangaProviderExtensionsEvent struct {
	hook_resolver.Event
	Data *extension_repo.MangaProviderExtensionItem `json:"data"`
}

// HandleListOnlinestreamProviderExtensionsRequestedEvent is triggered when ListOnlinestreamProviderExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleListOnlinestreamProviderExtensionsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.OnlinestreamProviderExtensionItem `json:"data"`
}

// HandleListOnlinestreamProviderExtensionsEvent is triggered after processing ListOnlinestreamProviderExtensions.
type HandleListOnlinestreamProviderExtensionsEvent struct {
	hook_resolver.Event
	Data *extension_repo.OnlinestreamProviderExtensionItem `json:"data"`
}

// HandleListAnimeTorrentProviderExtensionsRequestedEvent is triggered when ListAnimeTorrentProviderExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleListAnimeTorrentProviderExtensionsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.AnimeTorrentProviderExtensionItem `json:"data"`
}

// HandleListAnimeTorrentProviderExtensionsEvent is triggered after processing ListAnimeTorrentProviderExtensions.
type HandleListAnimeTorrentProviderExtensionsEvent struct {
	hook_resolver.Event
	Data *extension_repo.AnimeTorrentProviderExtensionItem `json:"data"`
}

// HandleGetPluginSettingsRequestedEvent is triggered when GetPluginSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetPluginSettingsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.StoredPluginSettingsData `json:"data"`
}

// HandleGetPluginSettingsEvent is triggered after processing GetPluginSettings.
type HandleGetPluginSettingsEvent struct {
	hook_resolver.Event
	Data *extension_repo.StoredPluginSettingsData `json:"data"`
}

// HandleSetPluginSettingsPinnedTraysRequestedEvent is triggered when SetPluginSettingsPinnedTrays is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSetPluginSettingsPinnedTraysRequestedEvent struct {
	hook_resolver.Event
	PinnedTrayPluginIds *[]string `json:"pinnedTrayPluginIds"`
}

// HandleGrantPluginPermissionsRequestedEvent is triggered when GrantPluginPermissions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGrantPluginPermissionsRequestedEvent struct {
	hook_resolver.Event
	ID string `json:"id"`
}

// HandleRunExtensionPlaygroundCodeRequestedEvent is triggered when RunExtensionPlaygroundCode is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRunExtensionPlaygroundCodeRequestedEvent struct {
	hook_resolver.Event
	Params *extension_playground.RunPlaygroundCodeParams `json:"params"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_playground.RunPlaygroundCodeResponse `json:"data"`
}

// HandleRunExtensionPlaygroundCodeEvent is triggered after processing RunExtensionPlaygroundCode.
type HandleRunExtensionPlaygroundCodeEvent struct {
	hook_resolver.Event
	Data *extension_playground.RunPlaygroundCodeResponse `json:"data"`
}

// HandleGetExtensionUserConfigRequestedEvent is triggered when GetExtensionUserConfig is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetExtensionUserConfigRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension_repo.ExtensionUserConfig `json:"data"`
}

// HandleGetExtensionUserConfigEvent is triggered after processing GetExtensionUserConfig.
type HandleGetExtensionUserConfigEvent struct {
	hook_resolver.Event
	Data *extension_repo.ExtensionUserConfig `json:"data"`
}

// HandleSaveExtensionUserConfigRequestedEvent is triggered when SaveExtensionUserConfig is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveExtensionUserConfigRequestedEvent struct {
	hook_resolver.Event
	ID      string             `json:"id"`
	Version int                `json:"version"`
	Values  *map[string]string `json:"values"`
}

// HandleGetMarketplaceExtensionsRequestedEvent is triggered when GetMarketplaceExtensions is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMarketplaceExtensionsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *extension.Extension `json:"data"`
}

// HandleGetMarketplaceExtensionsEvent is triggered after processing GetMarketplaceExtensions.
type HandleGetMarketplaceExtensionsEvent struct {
	hook_resolver.Event
	Data *extension.Extension `json:"data"`
}

// HandleGetFileCacheTotalSizeRequestedEvent is triggered when GetFileCacheTotalSize is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetFileCacheTotalSizeRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetFileCacheTotalSizeEvent is triggered after processing GetFileCacheTotalSize.
type HandleGetFileCacheTotalSizeEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleRemoveFileCacheBucketRequestedEvent is triggered when RemoveFileCacheBucket is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRemoveFileCacheBucketRequestedEvent struct {
	hook_resolver.Event
	Bucket string `json:"bucket"`
}

// HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent is triggered when GetFileCacheMediastreamVideoFilesTotalSize is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetFileCacheMediastreamVideoFilesTotalSizeRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent is triggered after processing GetFileCacheMediastreamVideoFilesTotalSize.
type HandleGetFileCacheMediastreamVideoFilesTotalSizeEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleClearFileCacheMediastreamVideoFilesRequestedEvent is triggered when ClearFileCacheMediastreamVideoFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleClearFileCacheMediastreamVideoFilesRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetLocalFilesRequestedEvent is triggered when GetLocalFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetLocalFilesRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleGetLocalFilesEvent is triggered after processing GetLocalFiles.
type HandleGetLocalFilesEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleImportLocalFilesRequestedEvent is triggered when ImportLocalFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleImportLocalFilesRequestedEvent struct {
	hook_resolver.Event
	DataFilePath string `json:"dataFilePath"`
}

// HandleLocalFileBulkActionRequestedEvent is triggered when LocalFileBulkAction is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleLocalFileBulkActionRequestedEvent struct {
	hook_resolver.Event
	Action string `json:"action"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleLocalFileBulkActionEvent is triggered after processing LocalFileBulkAction.
type HandleLocalFileBulkActionEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleUpdateLocalFileDataRequestedEvent is triggered when UpdateLocalFileData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateLocalFileDataRequestedEvent struct {
	hook_resolver.Event
	Path     string                   `json:"path"`
	Metadata *anime.LocalFileMetadata `json:"metadata"`
	Locked   bool                     `json:"locked"`
	Ignored  bool                     `json:"ignored"`
	MediaId  int                      `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleUpdateLocalFileDataEvent is triggered after processing UpdateLocalFileData.
type HandleUpdateLocalFileDataEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleUpdateLocalFilesRequestedEvent is triggered when UpdateLocalFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateLocalFilesRequestedEvent struct {
	hook_resolver.Event
	Paths   *[]string `json:"paths"`
	Action  string    `json:"action"`
	MediaId int       `json:"mediaId"`
}

// HandleDeleteLocalFilesRequestedEvent is triggered when DeleteLocalFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteLocalFilesRequestedEvent struct {
	hook_resolver.Event
	Paths *[]string `json:"paths"`
}

// HandleRemoveEmptyDirectoriesRequestedEvent is triggered when RemoveEmptyDirectories is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRemoveEmptyDirectoriesRequestedEvent struct {
	hook_resolver.Event
}

// HandleMALAuthRequestedEvent is triggered when MALAuth is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleMALAuthRequestedEvent struct {
	hook_resolver.Event
	Code         string `json:"code"`
	State        string `json:"state"`
	CodeVerifier string `json:"code_verifier"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *MalAuthResponse `json:"data"`
}

// HandleMALAuthEvent is triggered after processing MALAuth.
type HandleMALAuthEvent struct {
	hook_resolver.Event
	Data *MalAuthResponse `json:"data"`
}

// HandleEditMALListEntryProgressRequestedEvent is triggered when EditMALListEntryProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleEditMALListEntryProgressRequestedEvent struct {
	hook_resolver.Event
	MediaId  int `json:"mediaId"`
	Progress int `json:"progress"`
}

// HandleMALLogoutRequestedEvent is triggered when MALLogout is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleMALLogoutRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetAnilistMangaCollectionRequestedEvent is triggered when GetAnilistMangaCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetAnilistMangaCollectionRequestedEvent struct {
	hook_resolver.Event
	BypassCache bool `json:"bypassCache"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.MangaCollection `json:"data"`
}

// HandleGetAnilistMangaCollectionEvent is triggered after processing GetAnilistMangaCollection.
type HandleGetAnilistMangaCollectionEvent struct {
	hook_resolver.Event
	Data *anilist.MangaCollection `json:"data"`
}

// HandleGetRawAnilistMangaCollectionRequestedEvent is triggered when GetRawAnilistMangaCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetRawAnilistMangaCollectionRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.MangaCollection `json:"data"`
}

// HandleGetRawAnilistMangaCollectionEvent is triggered after processing GetRawAnilistMangaCollection.
type HandleGetRawAnilistMangaCollectionEvent struct {
	hook_resolver.Event
	Data *anilist.MangaCollection `json:"data"`
}

// HandleGetMangaCollectionRequestedEvent is triggered when GetMangaCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaCollectionRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.Collection `json:"data"`
}

// HandleGetMangaCollectionEvent is triggered after processing GetMangaCollection.
type HandleGetMangaCollectionEvent struct {
	hook_resolver.Event
	Data *manga.Collection `json:"data"`
}

// HandleGetMangaEntryRequestedEvent is triggered when GetMangaEntry is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaEntryRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.Entry `json:"data"`
}

// HandleGetMangaEntryEvent is triggered after processing GetMangaEntry.
type HandleGetMangaEntryEvent struct {
	hook_resolver.Event
	Data *manga.Entry `json:"data"`
}

// HandleGetMangaEntryDetailsRequestedEvent is triggered when GetMangaEntryDetails is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaEntryDetailsRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.MangaDetailsById_Media `json:"data"`
}

// HandleGetMangaEntryDetailsEvent is triggered after processing GetMangaEntryDetails.
type HandleGetMangaEntryDetailsEvent struct {
	hook_resolver.Event
	Data *anilist.MangaDetailsById_Media `json:"data"`
}

// HandleGetMangaLatestChapterNumbersMapRequestedEvent is triggered when GetMangaLatestChapterNumbersMap is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaLatestChapterNumbersMapRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.MangaLatestChapterNumberItem `json:"data"`
}

// HandleGetMangaLatestChapterNumbersMapEvent is triggered after processing GetMangaLatestChapterNumbersMap.
type HandleGetMangaLatestChapterNumbersMapEvent struct {
	hook_resolver.Event
	Data *manga.MangaLatestChapterNumberItem `json:"data"`
}

// HandleRefetchMangaChapterContainersRequestedEvent is triggered when RefetchMangaChapterContainers is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRefetchMangaChapterContainersRequestedEvent struct {
	hook_resolver.Event
	SelectedProviderMap *map[int]string `json:"selectedProviderMap"`
}

// HandleEmptyMangaEntryCacheRequestedEvent is triggered when EmptyMangaEntryCache is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleEmptyMangaEntryCacheRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleGetMangaEntryChaptersRequestedEvent is triggered when GetMangaEntryChapters is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaEntryChaptersRequestedEvent struct {
	hook_resolver.Event
	MediaId  int    `json:"mediaId"`
	Provider string `json:"provider"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.ChapterContainer `json:"data"`
}

// HandleGetMangaEntryChaptersEvent is triggered after processing GetMangaEntryChapters.
type HandleGetMangaEntryChaptersEvent struct {
	hook_resolver.Event
	Data *manga.ChapterContainer `json:"data"`
}

// HandleGetMangaEntryPagesRequestedEvent is triggered when GetMangaEntryPages is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaEntryPagesRequestedEvent struct {
	hook_resolver.Event
	MediaId    int    `json:"mediaId"`
	Provider   string `json:"provider"`
	ChapterId  string `json:"chapterId"`
	DoublePage bool   `json:"doublePage"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.PageContainer `json:"data"`
}

// HandleGetMangaEntryPagesEvent is triggered after processing GetMangaEntryPages.
type HandleGetMangaEntryPagesEvent struct {
	hook_resolver.Event
	Data *manga.PageContainer `json:"data"`
}

// HandleGetMangaEntryDownloadedChaptersRequestedEvent is triggered when GetMangaEntryDownloadedChapters is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaEntryDownloadedChaptersRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.ChapterContainer `json:"data"`
}

// HandleGetMangaEntryDownloadedChaptersEvent is triggered after processing GetMangaEntryDownloadedChapters.
type HandleGetMangaEntryDownloadedChaptersEvent struct {
	hook_resolver.Event
	Data *manga.ChapterContainer `json:"data"`
}

// HandleAnilistListMangaRequestedEvent is triggered when AnilistListManga is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleAnilistListMangaRequestedEvent struct {
	hook_resolver.Event
	Page                int                    `json:"page"`
	Search              string                 `json:"search"`
	PerPage             int                    `json:"perPage"`
	Sort                *[]anilist.MediaSort   `json:"sort"`
	Status              *[]anilist.MediaStatus `json:"status"`
	Genres              *[]string              `json:"genres"`
	AverageScoreGreater int                    `json:"averageScore_greater"`
	Year                int                    `json:"year"`
	CountryOfOrigin     string                 `json:"countryOfOrigin"`
	IsAdult             bool                   `json:"isAdult"`
	Format              *anilist.MediaFormat   `json:"format"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anilist.ListManga `json:"data"`
}

// HandleAnilistListMangaEvent is triggered after processing AnilistListManga.
type HandleAnilistListMangaEvent struct {
	hook_resolver.Event
	Data *anilist.ListManga `json:"data"`
}

// HandleUpdateMangaProgressRequestedEvent is triggered when UpdateMangaProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateMangaProgressRequestedEvent struct {
	hook_resolver.Event
	MediaId       int `json:"mediaId"`
	MalId         int `json:"malId"`
	ChapterNumber int `json:"chapterNumber"`
	TotalChapters int `json:"totalChapters"`
}

// HandleMangaManualSearchRequestedEvent is triggered when MangaManualSearch is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleMangaManualSearchRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	Query    string `json:"query"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *hibikemanga.SearchResult `json:"data"`
}

// HandleMangaManualSearchEvent is triggered after processing MangaManualSearch.
type HandleMangaManualSearchEvent struct {
	hook_resolver.Event
	Data *hibikemanga.SearchResult `json:"data"`
}

// HandleMangaManualMappingRequestedEvent is triggered when MangaManualMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleMangaManualMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
	MangaId  string `json:"mangaId"`
}

// HandleGetMangaMappingRequestedEvent is triggered when GetMangaMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.MappingResponse `json:"data"`
}

// HandleGetMangaMappingEvent is triggered after processing GetMangaMapping.
type HandleGetMangaMappingEvent struct {
	hook_resolver.Event
	Data *manga.MappingResponse `json:"data"`
}

// HandleRemoveMangaMappingRequestedEvent is triggered when RemoveMangaMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRemoveMangaMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
}

// HandleDownloadMangaChaptersRequestedEvent is triggered when DownloadMangaChapters is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDownloadMangaChaptersRequestedEvent struct {
	hook_resolver.Event
	MediaId    int       `json:"mediaId"`
	Provider   string    `json:"provider"`
	ChapterIds *[]string `json:"chapterIds"`
	StartNow   bool      `json:"startNow"`
}

// HandleGetMangaDownloadDataRequestedEvent is triggered when GetMangaDownloadData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaDownloadDataRequestedEvent struct {
	hook_resolver.Event
	MediaId int  `json:"mediaId"`
	Cached  bool `json:"cached"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.MediaDownloadData `json:"data"`
}

// HandleGetMangaDownloadDataEvent is triggered after processing GetMangaDownloadData.
type HandleGetMangaDownloadDataEvent struct {
	hook_resolver.Event
	Data *manga.MediaDownloadData `json:"data"`
}

// HandleGetMangaDownloadQueueRequestedEvent is triggered when GetMangaDownloadQueue is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaDownloadQueueRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.ChapterDownloadQueueItem `json:"data"`
}

// HandleGetMangaDownloadQueueEvent is triggered after processing GetMangaDownloadQueue.
type HandleGetMangaDownloadQueueEvent struct {
	hook_resolver.Event
	Data *models.ChapterDownloadQueueItem `json:"data"`
}

// HandleStartMangaDownloadQueueRequestedEvent is triggered when StartMangaDownloadQueue is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleStartMangaDownloadQueueRequestedEvent struct {
	hook_resolver.Event
}

// HandleStopMangaDownloadQueueRequestedEvent is triggered when StopMangaDownloadQueue is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleStopMangaDownloadQueueRequestedEvent struct {
	hook_resolver.Event
}

// HandleClearAllChapterDownloadQueueRequestedEvent is triggered when ClearAllChapterDownloadQueue is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleClearAllChapterDownloadQueueRequestedEvent struct {
	hook_resolver.Event
}

// HandleResetErroredChapterDownloadQueueRequestedEvent is triggered when ResetErroredChapterDownloadQueue is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleResetErroredChapterDownloadQueueRequestedEvent struct {
	hook_resolver.Event
}

// HandleDeleteMangaDownloadedChaptersRequestedEvent is triggered when DeleteMangaDownloadedChapters is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteMangaDownloadedChaptersRequestedEvent struct {
	hook_resolver.Event
	DownloadIds *[]chapter_downloader.DownloadID `json:"downloadIds"`
}

// HandleGetMangaDownloadsListRequestedEvent is triggered when GetMangaDownloadsList is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMangaDownloadsListRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *manga.DownloadListItem `json:"data"`
}

// HandleGetMangaDownloadsListEvent is triggered after processing GetMangaDownloadsList.
type HandleGetMangaDownloadsListEvent struct {
	hook_resolver.Event
	Data *manga.DownloadListItem `json:"data"`
}

// HandleTestDumpRequestedEvent is triggered when TestDump is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTestDumpRequestedEvent struct {
	hook_resolver.Event
}

// HandleStartDefaultMediaPlayerRequestedEvent is triggered when StartDefaultMediaPlayer is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleStartDefaultMediaPlayerRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetMediastreamSettingsRequestedEvent is triggered when GetMediastreamSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetMediastreamSettingsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.MediastreamSettings `json:"data"`
}

// HandleGetMediastreamSettingsEvent is triggered after processing GetMediastreamSettings.
type HandleGetMediastreamSettingsEvent struct {
	hook_resolver.Event
	Data *models.MediastreamSettings `json:"data"`
}

// HandleSaveMediastreamSettingsRequestedEvent is triggered when SaveMediastreamSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveMediastreamSettingsRequestedEvent struct {
	hook_resolver.Event
	Settings *models.MediastreamSettings `json:"settings"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.MediastreamSettings `json:"data"`
}

// HandleSaveMediastreamSettingsEvent is triggered after processing SaveMediastreamSettings.
type HandleSaveMediastreamSettingsEvent struct {
	hook_resolver.Event
	Data *models.MediastreamSettings `json:"data"`
}

// HandleRequestMediastreamMediaContainerRequestedEvent is triggered when RequestMediastreamMediaContainer is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRequestMediastreamMediaContainerRequestedEvent struct {
	hook_resolver.Event
	Path             string                  `json:"path"`
	StreamType       *mediastream.StreamType `json:"streamType"`
	AudioStreamIndex int                     `json:"audioStreamIndex"`
	ClientId         string                  `json:"clientId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *mediastream.MediaContainer `json:"data"`
}

// HandleRequestMediastreamMediaContainerEvent is triggered after processing RequestMediastreamMediaContainer.
type HandleRequestMediastreamMediaContainerEvent struct {
	hook_resolver.Event
	Data *mediastream.MediaContainer `json:"data"`
}

// HandlePreloadMediastreamMediaContainerRequestedEvent is triggered when PreloadMediastreamMediaContainer is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePreloadMediastreamMediaContainerRequestedEvent struct {
	hook_resolver.Event
	Path             string                  `json:"path"`
	StreamType       *mediastream.StreamType `json:"streamType"`
	AudioStreamIndex int                     `json:"audioStreamIndex"`
}

// HandleMediastreamShutdownTranscodeStreamRequestedEvent is triggered when MediastreamShutdownTranscodeStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleMediastreamShutdownTranscodeStreamRequestedEvent struct {
	hook_resolver.Event
}

// HandlePopulateTVDBEpisodesRequestedEvent is triggered when PopulateTVDBEpisodes is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePopulateTVDBEpisodesRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *tvdb.Episode `json:"data"`
}

// HandlePopulateTVDBEpisodesEvent is triggered after processing PopulateTVDBEpisodes.
type HandlePopulateTVDBEpisodesEvent struct {
	hook_resolver.Event
	Data *tvdb.Episode `json:"data"`
}

// HandleEmptyTVDBEpisodesRequestedEvent is triggered when EmptyTVDBEpisodes is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleEmptyTVDBEpisodesRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandlePopulateFillerDataRequestedEvent is triggered when PopulateFillerData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePopulateFillerDataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleRemoveFillerDataRequestedEvent is triggered when RemoveFillerData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRemoveFillerDataRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleGetOnlineStreamEpisodeListRequestedEvent is triggered when GetOnlineStreamEpisodeList is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetOnlineStreamEpisodeListRequestedEvent struct {
	hook_resolver.Event
	MediaId  int    `json:"mediaId"`
	Dubbed   bool   `json:"dubbed"`
	Provider string `json:"provider"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *onlinestream.EpisodeListResponse `json:"data"`
}

// HandleGetOnlineStreamEpisodeListEvent is triggered after processing GetOnlineStreamEpisodeList.
type HandleGetOnlineStreamEpisodeListEvent struct {
	hook_resolver.Event
	Data *onlinestream.EpisodeListResponse `json:"data"`
}

// HandleGetOnlineStreamEpisodeSourceRequestedEvent is triggered when GetOnlineStreamEpisodeSource is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetOnlineStreamEpisodeSourceRequestedEvent struct {
	hook_resolver.Event
	EpisodeNumber int    `json:"episodeNumber"`
	MediaId       int    `json:"mediaId"`
	Provider      string `json:"provider"`
	Dubbed        bool   `json:"dubbed"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *onlinestream.EpisodeSource `json:"data"`
}

// HandleGetOnlineStreamEpisodeSourceEvent is triggered after processing GetOnlineStreamEpisodeSource.
type HandleGetOnlineStreamEpisodeSourceEvent struct {
	hook_resolver.Event
	Data *onlinestream.EpisodeSource `json:"data"`
}

// HandleOnlineStreamEmptyCacheRequestedEvent is triggered when OnlineStreamEmptyCache is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleOnlineStreamEmptyCacheRequestedEvent struct {
	hook_resolver.Event
	MediaId int `json:"mediaId"`
}

// HandleOnlinestreamManualSearchRequestedEvent is triggered when OnlinestreamManualSearch is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleOnlinestreamManualSearchRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	Query    string `json:"query"`
	Dubbed   bool   `json:"dubbed"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *hibikeonlinestream.SearchResult `json:"data"`
}

// HandleOnlinestreamManualSearchEvent is triggered after processing OnlinestreamManualSearch.
type HandleOnlinestreamManualSearchEvent struct {
	hook_resolver.Event
	Data *hibikeonlinestream.SearchResult `json:"data"`
}

// HandleOnlinestreamManualMappingRequestedEvent is triggered when OnlinestreamManualMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleOnlinestreamManualMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
	AnimeId  string `json:"animeId"`
}

// HandleGetOnlinestreamMappingRequestedEvent is triggered when GetOnlinestreamMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetOnlinestreamMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *onlinestream.MappingResponse `json:"data"`
}

// HandleGetOnlinestreamMappingEvent is triggered after processing GetOnlinestreamMapping.
type HandleGetOnlinestreamMappingEvent struct {
	hook_resolver.Event
	Data *onlinestream.MappingResponse `json:"data"`
}

// HandleRemoveOnlinestreamMappingRequestedEvent is triggered when RemoveOnlinestreamMapping is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleRemoveOnlinestreamMappingRequestedEvent struct {
	hook_resolver.Event
	Provider string `json:"provider"`
	MediaId  int    `json:"mediaId"`
}

// HandlePlaybackPlayVideoRequestedEvent is triggered when PlaybackPlayVideo is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackPlayVideoRequestedEvent struct {
	hook_resolver.Event
	Path string `json:"path"`
}

// HandlePlaybackPlayRandomVideoRequestedEvent is triggered when PlaybackPlayRandomVideo is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackPlayRandomVideoRequestedEvent struct {
	hook_resolver.Event
}

// HandlePlaybackSyncCurrentProgressRequestedEvent is triggered when PlaybackSyncCurrentProgress is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackSyncCurrentProgressRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data int `json:"data"`
}

// HandlePlaybackSyncCurrentProgressEvent is triggered after processing PlaybackSyncCurrentProgress.
type HandlePlaybackSyncCurrentProgressEvent struct {
	hook_resolver.Event
	Data int `json:"data"`
}

// HandlePlaybackPlayNextEpisodeRequestedEvent is triggered when PlaybackPlayNextEpisode is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackPlayNextEpisodeRequestedEvent struct {
	hook_resolver.Event
}

// HandlePlaybackGetNextEpisodeRequestedEvent is triggered when PlaybackGetNextEpisode is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackGetNextEpisodeRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandlePlaybackGetNextEpisodeEvent is triggered after processing PlaybackGetNextEpisode.
type HandlePlaybackGetNextEpisodeEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandlePlaybackAutoPlayNextEpisodeRequestedEvent is triggered when PlaybackAutoPlayNextEpisode is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackAutoPlayNextEpisodeRequestedEvent struct {
	hook_resolver.Event
}

// HandlePlaybackStartPlaylistRequestedEvent is triggered when PlaybackStartPlaylist is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackStartPlaylistRequestedEvent struct {
	hook_resolver.Event
	DbId uint `json:"dbId"`
}

// HandlePlaybackCancelCurrentPlaylistRequestedEvent is triggered when PlaybackCancelCurrentPlaylist is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackCancelCurrentPlaylistRequestedEvent struct {
	hook_resolver.Event
}

// HandlePlaybackPlaylistNextRequestedEvent is triggered when PlaybackPlaylistNext is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackPlaylistNextRequestedEvent struct {
	hook_resolver.Event
}

// HandlePlaybackStartManualTrackingRequestedEvent is triggered when PlaybackStartManualTracking is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackStartManualTrackingRequestedEvent struct {
	hook_resolver.Event
	MediaId       int    `json:"mediaId"`
	EpisodeNumber int    `json:"episodeNumber"`
	ClientId      string `json:"clientId"`
}

// HandlePlaybackCancelManualTrackingRequestedEvent is triggered when PlaybackCancelManualTracking is requested.
// Prevent default to skip the default behavior and return your own data.
type HandlePlaybackCancelManualTrackingRequestedEvent struct {
	hook_resolver.Event
}

// HandleCreatePlaylistRequestedEvent is triggered when CreatePlaylist is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleCreatePlaylistRequestedEvent struct {
	hook_resolver.Event
	Name  string    `json:"name"`
	Paths *[]string `json:"paths"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.Playlist `json:"data"`
}

// HandleCreatePlaylistEvent is triggered after processing CreatePlaylist.
type HandleCreatePlaylistEvent struct {
	hook_resolver.Event
	Data *anime.Playlist `json:"data"`
}

// HandleGetPlaylistsRequestedEvent is triggered when GetPlaylists is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetPlaylistsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.Playlist `json:"data"`
}

// HandleGetPlaylistsEvent is triggered after processing GetPlaylists.
type HandleGetPlaylistsEvent struct {
	hook_resolver.Event
	Data *anime.Playlist `json:"data"`
}

// HandleUpdatePlaylistRequestedEvent is triggered when UpdatePlaylist is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdatePlaylistRequestedEvent struct {
	hook_resolver.Event
	Id    int       `json:"id"`
	DbId  uint      `json:"dbId"`
	Name  string    `json:"name"`
	Paths *[]string `json:"paths"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.Playlist `json:"data"`
}

// HandleUpdatePlaylistEvent is triggered after processing UpdatePlaylist.
type HandleUpdatePlaylistEvent struct {
	hook_resolver.Event
	Data *anime.Playlist `json:"data"`
}

// HandleDeletePlaylistRequestedEvent is triggered when DeletePlaylist is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeletePlaylistRequestedEvent struct {
	hook_resolver.Event
	DbId uint `json:"dbId"`
}

// HandleGetPlaylistEpisodesRequestedEvent is triggered when GetPlaylistEpisodes is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetPlaylistEpisodesRequestedEvent struct {
	hook_resolver.Event
	Id       int `json:"id"`
	Progress int `json:"progress"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleGetPlaylistEpisodesEvent is triggered after processing GetPlaylistEpisodes.
type HandleGetPlaylistEpisodesEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleInstallLatestUpdateRequestedEvent is triggered when InstallLatestUpdate is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleInstallLatestUpdateRequestedEvent struct {
	hook_resolver.Event
	FallbackDestination string `json:"fallback_destination"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleInstallLatestUpdateEvent is triggered after processing InstallLatestUpdate.
type HandleInstallLatestUpdateEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleGetLatestUpdateRequestedEvent is triggered when GetLatestUpdate is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetLatestUpdateRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *updater.Update `json:"data"`
}

// HandleGetLatestUpdateEvent is triggered after processing GetLatestUpdate.
type HandleGetLatestUpdateEvent struct {
	hook_resolver.Event
	Data *updater.Update `json:"data"`
}

// HandleGetChangelogRequestedEvent is triggered when GetChangelog is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetChangelogRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetChangelogEvent is triggered after processing GetChangelog.
type HandleGetChangelogEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleSaveIssueReportRequestedEvent is triggered when SaveIssueReport is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveIssueReportRequestedEvent struct {
	hook_resolver.Event
	ClickLogs           *[]report.ClickLog      `json:"clickLogs"`
	NetworkLogs         *[]report.NetworkLog    `json:"networkLogs"`
	ReactQueryLogs      *[]report.ReactQueryLog `json:"reactQueryLogs"`
	ConsoleLogs         *[]report.ConsoleLog    `json:"consoleLogs"`
	IsAnimeLibraryIssue bool                    `json:"isAnimeLibraryIssue"`
}

// HandleDownloadIssueReportRequestedEvent is triggered when DownloadIssueReport is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDownloadIssueReportRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *report.IssueReport `json:"data"`
}

// HandleDownloadIssueReportEvent is triggered after processing DownloadIssueReport.
type HandleDownloadIssueReportEvent struct {
	hook_resolver.Event
	Data *report.IssueReport `json:"data"`
}

// HandleScanLocalFilesRequestedEvent is triggered when ScanLocalFiles is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleScanLocalFilesRequestedEvent struct {
	hook_resolver.Event
	Enhanced         bool `json:"enhanced"`
	SkipLockedFiles  bool `json:"skipLockedFiles"`
	SkipIgnoredFiles bool `json:"skipIgnoredFiles"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *anime.LocalFile `json:"data"`
}

// HandleScanLocalFilesEvent is triggered after processing ScanLocalFiles.
type HandleScanLocalFilesEvent struct {
	hook_resolver.Event
	Data *anime.LocalFile `json:"data"`
}

// HandleGetScanSummariesRequestedEvent is triggered when GetScanSummaries is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetScanSummariesRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *summary.ScanSummaryItem `json:"data"`
}

// HandleGetScanSummariesEvent is triggered after processing GetScanSummaries.
type HandleGetScanSummariesEvent struct {
	hook_resolver.Event
	Data *summary.ScanSummaryItem `json:"data"`
}

// HandleGetSettingsRequestedEvent is triggered when GetSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetSettingsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.Settings `json:"data"`
}

// HandleGetSettingsEvent is triggered after processing GetSettings.
type HandleGetSettingsEvent struct {
	hook_resolver.Event
	Data *models.Settings `json:"data"`
}

// HandleGettingStartedRequestedEvent is triggered when GettingStarted is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGettingStartedRequestedEvent struct {
	hook_resolver.Event
	Library                *models.LibrarySettings      `json:"library"`
	MediaPlayer            *models.MediaPlayerSettings  `json:"mediaPlayer"`
	Torrent                *models.TorrentSettings      `json:"torrent"`
	Anilist                *models.AnilistSettings      `json:"anilist"`
	Discord                *models.DiscordSettings      `json:"discord"`
	Manga                  *models.MangaSettings        `json:"manga"`
	Notifications          *models.NotificationSettings `json:"notifications"`
	EnableTranscode        bool                         `json:"enableTranscode"`
	EnableTorrentStreaming bool                         `json:"enableTorrentStreaming"`
	DebridProvider         string                       `json:"debridProvider"`
	DebridApiKey           string                       `json:"debridApiKey"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleGettingStartedEvent is triggered after processing GettingStarted.
type HandleGettingStartedEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleSaveSettingsRequestedEvent is triggered when SaveSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveSettingsRequestedEvent struct {
	hook_resolver.Event
	Library       *models.LibrarySettings      `json:"library"`
	MediaPlayer   *models.MediaPlayerSettings  `json:"mediaPlayer"`
	Torrent       *models.TorrentSettings      `json:"torrent"`
	Anilist       *models.AnilistSettings      `json:"anilist"`
	Discord       *models.DiscordSettings      `json:"discord"`
	Manga         *models.MangaSettings        `json:"manga"`
	Notifications *models.NotificationSettings `json:"notifications"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleSaveSettingsEvent is triggered after processing SaveSettings.
type HandleSaveSettingsEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleSaveAutoDownloaderSettingsRequestedEvent is triggered when SaveAutoDownloaderSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveAutoDownloaderSettingsRequestedEvent struct {
	hook_resolver.Event
	Interval              int  `json:"interval"`
	Enabled               bool `json:"enabled"`
	DownloadAutomatically bool `json:"downloadAutomatically"`
	EnableEnhancedQueries bool `json:"enableEnhancedQueries"`
	EnableSeasonCheck     bool `json:"enableSeasonCheck"`
	UseDebrid             bool `json:"useDebrid"`
}

// HandleGetStatusRequestedEvent is triggered when GetStatus is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetStatusRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *Status `json:"data"`
}

// HandleGetStatusEvent is triggered after processing GetStatus.
type HandleGetStatusEvent struct {
	hook_resolver.Event
	Data *Status `json:"data"`
}

// HandleGetLogFilenamesRequestedEvent is triggered when GetLogFilenames is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetLogFilenamesRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetLogFilenamesEvent is triggered after processing GetLogFilenames.
type HandleGetLogFilenamesEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleDeleteLogsRequestedEvent is triggered when DeleteLogs is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleDeleteLogsRequestedEvent struct {
	hook_resolver.Event
	Filenames *[]string `json:"filenames"`
}

// HandleGetLatestLogContentRequestedEvent is triggered when GetLatestLogContent is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetLatestLogContentRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleGetLatestLogContentEvent is triggered after processing GetLatestLogContent.
type HandleGetLatestLogContentEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleSyncGetTrackedMediaItemsRequestedEvent is triggered when SyncGetTrackedMediaItems is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncGetTrackedMediaItemsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *sync.TrackedMediaItem `json:"data"`
}

// HandleSyncGetTrackedMediaItemsEvent is triggered after processing SyncGetTrackedMediaItems.
type HandleSyncGetTrackedMediaItemsEvent struct {
	hook_resolver.Event
	Data *sync.TrackedMediaItem `json:"data"`
}

// HandleSyncAddMediaRequestedEvent is triggered when SyncAddMedia is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncAddMediaRequestedEvent struct {
	hook_resolver.Event
	Media *[]struct {
		MediaId int    `json:"mediaId"`
		Type    string `json:"type"`
	} `json:"media"`
}

// HandleSyncRemoveMediaRequestedEvent is triggered when SyncRemoveMedia is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncRemoveMediaRequestedEvent struct {
	hook_resolver.Event
	MediaId int    `json:"mediaId"`
	Type    string `json:"type"`
}

// HandleSyncGetIsMediaTrackedRequestedEvent is triggered when SyncGetIsMediaTracked is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncGetIsMediaTrackedRequestedEvent struct {
	hook_resolver.Event
	Id   int    `json:"id"`
	Type string `json:"type"`
}

// HandleSyncLocalDataRequestedEvent is triggered when SyncLocalData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncLocalDataRequestedEvent struct {
	hook_resolver.Event
}

// HandleSyncGetQueueStateRequestedEvent is triggered when SyncGetQueueState is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncGetQueueStateRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *sync.QueueState `json:"data"`
}

// HandleSyncGetQueueStateEvent is triggered after processing SyncGetQueueState.
type HandleSyncGetQueueStateEvent struct {
	hook_resolver.Event
	Data *sync.QueueState `json:"data"`
}

// HandleSyncAnilistDataRequestedEvent is triggered when SyncAnilistData is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncAnilistDataRequestedEvent struct {
	hook_resolver.Event
}

// HandleSyncSetHasLocalChangesRequestedEvent is triggered when SyncSetHasLocalChanges is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncSetHasLocalChangesRequestedEvent struct {
	hook_resolver.Event
	Updated bool `json:"updated"`
}

// HandleSyncGetHasLocalChangesRequestedEvent is triggered when SyncGetHasLocalChanges is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncGetHasLocalChangesRequestedEvent struct {
	hook_resolver.Event
}

// HandleSyncGetLocalStorageSizeRequestedEvent is triggered when SyncGetLocalStorageSize is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSyncGetLocalStorageSizeRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data string `json:"data"`
}

// HandleSyncGetLocalStorageSizeEvent is triggered after processing SyncGetLocalStorageSize.
type HandleSyncGetLocalStorageSizeEvent struct {
	hook_resolver.Event
	Data string `json:"data"`
}

// HandleGetThemeRequestedEvent is triggered when GetTheme is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetThemeRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.Theme `json:"data"`
}

// HandleGetThemeEvent is triggered after processing GetTheme.
type HandleGetThemeEvent struct {
	hook_resolver.Event
	Data *models.Theme `json:"data"`
}

// HandleUpdateThemeRequestedEvent is triggered when UpdateTheme is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleUpdateThemeRequestedEvent struct {
	hook_resolver.Event
	Theme *models.Theme `json:"theme"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.Theme `json:"data"`
}

// HandleUpdateThemeEvent is triggered after processing UpdateTheme.
type HandleUpdateThemeEvent struct {
	hook_resolver.Event
	Data *models.Theme `json:"data"`
}

// HandleGetActiveTorrentListRequestedEvent is triggered when GetActiveTorrentList is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetActiveTorrentListRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *torrent_client.Torrent `json:"data"`
}

// HandleGetActiveTorrentListEvent is triggered after processing GetActiveTorrentList.
type HandleGetActiveTorrentListEvent struct {
	hook_resolver.Event
	Data *torrent_client.Torrent `json:"data"`
}

// HandleTorrentClientActionRequestedEvent is triggered when TorrentClientAction is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentClientActionRequestedEvent struct {
	hook_resolver.Event
	Hash   string `json:"hash"`
	Action string `json:"action"`
	Dir    string `json:"dir"`
}

// HandleTorrentClientDownloadRequestedEvent is triggered when TorrentClientDownload is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentClientDownloadRequestedEvent struct {
	hook_resolver.Event
	Torrents    *[]hibiketorrent.AnimeTorrent `json:"torrents"`
	Destination string                        `json:"destination"`
	SmartSelect *struct {
		Enabled               bool  `json:"enabled"`
		MissingEpisodeNumbers []int `json:"missingEpisodeNumbers"`
	} `json:"smartSelect"`
	Media *anilist.BaseAnime `json:"media"`
}

// HandleTorrentClientAddMagnetFromRuleRequestedEvent is triggered when TorrentClientAddMagnetFromRule is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentClientAddMagnetFromRuleRequestedEvent struct {
	hook_resolver.Event
	MagnetUrl    string `json:"magnetUrl"`
	RuleId       uint   `json:"ruleId"`
	QueuedItemId uint   `json:"queuedItemId"`
}

// HandleSearchTorrentRequestedEvent is triggered when SearchTorrent is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSearchTorrentRequestedEvent struct {
	hook_resolver.Event
	Type           string             `json:"type"`
	Provider       string             `json:"provider"`
	Query          string             `json:"query"`
	EpisodeNumber  int                `json:"episodeNumber"`
	Batch          bool               `json:"batch"`
	Media          *anilist.BaseAnime `json:"media"`
	AbsoluteOffset int                `json:"absoluteOffset"`
	Resolution     string             `json:"resolution"`
	BestRelease    bool               `json:"bestRelease"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *torrent.SearchData `json:"data"`
}

// HandleSearchTorrentEvent is triggered after processing SearchTorrent.
type HandleSearchTorrentEvent struct {
	hook_resolver.Event
	Data *torrent.SearchData `json:"data"`
}

// HandleGetTorrentstreamEpisodeCollectionRequestedEvent is triggered when GetTorrentstreamEpisodeCollection is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetTorrentstreamEpisodeCollectionRequestedEvent struct {
	hook_resolver.Event
	Id int `json:"id"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *torrentstream.EpisodeCollection `json:"data"`
}

// HandleGetTorrentstreamEpisodeCollectionEvent is triggered after processing GetTorrentstreamEpisodeCollection.
type HandleGetTorrentstreamEpisodeCollectionEvent struct {
	hook_resolver.Event
	Data *torrentstream.EpisodeCollection `json:"data"`
}

// HandleGetTorrentstreamSettingsRequestedEvent is triggered when GetTorrentstreamSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetTorrentstreamSettingsRequestedEvent struct {
	hook_resolver.Event
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.TorrentstreamSettings `json:"data"`
}

// HandleGetTorrentstreamSettingsEvent is triggered after processing GetTorrentstreamSettings.
type HandleGetTorrentstreamSettingsEvent struct {
	hook_resolver.Event
	Data *models.TorrentstreamSettings `json:"data"`
}

// HandleSaveTorrentstreamSettingsRequestedEvent is triggered when SaveTorrentstreamSettings is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleSaveTorrentstreamSettingsRequestedEvent struct {
	hook_resolver.Event
	Settings *models.TorrentstreamSettings `json:"settings"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *models.TorrentstreamSettings `json:"data"`
}

// HandleSaveTorrentstreamSettingsEvent is triggered after processing SaveTorrentstreamSettings.
type HandleSaveTorrentstreamSettingsEvent struct {
	hook_resolver.Event
	Data *models.TorrentstreamSettings `json:"data"`
}

// HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent is triggered when GetTorrentstreamTorrentFilePreviews is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetTorrentstreamTorrentFilePreviewsRequestedEvent struct {
	hook_resolver.Event
	Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
	EpisodeNumber int                         `json:"episodeNumber"`
	Media         *anilist.BaseAnime          `json:"media"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *torrentstream.FilePreview `json:"data"`
}

// HandleGetTorrentstreamTorrentFilePreviewsEvent is triggered after processing GetTorrentstreamTorrentFilePreviews.
type HandleGetTorrentstreamTorrentFilePreviewsEvent struct {
	hook_resolver.Event
	Data *torrentstream.FilePreview `json:"data"`
}

// HandleTorrentstreamStartStreamRequestedEvent is triggered when TorrentstreamStartStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentstreamStartStreamRequestedEvent struct {
	hook_resolver.Event
	MediaId       int                         `json:"mediaId"`
	EpisodeNumber int                         `json:"episodeNumber"`
	AniDBEpisode  string                      `json:"aniDBEpisode"`
	AutoSelect    bool                        `json:"autoSelect"`
	Torrent       *hibiketorrent.AnimeTorrent `json:"torrent"`
	FileIndex     int                         `json:"fileIndex"`
	PlaybackType  *torrentstream.PlaybackType `json:"playbackType"`
	ClientId      string                      `json:"clientId"`
}

// HandleTorrentstreamStopStreamRequestedEvent is triggered when TorrentstreamStopStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentstreamStopStreamRequestedEvent struct {
	hook_resolver.Event
}

// HandleTorrentstreamDropTorrentRequestedEvent is triggered when TorrentstreamDropTorrent is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentstreamDropTorrentRequestedEvent struct {
	hook_resolver.Event
}

// HandleGetTorrentstreamBatchHistoryRequestedEvent is triggered when GetTorrentstreamBatchHistory is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleGetTorrentstreamBatchHistoryRequestedEvent struct {
	hook_resolver.Event
	MediaID int `json:"mediaId"`
	// Empty data object, will be used if the hook prevents the default behavior
	Data *torrentstream.BatchHistoryResponse `json:"data"`
}

// HandleGetTorrentstreamBatchHistoryEvent is triggered after processing GetTorrentstreamBatchHistory.
type HandleGetTorrentstreamBatchHistoryEvent struct {
	hook_resolver.Event
	Data *torrentstream.BatchHistoryResponse `json:"data"`
}

// HandleTorrentstreamServeStreamRequestedEvent is triggered when TorrentstreamServeStream is requested.
// Prevent default to skip the default behavior and return your own data.
type HandleTorrentstreamServeStreamRequestedEvent struct {
	hook_resolver.Event
}
