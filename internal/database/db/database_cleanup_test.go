package db

import (
	"fmt"
	"os"
	"path/filepath"
	"seanime/internal/database/models"
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/dustin/go-humanize"
)

func TestDatabaseCleanupManager(t *testing.T) {
	tempDir := t.TempDir()
	logger := util.NewLogger()

	database, err := NewDatabase(tempDir, "cleanup_test", logger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	t.Log("Populating database with test data...")
	populateCleanupTestData(t, database)

	checkTableCounts(t, database, "before cleanup")

	dbPath := filepath.Join(tempDir, "cleanup_test.db")
	if stat, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file does not exist: %v", err)
	} else {
		t.Logf("Database file size: %s bytes", humanize.Bytes(uint64(stat.Size())))
	}

	t.Log("Running database cleanup...")
	database.RunDatabaseCleanup()

	// DEVNOTE: Locking issues occur when running this in parallel to many writes
	//go func() {
	//database.TrimLocalFileEntries()
	//database.TrimTorrentstreamHistory()
	//database.TrimScanSummaryEntries()
	//}()

	t.Log("Launching many write operations...")
	time.Sleep(100 * time.Millisecond)
	for i := 0; i < 1000; i++ {
		go database.Gorm().Create(&models.ScanSummary{Value: []byte(fmt.Sprintf("scan summary data %d - %s", i, generateCleanupTestData(5000)))})
	}

	time.Sleep(20 * time.Second)

	checkTableCounts(t, database, "after cleanup")

	if stat, err := os.Stat(dbPath); err != nil {
		t.Errorf("Database file does not exist: %v", err)
	} else {
		t.Logf("Database file size: %s bytes", humanize.Bytes(uint64(stat.Size())))
	}

	t.Log("Cleanup manager test completed successfully")
}

func populateCleanupTestData(t *testing.T, database *Database) {
	for i := 0; i < 10000; i++ {
		scanSummary := &models.ScanSummary{
			Value: []byte(fmt.Sprintf("scan summary data %d - %s", i, generateCleanupTestData(50000))),
		}
		err := database.Gorm().Create(scanSummary).Error
		if err != nil {
			t.Fatalf("Failed to create scan summary: %v", err)
		}
	}

	for i := 0; i < 10000; i++ {
		localFiles := &models.LocalFiles{
			Value: []byte(fmt.Sprintf("local files data %d - %s", i, generateCleanupTestData(500000))),
		}
		err := database.Gorm().Create(localFiles).Error
		if err != nil {
			t.Fatalf("Failed to create local files: %v", err)
		}
	}

	for i := 0; i < 1000; i++ {
		history := &models.TorrentstreamHistory{
			MediaId: i % 10,
			Torrent: []byte(fmt.Sprintf("torrent data %d - %s", i, generateCleanupTestData(50000))),
		}
		err := database.Gorm().Create(history).Error
		if err != nil {
			t.Fatalf("Failed to create torrent stream history: %v", err)
		}
	}
}

func checkTableCounts(t *testing.T, database *Database, phase string) {
	var scanCount, localCount, torrentCount int64

	err := database.Gorm().Model(&models.ScanSummary{}).Count(&scanCount).Error
	if err != nil {
		t.Errorf("Failed to count scan summaries: %v", err)
	}

	err = database.Gorm().Model(&models.LocalFiles{}).Count(&localCount).Error
	if err != nil {
		t.Errorf("Failed to count local files: %v", err)
	}

	err = database.Gorm().Model(&models.TorrentstreamHistory{}).Count(&torrentCount).Error
	if err != nil {
		t.Errorf("Failed to count torrent stream history: %v", err)
	}

	t.Logf("Record counts %s: ScanSummary=%d, LocalFiles=%d, TorrentstreamHistory=%d",
		phase, scanCount, localCount, torrentCount)

	if phase == "after cleanup" {
		//if scanCount > 10 {
		//	t.Errorf("Expected scan summaries to be trimmed to ≤10, got %d", scanCount)
		//}
		if localCount > 10 {
			t.Errorf("Expected local files to be trimmed to ≤10, got %d", localCount)
		}
		if torrentCount > 50 {
			t.Errorf("Expected torrent stream history to be trimmed to ≤50, got %d", torrentCount)
		}

		//var minScanSummary models.ScanSummary
		//if err := database.Gorm().Order("id asc").First(&minScanSummary).Error; err != nil {
		//	t.Errorf("Failed to get min scan summary: %v", err)
		//} else if minScanSummary.ID != 996 {
		//	t.Errorf("Expected min scan summary ID to be 991, got %d", minScanSummary.ID)
		//}

		var minLocalFiles models.LocalFiles
		if err := database.Gorm().Order("id asc").First(&minLocalFiles).Error; err != nil {
			t.Errorf("Failed to get min local files: %v", err)
		} else if minLocalFiles.ID != 9996 {
			t.Errorf("Expected min local files ID to be 991, got %d", minLocalFiles.ID)
		}

		var minTorrentHistory models.TorrentstreamHistory
		if err := database.Gorm().Order("id asc").First(&minTorrentHistory).Error; err != nil {
			t.Errorf("Failed to get min torrent history: %v", err)
		} else if minTorrentHistory.ID != 961 {
			t.Errorf("Expected min torrent history ID to be 951, got %d", minTorrentHistory.ID)
		}
	}
}

func generateCleanupTestData(length int) string {
	data := make([]byte, length)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	return string(data)
}
