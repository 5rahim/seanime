package extension

import (
	hibiketorrent "github.com/5rahim/hibike/pkg/extension/torrent"
)

type AnimeTorrentProviderExtension interface {
	BaseExtension
	GetProvider() hibiketorrent.AnimeProvider
}

type AnimeTorrentProviderExtensionImpl struct {
	ext      *Extension
	provider hibiketorrent.AnimeProvider
}

func NewAnimeTorrentProviderExtension(ext *Extension, provider hibiketorrent.AnimeProvider) AnimeTorrentProviderExtension {
	return &AnimeTorrentProviderExtensionImpl{
		ext:      ext,
		provider: provider,
	}
}

func (m *AnimeTorrentProviderExtensionImpl) GetProvider() hibiketorrent.AnimeProvider {
	return m.provider
}

func (m *AnimeTorrentProviderExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *AnimeTorrentProviderExtensionImpl) GetType() Type {
	return m.ext.Type
}

func (m *AnimeTorrentProviderExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *AnimeTorrentProviderExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *AnimeTorrentProviderExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *AnimeTorrentProviderExtensionImpl) GetManifestURI() string {
	return m.ext.ManifestURI
}

func (m *AnimeTorrentProviderExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *AnimeTorrentProviderExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *AnimeTorrentProviderExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *AnimeTorrentProviderExtensionImpl) GetPayload() string {
	return m.ext.Payload
}

func (m *AnimeTorrentProviderExtensionImpl) GetMeta() Meta {
	return m.ext.Meta
}
