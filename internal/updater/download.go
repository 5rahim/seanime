package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"seanime/internal/util"
)

var (
	ErrExtractionFailed = errors.New("could not extract assets")
)

// DownloadLatestRelease will download the latest release assets and extract them
// If the decompression fails, the returned string will be the directory to the compressed file
// If the decompression is successful, the returned string will be the directory to the extracted files
func (u *Updater) DownloadLatestRelease(assetUrl, dest string) (string, error) {
	if u.LatestRelease == nil {
		return "", errors.New("no new release found")
	}

	u.logger.Debug().Str("asset_url", assetUrl).Str("dest", dest).Msg("updater: Downloading latest release")

	fpath, err := u.downloadAsset(assetUrl, dest)
	if err != nil {
		return "", err
	}

	dest = filepath.Dir(fpath)

	u.logger.Info().Str("dest", dest).Msg("updater: Downloaded release assets")

	fp, err := u.decompressAsset(fpath, "")
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to decompress release assets")
		return fp, ErrExtractionFailed
	}
	dest = fp

	u.logger.Info().Str("dest", dest).Msg("updater: Successfully decompressed downloaded release assets")

	return dest, nil
}

func (u *Updater) DownloadLatestReleaseN(assetUrl, dest, folderName string) (string, error) {
	if u.LatestRelease == nil {
		return "", errors.New("no new release found")
	}

	u.logger.Debug().Str("asset_url", assetUrl).Str("dest", dest).Msg("updater: Downloading latest release")

	fpath, err := u.downloadAsset(assetUrl, dest)
	if err != nil {
		return "", err
	}

	dest = filepath.Dir(fpath)

	u.logger.Info().Str("dest", dest).Msg("updater: Downloaded release assets")

	fp, err := u.decompressAsset(fpath, folderName)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to decompress release assets")
		return fp, err
	}
	dest = fp

	u.logger.Info().Str("dest", dest).Msg("updater: Successfully decompressed downloaded release assets")

	return dest, nil
}

func (u *Updater) decompressZip(archivePath string, folderName string) (dest string, err error) {
	topFolderName := "seanime-" + u.LatestRelease.Version
	if folderName != "" {
		topFolderName = folderName
	}
	// "/seanime-repo/seanime-v1.0.0.zip" -> "/seanime-repo/seanime-1.0.0/"
	dest = filepath.Join(filepath.Dir(archivePath), topFolderName) // "/seanime-repo/seanime-v1.0.0"

	// Check if the destination folder already exists
	if _, err := os.Stat(dest); err == nil {
		return dest, errors.New("destination folder already exists")
	}

	// Create the destination folder
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return dest, err
	}

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return dest, err
	}
	defer r.Close()

	u.logger.Debug().Msg("updater: Decompressing release assets (zip)")

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return dest, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return dest, err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return dest, err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return dest, err
		}
	}

	r.Close()

	err = os.Remove(archivePath)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to remove compressed file")
		return dest, nil
	}

	u.logger.Debug().Msg("updater: Decompressed release assets (zip)")

	return dest, nil
}

func (u *Updater) decompressTarGz(archivePath string, folderName string) (dest string, err error) {
	topFolderName := "seanime-" + u.LatestRelease.Version
	if folderName != "" {
		topFolderName = folderName
	}
	dest = filepath.Join(filepath.Dir(archivePath), topFolderName)

	if _, err := os.Stat(dest); err == nil {
		return dest, errors.New("destination folder already exists")
	}

	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		return dest, err
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return dest, err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return dest, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	u.logger.Debug().Msg("updater: Decompressing release assets (gzip)")

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return dest, err
		}

		fpath := filepath.Join(dest, header.Name)
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return dest, err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return dest, err
			}

			outFile, err := os.Create(fpath)
			if err != nil {
				return dest, err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return dest, err
			}
			outFile.Close()
		}
	}

	gzr.Close()
	file.Close()

	err = os.Remove(archivePath)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to remove compressed file")
		return dest, nil
	}

	u.logger.Debug().Msg("updater: Decompressed release assets (gzip)")

	return dest, nil
}

// decompressAsset will uncompress the release assets and delete the compressed folder
//   - "/seanime-repo/seanime-v1.0.0.zip" -> "/seanime-repo/seanime-1.0.0/"
func (u *Updater) decompressAsset(archivePath string, folderName string) (dest string, err error) {

	defer util.HandlePanicInModuleWithError("updater/download/decompressAsset", &err)

	u.logger.Debug().Str("archive_path", archivePath).Msg("updater: Decompressing release assets")

	ext := filepath.Ext(archivePath)
	if ext == ".zip" {
		return u.decompressZip(archivePath, folderName)
	} else if ext == ".gz" {
		return u.decompressTarGz(archivePath, folderName)
	}

	u.logger.Error().Msg("updater: Failed to decompress release assets, unsupported archive format")

	return "", fmt.Errorf("unsupported archive format: %s", ext)

}

// downloadAsset will download the release assets to a folder
//   - "seanime-v1.zip" -> "/seanime-repo/seanime-v1.zip"
func (u *Updater) downloadAsset(assetUrl, dest string) (fp string, err error) {

	defer util.HandlePanicInModuleWithError("updater/download/downloadAsset", &err)

	u.logger.Debug().Msg("updater: Downloading assets")

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
	err = os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to download assets")
		return "", err
	}

	// Create the file
	out, err := os.Create(fp)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to download assets")
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		u.logger.Error().Err(err).Msg("updater: Failed to download assets")
		return "", err
	}

	return
}

func (u *Updater) getFilePath(url, dest string) string {
	// Get the file name from the URL
	fileName := filepath.Base(url)
	return filepath.Join(dest, fileName)
}
