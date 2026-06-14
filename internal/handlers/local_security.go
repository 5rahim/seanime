package handlers

import (
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"path/filepath"
	"runtime"
	"seanime/internal/database/models"
	"seanime/internal/security"
	"strings"

	"github.com/labstack/echo/v4"
)

// These helpers add baseline security for unauthenticated servers
// If the server password is not set, some actions like mutating certain settings will be denied unless the request comes from a trusted local origin (like denshi or localhost).
// This helps prevent some CSRF attacks from websites when running in passwordless mode

var errPrivilegedExecutionDenied = errors.New("this action requires either a server password or a trusted local origin")

var errStrictModeDenied = errors.New("this action is disabled when secure mode is strict")

var errStrictLocalOnlyDenied = errors.New("this action requires a trusted local origin when secure mode is strict")

var errStrictFilesystemPathDenied = errors.New("this path is not allowed when secure mode is strict")

var errGuardResponseWritten = errors.New("guard response written")

func respondWithAbort(c echo.Context, code int, err error) error {
	if c == nil {
		return err
	}

	if writeErr := c.JSON(code, NewErrorResponse(err)); writeErr != nil {
		return writeErr
	}

	return errGuardResponseWritten
}

func isStrictModeSensitive(req *http.Request, serverPassword string) bool {
	return serverPassword == "" && security.IsStrict() && !isRequestFromTrustedLocal(req)
}

func reqHasOriginMetadata(req *http.Request) bool {
	if req == nil {
		return false
	}

	return strings.TrimSpace(req.Header.Get("Origin")) != "" || strings.TrimSpace(req.Header.Get("Referer")) != ""
}

func isCrossSiteBrowserRequest(req *http.Request) bool {
	if req == nil {
		return false
	}

	return strings.EqualFold(strings.TrimSpace(req.Header.Get("Sec-Fetch-Site")), "cross-site")
}

func isPathNeedingTrustedLocalBoundary(path string) bool {
	return path == "/events" || strings.HasPrefix(path, "/api/")
}

func isHardenedTrustedRequestHost(req *http.Request) bool {
	view := createRequestBoundaryView(req)
	if view.hostname == "" {
		return false
	}

	host := view.hostname
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	return addr.IsLoopback()
}

