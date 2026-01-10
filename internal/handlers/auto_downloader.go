package handlers

import (
	"errors"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"strconv"

	"github.com/labstack/echo/v4"
)

// HandleRunAutoDownloader
//
//	@summary tells the AutoDownloader to check for new episodes if enabled.
//	@desc This will run the AutoDownloader if it is enabled.
//	@desc It does nothing if the AutoDownloader is disabled.
//	@route /api/v1/auto-downloader/run [POST]
//	@returns bool
func (h *Handler) HandleRunAutoDownloader(c echo.Context) error {

	h.App.AutoDownloader.Run()

	return h.RespondWithData(c, true)
}

// HandleGetAutoDownloaderRule
//
//	@summary returns the rule with the given DB id.
//	@desc This is used to get a specific rule, useful for editing.
//	@route /api/v1/auto-downloader/rule/{id} [GET]
//	@param id - int - true - "The DB id of the rule"
//	@returns anime.AutoDownloaderRule
func (h *Handler) HandleGetAutoDownloaderRule(c echo.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, errors.New("invalid id"))
	}

	rule, err := db_bridge.GetAutoDownloaderRule(h.App.Database, uint(id))
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, rule)
}

// HandleGetAutoDownloaderRulesByAnime
//
//	@summary returns the rules with the given media id.
//	@route /api/v1/auto-downloader/rule/anime/{id} [GET]
//	@param id - int - true - "The AniList anime id of the rules"
//	@returns []anime.AutoDownloaderRule
func (h *Handler) HandleGetAutoDownloaderRulesByAnime(c echo.Context) error {

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, errors.New("invalid id"))
	}

	rules := db_bridge.GetAutoDownloaderRulesByMediaId(h.App.Database, id)
	return h.RespondWithData(c, rules)
}

// HandleGetAutoDownloaderRules
//
//	@summary returns all rules.
//	@desc This is used to list all rules. It returns an empty slice if there are no rules.
//	@route /api/v1/auto-downloader/rules [GET]
//	@returns []anime.AutoDownloaderRule
func (h *Handler) HandleGetAutoDownloaderRules(c echo.Context) error {
	rules, err := db_bridge.GetAutoDownloaderRules(h.App.Database)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, rules)
}

// HandleCreateAutoDownloaderRule
//
//	@summary creates a new rule.
//	@desc The body should contain the same fields as entities.AutoDownloaderRule.
//	@desc It returns the created rule.
//	@route /api/v1/auto-downloader/rule [POST]
//	@returns anime.AutoDownloaderRule
func (h *Handler) HandleCreateAutoDownloaderRule(c echo.Context) error {
	type body struct {
		Enabled             bool                                        `json:"enabled"`
		MediaId             int                                         `json:"mediaId"`
		ReleaseGroups       []string                                    `json:"releaseGroups"`
		Resolutions         []string                                    `json:"resolutions"`
		AdditionalTerms     []string                                    `json:"additionalTerms"`
		ComparisonTitle     string                                      `json:"comparisonTitle"`
		TitleComparisonType anime.AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType"`
		EpisodeType         anime.AutoDownloaderRuleEpisodeType         `json:"episodeType"`
		EpisodeNumbers      []int                                       `json:"episodeNumbers,omitempty"`
		Destination         string                                      `json:"destination"`
	}

	var b body

	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Destination == "" {
		return h.RespondWithError(c, errors.New("destination is required"))
	}

	if !filepath.IsAbs(b.Destination) {
		return h.RespondWithError(c, errors.New("destination must be an absolute path"))
	}

	rule := &anime.AutoDownloaderRule{
		Enabled:             b.Enabled,
		MediaId:             b.MediaId,
		ReleaseGroups:       b.ReleaseGroups,
		Resolutions:         b.Resolutions,
		ComparisonTitle:     b.ComparisonTitle,
		TitleComparisonType: b.TitleComparisonType,
		EpisodeType:         b.EpisodeType,
		EpisodeNumbers:      b.EpisodeNumbers,
		Destination:         b.Destination,
		AdditionalTerms:     b.AdditionalTerms,
	}

	if err := db_bridge.InsertAutoDownloaderRule(h.App.Database, rule); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, rule)
}

// HandleUpdateAutoDownloaderRule
//
//	@summary updates a rule.
//	@desc The body should contain the same fields as entities.AutoDownloaderRule.
//	@desc It returns the updated rule.
//	@route /api/v1/auto-downloader/rule [PATCH]
//	@returns anime.AutoDownloaderRule
func (h *Handler) HandleUpdateAutoDownloaderRule(c echo.Context) error {

	type body struct {
		Rule *anime.AutoDownloaderRule `json:"rule"`
	}

	var b body

	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if b.Rule == nil {
		return h.RespondWithError(c, errors.New("invalid rule"))
	}

	if b.Rule.DbID == 0 {
		return h.RespondWithError(c, errors.New("invalid id"))
	}

	// Update the rule based on its DbID (primary key)
	if err := db_bridge.UpdateAutoDownloaderRule(h.App.Database, b.Rule.DbID, b.Rule); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, b.Rule)
}

// HandleDeleteAutoDownloaderRule
//
//	@summary deletes a rule.
//	@desc It returns 'true' if the rule was deleted.
//	@route /api/v1/auto-downloader/rule/{id} [DELETE]
//	@param id - int - true - "The DB id of the rule"
//	@returns bool
func (h *Handler) HandleDeleteAutoDownloaderRule(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return h.RespondWithError(c, errors.New("invalid id"))
	}

	// -1 deletes all no longer airing
	if id == -1 {
		animeCollection, err := h.App.GetAnimeCollection(false)
		if err != nil {
			return h.RespondWithError(c, err)
		}
		rules, err := db_bridge.GetAutoDownloaderRules(h.App.Database)
		if err != nil {
			return h.RespondWithError(c, err)
		}
		for _, rule := range rules {
			media, ok := animeCollection.FindAnime(rule.MediaId)
			if !ok {
				continue
			}
			if media.Status != nil && *media.Status == anilist.MediaStatusFinished {
				_ = db_bridge.DeleteAutoDownloaderRule(h.App.Database, rule.DbID)
			}
		}
		return h.RespondWithData(c, true)
	}

	if err := db_bridge.DeleteAutoDownloaderRule(h.App.Database, uint(id)); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetAutoDownloaderItems
//
//	@summary returns all queued items.
//	@desc Queued items are episodes that are downloaded but not scanned or not yet downloaded.
//	@desc The AutoDownloader uses these items in order to not download the same episode twice.
//	@route /api/v1/auto-downloader/items [GET]
//	@returns []models.AutoDownloaderItem
func (h *Handler) HandleGetAutoDownloaderItems(c echo.Context) error {
	rules, err := h.App.Database.GetAutoDownloaderItems()
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, rules)
}

// HandleDeleteAutoDownloaderItem
//
//	@summary delete a queued item.
//	@desc This is used to remove a queued item from the list.
//	@desc Returns 'true' if the item was deleted.
//	@route /api/v1/auto-downloader/item [DELETE]
//	@param id - int - true - "The DB id of the item"
//	@returns bool
func (h *Handler) HandleDeleteAutoDownloaderItem(c echo.Context) error {

	type body struct {
		ID uint `json:"id"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	if err := h.App.Database.DeleteAutoDownloaderItem(b.ID); err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
