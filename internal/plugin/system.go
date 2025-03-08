package plugin

import (
	"errors"
	"io/fs"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"seanime/internal/extension"
	util "seanime/internal/util"
	goja_util "seanime/internal/util/goja"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

var (
	ErrPathNotAuthorized = errors.New("path not authorized")
)

const (
	AllowPathRead  = 0
	AllowPathWrite = 1
)

// BindSystem binds the system module to the Goja runtime.
// Permissions needed: system + allowlist
func (a *AppContextImpl) BindSystem(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) {

	//////////////////////////////////////
	// OS
	//////////////////////////////////////

	osObj := vm.NewObject()

	// _ = osObj.Set("args", os.Args) // NOT INCLUDED
	// _ = osObj.Set("exit", os.Exit) // NOT INCLUDED
	// _ = osObj.Set("getenv", os.Getenv) // NOT INCLUDED
	// _ = osObj.Set("dirFS", os.DirFS) // NOT INCLUDED
	// _ = osObj.Set("getwd", os.Getwd) // NOT INCLUDED
	// _ = osObj.Set("chown", os.Chown) // NOT INCLUDED

	// e.g. $os.platform // "windows"
	_ = osObj.Set("platform", runtime.GOOS)

	// e.g. $os.arch // "amd64"
	_ = osObj.Set("arch", runtime.GOARCH)

	_ = osObj.Set("cmd", func(name string, arg ...string) (*exec.Cmd, error) {
		if !a.isAllowedCommand(ext, name, arg...) {
			return nil, errors.New("command not authorized")
		}

		return exec.Command(name, arg...), nil
	})
	_ = osObj.Set("readFile", func(path string) ([]byte, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, ErrPathNotAuthorized
		}

		return os.ReadFile(path)
	})
	_ = osObj.Set("writeFile", func(path string, data []byte, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.WriteFile(path, data, perm)
	})
	_ = osObj.Set("readDir", func(path string) ([]fs.DirEntry, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, ErrPathNotAuthorized
		}
		return os.ReadDir(path)
	})
	_ = osObj.Set("tempDir", func() string {
		if !a.isAllowedPath(ext, os.TempDir(), AllowPathRead) {
			return ""
		}
		return os.TempDir()
	})
	_ = osObj.Set("truncate", func(path string, size int64) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.Truncate(path, size)
	})
	_ = osObj.Set("mkdir", func(path string, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.Mkdir(path, perm)
	})
	_ = osObj.Set("mkdirAll", func(path string, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.MkdirAll(path, perm)
	})
	_ = osObj.Set("rename", func(oldpath, newpath string) error {
		if !a.isAllowedPath(ext, oldpath, AllowPathWrite) || !a.isAllowedPath(ext, newpath, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.Rename(oldpath, newpath)
	})
	_ = osObj.Set("remove", func(path string) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.Remove(path)
	})
	_ = osObj.Set("removeAll", func(path string) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return os.RemoveAll(path)
	})
	_ = osObj.Set("stat", func(path string) (fs.FileInfo, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, ErrPathNotAuthorized
		}
		return os.Stat(path)
	})

	fileModeObj := vm.NewObject()

	fileModeObj.Set("ModeDir", os.ModeDir)
	fileModeObj.Set("ModeAppend", os.ModeAppend)
	fileModeObj.Set("ModeExclusive", os.ModeExclusive)
	fileModeObj.Set("ModeTemporary", os.ModeTemporary)
	fileModeObj.Set("ModeSymlink", os.ModeSymlink)
	fileModeObj.Set("ModeDevice", os.ModeDevice)
	fileModeObj.Set("ModeNamedPipe", os.ModeNamedPipe)
	fileModeObj.Set("ModeSocket", os.ModeSocket)
	fileModeObj.Set("ModeSetuid", os.ModeSetuid)
	fileModeObj.Set("ModeSetgid", os.ModeSetgid)
	fileModeObj.Set("ModeCharDevice", os.ModeCharDevice)
	fileModeObj.Set("ModeSticky", os.ModeSticky)
	fileModeObj.Set("ModeIrregular", os.ModeIrregular)
	fileModeObj.Set("ModeType", os.ModeType)
	fileModeObj.Set("ModePerm", os.ModePerm)
	_ = vm.Set("$os.FileMode", fileModeObj)

	_ = vm.Set("$os", osObj)

	//////////////////////////////////////
	// Downloader
	//////////////////////////////////////

	a.bindDownloader(vm, logger, ext, scheduler)

	//////////////////////////////////////
	// Filepath
	//////////////////////////////////////

	filepathObj := vm.NewObject()

	filepathObj.Set("base", filepath.Base)
	filepathObj.Set("clean", filepath.Clean)
	filepathObj.Set("dir", filepath.Dir)
	filepathObj.Set("ext", filepath.Ext)
	filepathObj.Set("fromSlash", filepath.FromSlash)

	filepathObj.Set("glob", func(basePath string, pattern string) ([]string, error) {
		if !a.isAllowedPath(ext, basePath, AllowPathRead) {
			return nil, ErrPathNotAuthorized
		}
		return doublestar.Glob(os.DirFS(basePath), pattern)
	})
	filepathObj.Set("isAbs", filepath.IsAbs)
	filepathObj.Set("join", filepath.Join)
	filepathObj.Set("match", doublestar.Match)
	filepathObj.Set("rel", filepath.Rel)
	filepathObj.Set("split", filepath.Split)
	filepathObj.Set("splitList", filepath.SplitList)
	filepathObj.Set("toSlash", filepath.ToSlash)
	filepathObj.Set("walk", func(root string, walkFn filepath.WalkFunc) error {
		if !a.isAllowedPath(ext, root, AllowPathRead) {
			return ErrPathNotAuthorized
		}
		return filepath.Walk(root, walkFn)
	})
	filepathObj.Set("walkDir", func(root string, walkFn fs.WalkDirFunc) error {
		if !a.isAllowedPath(ext, root, AllowPathRead) {
			return ErrPathNotAuthorized
		}
		return filepath.WalkDir(root, walkFn)
	})

	_ = vm.Set("$filepath", filepathObj)

	//////////////////////////////////////
	// osextra
	//////////////////////////////////////

	osExtraObj := vm.NewObject()

	osExtraObj.Set("unwrapAndMove", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return util.UnwrapAndMove(src, dest)
	})

	osExtraObj.Set("unzip", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return util.UnzipFile(src, dest)
	})

	osExtraObj.Set("unrar", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return ErrPathNotAuthorized
		}
		return util.UnrarFile(src, dest)
	})

	_ = vm.Set("$osExtra", osExtraObj)

	//////////////////////////////////////
	// mime
	//////////////////////////////////////

	mimeObj := vm.NewObject()

	mimeObj.Set("parse", func(contentType string) (map[string]interface{}, error) {
		res, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"mediaType":  res,
			"parameters": params,
		}, nil
	})
	_ = vm.Set("$mime", mimeObj)

}