func isTrustedHardenedOriginURL(parsed *url.URL) bool {
	if parsed == nil {
		return false
	}

	if parsed.Scheme == "app" && parsed.Host == "-" {
		return true
	}

	scheme := strings.ToLower(parsed.Scheme)
	isLocal := scheme == "capacitor" || scheme == "ionic" || scheme == "app" || scheme == "file"
	if isLocal {
		host := strings.ToLower(parsed.Hostname())
		return host == "" || host == "localhost" || host == "-"
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	return addr.IsLoopback()
}

func isRequestFromTrustedHardenedOrigin(req *http.Request) bool {
	if req == nil {
		return false
	}

	rawOrigin := strings.TrimSpace(req.Header.Get("Origin"))
	if rawOrigin == "" {
		rawOrigin = strings.TrimSpace(req.Header.Get("Referer"))
	}
	parsed, ok := parseTrustedOrigin(rawOrigin)
	if !ok {
		return false
	}

	return isTrustedHardenedOriginURL(parsed)
}

func hasHardenedLocalClientBoundary(req *http.Request) bool {
	if req == nil {
		return false
	}

	view := createRequestBoundaryView(req)
	if !view.clientIP.IsValid() || !view.clientIP.IsLoopback() {
		return false
	}

	if hasForwardedHeaders(req) && !view.trustedProxy {
		return false
	}

	return true
}

func isRequestFromTrustedHardenedLocal(req *http.Request) bool {
	if !hasHardenedLocalClientBoundary(req) {
		return false
	}

	if !isHardenedTrustedRequestHost(req) {
		return false
	}

	return isRequestFromTrustedHardenedOrigin(req)
}

// isTrustedRequestHost checks if the request originates from a trusted host such as localhost, a loopback, or private network address.
func isTrustedRequestHost(req *http.Request) bool {
	view := createRequestBoundaryView(req)
	if view.hostname == "" {
		return false
	}

	host := view.hostname
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	return addr.IsLoopback() || addr.IsPrivate() || isTailscaleIP(addr)
}

// isRequestPermitted determines if an HTTP request is permitted based on server password, access allowlist, and request origin metadata.
func isRequestPermitted(req *http.Request, serverPassword string, accessAllowlist []string) bool {
	if serverPassword != "" || security.IsLax() {
		return true
	}

	if security.IsHardened() {
		allowlistedHost := isAllowlistedRequestHost(req, accessAllowlist)
		if !isHardenedTrustedRequestHost(req) && !allowlistedHost {
			return false
		}

		if !reqHasOriginMetadata(req) {
			if isCrossSiteBrowserRequest(req) {
				return false
			}
			if allowlistedHost {
				return true
			}
			return hasHardenedLocalClientBoundary(req)
		}

		if isRequestFromAllowlistedOrigin(req, accessAllowlist) {
			return true
		}

		return isRequestFromTrustedHardenedLocal(req)
	}

	if !isTrustedRequestHost(req) && !isAllowlistedRequestHost(req, accessAllowlist) {
		return false
	}

	if !reqHasOriginMetadata(req) {
		if isCrossSiteBrowserRequest(req) {
			return false
		}
		return true
	}

	return isRequestFromTrustedOrigin(req) || isRequestFromAllowlistedOrigin(req, accessAllowlist)
}

// isTrustedCORSOrigin determines if the provided CORS origin is trusted based on server security settings and allowlist rules.
func isTrustedCORSOrigin(rawOrigin string, serverPassword string, accessAllowlist []string) bool {
	if serverPassword != "" || security.IsLax() {
		return true
	}

	parsed, ok := parseTrustedOrigin(rawOrigin)
	if !ok {
		return false
	}

	if parsed.Scheme == "app" && parsed.Host == "-" {
		return true
	}
	if isAllowlistedOrigin(parsed, accessAllowlist) {
		return true
	}
	if security.IsHardened() {
		return isTrustedHardenedOriginURL(parsed)
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	return addr.IsLoopback() || addr.IsPrivate() || isTailscaleIP(addr)
}

func (h *Handler) trustedLocalRequestMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if h == nil || h.App == nil || h.App.Config == nil {
			return next(c)
		}

		req := c.Request()
		if req == nil || req.URL == nil || !isPathNeedingTrustedLocalBoundary(req.URL.Path) {
			return next(c)
		}

		if isRequestPermitted(req, h.App.Config.Server.Password, h.App.Config.Server.AccessAllowlist) {
			return next(c)
		}

		return h.RespondWithStatusError(c, http.StatusForbidden, errPrivilegedExecutionDenied)
	}
}

// isTrustedRequest determines whether the request is from a trusted source based on server password, security mode, or request origin.
func isTrustedRequest(req *http.Request, serverPassword string) bool {
	if serverPassword != "" || security.IsLax() {
		return true
	}
	if security.IsHardened() {
		return isRequestFromTrustedHardenedLocal(req)
	}

	return isRequestFromTrustedOrigin(req)
}

// guardPrivilegedSettingsMutation checks and denies unauthorized privileged settings modifications, returning an error if access is forbidden.
func (h *Handler) guardPrivilegedSettingsMutation(c echo.Context, prev *models.Settings, nextMedia *models.MediaPlayerSettings, nextTorrent *models.TorrentSettings) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if canMutatePrivilegedSettings(c.Request(), h.App.Config.Server.Password, prev, nextMedia, nextTorrent) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedExtensionManagement checks if a request meets the criteria for privileged extension management access.
func (h *Handler) guardPrivilegedExtensionManagement(c echo.Context) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if security.IsStrict() && !isRequestFromTrustedLocal(c.Request()) {
		return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
	}

	if canUsePrivilegedExtensionManagement(c.Request(), h.App.Config.Server.Password) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedMediastreamSettingsMutation ensures that mutations to privileged mediastream settings are only allowed by trusted or authorized sources.
