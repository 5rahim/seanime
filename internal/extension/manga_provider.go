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
	return TypeMangaProvider
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

func (m *MangaProviderExtensionImpl) GetRepositoryURI() string {
	return m.ext.RepositoryURI
}

func (m *MangaProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
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
