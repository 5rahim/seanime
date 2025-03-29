package handlers

import (
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/extension_playground"

	"github.com/labstack/echo/v4"
)

// HandleFetchExternalExtensionData
//
//	@summary returns the extension data from the given manifest uri.
//	@route /api/v1/extensions/external/fetch [POST]
//	@returns extension.Extension
func (h *Handler) HandleFetchExternalExtensionData(c echo.Context) error {
	type body struct {
		ManifestURI string `json:"manifestUri"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	extension, err := h.App.ExtensionRepository.FetchExternalExtensionData(b.ManifestURI)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, extension)
}

// HandleInstallExternalExtension
//
//	@summary installs the extension from the given manifest uri.
//	@route /api/v1/extensions/external/install [POST]
//	@returns extension_repo.ExtensionInstallResponse
func (h *Handler) HandleInstallExternalExtension(c echo.Context) error {
	type body struct {
		ManifestURI string `json:"manifestUri"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	res, err := h.App.ExtensionRepository.InstallExternalExtension(b.ManifestURI)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, res)
}

// HandleUninstallExternalExtension
//
//	@summary uninstalls the extension with the given ID.
//	@route /api/v1/extensions/external/uninstall [POST]
//	@returns bool
func (h *Handler) HandleUninstallExternalExtension(c echo.Context) error {
	type body struct {
		ID string `json:"id"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.ExtensionRepository.UninstallExternalExtension(b.ID)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleUpdateExtensionCode
//
//	@summary updates the extension code with the given ID and reloads the extensions.
//	@route /api/v1/extensions/external/edit-payload [POST]
//	@returns bool
func (h *Handler) HandleUpdateExtensionCode(c echo.Context) error {
	type body struct {
		ID      string `json:"id"`
		Payload string `json:"payload"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	err := h.App.ExtensionRepository.UpdateExtensionCode(b.ID, b.Payload)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}

// HandleReloadExternalExtensions
//
//	@summary reloads the external extensions.
//	@route /api/v1/extensions/external/reload [POST]
//	@returns bool
func (h *Handler) HandleReloadExternalExtensions(c echo.Context) error {
	h.App.ExtensionRepository.ReloadExternalExtensions()
	return h.RespondWithData(c, true)
}

// HandleReloadExternalExtension
//
//	@summary reloads the external extension with the given ID.
//	@route /api/v1/extensions/external/reload [POST]
//	@returns bool
func (h *Handler) HandleReloadExternalExtension(c echo.Context) error {
	type body struct {
		ID string `json:"id"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}
	h.App.ExtensionRepository.ReloadExternalExtension(b.ID)
	return h.RespondWithData(c, true)
}

// HandleListExtensionData
//
//	@summary returns the loaded extensions
//	@route /api/v1/extensions/list [GET]
//	@returns []extension.Extension
func (h *Handler) HandleListExtensionData(c echo.Context) error {
	extensions := h.App.ExtensionRepository.ListExtensionData()
	return h.RespondWithData(c, extensions)
}

// HandleGetExtensionPayload
//
//	@summary returns the payload of the extension with the given ID.
//	@route /api/v1/extensions/payload/{id} [GET]
//	@returns string
func (h *Handler) HandleGetExtensionPayload(c echo.Context) error {
	payload := h.App.ExtensionRepository.GetExtensionPayload(c.Param("id"))
	return h.RespondWithData(c, payload)
}

// HandleListDevelopmentModeExtensions
//
//	@summary returns the development mode extensions
//	@route /api/v1/extensions/list/development [GET]
//	@returns []extension.Extension
func (h *Handler) HandleListDevelopmentModeExtensions(c echo.Context) error {
	extensions := h.App.ExtensionRepository.ListDevelopmentModeExtensions()
	return h.RespondWithData(c, extensions)
}

// HandleGetAllExtensions
//
//	@summary returns all loaded and invalid extensions.
//	@route /api/v1/extensions/all [POST]
//	@returns extension_repo.AllExtensions
func (h *Handler) HandleGetAllExtensions(c echo.Context) error {
	type body struct {
		WithUpdates bool `json:"withUpdates"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	extensions := h.App.ExtensionRepository.GetAllExtensions(b.WithUpdates)
	return h.RespondWithData(c, extensions)
}

// HandleListMangaProviderExtensions
//
//	@summary returns the installed manga providers.
//	@route /api/v1/extensions/list/manga-provider [GET]
//	@returns []extension_repo.MangaProviderExtensionItem
func (h *Handler) HandleListMangaProviderExtensions(c echo.Context) error {
	extensions := h.App.ExtensionRepository.ListMangaProviderExtensions()
	return h.RespondWithData(c, extensions)
}

// HandleListOnlinestreamProviderExtensions
//
//	@summary returns the installed online streaming providers.
//	@route /api/v1/extensions/list/onlinestream-provider [GET]
//	@returns []extension_repo.OnlinestreamProviderExtensionItem
func (h *Handler) HandleListOnlinestreamProviderExtensions(c echo.Context) error {
	extensions := h.App.ExtensionRepository.ListOnlinestreamProviderExtensions()
	return h.RespondWithData(c, extensions)
}

// HandleListAnimeTorrentProviderExtensions
//
//	@summary returns the installed torrent providers.
//	@route /api/v1/extensions/list/anime-torrent-provider [GET]
//	@returns []extension_repo.AnimeTorrentProviderExtensionItem
func (h *Handler) HandleListAnimeTorrentProviderExtensions(c echo.Context) error {
	extensions := h.App.ExtensionRepository.ListAnimeTorrentProviderExtensions()
	return h.RespondWithData(c, extensions)
}

// HandleGetPluginSettings
//
//	@summary returns the plugin settings.
//	@route /api/v1/extensions/plugin-settings [GET]
//	@returns extension_repo.StoredPluginSettingsData
func (h *Handler) HandleGetPluginSettings(c echo.Context) error {
	settings := h.App.ExtensionRepository.GetPluginSettings()
	return h.RespondWithData(c, settings)
}

// HandleSetPluginSettingsPinnedTrays
//
//	@summary sets the pinned trays in the plugin settings.
//	@route /api/v1/extensions/plugin-settings/pinned-trays [POST]
//	@returns bool
func (h *Handler) HandleSetPluginSettingsPinnedTrays(c echo.Context) error {
	type body struct {
		PinnedTrayPluginIds []string `json:"pinnedTrayPluginIds"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.ExtensionRepository.SetPluginSettingsPinnedTrays(b.PinnedTrayPluginIds)
	return h.RespondWithData(c, true)
}

// HandleGrantPluginPermissions
//
//	@summary grants the plugin permissions to the extension with the given ID.
//	@route /api/v1/extensions/plugin-permissions/grant [POST]
//	@returns bool
func (h *Handler) HandleGrantPluginPermissions(c echo.Context) error {
	type body struct {
		ID string `json:"id"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	h.App.ExtensionRepository.GrantPluginPermissions(b.ID)
	return h.RespondWithData(c, true)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleRunExtensionPlaygroundCode
//
//	@summary runs the code in the extension playground.
//	@desc Returns the logs
//	@route /api/v1/extensions/playground/run [POST]
//	@returns extension_playground.RunPlaygroundCodeResponse
func (h *Handler) HandleRunExtensionPlaygroundCode(c echo.Context) error {
	type body struct {
		Params *extension_playground.RunPlaygroundCodeParams `json:"params"`
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	res, err := h.App.ExtensionPlaygroundRepository.RunPlaygroundCode(b.Params)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, res)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// HandleGetExtensionUserConfig
//
//	@summary returns the user config definition and current values for the extension with the given ID.
//	@route /api/v1/extensions/user-config/{id} [GET]
//	@returns extension_repo.ExtensionUserConfig
func (h *Handler) HandleGetExtensionUserConfig(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return h.RespondWithError(c, fmt.Errorf("id is required"))
	}
	config := h.App.ExtensionRepository.GetExtensionUserConfig(id)
	return h.RespondWithData(c, config)
}

// HandleSaveExtensionUserConfig
//
//	@summary saves the user config for the extension with the given ID and reloads it.
//	@route /api/v1/extensions/user-config [POST]
//	@returns bool
func (h *Handler) HandleSaveExtensionUserConfig(c echo.Context) error {
	type body struct {
		ID      string            `json:"id"`      // The extension ID
		Version int               `json:"version"` // The current extension user config definition version
		Values  map[string]string `json:"values"`  // The values
	}

	var b body
	if err := c.Bind(&b); err != nil {
		return h.RespondWithError(c, err)
	}

	config := &extension.SavedUserConfig{
		Version: b.Version,
		Values:  b.Values,
	}

	err := h.App.ExtensionRepository.SaveExtensionUserConfig(b.ID, config)
	if err != nil {
		return h.RespondWithError(c, err)
	}

	return h.RespondWithData(c, true)
}
