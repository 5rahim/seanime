package scanner

import (
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"github.com/seanime-app/seanime-server/internal/comparison"
	"github.com/seanime-app/seanime-server/internal/limiter"
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/sourcegraph/conc/pool"
	"strconv"
	"time"
)

type FileHydrator struct {
	localFiles     []*LocalFile
	media          []*anilist.BaseMedia
	baseMediaCache *anilist.BaseMediaCache
	anizipCache    *anizip.Cache
}

// HydrateMetadata will hydrate the metadata of each LocalFile with the metadata of the matched anilist.BaseMedia.
func (fh *FileHydrator) HydrateMetadata() {
	rateLimiter := limiter.NewLimiter(5*time.Second, 20)

	// Group local files by media ID
	groups := lop.GroupBy(fh.localFiles, func(localFile *LocalFile) int {
		return localFile.MediaId
	})

	// Remove the group with unmatched media
	delete(groups, 0)

	// Process each group in parallel
	p := pool.New()
	for mId, files := range groups {
		p.Go(func() {
			if len(files) > 0 {
				fh.hydrateGroupMetadata(mId, files, rateLimiter)
			}
		})
	}
	p.Wait()
}

func (fh *FileHydrator) hydrateGroupMetadata(
	mId int,
	lfs []*LocalFile,
	rateLimiter *limiter.Limiter,
) {

	// Get the media
	media, found := lo.Find(fh.media, func(media *anilist.BaseMedia) bool {
		return media.ID == mId
	})
	if !found {
		return
	}

	// Process each local file in the group sequentially
	lo.ForEach(lfs, func(lf *LocalFile, index int) {

		// Get episode number
		episode := -1
		if len(lf.ParsedData.Episode) > 0 {
			if ep, ok := util.StringToInt(lf.ParsedData.Episode); ok {
				episode = ep
			}
		}

		if comparison.ValueContainsSpecial(lf.Name) {
			lf.Metadata.IsSpecial = true
			if episode > -1 {
				lf.Metadata.Episode = episode
				lf.Metadata.AniDBEpisode = "S" + strconv.Itoa(episode)
			}
			return
		}
		if comparison.ValueContainsNC(lf.Name) {
			lf.Metadata.IsNC = true
			return
		}
		// Movie metadata
		if *media.Format == anilist.MediaFormatMovie {
			lf.Metadata.Episode = 1
			lf.Metadata.AniDBEpisode = "1"
			return
		}

		// No absolute episode count
		if episode <= media.GetTotalEpisodeCount() {
			// Episode 0 - Might be a special
			// By default, we will assume that AniDB doesn't include it as main episodes (which is often the case)
			if episode == 0 {
				// Leave episode number as 0, assuming that the client will handle tracking correctly
				lf.Metadata.Episode = 0
				lf.Metadata.AniDBEpisode = "S1"
			}
			return
		}

		// Episode number is higher but media only has 1 episode
		// - Might be a movie that was not correctly identified as such
		// - Or, the torrent files were divided into multiple episodes from a media that is listed as a movie on AniList
		if episode > media.GetTotalEpisodeCount() && media.GetTotalEpisodeCount() == 1 {
			lf.Metadata.Episode = 1 // Coerce episode number to 1 because it is used for tracking
			lf.Metadata.AniDBEpisode = "1"
			return
		}

		// Absolute episode count
		if episode > media.GetTotalEpisodeCount() {
			// TODO: Fetch media tree, normalize episode number
			return
		}

	})

}
