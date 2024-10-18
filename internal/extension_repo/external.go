package extension_repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/util"
	"sync"
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

	// Check manifest
	if err = manifestSanityCheck(&ext); err != nil {
		r.logger.Error().Err(err).Str("uri", manifestURI).Msg("extensions: Failed sanity check")
		return nil, fmt.Errorf("failed sanity check, %w", err)
	}

	return &ext, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
	//r.loadExternalExtensions()

	r.reloadExtension(ext.ID)

	if update {
		return &ExtensionInstallResponse{
			Message: fmt.Sprintf("Successfully updated %s", ext.Name),
		}, nil
	}

	return &ExtensionInstallResponse{
		Message: fmt.Sprintf("Successfully installed %s", ext.Name),
	}, nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
	//r.loadExternalExtensions()

	go func() {
		_ = r.deleteExtensionUserConfig(id)
	}()

	r.reloadExtension(id)

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// checkForUpdates checks all extensions for updates by querying their respective repositories.
// It returns a list of extension update data containing IDs and versions.
func (r *Repository) checkForUpdates() (ret []UpdateData) {

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}

	r.logger.Trace().Msg("extensions: Checking for updates")

	// Check for updates for all extensions
	r.extensionBank.Range(func(key string, ext extension.BaseExtension) bool {
		wg.Add(1)
		go func(ext extension.BaseExtension) {
			defer wg.Done()

			if ext.GetManifestURI() == "builtin" || ext.GetManifestURI() == "" {
				return
			}
			// Check for updates
			extFromRepo, err := r.fetchExternalExtensionData(ext.GetManifestURI())
			if err != nil {
				r.logger.Error().Err(err).Str("id", ext.GetID()).Str("url", ext.GetManifestURI()).Msg("extensions: Failed to fetch extension data while checking for update")
				return
			}

			// Sanity check, this checks for the version too
			if err = manifestSanityCheck(extFromRepo); err != nil {
				r.logger.Error().Err(err).Str("id", ext.GetID()).Str("url", ext.GetManifestURI()).Msg("extensions: Failed sanity check while checking for update")
				return
			}

			// If there's an update, send the update data to the channel
			if extFromRepo.Version != ext.GetVersion() {
				updateData := UpdateData{
					ExtensionID: extFromRepo.ID,
					Version:     extFromRepo.Version,
					ManifestURI: extFromRepo.ManifestURI,
				}
				mu.Lock()
				ret = append(ret, updateData)
				mu.Unlock()
			}
		}(ext)
		return true
	})

	wg.Wait()

	r.logger.Debug().Int("haveUpdates", len(ret)).Msg("extensions: Retrieved update info")

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// UpdateExtensionCode updates the code of an external application
func (r *Repository) UpdateExtensionCode(id string, payload string) error {

	if id == "" {
		r.logger.Error().Msg("extensions: ID is empty")
		return fmt.Errorf("id is empty")
	}

	if payload == "" {
		r.logger.Error().Msg("extensions: Payload is empty")
		return fmt.Errorf("payload is empty")
	}

	// We don't check if the extension existed in "loaded" extensions since the extension might be invalid
	// We check if the file exists

	filename := id + ".json"
	extensionFilepath := filepath.Join(r.extensionDir, filename)

	if _, err := os.Stat(extensionFilepath); err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("extensions: Extension not found")
		return fmt.Errorf("extension not found")
	}

	ext, err := extractExtensionFromFile(extensionFilepath)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("extensions: Failed to read extension file")
		return fmt.Errorf("failed to read extension file, %w", err)
	}

	// Update the payload
	ext.Payload = payload

	// Write the extension to the file
	file, err := os.Create(extensionFilepath)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("extensions: Failed to create extension file")
		return fmt.Errorf("failed to create extension file, %w", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	err = enc.Encode(ext)

	if err != nil {
		r.logger.Error().Err(err).Str("id", id).Msg("extensions: Failed to write extension to file")
		return fmt.Errorf("failed to write extension to file, %w", err)
	}

	// Reload the extensions
	//r.loadExternalExtensions()

	r.reloadExtension(id)

	return nil
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
	r.extensionBank.RemoveExternalExtensions()

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
	// Parse the ext
	ext, err := extractExtensionFromFile(filePath)
	if err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extensions: Failed to read extension file")
		return
	}

	ext.Lang = extension.GetExtensionLang(ext.Lang)

	var manifestError error

	// Sanity check
	if err = r.extensionSanityCheck(ext); err != nil {
		r.logger.Error().Err(err).Str("filepath", filePath).Msg("extensions: Failed sanity check")
		manifestError = err
	}

	invalidExtensionID := ext.ID
	if invalidExtensionID == "" {
		invalidExtensionID = uuid.NewString()
	}

	// If there was an error with the manifest, skip loading the extension,
	// add the extension to the InvalidExtensions list and return
	// The extension should be added to the InvalidExtensions list with an auto-generated ID.
	if manifestError != nil {
		r.invalidExtensions.Set(invalidExtensionID, &extension.InvalidExtension{
			ID:        invalidExtensionID,
			Reason:    manifestError.Error(),
			Path:      filePath,
			Code:      extension.InvalidExtensionManifestError,
			Extension: *ext,
		})
		return
	}

	var loadingErr error

	// Load user config
	configErr := r.loadUserConfig(ext)

	// If there was an error loading the user config, we add it to the InvalidExtensions list
	// BUT we still load the extension
	// DEVNOTE: Failure to load the user config is not a critical error
	if configErr != nil {
		r.invalidExtensions.Set(invalidExtensionID, &extension.InvalidExtension{
			ID:        invalidExtensionID,
			Reason:    configErr.Error(),
			Path:      filePath,
			Code:      extension.InvalidExtensionUserConfigError,
			Extension: *ext,
		})
	}

	// Load extension
	switch ext.Type {
	case extension.TypeMangaProvider:
		// Load manga provider
		loadingErr = r.loadExternalMangaExtension(ext)
	case extension.TypeOnlinestreamProvider:
		// Load online streaming provider
		loadingErr = r.loadExternalOnlinestreamProviderExtension(ext)
	case extension.TypeAnimeTorrentProvider:
		// Load torrent provider
		loadingErr = r.loadExternalAnimeTorrentProviderExtension(ext)
	default:
		r.logger.Error().Str("type", string(ext.Type)).Msg("extensions: Extension type not supported")
		loadingErr = fmt.Errorf("extension type not supported")
	}

	// If there was an error loading the extension, skip adding it to the extension bank
	// and add the extension to the InvalidExtensions list
	if loadingErr != nil {
		r.invalidExtensions.Set(invalidExtensionID, &extension.InvalidExtension{
			ID:        invalidExtensionID,
			Reason:    loadingErr.Error(),
			Path:      filePath,
			Code:      extension.InvalidExtensionPayloadError,
			Extension: *ext,
		})
		return
	}

	return
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Reload specific extension
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// ReloadExtension reloads a specific extension by ID.
func (r *Repository) ReloadExtension(id string) error {
	r.reloadExtension(id)
	return nil
}

