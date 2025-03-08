package extension

import (
	"strings"
)

type Consumer interface {
	InitExtensionBank(bank *UnifiedBank)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Type string

type Language string

type PluginPermission string

const (
	TypeAnimeTorrentProvider Type = "anime-torrent-provider"
	TypeMangaProvider        Type = "manga-provider"
	TypeOnlinestreamProvider Type = "onlinestream-provider"
	TypePlugin               Type = "plugin"
)

const (
	LanguageJavascript Language = "javascript"
	LanguageTypescript Language = "typescript"
	LanguageGo         Language = "go"
)

var (
	PluginPermissionStorage  PluginPermission = "storage"  // Allows the plugin to store its own data
	PluginPermissionDatabase PluginPermission = "database" // Allows the plugin to read non-auth data from the database and write to it
	PluginPermissionPlayback PluginPermission = "playback" // Allows the plugin to use the playback manager
	PluginPermissionAnilist  PluginPermission = "anilist"  // Allows the plugin to use the Anilist client
	PluginPermissionSystem   PluginPermission = "system"   // Allows the plugin to use the OS/Filesystem/Filepath functions. SystemPermissions must be granted additionally.
	PluginPermissionCron     PluginPermission = "cron"     // Allows the plugin to use the cron manager
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
	// List of permissions asked by the extension.
	// The user must grant these permissions before the extension can be loaded.
	Permissions []string    `json:"permissions,omitempty"` // NOT IMPLEMENTED
	UserConfig  *UserConfig `json:"userConfig,omitempty"`
	// Payload is the content of the extension.
	Payload string `json:"payload"`
	// PayloadURI is the URI to the extension payload.
	// It can be used as an alternative to the Payload field.
	PayloadURI string `json:"payloadURI,omitempty"`
	// Plugin is the manifest of the extension if it is a plugin.
	Plugin *PluginManifest `json:"plugin,omitempty"`
}

type PluginManifest struct {
	Version string `json:"version"`
	// Permissions is a list of permissions that the plugin is asking for.
	// The user must acknowledge these permissions before the plugin can be loaded.
	Permissions []PluginPermission `json:"permissions,omitempty"`
	// SystemAllowlist is a list of system permissions that the plugin is asking for.
	// The user must acknowledge these permissions before the plugin can be loaded.
	SystemAllowlist *PluginSystemAllowlist `json:"systemAllowlist,omitempty"`
}

// PluginSystemAllowlist is a list of system permissions that the plugin is asking for.
//
// The user must acknowledge these permissions before the plugin can be loaded.
type PluginSystemAllowlist struct {
	// AllowReadPaths is a list of paths that the plugin is allowed to read from.
	AllowReadPaths []string `json:"allowReadPaths,omitempty"`
	// AllowWritePaths is a list of paths that the plugin is allowed to write to.
	AllowWritePaths []string `json:"allowWritePaths,omitempty"`
	// CommandScopes defines the commands that the plugin is allowed to execute.
	// Each command scope has a unique identifier and configuration.
	CommandScopes []*CommandScope `json:"commandScopes,omitempty"`
	// AllowCommands is a list of commands that the plugin is allowed to execute.
	// This field is deprecated and kept for backward compatibility.
	// Use CommandScopes instead.
}

// CommandScope defines a specific command or set of commands that can be executed
// with specific arguments and validation rules.
type CommandScope struct {
	// Description explains why this command scope is needed
	Description string `json:"description,omitempty"`
	// Command is the executable program
	Command string `json:"command"`
	// Args defines the allowed arguments for this command
	// If nil or empty, no arguments are allowed
	// If contains "$ARGS", any arguments are allowed at that position
	Args []CommandArg `json:"args,omitempty"`
}

// CommandArg represents an argument for a command
type CommandArg struct {
	// Value is the fixed argument value
	// If empty, Validator must be set
	Value string `json:"value,omitempty"`
	// Validator is a Perl compatible regex pattern to validate dynamic argument values
	// Special values:
	// - "$ARGS" allows any arguments at this position
	// - "$PATH" allows any valid file path
	Validator string `json:"validator,omitempty"`
}

// ReadAllowCommands returns a human-readable representation of the commands
// that the plugin is allowed to execute.
func (p *PluginSystemAllowlist) ReadAllowCommands() []string {
	if p == nil {
		return []string{}
	}

	result := make([]string, 0)

	// Add commands from CommandScopes
	if len(p.CommandScopes) > 0 {
		for _, scope := range p.CommandScopes {
			cmd := scope.Command

			// Build argument string
			args := ""
			for i, arg := range scope.Args {
				if i > 0 {
					args += " "
				}

				if arg.Value != "" {
					args += arg.Value
				} else if arg.Validator == "$ARGS" {
					args += "[any arguments]"
				} else if arg.Validator == "$PATH" {
					args += "[any path]"
				} else if arg.Validator != "" {
					args += "[matching: " + arg.Validator + "]"
				}
			}

			if args != "" {
				cmd += " " + args
			}

			// Add description if available
			if scope.Description != "" {
				cmd += " - " + scope.Description
			}

			result = append(result, cmd)
		}
	}

	return result
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
	GetPayloadURI() string
	GetLang() string
	GetIcon() string
	GetWebsite() string
	GetPermissions() []string
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
		Permissions: ext.GetPermissions(),
		UserConfig:  ext.GetUserConfig(),
		Icon:        ext.GetIcon(),
		Website:     ext.GetWebsite(),
		Payload:     ext.GetPayload(),
		PayloadURI:  ext.GetPayloadURI(),
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

func (p *PluginPermission) String() string {
	return string(*p)
}

func (p *PluginPermission) Is(str string) bool {
	return strings.EqualFold(string(*p), str)
}
