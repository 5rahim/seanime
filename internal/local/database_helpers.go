package local

import (
	"seanime/internal/api/anilist"

	"github.com/goccy/go-json"
)

var CurrSettings *Settings

func (ldb *Database) SaveSettings(s *Settings) error {
	s.BaseModel.ID = 1
	CurrSettings = nil
	return ldb.gormdb.Save(s).Error
}

func (ldb *Database) GetSettings() *Settings {
	if CurrSettings != nil {
		return CurrSettings
	}
	var s Settings
	err := ldb.gormdb.First(&s).Error
	if err != nil {
		_ = ldb.SaveSettings(&Settings{
			BaseModel: BaseModel{
				ID: 1,
			},
			Updated: false,
		})
		return &Settings{
			BaseModel: BaseModel{
				ID: 1,
			},
			Updated: false,
		}
	}
	return &s
}

func (ldb *Database) SetTrackedMedia(sm *TrackedMedia) error {
	return ldb.gormdb.Save(sm).Error
}

// GetTrackedMedia returns the tracked media with the given mediaId and kind.
// This should only be used when adding/removing tracked media.
func (ldb *Database) GetTrackedMedia(mediaId int, kind string) (*TrackedMedia, bool) {
	var sm TrackedMedia
	err := ldb.gormdb.Where("media_id = ? AND type = ?", mediaId, kind).First(&sm).Error
	return &sm, err == nil
}

func (ldb *Database) GetAllTrackedMediaByType(kind string) ([]*TrackedMedia, bool) {
	var sm []*TrackedMedia
	err := ldb.gormdb.Where("type = ?", kind).Find(&sm).Error
	return sm, err == nil
}

func (ldb *Database) GetAllTrackedMedia() ([]*TrackedMedia, bool) {
	var sm []*TrackedMedia
	err := ldb.gormdb.Find(&sm).Error
	return sm, err == nil
}

func (ldb *Database) RemoveTrackedMedia(mediaId int, kind string) error {
	return ldb.gormdb.Where("media_id = ? AND type = ?", mediaId, kind).Delete(&TrackedMedia{}).Error
}

//----------------------------------------------------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------------------------------------------------

func (ldb *Database) SaveAnimeSnapshot(as *AnimeSnapshot) error {
	return ldb.gormdb.Save(as).Error
}

func (ldb *Database) GetAnimeSnapshot(mediaId int) (*AnimeSnapshot, bool) {
	var as AnimeSnapshot
	err := ldb.gormdb.Where("media_id = ?", mediaId).First(&as).Error
	return &as, err == nil
}

func (ldb *Database) RemoveAnimeSnapshot(mediaId int) error {
	return ldb.gormdb.Where("media_id = ?", mediaId).Delete(&AnimeSnapshot{}).Error
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (ldb *Database) SaveMangaSnapshot(ms *MangaSnapshot) error {
	return ldb.gormdb.Save(ms).Error
}

func (ldb *Database) GetMangaSnapshot(mediaId int) (*MangaSnapshot, bool) {
	var ms MangaSnapshot
	err := ldb.gormdb.Where("media_id = ?", mediaId).First(&ms).Error
	return &ms, err == nil
}

func (ldb *Database) RemoveMangaSnapshot(mediaId int) error {
	return ldb.gormdb.Where("media_id = ?", mediaId).Delete(&MangaSnapshot{}).Error
}

//----------------------------------------------------------------------------------------------------------------------------------------------------
//----------------------------------------------------------------------------------------------------------------------------------------------------

func (ldb *Database) GetAnimeSnapshots() ([]*AnimeSnapshot, bool) {
	var as []*AnimeSnapshot
	err := ldb.gormdb.Find(&as).Error
	return as, err == nil
}

func (ldb *Database) GetMangaSnapshots() ([]*MangaSnapshot, bool) {
	var ms []*MangaSnapshot
	err := ldb.gormdb.Find(&ms).Error
	return ms, err == nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (ldb *Database) SaveAnimeCollection(ac *anilist.AnimeCollection) error {
	return ldb._saveLocalCollection(AnimeType, ac)
}

func (ldb *Database) SaveMangaCollection(mc *anilist.MangaCollection) error {
	return ldb._saveLocalCollection(MangaType, mc)
}

func (ldb *Database) GetLocalAnimeCollection() (*anilist.AnimeCollection, bool) {
	lc, ok := ldb._getLocalCollection(AnimeType)
	if !ok {
		return nil, false
	}

	var ac anilist.AnimeCollection
	err := json.Unmarshal(lc.Value, &ac)

	return &ac, err == nil
}

func (ldb *Database) GetLocalMangaCollection() (*anilist.MangaCollection, bool) {
	lc, ok := ldb._getLocalCollection(MangaType)
	if !ok {
		return nil, false
	}

	var mc anilist.MangaCollection
	err := json.Unmarshal(lc.Value, &mc)

	return &mc, err == nil
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (ldb *Database) _getLocalCollection(collectionType string) (*LocalCollection, bool) {
	var lc LocalCollection
	err := ldb.gormdb.Where("type = ?", collectionType).First(&lc).Error
	return &lc, err == nil
}

func (ldb *Database) _saveLocalCollection(collectionType string, value interface{}) error {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Check if collection already exists
	lc, ok := ldb._getLocalCollection(collectionType)
	if ok {
		lc.Value = marshalledValue
		return ldb.gormdb.Save(&lc).Error
	}

	lcN := LocalCollection{
		Type:  collectionType,
		Value: marshalledValue,
	}

	return ldb.gormdb.Save(&lcN).Error
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Simulated collections
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (ldb *Database) _getSimulatedCollection(collectionType string) (*SimulatedCollection, bool) {
	var lc SimulatedCollection
	err := ldb.gormdb.Where("type = ?", collectionType).First(&lc).Error
	return &lc, err == nil
}

func (ldb *Database) _saveSimulatedCollection(collectionType string, value interface{}) error {

	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// Check if collection already exists
	lc, ok := ldb._getSimulatedCollection(collectionType)
	if ok {
		lc.Value = marshalledValue
		return ldb.gormdb.Save(&lc).Error
	}

	lcN := SimulatedCollection{
		Type:  collectionType,
		Value: marshalledValue,
	}

	return ldb.gormdb.Save(&lcN).Error
}

func (ldb *Database) SaveSimulatedAnimeCollection(ac *anilist.AnimeCollection) error {
	return ldb._saveSimulatedCollection(AnimeType, ac)
}

func (ldb *Database) SaveSimulatedMangaCollection(mc *anilist.MangaCollection) error {
	return ldb._saveSimulatedCollection(MangaType, mc)
}

func (ldb *Database) GetSimulatedAnimeCollection() (*anilist.AnimeCollection, bool) {
	lc, ok := ldb._getSimulatedCollection(AnimeType)
	if !ok {
		return nil, false
	}

	var ac anilist.AnimeCollection
	err := json.Unmarshal(lc.Value, &ac)

	return &ac, err == nil
}

func (ldb *Database) GetSimulatedMangaCollection() (*anilist.MangaCollection, bool) {
	lc, ok := ldb._getSimulatedCollection(MangaType)
	if !ok {
		return nil, false
	}

	var mc anilist.MangaCollection
	err := json.Unmarshal(lc.Value, &mc)

	return &mc, err == nil
}
