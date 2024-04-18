package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/library/entities"
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
//	@returns entities.AutoDownloaderRule
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
//	@returns []entities.AutoDownloaderRule
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
//	@returns entities.AutoDownloaderRule
func HandleCreateAutoDownloaderRule(c *RouteCtx) error {
	rule := new(entities.AutoDownloaderRule)
	if err := c.Fiber.BodyParser(rule); err != nil {
		return c.RespondWithError(err)
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
//	@returns entities.AutoDownloaderRule
func HandleUpdateAutoDownloaderRule(c *RouteCtx) error {
	rule := new(entities.AutoDownloaderRule)
	if err := c.Fiber.BodyParser(rule); err != nil {
		return c.RespondWithError(err)
	}

	if rule.DbID == 0 {
		return c.RespondWithError(errors.New("invalid id"))
	}

	// Update the rule based on its DbID (primary key)
	if err := c.App.Database.UpdateAutoDownloaderRule(rule.DbID, rule); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rule)
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
