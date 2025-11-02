package db_bridge

import (
	"seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/library/anime"

	"github.com/goccy/go-json"
)

// GetLegacyPlaylists
// DEPRECATED
func GetLegacyPlaylists(db *db.Database) ([]*anime.LegacyPlaylist, error) {
	var res []*models.PlaylistEntry
	err := db.Gorm().Find(&res).Error
	if err != nil {
		return nil, err
	}

	playlists := make([]*anime.LegacyPlaylist, 0)
	for _, p := range res {
		var localFiles []*anime.LocalFile
		if err := json.Unmarshal(p.Value, &localFiles); err == nil {
			playlist := anime.NewLegacyPlaylist(p.Name)
			playlist.SetLocalFiles(localFiles)
			playlist.DbId = p.ID
			playlists = append(playlists, playlist)
		}
	}
	return playlists, nil
}

// SaveLegacyPlaylist
// DEPRECATED
func SaveLegacyPlaylist(db *db.Database, playlist *anime.LegacyPlaylist) error {
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

// DeleteLegacyPlaylist
// DEPRECATED
func DeleteLegacyPlaylist(db *db.Database, id uint) error {
	return db.Gorm().Where("id = ?", id).Delete(&models.PlaylistEntry{}).Error
}

// UpdateLegacyPlaylist
// DEPRECATED
func UpdateLegacyPlaylist(db *db.Database, playlist *anime.LegacyPlaylist) error {
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

// GetLegacyPlaylist
// DEPRECATED
func GetLegacyPlaylist(db *db.Database, id uint) (*anime.LegacyPlaylist, error) {
	playlistEntry := &models.PlaylistEntry{}
	if err := db.Gorm().Where("id = ?", id).First(playlistEntry).Error; err != nil {
		return nil, err
	}

	var localFiles []*anime.LocalFile
	if err := json.Unmarshal(playlistEntry.Value, &localFiles); err != nil {
		return nil, err
	}

	playlist := anime.NewLegacyPlaylist(playlistEntry.Name)
	playlist.SetLocalFiles(localFiles)
	playlist.DbId = playlistEntry.ID

	return playlist, nil
}
