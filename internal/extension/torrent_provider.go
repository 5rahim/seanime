package extension

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

type TorrentProviderExtension interface {
	BaseExtension
	GetProvider() hibiketorrent.Provider
}

type TorrentProviderExtensionImpl struct {
	ext      *Extension
	provider hibiketorrent.Provider
}

func NewTorrentProviderExtension(ext *Extension, provider hibiketorrent.Provider) TorrentProviderExtension {
	return &TorrentProviderExtensionImpl{
		ext:      ext,
		provider: provider,
	}
}

func (m *TorrentProviderExtensionImpl) GetProvider() hibiketorrent.Provider {
	return m.provider
}

func (m *TorrentProviderExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *TorrentProviderExtensionImpl) GetType() Type {
	return TypeMangaProvider
}

func (m *TorrentProviderExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *TorrentProviderExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *TorrentProviderExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *TorrentProviderExtensionImpl) GetRepositoryURI() string {
	return m.ext.RepositoryURI
}

func (m *TorrentProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *TorrentProviderExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *TorrentProviderExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *TorrentProviderExtensionImpl) GetPayload() string {
	return m.ext.Payload
}
