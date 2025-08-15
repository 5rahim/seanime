package updater

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"seanime/internal/constants"
	"seanime/internal/util"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	tempReleaseDir = "seanime_new_release"
	backupDirName  = "backup_restore_if_failed"
)

type (
	SelfUpdater struct {
		logger          *zerolog.Logger
		breakLoopCh     chan struct{}
		originalExePath mo.Option[string]
		updater         *Updater
		fallbackDest    string

		tmpExecutableName string
	}
)

func NewSelfUpdater() *SelfUpdater {
	logger := util.NewLogger()
	ret := &SelfUpdater{
		logger:          logger,
		breakLoopCh:     make(chan struct{}),
		originalExePath: mo.None[string](),
		updater:         New(constants.Version, logger, nil),
	}

	ret.tmpExecutableName = "seanime.exe.old"
	switch runtime.GOOS {
	case "windows":
		ret.tmpExecutableName = "seanime.exe.old"
	default:
		ret.tmpExecutableName = "seanime.old"
	}

	go func() {
		// Delete all files with the .old extension
		exePath := getExePath()
		entries, err := os.ReadDir(filepath.Dir(exePath))
		if err != nil {
			return
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".old") {
				_ = os.RemoveAll(filepath.Join(filepath.Dir(exePath), entry.Name()))
			}
		}

	}()

	return ret
}

// Started returns a channel that will be closed when the app loop should be broken
func (su *SelfUpdater) Started() <-chan struct{} {
	return su.breakLoopCh
}

func (su *SelfUpdater) StartSelfUpdate(fallbackDestination string) {
	su.fallbackDest = fallbackDestination
	close(su.breakLoopCh)
}

// recover will just print a message and attempt to download the latest release
func (su *SelfUpdater) recover(assetUrl string) {

	if su.originalExePath.IsAbsent() {
		return
	}

	if su.fallbackDest != "" {
		su.logger.Info().Str("dest", su.fallbackDest).Msg("selfupdate: Attempting to download the latest release")
		_, _ = su.updater.DownloadLatestRelease(assetUrl, su.fallbackDest)
	}

	su.logger.Error().Msg("selfupdate: Failed to install update. Update downloaded to 'seanime_new_release'")
}

func getExePath() string {
	exe, err := os.Executable() // /path/to/seanime.exe
	if err != nil {
		return ""
	}
	exePath, err := filepath.EvalSymlinks(exe) // /path/to/seanime.exe
	if err != nil {
		return ""
	}

	return exePath
}

