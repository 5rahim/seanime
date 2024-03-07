package db

import (
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/models"
)

func (db *Database) GetPlaylists() ([]*entities.Playlist, error) {
	var res []*models.PlaylistEntry
	err := db.gormdb.Find(&res).Error
	if err != nil {
		return nil, err
	}

	var playlists []*entities.Playlist
	for _, p := range res {
		var localFiles []*entities.LocalFile
		if err := json.Unmarshal(p.Value, &localFiles); err != nil {
			playlist := entities.NewPlaylist(p.Name)
			playlist.SetLocalFiles(localFiles)
			playlist.DbId = p.ID
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

func (db *Database) SavePlaylist(playlist *entities.Playlist) error {
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

func (db *Database) UpdatePlaylist(playlist *entities.Playlist) error {
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

func (db *Database) GetPlaylist(id uint) (*entities.Playlist, error) {
	playlistEntry := &models.PlaylistEntry{}
	if err := db.gormdb.Where("id = ?", id).First(playlistEntry).Error; err != nil {
		return nil, err
	}

	var localFiles []*entities.LocalFile
	if err := json.Unmarshal(playlistEntry.Value, &localFiles); err != nil {
		return nil, err
	}

	playlist := entities.NewPlaylist(playlistEntry.Name)
	playlist.SetLocalFiles(localFiles)
	playlist.DbId = playlistEntry.ID

	return playlist, nil
}