func (r *Repository) reloadExtension(id string) {
	r.logger.Trace().Str("id", id).Msg("extensions: Reloading extension")

	// Remove pointers to the extension
	// Remove extension from bank
	r.extensionBank.Delete(id)
	// Kill Goja VM if it exists
	r.gojaExtensions.Range(func(key string, ext GojaExtension) bool {
		defer util.HandlePanicInModuleThen(fmt.Sprintf("extension_repo/reloadExtension/%s", key), func() {})
		if key != id {
			return true
		}
		ext.GetVM().ClearInterrupt()
		r.logger.Trace().Str("id", id).Msg("extensions: Killed extension JS VM")
		return false
	})
	r.gojaExtensions.Delete(id)
	// Remove from invalid extensions
	r.invalidExtensions.Delete(id)
	//r.invalidExtensions.Range(func(key string, ext *extension.InvalidExtension) bool {
	//	if ext.Extension.ID == id {
	//		r.invalidExtensions.Delete(key)
	//		return false
	//	}
	//	return true
	//})

	// Load the extension from the file
	extensionFilepath := filepath.Join(r.extensionDir, id+".json")

	// Check if the extension still exists
	if _, err := os.Stat(extensionFilepath); err != nil {
		// If the extension doesn't exist anymore, return silently - it was uninstalled

		r.wsEventManager.SendEvent(events.ExtensionsReloaded, nil)
		r.logger.Debug().Str("id", id).Msg("extensions: Extension removed")
		return
	}

	// If the extension still exist, load it back
	r.loadExternalExtension(extensionFilepath)

	r.logger.Debug().Str("id", id).Msg("extensions: Reloaded extension")
	r.wsEventManager.SendEvent(events.ExtensionsReloaded, nil)

	return
}
