package extension_repo

import (
	"context"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/util"
	"time"
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// External extensions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) FetchExternalExtensionData(manifestURI string) (*extension.Extension, error) {
	return r.fetchExternalExtensionData(manifestURI)
}

func (r *Repository) fetchExternalExtensionData(manifestURI string) (*extension.Extension, error) {

	// Fetch the manifest file
	client := &http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURI, nil)
	if err != nil {
		r.logger.Error().Err(err).Str("uri", manifestURI).Msg("extensions: Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create HTTP request, %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		r.logger.Error().Err(err).Str("uri", manifestURI).Msg("extensions: Failed to fetch extension manifest")
		return nil, fmt.Errorf("failed to fetch extension manifest, %w", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var ext extension.Extension
	err = json.NewDecoder(resp.Body).Decode(&ext)
	if err != nil {
		r.logger.Error().Err(err).Str("uri", manifestURI).Msg("extensions: Failed to parse extension manifest")
		return nil, fmt.Errorf("failed to parse extension manifest, %w", err)
	}

	//if err = r.extensionSanityCheck(&ext); err != nil {
	//	r.logger.Error().Err(err).Str("url", manifestURI).Msg("extensions: Failed sanity check")
	//	return nil, err
	//}

	return &ext, nil
}

type ExtensionInstallResponse struct {
	Message string `json:"message"`
}

func (r *Repository) InstallExternalExtension(manifestURI string) (*ExtensionInstallResponse, error) {

	ext, err := r.fetchExternalExtensionData(manifestURI)
	if err != nil {
		r.logger.Error().Err(err).Str("uri", manifestURI).Msg("extensions: Failed to fetch extension data")
		return nil, fmt.Errorf("failed to fetch extension data, %w", err)
	}

	filename := filepath.Join(r.extensionDir, ext.ID+".json")

	update := false

	// Check if the extension is already installed
	// i.e. a file with the same ID exists
	if _, err := os.Stat(filename); err == nil {
		r.logger.Debug().Str("id", ext.ID).Msg("extensions: Updating extension")
		// Delete the old extension
		err := os.Remove(filename)
		if err != nil {
			r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to remove old extension")
			return nil, fmt.Errorf("failed to remove old extension, %w", err)
		}
		update = true
	}

	// Add the extension as a json file
	file, err := os.Create(filename)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to create extension file")
		return nil, fmt.Errorf("failed to create extension file, %w", err)
	}
	defer file.Close()

	// Write the extension to the file
	enc := json.NewEncoder(file)
	err = enc.Encode(ext)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ext.ID).Msg("extensions: Failed to write extension to file")
		return nil, fmt.Errorf("failed to write extension to file, %w", err)
	}

	// Reload the extensions
	r.loadExternalExtensions()

	if update {
		return &ExtensionInstallResponse{
			Message: fmt.Sprintf("Successfully updated %s", ext.Name),
		}, nil
	}

	return &ExtensionInstallResponse{
		Message: fmt.Sprintf("Successfully installed %s", ext.Name),
	}, nil
}

func (r *Repository) UninstallExternalExtension(id string) error {

	// Check if the extension exists
	installedExt, found := r.GetLoadedExtension(id)
	if !found {
		r.logger.Error().Str("id", id).Msg("extensions: Extension not found")
		return fmt.Errorf("extension not found")
	}

	// Check if the extension is external
	if installedExt.GetManifestURI() == "builtin" {
		r.logger.Error().Str("id", id).Msg("extensions: Extension is built-in")
		return fmt.Errorf("extension is built-in")
	}

	// Uninstall the extension
	err := os.Remove(filepath.Join(r.extensionDir, id+".json"))
	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("extensions: Failed to uninstall extension")
		return fmt.Errorf("failed to uninstall extension, %w", err)
	}

	// Reload the extensions
	r.loadExternalExtensions()

	return nil
}

// CheckForUpdates checks all extensions for updates by querying their respective repositories
func (r *Repository) CheckForUpdates(manifestURI string) {

}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Loading/Reloading external extensions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ReloadExternalExtensions() {
	r.loadExternalExtensions()
}

