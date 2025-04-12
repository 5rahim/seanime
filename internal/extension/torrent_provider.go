package extension

import (
	hibiketorrent "seanime/internal/extension/hibike/torrent"
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

func (m *AnimeTorrentProviderExtensionImpl) GetLang() string {
	return GetExtensionLang(m.ext.Lang)
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

func (m *AnimeTorrentProviderExtensionImpl) GetWebsite() string {
	return m.ext.Website
}

func (m *AnimeTorrentProviderExtensionImpl) GetIcon() string {
	return m.ext.Icon
}

func (m *AnimeTorrentProviderExtensionImpl) GetPermissions() []string {
	return m.ext.Permissions
}

func (m *AnimeTorrentProviderExtensionImpl) GetUserConfig() *UserConfig {
	return m.ext.UserConfig
}

func (m *AnimeTorrentProviderExtensionImpl) GetPayloadURI() string {
	return m.ext.PayloadURI
}

func (m *AnimeTorrentProviderExtensionImpl) GetIsDevelopment() bool {
	return m.ext.IsDevelopment
}
