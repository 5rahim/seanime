package debrid_client

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"seanime/internal/util"

	"github.com/nwaples/rardecode/v2"
)

var renamePath = os.Rename

// Unzips a file to the destination
//
//	Example:
//	If "file.zip" contains `folder>file.text`, the file will be extracted to "/path/to/dest/{TMP}/folder/file.txt"
//	unzipFile("file.zip", "/path/to/dest")
func unzipFile(src, dest string) (string, error) {
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer r.Close()

	// Create a temporary folder to extract the files
	extractedDir, err := os.MkdirTemp(dest, "extracted-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp folder: %w", err)
	}

	// Iterate through the files in the archive
	for _, f := range r.File {
		mode := f.Mode()
		if mode&os.ModeSymlink != 0 || (!mode.IsRegular() && !f.FileInfo().IsDir()) {
			return "", fmt.Errorf("%w: %s", util.ErrUnsupportedArchiveEntry, f.Name)
		}

		fpath, err := util.ResolveArchiveEntryPath(extractedDir, f.Name)
		if err != nil {
			return "", fmt.Errorf("failed to resolve archive path: %w", err)
		}
		// If the file is a directory, create it in the destination
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}
		// Make sure the parent directory exists (will not return an error if it already exists)
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		// Open the file in the destination
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}
		// Open the file in the archive
		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return "", err
		}

		// Copy the file from the archive to the destination
		_, err = io.Copy(outFile, rc)
		_ = outFile.Close()
		_ = rc.Close()

		if err != nil {
			return "", err
		}
	}
	return extractedDir, nil
}

// Unrars a file to the destination
//
//	Example:
//	If "file.rar" contains a folder "folder" with a file "file.txt", the file will be extracted to "/path/to/dest/{TM}/folder/file.txt"
//	unrarFile("file.rar", "/path/to/dest")
func unrarFile(src, dest string) (string, error) {
	r, err := rardecode.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("failed to open rar file: %w", err)
	}
	defer r.Close()

	// Create a temporary folder to extract the files
	extractedDir, err := os.MkdirTemp(dest, "extracted-")
	if err != nil {
		return "", fmt.Errorf("failed to create temp folder: %w", err)
	}

	// Iterate through the files in the archive
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		fpath, err := util.ResolveArchiveEntryPath(extractedDir, header.Name)
		if err != nil {
			return "", fmt.Errorf("failed to resolve archive path: %w", err)
		}
		// If the file is a directory, create it in the destination
		if header.IsDir {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make sure the parent directory exists (will not return an error if it already exists)
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		// Open the file in the destination
		outFile, err := os.Create(fpath)
		if err != nil {
			return "", err
		}

		// Copy the file from the archive to the destination
		_, err = io.Copy(outFile, r)
		outFile.Close()

		if err != nil {
			return "", err
		}
	}
	return extractedDir, nil
}

// Moves a folder or file to the destination
//
//	Example:
//	moveFolderOrFileTo("/path/to/src/folder", "/path/to/dest") -> "/path/to/dest/folder"
func moveFolderOrFileTo(src, dest string) error {
	// Ensure the destination folder exists
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		err := os.MkdirAll(dest, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create destination folder: %v", err)
		}
	}

	destFolder := filepath.Join(dest, filepath.Base(src))

	// Move the folder by renaming it
	err := renamePath(src, destFolder)
	if err != nil {
		return fmt.Errorf("failed to move folder: %v", err)
	}

	return nil
}

func moveFolderOrFileToMobile(src, dest string) error {
	err := moveFolderOrFileTo(src, dest)
	if err == nil {
		return nil
	}

	destPath := filepath.Join(dest, filepath.Base(src))
	if copyErr := copyPath(src, destPath); copyErr != nil {
		return fmt.Errorf("failed to move folder: %v", copyErr)
	}
	if removeErr := os.RemoveAll(src); removeErr != nil {
		return fmt.Errorf("failed to remove source folder: %v", removeErr)
	}

	return nil
}

// Moves the contents of a folder to the destination
// It will move ONLY the folder containing multiple files or folders OR a single deeply nested file
//
//	Example:
//
//	Case 1:
//	src/
//		- Anime/
//			- Ep1.mkv
//			- Ep2.mkv
//	moveContentsTo("/path/to/src", "/path/to/dest") -> "/path/to/dest/Anime"
//
//	Case 2:
//	src/
//		- {HASH}/
//			- Anime/
//				- Ep1.mkv
//				- Ep2.mkv
//	moveContentsTo("/path/to/src", "/path/to/dest") -> "/path/to/dest/Anime"
//
//	Case 3:
//	src/
//		- {HASH}/
//			- Anime/
//				- Ep1.mkv
//	moveContentsTo("/path/to/src", "/path/to/dest") -> "/path/to/dest/Ep1.mkv"
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
//	moveContentsTo("/path/to/src", "/path/to/dest") -> "/path/to/dest/Anime"
func moveContentsTo(src, dest string) error {
	return moveContentsToWith(src, dest, moveFolderOrFileTo)
}

func moveContentsToMobile(src, dest string) error {
	return moveContentsToWith(src, dest, moveFolderOrFileToMobile)
}

func moveContentsToWith(src, dest string, move func(string, string) error) error {
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
			err := move(filepath.Join(src, srcEntry.Name()), dest)
			if err != nil {
				return err
			}
		}
		return nil
	}

	folderMap := make(map[string]int)
	err = findFolderChildCount(src, folderMap)
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

	//util.Spew(folderToMove)

	// It's a single file, move that file only
	if folderToMove == "" {
		fp := getDeeplyNestedFile(src)
		if fp == "" {
			return fmt.Errorf("no files found in the source directory")
		}
		return move(fp, dest)
	}

	// Move the folder containing multiple files or folders
	err = move(folderToMove, dest)
	if err != nil {
		return err
	}

	return nil
}

// Finds the folder to move to the destination
func findFolderChildCount(src string, folderMap map[string]int) error {
	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, srcEntry := range srcEntries {
		folderMap[src]++
		if srcEntry.IsDir() {
			err = findFolderChildCount(filepath.Join(src, srcEntry.Name()), folderMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getDeeplyNestedFile(src string) (fp string) {
	srcEntries, err := os.ReadDir(src)
	if err != nil {
		return ""
	}

	for _, srcEntry := range srcEntries {
		if srcEntry.IsDir() {
			return getDeeplyNestedFile(filepath.Join(src, srcEntry.Name()))
		}
		return filepath.Join(src, srcEntry.Name())
	}

	return ""
}

func copyPath(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("unsupported symlink: %s", src)
	}
	if info.IsDir() {
		return copyDir(src, dest, info.Mode())
	}
	return copyFile(src, dest, info.Mode())
}

func copyDir(src, dest string, mode os.FileMode) error {
	if err := os.MkdirAll(dest, mode.Perm()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if err := copyPath(srcPath, destPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dest string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
