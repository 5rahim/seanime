package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/limiter"
	"github.com/seanime-app/seanime/internal/mal"
	"github.com/seanime-app/seanime/internal/result"
	"github.com/seanime-app/seanime/internal/scanner"
	"github.com/sourcegraph/conc/pool"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
)

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
		MediaId:           mId,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
		AnilistCollection: anilistCollection,
		AnilistClient:     c.App.AnilistClient,
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

func HandleMediaEntryBulkAction(c *RouteCtx) error {

	type body struct {
		MediaId int    `json:"mediaId"`
		Action  string `json:"action"`
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
		// Convert the directory path to lowercase for case-insensitivity
		lowerCasePath := strings.ToLower(dir)
		args = []string{lowerCasePath}
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
		s := s
		p2.Go(func() *anilist.BasicMedia {
			anilistRateLimit.Wait()
			// Check if the media has already been fetched
			media, found := anilistCache.Get(s.ID)
			if found {
				return media
			}
			// Otherwise, fetch the media
			mediaRes, err := c.App.AnilistClient.BasicMediaByMalID(context.Background(), &s.ID)
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
	mediaRes, err := c.App.AnilistClient.BaseMediaByID(context.Background(), &b.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}

	fh := scanner.FileHydrator{
		LocalFiles:         selectedLfs,
		AllMedia:           []*anilist.BaseMedia{mediaRes.GetMedia()},
		BaseMediaCache:     anilist.NewBaseMediaCache(),
		AnizipCache:        anizip.NewCache(),
		AnilistClient:      c.App.AnilistClient,
		AnilistRateLimiter: limiter.NewAnilistLimiter(),
		Logger:             c.App.Logger,
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

	missingEps := entities.NewMissingEpisodes(&entities.NewMissingEpisodesOptions{
		AnilistCollection: anilistCollection,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
	})

	return c.RespondWithData(missingEps.Episodes)

}

//----------------------------------------------------------------------------------------------------------------------
