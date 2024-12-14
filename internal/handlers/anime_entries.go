package handlers

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"runtime"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/library/scanner"
	"seanime/internal/library/summary"
	"seanime/internal/util"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"
	"slices"
	"strconv"
	"strings"
)

// HandleGetAnimeEntry
//
//	@summary return a media entry for the given AniList anime media id.
//	@desc This is used by the anime media entry pages to get all the data about the anime.
//	@desc This includes episodes and metadata (if any), AniList list data, download info...
//	@route /api/v1/library/anime-entry/{id} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@returns anime.Entry
func HandleGetAnimeEntry(c *RouteCtx) error {

	mId, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	animeCollection, err := c.App.GetAnimeCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	if animeCollection == nil {
		return c.RespondWithError(errors.New("anime collection not found"))
	}

	// Create a new media entry
	entry, err := anime.NewEntry(&anime.NewEntryOptions{
		MediaId:          mId,
		LocalFiles:       lfs,
		AnimeCollection:  animeCollection,
		Platform:         c.App.AnilistPlatform,
		MetadataProvider: c.App.MetadataProvider,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	c.App.FillerManager.HydrateFillerData(entry)

	return c.RespondWithData(entry)
}

//----------------------------------------------------------------------------------------------------------------------

// HandleAnimeEntryBulkAction
//
//	@summary perform given action on all the local files for the given media id.
//	@desc This is used to unmatch or toggle the lock status of all the local files for a specific media entry
//	@desc The response is not used in the frontend. The client should just refetch the entire media entry data.
//	@route /api/v1/library/anime-entry/bulk-action [PATCH]
//	@returns []anime.LocalFile
func HandleAnimeEntryBulkAction(c *RouteCtx) error {

	type body struct {
		MediaId int    `json:"mediaId"`
		Action  string `json:"action"` // "unmatch" or "toggle-lock"
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by media id
	groupedLfs := anime.GroupLocalFilesByMediaID(lfs)

	selectLfs, ok := groupedLfs[p.MediaId]
	if !ok {
		return c.RespondWithError(errors.New("no local files found for media id"))
	}

	switch p.Action {
	case "unmatch":
		lfs = lop.Map(lfs, func(item *anime.LocalFile, _ int) *anime.LocalFile {
			if item.MediaId == p.MediaId && p.MediaId != 0 {
				item.MediaId = 0
				item.Locked = false
				item.Ignored = false
			}
			return item
		})
	case "toggle-lock":
		// Flip the locked status of all the local files for the given media
		allLocked := lo.EveryBy(selectLfs, func(item *anime.LocalFile) bool { return item.Locked })
		lfs = lop.Map(lfs, func(item *anime.LocalFile, _ int) *anime.LocalFile {
			if item.MediaId == p.MediaId && p.MediaId != 0 {
				item.Locked = !allLocked
			}
			return item
		})
	}

	// Save the local files
	retLfs, err := db_bridge.SaveLocalFiles(c.App.Database, lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleOpenAnimeEntryInExplorer
//
//	@summary opens the directory of a media entry in the file explorer.
//	@desc This finds a common directory for all media entry local files and opens it in the file explorer.
//	@desc Returns 'true' whether the operation was successful or not, errors are ignored.
//	@route /api/v1/library/anime-entry/open-in-explorer [POST]
//	@returns bool
func HandleOpenAnimeEntryInExplorer(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	lf, found := lo.Find(lfs, func(i *anime.LocalFile) bool {
		return i.MediaId == p.MediaId
	})
	if !found {
		return c.RespondWithError(errors.New("local file not found"))
	}

	dir := filepath.Dir(lf.GetNormalizedPath())
	cmd := ""
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "explorer"
		wPath := strings.ReplaceAll(strings.ToLower(dir), "/", "\\")
		args = []string{wPath}
	case "darwin":
		cmd = "open"
		args = []string{dir}
	case "linux":
		cmd = "xdg-open"
		args = []string{dir}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	cmdObj := util.NewCmd(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	err = cmdObj.Run()

	return c.RespondWithData(true)

}

//----------------------------------------------------------------------------------------------------------------------

var (
	entriesSuggestionsCache = result.NewCache[string, []*anilist.BaseAnime]()
)

// HandleFetchAnimeEntrySuggestions
//
//	@summary returns a list of media suggestions for files in the given directory.
//	@desc This is used by the "Resolve unmatched media" feature to suggest media entries for the local files in the given directory.
//	@desc If some matches files are found in the directory, it will ignore them and base the suggestions on the remaining files.
//	@route /api/v1/library/anime-entry/suggestions [POST]
//	@returns []anilist.BaseAnime
func HandleFetchAnimeEntrySuggestions(c *RouteCtx) error {

	type body struct {
		Dir string `json:"dir"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	b.Dir = strings.ToLower(b.Dir)

	suggestions, found := entriesSuggestionsCache.Get(b.Dir)
	if found {
		return c.RespondWithData(suggestions)
	}

	// Retrieve local files
	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by dir
	groupedLfs := lop.GroupBy(lfs, func(item *anime.LocalFile) string {
		return filepath.Dir(item.GetNormalizedPath())
	})

	selectedLfs, found := groupedLfs[b.Dir]
	if !found {
		return c.RespondWithError(errors.New("no local files found for selected directory"))
	}

	// Filter out local files that are already matched
	selectedLfs = lo.Filter(selectedLfs, func(item *anime.LocalFile, _ int) bool {
		return item.MediaId == 0
	})

	title := selectedLfs[0].GetParsedTitle()

	c.App.Logger.Info().Str("title", title).Msg("handlers: Fetching anime suggestions")

	res, err := anilist.ListAnimeM(
		lo.ToPtr(1),
		&title,
		lo.ToPtr(8),
		nil,
		[]*anilist.MediaStatus{lo.ToPtr(anilist.MediaStatusFinished), lo.ToPtr(anilist.MediaStatusReleasing), lo.ToPtr(anilist.MediaStatusCancelled), lo.ToPtr(anilist.MediaStatusHiatus)},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		c.App.Logger,
		c.App.GetAccountToken(),
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Cache the results
	entriesSuggestionsCache.Set(b.Dir, res.GetPage().GetMedia())

	return c.RespondWithData(res.GetPage().GetMedia())

}

//----------------------------------------------------------------------------------------------------------------------

// HandleAnimeEntryManualMatch
//
//	@summary matches un-matched local files in the given directory to the given media.
//	@desc It is used by the "Resolve unmatched media" feature to manually match local files to a specific media entry.
//	@desc Matching involves the use of scanner.FileHydrator. It will also lock the files.
//	@desc The response is not used in the frontend. The client should just refetch the entire library collection.
//	@route /api/v1/library/anime-entry/manual-match [POST]
//	@returns []anime.LocalFile
func HandleAnimeEntryManualMatch(c *RouteCtx) error {

	type body struct {
		Paths   []string `json:"paths"`
		MediaId int      `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	animeCollectionWithRelations, err := c.App.AnilistPlatform.GetAnimeCollectionWithRelations()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Retrieve local files
	lfs, lfsId, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	compPaths := make(map[string]struct{})
	for _, p := range b.Paths {
		compPaths[strings.ToLower(filepath.ToSlash(p))] = struct{}{}
	}

	selectedLfs := lo.Filter(lfs, func(item *anime.LocalFile, _ int) bool {
		_, found := compPaths[item.GetNormalizedPath()]
		return found && item.MediaId == 0
	})

	// Add the media id to the selected local files
	// Also, lock the files
	selectedLfs = lop.Map(selectedLfs, func(item *anime.LocalFile, _ int) *anime.LocalFile {
		item.MediaId = b.MediaId
		item.Locked = true
		item.Ignored = false
		return item
	})

	// Get the media
	media, err := c.App.AnilistPlatform.GetAnime(b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a slice of normalized media
	normalizedMedia := []*anime.NormalizedMedia{
		anime.NewNormalizedMedia(media),
	}

	scanLogger, err := scanner.NewScanLogger(c.App.Config.Logs.Dir)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create scan summary logger
	scanSummaryLogger := summary.NewScanSummaryLogger()

	fh := scanner.FileHydrator{
		LocalFiles:         selectedLfs,
		CompleteAnimeCache: anilist.NewCompleteAnimeCache(),
		Platform:           c.App.AnilistPlatform,
		MetadataProvider:   c.App.MetadataProvider,
		AnilistRateLimiter: limiter.NewAnilistLimiter(),
		Logger:             c.App.Logger,
		ScanLogger:         scanLogger,
		ScanSummaryLogger:  scanSummaryLogger,
		AllMedia:           normalizedMedia,
		ForceMediaId:       media.GetID(),
	}

	fh.HydrateMetadata()

	// Hydrate the summary logger before merging files
	fh.ScanSummaryLogger.HydrateData(selectedLfs, normalizedMedia, animeCollectionWithRelations)

	// Save the scan summary
	go func() {
		err = db_bridge.InsertScanSummary(c.App.Database, scanSummaryLogger.GenerateSummary())
	}()

	// Remove select local files from the database slice, we will add them (hydrated) later
	selectedPaths := lop.Map(selectedLfs, func(item *anime.LocalFile, _ int) string { return item.GetNormalizedPath() })
	lfs = lo.Filter(lfs, func(item *anime.LocalFile, _ int) bool {
		if slices.Contains(selectedPaths, item.GetNormalizedPath()) {
			return false
		}
		return true
	})

	// Add the hydrated local files to the slice
	lfs = append(lfs, selectedLfs...)

	// Update the local files
	retLfs, err := db_bridge.SaveLocalFiles(c.App.Database, lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

//var missingEpisodesMap = result.NewResultMap[string, *anime.MissingEpisodes]()

// HandleGetMissingEpisodes
//
//	@summary returns a list of episodes missing from the user's library collection
//	@desc It detects missing episodes by comparing the user's AniList collection 'next airing' data with the local files.
//	@desc This route can be called multiple times, as it does not bypass the cache.
//	@route /api/v1/library/missing-episodes [GET]
//	@returns anime.MissingEpisodes
func HandleGetMissingEpisodes(c *RouteCtx) error {

	// Get the user's anilist collection
	// Do not bypass the cache, since this handler might be called multiple times, and we don't want to spam the API
	// A cron job will refresh the cache every 10 minutes
	animeCollection, err := c.App.GetAnimeCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	lfs, _, err := db_bridge.GetLocalFiles(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the silenced media ids
	silencedMediaIds, _ := c.App.Database.GetSilencedMediaEntryIds()

	missingEps := anime.NewMissingEpisodes(&anime.NewMissingEpisodesOptions{
		AnimeCollection:  animeCollection,
		LocalFiles:       lfs,
		SilencedMediaIds: silencedMediaIds,
		MetadataProvider: c.App.MetadataProvider,
	})

	return c.RespondWithData(missingEps)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleGetAnimeEntrySilenceStatus
//
//	@summary returns the silence status of a media entry.
//	@param id - int - true - "The ID of the media entry."
//	@route /api/v1/library/anime-entry/silence/{id} [GET]
//	@returns models.SilencedMediaEntry
func HandleGetAnimeEntrySilenceStatus(c *RouteCtx) error {
	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(errors.New("invalid id"))
	}

	animeEntry, err := c.App.Database.GetSilencedMediaEntry(uint(mId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.RespondWithData(false)
		} else {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(animeEntry)
}

// HandleToggleAnimeEntrySilenceStatus
//
//	@summary toggles the silence status of a media entry.
//	@desc The missing episodes should be re-fetched after this.
//	@route /api/v1/library/anime-entry/silence [POST]
//	@returns bool
func HandleToggleAnimeEntrySilenceStatus(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	animeEntry, err := c.App.Database.GetSilencedMediaEntry(uint(b.MediaId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = c.App.Database.InsertSilencedMediaEntry(uint(b.MediaId))
			if err != nil {
				return c.RespondWithError(err)
			}
			return c.RespondWithData(true)
		} else {
			return c.RespondWithError(err)
		}
	}

	err = c.App.Database.DeleteSilencedMediaEntry(animeEntry.ID)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

//-----------------------------------------------------------------------------------------------------------------------------

// HandleUpdateAnimeEntryProgress
//
//	@summary update the progress of the given anime media entry.
//	@desc This is used to update the progress of the given anime media entry on AniList.
//	@desc The response is not used in the frontend, the client should just refetch the entire media entry data.
//	@desc NOTE: This is currently only used by the 'Online streaming' feature since anime progress updates are handled by the Playback Manager.
//	@route /api/v1/library/anime-entry/update-progress [POST]
//	@returns bool
func HandleUpdateAnimeEntryProgress(c *RouteCtx) error {

	type body struct {
		MediaId       int `json:"mediaId"`
		MalId         int `json:"malId,omitempty"`
		EpisodeNumber int `json:"episodeNumber"`
		TotalEpisodes int `json:"totalEpisodes"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Update the progress on AniList
	err := c.App.AnilistPlatform.UpdateEntryProgress(
		b.MediaId,
		b.EpisodeNumber,
		&b.TotalEpisodes,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	_, _ = c.App.RefreshAnimeCollection() // Refresh the AniList collection

	return c.RespondWithData(true)
}

//-----------------------------------------------------------------------------------------------------------------------------

// HandleUpdateAnimeEntryRepeat
//
//	@summary update the repeat value of the given anime media entry.
//	@desc This is used to update the repeat value of the given anime media entry on AniList.
//	@desc The response is not used in the frontend, the client should just refetch the entire media entry data.
//	@route /api/v1/library/anime-entry/update-repeat [POST]
//	@returns bool
func HandleUpdateAnimeEntryRepeat(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
		Repeat  int `json:"repeat"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.AnilistPlatform.UpdateEntryRepeat(
		b.MediaId,
		b.Repeat,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	_, _ = c.App.RefreshAnimeCollection() // Refresh the AniList collection

	return c.RespondWithData(true)
}
