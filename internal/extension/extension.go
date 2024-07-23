package extension

type Type string

type Language string

const (
	TypeTorrentProvider      Type = "torrent-provider"
	TypeMangaProvider        Type = "manga-provider"
	TypeOnlinestreamProvider Type = "onlinestream-provider"
)

const (
	LanguageJavascript Language = "javascript"
	LanguageGo         Language = "go"
)

type Extension struct {
	// ID is the unique identifier of the extension
	// It must be unique across all extensions
	// It must start with a letter and contain only alphanumeric characters
	ID      string `json:"id"`      // e.g. "nyaa"
	Name    string `json:"name"`    // e.g. "Nyaa"
	Version string `json:"version"` // e.g. "1.0.0"
	// RepositoryURI is the URI to the extension
	// It can be a URL or a local file path, depending on the extension origin
	// This is empty if the extension is built-in
	RepositoryURI string `json:"repositoryURI"` // e.g. "http://cdn.something.app/extensions/nyaa/latest.js"
	// Language is the programming language of the extension
	// It is used to determine how to interpret the extension
	Language Language `json:"language"` // e.g. "go"
	// Type is the area of the application the extension is targeting
	Type        Type   `json:"type"`        // e.g. "torrent-provider"
	Description string `json:"description"` // e.g. "Nyaa torrent search extension"
	Author      string `json:"author"`      // e.g. "Seanime"
	// Payload is the content of the extension
	Payload string `json:"payload"`
}

// BaseExtension is the base interface for all extensions
// An extension is a JS file that is loaded by HTTP request
type BaseExtension interface {
	GetID() string
	// GetName returns the name of the extension
	GetName() string
	// GetVersion returns the version of the extension
	GetVersion() string
	// GetRepositoryURI returns the URI to the extension
	GetRepositoryURI() string
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
}
