package handlers

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	urlpkg "net/url"
	"os"
	"path"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/updater"
	"seanime/internal/util"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

type downloadGrantChallenge struct {
	code      string
	clientId  string
	createdAt time.Time
}

var (
	downloadGrantChallenges   = make(map[string]*downloadGrantChallenge) // keyed by challenge ID
	downloadGrantChallengesMu sync.Mutex
	downloadGrantChallengeTTL = 2 * time.Minute
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
		ClientId     string             `json:"clientId"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	if b.Destination == "" {
		return h.RespondWithError(c, errors.New("destination not found"))
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination path must be absolute"))
	}

	if err := h.guardStrictFilesystemPath(c, b.Destination); err != nil {
		return err
	}

	contextClientId := getContextClientId(c)
	if contextClientId == "" {
		return h.RespondWithError(c, fmt.Errorf("client session not found"))
	}

	if !strings.HasPrefix(b.ClientId, "CODE:") {
		challengeID := util.RandomStringWithAlphabet(16, "abcdefghijklmnopqrstuvwxyz0123456789")
		randomCode := util.RandomStringWithAlphabet(32, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

		downloadGrantChallengesMu.Lock()
		for k, ch := range downloadGrantChallenges {
			if time.Since(ch.createdAt) > downloadGrantChallengeTTL {
				delete(downloadGrantChallenges, k)
			}
		}
		downloadGrantChallenges[challengeID] = &downloadGrantChallenge{
			code:      randomCode,
			clientId:  contextClientId,
			createdAt: time.Now(),
		}
		downloadGrantChallengesMu.Unlock()

		h.App.WSEventManager.SendEventTo(contextClientId, "download-torrent-file-permission-check", challengeID+":"+randomCode)
		return h.RespondWithData(c, false)
	}

	payload := strings.TrimPrefix(b.ClientId, "CODE:")
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		return h.RespondWithError(c, fmt.Errorf("invalid verification format"))
	}
	challengeID := parts[0]
	submittedCode := parts[1]

	downloadGrantChallengesMu.Lock()
	challenge, exists := downloadGrantChallenges[challengeID]
	if exists {
		delete(downloadGrantChallenges, challengeID)
	}
	downloadGrantChallengesMu.Unlock()

	if !exists {
		return h.RespondWithError(c, fmt.Errorf("no pending verification found"))
	}

	if time.Since(challenge.createdAt) > downloadGrantChallengeTTL {
		return h.RespondWithError(c, fmt.Errorf("verification code expired"))
	}

	if challenge.code != submittedCode {
		return h.RespondWithError(c, fmt.Errorf("invalid verification code"))
	}

	if challenge.clientId != contextClientId {
		return h.RespondWithError(c, fmt.Errorf("verification code does not belong to this client session"))
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

	fileName := getTorrentFileName(resp, url)
	if fileName == "" {
		return fmt.Errorf("failed to determine file name")
	}
	filePath := filepath.Join(dest, fileName)

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

func getTorrentFileName(resp *http.Response, downloadURL string) string {
	if resp != nil {
		contentDisposition := resp.Header.Get("Content-Disposition")
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			if fileName := cleanTorrentFileName(params["filename"]); fileName != "" {
				return fileName
			}
		}
	}

	if parsedURL, err := urlpkg.Parse(downloadURL); err == nil {
		fileName := path.Base(parsedURL.Path)
		if unescaped, err := urlpkg.PathUnescape(fileName); err == nil {
			fileName = unescaped
		}
		if fileName := cleanTorrentFileName(fileName); fileName != "" && resp != nil {
			contentType := resp.Header.Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			isTorrentType := err == nil && mediaType == "application/x-bittorrent"
			if filepath.Ext(fileName) == "" && isTorrentType {
				return fileName + ".torrent"
			}
			return fileName
		}
	}

	return cleanTorrentFileName(downloadURL)
}

func cleanTorrentFileName(f string) string {
	f = strings.TrimSpace(f)
	f = strings.ReplaceAll(f, "\\", "/")
	f = path.Base(f)
	if f == "." || f == ".." || f == "/" {
		return ""
	}
	return f
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

	if err := h.guardStrictLocalOnlyAction(c); err != nil {
		return err
	}

	if err := h.guardStrictFilesystemPath(c, b.Destination); err != nil {
		return err
	}

	if err := util.ValidateReleaseUrl(b.DownloadUrl); err != nil {
		return h.RespondWithError(c, fmt.Errorf("invalid download URL: %w", err))
	}

	if b.Destination == "" {
		return h.RespondWithError(c, errors.New("destination not found"))
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination path must be absolute"))
	}

	if err := h.guardStrictFilesystemPath(c, b.Destination); err != nil {
		return err
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
	if err := h.guardPrivilegedLocalExecution(c); err != nil {
		return err
	}

	type body struct {
		DownloadUrl string `json:"download_url"`
		Version     string `json:"version"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := util.ValidateReleaseUrl(b.DownloadUrl); err != nil {
		return h.RespondWithError(c, fmt.Errorf("invalid download URL: %w", err))
	}

	if strings.ContainsAny(b.Version, "/\\") || strings.Contains(b.Version, "..") || b.Version == "" {
		return h.RespondWithError(c, fmt.Errorf("invalid version string"))
	}

	stageDir, err := os.MkdirTemp("", "seanime-denshi-update-")
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create update staging directory: %w", err))
	}
	defer os.RemoveAll(stageDir)

	// Download the file
	h.App.Logger.Info().Str("url", b.DownloadUrl).Msg("app: Downloading macOS update")
	resp, err := http.Get(b.DownloadUrl)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to download update: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return h.RespondWithError(c, fmt.Errorf("failed to download update: status %d", resp.StatusCode))
	}

	// Create temp file for download
	zipPath := filepath.Join(stageDir, fmt.Sprintf("seanime-denshi-%s_MacOS_arm64.zip", b.Version))
	zipFile, err := os.OpenFile(zipPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create zip file: %w", err))
	}

	// Copy download to file
	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		_ = zipFile.Close()
		return h.RespondWithError(c, fmt.Errorf("failed to write zip file: %w", err))
	}
	if err := zipFile.Close(); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to finalize zip file: %w", err))
	}

	h.App.Logger.Info().Str("path", zipPath).Msg("app: Downloaded update")
	if err := validateMacAppArchive(zipPath, "Seanime Denshi.app"); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to validate update archive: %w", err))
	}

	// Extract the zip file
	extractDir := filepath.Join(stageDir, "extracted")
	err = os.MkdirAll(extractDir, 0755)
	if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to create extract directory: %w", err))
	}

	h.App.Logger.Info().Str("path", extractDir).Msg("app: Extracting update")
	if err := extractMacAppArchive(zipPath, extractDir, "Seanime Denshi.app"); err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to extract zip: %w", err))
	}

	// Find the .app bundle
	appPath := filepath.Join(extractDir, "Seanime Denshi.app")
	appInfo, err := os.Stat(appPath)
	if os.IsNotExist(err) {
		return h.RespondWithError(c, fmt.Errorf("app: Seanime Denshi.app not found in extracted files"))
	} else if err != nil {
		return h.RespondWithError(c, fmt.Errorf("failed to inspect extracted app: %w", err))
	}
	if !appInfo.IsDir() {
		return h.RespondWithError(c, fmt.Errorf("app: Seanime Denshi.app is not a directory"))
	}

	// Run xattr -c to remove quarantine attributes
	h.App.Logger.Info().Str("path", appPath).Msg("app: Removing quarantine attributes")
	if err := util.ClearMacAppQuarantine(appPath); err != nil {
		h.App.Logger.Warn().Err(err).Msg("app: Failed to remove quarantine attributes natively, falling back to xattr")
		xattrCmd := util.NewCmd("xattr", "-cr", appPath)
		if err := xattrCmd.Run(); err != nil {
			h.App.Logger.Warn().Err(err).Msg("app: Failed to remove quarantine attributes, continuing anyway")
		}
	}

	// Move to Applications folder
	applicationsPath := "/Applications/Seanime Denshi.app"
	h.App.Logger.Info().Str("destination", applicationsPath).Msg("app: Installing update into Applications")
	if err := installMacAppBundle(appPath, applicationsPath, nil); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.Logger.Info().Msg("app: macOS update installed successfully")

	return h.RespondWithData(c, DownloadReleaseResponse{Destination: applicationsPath})
}

