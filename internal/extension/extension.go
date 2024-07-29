package extension

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
	Meta        Meta   `json:"meta"`
	// Payload is the content of the extension.
	// When returning the extension, this field should be emptied so the client knows it is installed.
	Payload string `json:"payload"`
}

type InvalidExtensionErrorCode string

const (
	// InvalidExtensionManifestError is returned when the extension manifest is invalid
	InvalidExtensionManifestError InvalidExtensionErrorCode = "invalid_manifest"
	// InvalidExtensionPayloadError is returned when the extension code is invalid / obsolete
	InvalidExtensionPayloadError InvalidExtensionErrorCode = "invalid_payload"
)

type InvalidExtension struct {
	// Auto-generated ID
	ID        string                    `json:"id"`
	Path      string                    `json:"path"`
	Extension Extension                 `json:"extension"`
	Reason    string                    `json:"reason"`
	Code      InvalidExtensionErrorCode `json:"code"`
}

type Meta struct {
	// Icon is the URL to the extension icon
	Icon string `json:"icon"`
	// Website is the URL to the extension website
	Website string `json:"website"`
}

// BaseExtension is the base interface for all extensions
// An extension is a JS file that is loaded by HTTP request
type BaseExtension interface {
	GetID() string
	// GetName returns the name of the extension
	GetName() string
	// GetVersion returns the version of the extension
	GetVersion() string
	// GetManifestURI returns the URI to the extension
	GetManifestURI() string
	// GetLanguage returns the language of the extension
	GetLanguage() Language
	// GetType returns the type of the extension
	GetType() Type
	// GetDescription returns the description of the extension
	GetDescription() string
	// GetAuthor returns the author of the extension
	GetAuthor() string
	// GetPayload returns the content of the extension
	GetPayload() string
	// GetMeta returns the meta information of the extension
	GetMeta() Meta
}

func InstalledToExtensionData(ext BaseExtension) *Extension {
	return &Extension{
		ID:          ext.GetID(),
		Name:        ext.GetName(),
		Version:     ext.GetVersion(),
		ManifestURI: ext.GetManifestURI(),
		Language:    ext.GetLanguage(),
		Type:        ext.GetType(),
		Description: ext.GetDescription(),
		Author:      ext.GetAuthor(),
		Meta:        ext.GetMeta(),
		// We do not return the payload
	}
}
