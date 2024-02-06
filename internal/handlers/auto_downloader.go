package handlers

import (
	"errors"
	"github.com/seanime-app/seanime/internal/entities"
	"strconv"
)

//v1.Get("/auto-downloader/rule/:id", makeHandler(app, HandleGetAutoDownloaderRule))

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

//v1.Get("/auto-downloader/rules", makeHandler(app, HandleGetAutoDownloaderRules))

func HandleGetAutoDownloaderRules(c *RouteCtx) error {
	rules, err := c.App.Database.GetAutoDownloaderRules()
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rules)
}

//v1.Post("/auto-downloader/rule", makeHandler(app, HandleCreateAutoDownloaderRule))

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

//v1.Patch("/auto-downloader/rule", makeHandler(app, HandleUpdateAutoDownloaderRule))

func HandleUpdateAutoDownloaderRule(c *RouteCtx) error {
	rule := new(entities.AutoDownloaderRule)
	if err := c.Fiber.BodyParser(rule); err != nil {
		return c.RespondWithError(err)
	}

	if err := c.App.Database.UpdateAutoDownloaderRule(rule.DbID, rule); err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(rule)
}

//v1.Delete("/auto-downloader/rule/:id", makeHandler(app, HandleDeleteAutoDownloaderRule))

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