func validateMacAppArchive(archivePath string, bundleName string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	validationRoot := filepath.Join(os.TempDir(), "seanime-denshi-update-archive-validation")
	bundleRoot := path.Clean("/" + bundleName)
	foundBundle := false

	for _, file := range r.File {
		normalizedEntry, err := normalizeArchiveEntryPath(file.Name)
		if err != nil {
			return err
		}

		mode := file.Mode()
		if _, err := util.ResolveArchiveEntryPath(validationRoot, file.Name); err != nil {
			return fmt.Errorf("failed to resolve archive path: %w", err)
		}

		switch {
		case mode&os.ModeSymlink != 0:
			if err := validateMacAppArchiveSymlink(file, normalizedEntry, bundleRoot); err != nil {
				return err
			}
		case mode.IsRegular() || file.FileInfo().IsDir():
		default:
			return fmt.Errorf("%w: %s", util.ErrUnsupportedArchiveEntry, file.Name)
		}

		if normalizedEntry == bundleRoot || strings.HasPrefix(normalizedEntry, bundleRoot+"/") {
			foundBundle = true
		}
	}

	if !foundBundle {
		return fmt.Errorf("app: %s not found in archive", bundleName)
	}

	return nil
}

func normalizeArchiveEntryPath(entryName string) (string, error) {
	normalizedEntry := path.Clean("/" + strings.ReplaceAll(entryName, `\\`, "/"))
	if normalizedEntry == "/" {
		return "", fmt.Errorf("invalid archive entry path: %q", entryName)
	}

	return normalizedEntry, nil
}

