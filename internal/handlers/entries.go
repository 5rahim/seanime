package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/adrg/strutil/metrics"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/api/mal"
	"github.com/seanime-app/seanime/internal/events"
	"github.com/seanime-app/seanime/internal/library/anime"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"github.com/seanime-app/seanime/internal/util/result"
	"github.com/sourcegraph/conc/pool"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
)

// HandleGetMediaEntry
//
//	@summary return a media entry for the given AniList anime media id.
//	@desc This is used by the anime media entry pages to get all the data about the anime.
//	@desc This includes episodes and metadata (if any), AniList list data, download info...
//	@route /api/v1/library/media-entry/{id} [GET]
//	@param id - int - true - "AniList anime media ID"
//	@returns entities.MediaEntry
func HandleGetMediaEntry(c *RouteCtx) error {

	mId, err := c.Fiber.ParamsInt("id")
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
	entry, err := anime.NewMediaEntry(&anime.NewMediaEntryOptions{
		MediaId:              mId,
		LocalFiles:           lfs,
		AnizipCache:          c.App.AnizipCache,
		AnilistCollection:    anilistCollection,
		AnilistClientWrapper: c.App.AnilistClientWrapper,
		MetadataProvider:     c.App.MetadataProvider,
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

// HandleMediaEntryBulkAction
//
//	@summary perform given action on all the local files for the given media id.
//	@desc This is used to unmatch or toggle the lock status of all the local files for a specific media entry
//	@desc The response is not used in the frontend. The client should just refetch the entire media entry data.
//	@route /api/v1/library/media-entry/bulk-action [PATCH]
//	@returns []entities.LocalFile
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
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleOpenMediaEntryInExplorer
//
//	@summary opens the directory of a media entry in the file explorer.
//	@desc This finds a common directory for all media entry local files and opens it in the file explorer.
//	@desc Returns 'true' whether the operation was successful or not, errors are ignored.
//	@route /api/v1/library/media-entry/open-in-explorer [POST]
//	@returns boolean
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
	cmdObj := exec.Command(cmd, args...)
	cmdObj.Stdout = os.Stdout
	cmdObj.Stderr = os.Stderr
	err = cmdObj.Run()

	return c.RespondWithData(true)

}

//----------------------------------------------------------------------------------------------------------------------

var (
	entriesMalCache               = result.NewCache[string, []*mal.SearchResultAnime]()
	entriesAnilistBasicMediaCache = result.NewCache[int, *anilist.BasicMedia]()
)

// HandleFindProspectiveMediaEntrySuggestions
//
//	@summary returns a list of media suggestions for files in the given directory.
//	@desc This is used by the "Resolve unmatched media" feature to suggest media entries for the local files in the given directory.
//	@desc If some matches files are found in the directory, it will ignore them and base the suggestions on the remaining files.
//	@route /api/v1/library/media-entry/suggestions [POST]
//	@returns []anilist.BasicMedia
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

	// Fetch 8 suggestions from MAL
	malSuggestions, err := entriesMalCache.GetOrSet(title, func() ([]*mal.SearchResultAnime, error) {
		malSuggestions, err := mal.SearchWithMAL(title, 8)
		if err != nil {
			return nil, err
		}
		// Cache the results
		entriesMalCache.Set(title, malSuggestions)
		return malSuggestions, nil
	})
	if err != nil {
		return c.RespondWithError(err)
	}
	if len(malSuggestions) == 0 {
		return c.RespondWithData([]*anilist.BasicMedia{})
	}

	dice := metrics.NewSorensenDice()
	dice.CaseSensitive = false
	// Sort by top 4 suggestions
	malRatings := lo.Map(malSuggestions, func(item *mal.SearchResultAnime, _ int) struct {
		OriginalValue string
		Rating        float64
	} {
		return struct {
			OriginalValue string
			Rating        float64
		}{
			OriginalValue: item.Name,
			Rating:        dice.Compare(title, item.Name),
		}
	})
	// Sort by top 4 suggestions
	sort.SliceStable(malRatings, func(i, j int) bool {
		return malRatings[i].Rating > malRatings[j].Rating
	})

	_malSuggestions := make([]*mal.SearchResultAnime, 0)
	for idx, item := range malRatings {
		if idx < 4 {
			s, ok := lo.Find(malSuggestions, func(i *mal.SearchResultAnime) bool {
				return i.Name == item.OriginalValue
			})
			if ok {
				_malSuggestions = append(_malSuggestions, s)
			}
		}
	}
	malSuggestions = _malSuggestions

	anilistRateLimit := limiter.NewAnilistLimiter()
	p2 := pool.NewWithResults[*anilist.BasicMedia]()
	for _, s := range malSuggestions {
		p2.Go(func() *anilist.BasicMedia {
			anilistRateLimit.Wait()
			// Check if the media has already been fetched
			media, found := entriesAnilistBasicMediaCache.Get(s.ID)
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
			entriesAnilistBasicMediaCache.Set(s.ID, media)
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

// HandleMediaEntryManualMatch
//
//	@summary matches un-matched local files in the given directory to the given media.
//	@desc It is used by the "Resolve unmatched media" feature to manually match local files to a specific media entry.
//	@desc Matching involves the use of scanner.FileHydrator. It will also lock the files.
//	@desc The response is not used in the frontend. The client should just refetch the entire library collection.
//	@route /api/v1/library/media-entry/manual-match [POST]
//	@returns []entities.LocalFile
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
	groupedLfs := lop.GroupBy(lfs, func(item *anime.LocalFile) string {
		return filepath.Dir(item.GetNormalizedPath())
	})

	selectedLfs, found := groupedLfs[strings.ToLower(b.Dir)]
	if !found {
		return c.RespondWithError(errors.New("no local files found for selected directory"))
	}

	// Add the media id to the selected local files
	// Also, lock the files
	selectedLfs = lop.Map(selectedLfs, func(item *anime.LocalFile, _ int) *anime.LocalFile {
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
		AllMedia: []*anime.NormalizedMedia{
			anime.NewNormalizedMedia(mediaRes.GetMedia().ToBasicMedia()),
		},
		ForceMediaId: mediaRes.GetMedia().GetID(),
	}

	fh.HydrateMetadata()

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
	retLfs, err := c.App.Database.SaveLocalFiles(lfsId, lfs)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(retLfs)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleGetMissingEpisodes
//
//	@summary returns a list of episodes missing from the user's library collection
//	@desc It detects missing episodes by comparing the user's AniList collection 'next airing' data with the local files.
//	@desc This route can be called multiple times, as it does not bypass the cache.
//	@route /api/v1/library/missing-episodes [GET]
//	@returns entities.MissingEpisodes
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

	missingEps := anime.NewMissingEpisodes(&anime.NewMissingEpisodesOptions{
		AnilistCollection: anilistCollection,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
		SilencedMediaIds:  silencedMediaIds,
		MetadataProvider:  c.App.MetadataProvider,
	})

	return c.RespondWithData(missingEps)

}

//----------------------------------------------------------------------------------------------------------------------

// HandleAddUnknownMedia
//
//	@summary adds the given media to the user's AniList planning collections
//	@desc Since media not found in the user's AniList collection are not displayed in the library, this route is used to add them.
//	@desc The response is ignored in the frontend, the client should just refetch the entire library collection.
//	@route /api/v1/media-entry/unknown-media [POST]
//	@returns anilist.AnimeCollection
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

//-----------------------------------------------------------------------------------------------------------------------------

// HandleUpdateProgress
//
//	@summary update the progress of the given anime media entry.
//	@desc This is used to update the progress of the given anime media entry on AniList and MyAnimeList (if an account is linked).
//	@desc The response is not used in the frontend, the client should just refetch the entire media entry data.
//	@desc NOTE: This is currently only used by the 'Online streaming' feature since anime progress updates are handled by the Playback Manager.
//	@route /api/v1/media-entry/update-progress [POST]
//	@returns boolean
func HandleUpdateProgress(c *RouteCtx) error {

	type body struct {
		MediaId       int `json:"mediaId"`
		MalId         int `json:"malId"`
		EpisodeNumber int `json:"episodeNumber"`
		TotalEpisodes int `json:"totalEpisodes"`
	}

	b := new(body)
	if err := c.Fiber.BodyParser(b); err != nil {
		return c.RespondWithError(err)
	}

	// Update the progress on AniList
	err := c.App.AnilistClientWrapper.UpdateMediaListEntryProgress(
		context.Background(),
		&b.MediaId,
		&b.EpisodeNumber,
		&b.TotalEpisodes,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	_, _ = c.App.RefreshAnilistCollection() // Refresh the AniList collection

	go func() {
		// Update the progress on MAL if an account is linked
		malInfo, _ := c.App.Database.GetMalInfo()
		if malInfo != nil && malInfo.AccessToken != "" && b.MalId > 0 {

			// Verify MAL auth
			malInfo, err = mal.VerifyMALAuth(malInfo, c.App.Database, c.App.Logger)
			if err != nil {
				c.App.WSEventManager.SendEvent(events.WarningToast, "Failed to update progress on MyAnimeList")
				return
			}

			client := mal.NewWrapper(malInfo.AccessToken, c.App.Logger)
			err = client.UpdateAnimeProgress(&mal.AnimeListProgressParams{
				NumEpisodesWatched: &b.EpisodeNumber,
			}, b.MalId)
			if err != nil {
				c.App.WSEventManager.SendEvent(events.WarningToast, "Failed to update progress on MyAnimeList")
			}
		}
	}()

	return c.RespondWithData(true)
}