// resolveEnvironmentPaths resolves environment paths in the form of $NAME
// to the actual path.
//
// e.g. $SEANIME_LIBRARY_PATH -> /home/user/anime
func (a *AppContextImpl) resolveEnvironmentPaths(name string) []string {

	switch name {
	case "SEANIME_ANIME_LIBRARY":
		if animeLibraryPaths := a.animeLibraryPaths.MustGet(); len(animeLibraryPaths) > 0 {
			return animeLibraryPaths
		}
		return []string{}
	case "HOME": // %USERPROFILE% on Windows
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return []string{}
		}
		return []string{homeDir}
	case "CACHE": // LocalAppData on Windows
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return []string{}
		}
		return []string{cacheDir}
	case "TEMP": // %TMP%, %TEMP% or %USERPROFILE% on Windows
		tempDir := os.TempDir()
		if tempDir == "" {
			return []string{}
		}
		return []string{tempDir}
	case "CONFIG": // AppData on Windows
		configDir, err := os.UserConfigDir()
		if err != nil {
			return []string{}
		}
		return []string{configDir}
	}

	return []string{}
}

func (a *AppContextImpl) isAllowedPath(ext *extension.Extension, path string, mode int) bool {
	// If the extension doesn't have a plugin manifest or system allowlist, deny access
	if ext == nil || ext.Plugin == nil || ext.Plugin.SystemAllowlist == nil {
		return false
	}

	// Get the appropriate patterns based on the mode
	var patterns []string
	if mode == AllowPathRead {
		patterns = ext.Plugin.SystemAllowlist.AllowReadPaths
	} else if mode == AllowPathWrite {
		patterns = ext.Plugin.SystemAllowlist.AllowWritePaths
	} else {
		// Unknown mode
		return false
	}

	// If no patterns are defined, deny access
	if len(patterns) == 0 {
		return false
	}

	// Normalize the path to use forward slashes and absolute path
	normalizedPath := path
	if !filepath.IsAbs(normalizedPath) {
		// Convert to absolute path
		absPath, err := filepath.Abs(normalizedPath)
		if err != nil {
			return false
		}
		normalizedPath = absPath
	}
	normalizedPath = filepath.ToSlash(normalizedPath)

	// Check if the path matches any of the allowed patterns
	for _, pattern := range patterns {
		// Resolve environment variables in the pattern, which may result in multiple patterns
		resolvedPatterns := a.resolvePattern(pattern)

		for _, resolvedPattern := range resolvedPatterns {
			// Convert to absolute path if needed
			if !filepath.IsAbs(resolvedPattern) && !strings.HasPrefix(resolvedPattern, "*") {
				resolvedPattern = filepath.Join(filepath.Dir(normalizedPath), resolvedPattern)
			}

			// Use doublestar for glob pattern matching
			matched, err := doublestar.Match(resolvedPattern, normalizedPath)
			if err == nil && matched {
				return true
			}
		}
	}

	// No matching pattern found, deny access
	return false
}

