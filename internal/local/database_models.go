package local

import (
	"database/sql/driver"
	"errors"
	"seanime/internal/api/metadata"
	"seanime/internal/manga"
	"time"

	"github.com/goccy/go-json"
)

type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Settings struct {
	BaseModel
	// Flag to determine if there are local changes that need to be synced with AniList.
	Updated bool `gorm:"column:updated" json:"updated"`
}

// +---------------------+
// |      Offline        |
// +---------------------+

// LocalCollection is an anilist collection that is stored locally for offline use.
// It is meant to be kept in sync with the real AniList collection when online.
type LocalCollection struct {
	BaseModel
	Type  string `gorm:"column:type" json:"type"`   // "anime" or "manga"
	Value []byte `gorm:"column:value" json:"value"` // Marshalled struct
}

// TrackedMedia tracks media that should be stored locally.
type TrackedMedia struct {
	BaseModel
	MediaId int    `gorm:"column:media_id" json:"mediaId"`
	Type    string `gorm:"column:type" json:"type"` // "anime" or "manga"
}

type AnimeSnapshot struct {
	BaseModel
	MediaId int `gorm:"column:media_id" json:"mediaId"`
	//ListEntry         LocalAnimeListEntry `gorm:"column:list_entry" json:"listEntry"`
	AnimeMetadata     LocalAnimeMetadata `gorm:"column:anime_metadata" json:"animeMetadata"`
	BannerImagePath   string             `gorm:"column:banner_image_path" json:"bannerImagePath"`
	CoverImagePath    string             `gorm:"column:cover_image_path" json:"coverImagePath"`
	EpisodeImagePaths StringMap          `gorm:"column:episode_image_paths" json:"episodeImagePaths"`
	// ReferenceKey is used to compare the snapshot with the current data.
	ReferenceKey string `gorm:"column:reference_key" json:"referenceKey"`
}

type MangaSnapshot struct {
	BaseModel
	MediaId int `gorm:"column:media_id" json:"mediaId"`
	//ListEntry         LocalMangaListEntry         `gorm:"column:list_entry" json:"listEntry"`
	ChapterContainers LocalMangaChapterContainers `gorm:"column:chapter_Containers" json:"chapterContainers"`
	BannerImagePath   string                      `gorm:"column:banner_image_path" json:"bannerImagePath"`
	CoverImagePath    string                      `gorm:"column:cover_image_path" json:"coverImagePath"`
	// ReferenceKey is used to compare the snapshot with the current data.
	ReferenceKey string `gorm:"column:reference_key" json:"referenceKey"`
}

// +---------------------+
// |      Simulated      |
// +---------------------+

// SimulatedCollection is used for users without an account.
type SimulatedCollection struct {
	BaseModel
	Type  string `gorm:"column:type" json:"type"`   // "anime" or "manga"
	Value []byte `gorm:"column:value" json:"value"` // Marshalled struct
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type StringMap map[string]string

func (o *StringMap) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	var ret map[string]string
	err := json.Unmarshal(bytes, &ret)
	if err != nil {
		return err
	}
	*o = ret
	return nil
}

func (o StringMap) Value() (driver.Value, error) {
	return json.Marshal(o)
}

type LocalAnimeMetadata metadata.AnimeMetadata

//----------------------------------------------------------------------------------------------------------------------------------------------------

func (o *LocalAnimeMetadata) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	var ret metadata.AnimeMetadata
	err := json.Unmarshal(bytes, &ret)
	if err != nil {
		return err
	}
	*o = LocalAnimeMetadata(ret)
	return nil
}

func (o LocalAnimeMetadata) Value() (driver.Value, error) {
	return json.Marshal(o)
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

type LocalMangaChapterContainers []*manga.ChapterContainer

func (o *LocalMangaChapterContainers) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return errors.New("src value cannot cast to []byte")
	}
	var ret []*manga.ChapterContainer
	err := json.Unmarshal(bytes, &ret)
	if err != nil {
		return err
	}
	*o = LocalMangaChapterContainers(ret)
	return nil
}

func (o LocalMangaChapterContainers) Value() (driver.Value, error) {
	return json.Marshal(o)
}

//----------------------------------------------------------------------------------------------------------------------------------------------------

//type LocalMangaListEntry anilist.MangaListEntry
//
//func (o *LocalMangaListEntry) Scan(src interface{}) error {
//	bytes, ok := src.([]byte)
//	if !ok {
//		return errors.New("src value cannot cast to []byte")
//	}
//	var ret anilist.MangaListEntry
//	err := json.Unmarshal(bytes, &ret)
//	if err != nil {
//		return err
//	}
//	*o = LocalMangaListEntry(ret)
//	return nil
//}
//
//func (o LocalMangaListEntry) Value() (driver.Value, error) {
//	if o.ID == 0 {
//		return nil, nil
//	}
//	return json.Marshal(o)
//}

//----------------------------------------------------------------------------------------------------------------------------------------------------

//type LocalAnimeListEntry anilist.AnimeListEntry
//
//func (o *LocalAnimeListEntry) Scan(src interface{}) error {
//	bytes, ok := src.([]byte)
//	if !ok {
//		return errors.New("src value cannot cast to []byte")
//	}
//	var ret anilist.AnimeListEntry
//	err := json.Unmarshal(bytes, &ret)
//	if err != nil {
//		return err
//	}
//	*o = LocalAnimeListEntry(ret)
//	return nil
//}
//
//func (o LocalAnimeListEntry) Value() (driver.Value, error) {
//	if o.ID == 0 {
//		return nil, nil
//	}
//	return json.Marshal(o)
//}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Local account
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
