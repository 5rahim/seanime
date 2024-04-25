package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/library/anime"
	"strconv"
)

// HandleRunAutoDownloader
//
//	@summary tells the AutoDownloader to check for new episodes if enabled.
//	@desc This will run the AutoDownloader if it is enabled.
//	@desc It does nothing if the AutoDownloader is disabled.
//	@route /api/v1/auto-downloader/run [POST]
//	@returns bool
func HandleRunAutoDownloader(c *RouteCtx) error {

	c.App.AutoDownloader.Run()

	return c.RespondWithData(true)
}

// HandleGetAutoDownloaderRule
//
//	@summary returns the rule with the given DB id.
//	@desc This is used to get a specific rule, useful for editing.
//	@route /api/v1/auto-downloader/rule/{id} [GET]
//	@param id - int - true - "The DB id of the rule"
//	@returns anime.AutoDownloaderRule
func HandleGetAutoDownloaderRule(c *RouteCtx) error {

	id, err := c.Fiber.ParamsInt("id")
	if err != nil {
		return c.RespondWithError(errors.New("invalid id"))
	}

	rule, err := c.App.Database.GetAutoDownloaderRule(uint(id))
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rule)
}

// HandleGetAutoDownloaderRules
//
//	@summary returns all rules.
//	@desc This is used to list all rules. It returns an empty slice if there are no rules.
//	@route /api/v1/auto-downloader/rules [GET]
//	@returns []anime.AutoDownloaderRule
func HandleGetAutoDownloaderRules(c *RouteCtx) error {
	rules, err := c.App.Database.GetAutoDownloaderRules()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rules)
}

// HandleCreateAutoDownloaderRule
//
//	@summary creates a new rule.
//	@desc The body should contain the same fields as entities.AutoDownloaderRule.
//	@desc It returns the created rule.
//	@route /api/v1/auto-downloader/rule [POST]
//	@returns anime.AutoDownloaderRule
func HandleCreateAutoDownloaderRule(c *RouteCtx) error {
	type body struct {
		Enabled             bool                                        `json:"enabled"`
		MediaId             int                                         `json:"mediaId"`
		ReleaseGroups       []string                                    `json:"releaseGroups"`
		Resolutions         []string                                    `json:"resolutions"`
		ComparisonTitle     string                                      `json:"comparisonTitle"`
		TitleComparisonType anime.AutoDownloaderRuleTitleComparisonType `json:"titleComparisonType"`
		EpisodeType         anime.AutoDownloaderRuleEpisodeType         `json:"episodeType"`
		EpisodeNumbers      []int                                       `json:"episodeNumbers,omitempty"`
		Destination         string                                      `json:"destination"`
	}

	var b body

	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
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
	}

	if err := c.App.Database.InsertAutoDownloaderRule(rule); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rule)
}

// HandleUpdateAutoDownloaderRule
//
//	@summary updates a rule.
//	@desc The body should contain the same fields as entities.AutoDownloaderRule.
//	@desc It returns the updated rule.
//	@route /api/v1/auto-downloader/rule [PATCH]
//	@returns anime.AutoDownloaderRule
func HandleUpdateAutoDownloaderRule(c *RouteCtx) error {

	type body struct {
		Rule *anime.AutoDownloaderRule `json:"rule"`
	}

	var b body

	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if b.Rule == nil {
		return c.RespondWithError(errors.New("invalid rule"))
	}

	if b.Rule.DbID == 0 {
		return c.RespondWithError(errors.New("invalid id"))
	}

	// Update the rule based on its DbID (primary key)
	if err := c.App.Database.UpdateAutoDownloaderRule(b.Rule.DbID, b.Rule); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(b.Rule)
}

// HandleDeleteAutoDownloaderRule
//
//	@summary deletes a rule.
//	@desc It returns 'true' if the rule was deleted.
//	@route /api/v1/auto-downloader/rule/{id} [DELETE]
//	@param id - int - true - "The DB id of the rule"
//	@returns bool
func HandleDeleteAutoDownloaderRule(c *RouteCtx) error {
	id, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(errors.New("invalid id"))
	}

	if err := c.App.Database.DeleteAutoDownloaderRule(uint(id)); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetAutoDownloaderItems
//
//	@summary returns all queued items.
//	@desc Queued items are episodes that are downloaded but not scanned or not yet downloaded.
//	@desc The AutoDownloader uses these items in order to not download the same episode twice.
//	@route /api/v1/auto-downloader/items [GET]
//	@returns []models.AutoDownloaderItem
func HandleGetAutoDownloaderItems(c *RouteCtx) error {
	rules, err := c.App.Database.GetAutoDownloaderItems()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rules)
}

// HandleDeleteAutoDownloaderItem
//
//	@summary delete a queued item.
//	@desc This is used to remove a queued item from the list.
//	@desc Returns 'true' if the item was deleted.
//	@route /api/v1/auto-downloader/item [DELETE]
//	@param id - int - true - "The DB id of the item"
//	@returns bool
func HandleDeleteAutoDownloaderItem(c *RouteCtx) error {

	type body struct {
		ID uint `json:"id"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	if err := c.App.Database.DeleteAutoDownloaderItem(b.ID); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}
