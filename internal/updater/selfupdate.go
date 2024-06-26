package updater

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/seanime-app/seanime/internal/constants"
	"github.com/seanime-app/seanime/internal/util"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	tempReleaseDir = "seanime_tmp"
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
		updater:         New(constants.Version, logger),
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

// recover will attempt to recover the application from a failed update AFTER renaming the executable
func (su *SelfUpdater) recover(assetUrl string) {

	if su.originalExePath.IsAbsent() {
		return
	}

	exeDir := filepath.Dir(su.originalExePath.MustGet())

	// Remove all files that do not have the .old extension
	entries, err := os.ReadDir(exeDir)
	if err != nil {
		su.logger.Error().Err(err).Msg("selfupdate: Failed to recover, please manually update application")
		return
	}

	for _, entry := range entries {
		//if entry.Name() == tempReleaseDir {
		//	_ = os.RemoveAll(filepath.Join(exeDir, entry.Name()))
		//	continue
		//}

		if entry.Name() == backupDirName {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".old") {
			_ = os.RemoveAll(filepath.Join(exeDir, entry.Name()))
			continue
		}

		// Trim the .old extension
		if strings.HasSuffix(entry.Name(), ".old") {
			newName := strings.TrimSuffix(entry.Name(), ".old")
			_ = os.Rename(filepath.Join(exeDir, entry.Name()), filepath.Join(exeDir, newName))
		}
	}

	if su.fallbackDest != "" {
		su.logger.Info().Str("dest", su.fallbackDest).Msg("selfupdate: Attempting to download the latest release")
		_, err = su.updater.DownloadLatestRelease(assetUrl, su.fallbackDest)
	}

	su.logger.Info().Msg("selfupdate: Recovered")
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

	// Download the asset
	// The asset will be downloaded to exeDir/seanime_tmp
	newReleaseDir, err := su.updater.DownloadLatestReleaseN(asset.BrowserDownloadUrl, exeDir, tempReleaseDir)
	if err != nil {
		su.logger.Error().Err(err).Msg("selfupdate: Failed to download latest release")
		return err
	}

	// Rename the executable
	//newExePath := filepath.Join(exeDir, su.tmpExecutableName) // /path/to/seanime.exe.old
	//err = os.Rename(exePath, newExePath)
	//if err != nil {
	//	return err
	//}

	// DEVNOTE: Past this point, the application will be broken
	// Use "recover" to attempt to recover the application

	// Get the contents of the directory
	// - web -> to rename
	// - LICENSE -> to rename
	// - seanime.exe -> to rename

	su.logger.Info().Msg("selfupdate: Creating backup")

	entries, err := os.ReadDir(exeDir)
	if err != nil {
		su.recover(asset.BrowserDownloadUrl)
		su.logger.Error().Err(err).Msg("selfupdate: Failed to read directory")
		return err
	}

	// Delete the backup directory if it exists
	_ = os.RemoveAll(filepath.Join(exeDir, backupDirName))
	// Create the backup directory
	backupDir := filepath.Join(exeDir, backupDirName)
	_ = os.MkdirAll(backupDir, 0755)
	// Backup the current assets
	for _, entry := range entries {
		if entry.Name() == tempReleaseDir || entry.Name() == backupDirName {
			continue
		}
		if !entry.IsDir() {
			_ = copyFile(filepath.Join(exeDir, entry.Name()), filepath.Join(backupDir, entry.Name()))
		}
		if entry.IsDir() {
			_ = copyDir(filepath.Join(exeDir, entry.Name()), filepath.Join(backupDir, entry.Name()))
		}
	}

	su.logger.Info().Msg("selfupdate: Renaming assets")
	time.Sleep(5 * time.Second)

	renamingFailed := false
	failedEntryNames := make([]string, 0)

	for _, entry := range entries {
		su.logger.Info().Str("entry", entry.Name()).Msg("selfupdate: Found entry")

		// Do not rename the new release directory
		if entry.Name() == tempReleaseDir || entry.Name() == backupDirName {
			continue
		}

		// Rename the contents
		// - LICENSE -> LICENSE.old
		// This will fail on Windows due to some files inside the directory being in use, this can happen for many reasons
		err = os.Rename(filepath.Join(exeDir, entry.Name()), filepath.Join(exeDir, entry.Name()+".old"))
		if err != nil {
			renamingFailed = true
			failedEntryNames = append(failedEntryNames, entry.Name())
			//su.recover()
			su.logger.Error().Err(err).Msg("selfupdate: Failed to rename entry")
			//return err
		}

		// TESTONLY: Simulate a failed rename
		//var err error
		//if entry.Name() == "web" {
		//	err = fmt.Errorf("Access is denied")
		//} else {
		//	err = os.Rename(filepath.Join(exeDir, entry.Name()), filepath.Join(exeDir, entry.Name()+".old"))
		//}
		//if err != nil {
		//	renamingFailed = true
		//	failedEntryNames = append(failedEntryNames, entry.Name())
		//	su.logger.Error().Err(err).Msg("selfupdate: Failed to rename entry")
		//	su.recover(asset.BrowserDownloadUrl)
		//	return err
		//}
	}

	if renamingFailed {
		fmt.Println("---------------------------------")
		fmt.Println("Please close your browser and the file explorer, a second attempt will be made in 30 seconds")
		fmt.Println("---------------------------------")
		time.Sleep(30 * time.Second)
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
	entries, err = os.ReadDir(exeDir)
	if err != nil {
		su.logger.Warn().Err(err).Msg("selfupdate: Failed to read directory")
		os.Exit(0)
		return nil
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".old") {
			_ = os.RemoveAll(filepath.Join(exeDir, entry.Name()))
		}
	}

	_ = os.RemoveAll(backupDir)

	os.Exit(0)
	return nil
}

func openWindows(path string) error {
	cmd := exec.Command("cmd", "/c", "start", "cmd", "/k", path)
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
	terminals := []string{
		"gnome-terminal", "--", path,
		"konsole", "-e", path,
		"xfce4-terminal", "-e", path,
		"xterm", "-hold", "-e", path,
		"lxterminal", "-e", path,
	}

	for i := 0; i < len(terminals); i += 2 {
		if exec.Command("which", terminals[i]).Run() == nil {
			cmd := exec.Command(terminals[i], terminals[i+1:]...)
			return cmd.Start()
		}
	}

	return fmt.Errorf("no supported terminal emulator found")
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
