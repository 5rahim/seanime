package sync

import (
	"github.com/pkg/errors"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/api/tvdb"
	sync_util "seanime/internal/sync/util"
	"seanime/internal/util/result"
	"strconv"
)

// LocalMetadataProvider replaces the metadata provider only when offline
type LocalMetadataProvider struct {
	manager            *ManagerImpl
	animeSnapshots     map[int]*AnimeSnapshot
	animeMetadataCache *result.Cache[string, *metadata.AnimeMetadata]
}

type LocalAnimeMetadataWrapper struct {
	anime    *anilist.BaseAnime
	metadata *metadata.AnimeMetadata
}

func NewLocalMetadataProvider(manager *ManagerImpl) metadata.Provider {
	ret := &LocalMetadataProvider{
		manager:            manager,
		animeSnapshots:     make(map[int]*AnimeSnapshot),
		animeMetadataCache: result.NewCache[string, *metadata.AnimeMetadata](),
	}

	// Load the anime snapshots
	// DEVNOTE: We assume that it will be loaded once since it's used only when offline
	ret.loadAnimeSnapshots()

	return ret
}

func (mp *LocalMetadataProvider) loadAnimeSnapshots() {
	animeSnapshots, ok := mp.manager.localDb.GetAnimeSnapshots()
	if !ok {
		return
	}

	for _, snapshot := range animeSnapshots {
		mp.animeSnapshots[snapshot.MediaId] = snapshot
	}
}

func (mp *LocalMetadataProvider) GetAnimeMetadata(platform metadata.Platform, mId int) (*metadata.AnimeMetadata, error) {
	if platform != metadata.AnilistPlatform {
		return nil, errors.New("unsupported platform")
	}

	if snapshot, ok := mp.animeSnapshots[mId]; ok {
		localAnimeMetadata := snapshot.AnimeMetadata
		for _, episode := range localAnimeMetadata.Episodes {
			if imgUrl, ok := snapshot.EpisodeImagePaths[episode.Episode]; ok {
				episode.Image = *sync_util.FormatAssetUrl(mId, imgUrl)
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

func (mp *LocalMetadataProvider) GetCache() *result.Cache[string, *metadata.AnimeMetadata] {
	return mp.animeMetadataCache
}

func (mp *LocalMetadataProvider) GetAnimeMetadataWrapper(anime *anilist.BaseAnime, metadata *metadata.AnimeMetadata) metadata.AnimeMetadataWrapper {
	return &LocalAnimeMetadataWrapper{
		anime:    anime,
		metadata: metadata,
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (mw *LocalAnimeMetadataWrapper) GetEpisodeMetadata(episodeNumber int) (ret metadata.EpisodeMetadata) {
	episodeMetadata, found := mw.metadata.FindEpisode(strconv.Itoa(episodeNumber))
	if found {
		ret = *episodeMetadata
	}
	return
}

func (mw *LocalAnimeMetadataWrapper) EmptyTVDBEpisodesBucket(mediaId int) error {
	return nil
}

func (mw *LocalAnimeMetadataWrapper) GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error) {
	return make([]*tvdb.Episode, 0), nil
}

func (mw *LocalAnimeMetadataWrapper) GetTVDBEpisodeByNumber(episodeNumber int) (*tvdb.Episode, bool) {
	return nil, false
}
