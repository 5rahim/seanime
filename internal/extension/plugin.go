package extension

type PluginExtension interface {
	BaseExtension
}

type PluginExtensionImpl struct {
	ext *Extension
}

func NewPluginExtension(ext *Extension) PluginExtension {
	return &PluginExtensionImpl{
		ext: ext,
	}
}

// func (m *PluginExtensionImpl) GetProvider() hibikemanga.Provider {
// 	return m.provider
// }

func (m *PluginExtensionImpl) GetExtension() *Extension {
	return m.ext
}

func (m *PluginExtensionImpl) GetType() Type {
	return m.ext.Type
}

func (m *PluginExtensionImpl) GetID() string {
	return m.ext.ID
}

func (m *PluginExtensionImpl) GetName() string {
	return m.ext.Name
}

func (m *PluginExtensionImpl) GetVersion() string {
	return m.ext.Version
}

func (m *PluginExtensionImpl) GetManifestURI() string {
	return m.ext.ManifestURI
}

func (m *PluginExtensionImpl) GetLanguage() Language {
	return m.ext.Language
}

func (m *PluginExtensionImpl) GetLang() string {
	return GetExtensionLang(m.ext.Lang)
}

func (m *PluginExtensionImpl) GetDescription() string {
	return m.ext.Description
}

func (m *PluginExtensionImpl) GetAuthor() string {
	return m.ext.Author
}

func (m *PluginExtensionImpl) GetPayload() string {
	return m.ext.Payload
}

func (m *PluginExtensionImpl) GetWebsite() string {
	return m.ext.Website
}

func (m *PluginExtensionImpl) GetIcon() string {
	return m.ext.Icon
}

func (m *PluginExtensionImpl) GetScopes() []string {
	return m.ext.Scopes
}

func (m *PluginExtensionImpl) GetUserConfig() *UserConfig {
	return m.ext.UserConfig
}

func (m *PluginExtensionImpl) GetPayloadURI() string {
	return m.ext.PayloadURI
}
