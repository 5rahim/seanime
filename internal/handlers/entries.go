package handlers

import (
	"errors"
	"fmt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/constants"
	"github.com/seanime-app/seanime-server/internal/entities"
	"github.com/seanime-app/seanime-server/internal/result"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func HandleGetMediaEntry(c *RouteCtx) error {

	type query struct {
		MediaId int `query:"mediaId" json:"mediaId"`
	}

	p := new(query)
	if err := c.Fiber.QueryParser(p); err != nil {
		return c.RespondWithError(err)
	}

	// Get all the local files
	lfs, err := getLocalFilesFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Get the user's anilist collection
	anilistCollection, err := c.App.GetAnilistCollection()
	if err != nil {
		return c.RespondWithError(err)
	}

	// Create a new media entry
	entry, err := entities.NewMediaEntry(&entities.NewMediaEntryOptions{
		MediaId:           p.MediaId,
		LocalFiles:        lfs,
		AnizipCache:       c.App.AnizipCache,
		AnilistCollection: anilistCollection,
		AnilistClient:     c.App.AnilistClient,
	})
	if err != nil {
		return c.RespondWithError(err)
	}

	// Fetch media details in the background and send them via websocket
	go func() {
		details, err := c.App.AnilistClient.MediaDetailsByID(c.Fiber.Context(), &p.MediaId)
		if err == nil {
			c.App.WSEventManager.SendEvent(constants.EventMediaDetails, details)
		}
	}()

	return c.RespondWithData(entry)
}

//----------------------------------------------------------------------------------------------------------------------

var (
	detailsCache = result.NewCache[int, *anilist.MediaDetailsById_Media]()
)

func HandleGetMediaDetails(c *RouteCtx) error {
	type query struct {
		MediaId int `query:"mediaId" json:"mediaId"`
	}

	p := new(query)
	if err := c.Fiber.QueryParser(p); err != nil {
		return c.RespondWithError(err)
	}

	if details, ok := detailsCache.Get(p.MediaId); ok {
		return c.RespondWithData(details)
	}
	details, err := c.App.AnilistClient.MediaDetailsByID(c.Fiber.Context(), &p.MediaId)
	if err != nil {
		return c.RespondWithError(err)
	}
	detailsCache.Set(p.MediaId, details.GetMedia())

	return c.RespondWithData(details.GetMedia())
}

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
	lfs, dbId, err := getLocalFilesAndIdFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Group local files by media id
	groupedLfs := entities.GetGroupedLocalFiles(lfs)

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
	retLfs, err := saveLocalFilesInDB(c.App.Database, dbId, lfs)
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
	lfs, _, err := getLocalFilesAndIdFromDB(c.App.Database)
	if err != nil {
		return c.RespondWithError(err)
	}

	lf, found := lo.Find(lfs, func(i *entities.LocalFile) bool {
		return i.MediaId == p.MediaId
	})
	if !found {
		return c.RespondWithError(errors.New("local file not found"))
	}

	dir := filepath.Dir(lf.Path)
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

func HandleStartDefaultPlayer(c *RouteCtx) error {

	settings, err := c.App.Database.GetSettings()
	if err != nil {
		return c.RespondWithError(err)
	}

	switch settings.MediaPlayer.Default {
	case "vlc":
		err = c.App.MediaPlayer.VLC.Start()
		if err != nil {
			return c.RespondWithError(err)
		}
	case "mpc-hc":
		err = c.App.MediaPlayer.MpcHc.Start()
		if err != nil {
			return c.RespondWithError(err)
		}
	}

	return c.RespondWithData(true)

}

//----------------------------------------------------------------------------------------------------------------------

func HandleEditMediaListData(c *RouteCtx) error {

	type body struct {
		MediaId   *int                     `json:"mediaId"`
		Status    *anilist.MediaListStatus `json:"status"`
		Score     *int                     `json:"score"`
		Progress  *int                     `json:"progress"`
		StartDate *anilist.FuzzyDateInput  `json:"startedAt"`
		EndDate   *anilist.FuzzyDateInput  `json:"completedAt"`
	}

	p := new(body)
	if err := c.Fiber.BodyParser(p); err != nil {
		return c.RespondWithError(err)
	}

	ret, err := c.App.AnilistClient.UpdateMediaListEntry(
		c.Fiber.Context(),
		p.MediaId,
		p.Status,
		p.Score,
		p.Progress,
		p.StartDate,
		p.EndDate,
	)
	if err != nil {
		return c.RespondWithError(err)
	}

	// Refresh the anilist collection
	_, _ = c.App.RefreshAnilistCollection()

	return c.RespondWithData(ret)
}
