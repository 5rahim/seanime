package handlers

// HandleListMangaProviderExtensions
//
//	@summary returns the available manga providers.
//	@route /api/v1/extensions/list/manga-provider [GET]
//	@returns []extension_repo.MangaProviderExtensionItem
func HandleListMangaProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListMangaProviderExtensions()
	return c.RespondWithData(extensions)
}

// HandleListOnlinestreamProviderExtensions
//
//	@summary returns the available manga providers.
//	@route /api/v1/extensions/list/onlinestream-provider [GET]
//	@returns []extension_repo.OnlinestreamProviderExtensionItem
func HandleListOnlinestreamProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListOnlinestreamProviderExtensions()
	return c.RespondWithData(extensions)
}

// HandleListAnimeTorrentProviderExtensions
//
//	@summary returns the available manga providers.
//	@route /api/v1/extensions/list/anime-torrent-provider [GET]
//	@returns []extension_repo.AnimeTorrentProviderExtensionItem
func HandleListAnimeTorrentProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListAnimeTorrentProviderExtensions()
	return c.RespondWithData(extensions)
}
