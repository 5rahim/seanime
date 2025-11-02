package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"seanime/internal/updater"
	"seanime/internal/util/result"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

// HandleInstallLatestUpdate
//
//	@summary installs the latest update.
//	@desc This will install the latest update and launch the new version.
//	@route /api/v1/install-update [POST]
//	@returns handlers.Status
func (h *Handler) HandleInstallLatestUpdate(c echo.Context) error {
	type body struct {
		FallbackDestination string `json:"fallback_destination"`
	}
	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		h.App.SelfUpdater.StartSelfUpdate(b.FallbackDestination)
	}()

	status := h.NewStatus(c)
	status.Updating = true

	time.Sleep(1 * time.Second)

	return h.RespondWithData(c, status)
}

// HandleGetLatestUpdate
//
//	@summary returns the latest update.
//	@desc This will return the latest update.
//	@desc If an error occurs, it will return an empty update.
//	@route /api/v1/latest-update [GET]
//	@returns updater.Update
func (h *Handler) HandleGetLatestUpdate(c echo.Context) error {
	update, err := h.App.Updater.GetLatestUpdate()
	if err != nil {
		return h.RespondWithData(c, &updater.Update{})
	}

	return h.RespondWithData(c, update)
}

type changelogItem struct {
	Version string   `json:"version"`
	Lines   []string `json:"lines"`
}

var changelogCache = result.NewCache[string, []*changelogItem]()

// HandleGetChangelog
//
//	@summary returns the changelog for versions greater than or equal to the given version.
//	@route /api/v1/changelog [GET]
//	@param before query string true "The version to get the changelog for."
//	@returns string
func (h *Handler) HandleGetChangelog(c echo.Context) error {
	before := c.QueryParam("before")
	after := c.QueryParam("after")

	key := fmt.Sprintf("%s-%s", before, after)

	cached, ok := changelogCache.Get(key)
	if ok {
		return h.RespondWithData(c, cached)
	}

	changelogBody, err := http.Get("https://raw.githubusercontent.com/5rahim/seanime/main/CHANGELOG.md")
	if err != nil {
		return h.RespondWithData(c, []*changelogItem{})
	}
	defer changelogBody.Body.Close()

	changelog := []*changelogItem{}

	scanner := bufio.NewScanner(changelogBody.Body)

	var version string
	var body []string
	var blockOpen bool

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			if blockOpen {
				changelog = append(changelog, &changelogItem{
					Version: version,
					Lines:   body,
				})
			}

			version = strings.TrimPrefix(line, "## ")
			version = strings.TrimLeft(version, "v")
			body = []string{}
			blockOpen = true
		} else if blockOpen {
			if strings.TrimSpace(line) == "" {
				continue
			}

			body = append(body, line)
		}
	}

	if blockOpen {
		changelog = append(changelog, &changelogItem{
			Version: version,
			Lines:   body,
		})
	}

	// e.g. get changelog after 2.7.0
	if after != "" {
		changelog = lo.Filter(changelog, func(item *changelogItem, index int) bool {
			afterVersion, err := semver.NewVersion(after)
			if err != nil {
				return false
			}

			version, err := semver.NewVersion(item.Version)
			if err != nil {
				return false
			}

			return version.GreaterThan(afterVersion)
		})
	}

	// e.g. get changelog before 2.7.0
	if before != "" {
		changelog = lo.Filter(changelog, func(item *changelogItem, index int) bool {
			beforeVersion, err := semver.NewVersion(before)
			if err != nil {
				return false
			}

			version, err := semver.NewVersion(item.Version)
			if err != nil {
				return false
			}

			return version.LessThan(beforeVersion)
		})
	}

	changelogCache.Set(key, changelog)

	return h.RespondWithData(c, changelog)
}
