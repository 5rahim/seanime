package offline

import "seanime/internal/database/models"

type SnapshotEntry struct {
	models.BaseModel
	User        []byte `gorm:"column:user" json:"user"`
	Collections []byte `gorm:"column:collections" json:"collections"`
	AssetMap    []byte `gorm:"column:asset_map" json:"assetMap"`
	Synced      bool   `gorm:"column:synced" json:"synced"`
	Used        bool   `gorm:"column:used" json:"used"`
}

type SnapshotMediaEntryType string

const (
	SnapshotMediaEntryTypeAnime SnapshotMediaEntryType = "anime"
	SnapshotMediaEntryTypeManga SnapshotMediaEntryType = "manga"
)

type SnapshotMediaEntry struct {
	models.BaseModel
	Type           string `gorm:"column:type" json:"type"`
	MediaId        int    `gorm:"column:media_id" json:"mediaId"`
	SnapshotItemId uint   `gorm:"column:snapshot_item_id" json:"snapshotItemId"`
	Value          []byte `gorm:"column:value" json:"value"`
}
