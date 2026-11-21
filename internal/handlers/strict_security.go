package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/security"
	"seanime/internal/util"
	"slices"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
)

// guardStrictLocalOnlyAction ensures an action is restricted to trusted local requests when in strict security mode, returning an error if denied.
func (h *Handler) guardStrictLocalOnlyAction(c echo.Context) error {
	if !security.IsStrict() || c == nil || isRequestFromTrustedLocal(c.Request()) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
}

// guardStrictFilesystemPath validates if a filesystem path is allowed under strict security mode based on origin and root constraints.
func (h *Handler) guardStrictFilesystemPath(c echo.Context, rawPath string) error {
	if !security.IsStrict() || c == nil {
		return nil
	}

	if isRequestFromTrustedLocal(c.Request()) {
		return nil
	}

	if isStrictFsPathAllowed(rawPath, h.strictFilesystemRoots()) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errStrictFilesystemPathDenied)
}

func (h *Handler) guardStrictSettingsMutation(c echo.Context, prev *models.Settings, nextLibrary *models.LibrarySettings, nextManga *models.MangaSettings) error {
	if !security.IsStrict() || c == nil || isRequestFromTrustedLocal(c.Request()) {
		return nil
	}

	if !libraryRootsChanged(prev.GetLibrary(), nextLibrary) && !mangaSourceChanged(prev.GetManga(), nextManga) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
}

func (h *Handler) guardStrictMediastreamRootMutation(c echo.Context, prev *models.MediastreamSettings, next *models.MediastreamSettings) error {
	if !security.IsStrict() || c == nil || isRequestFromTrustedLocal(c.Request()) {
		return nil
	}

	if !mediastreamRootsChanged(prev, next) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
}

func (h *Handler) guardStrictTorrentstreamRootMutation(c echo.Context, prev *models.TorrentstreamSettings, next *models.TorrentstreamSettings) error {
	if !security.IsStrict() || c == nil || isRequestFromTrustedLocal(c.Request()) {
		return nil
	}

	if !torrentstreamRootsChanged(prev, next) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
}

func (h *Handler) strictFilesystemRoots() []string {
	if h == nil || h.App == nil {
		return nil
	}

	roots := make([]string, 0, 8)

	if h.App.Settings != nil {
		for _, root := range h.App.Settings.GetLibrary().GetLibraryPaths() {
			roots = appendStrictRoot(roots, root)
		}
		roots = appendStrictRoot(roots, h.App.Settings.GetManga().LocalSourceDirectory)
	}

	if h.App.SecondarySettings.Torrentstream != nil {
		roots = appendStrictRoot(roots, h.App.SecondarySettings.Torrentstream.DownloadDir)
	}

	if h.App.SecondarySettings.Mediastream != nil {
		roots = appendStrictRoot(roots, h.App.SecondarySettings.Mediastream.PreTranscodeLibraryDir)
	}

	if downloadsDir, ok := userDownloadsDir(); ok {
		roots = appendStrictRoot(roots, downloadsDir)
	}

	return roots
}

func isStrictFsPathAllowed(rawPath string, roots []string) bool {
	path := normalizeStrictPath(rawPath)
	if path == "" {
		return false
	}

	for _, root := range roots {
		if root == "" {
			continue
		}
		if util.IsSameDir(root, path) || util.IsSubdirectory(root, path) {
			return true
		}
	}

	return false
}

func appendStrictRoot(roots []string, rawPath string) []string {
	path := normalizeStrictPath(rawPath)
	if path == "" {
		return roots
	}

	for _, existing := range roots {
		if util.IsSameDir(existing, path) {
			return roots
		}
	}

	return append(roots, path)
}

func normalizeStrictPath(rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return ""
	}

	absPath, err := filepath.Abs(filepath.Clean(trimmed))
	if err != nil {
		return ""
	}

	return filepath.Clean(absPath)
}

func userDownloadsDir() (string, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(homeDir) == "" {
		return "", false
	}

	return filepath.Join(homeDir, "Downloads"), true
}

func libraryRootsChanged(prev *models.LibrarySettings, next *models.LibrarySettings) bool {
	var prevPaths []string
	var nextPaths []string
	if prev != nil {
		prevPaths = prev.GetLibraryPaths()
	}
	if next != nil {
		nextPaths = next.GetLibraryPaths()
	}

	return !slices.Equal(normalizedDirectoryList(prevPaths), normalizedDirectoryList(nextPaths))
}

func mangaSourceChanged(prev *models.MangaSettings, next *models.MangaSettings) bool {
	prevPath := ""
	nextPath := ""
	if prev != nil {
		prevPath = prev.LocalSourceDirectory
	}
	if next != nil {
		nextPath = next.LocalSourceDirectory
	}

	if normalizeStrictPath(prevPath) != normalizeStrictPath(nextPath) {
		return true
	}
	prevURL, nextURL := "", ""
	if prev != nil {
		prevURL = prev.SuwayomiURL
	}
	if next != nil {
		nextURL = next.SuwayomiURL
	}
	return prevURL != nextURL
}

func mediastreamRootsChanged(prev *models.MediastreamSettings, next *models.MediastreamSettings) bool {
	prevPath := ""
	nextPath := ""
	if prev != nil {
		prevPath = prev.PreTranscodeLibraryDir
	}
	if next != nil {
		nextPath = next.PreTranscodeLibraryDir
	}

	return normalizeStrictPath(prevPath) != normalizeStrictPath(nextPath)
}

func torrentstreamRootsChanged(prev *models.TorrentstreamSettings, next *models.TorrentstreamSettings) bool {
	prevPath := ""
	nextPath := ""
	if prev != nil {
		prevPath = prev.DownloadDir
	}
	if next != nil {
		nextPath = next.DownloadDir
	}

	return normalizeStrictPath(prevPath) != normalizeStrictPath(nextPath)
}

func usesExternalMediaPlayer(settings *models.Settings) bool {
	switch settings.GetMediaPlayer().Default {
	case "vlc", "mpc-hc", "mpv", "iina":
		return true
	default:
		return false
	}
}

func usesExternalTorrentClient(settings *models.Settings) bool {
	switch settings.GetTorrent().Default {
	case "qbittorrent", "transmission":
		return true
	default:
		return false
	}
}

func normalizedDirectoryList(paths []string) []string {
	ret := make([]string, 0, len(paths))
	for _, path := range paths {
		normalized := normalizeStrictPath(path)
		if normalized == "" {
			continue
		}
		ret = append(ret, normalized)
	}

	sort.Strings(ret)
	ret = slices.Compact(ret)
	return ret
}
