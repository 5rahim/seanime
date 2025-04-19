package extension_repo

import (
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Built-in extensions
// - Built-in extensions are loaded once, on application startup
// - The "manifestURI" field is set to "builtin" to indicate that the extension is not external
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ReloadBuiltInExtension(ext extension.Extension, provider interface{}) {
	r.reloadBuiltInExtension(ext, provider)
}

func (r *Repository) reloadBuiltInExtension(ext extension.Extension, provider interface{}) {

	// Unload the extension
	// Remove extension from bank
	r.extensionBank.Delete(ext.ID)

	// Kill Goja VM if it exists
	gojaExtension, ok := r.gojaExtensions.Get(ext.ID)
	if ok {
		// Interrupt the extension's runtime and running processed before unloading
		gojaExtension.ClearInterrupt()
		r.logger.Trace().Str("id", ext.ID).Msg("extensions: Killed built-in extension's runtime")
		r.gojaExtensions.Delete(ext.ID)
	}
	// Remove from invalid extensions
	r.invalidExtensions.Delete(ext.ID)

	// Load the extension
	r.loadBuiltInExtension(ext, provider)
}

func saveUserConfigInProvider(ext *extension.Extension, provider interface{}) {
	if provider == nil {
		return
	}

	if ext.SavedUserConfig == nil {
		return
	}

	if configurableProvider, ok := provider.(extension.Configurable); ok {
		configurableProvider.SetSavedUserConfig(*ext.SavedUserConfig)
	}
}

func (r *Repository) loadBuiltInExtension(ext extension.Extension, provider interface{}) {

	r.builtinExtensions.Set(ext.ID, &builtinExtension{
		Extension: ext,
		provider:  provider,
	})

	// Load user config in the struct
	configErr := r.loadUserConfig(&ext)
	if configErr != nil {
		r.invalidExtensions.Set(ext.ID, &extension.InvalidExtension{
			ID:        ext.ID,
			Reason:    configErr.Error(),
			Path:      "",
			Code:      extension.InvalidExtensionUserConfigError,
			Extension: ext,
		})
		r.logger.Warn().Err(configErr).Str("id", ext.ID).Msg("extensions: Failed to load user config")
	}

	switch ext.Type {
	case extension.TypeMangaProvider:
		switch ext.Language {
		// Go
		case extension.LanguageGo:
			if provider == nil {
				r.logger.Error().Str("id", ext.ID).Msg("extensions: Built-in manga provider extension requires a provider")
				return
			}
			saveUserConfigInProvider(&ext, provider)
			if mangaProvider, ok := provider.(hibikemanga.Provider); ok {
				r.loadBuiltInMangaProviderExtension(ext, mangaProvider)
			}
		}
	case extension.TypeAnimeTorrentProvider:
		switch ext.Language {
		// Go
		case extension.LanguageGo:
			if provider == nil {
				r.logger.Error().Str("id", ext.ID).Msg("extensions: Built-in anime torrent provider extension requires a provider")
				return
			}
			saveUserConfigInProvider(&ext, provider)
			if animeProvider, ok := provider.(hibiketorrent.AnimeProvider); ok {
				r.loadBuiltInAnimeTorrentProviderExtension(ext, animeProvider)
			}
		}
	case extension.TypeOnlinestreamProvider:
		switch ext.Language {
		// Go
		case extension.LanguageGo:
			if provider == nil {
				r.logger.Error().Str("id", ext.ID).Msg("extensions: Built-in onlinestream provider extension requires a provider")
				return
			}
			saveUserConfigInProvider(&ext, provider)
			if onlinestreamProvider, ok := provider.(hibikeonlinestream.Provider); ok {
				r.loadBuiltInOnlinestreamProviderExtension(ext, onlinestreamProvider)
			}
		case extension.LanguageJavascript, extension.LanguageTypescript:
			r.loadBuiltInOnlinestreamProviderExtensionJS(ext)
		}
	case extension.TypePlugin:
		// TODO: Implement
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded built-in extension")
	r.wsEventManager.SendEvent(events.ExtensionsReloaded, nil)
}

func (r *Repository) loadBuiltInMangaProviderExtension(ext extension.Extension, provider hibikemanga.Provider) {
	r.extensionBank.Set(ext.ID, extension.NewMangaProviderExtension(&ext, provider))
	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded built-in manga provider extension")
}

func (r *Repository) loadBuiltInAnimeTorrentProviderExtension(ext extension.Extension, provider hibiketorrent.AnimeProvider) {
	r.extensionBank.Set(ext.ID, extension.NewAnimeTorrentProviderExtension(&ext, provider))
	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded built-in anime torrent provider extension")
}

func (r *Repository) loadBuiltInOnlinestreamProviderExtension(ext extension.Extension, provider hibikeonlinestream.Provider) {
	r.extensionBank.Set(ext.ID, extension.NewOnlinestreamProviderExtension(&ext, provider))
	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

func (r *Repository) loadBuiltInOnlinestreamProviderExtensionJS(ext extension.Extension) {
	// Load the extension as if it was an external extension
	err := r.loadExternalOnlinestreamExtensionJS(&ext, ext.Language)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to load built-in JS onlinestream provider extension")
		return
	}
	r.logger.Debug().Str("id", ext.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}
