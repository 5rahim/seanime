package offline

import "github.com/goccy/go-json"

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

	// Delete the previous snapshots
	if db.HasSnapshots() {
		snapshots, err := db.GetSnapshots()
		db.logger.Debug().Msgf("offline hub: Deleting %d snapshot(s)", len(snapshots))
		if err == nil {
			for _, snapshot := range snapshots {
				_ = db.DeleteSnapshot(snapshot.ID)
			}
		}
	}

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
	err := db.gormdb.Delete(&SnapshotEntry{}, id).Error

	if err == nil {
		_ = db.DeleteSnapshotMediaEntries(id)
	}

	return err
}

func (db *database) GetLatestSnapshot() (snapshot *SnapshotEntry, err error) {
	err = db.gormdb.Last(&snapshot).Error
	return
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

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (sme *SnapshotMediaEntry) GetAnimeEntry() *AnimeEntry {
	var entry *AnimeEntry
	_ = json.Unmarshal(sme.Value, &entry)
	return entry
}

func (sme *SnapshotMediaEntry) GetMangaEntry() *MangaEntry {
	var entry *MangaEntry
	_ = json.Unmarshal(sme.Value, &entry)
	return entry
}