func (h *Handler) guardPrivilegedMediastreamSettingsMutation(c echo.Context, prev *models.MediastreamSettings, next *models.MediastreamSettings) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if canMutatePrivilegedMediastreamSettings(c.Request(), h.App.Config.Server.Password, prev, next) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// canMutatePrivilegedSettings determines if privileged settings modifications can proceed based on request origin, server password, and settings changes.
func canMutatePrivilegedSettings(req *http.Request, serverPassword string, prev *models.Settings, nextMedia *models.MediaPlayerSettings, nextTorrent *models.TorrentSettings) bool {
	if security.IsStrict() && !isRequestFromTrustedLocal(req) && privilegedSettingsChanged(prev, nextMedia, nextTorrent) {
		return false
	}

	if isTrustedRequest(req, serverPassword) {
		return true
	}

	if !privilegedSettingsChanged(prev, nextMedia, nextTorrent) {
		return true
	}

	return false
}

// canMutatePrivilegedMediastreamSettings determines if privileged mediastream settings can be modified based on request trust and setting changes.
func canMutatePrivilegedMediastreamSettings(req *http.Request, serverPassword string, prev *models.MediastreamSettings, next *models.MediastreamSettings) bool {
	if security.IsStrict() && !isRequestFromTrustedLocal(req) && privilegedMediastreamSettingsChanged(prev, next) {
		return false
	}

	if isTrustedRequest(req, serverPassword) {
		return true
	}

	if !privilegedMediastreamSettingsChanged(prev, next) {
		return true
	}

	return false
}

// canUsePrivilegedExtensionManagement determines if the request can access privileged extension management based on security mode, origin, and server password.
func canUsePrivilegedExtensionManagement(req *http.Request, serverPassword string) bool {
	if security.IsStrict() && !isRequestFromTrustedLocal(req) {
		return false
	}

	return isTrustedRequest(req, serverPassword)
}

func canConsumeMedia(req *http.Request, serverPassword string, accessAllowlist []string) bool {
	return isRequestPermitted(req, serverPassword, accessAllowlist)
}

func (h *Handler) guardMediaConsumption(c echo.Context) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if canConsumeMedia(c.Request(), h.App.Config.Server.Password, h.App.Config.Server.AccessAllowlist) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedMediaPlayer restricts access to privileged media player actions based on security settings and request origin validation.
func (h *Handler) guardPrivilegedMediaPlayer(c echo.Context, settings *models.Settings) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if isTrustedRequest(c.Request(), h.App.Config.Server.Password) || !isPrivilegedMediaPlayer(settings) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedTorrentClient ensures that only trusted or authorized requests can execute privileged torrent client actions.
func (h *Handler) guardPrivilegedTorrentClient(c echo.Context, settings *models.Settings) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if security.IsStrict() && usesExternalTorrentClient(settings) && !isRequestFromTrustedLocal(c.Request()) {
		return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
	}

	if isTrustedRequest(c.Request(), h.App.Config.Server.Password) || !isPrivilegedTorrentClient(settings) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedMediastream ensures that privileged mediastream actions are restricted to trusted requests or server password authorization.
func (h *Handler) guardPrivilegedMediastream(c echo.Context, settings *models.MediastreamSettings) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if isTrustedRequest(c.Request(), h.App.Config.Server.Password) || !isPrivilegedMediastream(settings) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// guardPrivilegedLocalExecution enforces security checks for privileged actions, ensuring the request originates from a trusted or authorized source.
func (h *Handler) guardPrivilegedLocalExecution(c echo.Context) error {
	if h == nil || h.App == nil || h.App.Config == nil {
		return nil
	}

	if security.IsStrict() && !isRequestFromTrustedLocal(c.Request()) {
		return respondWithAbort(c, http.StatusForbidden, errStrictLocalOnlyDenied)
	}

	if isTrustedRequest(c.Request(), h.App.Config.Server.Password) {
		return nil
	}

	return respondWithAbort(c, http.StatusForbidden, errPrivilegedExecutionDenied)
}

