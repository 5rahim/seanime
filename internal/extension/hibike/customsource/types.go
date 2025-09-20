package hibikecustomsource

import (
	"context"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
)

// Custom sources allow users to add custom media, often things not available on AniList, to the app.
// At runtime the extension will be assigned a unique extension identifier (int). Extension IDs use bit-based separation:
// AniList IDs: 0 to 2^31-1, Extension IDs: 2^31+ with embedded extension identifier and local ID.
// A custom source can be identified by its ID and SiteUrl property (e.g. "ext_custom_source_{extId}|END|https://example.com")
//
// Custom source media are fetched in the Platform and Metadata provider. The customsource.Manager is tasked with storing tracking info.

type (
	Settings struct {
		SupportsAnime bool `json:"supportsAnime"`
		SupportsManga bool `json:"supportsManga"`
	}

	ListAnimeResponse struct {
		Media      []*anilist.BaseAnime `json:"media"`
		Page       int                  `json:"page"`
		TotalPages int                  `json:"totalPages"`
		Total      int                  `json:"total"`
	}

	ListMangaResponse struct {
		Media      []*anilist.BaseManga `json:"media"`
		Page       int                  `json:"page"`
		TotalPages int                  `json:"totalPages"`
		Total      int                  `json:"total"`
	}

	Provider interface {
		GetExtensionIdentifier() int
		GetSettings() Settings
		GetAnime(ctx context.Context, id []int) ([]*anilist.BaseAnime, error)
		ListAnime(ctx context.Context, search string, page int, perPage int) (*ListAnimeResponse, error)
		GetAnimeWithRelations(ctx context.Context, id int) (*anilist.CompleteAnime, error)
		GetAnimeMetadata(ctx context.Context, id int) (*metadata.AnimeMetadata, error)
		GetAnimeDetails(ctx context.Context, id int) (*anilist.AnimeDetailsById_Media, error)
		GetManga(ctx context.Context, id []int) ([]*anilist.BaseManga, error)
		ListManga(ctx context.Context, search string, page int, perPage int) (*ListMangaResponse, error)
		GetMangaDetails(ctx context.Context, id int) (*anilist.MangaDetailsById_Media, error)
	}
)