func (su *SelfUpdater) Run() error {

	exePath := getExePath()

	su.originalExePath = mo.Some(exePath)

	exeDir := filepath.Dir(exePath) // /path/to

	var files []string

	switch runtime.GOOS {
	case "windows":
		files = []string{
			"seanime.exe",
			"LICENSE",
		}
	default:
		files = []string{
			"seanime",
			"LICENSE",
		}
	}

	// Get the new assets
	su.logger.Info().Msg("selfupdate: Fetching latest release info")

	// Get the latest release
	release, err := su.updater.GetLatestRelease()
	if err != nil {
		su.logger.Error().Err(err).Msg("selfupdate: Failed to get latest release")
		return err
	}

	// Find the asset
	assetName := su.updater.GetReleaseName(release.Version)
	asset, ok := lo.Find(release.Assets, func(asset ReleaseAsset) bool {
		return asset.Name == assetName
	})
	if !ok {
		su.logger.Error().Msg("selfupdate: Asset not found")
		return err
	}

	su.logger.Info().Msg("selfupdate: Downloading latest release")

	// Download the asset to exeDir/seanime_tmp
	newReleaseDir, err := su.updater.DownloadLatestReleaseN(asset.BrowserDownloadUrl, exeDir, tempReleaseDir)
	if err != nil {
		su.logger.Error().Err(err).Msg("selfupdate: Failed to download latest release")
		return err
	}

	// DEVNOTE: Past this point, the application will be broken
	// Use "recover" to attempt to recover the application

	su.logger.Info().Msg("selfupdate: Creating backup")

	// Delete the backup directory if it exists
	_ = os.RemoveAll(filepath.Join(exeDir, backupDirName))
	// Create the backup directory
	backupDir := filepath.Join(exeDir, backupDirName)
	_ = os.MkdirAll(backupDir, 0755)

	// Backup the current assets
	// Copy the files to the backup directory
	// seanime.exe + /backup_restore_if_failed/seanime.exe
	// LICENSE + /backup_restore_if_failed/LICENSE
	for _, file := range files {
		// We don't check for errors here because we don't want to stop the update process if LICENSE is not found for example
		_ = copyFile(filepath.Join(exeDir, file), filepath.Join(backupDir, file))
	}

	su.logger.Info().Msg("selfupdate: Renaming assets")
	time.Sleep(2 * time.Second)

	renamingFailed := false
	failedEntryNames := make([]string, 0)

	// Rename the current assets
	// seanime.exe -> seanime.exe.old
	// LICENSE -> LICENSE.old
	for _, file := range files {
		err = os.Rename(filepath.Join(exeDir, file), filepath.Join(exeDir, file+".old"))
		// We care about the error ONLY if the file is the executable
		if err != nil && (file == "seanime" || file == "seanime.exe") {
			renamingFailed = true
			failedEntryNames = append(failedEntryNames, file)
			//su.recover()
			su.logger.Error().Err(err).Msg("selfupdate: Failed to rename entry")
			//return err
		}
	}

	if renamingFailed {
		fmt.Println("---------------------------------")
		fmt.Println("A second attempt will be made in 30 seconds")
		fmt.Println("---------------------------------")
		time.Sleep(30 * time.Second)
		// Here `failedEntryNames` should only contain NECESSARY files that failed to rename
		for _, entry := range failedEntryNames {
			err = os.Rename(filepath.Join(exeDir, entry), filepath.Join(exeDir, entry+".old"))
			if err != nil {
				su.logger.Error().Err(err).Msg("selfupdate: Failed to rename entry")
				su.recover(asset.BrowserDownloadUrl)
				return err
			}
		}
	}

	// Now all the files have been renamed, we can move the new release to the exeDir

	su.logger.Info().Msg("selfupdate: Moving assets")

	// Move the new release elements to the exeDir
	err = moveContents(newReleaseDir, exeDir)
	if err != nil {
		su.recover(asset.BrowserDownloadUrl)
		su.logger.Error().Err(err).Msg("selfupdate: Failed to move assets")
		return err
	}

	_ = os.Chmod(su.originalExePath.MustGet(), 0755)

	// Delete the new release directory
	_ = os.RemoveAll(newReleaseDir)

	// Start the new executable
	su.logger.Info().Msg("selfupdate: Starting new executable")

	switch runtime.GOOS {
	case "windows":
		err = openWindows(su.originalExePath.MustGet())
	case "darwin":
		err = openMacOS(su.originalExePath.MustGet())
	case "linux":
		err = openLinux(su.originalExePath.MustGet())
	default:
		su.logger.Fatal().Msg("selfupdate: Unsupported platform")
	}

	// Remove .old files (will fail on Windows for executable)
	// Remove seanime.exe.old and LICENSE.old
	for _, file := range files {
		_ = os.RemoveAll(filepath.Join(exeDir, file+".old"))
	}

	// Remove the backup directory
	_ = os.RemoveAll(backupDir)

	os.Exit(0)
	return nil
}

func openWindows(path string) error {
	cmd := util.NewCmd("cmd", "/c", "start", "cmd", "/k", path)
	return cmd.Start()
}

func openMacOS(path string) error {
	script := fmt.Sprintf(`
    tell application "Terminal"
        do script "%s"
        activate
    end tell`, path)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Start()
}

func openLinux(path string) error {
	// Filter out the -update flag or we end up in an infinite update loop
	filteredArgs := slices.DeleteFunc(os.Args, func(s string) bool { return s == "-update" })

	// Replace the current process with the updated executable
	return syscall.Exec(path, filteredArgs, os.Environ())
}

// moveContents moves contents of newReleaseDir to exeDir without deleting existing files
func moveContents(newReleaseDir, exeDir string) error {
	// Ensure exeDir exists
	if err := os.MkdirAll(exeDir, 0755); err != nil {
		return err
	}

	// Copy contents of newReleaseDir to exeDir
	return copyDir(newReleaseDir, exeDir)
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it does not exist
	if err = os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// copyDir recursively copies a directory tree, attempting to preserve permissions.
func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
