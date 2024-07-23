package extension_repo

import (
	"github.com/goccy/go-json"
	"io/fs"
	"os"
	"path/filepath"
	"seanime/internal/extension"
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
	case extension.TypeTorrentProvider:
		// Load torrent provider
		r.loadExternalTorrentProviderExtension(&ext)
	default:
		r.logger.Error().Str("type", string(ext.Type)).Msg("extension repo: Extension type not supported")
	}

}