// resolvePattern resolves environment variables and special placeholders in a pattern
// Returns a slice of resolved patterns to account for placeholders that can expand to multiple paths
func (a *AppContextImpl) resolvePattern(pattern string) []string {
	// Start with the original pattern
	patterns := []string{pattern}

	// Replace special placeholders with their actual values
	placeholders := []string{"$SEANIME_ANIME_LIBRARY", "$HOME", "$CACHE", "$TEMP", "$CONFIG"}

	for _, placeholder := range placeholders {
		// Extract the placeholder name without the $ prefix
		name := strings.TrimPrefix(placeholder, "$")
		paths := a.resolveEnvironmentPaths(name)

		if len(paths) == 0 {
			continue
		}

		// If the placeholder exists in the pattern and expands to multiple paths,
		// we need to create multiple patterns
		if strings.Contains(pattern, placeholder) && len(paths) > 1 {
			// Create a new set of patterns for each path
			newPatterns := []string{}

			for _, existingPattern := range patterns {
				for _, path := range paths {
					newPattern := strings.ReplaceAll(existingPattern, placeholder, path)
					newPatterns = append(newPatterns, newPattern)
				}
			}

			// Replace the old patterns with the new ones
			patterns = newPatterns
		} else if len(paths) > 0 {
			// If there's only one path or the placeholder doesn't exist,
			// just replace it in all existing patterns
			for i := range patterns {
				patterns[i] = strings.ReplaceAll(patterns[i], placeholder, paths[0])
			}
		}
	}

	// Replace environment variables in all patterns
	for i := range patterns {
		patterns[i] = os.ExpandEnv(patterns[i])
	}

	return patterns
}

func (a *AppContextImpl) isAllowedCommand(ext *extension.Extension, name string, arg ...string) bool {
	// If the extension doesn't have a plugin manifest or system allowlist, deny access
	if ext == nil || ext.Plugin == nil || ext.Plugin.SystemAllowlist == nil {
		return false
	}

	// Get the system allowlist
	allowlist := ext.Plugin.SystemAllowlist

	// Check if the command is allowed in any of the command scopes
	for _, scope := range allowlist.CommandScopes {
		// Check if the command name matches
		if scope.Command != name {
			continue
		}

		// If no args are defined in the scope but args are provided, deny access
		if len(scope.Args) == 0 && len(arg) > 0 {
			continue
		}

		// Check if the arguments match the allowed pattern
		if !a.validateCommandArgs(ext, scope.Args, arg) {
			continue
		}

		// Command and args match the scope, allow access
		return true
	}

	// No matching command scope found, deny access
	return false
}

// validateCommandArgs checks if the provided arguments match the allowed pattern
func (a *AppContextImpl) validateCommandArgs(ext *extension.Extension, allowedArgs []extension.CommandArg, providedArgs []string) bool {
	// If more args are provided than allowed, deny access
	if len(providedArgs) > len(allowedArgs) {
		return false
	}

	// Check each argument
	for i, allowedArg := range allowedArgs {
		// If we've reached the end of the provided args, deny access
		if i >= len(providedArgs) {
			return false
		}

		// If the argument has a fixed value, check if it matches exactly
		if allowedArg.Value != "" {
			if allowedArg.Value != providedArgs[i] {
				return false
			}
			continue
		}

		// If the argument has a validator, check if it matches
		if allowedArg.Validator != "" {
			// Special case: $ARGS allows any value
			if allowedArg.Validator == "$ARGS" {
				continue
			}

			// Special case: $PATH allows any valid file path
			if allowedArg.Validator == "$PATH" {
				// Simple path validation - could be enhanced
				if providedArgs[i] == "" {
					return false
				}

				// Check if the path is allowed
				if !a.isAllowedPath(ext, providedArgs[i], AllowPathWrite) {
					return false
				}

				// Check if the path exists
				if _, err := os.Stat(providedArgs[i]); os.IsNotExist(err) {
					return false
				}

				continue
			}

			// Use regex validation for other validators
			matched, err := regexp.MatchString(allowedArg.Validator, providedArgs[i])
			if err != nil || !matched {
				return false
			}
			continue
		}

		// If neither value nor validator is specified, deny access
		return false
	}

	return true
}