func validateMacAppArchiveSymlink(file *zip.File, entryPath string, bundleRoot string) error {
	targetPath, err := readZipSymlinkTarget(file)
	if err != nil {
		return err
	}
	if targetPath == "" || path.IsAbs(targetPath) {
		return fmt.Errorf("%w: %s", util.ErrArchivePathTraversal, file.Name)
	}

	resolvedTarget := path.Clean(path.Join(path.Dir(entryPath), targetPath))
	if resolvedTarget != bundleRoot && !strings.HasPrefix(resolvedTarget, bundleRoot+"/") {
		return fmt.Errorf("%w: %s", util.ErrArchivePathTraversal, file.Name)
	}

	return nil
}

func readZipSymlinkTarget(file *zip.File) (string, error) {
	if file.UncompressedSize64 == 0 || file.UncompressedSize64 > 4096 {
		return "", fmt.Errorf("%w: %s", util.ErrUnsupportedArchiveEntry, file.Name)
	}

	rc, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open symlink entry %s: %w", file.Name, err)
	}
	defer rc.Close()

	targetBytes, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("failed to read symlink target %s: %w", file.Name, err)
	}

	return strings.ReplaceAll(string(targetBytes), `\\`, "/"), nil
}

func extractMacAppArchive(archivePath string, dest string, bundleName string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	bundleRoot := path.Clean("/" + bundleName)

	for _, file := range r.File {
		entryPath, err := normalizeArchiveEntryPath(file.Name)
		if err != nil {
			return err
		}

		outputPath, err := util.ResolveArchiveEntryPath(dest, file.Name)
		if err != nil {
			return fmt.Errorf("failed to resolve archive path: %w", err)
		}

		mode := file.Mode()
		switch {
		case mode&os.ModeSymlink != 0:
			if err := validateMacAppArchiveSymlink(file, entryPath, bundleRoot); err != nil {
				return err
			}

			targetPath, err := readZipSymlinkTarget(file)
			if err != nil {
				return err
			}

			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}
			if err := os.Symlink(targetPath, outputPath); err != nil {
				return fmt.Errorf("failed to create symlink %s: %w", file.Name, err)
			}
		case file.FileInfo().IsDir():
			perm := mode.Perm()
			if perm == 0 {
				perm = 0755
			}
			if err := os.MkdirAll(outputPath, perm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", file.Name, err)
			}
		case mode.IsRegular():
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", file.Name, err)
			}

			rc, err := file.Open()
			if err != nil {
				return fmt.Errorf("failed to open archive entry %s: %w", file.Name, err)
			}

			outFile, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
			if err != nil {
				_ = rc.Close()
				return fmt.Errorf("failed to create extracted file %s: %w", file.Name, err)
			}

			_, copyErr := io.Copy(outFile, rc)
			closeErr := outFile.Close()
			_ = rc.Close()
			if copyErr != nil {
				return fmt.Errorf("failed to extract file %s: %w", file.Name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("failed to finalize extracted file %s: %w", file.Name, closeErr)
			}
		default:
			return fmt.Errorf("%w: %s", util.ErrUnsupportedArchiveEntry, file.Name)
		}
	}

	return nil
}

type movePathFunc func(src string, dst string) error

func movePathWithCopyFallback(src string, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return err
	}

	if err := copyPathRecursive(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

func copyPathRecursive(src string, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return err
		}

		target, err := os.Readlink(src)
		if err != nil {
			return err
		}

		return os.Symlink(target, dst)
	}

	if info.IsDir() {
		if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
			return err
		}

		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if err := copyPathRecursive(filepath.Join(src, entry.Name()), filepath.Join(dst, entry.Name())); err != nil {
				return err
			}
		}

		return nil
	}

	return copyRegularFile(src, dst, info.Mode())
}

func copyRegularFile(src string, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return os.Chmod(dst, mode.Perm())
}

func installMacAppBundle(appPath string, applicationsPath string, movePath movePathFunc) error {
	if movePath == nil {
		movePath = movePathWithCopyFallback
	}

	backupPath := ""
	if _, err := os.Stat(applicationsPath); err == nil {
		backupPath = applicationsPath + ".backup-" + strconv.FormatInt(time.Now().UnixNano(), 10)
		if err := movePath(applicationsPath, backupPath); err != nil {
			return fmt.Errorf("failed to stage existing app: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect existing app: %w", err)
	}

	if err := movePath(appPath, applicationsPath); err != nil {
		if backupPath != "" {
			if restoreErr := movePath(backupPath, applicationsPath); restoreErr != nil {
				return fmt.Errorf("failed to move app to Applications: %w; additionally failed to restore previous app: %v", err, restoreErr)
			}
		}

		return fmt.Errorf("failed to move app to Applications: %w", err)
	}

	if backupPath != "" {
		_ = os.RemoveAll(backupPath)
	}

	return nil
}