// getContextClientId retrieves the client ID from the echo.Context by checking a header or a cookie, returning an empty string if not found.
func getContextClientId(c echo.Context) string {
	if c == nil {
		return ""
	}

	if value := c.Get("Seanime-Client-Id"); value != nil {
		if clientID, ok := value.(string); ok {
			clientID = strings.TrimSpace(clientID)
			if clientID != "" {
				return clientID
			}
		}
	}

	cookie, err := c.Cookie(clientIdCookieName)
	if err == nil {
		if clientID := strings.TrimSpace(cookie.Value); clientID != "" {
			return clientID
		}
	}

	return ""
}

func getClientPlatformFromContext(c echo.Context) string {
	if c == nil {
		return ""
	}

	if value := c.Get(clientPlatformHeader); value != nil {
		if platform, ok := value.(string); ok {
			return normalizeClientPlatform(platform)
		}
	}

	return ""
}

// getRequestClientId retrieves the client ID from the context or falls back to the claimed value after trimming whitespace.
func getRequestClientId(c echo.Context, claimed string) string {
	if contextClientID := getContextClientId(c); contextClientID != "" {
		return contextClientID
	}

	return strings.TrimSpace(claimed)
}

// isSameContextClientId checks if the claimed client ID matches the context's client ID after trimming spaces and ensuring both are non-empty.
func isSameContextClientId(c echo.Context, claimed string) bool {
	contextClientId := getContextClientId(c)
	claimed = strings.TrimSpace(claimed)
	return contextClientId != "" && claimed != "" && contextClientId == claimed
}

type accessAllowlistEntry struct {
	scheme string
	host   string
	port   string
}

func parseAccessAllowlistEntry(raw string) (*accessAllowlistEntry, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}

	if strings.Contains(raw, "://") {
		parsed, ok := parseTrustedOrigin(raw)
		if !ok {
			return nil, false
		}

		return &accessAllowlistEntry{
			scheme: strings.ToLower(parsed.Scheme),
			host:   strings.ToLower(parsed.Hostname()),
			port:   getEffectivePort(parsed.Scheme, parsed.Port()),
		}, true
	}

	parsed, err := url.Parse("//" + raw)
	if err != nil || parsed.Hostname() == "" {
		return nil, false
	}

	return &accessAllowlistEntry{
		host: strings.ToLower(parsed.Hostname()),
		port: parsed.Port(),
	}, true
}

func parseTrustedOrigin(rawOrigin string) (*url.URL, bool) {
	rawOrigin = strings.TrimSpace(rawOrigin)
	if rawOrigin == "" {
		return nil, false
	}

	parsed, err := url.Parse(rawOrigin)
	if err != nil {
		return nil, false
	}

	if parsed.Scheme == "app" && parsed.Host == "-" {
		return parsed, true
	}

	scheme := strings.ToLower(parsed.Scheme)
	isLocal := scheme == "capacitor" || scheme == "ionic" || scheme == "app" || scheme == "file"
	if parsed.Scheme != "http" && parsed.Scheme != "https" && !isLocal {
		return nil, false
	}

	if parsed.Scheme != "file" && parsed.Hostname() == "" && parsed.Host != "-" {
		return nil, false
	}

	return parsed, true
}

