package extension

const (
	PluginManifestVersion = "1"
)

var (
	PluginPermissionStorage      PluginPermissionScope = "storage"       // Allows the plugin to store its own data
	PluginPermissionDatabase     PluginPermissionScope = "database"      // Allows the plugin to read non-auth data from the database and write to it
	PluginPermissionPlayback     PluginPermissionScope = "playback"      // Allows the plugin to use the playback manager
	PluginPermissionAnilist      PluginPermissionScope = "anilist"       // Allows the plugin to use the Anilist client
	PluginPermissionAnilistToken PluginPermissionScope = "anilist-token" // Allows the plugin to see and use the Anilist token
	PluginPermissionSystem       PluginPermissionScope = "system"        // Allows the plugin to use the OS/Filesystem/Filepath functions. SystemPermissions must be granted additionally.
	PluginPermissionCron         PluginPermissionScope = "cron"          // Allows the plugin to use the cron manager
	PluginPermissionNotification PluginPermissionScope = "notification"  // Allows the plugin to use the notification manager
	PluginPermissionDiscord      PluginPermissionScope = "discord"       // Allows the plugin to use the discord rpc
)

type PluginManifest struct {
	Version string `json:"version"`
	// Permissions is a list of permissions that the plugin is asking for.
	// The user must acknowledge these permissions before the plugin can be loaded.
	Permissions PluginPermissions `json:"permissions,omitempty"`
}

type PluginPermissions struct {
	Scopes []PluginPermissionScope `json:"scopes,omitempty"`
	Allow  PluginAllowlist         `json:"allow,omitempty"`
}

// PluginAllowlist is a list of system permissions that the plugin is asking for.
//
// The user must acknowledge these permissions before the plugin can be loaded.
type PluginAllowlist struct {
	// ReadPaths is a list of paths that the plugin is allowed to read from.
	ReadPaths []string `json:"readPaths,omitempty"`
	// WritePaths is a list of paths that the plugin is allowed to write to.
	WritePaths []string `json:"writePaths,omitempty"`
	// CommandScopes defines the commands that the plugin is allowed to execute.
	// Each command scope has a unique identifier and configuration.
	CommandScopes []CommandScope `json:"commandScopes,omitempty"`
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
func (p *PluginAllowlist) ReadAllowCommands() []string {
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

////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginExtension interface {
	BaseExtension
	IsPlugin() bool
}

type PluginExtensionImpl struct {
	ext *Extension
}

func NewPluginExtension(ext *Extension) PluginExtension {
	return &PluginExtensionImpl{
		ext: ext,
	}
}

func (m *PluginExtensionImpl) IsPlugin() bool {
	return true
}

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

func (m *PluginExtensionImpl) GetPermissions() []string {
	return m.ext.Permissions
}

func (m *PluginExtensionImpl) GetUserConfig() *UserConfig {
	return m.ext.UserConfig
}

func (m *PluginExtensionImpl) GetPayloadURI() string {
	return m.ext.PayloadURI
}

func (m *PluginExtensionImpl) GetIsDevelopment() bool {
	return m.ext.IsDevelopment
}
