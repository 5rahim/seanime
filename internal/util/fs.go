package util

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nwaples/rardecode/v2"
)

func DirSize(path string) (uint64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return uint64(size), err
}

func IsValidMediaFile(path string) bool {
	return !strings.HasPrefix(path, "._")
}
func IsValidVideoExtension(ext string) bool {
	validExtensions := map[string]struct{}{
		".mp4": {}, ".avi": {}, ".mkv": {}, ".mov": {}, ".flv": {}, ".wmv": {}, ".webm": {},
		".mpeg": {}, ".mpg": {}, ".m4v": {}, ".3gp": {}, ".3g2": {}, ".ogg": {}, ".ogv": {},
		".vob": {}, ".mts": {}, ".m2ts": {}, ".ts": {}, ".f4v": {}, ".ogm": {}, ".rm": {},
		".rmvb": {}, ".drc": {}, ".yuv": {}, ".asf": {}, ".amv": {}, ".m2v": {}, ".mpe": {},
		".mpv": {}, ".mp2": {}, ".svi": {}, ".mxf": {}, ".roq": {}, ".nsv": {}, ".f4p": {},
		".f4a": {}, ".f4b": {},
	}
	ext = strings.ToLower(ext)
	_, exists := validExtensions[ext]
	return exists
}

func IsSubdirectory(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

func IsSubdirectoryOfAny(dirs []string, child string) bool {
	for _, dir := range dirs {
		if IsSubdirectory(dir, child) {
			return true
		}
	}
	return false
}

func IsSameDir(dir1, dir2 string) bool {
	if runtime.GOOS == "windows" {
		dir1 = strings.ToLower(dir1)
		dir2 = strings.ToLower(dir2)
	}

	absDir1, err := filepath.Abs(dir1)
	if err != nil {
		return false
	}
	absDir2, err := filepath.Abs(dir2)
	if err != nil {
		return false
	}
	return absDir1 == absDir2
}

func IsFileUnderDir(filePath, dir string) bool {
	// Get the absolute path of the file
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	// Get the absolute path of the directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		absFilePath = strings.ToLower(absFilePath)
		absDir = strings.ToLower(absDir)
	}

	// Check if the file path starts with the directory path
	return strings.HasPrefix(absFilePath, absDir+string(os.PathSeparator))
}

