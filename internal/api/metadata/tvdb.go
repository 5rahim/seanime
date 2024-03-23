package metadata

import (
	"errors"
	"github.com/seanime-app/seanime/internal/api/tvdb"
	"github.com/seanime-app/seanime/internal/util/filecache"
	"strconv"
	"time"
)

var (
	fcTVDBEpisodesBucket = filecache.NewBucket("tvdb_episodes", time.Hour*24*7*365) // Store TVDB episodes permanently
)

func (mw *MediaWrapper) EmptyTVDBEpisodesBucket() error {

	// Get TVDB ID
	tvdbId := mw.anizipMedia.Mappings.ThetvdbID
	if tvdbId == 0 {
		return errors.New("metadata: could not find tvdb id")
	}

	return mw.fileCacher.Delete(fcTVDBEpisodesBucket, strconv.Itoa(tvdbId))
}

func (mw *MediaWrapper) GetTVDBEpisodes(populate bool) ([]*tvdb.Episode, error) {

	// Get TVDB ID
	tvdbId := mw.anizipMedia.Mappings.ThetvdbID
	if tvdbId == 0 {
		return nil, errors.New("metadata: could not find tvdb id")
	}

	// Find episodes in cache
	var episodes []*tvdb.Episode
	found, _ := mw.fileCacher.Get(fcTVDBEpisodesBucket, strconv.Itoa(tvdbId), &episodes)
	if found && episodes != nil {
		return episodes, nil
	}

	// Fetch episodes only if we need to populate
	if populate {
		var err error

		tv := tvdb.NewTVDB(&tvdb.NewTVDBOptions{
			Logger: mw.logger,
		})

		episodes, err = tv.FetchSeriesEpisodes(tvdbId)
		if err != nil {
			return nil, err
		}

		key := strconv.Itoa(tvdbId)
		err = mw.fileCacher.Set(fcTVDBEpisodesBucket, key, episodes)

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
