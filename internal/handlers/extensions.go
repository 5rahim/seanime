package handlers

// HandleFetchExternalExtensionData
//
//	@summary returns the extension data from the given manifest uri.
//	@route /api/v1/extensions/external/fetch [POST]
//	@returns extension.Extension
func HandleFetchExternalExtensionData(c *RouteCtx) error {
	type body struct {
		ManifestURI string `json:"manifestUri"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	extension, err := c.App.ExtensionRepository.FetchExternalExtensionData(b.ManifestURI)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(extension)
}

// HandleInstallExternalExtension
//
//	@summary installs the extension from the given manifest uri.
//	@route /api/v1/extensions/external/install [POST]
//	@returns extension_repo.ExtensionInstallResponse
func HandleInstallExternalExtension(c *RouteCtx) error {
	type body struct {
		ManifestURI string `json:"manifestUri"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	res, err := c.App.ExtensionRepository.InstallExternalExtension(b.ManifestURI)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(res)
}

// HandleUninstallExternalExtension
//
//	@summary uninstalls the extension with the given ID.
//	@route /api/v1/extensions/external/uninstall [POST]
//	@returns bool
func HandleUninstallExternalExtension(c *RouteCtx) error {
	type body struct {
		ID string `json:"id"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	err := c.App.ExtensionRepository.UninstallExternalExtension(b.ID)
	if err != nil {
		return c.RespondWithError(err)
	}

	return c.RespondWithData(true)
}

// HandleReloadExternalExtensions
//
//	@summary reloads the external extensions.
//	@route /api/v1/extensions/external/reload [POST]
//	@returns bool
func HandleReloadExternalExtensions(c *RouteCtx) error {
	c.App.ExtensionRepository.ReloadExternalExtensions()
	return c.RespondWithData(true)
}

// HandleListExtensionData
//
//	@summary returns the loaded extensions
//	@route /api/v1/extensions/list [GET]
//	@returns []extension.Extension
func HandleListExtensionData(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListExtensionData()
	return c.RespondWithData(extensions)
}

// HandleGetAllExtensions
//
//	@summary returns all loaded and invalid extensions.
//	@route /api/v1/extensions/all [POST]
//	@returns extension_repo.AllExtensions
func HandleGetAllExtensions(c *RouteCtx) error {
	type body struct {
		WithUpdates bool `json:"withUpdates"`
	}

	var b body
	if err := c.Fiber.BodyParser(&b); err != nil {
		return c.RespondWithError(err)
	}

	extensions := c.App.ExtensionRepository.GetAllExtensions(b.WithUpdates)
	return c.RespondWithData(extensions)
}

// HandleListMangaProviderExtensions
//
//	@summary returns the installed manga providers.
//	@route /api/v1/extensions/list/manga-provider [GET]
//	@returns []extension_repo.MangaProviderExtensionItem
func HandleListMangaProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListMangaProviderExtensions()
	return c.RespondWithData(extensions)
}

// HandleListOnlinestreamProviderExtensions
//
//	@summary returns the installed online streaming providers.
//	@route /api/v1/extensions/list/onlinestream-provider [GET]
//	@returns []extension_repo.OnlinestreamProviderExtensionItem
func HandleListOnlinestreamProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListOnlinestreamProviderExtensions()
	return c.RespondWithData(extensions)
}

// HandleListAnimeTorrentProviderExtensions
//
//	@summary returns the installed torrent providers.
//	@route /api/v1/extensions/list/anime-torrent-provider [GET]
//	@returns []extension_repo.AnimeTorrentProviderExtensionItem
func HandleListAnimeTorrentProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListAnimeTorrentProviderExtensions()
	return c.RespondWithData(extensions)
}
