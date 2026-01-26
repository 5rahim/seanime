package scanner

import (
	"os"
	"path/filepath"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata_provider"
	"seanime/internal/database/db"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner_Shelving(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatal(err)
	}

	// a temporary directory for the library
	tempDir, err := os.MkdirTemp("", "seanime_test_library")
	if err != nil {
		t.Fatal(err)
	}
	// fix path mismatch on macos
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	filename := "[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv"
	filePath := filepath.Join(tempDir, filename)

	anilistClientRef := util.NewRef[anilist.AnilistClient](anilistClient)
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClientRef, extensionBankRef, logger, database)
	anilistPlatform.SetUsername(test_utils.ConfigData.Provider.AnilistUsername)
	metadataProvider := metadata_provider.GetFakeProvider(t, database)
	wsEventManager := events.NewMockWSEventManager(util.NewLogger())

	t.Run("Shelve missing locked file", func(t *testing.T) {
		// Create a locked file that exists in the DB but not on disk
		lockedLf := anime.NewLocalFile(filePath, tempDir)
		lockedLf.Locked = true
		lockedLf.MediaId = 1

		existingLfs := []*anime.LocalFile{lockedLf}

		scanner := &Scanner{
			DirPath:              tempDir,
			Enhanced:             false,
			PlatformRef:          util.NewRef(anilistPlatform),
			MetadataProviderRef:  util.NewRef(metadataProvider),
			Logger:               logger,
			WSEventManager:       wsEventManager,
			ExistingLocalFiles:   existingLfs,
			SkipLockedFiles:      true,
			SkipIgnoredFiles:     false,
			WithShelving:         true,
			ExistingShelvedFiles: []*anime.LocalFile{},
		}

		// Run Scan
		lfs, err := scanner.Scan(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		// Verify results
		// The file is missing from disk, so it should NOT be in lfs
		assert.Equal(t, 0, len(lfs), "Expected returned local files to be empty")

		// It should be in ShelvedLocalFiles because it was locked and missing
		shelved := scanner.GetShelvedLocalFiles()
		assert.Equal(t, 1, len(shelved), "Expected 1 shelved file")
		if len(shelved) > 0 {
			assert.Equal(t, filePath, shelved[0].Path, "Shelved file path mismatch")
		}
	})

	t.Run("Unshelve reappearing file", func(t *testing.T) {
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()

		shelvedLf := anime.NewLocalFile(filePath, tempDir)
		shelvedLf.Locked = true

		existingShelvedFiles := []*anime.LocalFile{shelvedLf}
		var existingLfs []*anime.LocalFile

		scanner := &Scanner{
			DirPath:              tempDir,
			Enhanced:             false,
			PlatformRef:          util.NewRef(anilistPlatform),
			MetadataProviderRef:  util.NewRef(metadataProvider),
			Logger:               logger,
			WSEventManager:       wsEventManager,
			ExistingLocalFiles:   existingLfs,
			SkipLockedFiles:      true,
			SkipIgnoredFiles:     false,
			WithShelving:         true,
			ExistingShelvedFiles: existingShelvedFiles,
		}

		lfs, err := scanner.Scan(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		// it should be in lfs
		assert.Equal(t, 1, len(lfs), "Expected 1 local file found")
		if len(lfs) > 0 {
			assert.Equal(t, filePath, lfs[0].Path)
		}

		// It should NOT be in ShelvedLocalFiles anymore
		shelved := scanner.GetShelvedLocalFiles()
		assert.Equal(t, 0, len(shelved), "Expected 0 shelved files")
	})

	t.Run("Yeet deleted shelved file", func(t *testing.T) {
		// Prepare shelved file
		shelvedLf := anime.NewLocalFile(filePath, tempDir)
		shelvedLf.Locked = true

		existingShelvedFiles := []*anime.LocalFile{shelvedLf}
		var existingLfs []*anime.LocalFile

		// File is shelved but we remove it
		_ = os.Remove(filePath)
		assert.DirExists(t, tempDir)

		scanner := &Scanner{
			DirPath:              tempDir,
			Enhanced:             false,
			PlatformRef:          util.NewRef(anilistPlatform),
			MetadataProviderRef:  util.NewRef(metadataProvider),
			Logger:               logger,
			WSEventManager:       wsEventManager,
			ExistingLocalFiles:   existingLfs,
			SkipLockedFiles:      true,
			SkipIgnoredFiles:     false,
			WithShelving:         true,
			ExistingShelvedFiles: existingShelvedFiles,
		}

		lfs, err := scanner.Scan(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 0, len(lfs), "Expected 0 local files found")

		// It should NOT be in ShelvedLocalFiles because the library path exists but the file is gone
		shelved := scanner.GetShelvedLocalFiles()
		assert.Equal(t, 0, len(shelved), "Expected 0 shelved files, file should be removed")
	})

	t.Run("Keep shelved file if library path missing", func(t *testing.T) {

		// non-existent library path
		missingDir := filepath.Join(tempDir, "missing_drive")
		missingFilePath := filepath.Join(missingDir, filename)

		// Prepare shelved file
		shelvedLf := anime.NewLocalFile(missingFilePath, missingDir)
		shelvedLf.Locked = true

		existingShelvedFiles := []*anime.LocalFile{shelvedLf}
		var existingLfs []*anime.LocalFile

		// library path is missing
		assert.NoDirExists(t, missingDir)

		scanner := &Scanner{
			DirPath:              tempDir,              // The main scan dir is tempDir
			OtherDirPaths:        []string{missingDir}, // We include the missing dir as a library path
			Enhanced:             false,
			PlatformRef:          util.NewRef(anilistPlatform),
			MetadataProviderRef:  util.NewRef(metadataProvider),
			Logger:               logger,
			WSEventManager:       wsEventManager,
			ExistingLocalFiles:   existingLfs,
			SkipLockedFiles:      true,
			SkipIgnoredFiles:     false,
			WithShelving:         true,
			ExistingShelvedFiles: existingShelvedFiles,
		}

		lfs, err := scanner.Scan(t.Context())
		if err != nil {
			t.Fatal(err)
		}

		// Verify results
		assert.Equal(t, 0, len(lfs), "Expected 0 local files found")

		// File should be kep shelved because the library path is missing
		shelved := scanner.GetShelvedLocalFiles()
		assert.Equal(t, 1, len(shelved), "Expected 1 shelved file")
		if len(shelved) > 0 {
			assert.Equal(t, missingFilePath, shelved[0].Path)
		}
	})
}
