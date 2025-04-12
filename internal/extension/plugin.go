package extension

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

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

func (p *PluginPermissions) GetHash() string {
	if p == nil {
		return ""
	}

	if len(p.Scopes) == 0 &&
		len(p.Allow.ReadPaths) == 0 &&
		len(p.Allow.WritePaths) == 0 &&
		len(p.Allow.CommandScopes) == 0 {
		return ""
	}

	h := sha256.New()

	// Hash scopes
	for _, scope := range p.Scopes {
		h.Write([]byte(scope))
	}

	// Hash allowlist read paths
	for _, path := range p.Allow.ReadPaths {
		h.Write([]byte("read:" + path))
	}

	// Hash allowlist write paths
	for _, path := range p.Allow.WritePaths {
		h.Write([]byte("write:" + path))
	}

	// Hash command scopes
	for _, cmd := range p.Allow.CommandScopes {
		h.Write([]byte("cmd:" + cmd.Command + ":" + cmd.Description))
		for _, arg := range cmd.Args {
			h.Write([]byte("arg:" + arg.Value + ":" + arg.Validator))
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (p *PluginPermissions) GetDescription() string {
	if p == nil {
		return ""
	}

	// Check if any permissions exist
	if len(p.Scopes) == 0 &&
		len(p.Allow.ReadPaths) == 0 &&
		len(p.Allow.WritePaths) == 0 &&
		len(p.Allow.CommandScopes) == 0 {
		return "No permissions requested."
	}

	var desc strings.Builder

	// Add scopes section if any exist
	if len(p.Scopes) > 0 {
		desc.WriteString("Application:\n")
		for _, scope := range p.Scopes {
			desc.WriteString("• ")
			switch scope {
			case PluginPermissionStorage:
				desc.WriteString("Storage: Store plugin data locally\n")
			case PluginPermissionDatabase:
				desc.WriteString("Database: Read and write non-auth data\n")
			case PluginPermissionPlayback:
				desc.WriteString("Playback: Control media playback and media players\n")
			case PluginPermissionAnilist:
				desc.WriteString("AniList: View and edit your AniList lists\n")
			case PluginPermissionAnilistToken:
				desc.WriteString("AniList Token: View and use your AniList token\n")
			case PluginPermissionSystem:
				desc.WriteString("System: Access OS functions (accessing files, running commands, etc.)\n")
			case PluginPermissionCron:
				desc.WriteString("Cron: Schedule automated tasks\n")
			case PluginPermissionNotification:
				desc.WriteString("Notification: Send system notifications\n")
			case PluginPermissionDiscord:
				desc.WriteString("Discord: Set Discord Rich Presence\n")
			default:
				desc.WriteString(string(scope) + "\n")
			}
		}
		desc.WriteString("\n")
	}

	// Add file permissions if any exist
	hasFilePaths := len(p.Allow.ReadPaths) > 0 || len(p.Allow.WritePaths) > 0
	if hasFilePaths {
		desc.WriteString("File System:\n")

		if len(p.Allow.ReadPaths) > 0 {
			desc.WriteString("• Read from:\n")
			for _, path := range p.Allow.ReadPaths {
				desc.WriteString("\t  - " + explainPath(path) + "\n")
			}
		}

		if len(p.Allow.WritePaths) > 0 {
			desc.WriteString("• Write to:\n")
			for _, path := range p.Allow.WritePaths {
				desc.WriteString("\t  - " + explainPath(path) + "\n")
			}
		}
		desc.WriteString("\n")
	}

	// Add command permissions if any exist
	if len(p.Allow.CommandScopes) > 0 {
		desc.WriteString("Commands:\n")
		for _, cmd := range p.Allow.CommandScopes {
			cmdDesc := "• " + cmd.Command

			// Format arguments
			if len(cmd.Args) > 0 {
				argsDesc := ""
				for _, arg := range cmd.Args {
					if arg.Value != "" {
						argsDesc += " " + arg.Value
					} else if arg.Validator == "$ARGS" {
						argsDesc += " [any arguments]"
					} else if arg.Validator == "$PATH" {
						argsDesc += " [any file path]"
					} else if arg.Validator != "" {
						argsDesc += " [pattern: " + arg.Validator + "]"
					}
				}
				cmdDesc += argsDesc
			}

			// Add command description if available
			if cmd.Description != "" {
				cmdDesc += "\n\t  Purpose: " + cmd.Description
			}

			desc.WriteString(cmdDesc + "\n")
		}
	}

	return strings.TrimSpace(desc.String())
}

// explainPath adds human-readable descriptions to paths containing environment variables
func explainPath(path string) string {
	environmentVars := map[string]string{
		"$SEANIME_ANIME_LIBRARY": "Your anime library directories",
		"$HOME":                  "Your system's Home directory",
		"$CACHE":                 "Your system's Cache directory",
		"$TEMP":                  "Your system's Temporary directory",
		"$CONFIG":                "Your system's Config directory",
		"$DOWNLOAD":              "Your system's Downloads directory",
		"$DESKTOP":               "Your system's Desktop directory",
		"$DOCUMENT":              "Your system's Documents directory",
	}

	result := path

	// Check if we need to add an explanation
	needsExplanation := false
	explanation := ""

	for envVar, description := range environmentVars {
		if strings.Contains(path, envVar) {
			if explanation != "" {
				explanation += ", "
			}
			explanation += fmt.Sprintf("%s = %s", envVar, description)
			needsExplanation = true
		}
	}

	if needsExplanation {
		result += " (" + explanation + ")"
	}

	return result
}

////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginExtension interface {
	BaseExtension
	GetPermissionHash() string
}

type PluginExtensionImpl struct {
	ext *Extension
}

func NewPluginExtension(ext *Extension) PluginExtension {
	return &PluginExtensionImpl{
		ext: ext,
	}
}

func (m *PluginExtensionImpl) GetPermissionHash() string {
	if m.ext.Plugin == nil {
		return ""
	}

	return m.ext.Plugin.Permissions.GetHash()
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
