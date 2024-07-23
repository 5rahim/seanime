package handlers

// HandleGetMangaProviderExtensions
//
//	@summary returns the available manga providers.
//	@route /api/v1/manga/provider-extensions [GET]
//	@returns []extension.MangaProviderExtensionItem
func HandleGetMangaProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListMangaProviderExtensions()
	return c.RespondWithData(extensions)
}
