package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/mal"
	"github.com/seanime-app/seanime/internal/library/entities"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/seanime-app/seanime/internal/util/result"
	"github.com/sourcegraph/conc/pool"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

// HandleGetMediaEntry will return the media entry (entities.MediaEntry) with the given media id.
//
//	GET /v1/library/media-entry/:id
func HandleGetMediaEntry(c *RouteCtx) error {

	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media entry
	entry, err := entities.NewMediaEntry(&entities.NewMediaEntryOptions{
		MediaId:              mId,
		LocalFiles:           lfs,
		AnizipCache:          c.App.AnizipCache,
		AnilistCollection:    anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(entry)
}

// HandleGetSimpleMediaEntry will return the simple media entry (entities.SimpleMediaEntry) with the given media id.
//
//	GET /v1/simple-media-entry/:id
//
// DEPRECATED: Use HandleGetMediaEntry instead.
func HandleGetSimpleMediaEntry(c *RouteCtx) error {

	mId, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media entry
	entry, err := entities.NewSimpleMediaEntry(&entities.NewSimpleMediaEntryOptions{
		MediaId:              mId,
		LocalFiles:           lfs,
		AnilistCollection:    anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(entry)
}

//----------------------------------------------------------------------------------------------------------------------

var (
	detailsCache = result.NewCache[int, *anilist.MediaDetailsById_Media]()
)

//----------------------------------------------------------------------------------------------------------------------

// HandleMediaEntryBulkAction will perform the given action on all the local files for the given media id.
// It will return the updated local files.
//
//	PATCH /v1/library/media-entry/bulk-action
func HandleMediaEntryBulkAction(c *RouteCtx) error {

	type body struct {
		MediaId int    `json:"mediaId"`
		Action  string `json:"action"` // "unmatch" or "toggle-lock"
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, lfsId, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by media id
	groupedLfs := entities.GroupLocalFilesByMediaID(lfs)

	selectLfs, ok := groupedLfs[p.MediaId]
	if !ok {
		return c.RespondWithError(errors.New("no local files found for media id"))
	}

	switch p.Action {
	case "unmatch":
		lfs = lop.Map(lfs, func(item *entities.LocalFile, _ int) *entities.LocalFile {
			if item.MediaId == p.MediaId && p.MediaId != 0 {
				item.MediaId = 0
				item.Locked = false
				item.Ignored = false
			}
			return item
		})
	case "toggle-lock":
		// Flip the locked status of all the local files for the given media
		allLocked := lo.EveryBy(selectLfs, func(item *entities.LocalFile) bool { return item.Locked })
		lfs = lop.Map(lfs, func(item *entities.LocalFile, _ int) *entities.LocalFile {
			if item.MediaId == p.MediaId && p.MediaId != 0 {
				item.Locked = !allLocked
			}
			return item
		})
	}

	// Save the local files
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleOpenMediaEntryInExplorer will open the directory of the local files for the given media id in the file explorer.
// It will return true if the operation was successful. (Note: the operation can still fail even if true is returned)
//
//	POST /v1/library/media-entry/open-in-explorer
func HandleOpenMediaEntryInExplorer(c *RouteCtx) error {

	type body struct {
		MediaId int `json:"mediaId"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	lf, found := lo.Find(lfs, func(i *entities.LocalFile) bool {
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
	cmdObj := exec.Command(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	err = cmdObj.Run()

	return c.RespondWithData(true)

}

//----------------------------------------------------------------------------------------------------------------------

var (
	malCache     = result.NewCache[string, []*mal.SearchResultAnime]()
	anilistCache = result.NewCache[int, *anilist.BasicMedia]()
)

// HandleFindProspectiveMediaEntrySuggestions will return a list of media suggestions for files in the given directory.
// This is used by the "Resolve unmatched media" feature to suggest media entries for the local files in the given directory.
//
// It uses the title of the first local file in the directory to fetch suggestions from MAL.
// It will return a list of anilist.BasicMedia.
//
//	POST /v1/library/media-entry/suggestions
func HandleFindProspectiveMediaEntrySuggestions(c *RouteCtx) error {

	type body struct {
		Dir string `json:"dir"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	b.Dir = strings.ToLower(b.Dir)

	// Retrieve local files
	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by dir
	groupedLfs := lop.GroupBy(lfs, func(item *entities.LocalFile) string {
		return filepath.Dir(item.GetNormalizedPath())
	})

	selectedLfs, found := groupedLfs[b.Dir]
	if !found {
		return c.RespondWithError(errors.New("no local files found for selected directory"))
	}

	// Filter out local files that are already matched
	selectedLfs = lo.Filter(selectedLfs, func(item *entities.LocalFile, _ int) bool {
		return item.MediaId == 0
	})

	title := selectedLfs[0].GetParsedTitle()

	// Fetch 8 suggestions from MAL
	malSuggestions, err := mal.SearchWithMAL(title, 8)
	if err != nil {
		return c.RespondWithError(err)
	}
	if len(malSuggestions) == 0 {
		return c.RespondWithData([]*anilist.BasicMedia{})
	}

	// Cache the results (10 minutes)
	malCache.Set(title, malSuggestions)

	anilistRateLimit := limiter.NewAnilistLimiter()
	p2 := pool.NewWithResults[*anilist.BasicMedia]()
	for _, s := range malSuggestions {
		p2.Go(func() *anilist.BasicMedia {
			anilistRateLimit.Wait()
			// Check if the media has already been fetched
			media, found := anilistCache.Get(s.ID)
			if found {
				return media
			}
			// Otherwise, fetch the media
			mediaRes, err := c.App.AnilistClientWrapper.BasicMediaByMalID(context.Background(), &s.ID)
			if err != nil {
				return nil
			}
			media = mediaRes.GetMedia()
			// Cache the media
			anilistCache.Set(s.ID, media)
			return media
		})
	}
	anilistMedia := p2.Wait()
	anilistMedia = lo.Filter(anilistMedia, func(item *anilist.BasicMedia, _ int) bool {
		return item != nil
	})

	return c.RespondWithData(anilistMedia)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleMediaEntryManualMatch will match the local files in the given directory to the given media.
// It is used by the "Resolve unmatched media" feature to manually match local files to media entries.
//
//   - It will hydrate the local files with the appropriate metadata by using scanner.FileHydrator.
//   - It will also add the media id to the selected local files and lock them.
//
// It will return the updated local files.
//
//	POST /v1/library/media-entry/manual-match
func HandleMediaEntryManualMatch(c *RouteCtx) error {

	type body struct {
		Dir     string `json:"dir"`
		MediaId int    `json:"mediaId"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Retrieve local files
	lfs, lfsId, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by dir
	groupedLfs := lop.GroupBy(lfs, func(item *entities.LocalFile) string {
		return filepath.Dir(item.GetNormalizedPath())
	})

	selectedLfs, found := groupedLfs[strings.ToLower(b.Dir)]
	if !found {
		return c.RespondWithError(errors.New("no local files found for selected directory"))
	}

	// Add the media id to the selected local files
	// Also, lock the files
	selectedLfs = lop.Map(selectedLfs, func(item *entities.LocalFile, _ int) *entities.LocalFile {
		item.MediaId = b.MediaId
		item.Locked = true
		item.Ignored = false
		return item
	})

	// Get the media
	mediaRes, err := c.App.AnilistClientWrapper.BaseMediaByID(context.Background(), &b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	scanLogger, err := scanner.NewScanLogger(c.App.Config.Logs.Dir)
	if err != nil {
		return c.RespondWithError(err)
	}

	fh := scanner.FileHydrator{
		LocalFiles:           selectedLfs,
		BaseMediaCache:       anilist.NewBaseMediaCache(),
		AnizipCache:          anizip.NewCache(),
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		AnilistRateLimiter:   limiter.NewAnilistLimiter(),
		Logger:               c.App.Logger,
		ScanLogger:           scanLogger,
		AllMedia: []*entities.NormalizedMedia{
			entities.NewNormalizedMedia(mediaRes.GetMedia().ToBasicMedia()),
		},
	}

	fh.HydrateMetadata()

	// Remove select local files from the database slice, we will add them (hydrated) later
	selectedPaths := lop.Map(selectedLfs, func(item *entities.LocalFile, _ int) string { return item.GetNormalizedPath() })
	lfs = lo.Filter(lfs, func(item *entities.LocalFile, _ int) bool {
		if slices.Contains(selectedPaths, item.GetNormalizedPath()) {
			return false
		}
		return true
	})

	// Add the hydrated local files to the slice
	lfs = append(lfs, selectedLfs...)

	// Update the local files
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleGetMissingEpisodes will return a list of missing episodes from the user's library collection.
// Missing episodes are detected using data coming from the user's AniList collection.
//
//	GET /v1/library/missing-episodes
func HandleGetMissingEpisodes(c *RouteCtx) error {

	lfs, _, err := c.App.Database.GetLocalFiles()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	// Do not bypass the cache, since this handler might be called multiple times, and we don't want to spam the API
	// A cron job will refresh the cache every 10 minutes
	anilistCollection, err := c.App.GetAnilistCollection(false)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the silenced media ids
	silencedMediaIds, _ := c.App.Database.GetSilencedMediaEntryIds()

	missingEps := entities.NewMissingEpisodes(&entities.NewMissingEpisodesOptions{
		AnilistCollection: anilistCollection,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
		SilencedMediaIds:  silencedMediaIds,
	})

	return c.RespondWithData(missingEps)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleAddUnknownMedia will add the given media ids to the user's AniList planning collection.
//
//	POST /v1/library/unknown-media
func HandleAddUnknownMedia(c *RouteCtx) error {

	type body struct {
		MediaIds []int `json:"mediaIds"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Add non-added media entries to AniList collection
	if err := c.App.AnilistClientWrapper.AddMediaToPlanning(b.MediaIds, limiter.NewAnilistLimiter(), c.App.Logger); err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, this is most likely a rate limit issue"))
	}

	// Bypass the cache
	anilistCollection, err := c.App.GetAnilistCollection(true)
	if err != nil {
		return c.RespondWithError(errors.New("error: Anilist responded with an error, wait one minute before refreshing"))
	}

	return c.RespondWithData(anilistCollection)

}
