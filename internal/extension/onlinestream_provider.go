package extension

import (
	hibikeonlinestream "github.com/5rahim/hibike/pkg/extension/onlinestream"
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
	return TypeMangaProvider
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

func (m *OnlinestreamProviderExtensionImpl) GetRepositoryURI() string {
	return m.ext.RepositoryURI
}

func (m *OnlinestreamProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
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
