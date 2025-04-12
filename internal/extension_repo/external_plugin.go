package extension_repo

import (
	"fmt"
	"path/filepath"
	"slices"

	"seanime/internal/extension"
	"seanime/internal/util"
	"seanime/internal/util/filecache"

	"github.com/samber/lo"
)

const PluginSettingsKey = "1"
const PluginSettingsBucket = "plugin-settings"

var (
	ErrPluginPermissionsNotGranted = fmt.Errorf("plugin: permissions not granted")
)

type (
	StoredPluginSettingsData struct {
		PinnedTrayPluginIds      []string          `json:"pinnedTrayPluginIds"`
		PluginGrantedPermissions map[string]string `json:"pluginGrantedPermissions"` // Extension ID -> Permission Hash
	}
)

var DefaultStoredPluginSettingsData = StoredPluginSettingsData{
	PinnedTrayPluginIds:      []string{},
	PluginGrantedPermissions: map[string]string{},
}

// GetPluginSettings returns the stored plugin settings.
// If no settings are found, it will return the default settings.
func (r *Repository) GetPluginSettings() *StoredPluginSettingsData {
	bucket := filecache.NewPermanentBucket(PluginSettingsBucket)

	var settings StoredPluginSettingsData
	found, _ := r.fileCacher.GetPerm(bucket, PluginSettingsKey, &settings)
	if !found {
		r.fileCacher.SetPerm(bucket, PluginSettingsKey, DefaultStoredPluginSettingsData)
		return &DefaultStoredPluginSettingsData
	}

	return &settings
}

// SetPluginSettingsPinnedTrays sets the pinned tray plugin IDs.
func (r *Repository) SetPluginSettingsPinnedTrays(pinnedTrayPluginIds []string) {
	bucket := filecache.NewPermanentBucket(PluginSettingsBucket)

	settings := r.GetPluginSettings()
	settings.PinnedTrayPluginIds = pinnedTrayPluginIds

	r.fileCacher.SetPerm(bucket, PluginSettingsKey, settings)
}

func (r *Repository) GrantPluginPermissions(pluginId string) {
	// Parse the ext
	ext, err := extractExtensionFromFile(filepath.Join(r.extensionDir, pluginId+".json"))
	if err != nil {
		r.logger.Error().Err(err).Str("filepath", filepath.Join(r.extensionDir, pluginId+".json")).Msg("extensions: Failed to read extension file")
		return
	}

	// Check if the extension is a plugin
	if ext.Type != extension.TypePlugin {
		r.logger.Error().Str("id", pluginId).Msg("extensions: Extension is not a plugin")
		return
	}

	// Grant the plugin permissions
	permissionHash := ext.Plugin.Permissions.GetHash()

	r.setPluginGrantedPermissions(pluginId, permissionHash)

	r.logger.Debug().Str("id", pluginId).Msg("extensions: Granted plugin permissions")

	// Reload the extension
	r.reloadExtension(pluginId)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// setPluginGrantedPermissions sets the granted permissions for a plugin.
func (r *Repository) setPluginGrantedPermissions(pluginId string, permissionHash string) {
	bucket := filecache.NewPermanentBucket(PluginSettingsBucket)

	settings := r.GetPluginSettings()
	if settings.PluginGrantedPermissions == nil {
		settings.PluginGrantedPermissions = make(map[string]string)
	}
	settings.PluginGrantedPermissions[pluginId] = permissionHash

	r.fileCacher.SetPerm(bucket, PluginSettingsKey, settings)
}

// removePluginFromStoredSettings removes a plugin from the stored settings.
func (r *Repository) removePluginFromStoredSettings(pluginId string) {
	bucket := filecache.NewPermanentBucket(PluginSettingsBucket)

	settings := r.GetPluginSettings()
	delete(settings.PluginGrantedPermissions, pluginId)

	if slices.Contains(settings.PinnedTrayPluginIds, pluginId) {
		settings.PinnedTrayPluginIds = lo.Filter(settings.PinnedTrayPluginIds, func(id string, _ int) bool {
			return id != pluginId
		})
	}

	r.fileCacher.SetPerm(bucket, PluginSettingsKey, settings)
}

func (r *Repository) checkPluginPermissions(ext *extension.Extension) (err error) {
	defer util.HandlePanicInModuleWithError("extension_repo/checkPluginPermissions", &err)

	if ext.Type != extension.TypePlugin {
		return nil
	}

	if ext.Plugin == nil {
		return nil
	}

	// Get current plugin permission hash
	pluginPermissionHash := ext.Plugin.Permissions.GetHash()

	// If the plugin has no permissions, skip the check
	if pluginPermissionHash == "" {
		return nil
	}

	// Get stored plugin permission hash
	permissionMap := r.GetPluginSettings().PluginGrantedPermissions

	// Check if the plugin has been granted the required permissions
	granted, found := permissionMap[ext.ID]
	if !found {
		return ErrPluginPermissionsNotGranted
	}

	if granted != pluginPermissionHash {
		return ErrPluginPermissionsNotGranted
	}

	return nil
}
