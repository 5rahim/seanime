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

// func (h *Hub) UpdateAnimeListStatus

// func (h *Hub) UpdateMangaListStatus
