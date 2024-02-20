package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/entities"
	"strconv"
)

// HandleRunAutoDownloader will run the auto downloader.
// It returns true.
//
//	POST /v1/auto-downloader/run
func HandleRunAutoDownloader(c *RouteCtx) error {

	c.App.AutoDownloader.Run()

	return c.RespondWithData(true)
}

// HandleGetAutoDownloaderRule will return the rule with the given id (primary key).
//
//	GET /v1/auto-downloader/rule/:id
func HandleGetAutoDownloaderRule(c *RouteCtx) error {

	id, err := strconv.Atoi(c.Fiber.Params("id"))
	if err != nil {
		return c.RespondWithError(errors.New("invalid id"))
	}

	rule, err := c.App.Database.GetAutoDownloaderRule(uint(id))
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rule)
}

// HandleGetAutoDownloaderRules will return all rules.
//
//	GET	/v1/auto-downloader/rules
func HandleGetAutoDownloaderRules(c *RouteCtx) error {
	rules, err := c.App.Database.GetAutoDownloaderRules()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rules)
}

// HandleCreateAutoDownloaderRule will create a new rule.
// It returns the created rule.
//
//	POST /v1/auto-downloader/rule
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

// HandleUpdateAutoDownloaderRule will update the rule passed in the request body.
// It returns the updated rule.
//
//	POST /v1/auto-downloader/rule
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

// HandleDeleteAutoDownloaderRule will delete the rule with the given id (primary key).
// It returns true.
//
//	DELETE /v1/auto-downloader/rule/:id
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

// HandleGetAutoDownloaderItems will return all items.
//
//	GET	/v1/auto-downloader/items
func HandleGetAutoDownloaderItems(c *RouteCtx) error {
	rules, err := c.App.Database.GetAutoDownloaderItems()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rules)
}

// HandleDeleteAutoDownloaderItem will delete the item with the given id (primary key).
// It returns true.
//
//	DELETE /v1/auto-downloader/item
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
