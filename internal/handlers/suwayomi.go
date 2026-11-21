package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

var suwayomiImagePath = regexp.MustCompile(`^/api/v1/manga/[0-9]+/(thumbnail|chapter/[0-9]+/page/[0-9]+)$`)

func validateSuwayomiImageURL(baseURL, imageURL string) (*url.URL, error) {
	base, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil || base == nil || (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" {
		return nil, errors.New("invalid Suwayomi server URL")
	}
	target, err := url.Parse(imageURL)
	if err != nil || target == nil || target.User != nil || target.Scheme != base.Scheme || target.Host != base.Host || target.RawQuery != "" || target.Fragment != "" {
		return nil, errors.New("invalid Suwayomi image URL")
	}
	prefix := strings.TrimRight(base.Path, "/")
	if !strings.HasPrefix(target.Path, prefix+"/") || !suwayomiImagePath.MatchString(strings.TrimPrefix(target.Path, prefix)) {
		return nil, errors.New("invalid Suwayomi image path")
	}
	return target, nil
}

// HandleGetSuwayomiImage only forwards image paths on the configured server.
func (h *Handler) HandleGetSuwayomiImage(c echo.Context) error {
	settings, err := h.App.Database.GetSettings()
	if err != nil || settings == nil {
		return h.RespondWithError(c, errors.New("Suwayomi is not configured"))
	}
	target, err := validateSuwayomiImageURL(settings.GetManga().SuwayomiURL, c.QueryParam("url"))
	if err != nil {
		return h.RespondWithStatusError(c, http.StatusBadRequest, err)
	}
	client := &http.Client{Timeout: 45 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, target.String(), nil)
	if err != nil {
		return h.RespondWithError(c, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return h.RespondWithStatusError(c, http.StatusBadGateway, errors.New("could not load Suwayomi image"))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return h.RespondWithStatusError(c, http.StatusBadGateway, errors.New("Suwayomi could not load the image"))
	}
	return c.Stream(http.StatusOK, resp.Header.Get("Content-Type"), resp.Body)
}
