package extension_repo

import (
	"fmt"
	"seanime/internal/extension"
	"seanime/internal/plugin"
	"seanime/internal/util"
	"seanime/internal/util/filecache"
	"strings"
)

func getExtensionUserConfigBucketKey(extId string) string {
	return fmt.Sprintf("ext_user_config_%s", extId)
}

var (
	ErrMissingUserConfig      = fmt.Errorf("extension: user config is missing")
	ErrIncompatibleUserConfig = fmt.Errorf("extension: user config is incompatible")
)

// loadUserConfig loads the user config for the given extension by getting it from the cache and modifying the payload.
// This should be called before loading the extension.
// If the user config is absent OR the current user config is outdated, it will return an error.
// When an error is returned, the extension will not be loaded and the user will be prompted to update the extension on the frontend.
func (r *Repository) loadUserConfig(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleThen("extension_repo/loadUserConfig", func() {
		err = nil
	})

	// If the extension doesn't define a user config, skip this step
	if ext.UserConfig == nil {
		return nil
	}

	bucket := filecache.NewPermanentBucket(getExtensionUserConfigBucketKey(ext.ID))

	// Get the user config from the cache
	var savedConfig extension.SavedUserConfig
	found, _ := r.fileCacher.GetPerm(bucket, ext.ID, &savedConfig)

	// No user config found but the extension requires it
	if !found && ext.UserConfig.RequiresConfig {
		return ErrMissingUserConfig
	}

	// If the user config is outdated, return an error
	if found && savedConfig.Version != ext.UserConfig.Version {
		return ErrIncompatibleUserConfig
	}

	if found {
		// Replace the placeholders in the payload with the saved values
		for _, field := range ext.UserConfig.Fields {
			savedValue, found := savedConfig.Values[field.Name]
			if !found {
				ext.Payload = strings.ReplaceAll(ext.Payload, fmt.Sprintf("{{%s}}", field.Name), field.Default)
			} else {
				ext.Payload = strings.ReplaceAll(ext.Payload, fmt.Sprintf("{{%s}}", field.Name), savedValue)
			}
		}
		return nil
	} else {
		// If the user config is missing but isn't required, replace the placeholders with the default values
		for _, field := range ext.UserConfig.Fields {
			ext.Payload = strings.ReplaceAll(ext.Payload, fmt.Sprintf("{{%s}}", field.Name), field.Default)
		}
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type ExtensionUserConfig struct {
	UserConfig      *extension.UserConfig      `json:"userConfig"`
	SavedUserConfig *extension.SavedUserConfig `json:"savedUserConfig"`
}

func (r *Repository) GetExtensionUserConfig(id string) (ret *ExtensionUserConfig) {
	ret = &ExtensionUserConfig{
		UserConfig:      nil,
		SavedUserConfig: nil,
	}

	defer util.HandlePanicInModuleThen("extension_repo/GetExtensionUserConfig", func() {})

	ext, found := r.extensionBank.Get(id)
	if !found {
		return
	}

	ret.UserConfig = ext.GetUserConfig()

	bucket := filecache.NewPermanentBucket(getExtensionUserConfigBucketKey(id))

	var savedConfig extension.SavedUserConfig
	found, _ = r.fileCacher.GetPerm(bucket, id, &savedConfig)

	if found {
		ret.SavedUserConfig = &savedConfig
	}

	return
}

func (r *Repository) SaveExtensionUserConfig(id string, savedConfig *extension.SavedUserConfig) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/SaveExtensionUserConfig", &err)

	// Save the config
	bucket := filecache.NewPermanentBucket(getExtensionUserConfigBucketKey(id))
	err = r.fileCacher.SetPerm(bucket, id, savedConfig)
	if err != nil {
		return err
	}

	// Reload the extension
	r.reloadExtension(id)

	return nil
}

// This should be called when the extension is uninstalled
func (r *Repository) deleteExtensionUserConfig(id string) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/deleteExtensionUserConfig", &err)

	// Delete the config
	bucket := filecache.NewPermanentBucket(getExtensionUserConfigBucketKey(id))
	err = r.fileCacher.RemovePerm(bucket.Name())
	if err != nil {
		return err
	}

	return nil
}

// This should be called when the extension is uninstalled
func (r *Repository) deletePluginData(id string) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/deletePluginData", &err)

	plugin.GlobalAppContext.DropPluginData(id)

	return nil
}
