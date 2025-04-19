package extension

import (
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
)

type OnlinestreamProviderExtension interface {
	BaseExtension
	GetProvider() hibikeonlinestream.Provider
}

type OnlinestreamProviderExtensionImpl struct {
	ext      *Extension
	provider hibikeonlinestream.Provider
}

func NewOnlinestreamProviderExtension(ext *Extension, provider hibikeonlinestream.Provider) OnlinestreamProviderExtension {
	return &OnlinestreamProviderExtensionImpl{
		ext:      ext,
		provider: provider,
	}
}

func (m *OnlinestreamProviderExtensionImpl) GetProvider() hibikeonlinestream.Provider {
	return m.provider
}

func (m *OnlinestreamProviderExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *OnlinestreamProviderExtensionImpl) GetType() Type {
	return m.ext.Type
}

func (m *OnlinestreamProviderExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *OnlinestreamProviderExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *OnlinestreamProviderExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *OnlinestreamProviderExtensionImpl) GetManifestURI() string {
	return m.ext.ManifestURI
}

func (m *OnlinestreamProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *OnlinestreamProviderExtensionImpl) GetLang() string {
	return GetExtensionLang(m.ext.Lang)
}

func (m *OnlinestreamProviderExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *OnlinestreamProviderExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *OnlinestreamProviderExtensionImpl) GetPayload() string {
	return m.ext.Payload
}

func (m *OnlinestreamProviderExtensionImpl) GetWebsite() string {
	return m.ext.Website
}

func (m *OnlinestreamProviderExtensionImpl) GetIcon() string {
	return m.ext.Icon
}

func (m *OnlinestreamProviderExtensionImpl) GetPermissions() []string {
	return m.ext.Permissions
}

func (m *OnlinestreamProviderExtensionImpl) GetUserConfig() *UserConfig {
	return m.ext.UserConfig
}

func (m *OnlinestreamProviderExtensionImpl) GetSavedUserConfig() *SavedUserConfig {
	return m.ext.SavedUserConfig
}

func (m *OnlinestreamProviderExtensionImpl) GetPayloadURI() string {
	return m.ext.PayloadURI
}

func (m *OnlinestreamProviderExtensionImpl) GetIsDevelopment() bool {
	return m.ext.IsDevelopment
}
