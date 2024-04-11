package offline

func (db *database) HasSnapshots() bool {
	var count int64
	db.gormdb.Model(&SnapshotEntry{}).Count(&count)
	return count > 0
}

func (db *database) GetSnapshots() ([]*SnapshotEntry, error) {
	var snapshots []*SnapshotEntry
	err := db.gormdb.Find(&snapshots).Error
	return snapshots, err
}

func (db *database) GetSnapshot(id uint) (*SnapshotEntry, error) {
	var snapshot SnapshotEntry
	err := db.gormdb.First(&snapshot, id).Error
	return &snapshot, err
}

func (db *database) InsertSnapshot(user []byte, collections []byte, assetMap []byte) (*SnapshotEntry, error) {
	snapshot := &SnapshotEntry{
		User:        user,
		Collections: collections,
		AssetMap:    assetMap,
	}
	err := db.gormdb.Create(snapshot).Error
	return snapshot, err
}

func (db *database) UpdateSnapshot(id uint, collections []byte, assetMap []byte) (*SnapshotEntry, error) {
	snapshot := &SnapshotEntry{
		Collections: collections,
		AssetMap:    assetMap,
	}
	err := db.gormdb.Model(&SnapshotEntry{}).Where("id = ?", id).Updates(snapshot).Error
	return snapshot, err
}

func (db *database) DeleteSnapshot(id uint) error {
	return db.gormdb.Delete(&SnapshotEntry{}, id).Error
}

func (db *database) GetLatestSnapshot() (*SnapshotEntry, error) {
	var snapshot SnapshotEntry
	err := db.gormdb.Last(&snapshot).Error
	return &snapshot, err
}

//
// SnapshotMediaEntry
//

func (db *database) InsertSnapshotMediaEntry(snapshotItemId uint, t SnapshotMediaEntryType, mediaId int, value []byte) (*SnapshotMediaEntry, error) {
	entry := &SnapshotMediaEntry{
		MediaId:        mediaId,
		SnapshotItemId: snapshotItemId,
		Value:          value,
		Type:           string(t),
	}
	err := db.gormdb.Create(entry).Error
	return entry, err
}

func (db *database) GetSnapshotMediaEntry(mediaId int, snapshotItemId uint) (*SnapshotMediaEntry, error) {
	var entry SnapshotMediaEntry
	err := db.gormdb.Where("media_id = ? AND snapshot_item_id = ?", mediaId, snapshotItemId).First(&entry).Error
	return &entry, err
}

func (db *database) UpdateSnapshotMediaEntry(mediaId int, snapshotItemId uint, value []byte) (*SnapshotMediaEntry, error) {
	entry := &SnapshotMediaEntry{
		MediaId:        mediaId,
		SnapshotItemId: snapshotItemId,
		Value:          value,
	}
	err := db.gormdb.Model(&SnapshotMediaEntry{}).Where("media_id = ? AND snapshot_item_id = ?", mediaId, snapshotItemId).Updates(entry).Error
	return entry, err
}

func (db *database) DeleteSnapshotMediaEntry(mediaId int, snapshotItemId uint) error {
	return db.gormdb.Where("media_id = ? AND snapshot_item_id = ?", mediaId, snapshotItemId).Delete(&SnapshotMediaEntry{}).Error
}

func (db *database) GetSnapshotMediaEntries(snapshotItemId uint) ([]*SnapshotMediaEntry, error) {
	var entries []*SnapshotMediaEntry
	err := db.gormdb.Where("snapshot_item_id = ?", snapshotItemId).Find(&entries).Error
	return entries, err
}

func (db *database) DeleteSnapshotMediaEntries(snapshotItemId uint) error {
	return db.gormdb.Where("snapshot_item_id = ?", snapshotItemId).Delete(&SnapshotMediaEntry{}).Error
}
