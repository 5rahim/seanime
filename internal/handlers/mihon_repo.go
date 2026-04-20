package handlers

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed mihon_repo/*.apk
var mihonRepoFS embed.FS

type mihonRepoEntry struct {
	Name    string   `json:"name"`
	Pkg     string   `json:"pkg"`
	Apk     string   `json:"apk"`
	Lang    string   `json:"lang"`
	Code    int      `json:"code"`
	Version string   `json:"version"`
	NSFW    int      `json:"nsfw"`
	Sources []mihonRepoSource `json:"sources"`
}

type mihonRepoSource struct {
	Name string `json:"name"`
	Lang string `json:"lang"`
	ID   int64  `json:"id"`
}

// HandleMihonRepoIndex serves the extension repo index for Mihon.
//
//	@summary returns the Mihon extension repo index.
//	@route /api/v1/mihon/repo/index.min.json [GET]
func (h *Handler) HandleMihonRepoIndex(c echo.Context) error {
	index := []mihonRepoEntry{
		{
			Name:    "Seanime",
			Pkg:     "eu.kanade.tachiyomi.extension.all.seanime",
			Apk:     "tachiyomi-all.seanime-v1.4.1.apk",
			Lang:    "all",
			Code:    1,
			Version: "1.4.1",
			NSFW:    0,
			Sources: []mihonRepoSource{
				{
					Name: "Seanime",
					Lang: "all",
					ID:   0,
				},
			},
		},
	}
	return c.JSON(http.StatusOK, index)
}

// HandleMihonRepoAPK serves the extension APK file.
//
//	@summary serves the Mihon extension APK.
//	@route /api/v1/mihon/repo/apk/:name [GET]
func (h *Handler) HandleMihonRepoAPK(c echo.Context) error {
	name := c.Param("name")

	data, err := fs.ReadFile(mihonRepoFS, "mihon_repo/"+name)
	if err != nil {
		return c.String(http.StatusNotFound, "APK not found")
	}

	return c.Blob(http.StatusOK, "application/vnd.android.package-archive", data)
}
