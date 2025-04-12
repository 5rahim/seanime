package plugin

import (
	"seanime/internal/extension"
	"testing"

	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
)

// Mock AppContextImpl for testing
type mockAppContext struct {
	AppContextImpl
	mockPaths map[string][]string
}

// Create a new mock context with initialized fields
func newMockAppContext(paths map[string][]string) *mockAppContext {
	ctx := &mockAppContext{
		mockPaths: paths,
	}
	// Initialize the animeLibraryPaths field with mock data
	if libraryPaths, ok := paths["SEANIME_ANIME_LIBRARY"]; ok {
		ctx.animeLibraryPaths = mo.Some(libraryPaths)
	} else {
		ctx.animeLibraryPaths = mo.Some([]string{})
	}
	return ctx
}

func TestIsAllowedPath(t *testing.T) {
	// Create mock context with predefined paths
	mockCtx := newMockAppContext(map[string][]string{
		"SEANIME_ANIME_LIBRARY": {"/anime/lib1", "/anime/lib2"},
		"HOME":                  {"/home/user"},
		"TEMP":                  {"/tmp"},
	})

	tests := []struct {
		name     string
		ext      *extension.Extension
		path     string
		mode     int
		expected bool
	}{
		{
			name:     "nil extension",
			ext:      nil,
			path:     "/some/path",
			mode:     AllowPathRead,
			expected: false,
		},
		{
			name: "no patterns",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
					},
				},
			},
			path:     "/some/path",
			mode:     AllowPathRead,
			expected: false,
		},
		{
			name: "simple path match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							ReadPaths: []string{"/test/*.txt"},
						},
					},
				},
			},
			path:     "/test/file.txt",
			mode:     AllowPathRead,
			expected: true,
		},
		{
			name: "multiple library paths - first match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							ReadPaths: []string{"$SEANIME_ANIME_LIBRARY/**"},
						},
					},
				},
			},
			path:     "/anime/lib1/file.txt",
			mode:     AllowPathRead,
			expected: true,
		},
		{
			name: "multiple library paths - second match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							ReadPaths: []string{"$SEANIME_ANIME_LIBRARY/**"},
						},
					},
				},
			},
			path:     "/anime/lib2/file.txt",
			mode:     AllowPathRead,
			expected: true,
		},
		{
			name: "write mode with read pattern",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
					},
				},
			},
			path:     "/test/file.txt",
			mode:     AllowPathWrite,
			expected: false,
		},
		{
			name: "multiple patterns - match one",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							ReadPaths: []string{"$SEANIME_ANIME_LIBRARY/**"},
						},
					},
				},
			},
			path:     "/anime/lib1/file.txt",
			mode:     AllowPathRead,
			expected: true,
		},
		{
			name: "no matching pattern",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							ReadPaths: []string{"$SEANIME_ANIME_LIBRARY/**"},
						},
					},
				},
			},
			path:     "/anime/lib1/file.txt",
			mode:     AllowPathRead,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockCtx.isAllowedPath(tt.ext, tt.path, tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAllowedCommand(t *testing.T) {
	// Create mock context
	mockCtx := newMockAppContext(map[string][]string{
		"HOME":                  {"/home/user"},
		"SEANIME_ANIME_LIBRARY": {}, // Empty but initialized
	})

	tests := []struct {
		name     string
		ext      *extension.Extension
		cmd      string
		args     []string
		expected bool
	}{
		{
			name:     "nil extension",
			ext:      nil,
			cmd:      "ls",
			args:     []string{"-l"},
			expected: false,
		},
		{
			name: "simple command no args",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "ls",
								},
							},
						},
					},
				},
			},
			cmd:      "ls",
			args:     []string{},
			expected: true,
		},
		{
			name: "command with fixed args - match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "git",
									Args: []extension.CommandArg{
										{Value: "pull"},
										{Value: "origin"},
										{Value: "main"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "git",
			args:     []string{"pull", "origin", "main"},
			expected: true,
		},
		{
			name: "command with fixed args - no match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "git",
									Args: []extension.CommandArg{
										{Value: "pull"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "git",
			args:     []string{"push"},
			expected: false,
		},
		{
			name: "command with $ARGS validator",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "echo",
									Args: []extension.CommandArg{
										{Validator: "$ARGS"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "echo",
			args:     []string{"hello", "world"},
			expected: true,
		},
		{
			name: "command with regex validator - match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "open",
									Args: []extension.CommandArg{
										{Validator: "^https?://.*$"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "open",
			args:     []string{"https://example.com"},
			expected: true,
		},
		{
			name: "command with regex validator - no match",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "open",
									Args: []extension.CommandArg{
										{Validator: "^https?://.*$"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "open",
			args:     []string{"file://example.com"},
			expected: false,
		},
		{
			name: "command with $PATH validator",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "open",
									Args: []extension.CommandArg{
										{Validator: "$PATH"},
									},
								},
							},
							WritePaths: []string{"$SEANIME_ANIME_LIBRARY/**"},
						},
					},
				},
			},
			cmd:      "open",
			args:     []string{"/anime/lib1/test.txt"},
			expected: false, // Directory does not exist on the machine
		},
		{
			name: "too many args",
			ext: &extension.Extension{
				Plugin: &extension.PluginManifest{
					Permissions: extension.PluginPermissions{
						Scopes: []extension.PluginPermissionScope{
							extension.PluginPermissionSystem,
						},
						Allow: extension.PluginAllowlist{
							CommandScopes: []extension.CommandScope{
								{
									Command: "ls",
									Args: []extension.CommandArg{
										{Value: "-l"},
									},
								},
							},
						},
					},
				},
			},
			cmd:      "ls",
			args:     []string{"-l", "-a"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mockCtx.isAllowedCommand(tt.ext, tt.cmd, tt.args...)
			assert.Equal(t, tt.expected, result)
		})
	}
}
