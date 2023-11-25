package db

import (
	"errors"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime-server/internal/anify"
	"github.com/seanime-app/seanime-server/internal/models"
	"gorm.io/gorm/clause"
)

func (db *Database) UpsertAnifyMediaEpisodeImages(entry *anify.MediaEpisodeImagesEntry) error {

	marshaled, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	sch := &models.AnifyMediaEpisodeImages{
		BaseModel: models.BaseModel{
			ID: uint(entry.MediaId),
		},
		Value: marshaled,
	}

	err = db.gormdb.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(sch).Error

	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetAnifyMediaEpisodeImages(mId int) (*anify.MediaEpisodeImagesEntry, error) {
	var r models.AnifyMediaEpisodeImages
	err := db.gormdb.Model(&r).Where("id = ?", mId).Error
	if err != nil {
		return nil, err
	}
	if r.Value == nil {
		return nil, errors.New("media episode images entry does not exist")
	}

	var entry *anify.MediaEpisodeImagesEntry
	if err := json.Unmarshal(r.Value, &entry); err != nil {
		return nil, err
	}

	return entry, nil
}

func (db *Database) GetAnifyMediaEpisodeImagesEntries() ([]*anify.MediaEpisodeImagesEntry, error) {
	var el []*models.AnifyMediaEpisodeImages
	err := db.gormdb.Model(&el).Error
	if err != nil {
		return nil, err
	}

	ret := make([]*anify.MediaEpisodeImagesEntry, 0)

	for _, r := range el {
		if r.Value == nil {
			return nil, errors.New("media episode images entry does not exist")
		}

		var entry *anify.MediaEpisodeImagesEntry
		if err := json.Unmarshal(r.Value, &entry); err != nil {
			return nil, err
		}

		ret = append(ret, entry)
	}

	return ret, nil
}