// killGojaVMs kills all VMs from currently loaded Goja extensions & clears the Goja extensions map.
func (r *Repository) killGojaVMs() {
	defer util.HandlePanicInModuleThen("extension_repo/killGojaVMs", func() {})

	r.logger.Trace().Msg("extensions: Killing Goja VMs")

	r.gojaExtensions.Range(func(key string, ext GojaExtension) bool {
		defer util.HandlePanicInModuleThen(fmt.Sprintf("extension_repo/killGojaVMs/%s", key), func() {})

		ext.GetVM().ClearInterrupt()
		return true
	})

	// Clear the Goja extensions map
	r.gojaExtensions.Clear()

	r.logger.Debug().Msg("extensions: Killed Goja VMs")
}

// unloadExternalExtensions unloads all external extensions from the extension banks.
func (r *Repository) unloadExternalExtensions() {
	r.logger.Trace().Msg("extensions: Unloading external extensions")
	// We also clear the invalid extensions list, assuming the extensions are reloaded
	r.invalidExtensions.Clear()
	r.mangaProviderExtensionBank.RemoveExternalExtensions()
	r.animeTorrentProviderExtensionBank.RemoveExternalExtensions()
	r.onlinestreamProviderExtensionBank.RemoveExternalExtensions()

	r.logger.Debug().Msg("extensions: Unloaded external extensions")
}

// loadExternalExtensions loads all external extensions from the extension directory.
// This should be called after the built-in extensions are loaded.
func (r *Repository) loadExternalExtensions() {
	r.logger.Trace().Msg("extensions: Loading external extensions")

	// Kill all Goja VMs
	r.killGojaVMs()

	// Unload all external extensions
	r.unloadExternalExtensions()

	//
	// Load external extensions
	//

	err := filepath.WalkDir(r.extensionDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			r.logger.Error().Err(err).Msg("extensions: Failed to walk directory")
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
		r.logger.Error().Err(err).Msg("extensions: Failed to load extensions")
		return
	}

	r.logger.Debug().Msg("extensions: Loaded external extensions")

	r.wsEventManager.SendEvent(events.ExtensionsReloaded, nil)
}

// Loads an external extension from a file path
func (r *Repository) loadExternalExtension(filePath string) {
	// Get the content of the file
	var ext extension.Extension
	// Parse the ext
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extensions: Failed to read extension file")
		return
	}

	err = json.Unmarshal(fileContent, &ext)
	if err != nil {
		// If the extension file is corrupted or not a valid extension, skip loading the extension.
		// We don't add it to the InvalidExtensions list because there's not enough information to
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extensions: Failed to parse extension file")
		return
	}

	var manifestError error

	// Sanity check
	if err = r.extensionSanityCheck(&ext); err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extensions: Failed sanity check")
		manifestError = err
	}

	// If there was an error with the manifest, skip loading the extension,
	// add the extension to the InvalidExtensions list and return
	// The extension should be added to the InvalidExtensions list with an auto-generated ID.
	if manifestError != nil {
		id := uuid.NewString()
		r.invalidExtensions.Set(id, &extension.InvalidExtension{
			ID:        id,
			Reason:    manifestError.Error(),
			Path:      filePath,
			Code:      extension.InvalidExtensionManifestError,
			Extension: ext,
		})
		return
	}

	var loadingErr error

	// Load extension
	switch ext.Type {
	case extension.TypeMangaProvider:
		// Load manga provider
		loadingErr = r.loadExternalMangaExtension(&ext)
	case extension.TypeOnlinestreamProvider:
		// Load online streaming provider
		loadingErr = r.loadExternalOnlinestreamProviderExtension(&ext)
	case extension.TypeAnimeTorrentProvider:
		// Load torrent provider
		loadingErr = r.loadExternalAnimeTorrentProviderExtension(&ext)
	default:
		r.logger.Error().Str("type", string(ext.Type)).Msg("extensions: Extension type not supported")
		loadingErr = fmt.Errorf("extension type not supported")
	}

	// If there was an error loading the extension, skip adding it to the extension bank
	// and add the extension to the InvalidExtensions list
	if loadingErr != nil {
		id := uuid.NewString()
		r.invalidExtensions.Set(id, &extension.InvalidExtension{
			ID:        id,
			Reason:    loadingErr.Error(),
			Path:      filePath,
			Code:      extension.InvalidExtensionPayloadError,
			Extension: ext,
		})
		return
	}

	return
}
