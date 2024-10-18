package extension

type Consumer interface {
	InitExtensionBank(bank *UnifiedBank)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Type string

type Language string

const (
	TypeAnimeTorrentProvider Type = "anime-torrent-provider"
	TypeMangaProvider        Type = "manga-provider"
	TypeOnlinestreamProvider Type = "onlinestream-provider"
)

const (
	LanguageJavascript Language = "javascript"
	LanguageTypescript Language = "typescript"
	LanguageGo         Language = "go"
)

type Extension struct {
	// ID is the unique identifier of the extension
	// It must be unique across all extensions
	// It must start with a letter and contain only alphanumeric characters
	ID      string `json:"id"`      // e.g. "extension-example"
	Name    string `json:"name"`    // e.g. "Extension"
	Version string `json:"version"` // e.g. "1.0.0"
	// The URI to the extension manifest file.
	// This is "builtin" if the extension is built-in and "" if the extension is local.
	ManifestURI string `json:"manifestURI"` // e.g. "http://cdn.something.app/extensions/extension-example/manifest.json"
	// The programming language of the extension
	// It is used to determine how to interpret the extension
	Language Language `json:"language"` // e.g. "go"
	// Type is the area of the application the extension is targeting
	Type        Type   `json:"type"`        // e.g. "anime-torrent-provider"
	Description string `json:"description"` // e.g. "This extension provides torrents"
	Author      string `json:"author"`      // e.g. "Seanime"
	// Icon is the URL to the extension icon
	Icon string `json:"icon"`
	// Website is the URL to the extension website
	Website string `json:"website"`
	// ISO 639-1 language code.
	// Set this to "multi" if the extension supports multiple languages.
	// Defaults to "en".
	Lang string `json:"lang"`
	// List of authorization scopes required by the extension.
	// The user must grant these permissions before the extension can be loaded.
	Scopes     []string    `json:"scopes,omitempty"` // NOT IMPLEMENTED
	UserConfig *UserConfig `json:"userConfig,omitempty"`
	// Payload is the content of the extension.
	Payload string `json:"payload"`
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// BaseExtension is the base interface for all extensions
// An extension is a JS file that is loaded by HTTP request
type BaseExtension interface {
	GetID() string
	GetName() string
	GetVersion() string
	GetManifestURI() string
	GetLanguage() Language
	GetType() Type
	GetDescription() string
	GetAuthor() string
	GetPayload() string
	GetLang() string
	GetIcon() string
	GetWebsite() string
	GetScopes() []string
	GetUserConfig() *UserConfig
}

func ToExtensionData(ext BaseExtension) *Extension {
	return &Extension{
		ID:          ext.GetID(),
		Name:        ext.GetName(),
		Version:     ext.GetVersion(),
		ManifestURI: ext.GetManifestURI(),
		Language:    ext.GetLanguage(),
		Lang:        GetExtensionLang(ext.GetLang()),
		Type:        ext.GetType(),
		Description: ext.GetDescription(),
		Author:      ext.GetAuthor(),
		Scopes:      ext.GetScopes(),
		UserConfig:  ext.GetUserConfig(),
		Icon:        ext.GetIcon(),
		Website:     ext.GetWebsite(),
		Payload:     ext.GetPayload(),
	}
}

func GetExtensionLang(lang string) string {
	if lang == "" {
		return "en"
	}
	if lang == "all" {
		return "multi"
	}
	return lang
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type InvalidExtensionErrorCode string

const (
	// InvalidExtensionManifestError is returned when the extension manifest is invalid
	InvalidExtensionManifestError InvalidExtensionErrorCode = "invalid_manifest"
	// InvalidExtensionPayloadError is returned when the extension code is invalid / obsolete
	InvalidExtensionPayloadError    InvalidExtensionErrorCode = "invalid_payload"
	InvalidExtensionUserConfigError InvalidExtensionErrorCode = "user_config_error"
	// InvalidExtensionAuthorizationError is returned when some authorization scopes have not been granted
	InvalidExtensionAuthorizationError InvalidExtensionErrorCode = "invalid_authorization"
)

type InvalidExtension struct {
	// Auto-generated ID
	ID        string                    `json:"id"`
	Path      string                    `json:"path"`
	Extension Extension                 `json:"extension"`
	Reason    string                    `json:"reason"`
	Code      InvalidExtensionErrorCode `json:"code"`
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type UserConfig struct {
	// The version of the extension configuration.
	// This is used to determine if the configuration has changed.
	Version int `json:"version"`
	// Whether the extension requires user configuration.
	RequiresConfig bool `json:"requiresConfig"`
	// This will be used to generate the user configuration form, and the values will be passed to the extension.
	Fields []ConfigField `json:"fields"`
}

type SavedUserConfig struct {
	// The version of the extension configuration.
	Version int `json:"version"`
	// The values of the user configuration fields.
	Values map[string]string `json:"values"`
}

const (
	ConfigFieldTypeText   ConfigFieldType = "text"
	ConfigFieldTypeSwitch ConfigFieldType = "switch"
	ConfigFieldTypeSelect ConfigFieldType = "select"
)

type (

	// ConfigField represents a field in an extension's configuration.
	// The fields are defined in the manifest file.
	ConfigField struct {
		Type    ConfigFieldType           `json:"type"`
		Name    string                    `json:"name"`
		Label   string                    `json:"label"`
		Options []ConfigFieldSelectOption `json:"options,omitempty"`
		Default string                    `json:"default,omitempty"`
	}

	ConfigFieldType string

	ConfigFieldSelectOption struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}

	ConfigFieldValueValidator func(value string) error
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
