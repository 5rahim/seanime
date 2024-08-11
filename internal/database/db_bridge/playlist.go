package db_bridge

import (
	"github.com/goccy/go-json"
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"
)

func GetPlaylists(db *db.Database) ([]*anime.Playlist, error) {
	var res []*models.PlaylistEntry
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	playlists := make([]*anime.Playlist, 0)
	for _, p := range res {
		var localFiles []*anime.LocalFile
		if err := json.Unmarshal(p.Value, &localFiles); err == nil {
			playlist := anime.NewPlaylist(p.Name)
			playlist.SetLocalFiles(localFiles)
			playlist.DbId = p.ID
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

func SavePlaylist(db *db.Database, playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.LocalFiles)
	if err != nil {
		return err
	}
	playlistEntry := &models.PlaylistEntry{
		Name:  playlist.Name,
		Value: data,
	}

	return db.Gorm().Save(playlistEntry).Error
}

func DeletePlaylist(db *db.Database, id uint) error {
	return db.Gorm().Where("id = ?", id).Delete(&models.PlaylistEntry{}).Error
}

func UpdatePlaylist(db *db.Database, playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.LocalFiles)
	if err != nil {
		return err
	}

	// Get the playlist entry
	playlistEntry := &models.PlaylistEntry{}
	if err := db.Gorm().Where("id = ?", playlist.DbId).First(playlistEntry).Error; err != nil {
		return err
	}

	// Update the playlist entry
	playlistEntry.Name = playlist.Name
	playlistEntry.Value = data

	return db.Gorm().Save(playlistEntry).Error
}

func GetPlaylist(db *db.Database, id uint) (*anime.Playlist, error) {
	playlistEntry := &models.PlaylistEntry{}
	if err := db.Gorm().Where("id = ?", id).First(playlistEntry).Error; err != nil {
		return nil, err
	}

	var localFiles []*anime.LocalFile
	if err := json.Unmarshal(playlistEntry.Value, &localFiles); err != nil {
		return nil, err
	}

	playlist := anime.NewPlaylist(playlistEntry.Name)
	playlist.SetLocalFiles(localFiles)
	playlist.DbId = playlistEntry.ID

	return playlist, nil
}
