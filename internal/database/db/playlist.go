package db

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/database/models"
	"github.com/seanime-app/seanime/internal/library/anime"
)

func (db *Database) GetPlaylists() ([]*anime.Playlist, error) {
	var res []*models.PlaylistEntry
	err := db.gormdb.Find(&res).Error
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

func (db *Database) SavePlaylist(playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.LocalFiles)
	if err != nil {
		return err
	}
	playlistEntry := &models.PlaylistEntry{
		Name:  playlist.Name,
		Value: data,
	}

	return db.gormdb.Save(playlistEntry).Error
}

func (db *Database) DeletePlaylist(id uint) error {
	return db.gormdb.Where("id = ?", id).Delete(&models.PlaylistEntry{}).Error
}

func (db *Database) UpdatePlaylist(playlist *anime.Playlist) error {
	data, err := json.Marshal(playlist.LocalFiles)
	if err != nil {
		return err
	}

	// Get the playlist entry
	playlistEntry := &models.PlaylistEntry{}
	if err := db.gormdb.Where("id = ?", playlist.DbId).First(playlistEntry).Error; err != nil {
		return err
	}

	// Update the playlist entry
	playlistEntry.Name = playlist.Name
	playlistEntry.Value = data

	return db.gormdb.Save(playlistEntry).Error
}

func (db *Database) GetPlaylist(id uint) (*anime.Playlist, error) {
	playlistEntry := &models.PlaylistEntry{}
	if err := db.gormdb.Where("id = ?", id).First(playlistEntry).Error; err != nil {
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
