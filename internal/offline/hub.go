package offline

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/database/db"
	"github.com/seanime-app/seanime/internal/util/filecache"
)

type (
	// Hub is a struct that holds all the offline modules.
	Hub struct {
		anilistClientWrapper anilist.ClientWrapperInterface // Used to fetch anime and manga data from AniList
		db                   *db.Database
		fileCacher           *filecache.Cacher
		logger               *zerolog.Logger
	}
)

// NewHub creates a new offline hub.

// Snapshot populates offline data
func (h *Hub) Snapshot() error {

	// Use NewMediaEntry (anime)
	// Modify NewMediaEntry, so we don't concern ourselves with the DownloadInfo

	panic("not implemented")
}

// func (h *Hub) UpdateAnimeListStatus

// func (h *Hub) UpdateMangaListStatus
