package plugin

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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
	"sync"

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

type AsyncCmd struct {
	cmd *exec.Cmd

	appContext *AppContextImpl
	scheduler  *goja_util.Scheduler
	vm         *goja.Runtime
}

type CmdHelper struct {
	cmd *exec.Cmd

	stdout io.ReadCloser
	stderr io.ReadCloser

	appContext *AppContextImpl
	scheduler  *goja_util.Scheduler
	vm         *goja.Runtime
}

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
			return nil, fmt.Errorf("command (%s) not authorized", fmt.Sprintf("%s %s", name, strings.Join(arg, " ")))
		}

		return util.NewCmdCtx(context.Background(), name, arg...), nil
	})

	_ = osObj.Set("Interrupt", os.Interrupt)
	_ = osObj.Set("Kill", os.Kill)

	_ = osObj.Set("readFile", func(path string) ([]byte, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, fmt.Errorf("$os.readFile: path (%s) not authorized for read", path)
		}

		return os.ReadFile(path)
	})
	_ = osObj.Set("writeFile", func(path string, data []byte, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.writeFile: path (%s) not authorized for write", path)
		}
		return os.WriteFile(path, data, perm)
	})
	_ = osObj.Set("readDir", func(path string) ([]fs.DirEntry, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, fmt.Errorf("$os.readDir: path (%s) not authorized for read", path)
		}
		return os.ReadDir(path)
	})
	_ = osObj.Set("tempDir", func() (string, error) {
		if !a.isAllowedPath(ext, os.TempDir(), AllowPathRead) {
			return "", fmt.Errorf("$os.tempDir: path (%s) not authorized for read", os.TempDir())
		}
		return os.TempDir(), nil
	})
	_ = osObj.Set("configDir", func() (string, error) {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, configDir, AllowPathRead) {
			return "", fmt.Errorf("$os.configDir: path (%s) not authorized for read", configDir)
		}
		return configDir, nil
	})
	_ = osObj.Set("homeDir", func() (string, error) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, homeDir, AllowPathRead) {
			return "", fmt.Errorf("$os.homeDir: path (%s) not authorized for read", homeDir)
		}
		return homeDir, nil
	})
	_ = osObj.Set("cacheDir", func() (string, error) {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, cacheDir, AllowPathRead) {
			return "", fmt.Errorf("$os.cacheDir: path (%s) not authorized for read", cacheDir)
		}
		return cacheDir, nil
	})
	_ = osObj.Set("truncate", func(path string, size int64) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.truncate: path (%s) not authorized for write", path)
		}
		return os.Truncate(path, size)
	})
	_ = osObj.Set("mkdir", func(path string, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.mkdir: path (%s) not authorized for write", path)
		}
		return os.Mkdir(path, perm)
	})
	_ = osObj.Set("mkdirAll", func(path string, perm fs.FileMode) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.mkdirAll: path (%s) not authorized for write", path)
		}
		return os.MkdirAll(path, perm)
	})
	_ = osObj.Set("rename", func(oldpath, newpath string) error {
		if !a.isAllowedPath(ext, oldpath, AllowPathWrite) || !a.isAllowedPath(ext, newpath, AllowPathWrite) {
			return fmt.Errorf("$os.rename: path (%s) not authorized for write", oldpath)
		}
		return os.Rename(oldpath, newpath)
	})
	_ = osObj.Set("remove", func(path string) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.remove: path (%s) not authorized for write", path)
		}
		return os.Remove(path)
	})
	_ = osObj.Set("removeAll", func(path string) error {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return fmt.Errorf("$os.removeAll: path (%s) not authorized for write", path)
		}
		return os.RemoveAll(path)
	})
	_ = osObj.Set("stat", func(path string) (fs.FileInfo, error) {
		if !a.isAllowedPath(ext, path, AllowPathRead) {
			return nil, fmt.Errorf("$os.stat: path (%s) not authorized for read", path)
		}
		return os.Stat(path)
	})

	_ = osObj.Set("O_RDONLY", os.O_RDONLY)
	_ = osObj.Set("O_WRONLY", os.O_WRONLY)
	_ = osObj.Set("O_RDWR", os.O_RDWR)
	_ = osObj.Set("O_APPEND", os.O_APPEND)
	_ = osObj.Set("O_CREATE", os.O_CREATE)
	_ = osObj.Set("O_EXCL", os.O_EXCL)
	_ = osObj.Set("O_SYNC", os.O_SYNC)
	_ = osObj.Set("O_TRUNC", os.O_TRUNC)

	// Example:
	//	const file = $os.openFile("path/to/file.txt", $os.O_RDWR|$os.O_CREATE, 0644)
	//	file.writeString("Hello, world!")
	//	file.close()
	_ = osObj.Set("openFile", func(path string, flag int, perm fs.FileMode) (*os.File, error) {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return nil, fmt.Errorf("$os.openFile: path (%s) not authorized for write", path)
		}
		return os.OpenFile(path, flag, perm)
	})

	// Example:
	//	const file = $os.create("path/to/file.txt")
	//	file.writeString("Hello, world!")
	//	file.close()
	_ = osObj.Set("create", func(path string) (*os.File, error) {
		if !a.isAllowedPath(ext, path, AllowPathWrite) {
			return nil, fmt.Errorf("$os.create: path (%s) not authorized for write", path)
		}
		return os.Create(path)
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
	_ = osObj.Set("FileMode", fileModeObj)

	_ = vm.Set("$os", osObj)

	//////////////////////////////////////
	// IO
	//////////////////////////////////////

	ioObj := vm.NewObject()

	ioObj.Set("copy", func(dst io.Writer, src io.Reader) (int64, error) {
		return io.Copy(dst, src)
	})

	ioObj.Set("readAll", func(r io.Reader) ([]byte, error) {
		return io.ReadAll(r)
	})

	ioObj.Set("writeString", func(w io.Writer, s string) (int, error) {
		return io.WriteString(w, s)
	})

	ioObj.Set("readAtLeast", func(r io.Reader, buf []byte, min int) (int, error) {
		return io.ReadAtLeast(r, buf, min)
	})

	ioObj.Set("readFull", func(r io.Reader, buf []byte) (int, error) {
		return io.ReadFull(r, buf)
	})

	ioObj.Set("copyN", func(dst io.Writer, src io.Reader, n int64) (int64, error) {
		return io.CopyN(dst, src, n)
	})

	ioObj.Set("copyBuffer", func(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
		return io.CopyBuffer(dst, src, buf)
	})

	ioObj.Set("limitReader", func(r io.Reader, n int64) io.Reader {
		return io.LimitReader(r, n)
	})

	ioObj.Set("newSectionReader", func(r io.ReaderAt, off int64, n int64) io.Reader {
		return io.NewSectionReader(r, off, n)
	})

	ioObj.Set("nopCloser", func(r io.Reader) io.ReadCloser {
		return io.NopCloser(r)
	})

	_ = vm.Set("$io", ioObj)

	//////////////////////////////////////
	// bufio
	//////////////////////////////////////

	bufioObj := vm.NewObject()

	bufioObj.Set("newReader", func(r io.Reader) *bufio.Reader {
		return bufio.NewReader(r)
	})

	bufioObj.Set("newReaderSize", func(r io.Reader, size int) *bufio.Reader {
		return bufio.NewReaderSize(r, size)
	})

	bufioObj.Set("newWriter", func(w io.Writer) *bufio.Writer {
		return bufio.NewWriter(w)
	})

	bufioObj.Set("newWriterSize", func(w io.Writer, size int) *bufio.Writer {
		return bufio.NewWriterSize(w, size)
	})

	bufioObj.Set("newScanner", func(r io.Reader) *bufio.Scanner {
		return bufio.NewScanner(r)
	})

	bufioObj.Set("scanLines", func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return bufio.ScanLines(data, atEOF)
	})

	bufioObj.Set("scanWords", func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return bufio.ScanWords(data, atEOF)
	})

	bufioObj.Set("scanRunes", func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return bufio.ScanRunes(data, atEOF)
	})

	bufioObj.Set("scanBytes", func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return bufio.ScanBytes(data, atEOF)
	})

	_ = vm.Set("$bufio", bufioObj)

	//////////////////////////////////////
	// bytes
	//////////////////////////////////////

	bytesObj := vm.NewObject()

	bytesObj.Set("newBuffer", func(buf []byte) *bytes.Buffer {
		return bytes.NewBuffer(buf)
	})

	bytesObj.Set("newBufferString", func(s string) *bytes.Buffer {
		return bytes.NewBufferString(s)
	})

	bytesObj.Set("newReader", func(b []byte) *bytes.Reader {
		return bytes.NewReader(b)
	})

	_ = vm.Set("$bytes", bytesObj)

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
			return nil, fmt.Errorf("$filepath.glob: path (%s) not authorized for read", basePath)
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
			return fmt.Errorf("$filepath.walk: path (%s) not authorized for read", root)
		}
		return filepath.Walk(root, walkFn)
	})
	filepathObj.Set("walkDir", func(root string, walkFn fs.WalkDirFunc) error {
		if !a.isAllowedPath(ext, root, AllowPathRead) {
			return fmt.Errorf("$filepath.walkDir: path (%s) not authorized for read", root)
		}
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return walkFn(path, d, err)
		})
	})

	skipDir := filepath.SkipDir
	filepathObj.Set("skipDir", skipDir)

	_ = vm.Set("$filepath", filepathObj)

	//////////////////////////////////////
	// osextra
	//////////////////////////////////////

	osExtraObj := vm.NewObject()

	osExtraObj.Set("unwrapAndMove", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return fmt.Errorf("$osExtra.unwrapAndMove: path (%s) not authorized for write", src)
		}
		return util.UnwrapAndMove(src, dest)
	})

	osExtraObj.Set("unzip", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return fmt.Errorf("$osExtra.unzip: path (%s) not authorized for write", src)
		}
		return util.UnzipFile(src, dest)
	})

	osExtraObj.Set("unrar", func(src string, dest string) error {
		if !a.isAllowedPath(ext, src, AllowPathWrite) || !a.isAllowedPath(ext, dest, AllowPathWrite) {
			return fmt.Errorf("$osExtra.unrar: path (%s) not authorized for write", src)
		}
		return util.UnrarFile(src, dest)
	})

	osExtraObj.Set("downloadDir", func() (string, error) {
		donwloadDir, err := util.DownloadDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, donwloadDir, AllowPathRead) {
			return "", fmt.Errorf("$osExtra.downloadDir: path not authorized for read")
		}
		return donwloadDir, nil
	})

	osExtraObj.Set("desktopDir", func() (string, error) {
		desktopDir, err := util.DesktopDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, desktopDir, AllowPathRead) {
			return "", fmt.Errorf("$osExtra.desktopDir: path not authorized for read")
		}
		return desktopDir, nil
	})

	osExtraObj.Set("documentsDir", func() (string, error) {
		documentsDir, err := util.DocumentsDir()
		if err != nil {
			return "", err
		}
		if !a.isAllowedPath(ext, documentsDir, AllowPathRead) {
			return "", fmt.Errorf("$osExtra.documentsDir: path not authorized for read")
		}
		return documentsDir, nil
	})

	osExtraObj.Set("libraryDirs", func() ([]string, error) {
		libraryDirs := []string{}
		if animeLibraryPaths, ok := a.animeLibraryPaths.Get(); ok && len(animeLibraryPaths) > 0 {
			for _, path := range animeLibraryPaths {
				if !a.isAllowedPath(ext, path, AllowPathRead) {
					return nil, fmt.Errorf("$osExtra.libraryDirs: path not authorized for read")
				}
				libraryDirs = append(libraryDirs, path)
			}
		}
		return libraryDirs, nil
	})

	_ = osExtraObj.Set("asyncCmd", func(name string, arg ...string) *AsyncCmd {
		return &AsyncCmd{
			cmd:        util.NewCmdCtx(context.Background(), name, arg...),
			appContext: a,
			scheduler:  scheduler,
			vm:         vm,
		}
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
	mimeObj.Set("format", mime.FormatMediaType)
	_ = vm.Set("$mime", mimeObj)

}

// resolveEnvironmentPaths resolves environment paths in the form of $NAME
// to the actual path.
//
// e.g. $SEANIME_LIBRARY_PATH -> /home/user/anime
func (a *AppContextImpl) resolveEnvironmentPaths(name string) []string {

	switch name {
	case "SEANIME_ANIME_LIBRARY":
		if animeLibraryPaths, ok := a.animeLibraryPaths.Get(); ok && len(animeLibraryPaths) > 0 {
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
	case "DOWNLOAD":
		downloadDir, err := util.DownloadDir()
		if err != nil {
			return []string{}
		}
		return []string{downloadDir}
	case "DESKTOP":
		desktopDir, err := util.DesktopDir()
		if err != nil {
			return []string{}
		}
		return []string{desktopDir}
	case "DOCUMENT":
		documentsDir, err := util.DocumentsDir()
		if err != nil {
			return []string{}
		}
		return []string{documentsDir}
	}

	return []string{}
}

func (a *AppContextImpl) isAllowedPath(ext *extension.Extension, path string, mode int) bool {
	// If the extension doesn't have a plugin manifest or system allowlist, deny access
	if ext == nil || ext.Plugin == nil {
		return false
	}

	// Get the appropriate patterns based on the mode
	var patterns []string
	if mode == AllowPathRead {
		patterns = ext.Plugin.Permissions.Allow.ReadPaths
	} else if mode == AllowPathWrite {
		patterns = ext.Plugin.Permissions.Allow.WritePaths
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
		absPath, err := filepath.Abs(normalizedPath)
		if err != nil {
			return false
		}
		normalizedPath = absPath
	}
	normalizedPath = filepath.ToSlash(normalizedPath)

	// Check if the path is a directory
	isDir := false
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		isDir = true
		// Ensure directory paths end with a slash for proper matching
		if !strings.HasSuffix(normalizedPath, "/") {
			normalizedPath += "/"
		}
	}

	// Check if the path matches any of the allowed patterns
	for _, pattern := range patterns {
		// Resolve environment variables in the pattern, which may result in multiple patterns
		resolvedPatterns := a.resolvePattern(pattern)

		for _, resolvedPattern := range resolvedPatterns {
			// Convert to absolute path if needed
			if !filepath.IsAbs(resolvedPattern) && !strings.HasPrefix(resolvedPattern, "*") {
				resolvedPattern = filepath.Join(filepath.Dir(normalizedPath), resolvedPattern)
			}

			// Direct match attempt
			matched, err := doublestar.Match(resolvedPattern, normalizedPath)
			if err == nil && matched {
				return true
			}

			// For directories, we need special handling
			if isDir {
				// Case 1: Check if this directory is explicitly allowed by a pattern ending with "/"
				if !strings.HasSuffix(resolvedPattern, "/") {
					dirPattern := resolvedPattern
					if !strings.HasSuffix(dirPattern, "/") {
						dirPattern += "/"
					}
					matched, err = doublestar.Match(dirPattern, normalizedPath)
					if err == nil && matched {
						return true
					}
				}

				// Case 2: Check if this directory is covered by a wildcard pattern
				// Strip trailing wildcards to get the base directory pattern
				basePattern := resolvedPattern
				basePattern = strings.TrimSuffix(basePattern, "/**/*")
				basePattern = strings.TrimSuffix(basePattern, "/**")
				basePattern = strings.TrimSuffix(basePattern, "/*")

				// Ensure the base pattern ends with a slash for directory comparison
				if !strings.HasSuffix(basePattern, "/") {
					basePattern += "/"
				}

				// If the path is exactly the base directory or a subdirectory of it
				// AND the original pattern had a wildcard
				if (normalizedPath == basePattern || strings.HasPrefix(normalizedPath, basePattern)) &&
					(strings.HasSuffix(resolvedPattern, "/**") ||
						strings.HasSuffix(resolvedPattern, "/**/*") ||
						strings.HasSuffix(resolvedPattern, "/*")) {
					return true
				}

				// Case 3: Check if the pattern is for a subdirectory of this directory
				// This handles the case where we're checking access to a parent directory
				// when a subdirectory is explicitly allowed
				if strings.HasPrefix(basePattern, normalizedPath) &&
					(strings.HasSuffix(resolvedPattern, "/**") ||
						strings.HasSuffix(resolvedPattern, "/**/*") ||
						strings.HasSuffix(resolvedPattern, "/*")) {
					return true
				}
			} else {
				// For files, check if any parent directory is allowed with wildcards
				parentDir := filepath.Dir(normalizedPath)
				if !strings.HasSuffix(parentDir, "/") {
					parentDir += "/"
				}

				// Check if the file's parent directory matches a directory wildcard pattern
				for _, suffix := range []string{"/**/*", "/**", "/*"} {
					if strings.HasSuffix(resolvedPattern, suffix) {
						basePattern := strings.TrimSuffix(resolvedPattern, suffix)
						if !strings.HasSuffix(basePattern, "/") {
							basePattern += "/"
						}

						if strings.HasPrefix(parentDir, basePattern) {
							return true
						}
					}
				}
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
					// Ensure proper path separator handling
					cleanPath := filepath.ToSlash(path)
					// Replace the placeholder with the path, ensuring no double slashes
					newPattern := strings.ReplaceAll(existingPattern, placeholder, cleanPath)
					newPatterns = append(newPatterns, newPattern)
				}
			}

			// Replace the old patterns with the new ones
			patterns = newPatterns
		} else if len(paths) > 0 {
			// If there's only one path or the placeholder doesn't exist,
			// just replace it in all existing patterns
			for i := range patterns {
				// Ensure proper path separator handling
				cleanPath := filepath.ToSlash(paths[0])
				// Replace the placeholder with the path, ensuring no double slashes
				patterns[i] = strings.ReplaceAll(patterns[i], placeholder, cleanPath)
			}
		}
	}

	// Replace environment variables in all patterns
	for i := range patterns {
		patterns[i] = os.ExpandEnv(patterns[i])
	}

	// Clean up any potential double slashes that might have been introduced
	for i := range patterns {
		// Replace any double slashes with single slashes
		for strings.Contains(patterns[i], "//") {
			patterns[i] = strings.ReplaceAll(patterns[i], "//", "/")
		}
	}

	return patterns
}