func isRequestFromAllowlistedOrigin(req *http.Request, accessAllowlist []string) bool {
	if req == nil {
		return false
	}

	rawOrigin := strings.TrimSpace(req.Header.Get("Origin"))
	if rawOrigin == "" {
		rawOrigin = strings.TrimSpace(req.Header.Get("Referer"))
	}
	parsed, ok := parseTrustedOrigin(rawOrigin)
	if !ok {
		return false
	}

	return isAllowlistedOrigin(parsed, accessAllowlist)
}

func isRequestFromTrustedOrigin(req *http.Request) bool {
	if req == nil {
		return false
	}

	rawOrigin := strings.TrimSpace(req.Header.Get("Origin"))
	if rawOrigin == "" {
		rawOrigin = strings.TrimSpace(req.Header.Get("Referer"))
	}
	parsed, ok := parseTrustedOrigin(rawOrigin)
	if !ok {
		return false
	}

	if parsed.Scheme == "app" && parsed.Host == "-" {
		return true
	}

	scheme := strings.ToLower(parsed.Scheme)
	isLocal := scheme == "capacitor" || scheme == "ionic" || scheme == "app" || scheme == "file"
	if isLocal {
		host := strings.ToLower(parsed.Hostname())
		return host == "" || host == "localhost" || host == "-"
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" {
		return true
	}

	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}

	if addr.IsLoopback() {
		return true
	}

	if !addr.IsPrivate() && !isTailscaleIP(addr) {
		return false
	}

	return isReqSameLiteralHost(req, parsed)
}

// note: this rejects requests from hosted instances
func isRequestFromTrustedLocal(req *http.Request) bool {
	if req == nil {
		return false
	}

	view := createRequestBoundaryView(req)
	if !isTrustedLocalClient(view) {
		return false
	}

	if hasForwardedHeaders(req) && !view.trustedProxy {
		return false
	}

	if !isTrustedRequestHost(req) {
		return false
	}

	return isRequestFromTrustedOrigin(req)
}

func isTrustedLocalClient(view requestBoundaryView) bool {
	if !view.clientIP.IsValid() {
		return false
	}

	return view.clientIP.IsLoopback() || view.clientIP.IsPrivate() || isTailscaleIP(view.clientIP)
}

func hasForwardedHeaders(req *http.Request) bool {
	if req == nil {
		return false
	}

	return strings.TrimSpace(req.Header.Get("Forwarded")) != "" ||
		strings.TrimSpace(req.Header.Get("X-Forwarded-For")) != "" ||
		strings.TrimSpace(req.Header.Get("X-Forwarded-Host")) != "" ||
		strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")) != "" ||
		strings.TrimSpace(req.Header.Get("X-Real-IP")) != ""
}

func isAllowlistedRequestHost(req *http.Request, accessAllowlist []string) bool {
	view := createRequestBoundaryView(req)
	if view.hostname == "" {
		return false
	}

	return isAllowlistedHost(view.hostname, view.port, "", accessAllowlist)
}

func isAllowlistedOrigin(origin *url.URL, accessAllowlist []string) bool {
	if origin == nil {
		return false
	}

	return isAllowlistedHost(strings.ToLower(origin.Hostname()), getEffectivePort(origin.Scheme, origin.Port()), strings.ToLower(origin.Scheme), accessAllowlist)
}

func isAllowlistedHost(host string, port string, scheme string, accessAllowlist []string) bool {
	if host == "" {
		return false
	}

	for _, rawEntry := range accessAllowlist {
		entry, ok := parseAccessAllowlistEntry(rawEntry)
		if !ok {
			continue
		}
		if entry.scheme != "" && scheme != "" && entry.scheme != scheme {
			continue
		}
		if !isAllowlistHostMatch(entry.host, host) {
			continue
		}
		if entry.port != "" && entry.port != port && !(port == "" && (entry.port == "80" || entry.port == "443")) {
			continue
		}

		return true
	}

	return false
}

