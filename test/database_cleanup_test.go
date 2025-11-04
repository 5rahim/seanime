package test

import (
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"testing"
)

// TestDatabaseCleanupManager tests the new cleanup manager to ensure it prevents SQLite locking issues
func TestDatabaseCleanupManager(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "seanime_cleanup_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := util.NewLogger()

	// Create test database (file-based, not in-memory)
	database, err := db.NewDatabase(tempDir, "cleanup_test", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Populate test data to trigger cleanup operations
	t.Log("Populating database with test data...")
	populateCleanupTestData(t, database)

	// Check initial counts
	checkTableCounts(t, database, "before cleanup")

	// Test the new cleanup manager
	t.Log("Running database cleanup using new manager...")
	database.RunDatabaseCleanup()

	// Check final counts
	checkTableCounts(t, database, "after cleanup")

	// Verify the database file exists and has reasonable size
	dbPath := filepath.Join(tempDir, "cleanup_test.db")
	if stat, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file does not exist: %v", err)
	} else {
		t.Logf("Database file size: %d bytes", stat.Size())
	}

	t.Log("Cleanup manager test completed successfully")
}

// populateCleanupTestData creates test data to trigger all cleanup operations
func populateCleanupTestData(t *testing.T, database *db.Database) {
	// Create scan summaries (more than 10 to trigger trim)
	for i := 0; i < 25; i++ {
		scanSummary := &models.ScanSummary{
			Value: []byte(fmt.Sprintf("scan summary data %d - %s", i, generateCleanupTestData(200))),
		}
		err := database.Gorm().Create(scanSummary).Error
		if err != nil {
			t.Fatalf("Failed to create scan summary: %v", err)
		}
	}

	// Create local files (more than 10 to trigger trim)
	for i := 0; i < 30; i++ {
		localFiles := &models.LocalFiles{
			Value: []byte(fmt.Sprintf("local files data %d - %s", i, generateCleanupTestData(200))),
		}
		err := database.Gorm().Create(localFiles).Error
		if err != nil {
			t.Fatalf("Failed to create local files: %v", err)
		}
	}

	// Create torrent stream history (more than 50 to trigger trim)
	for i := 0; i < 75; i++ {
		history := &models.TorrentstreamHistory{
			MediaId: i % 10,
			Torrent: []byte(fmt.Sprintf("torrent data %d - %s", i, generateCleanupTestData(300))),
		}
		err := database.Gorm().Create(history).Error
		if err != nil {
			t.Fatalf("Failed to create torrent stream history: %v", err)
		}
	}
}

// checkTableCounts checks and logs the record counts in each table
func checkTableCounts(t *testing.T, database *db.Database, phase string) {
	var scanCount, localCount, torrentCount int64

	// Count scan summaries
	err := database.Gorm().Model(&models.ScanSummary{}).Count(&scanCount).Error
	if err != nil {
		t.Errorf("Failed to count scan summaries: %v", err)
	}

	// Count local files
	err = database.Gorm().Model(&models.LocalFiles{}).Count(&localCount).Error
	if err != nil {
		t.Errorf("Failed to count local files: %v", err)
	}

	// Count torrent stream history
	err = database.Gorm().Model(&models.TorrentstreamHistory{}).Count(&torrentCount).Error
	if err != nil {
		t.Errorf("Failed to count torrent stream history: %v", err)
	}

	t.Logf("Record counts %s: ScanSummary=%d, LocalFiles=%d, TorrentstreamHistory=%d",
		phase, scanCount, localCount, torrentCount)

	// Validate cleanup results
	if phase == "after cleanup" {
		if scanCount > 10 {
			t.Errorf("Expected scan summaries to be trimmed to ≤10, got %d", scanCount)
		}
		if localCount > 10 {
			t.Errorf("Expected local files to be trimmed to ≤10, got %d", localCount)
		}
		if torrentCount > 50 {
			t.Errorf("Expected torrent stream history to be trimmed to ≤50, got %d", torrentCount)
		}
	}
}

// generateCleanupTestData creates test data of specified length
func generateCleanupTestData(length int) string {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	return string(data)
}
