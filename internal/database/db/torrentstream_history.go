package db

func (db *Database) TrimTorrentstreamHistory() {
	// Use the cleanup manager to avoid concurrent access issues
	db.cleanupManager.trimTorrentstreamHistory()
}
