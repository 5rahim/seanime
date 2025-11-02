package db_bridge

import (
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"

	"github.com/goccy/go-json"
)

func GetPlaylists(db *db.Database) ([]*anime.Playlist, error) {
	var res []*models.Playlist
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	playlists := make([]*anime.Playlist, 0)
	for _, p := range res {
		var eps []*anime.PlaylistEpisode
		if err := json.Unmarshal(p.Value, &eps); err == nil {
			playlist := anime.NewPlaylist(p.Name)
			playlist.SetEpisodes(eps)
			playlist.DbId = p.ID
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

func GetPlaylistsWithoutEpisodes(db *db.Database) ([]*anime.Playlist, error) {
	var res []*models.Playlist
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	playlists := make([]*anime.Playlist, 0)
	for _, p := range res {
		var eps []*anime.PlaylistEpisode
		if err := json.Unmarshal(p.Value, &eps); err == nil {
			playlist := anime.NewPlaylist(p.Name)
			playlist.DbId = p.ID
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

func SavePlaylist(db *db.Database, playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.Episodes)
	if err != nil {
		return err
	}
	entry := &models.Playlist{
		Name:  playlist.Name,
		Value: data,
	}

	return db.Gorm().Save(entry).Error
}

func DeletePlaylist(db *db.Database, id uint) error {
	return db.Gorm().Where("id = ?", id).Delete(&models.Playlist{}).Error
}

func UpdatePlaylist(db *db.Database, playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.Episodes)
	if err != nil {
		return err
	}

	// Get the playlist entry
	entry := &models.Playlist{}
	if err := db.Gorm().Where("id = ?", playlist.DbId).First(entry).Error; err != nil {
		return err
	}

	// Update the playlist entry
	entry.Name = playlist.Name
	entry.Value = data

	return db.Gorm().Save(entry).Error
}

func GetPlaylist(db *db.Database, id uint) (*anime.Playlist, error) {
	entry := &models.Playlist{}
	if err := db.Gorm().Where("id = ?", id).First(entry).Error; err != nil {
		return nil, err
	}

	var eps []*anime.PlaylistEpisode
	if err := json.Unmarshal(entry.Value, &eps); err != nil {
		return nil, err
	}

	playlist := anime.NewPlaylist(entry.Name)
	playlist.SetEpisodes(eps)
	playlist.DbId = entry.ID

	return playlist, nil
}
