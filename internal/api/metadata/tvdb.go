package metadata

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/mappings"
	"github.com/seanime-app/seanime/internal/api/tvdb"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"time"
)

var (
	fcTVDBEpisodesBucket = filecache.NewBucket("tvdb_episodes", time.Hour*24*7*365) // Store TVDB episodes permanently
)

func getTvdbIDFromAnimeLists(anidbID int) (tvdbID int, ok bool) {
	res, err := mappings.GetReducedAnimeLists()
	if err != nil {
		return 0, false
	}
	return res.FindTvdbIDFromAnidbID(anidbID)
}

func (mw *MediaWrapper) EmptyTVDBEpisodesBucket() error {
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
		return errors.New("metadata: could not find tvdb id")
	}

	return mw.fileCacher.Delete(fcTVDBEpisodesBucket, strconv.Itoa(key))
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

	// Find episodes in cache
	var episodes []*tvdb.Episode
	found, _ := mw.fileCacher.Get(fcTVDBEpisodesBucket, strconv.Itoa(key), &episodes)
	if found && episodes != nil {
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

		err = mw.fileCacher.Set(fcTVDBEpisodesBucket, strconv.Itoa(key), episodes)

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