func isAllowlistHostMatch(pattern string, host string) bool {
	pattern = strings.ToLower(strings.TrimSpace(pattern))
	host = strings.ToLower(strings.TrimSpace(host))
	if pattern == "" || host == "" {
		return false
	}
	if pattern == host {
		return true
	}
	if !strings.HasPrefix(pattern, "*.") {
		return false
	}

	suffix := strings.TrimPrefix(pattern, "*.")
	return host != suffix && strings.HasSuffix(host, "."+suffix)
}

func getEffectivePort(scheme string, port string) string {
	if port != "" {
		return port
	}

	switch strings.ToLower(strings.TrimSpace(scheme)) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

func isReqSameLiteralHost(req *http.Request, origin *url.URL) bool {
	if req == nil || origin == nil {
		return false
	}

	view := createRequestBoundaryView(req)
	if view.hostname == "" {
		return false
	}

	if !strings.EqualFold(origin.Hostname(), view.hostname) {
		return false
	}

	return getEffectivePort(origin.Scheme, origin.Port()) == getEffectivePort(view.scheme, view.port)
}

func privilegedSettingsChanged(prev *models.Settings, nextMedia *models.MediaPlayerSettings, nextTorrent *models.TorrentSettings) bool {
	if nextMedia != nil {
		prevMedia := prev.GetMediaPlayer()
		if prevMedia.Default != nextMedia.Default ||
			prevMedia.VlcPath != nextMedia.VlcPath ||
			prevMedia.MpcPath != nextMedia.MpcPath ||
			prevMedia.MpvPath != nextMedia.MpvPath ||
			prevMedia.MpvArgs != nextMedia.MpvArgs ||
			prevMedia.IinaPath != nextMedia.IinaPath ||
			prevMedia.IinaArgs != nextMedia.IinaArgs ||
			translateEndpointChanged(prevMedia, nextMedia) {
			return true
		}
	}

	if nextTorrent != nil {
		prevTorrent := prev.GetTorrent()
		if prevTorrent.Default != nextTorrent.Default ||
			prevTorrent.QBittorrentPath != nextTorrent.QBittorrentPath ||
			prevTorrent.TransmissionPath != nextTorrent.TransmissionPath {
			return true
		}
	}

	return false
}

func translateEndpointChanged(prevMedia *models.MediaPlayerSettings, nextMedia *models.MediaPlayerSettings) bool {
	if prevMedia == nil || nextMedia == nil {
		return false
	}

	prevCompatible := strings.EqualFold(prevMedia.VcTranslateProvider, "openai-compatible")
	nextCompatible := strings.EqualFold(nextMedia.VcTranslateProvider, "openai-compatible")
	if prevCompatible != nextCompatible {
		return true
	}
	if prevCompatible || nextCompatible {
		return prevMedia.VcTranslate != nextMedia.VcTranslate ||
			prevMedia.VcTranslateBaseUrl != nextMedia.VcTranslateBaseUrl
	}

	return false
}

// privilegedMediastreamSettingsChanged checks if privileged mediastream settings differ between the previous and the next configuration.
func privilegedMediastreamSettingsChanged(prev *models.MediastreamSettings, next *models.MediastreamSettings) bool {
	if next == nil {
		return false
	}

	if prev == nil {
		return isPrivilegedMediastream(next)
	}

	return prev.FfmpegPath != next.FfmpegPath || prev.FfprobePath != next.FfprobePath
}

func isPrivilegedMediaPlayer(settings *models.Settings) bool {
	media := settings.GetMediaPlayer()

	switch media.Default {
	case "vlc":
		return hasCustomExecutablePath(media.VlcPath, defaultVLCPaths()...)
	case "mpc-hc":
		return hasCustomExecutablePath(media.MpcPath, defaultMpcHcPaths()...)
	case "mpv":
		return strings.TrimSpace(media.MpvArgs) != "" || hasCustomExecutablePath(media.MpvPath, "mpv")
	case "iina":
		return strings.TrimSpace(media.IinaArgs) != "" || hasCustomExecutablePath(media.IinaPath, "iina-cli")
	default:
		return false
	}
}

func isPrivilegedTorrentClient(settings *models.Settings) bool {
	torrent := settings.GetTorrent()

	switch torrent.Default {
	case "qbittorrent":
		return hasCustomExecutablePath(torrent.QBittorrentPath, defaultQBittorrentPaths()...)
	case "transmission":
		return hasCustomExecutablePath(torrent.TransmissionPath, defaultTransmissionPaths()...)
	default:
		return false
	}
}

func isPrivilegedMediastream(settings *models.MediastreamSettings) bool {
	if settings == nil {
		return false
	}

	return hasCustomExecutablePath(settings.FfmpegPath, defaultFFmpegPaths()...) || hasCustomExecutablePath(settings.FfprobePath, defaultFFprobePaths()...)
}

// hasCustomExecutablePath checks if the given path differs from the provided default executable paths. Returns true if a custom path is detected, false otherwise.
func hasCustomExecutablePath(path string, defaults ...string) bool {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false
	}

	for _, defaultPath := range defaults {
		if sameExecutablePath(trimmed, defaultPath) {
			return false
		}
	}

	return true
}

