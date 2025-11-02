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

	// Return success - Electron will handle closing the app
	return h.RespondWithData(c, DownloadReleaseResponse{Destination: path})
}

// HandleDownloadMacDenshiUpdate
//
//	@summary downloads, extracts, and installs macOS update, then closes the app
//	@route /api/v1/download-mac-denshi-update [POST]
//	@returns handlers.DownloadReleaseResponse
func (h *Handler) HandleDownloadMacDenshiUpdate(c echo.Context) error {

	type body struct {
		DownloadUrl string `json:"download_url"`
		Version     string `json:"version"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	// Get downloads directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to get home directory: %w", err))
	}
	downloadsDir := filepath.Join(homeDir, "Downloads")

	// Download the file
	h.App.Logger.Info().Str("url", b.DownloadUrl).Msg("Downloading macOS update")
	resp, err := http.Get(b.DownloadUrl)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to download update: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.RespondWithError(c, fmt.Errorf("failed to download update: status %d", resp.StatusCode))
	}

	// Create temp file for download
	zipPath := filepath.Join(downloadsDir, fmt.Sprintf("seanime-denshi-%s_MacOS_arm64.zip", b.Version))
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create zip file: %w", err))
	}
	defer zipFile.Close()

	// Copy download to file
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to write zip file: %w", err))
	}
	zipFile.Close()

	h.App.Logger.Info().Str("path", zipPath).Msg("Downloaded update")

	// Extract the zip file
	extractDir := filepath.Join(downloadsDir, fmt.Sprintf("seanime-denshi-%s", b.Version))
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create extract directory: %w", err))
	}

	h.App.Logger.Info().Str("path", extractDir).Msg("Extracting update")
	cmd := util.NewCmd("unzip", "-o", zipPath, "-d", extractDir)
	if err := cmd.Run(); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to extract zip: %w", err))
	}

	// Find the .app bundle
	appPath := filepath.Join(extractDir, "Seanime Denshi.app")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return h.RespondWithError(c, fmt.Errorf("Seanime Denshi.app not found in extracted files"))
	}

	// Run xattr -c to remove quarantine attributes
	h.App.Logger.Info().Str("path", appPath).Msg("Removing quarantine attributes")
	xattrCmd := util.NewCmd("xattr", "-cr", appPath)
	if err := xattrCmd.Run(); err != nil {
		h.App.Logger.Warn().Err(err).Msg("Failed to remove quarantine attributes, continuing anyway")
	}

	// Move to Applications folder
	applicationsPath := "/Applications/Seanime Denshi.app"
	h.App.Logger.Info().Str("destination", applicationsPath).Msg("Moving to Applications")

	// Remove existing app if it exists
	if _, err := os.Stat(applicationsPath); err == nil {
		if err := os.RemoveAll(applicationsPath); err != nil {
			return h.RespondWithError(c, fmt.Errorf("failed to remove existing app: %w", err))
		}
	}

	// Move new app to Applications
	moveCmd := util.NewCmd("mv", appPath, applicationsPath)
	if err := moveCmd.Run(); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to move app to Applications: %w", err))
	}

	// Clean up downloaded files
	os.Remove(zipPath)
	os.RemoveAll(extractDir)

	h.App.Logger.Info().Msg("macOS update installed successfully")

	return h.RespondWithData(c, DownloadReleaseResponse{Destination: applicationsPath})
}
