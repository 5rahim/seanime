package test

import (
	"os"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"
)

// TestDatabaseTrimConcurrency tests the concurrent execution of database trim operations
// to reproduce the SQLite lock issue that occurs during app startup
func TestDatabaseTrimConcurrency(t *testing.T) {
	// Set test environment
	os.Setenv("TEST_ENV", "true")
	defer os.Unsetenv("TEST_ENV")

	logger := util.NewLogger()

	// Create test database
	database, err := db.NewDatabase("", "test", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Populate test data
	populateTestData(t, database)

	// Test concurrent trim operations (simulating app startup)
	var wg sync.WaitGroup
	errors := make(chan error, 3)

	// Start all three trim operations concurrently
	wg.Add(3)

	go func() {
		defer wg.Done()
		database.TrimScanSummaryEntries()
		// Wait a bit to ensure the goroutine inside TrimScanSummaryEntries completes
		time.Sleep(100 * time.Millisecond)
	}()

	go func() {
		defer wg.Done()
		database.TrimLocalFileEntries()
		time.Sleep(100 * time.Millisecond)
	}()

	go func() {
		defer wg.Done()
		database.TrimTorrentstreamHistory()
		time.Sleep(100 * time.Millisecond)
	}()

	// Wait for all operations to complete
	wg.Wait()

	// Additional wait to ensure all internal goroutines complete
	time.Sleep(200 * time.Millisecond)

	// Check if any errors occurred (this test should pass after the fix)
	select {
	case err := <-errors:
		t.Errorf("Database operation failed: %v", err)
	default:
		t.Log("All database trim operations completed successfully")
	}
}

// populateTestData creates test data to trigger the trim operations
func populateTestData(t *testing.T, database *db.Database) {
	// Create scan summaries (more than 10 to trigger trim)
	for i := 0; i < 15; i++ {
		scanSummary := &models.ScanSummary{
			Value: []byte("test scan summary data"),
		}
		err := database.Gorm().Create(scanSummary).Error
		if err != nil {
			t.Fatalf("Failed to create scan summary: %v", err)
		}
	}

	// Create local files (more than 10 to trigger trim)
	for i := 0; i < 15; i++ {
		localFiles := &models.LocalFiles{
			Value: []byte("test local files data"),
		}
		err := database.Gorm().Create(localFiles).Error
		if err != nil {
			t.Fatalf("Failed to create local files: %v", err)
		}
	}

	// Create torrent stream history (more than 50 to trigger trim)
	for i := 0; i < 60; i++ {
		history := &models.TorrentstreamHistory{
			MediaId: 1,
			Torrent: []byte("test torrent data"),
		}
		err := database.Gorm().Create(history).Error
		if err != nil {
			t.Fatalf("Failed to create torrent stream history: %v", err)
		}
	}
}
