package metadata

import (
	"errors"
	"fmt"
	"github.com/seanime-app/seanime/internal/api/mappings"
	"github.com/seanime-app/seanime/internal/api/tvdb"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"time"
)

func getTvdbIDFromAnimeLists(anidbID int) (tvdbID int, ok bool) {
	res, err := mappings.GetReducedAnimeLists()
	if err != nil {
		return 0, false
	}
	return res.FindTvdbIDFromAnidbID(anidbID)
}

func (mw *MediaWrapper) EmptyTVDBEpisodesBucket(mediaId int) error {

	// Get TVDB ID
	var tvdbId int
	tvdbId = mw.anizipMedia.Mappings.ThetvdbID
	if tvdbId == 0 {
		if mw.anizipMedia.Mappings.AnidbID > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(mw.anizipMedia.Mappings.AnidbID)
		}
	}

	if tvdbId == 0 {
		return errors.New("metadata: could not find tvdb id")
	}

	return mw.fileCacher.Remove(fmt.Sprintf("tvdb_episodes_%d", mediaId))
}

func (mw *MediaWrapper) GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error) {
	key := mw.baseMedia.GetID()

	// Get TVDB ID
	var tvdbId int
	tvdbId = mw.anizipMedia.Mappings.ThetvdbID
	if tvdbId == 0 {
		if mw.anizipMedia.Mappings.AnidbID > 0 {
			// Try to get it from the mappings
			tvdbId, _ = getTvdbIDFromAnimeLists(mw.anizipMedia.Mappings.AnidbID)
		}
	}

	if tvdbId == 0 {
		return nil, errors.New("metadata: could not find tvdb id")
	}

	bucket := filecache.NewBucket(fmt.Sprintf("tvdb_episodes_%d", mw.baseMedia.GetID()), time.Hour*24*7*365)

	// Find episodes in cache
	var episodes []*tvdb.Episode
	found, _ := mw.fileCacher.Get(bucket, strconv.Itoa(key), &episodes)
	if !populate && found && episodes != nil {
		return episodes, nil
	}

	// Fetch episodes only if we need to populate
	if populate {
		var err error

		tv := tvdb.NewTVDB(&tvdb.NewTVDBOptions{
			Logger: mw.logger,
		})

		episodes, err = tv.FetchSeriesEpisodes(tvdbId, tvdb.FilterEpisodeMediaInfo{
			Year:           mw.baseMedia.GetStartDate().GetYear(),
			Month:          mw.baseMedia.GetStartDate().GetMonth(),
			TotalEp:        mw.anizipMedia.GetMainEpisodeCount(),
			AbsoluteOffset: mw.anizipMedia.GetOffset(),
		})
		if err != nil {
			return nil, err
		}

		err = mw.fileCacher.Set(bucket, strconv.Itoa(key), episodes)

		if err != nil {
			return nil, err
		}
	}

	return episodes, nil
}

func (mw *MediaWrapper) GetTVDBEpisodeByNumber(number int) (*tvdb.Episode, bool) {
	if mw == nil || mw.tvdbEpisodes == nil {
		return nil, false
	}

	for _, e := range mw.tvdbEpisodes {
		if e.Number == number {
			return e, true
		}
	}

	return nil, false
}
