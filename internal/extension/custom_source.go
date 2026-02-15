package extension

import (
	hibikecustomsource "seanime/internal/extension/hibike/customsource"
)

type CustomSourceExtension interface {
	BaseExtension
	GetProvider() hibikecustomsource.Provider
	GetExtensionIdentifier() int
	SetExtensionIdentifier(int)
}

type CustomSourceExtensionImpl struct {
	ext                 *Extension
	provider            hibikecustomsource.Provider
	extensionIdentifier int
}

func NewCustomSourceExtension(ext *Extension, provider hibikecustomsource.Provider) CustomSourceExtension {
	return &CustomSourceExtensionImpl{
		ext:      ext,
		provider: provider,
	}
}

func (m *CustomSourceExtensionImpl) GetProvider() hibikecustomsource.Provider {
	return m.provider
}

func (m *CustomSourceExtensionImpl) SetExtensionIdentifier(identifier int) {
	m.extensionIdentifier = identifier
}

func (m *CustomSourceExtensionImpl) GetExtensionIdentifier() int {
	return m.extensionIdentifier
}

func (m *CustomSourceExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *CustomSourceExtensionImpl) GetType() Type {
	return m.ext.Type
}

func (m *CustomSourceExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *CustomSourceExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *CustomSourceExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *CustomSourceExtensionImpl) GetManifestURI() string {
	return m.ext.ManifestURI
}

func (m *CustomSourceExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *CustomSourceExtensionImpl) GetLang() string {
	return GetExtensionLang(m.ext.Lang)
}

func (m *CustomSourceExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *CustomSourceExtensionImpl) GetNotes() string {
	return m.ext.Notes
}

func (m *CustomSourceExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *CustomSourceExtensionImpl) GetPayload() string {
	return m.ext.Payload
}

func (m *CustomSourceExtensionImpl) GetWebsite() string {
	return m.ext.Website
}

func (m *CustomSourceExtensionImpl) GetReadme() string {
	return m.ext.Readme
}

func (m *CustomSourceExtensionImpl) GetIcon() string {
	return m.ext.Icon
}

func (m *CustomSourceExtensionImpl) GetPermissions() []string {
	return m.ext.Permissions
}

func (m *CustomSourceExtensionImpl) GetUserConfig() *UserConfig {
	return m.ext.UserConfig
}

func (m *CustomSourceExtensionImpl) GetSavedUserConfig() *SavedUserConfig {
	return m.ext.SavedUserConfig
}

func (m *CustomSourceExtensionImpl) GetPayloadURI() string {
	return m.ext.PayloadURI
}

func (m *CustomSourceExtensionImpl) GetIsDevelopment() bool {
	return m.ext.IsDevelopment
}

func (m *CustomSourceExtensionImpl) GetPluginManifest() *PluginManifest {
	return m.ext.Plugin
}