func (a *AppContextImpl) isAllowedCommand(ext *extension.Extension, name string, arg ...string) bool {
	// If the extension doesn't have a plugin manifest or system allowlist, deny access
	if ext == nil || ext.Plugin == nil {
		return false
	}

	// Get the system allowlist
	allowlist := ext.Plugin.Permissions.Allow

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
		if len(allowedArgs) > 0 && allowedArgs[len(allowedArgs)-1].Validator == "$ARGS" {
			return true
		}
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
			// Special case: $ARGS allows any value for the rest of the arguments
			if allowedArg.Validator == "$ARGS" {
				return true
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetCommand returns the underlying exec.Cmd
func (c *AsyncCmd) GetCommand() *exec.Cmd {
	return c.cmd
}

func (c *AsyncCmd) Run(callback goja.Callable) error {

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	err = c.cmd.Start()
	if err != nil {
		return err
	}

	// Use WaitGroup to ensure all goroutines complete before Wait() is called
	var wg sync.WaitGroup
	wg.Add(2) // One for stdout, one for stderr

	// Handle stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			data := scanner.Bytes()
			dataBytes := make([]byte, len(data))
			copy(dataBytes, data)

			c.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), c.vm.ToValue(dataBytes), goja.Undefined(), goja.Undefined(), goja.Undefined())
				return err
			})
		}
	}()

	// Handle stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			data := scanner.Bytes()
			dataBytes := make([]byte, len(data))
			copy(dataBytes, data)

			c.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), goja.Undefined(), c.vm.ToValue(dataBytes), goja.Undefined(), goja.Undefined())
				return err
			})
		}
	}()

	// Wait for both stdout and stderr to be fully processed in a separate goroutine
	go func() {
		// Wait for stdout and stderr goroutines to complete
		wg.Wait()

		// Now wait for the command to finish
		err := c.cmd.Wait()

		_ = stdout.Close()
		_ = stderr.Close()

		// Process exit code and signal
		exitCode := 0
		signal := ""

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
			signal = exitErr.String()
		} else if c.cmd.ProcessState != nil {
			exitCode = c.cmd.ProcessState.ExitCode()
			signal = c.cmd.ProcessState.String()
		}

		// Call callback with exit code and signal
		c.scheduler.ScheduleAsync(func() error {
			_, err = callback(goja.Undefined(), goja.Undefined(), goja.Undefined(), c.vm.ToValue(exitCode), c.vm.ToValue(signal))
			return err
		})
	}()

	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (c *CmdHelper) Run(callback goja.Callable) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// // OnData registers a callback to be called when data is available from the command's stdout
