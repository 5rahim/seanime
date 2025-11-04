package test

import (
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"
)

// TestDatabaseFileBasedTrimConcurrency tests the concurrent execution of database trim operations
// using a file-based SQLite database to better reproduce the real-world locking issue
func TestDatabaseFileBasedTrimConcurrency(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "seanime_test_db")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logger := util.NewLogger()

	// Create test database (file-based, not in-memory)
	database, err := db.NewDatabase(tempDir, "test_file", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Check what tables were created
	t.Log("Checking database schema...")
	checkDatabaseSchema(t, database)

	// Populate test data
	t.Log("Populating database with test data...")
	populateFileTestData(t, database)
	t.Log("Database population completed")

	// Test concurrent trim operations (simulating app startup)
	var wg sync.WaitGroup
	errorChan := make(chan error, 3)

	// Start all three trim operations concurrently
	wg.Add(3)

	go func() {
		defer wg.Done()
		t.Log("Starting TrimScanSummaryEntries...")
		database.TrimScanSummaryEntries()
		// Wait a bit to ensure the goroutine inside TrimScanSummaryEntries completes
		time.Sleep(500 * time.Millisecond)
		t.Log("TrimScanSummaryEntries completed")
	}()

	go func() {
		defer wg.Done()
		t.Log("Starting TrimLocalFileEntries...")
		database.TrimLocalFileEntries()
		time.Sleep(500 * time.Millisecond)
		t.Log("TrimLocalFileEntries completed")
	}()

	go func() {
		defer wg.Done()
		t.Log("Starting TrimTorrentstreamHistory...")
		database.TrimTorrentstreamHistory()
		time.Sleep(500 * time.Millisecond)
		t.Log("TrimTorrentstreamHistory completed")
	}()

	// Wait for all operations to complete
	wg.Wait()

	// Additional wait to ensure all internal goroutines complete
	time.Sleep(1000 * time.Millisecond)

	// Check if any errors occurred
	select {
	case err := <-errorChan:
		t.Errorf("Database operation failed: %v", err)
	default:
		t.Log("All database trim operations completed successfully")
	}

	// Verify the database file exists and has data
	dbPath := filepath.Join(tempDir, "test_file.db")
	if stat, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file does not exist: %v", err)
	} else {
		t.Logf("Database file size: %d bytes", stat.Size())
	}
}

// checkDatabaseSchema inspects the database schema to see what tables exist
func checkDatabaseSchema(t *testing.T, database *db.Database) {
	var tables []string
	err := database.Gorm().Raw("SELECT name FROM sqlite_master WHERE type='table'").Scan(&tables).Error
	if err != nil {
		t.Errorf("Failed to get table names: %v", err)
		return
	}

	t.Logf("Database tables: %v", tables)

	// Check specific table counts
	for _, table := range tables {
		var count int64
		err := database.Gorm().Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count).Error
		if err != nil {
			t.Logf("Failed to count records in table %s: %v", table, err)
		} else {
			t.Logf("Table %s has %d records", table, count)
		}
	}
}

// populateFileTestData creates test data to trigger the trim operations
func populateFileTestData(t *testing.T, database *db.Database) {
	// Create scan summaries (more than 10 to trigger trim)
	t.Log("Creating scan summaries...")
	for i := 0; i < 50; i++ {
		scanSummary := &models.ScanSummary{
			Value: []byte(fmt.Sprintf("test scan summary data %d - %s", i, generateTestData(100))),
		}
		err := database.Gorm().Create(scanSummary).Error
		if err != nil {
			t.Fatalf("Failed to create scan summary: %v", err)
		}
	}

	// Create local files (more than 10 to trigger trim)
	t.Log("Creating local files...")
	for i := 0; i < 50; i++ {
		localFiles := &models.LocalFiles{
			Value: []byte(fmt.Sprintf("test local files data %d - %s", i, generateTestData(100))),
		}
		err := database.Gorm().Create(localFiles).Error
		if err != nil {
			t.Fatalf("Failed to create local files: %v", err)
		}
	}

	// Create torrent stream history (more than 50 to trigger trim)
	t.Log("Creating torrent stream history...")
	for i := 0; i < 100; i++ {
		history := &models.TorrentstreamHistory{
			MediaId: i % 10,
			Torrent: []byte(fmt.Sprintf("test torrent data %d - %s", i, generateTestData(200))),
		}
		err := database.Gorm().Create(history).Error
		if err != nil {
			t.Fatalf("Failed to create torrent stream history: %v", err)
		}
	}

	t.Log("Test data population completed")
}

// generateTestData creates test data of specified length
func generateTestData(length int) string {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	return string(data)
}
