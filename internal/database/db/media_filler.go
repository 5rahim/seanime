package db

import (
	"github.com/goccy/go-json"
	"github.com/samber/mo"
	"seanime/internal/api/filler"
	"seanime/internal/database/models"
	"time"
)

type MediaFillerItem struct {
	DbId           uint      `json:"dbId"`
	Provider       string    `json:"provider"`
	Slug           string    `json:"slug"`
	MediaId        int       `json:"mediaId"`
	LastFetchedAt  time.Time `json:"lastFetchedAt"`
	FillerEpisodes []string  `json:"fillerEpisodes"`
}

// GetCachedMediaFillers will return all the media fillers (cache-first).
// If the cache is empty, it will fetch the media fillers from the database.
func (db *Database) GetCachedMediaFillers() (map[int]*MediaFillerItem, error) {

	if db.CurrMediaFillers.IsPresent() {
		return db.CurrMediaFillers.MustGet(), nil
	}

	var res []*models.MediaFiller
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return nil, err
	}

	// Unmarshal the media fillers
	mediaFillers := make(map[int]*MediaFillerItem)
	for _, mf := range res {

		var fillerData filler.Data
		if err := json.Unmarshal(mf.Data, &fillerData); err != nil {
			return nil, err
		}

		// Get the filler episodes
		var fillerEpisodes []string
		if fillerData.FillerEpisodes != nil || len(fillerData.FillerEpisodes) > 0 {
			fillerEpisodes = fillerData.FillerEpisodes
		}

		mediaFillers[mf.MediaID] = &MediaFillerItem{
			DbId:           mf.ID,
			Provider:       mf.Provider,
			MediaId:        mf.MediaID,
			Slug:           mf.Slug,
			LastFetchedAt:  mf.LastFetchedAt,
			FillerEpisodes: fillerEpisodes,
		}
	}

	// Cache the media fillers
	db.CurrMediaFillers = mo.Some(mediaFillers)

	return db.CurrMediaFillers.MustGet(), nil
}

func (db *Database) GetMediaFillerItem(mediaId int) (*MediaFillerItem, bool) {

	mediaFillers, err := db.GetCachedMediaFillers()
	if err != nil {
		return nil, false
	}

	item, ok := mediaFillers[mediaId]

	return item, ok
}

func (db *Database) InsertMediaFiller(
	provider string,
	mediaId int,
	slug string,
	lastFetchedAt time.Time,
	fillerEpisodes []string,
) error {

	// Marshal the filler data
	fillerData := filler.Data{
		FillerEpisodes: fillerEpisodes,
	}

	fillerDataBytes, err := json.Marshal(fillerData)
	if err != nil {
		return err
	}

	// Delete the existing media filler
	_ = db.DeleteMediaFiller(mediaId)

	// Save the media filler
	err = db.gormdb.Create(&models.MediaFiller{
		Provider:      provider,
		MediaID:       mediaId,
		Slug:          slug,
		LastFetchedAt: lastFetchedAt,
		Data:          fillerDataBytes,
	}).Error
	if err != nil {
		return err
	}

	// Update the cache
	db.CurrMediaFillers = mo.None[map[int]*MediaFillerItem]()

	return nil
}

// SaveCachedMediaFillerItems will save the cached media filler items in the database.
// Call this function after editing the cached media filler items.
func (db *Database) SaveCachedMediaFillerItems() error {

	if db.CurrMediaFillers.IsAbsent() {
		return nil
	}

	mediaFillers, err := db.GetCachedMediaFillers()
	if err != nil {
		return err
	}

	for _, mf := range mediaFillers {
		if len(mf.FillerEpisodes) == 0 {
			continue
		}
		// Marshal the filler data
		fillerData := filler.Data{
			FillerEpisodes: mf.FillerEpisodes,
		}

		fillerDataBytes, err := json.Marshal(fillerData)
		if err != nil {
			return err
		}

		// Save the media filler
		err = db.gormdb.Model(&models.MediaFiller{}).
			Where("id = ?", mf.DbId).
			Updates(map[string]interface{}{
				"last_fetched_at": mf.LastFetchedAt,
				"data":            fillerDataBytes,
			}).Error
		if err != nil {
			return err
		}
	}

	// Update the cache
	db.CurrMediaFillers = mo.None[map[int]*MediaFillerItem]()

	return nil
}

func (db *Database) DeleteMediaFiller(mediaId int) error {

	mediaFillers, err := db.GetCachedMediaFillers()
	if err != nil {
		return err
	}

	item, ok := mediaFillers[mediaId]
	if !ok {
		return nil
	}

	err = db.gormdb.Delete(&models.MediaFiller{}, item.DbId).Error
	if err != nil {
		return err
	}

	// Update the cache
	db.CurrMediaFillers = mo.None[map[int]*MediaFillerItem]()

	return nil
}
