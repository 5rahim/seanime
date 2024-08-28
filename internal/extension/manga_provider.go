package extension

import (
	hibikemanga "github.com/5rahim/hibike/pkg/extension/manga"
)

type MangaProviderExtension interface {
	BaseExtension
	GetProvider() hibikemanga.Provider
}

type MangaProviderExtensionImpl struct {
	ext      *Extension
	provider hibikemanga.Provider
}

func NewMangaProviderExtension(ext *Extension, provider hibikemanga.Provider) MangaProviderExtension {
	return &MangaProviderExtensionImpl{
		ext:      ext,
		provider: provider,
	}
}

func (m *MangaProviderExtensionImpl) GetProvider() hibikemanga.Provider {
	return m.provider
}

func (m *MangaProviderExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *MangaProviderExtensionImpl) GetType() Type {
	return m.ext.Type
}

func (m *MangaProviderExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *MangaProviderExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *MangaProviderExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *MangaProviderExtensionImpl) GetManifestURI() string {
	return m.ext.ManifestURI
}

func (m *MangaProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *MangaProviderExtensionImpl) GetLang() string {
	return GetExtensionLang(m.ext.Lang)
}

func (m *MangaProviderExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *MangaProviderExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *MangaProviderExtensionImpl) GetPayload() string {
	return m.ext.Payload
}

func (m *MangaProviderExtensionImpl) GetWebsite() string {
	return m.ext.Website
}

func (m *MangaProviderExtensionImpl) GetIcon() string {
	return m.ext.Icon
}

func (m *MangaProviderExtensionImpl) GetScopes() []string {
	return m.ext.Scopes
}

func (m *MangaProviderExtensionImpl) GetConfig() Config {
	return m.ext.Config
}