// func (c *AsyncCmd) OnData(callback func(data []byte)) error {
// 	stdout, err := c.cmd.StdoutPipe()
// 	if err != nil {
// 		return err
// 	}

// 	go func() {
// 		scanner := bufio.NewScanner(stdout)
// 		for scanner.Scan() {
// 			c.scheduler.ScheduleAsync(func() error {
// 				callback(scanner.Bytes())
// 				return nil
// 			})
// 		}
// 	}()

// 	return nil
// }

// // OnError registers a callback to be called when data is available from the command's stderr
// func (c *AsyncCmd) OnError(callback func(data []byte)) error {
// 	stderr, err := c.cmd.StderrPipe()
// 	if err != nil {
// 		return err
// 	}

// 	go func() {
// 		scanner := bufio.NewScanner(stderr)
// 		for scanner.Scan() {
// 			c.scheduler.ScheduleAsync(func() error {
// 				callback(scanner.Bytes())
// 				return nil
// 			})
// 		}
// 	}()

// 	return nil
// }

// // OnExit registers a callback to be called when the command exits
// func (c *AsyncCmd) OnExit(callback func(code int, signal string)) error {
// 	go func() {
// 		err := c.cmd.Wait()
// 		if err != nil {
// 			return
// 		}
// 		c.scheduler.ScheduleAsync(func() error {
// 			callback(c.cmd.ProcessState.ExitCode(), c.cmd.ProcessState.String())
// 			return nil
// 		})
// 	}()
// 	return nil
// }