// UnzipFile unzips a file to the destination.
//
//	Example:
//	// If "file.zip" contains `folder > file.text`
//	UnzipFile("file.zip", "/path/to/dest") // -> "/path/to/dest/folder/file.txt"
//	// If "file.zip" contains `file.txt`
//	UnzipFile("file.zip", "/path/to/dest") // -> "/path/to/dest/file.txt"
func UnzipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// Create a temporary folder to extract the files
	extractedDir, err := os.MkdirTemp(filepath.Dir(dest), ".extracted-")
	if err != nil {
		return fmt.Errorf("failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(extractedDir)

	// Iterate through the files in the archive
	for _, f := range r.File {
		// Get the full path of the file in the destination
		fpath := filepath.Join(extractedDir, f.Name)
		// If the file is a directory, create it in the destination
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		// Make sure the parent directory exists (will not return an error if it already exists)
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Open the file in the destination
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		// Open the file in the archive
		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		// Copy the file from the archive to the destination
		_, err = io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the contents of the extracted directory to the destination
	entries, err := os.ReadDir(extractedDir)
	if err != nil {
		return fmt.Errorf("failed to read extracted directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(extractedDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		// Remove existing file/directory at destination if it exists
		_ = os.RemoveAll(destPath)

		// Move the file/directory to the destination
		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move extracted item %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// UnrarFile unzips a rar file to the destination.
func UnrarFile(src, dest string) error {
	r, err := rardecode.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open rar file: %w", err)
	}
	defer r.Close()

	// Create a temporary folder to extract the files
	extractedDir, err := os.MkdirTemp(filepath.Dir(dest), ".extracted-")
	if err != nil {
		return fmt.Errorf("failed to create temp folder: %w", err)
	}
	defer os.RemoveAll(extractedDir)

	// Iterate through the files in the archive
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to get next file in archive: %w", err)
		}

		// Get the full path of the file in the destination
		fpath := filepath.Join(extractedDir, header.Name)
		// If the file is a directory, create it in the destination
		if header.IsDir {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make sure the parent directory exists (will not return an error if it already exists)
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		// Open the file in the destination
		outFile, err := os.Create(fpath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Copy the file from the archive to the destination
		_, err = io.Copy(outFile, r)
		outFile.Close()

		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// Ensure the destination directory exists
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the contents of the extracted directory to the destination
	entries, err := os.ReadDir(extractedDir)
	if err != nil {
		return fmt.Errorf("failed to read extracted directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(extractedDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		// Remove existing file/directory at destination if it exists
		_ = os.RemoveAll(destPath)

		// Move the file/directory to the destination
		if err := os.Rename(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to move extracted item %s: %w", entry.Name(), err)
		}
	}

	return nil
}

// MoveToDestination moves a folder or file to the destination
//
//	Example:
//	MoveToDestination("/path/to/src/folder", "/path/to/dest") // -> "/path/to/dest/folder"
func MoveToDestination(src, dest string) error {
	// Ensure the destination folder exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err := os.MkdirAll(dest, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create destination folder: %v", err)
		}
	}

	destFolder := filepath.Join(dest, filepath.Base(src))

	// Move the folder by renaming it
	err := os.Rename(src, destFolder)
	if err != nil {
		return fmt.Errorf("failed to move folder: %v", err)
	}

	return nil
}

// UnwrapAndMove moves the last subfolder containing the files to the destination.
// If there is a single file, it will move that file only.
//
//	Example:
//
//	Case 1:
//	src/
//		- Anime/
//			- Ep1.mkv
//			- Ep2.mkv
//	UnwrapAndMove("/path/to/src", "/path/to/dest") // -> "/path/to/dest/Anime"
//
//	Case 2:
//	src/
//		- {HASH}/
//			- Anime/
//				- Ep1.mkv
//				- Ep2.mkv
//	UnwrapAndMove("/path/to/src", "/path/to/dest") // -> "/path/to/dest/Anime"
//
//	Case 3:
//	src/
//		- {HASH}/
//			- Anime/
//				- Ep1.mkv
//	UnwrapAndMove("/path/to/src", "/path/to/dest") // -> "/path/to/dest/Ep1.mkv"
//
//	Case 4:
//	src/
//		- {HASH}/
//			- Anime/
//				- Anime 1/
//					- Ep1.mkv
//					- Ep2.mkv
//				- Anime 2/
//					- Ep1.mkv
//					- Ep2.mkv
//	UnwrapAndMove("/path/to/src", "/path/to/dest") // -> "/path/to/dest/Anime"
func UnwrapAndMove(src, dest string) error {
	// Ensure the source and destination directories exist
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source directory does not exist: %s", src)
	}
	_ = os.MkdirAll(dest, os.ModePerm)

	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// If the source folder contains multiple files or folders, move its contents to the destination
	if len(srcEntries) > 1 {
		for _, srcEntry := range srcEntries {
			err := MoveToDestination(filepath.Join(src, srcEntry.Name()), dest)
			if err != nil {
				return err
			}
		}
		return nil
	}

	folderMap := make(map[string]int)
	err = FindFolderChildCount(src, folderMap)
	if err != nil {
		return err
	}

	var folderToMove string
	for folder, count := range folderMap {
		if count > 1 {
			if folderToMove == "" || len(folder) < len(folderToMove) {
				folderToMove = folder
			}
			continue
		}
	}

	// It's a single file, move that file only
	if folderToMove == "" {
		fp := GetDeepestFile(src)
		if fp == "" {
			return fmt.Errorf("no files found in the source directory")
		}
		return MoveToDestination(fp, dest)
	}

	// Move the folder containing multiple files or folders
	err = MoveToDestination(folderToMove, dest)
	if err != nil {
		return err
	}

	return nil
}

// Finds the folder to move to the destination
func FindFolderChildCount(src string, folderMap map[string]int) error {
	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, srcEntry := range srcEntries {
		folderMap[src]++
		if srcEntry.IsDir() {
			err = FindFolderChildCount(filepath.Join(src, srcEntry.Name()), folderMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func GetDeepestFile(src string) (fp string) {
	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return ""
	}

	for _, srcEntry := range srcEntries {
		if srcEntry.IsDir() {
			return GetDeepestFile(filepath.Join(src, srcEntry.Name()))
		}
		return filepath.Join(src, srcEntry.Name())
	}

	return ""
}
