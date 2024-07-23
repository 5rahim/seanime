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

// HandleListTorrentProviderExtensions
//
//	@summary returns the available manga providers.
//	@route /api/v1/extensions/list/torrent-provider [GET]
//	@returns []extension_repo.TorrentProviderExtensionItem
func HandleListTorrentProviderExtensions(c *RouteCtx) error {
	extensions := c.App.ExtensionRepository.ListTorrentProviderExtensions()
	return c.RespondWithData(extensions)
}
