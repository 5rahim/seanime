package extension_repo

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"io/fs"
	"os"
	"path/filepath"
	"seanime/internal/extension"
	"seanime/internal/util"
	"strings"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Loading external extensions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// LoadExternalExtensions loads all external extensions from the extension directory.
// This should be called after the built-in extensions are loaded.
func (r *Repository) LoadExternalExtensions() {
	r.logger.Trace().Msg("extension repo: Loading external extensions")

	//
	// Load external extensions
	//

	err := filepath.WalkDir(r.extensionDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			r.logger.Error().Err(err).Msg("extension repo: Failed to walk directory")
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Check if the file is a .json file
		// If it is, parse the json and install the extension
		if filepath.Ext(path) != ".json" {
			return nil
		}

		r.loadExternalExtension(path)

		return nil

	})
	if err != nil {
		r.logger.Error().Err(err).Msg("extension repo: Failed to load extensions")
		return
	}

	r.logger.Debug().Msg("extension repo: Loaded external extensions")
}

// Loads an external extension from a file path
func (r *Repository) loadExternalExtension(filePath string) {
	// Get the content of the file
	var ext extension.Extension
	// Parse the ext
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extension repo: Failed to read extension file")
		return
	}

	err = json.Unmarshal(fileContent, &ext)
	if err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extension repo: Failed to parse extension file")
		return
	}

	// Sanity check
	if ext.ID == "" || ext.Name == "" || ext.Version == "" || ext.Language == "" || ext.Type == "" || ext.Author == "" || ext.Payload == "" {
		r.logger.Error().Str("filepath", filePath).Msg("extension repo: Extension is missing required fields")
		return
	}

	// Check the ID
	if !r.isValidExtensionID(ext.ID) {
		r.logger.Error().Str("id", ext.ID).Msg("extension repo: Invalid extension ID")
		return
	}

	// Load extension
	switch ext.Type {
	case extension.TypeMangaProvider:
		// Load manga provider
		r.loadExternalMangaExtension(&ext)
	case extension.TypeOnlinestreamProvider:
		// Load online streaming provider
		r.loadExternalOnlinestreamProviderExtension(&ext)
	default:
		r.logger.Error().Str("type", string(ext.Type)).Msg("extension repo: Extension type not supported")
	}

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Manga
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalMangaExtension(ext *extension.Extension) {

	switch ext.Language {
	case extension.LanguageGo:
		r.loadExternalMangaExtensionGo(ext)
	case extension.LanguageJavascript:
		// TODO
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extension repo: Loaded manga provider extension")
}

//
// Go
//

func (r *Repository) loadExternalMangaExtensionGo(ext *extension.Extension) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Debug().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extension repo: Loading manga provider")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := r.yaegiInterp.Eval(payload)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load extension payload")
		return
	}

	// Get the provider
	newProviderFuncVal, err := r.yaegiInterp.Eval(extensionPackageName + `.NewProvider`)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load manga provider from extension")
		return
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikemanga.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg("extension repo: Failed to invoke provider constructor")
		return
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.mangaProviderExtensions.Set(ext.ID, extension.NewMangaProviderExtension(ext, provider))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Online streaming
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalOnlinestreamProviderExtension(ext *extension.Extension) {

	switch ext.Language {
	case extension.LanguageGo:
		r.loadExternalOnlinestreamProviderExtensionGo(ext)
	case extension.LanguageJavascript:
		// TODO
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extension repo: Loaded online streaming provider extension")
}

//
// Go
//

func (r *Repository) loadExternalOnlinestreamProviderExtensionGo(ext *extension.Extension) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Debug().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extension repo: Loading online streaming provider")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := r.yaegiInterp.Eval(payload)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load extension payload")
		return
	}

	// Get the provider
	newProviderFuncVal, err := r.yaegiInterp.Eval(extensionPackageName + `.NewProvider`)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load online streaming provider from extension")
		return
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibikeonlinestream.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg("extension repo: Failed to invoke provider constructor")
		return
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.onlinestreamProviderExtensions.Set(ext.ID, extension.NewOnlinestreamProviderExtension(ext, provider))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Torrent provider
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadExternalTorrentProviderExtension(ext *extension.Extension) {

	switch ext.Language {
	case extension.LanguageGo:
		r.loadExternalOnlinestreamProviderExtensionGo(ext)
	case extension.LanguageJavascript:
		// TODO
	}

	r.logger.Debug().Str("id", ext.ID).Msg("extension repo: Loaded online streaming provider extension")
}

//
// Go
//

func (r *Repository) loadExternalTorrentProviderExtensionGo(ext *extension.Extension) {

	extensionPackageName := "ext_" + util.GenerateCryptoID()

	r.logger.Debug().Str("id", ext.ID).Str("packageName", extensionPackageName).Msg("extension repo: Loading torrent provider")

	payload := strings.Replace(ext.Payload, "package main", "package "+extensionPackageName, 1)

	// Load the extension payload
	_, err := r.yaegiInterp.Eval(payload)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load extension payload")
		return
	}

	// Get the provider
	newProviderFuncVal, err := r.yaegiInterp.Eval(extensionPackageName + `.NewProvider`)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extension repo: Failed to load torrent provider from extension")
		return
	}

	newProviderFunc, ok := newProviderFuncVal.Interface().(func(logger *zerolog.Logger) hibiketorrent.Provider)
	if !ok {
		r.logger.Error().Str("id", ext.ID).Msg("extension repo: Failed to invoke provider constructor")
		return
	}

	provider := newProviderFunc(r.logger)

	// Add the extension to the map
	r.torrentProviderExtensions.Set(ext.ID, extension.NewTorrentProviderExtension(ext, provider))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
