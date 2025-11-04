package test

import (
	"crypto/rand"
	"fmt"
	"os"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"sync"
	"testing"
	"time"
)

// TestDatabaseLargeDataTrimConcurrency tests the concurrent execution of database trim operations
// with a large dataset (400MB) to reproduce the SQLite lock issue that occurs during app startup
func TestDatabaseLargeDataTrimConcurrency(t *testing.T) {
	// Set test environment
	os.Setenv("TEST_ENV", "true")
	defer os.Unsetenv("TEST_ENV")

	logger := util.NewLogger()

	// Create test database
	database, err := db.NewDatabase("", "test_large", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Populate test data with ~400MB of synthetic data
	t.Log("Populating database with 400MB of synthetic data...")
	populateLargeTestData(t, database)
	t.Log("Database population completed")

	// Test concurrent trim operations (simulating app startup)
	var wg sync.WaitGroup
	errors := make(chan error, 3)

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

	// Check if any errors occurred (this test should pass after the fix)
	select {
	case err := <-errors:
		t.Errorf("Database operation failed: %v", err)
	default:
		t.Log("All database trim operations completed successfully")
	}
}

// populateLargeTestData creates ~400MB of test data to trigger the trim operations
func populateLargeTestData(t *testing.T, database *db.Database) {
	// Target: ~400MB of data
	// Each record will be roughly 1KB, so we need ~400,000 records total

	// Create scan summaries (100,000 records, ~100MB)
	t.Log("Creating scan summaries...")
	for i := 0; i < 100000; i++ {
		scanSummary := &models.ScanSummary{
			Value: generateLargeData(1024), // 1KB per record
		}
		err := database.Gorm().Create(scanSummary).Error
		if err != nil {
			t.Fatalf("Failed to create scan summary %d: %v", i, err)
		}
		if i%10000 == 0 {
			t.Logf("Created %d scan summaries", i)
		}
	}

	// Create local files (100,000 records, ~100MB)
	t.Log("Creating local files...")
	for i := 0; i < 100000; i++ {
		localFiles := &models.LocalFiles{
			Value: generateLargeData(1024), // 1KB per record
		}
		err := database.Gorm().Create(localFiles).Error
		if err != nil {
			t.Fatalf("Failed to create local files %d: %v", i, err)
		}
		if i%10000 == 0 {
			t.Logf("Created %d local files", i)
		}
	}

	// Create torrent stream history (200,000 records, ~200MB)
	t.Log("Creating torrent stream history...")
	for i := 0; i < 200000; i++ {
		history := &models.TorrentstreamHistory{
			MediaId: i % 1000,                // Vary media IDs
			Torrent: generateLargeData(1024), // 1KB per record
		}
		err := database.Gorm().Create(history).Error
		if err != nil {
			t.Fatalf("Failed to create torrent stream history %d: %v", i, err)
		}
		if i%20000 == 0 {
			t.Logf("Created %d torrent stream history records", i)
		}
	}

	// Add some additional data types to increase database complexity
	t.Log("Creating additional data...")

	// Create media fillers
	for i := 0; i < 10000; i++ {
		mediaFiller := &models.MediaFiller{
			Provider:      "test_provider",
			Slug:          fmt.Sprintf("test-slug-%d", i),
			MediaID:       i,
			LastFetchedAt: time.Now(),
			Data:          generateLargeData(512), // 512 bytes per record
		}
		err := database.Gorm().Create(mediaFiller).Error
		if err != nil {
			t.Fatalf("Failed to create media filler %d: %v", i, err)
		}
	}

	// Create auto downloader items
	for i := 0; i < 10000; i++ {
		autoDownloaderItem := &models.AutoDownloaderItem{
			RuleID:      uint(i % 100),
			MediaID:     i,
			Episode:     i % 24,
			Link:        fmt.Sprintf("https://example.com/torrent/%d", i),
			Hash:        fmt.Sprintf("hash_%d_%s", i, generateRandomString(32)),
			Magnet:      fmt.Sprintf("magnet:?xt=urn:btih:%s", generateRandomString(40)),
			TorrentName: fmt.Sprintf("Test Torrent %d", i),
			Downloaded:  i%2 == 0,
		}
		err := database.Gorm().Create(autoDownloaderItem).Error
		if err != nil {
			t.Fatalf("Failed to create auto downloader item %d: %v", i, err)
		}
	}

	t.Log("Large test data population completed")
}

// generateLargeData creates a byte slice of the specified size filled with random data
func generateLargeData(size int) []byte {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		// Fallback to deterministic data if random fails
		for i := range data {
			data[i] = byte(i % 256)
		}
	}
	return data
}

// generateRandomString creates a random string of the specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
