package local

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/util/result"
	"strconv"

	"github.com/pkg/errors"
)

// OfflineMetadataProvider replaces the metadata provider only when offline
type OfflineMetadataProvider struct {
	manager            *ManagerImpl
	animeSnapshots     map[int]*AnimeSnapshot
	animeMetadataCache *result.BoundedCache[string, *metadata.AnimeMetadata]
}

type OfflineAnimeMetadataWrapper struct {
	anime    *anilist.BaseAnime
	metadata *metadata.AnimeMetadata
}

func NewOfflineMetadataProvider(manager *ManagerImpl) metadata.Provider {
	ret := &OfflineMetadataProvider{
		manager:            manager,
		animeSnapshots:     make(map[int]*AnimeSnapshot),
		animeMetadataCache: result.NewBoundedCache[string, *metadata.AnimeMetadata](500),
	}

	// Load the anime snapshots
	// DEVNOTE: We assume that it will be loaded once since it's used only when offline
	ret.loadAnimeSnapshots()

	return ret
}

func (mp *OfflineMetadataProvider) loadAnimeSnapshots() {
	animeSnapshots, ok := mp.manager.localDb.GetAnimeSnapshots()
	if !ok {
		return
	}

	for _, snapshot := range animeSnapshots {
		mp.animeSnapshots[snapshot.MediaId] = snapshot
	}
}

func (mp *OfflineMetadataProvider) GetAnimeMetadata(platform metadata.Platform, mId int) (*metadata.AnimeMetadata, error) {
	if platform != metadata.AnilistPlatform {
		return nil, errors.New("unsupported platform")
	}

	if snapshot, ok := mp.animeSnapshots[mId]; ok {
		localAnimeMetadata := snapshot.AnimeMetadata
		for _, episode := range localAnimeMetadata.Episodes {
			if imgUrl, ok := snapshot.EpisodeImagePaths[episode.Episode]; ok {
				episode.Image = *FormatAssetUrl(mId, imgUrl)
			}
		}

		return &metadata.AnimeMetadata{
			Titles:       localAnimeMetadata.Titles,
			Episodes:     localAnimeMetadata.Episodes,
			EpisodeCount: localAnimeMetadata.EpisodeCount,
			SpecialCount: localAnimeMetadata.SpecialCount,
			Mappings:     localAnimeMetadata.Mappings,
		}, nil
	}

	return nil, errors.New("anime metadata not found")
}

func (mp *OfflineMetadataProvider) GetCache() *result.BoundedCache[string, *metadata.AnimeMetadata] {
	return mp.animeMetadataCache
}

func (mp *OfflineMetadataProvider) GetAnimeMetadataWrapper(anime *anilist.BaseAnime, metadata *metadata.AnimeMetadata) metadata.AnimeMetadataWrapper {
	return &OfflineAnimeMetadataWrapper{
		anime:    anime,
		metadata: metadata,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (mw *OfflineAnimeMetadataWrapper) GetEpisodeMetadata(episodeNumber int) (ret metadata.EpisodeMetadata) {
	episodeMetadata, found := mw.metadata.FindEpisode(strconv.Itoa(episodeNumber))
	if found {
		ret = *episodeMetadata
	}
	return
}
