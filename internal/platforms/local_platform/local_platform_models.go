package local_platform

import (
	"github.com/goccy/go-json"
	"seanime/internal/api/anilist"
	"time"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type LocalCollection struct {
	BaseModel
	Type  string `gorm:"column:type" json:"type"`   // "anime" or "manga"
	Value []byte `gorm:"column:value" json:"value"` // Marshalled struct
}

func (ldb *LocalPlatformDatabase) getLocalCollection(collectionType string) (*LocalCollection, bool) {
	var lc LocalCollection
	err := ldb.gormdb.Where("type = ?", collectionType).First(&lc).Error
	return &lc, err == nil
}

func (ldb *LocalPlatformDatabase) saveLocalCollection(collectionType string, value interface{}) error {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	lc := LocalCollection{
		BaseModel: BaseModel{
			ID: 1,
		},
		Type:  collectionType,
		Value: marshalledValue,
	}

	return ldb.gormdb.Save(&lc).Error
}

func (ldb *LocalPlatformDatabase) getLocalAnimeCollection() (*anilist.AnimeCollection, bool) {
	lc, ok := ldb.getLocalCollection("anime")
	if !ok {
		return nil, false
	}

	var ac anilist.AnimeCollection
	err := json.Unmarshal(lc.Value, &ac)

	return &ac, err == nil
}

func (ldb *LocalPlatformDatabase) getLocalMangaCollection() (*anilist.MangaCollection, bool) {
	lc, ok := ldb.getLocalCollection("manga")
	if !ok {
		return nil, false
	}

	var mc anilist.MangaCollection
	err := json.Unmarshal(lc.Value, &mc)

	return &mc, err == nil
}
