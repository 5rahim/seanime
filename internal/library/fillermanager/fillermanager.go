package fillermanager

import (
	"seanime/internal/api/filler"
	"seanime/internal/database/db"
	"seanime/internal/library/anime"
	"seanime/internal/onlinestream"
	"seanime/internal/util"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	lop "github.com/samber/lo/parallel"
)

type (
	Interface interface {
		// RefetchFillerData re-fetches the fillers for the given media IDs
		RefetchFillerData() error
		// HasFillerFetched checks if the fillers for the given media ID have been fetched
		HasFillerFetched(mediaId int) bool
		// FetchAndStoreFillerData fetches the filler data for the given media ID
		FetchAndStoreFillerData(mediaId int, titles []string) error
		// RemoveFillerData removes the filler data for the given media ID
		RemoveFillerData(mediaId int) error
		// IsEpisodeFiller checks if the given episode number is a filler for the given media ID
		IsEpisodeFiller(mediaId int, episodeNumber int) bool
	}

	FillerManager struct {
		db        *db.Database
		logger    *zerolog.Logger
		fillerApi filler.API
	}

	NewFillerManagerOptions struct {
		DB     *db.Database
		Logger *zerolog.Logger
	}
)

func New(opts *NewFillerManagerOptions) *FillerManager {
	return &FillerManager{
		db:        opts.DB,
		logger:    opts.Logger,
		fillerApi: filler.NewAnimeFillerList(opts.Logger),
	}
}

func (fm *FillerManager) RefetchFillerData() error {

	defer util.HandlePanicInModuleThen("library/fillermanager/RefetchFillerData", func() {
		fm.logger.Error().Msg("fillermanager: Failed to re-fetch filler data")
	})

	wg := sync.WaitGroup{}

	fm.logger.Debug().Msg("fillermanager: Re-fetching filler data")

	mediaFillers, err := fm.db.GetCachedMediaFillers()
	if err != nil {
		return err
	}

	for _, mf := range mediaFillers {
		wg.Add(1)
		go func(*db.MediaFillerItem) {
			defer wg.Done()
			// Fetch the db data

			// Fetch the filler data
			fillerData, err := fm.fillerApi.FindFillerData(mf.Slug)
			if err != nil {
				fm.logger.Error().Err(err).Int("mediaId", mf.MediaId).Msg("fillermanager: Failed to fetch filler data")
				return
			}

			// Update the filler data
			mf.FillerEpisodes = fillerData.FillerEpisodes

		}(mf)
	}
	wg.Wait()

	err = fm.db.SaveCachedMediaFillerItems()
	if err != nil {
		return err
	}

	fm.logger.Debug().Msg("fillermanager: Re-fetched filler data")

	return nil
}

func (fm *FillerManager) HasFillerFetched(mediaId int) bool {

	defer util.HandlePanicInModuleThen("library/fillermanager/HasFillerFetched", func() {
	})

	_, ok := fm.db.GetMediaFillerItem(mediaId)
	return ok
}

func (fm *FillerManager) GetFillerEpisodes(mediaId int) ([]string, bool) {

	defer util.HandlePanicInModuleThen("library/fillermanager/GetFillerEpisodes", func() {
	})

	fillerItem, ok := fm.db.GetMediaFillerItem(mediaId)
	if !ok {
		return nil, false
	}

	return fillerItem.FillerEpisodes, true
}

func (fm *FillerManager) FetchAndStoreFillerData(mediaId int, titles []string) error {

	defer util.HandlePanicInModuleThen("library/fillermanager/FetchAndStoreFillerData", func() {
	})

	fm.logger.Debug().Int("mediaId", mediaId).Msg("fillermanager: Fetching filler data")

	res, err := fm.fillerApi.Search(filler.SearchOptions{
		Titles: titles,
	})
	if err != nil {
		return err
	}

	fm.logger.Debug().Int("mediaId", mediaId).Str("slug", res.Slug).Msg("fillermanager: Fetched filler data")

	return fm.fetchAndStoreFillerDataFromSlug(mediaId, res.Slug)
}

func (fm *FillerManager) fetchAndStoreFillerDataFromSlug(mediaId int, slug string) error {

	defer util.HandlePanicInModuleThen("library/fillermanager/FetchAndStoreFillerDataFromSlug", func() {
	})

	fillerData, err := fm.fillerApi.FindFillerData(slug)
	if err != nil {
		return err
	}

	err = fm.db.InsertMediaFiller(
		"animefillerlist",
		mediaId,
		slug,
		time.Now(),
		fillerData.FillerEpisodes,
	)
	if err != nil {
		return err
	}

	return nil
}

func (fm *FillerManager) StoreFillerData(source string, slug string, mediaId int, fillerEpisodes []string) error {

	defer util.HandlePanicInModuleThen("library/fillermanager/StoreFillerDataForMedia", func() {
	})

	return fm.db.InsertMediaFiller(
		source,
		mediaId,
		slug,
		time.Now(),
		fillerEpisodes,
	)
}

func (fm *FillerManager) RemoveFillerData(mediaId int) error {

	defer util.HandlePanicInModuleThen("library/fillermanager/RemoveFillerData", func() {
	})

	fm.logger.Debug().Int("mediaId", mediaId).Msg("fillermanager: Removing filler data")

	return fm.db.DeleteMediaFiller(mediaId)
}

func (fm *FillerManager) IsEpisodeFiller(mediaId int, episodeNumber int) bool {

	defer util.HandlePanicInModuleThen("library/fillermanager/IsEpisodeFiller", func() {
	})

	mediaFillerData, ok := fm.db.GetMediaFillerItem(mediaId)
	if !ok {
		return false
	}

	if len(mediaFillerData.FillerEpisodes) == 0 {
		return false
	}

	for _, ep := range mediaFillerData.FillerEpisodes {
		if ep == strconv.Itoa(episodeNumber) {
			return true
		}
	}

	return false
}

func (fm *FillerManager) HydrateFillerData(e *anime.Entry) {
	if fm == nil {
		return
	}
	if e == nil || e.Media == nil || e.Episodes == nil || len(e.Episodes) == 0 {
		return
	}

	// Check if the filler data has been fetched
	if !fm.HasFillerFetched(e.Media.ID) {
		return
	}

	lop.ForEach(e.Episodes, func(ep *anime.Episode, _ int) {
		if ep == nil || ep.EpisodeMetadata == nil {
			return
		}
		ep.EpisodeMetadata.IsFiller = fm.IsEpisodeFiller(e.Media.ID, ep.EpisodeNumber)
	})
}

func (fm *FillerManager) HydrateOnlinestreamFillerData(mId int, episodes []*onlinestream.Episode) {
	if fm == nil {
		return
	}
	if episodes == nil || len(episodes) == 0 {
		return
	}

	// Check if the filler data has been fetched
	if !fm.HasFillerFetched(mId) {
		return
	}

	for _, ep := range episodes {
		ep.IsFiller = fm.IsEpisodeFiller(mId, ep.Number)
	}
}

func (fm *FillerManager) HydrateEpisodeFillerData(mId int, e *anime.Episode) {
	if fm == nil || e == nil {
		return
	}

	// Check if the filler data has been fetched
	if !fm.HasFillerFetched(mId) {
		return
	}

	e.EpisodeMetadata.IsFiller = fm.IsEpisodeFiller(mId, e.EpisodeNumber)
}
