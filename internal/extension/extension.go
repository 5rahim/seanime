package extension

type Type string

type Language string

const (
	TypeAnimeTorrentProvider Type = "anime-torrent-provider"
	TypeMangaProvider        Type = "manga-provider"
	TypeOnlinestreamProvider Type = "onlinestream-provider"
	TypeMediaPlayer          Type = "mediaplayer"
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
	// ManifestURI is the URI to the extension
	// It can be a URL or a local file path, depending on the extension origin
	// This is "builtin" if the extension is built-in and "" if the extension is local
	ManifestURI string `json:"manifestURI"` // e.g. "http://cdn.something.app/extensions/extension-example/manifest.json"
	// Language is the programming language of the extension
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
	// List of authorization scopes required by the extension.
	// The user must grant these permissions before the extension can be loaded.
	Scopes []string `json:"scopes,omitempty"`
	Config Config   `json:"config,omitempty"`
	// Payload is the content of the extension.
	Payload string `json:"payload"`
}

type Config struct {
	// Whether the extension requires user configuration.
	RequiresConfig bool `json:"requiresConfig"`
	// This will be used to generate the user configuration form, and the values will be passed to the extension.
	Fields []ConfigField `json:"fields"`
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const (
	ConfigFieldTypeText   ConfigFieldType = "text"
	ConfigFieldTypeSwitch ConfigFieldType = "switch"
	ConfigFieldTypeSelect ConfigFieldType = "select"
	ConfigFieldTypeNumber ConfigFieldType = "number"
)

type (
	// ConfigField represents a field in an extension's configuration.
	// The fields are defined in the manifest file.
	ConfigField struct {
		Type    ConfigFieldType           `json:"type"`
		Name    string                    `json:"name"`
		Options []ConfigFieldSelectOption `json:"options"`
		Default string                    `json:"default"`
	}

	ConfigFieldType string

	ConfigFieldSelectOption struct {
		Value string `json:"value"`
		Label string `json:"label"`
	}

	ConfigFieldValueValidator func(value string) error
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type InvalidExtensionErrorCode string

const (
	// InvalidExtensionManifestError is returned when the extension manifest is invalid
	InvalidExtensionManifestError InvalidExtensionErrorCode = "invalid_manifest"
	// InvalidExtensionPayloadError is returned when the extension code is invalid / obsolete
	InvalidExtensionPayloadError InvalidExtensionErrorCode = "invalid_payload"
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
	GetIcon() string
	GetWebsite() string
	GetScopes() []string
	GetConfig() Config
}

func ToExtensionData(ext BaseExtension) *Extension {
	return &Extension{
		ID:          ext.GetID(),
		Name:        ext.GetName(),
		Version:     ext.GetVersion(),
		ManifestURI: ext.GetManifestURI(),
		Language:    ext.GetLanguage(),
		Type:        ext.GetType(),
		Description: ext.GetDescription(),
		Author:      ext.GetAuthor(),
		Scopes:      ext.GetScopes(),
		Config:      ext.GetConfig(),
		Icon:        ext.GetIcon(),
		Website:     ext.GetWebsite(),
		Payload:     ext.GetPayload(),
	}
}
