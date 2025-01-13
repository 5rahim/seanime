package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/updater"
	"seanime/internal/util"

	"github.com/labstack/echo/v4"
)

// HandleDownloadTorrentFile
//
//	@summary downloads torrent files to the destination folder
//	@route /api/v1/download-torrent-file [POST]
//	@returns bool
func (h *Handler) HandleDownloadTorrentFile(c echo.Context) error {

	type body struct {
		DownloadUrls []string           `json:"download_urls"`
		Destination  string             `json:"destination"`
		Media        *anilist.BaseAnime `json:"media"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	errs := make([]error, 0)
	for _, url := range b.DownloadUrls {
		err := downloadTorrentFile(url, b.Destination)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return h.RespondWithError(c, errs[0])
	} else if len(errs) > 1 {
		return h.RespondWithError(c, errors.New("failed to download multiple files"))
	}

	return h.RespondWithData(c, true)
}

func downloadTorrentFile(url string, dest string) (err error) {

	defer util.HandlePanicInModuleWithError("handlers/download/downloadTorrentFile", &err)

	// Get the file name from the URL
	fileName := filepath.Base(url)
	filePath := filepath.Join(dest, fileName)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download file, %s", resp.Status)
	}

	// Create the destination folder if it doesn't exist
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

type DownloadReleaseResponse struct {
	Destination string `json:"destination"`
	Error       string `json:"error,omitempty"`
}

// HandleDownloadRelease
//
//	@summary downloads selected release asset to the destination folder.
//	@desc Downloads the selected release asset to the destination folder and extracts it if possible.
//	@desc If the extraction fails, the error message will be returned in the successful response.
//	@desc The successful response will contain the destination path of the extracted files.
//	@desc It only returns an error if the download fails.
//	@route /api/v1/download-release [POST]
//	@returns handlers.DownloadReleaseResponse
func (h *Handler) HandleDownloadRelease(c echo.Context) error {

	type body struct {
		DownloadUrl string `json:"download_url"`
		Destination string `json:"destination"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	path, err := h.App.Updater.DownloadLatestRelease(b.DownloadUrl, b.Destination)

	if err != nil {
		if errors.Is(err, updater.ErrExtractionFailed) {
			return h.RespondWithData(c, DownloadReleaseResponse{Destination: path, Error: err.Error()})
		}
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, DownloadReleaseResponse{Destination: path})
}
