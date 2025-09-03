package db

func (db *Database) TrimScanSummaryEntries() {
	// Use the cleanup manager to avoid concurrent access issues
	db.cleanupManager.trimScanSummaryEntries()
}