func sameExecutablePath(left string, right string) bool {
	leftPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(left)))
	rightPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(right)))

	if runtime.GOOS == "windows" {
		return strings.EqualFold(leftPath, rightPath)
	}

	return leftPath == rightPath
}

func defaultVLCPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"C:\\Program Files\\VideoLAN\\VLC\\vlc.exe"}
	case "linux":
		return []string{"/usr/bin/vlc"}
	case "darwin":
		return []string{"/Applications/VLC.app/Contents/MacOS/VLC"}
	default:
		return []string{"vlc"}
	}
}

func defaultMpcHcPaths() []string {
	return []string{"C:\\Program Files\\MPC-HC\\mpc-hc64.exe"}
}

func defaultQBittorrentPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"C:/Program Files/qBittorrent/qbittorrent.exe"}
	case "linux":
		return []string{"/usr/bin/qbittorrent"}
	case "darwin":
		return []string{"/Applications/qbittorrent.app/Contents/MacOS/qbittorrent"}
	default:
		return []string{"qbittorrent"}
	}
}

func defaultTransmissionPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"C:/Program Files/Transmission/transmission-qt.exe"}
	case "linux":
		return []string{"/usr/bin/transmission-qt", "/usr/bin/transmission-gtk"}
	case "darwin":
		return []string{
			"/Applications/Transmission.app/Contents/MacOS/transmission-qt",
			"/Applications/Transmission.app/Contents/MacOS/Transmission",
		}
	default:
		return []string{"transmission-qt"}
	}
}

func defaultFFmpegPaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"ffmpeg.exe", "ffmpeg"}
	case "linux":
		return []string{"/usr/bin/ffmpeg", "/usr/local/bin/ffmpeg", "ffmpeg"}
	case "darwin":
		return []string{"/opt/homebrew/bin/ffmpeg", "/usr/local/bin/ffmpeg", "ffmpeg"}
	default:
		return []string{"ffmpeg"}
	}
}

func defaultFFprobePaths() []string {
	switch runtime.GOOS {
	case "windows":
		return []string{"ffprobe.exe", "ffprobe"}
	case "linux":
		return []string{"/usr/bin/ffprobe", "/usr/local/bin/ffprobe", "ffprobe"}
	case "darwin":
		return []string{"/opt/homebrew/bin/ffprobe", "/usr/local/bin/ffprobe", "ffprobe"}
	default:
		return []string{"ffprobe"}
	}
}

func isTailscaleIP(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	if addr.Is4() {
		ip := addr.As4()
		return ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127
	}
	if addr.Is6() {
		ip := addr.As16()
		return ip[0] == 0xfd && ip[1] == 0x7a &&
			ip[2] == 0x11 && ip[3] == 0x5c &&
			ip[4] == 0xa1 && ip[5] == 0xe0
	}
	return false
}
