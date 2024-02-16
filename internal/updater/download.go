package updater

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/util"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var (
	ErrExtractionFailed = errors.New("could not extract assets")
)

func (u *Updater) DownloadLatestRelease(assetUrl, dest string) (string, error) {
	if u.LatestRelease == nil {
		return "", errors.New("no new release found")
	}

	//asset, ok := lo.Find(u.LatestRelease.Assets, func(asset ReleaseAsset) bool {
	//	return asset.BrowserDownloadUrl == assetUrl
	//})
	//if !ok {
	//	return "", errors.New("could not find release asset")
	//}

	fpath, err := u.downloadAsset(assetUrl, dest)
	if err != nil {
		return "", err
	}

	folderPath := filepath.Dir(fpath)

	// Uncompress asset on Windows or Mac
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		fp, err := u.decompressAsset(fpath)
		if err != nil {
			return fp, ErrExtractionFailed
		}
		folderPath = fp
	}

	return folderPath, nil
}

// decompressAsset will uncompress the release assets and delete the compressed folder
//   - "/seanime-repo/seanime-v1.0.0.zip" -> "/seanime-repo/seanime-1.0.0/"
func (u *Updater) decompressAsset(archivePath string) (folderPath string, err error) {

	defer util.HandlePanicInModuleWithError("updater/download/decompressAsset", &err)

	// Get the destination folder, it should be the same as the compressed file
	topFolderName := "seanime-" + u.LatestRelease.Version
	folderPath = filepath.Dir(archivePath)
	dest := filepath.Join(filepath.Dir(archivePath), topFolderName)

	// Check if the destination folder already exists
	if _, err := os.Stat(dest); err == nil {
		return folderPath, errors.New("destination folder already exists")
	}

	// Create the destination folder if it doesn't exist
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return folderPath, err
	}

	folderPath = dest

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return folderPath, err
	}
	defer r.Close()

	// Iterate through the files in the archive
	for _, f := range r.File {

		// New file path
		fpath := filepath.Join(dest, f.Name)

		// If the file is a directory, create it
		if f.FileInfo().IsDir() {
			// Creating a new folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Creating the files in the target directory
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return folderPath, err
		}

		// Create the file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return folderPath, err
		}

		// Open the file in the archive
		rc, err := f.Open()

		// Copy the file
		_, err = io.Copy(outFile, rc)

		outFile.Close() // Close the file
		rc.Close()      // Close the file in the archive

		if err != nil {
			return folderPath, err
		}
	}

	r.Close() // Close the archive before deleting

	// Delete the compressed file
	err = os.Remove(archivePath)
	if err != nil {
		return folderPath, err
	}

	return folderPath, nil
}

// downloadAsset will download the release assets to a folder
//   - "seanime-v1.zip" -> "/seanime-repo/seanime-v1.zip"
func (u *Updater) downloadAsset(assetUrl, dest string) (fp string, err error) {

	defer util.HandlePanicInModuleWithError("updater/download/downloadAsset", &err)

	fp = u.getFilePath(assetUrl, dest)

	// Get the data
	resp, err := http.Get(assetUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if the request was successful (status code 200)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file, %s", resp.Status)
	}

	// Create the destination folder if it doesn't exist
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		return "", err
	}

	// Create the file
	out, err := os.Create(fp)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return
}

func (u *Updater) getFilePath(url, dest string) string {
	// Get the file name from the URL
	fileName := filepath.Base(url)
	return filepath.Join(dest, fileName)
}
