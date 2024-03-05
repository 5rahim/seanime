package torrent_client

import (
	"errors"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/comparison"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/nyaa"
	"github.com/seanime-app/seanime/internal/qbittorrent/model"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/sourcegraph/conc/pool"
	"math"
	"slices"
	"strconv"
	"time"
)

// SmartSelect will select only episodes that are missing.
// It will return an error if SmartSelect.Magnets has more than one magnet link.
// The torrent will be deleted when an error occurs.
// SmartSelect will block until it is done.
//
// TODO: Add support for transmission
func (r *Repository) SmartSelect(opts *SmartSelect) error {
	if !opts.Enabled {
		return nil
	}

	if len(opts.Magnets) != 1 {
		return errors.New("incorrect number of magnets")
	}

	if opts.Enabled && opts.Media == nil {
		return errors.New("no media found")
	}

	if r.Provider == TransmissionProvider {
		return errors.New("automatic file selection not supported for transmission")
	}

	magnet := opts.Magnets[0]
	// get hash
	hash, ok := nyaa.ExtractHashFromMagnet(magnet)
	if !ok {
		return errors.New("could not extract hash")
	}

	// ticker
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Set a timeout of 1 minute
	timeout := time.After(time.Minute)

	// exit
	done := make(chan struct{})

	var err error

	var contents []*qbittorrent_model.TorrentContent

	contentsChan := make(chan []*qbittorrent_model.TorrentContent)

	// get torrent contents when it's done loading
	go func() {
		for {
			select {
			case <-ticker.C:
				ret, _ := r.QbittorrentClient.Torrent.GetContents(hash)
				if ret != nil && len(ret) > 0 {
					contentsChan <- ret
					return
				}
			case <-timeout:
				return
			}
		}
	}()

workDone:
	for {
		select {
		case <-done:
			break workDone
		case <-timeout:
			err = errors.New("timeout occurred: unable to retrieve torrent content")
			_ = r.RemoveTorrents([]string{hash})
			close(done)
		case contents = <-contentsChan:
			close(done)
		}
	}

	if err != nil {
		_ = r.RemoveTorrents([]string{hash})
		return err
	}

	// pause the torrent
	err = r.PauseTorrents([]string{hash})
	if err != nil {
		_ = r.RemoveTorrents([]string{hash})
		return err
	}

	tmpLfs := r.getBestTempLocalFiles(contents, opts)

	// filter out files that are not main and without episode numbers
	tmpLfs = lo.Filter(tmpLfs, func(tmpLf *TmpLocalFile, _ int) bool {
		// remove files that are not main
		if comparison.ValueContainsSpecial(tmpLf.localFile.Name) || comparison.ValueContainsNC(tmpLf.localFile.Name) {
			return false
		}
		// remove files that don't have an episode number
		if tmpLf.localFile.ParsedData.Episode == "" {
			return false
		}
		return true
	})

	hasAtLeastOneAbsoluteEpisode := lo.SomeBy(tmpLfs, func(tmpLf *TmpLocalFile) bool {
		episode, _ := util.StringToInt(tmpLf.localFile.ParsedData.Episode)
		return episode > opts.Media.GetCurrentEpisodeCount()
	})

	// detect absolute episode number
	tmpLfs = lop.Map(tmpLfs, func(tmpLf *TmpLocalFile, _ int) *TmpLocalFile {
		episode, ok := util.StringToInt(tmpLf.localFile.ParsedData.Episode)
		if !ok {
			return tmpLf
		}
		// remove absolute offset from episode number ONLY if one absolute episode number is found
		if hasAtLeastOneAbsoluteEpisode {
			episode = episode - opts.AbsoluteOffset
		}
		tmpLf.localFile.Metadata.Episode = episode
		return tmpLf
	})

	// find episode number duplicates
	// if there are duplicate episode numbers (more than 3 duplicates), return error
	// we choose 3 as the threshold because sometimes there might be 1or2 episodes with different versions
	// this is used to prevent incorrect selections
	duplicates := lo.FindDuplicatesBy(tmpLfs, func(item *TmpLocalFile) int {
		return item.localFile.Metadata.Episode
	})
	if len(duplicates) > 2 {
		return errors.New("automatic torrent file selection not supported")
	}

	// remove files whose episode number is not in the missing episode numbers list
	toRemove := lo.Filter(tmpLfs, func(tmpLf *TmpLocalFile, _ int) bool {
		return !slices.Contains(opts.MissingEpisodeNumbers, tmpLf.localFile.Metadata.Episode)
	})

	// get the indices of the files that we will deselect
	toRemoveIndices := lop.Map(toRemove, func(tmpLf *TmpLocalFile, _ int) string {
		return strconv.Itoa(tmpLf.index)
	})

	// set priority to 0 for files that are not in the missing episode numbers list
	err = r.QbittorrentClient.Torrent.SetFilePriorities(hash, toRemoveIndices, 0)
	if err != nil {
		_ = r.RemoveTorrents([]string{hash})
		return err
	}

	// pause the torrent
	err = r.ResumeTorrents([]string{hash})
	if err != nil {
		_ = r.RemoveTorrents([]string{hash})
		return err
	}

	return nil
}

// getBestTempLocalFiles returns the best local files that match the media
func (r *Repository) getBestTempLocalFiles(contents []*qbittorrent_model.TorrentContent, opts *SmartSelect) []*TmpLocalFile {

	// get local files from contents
	tmpLfs := lop.Map(contents, func(content *qbittorrent_model.TorrentContent, idx int) *TmpLocalFile {
		return &TmpLocalFile{
			torrentContent: content,
			localFile:      entities.NewLocalFile(content.Name, opts.Destination),
			index:          idx,
		}
	})

	type comparisonRes struct {
		tmpLocalFile *TmpLocalFile
		rating       float64
	}

	titles := opts.Media.GetAllTitles()

	// compare each local file title variations with the media titles and synonyms
	p := pool.NewWithResults[*comparisonRes]()
	for _, tmpLf := range tmpLfs {
		p.Go(func() *comparisonRes {
			comparisons := lop.Map(titles, func(title *string, _ int) *comparison.SorensenDiceResult {

				titleVariations := tmpLf.localFile.GetTitleVariations()

				comps := make([]*comparison.SorensenDiceResult, 0)
				if eng, found := comparison.FindBestMatchWithSorensenDice(title, titleVariations); found {
					comps = append(comps, eng)
				}
				if rom, found := comparison.FindBestMatchWithSorensenDice(title, titleVariations); found {
					comps = append(comps, rom)
				}
				if syn, found := comparison.FindBestMatchWithSorensenDice(title, titleVariations); found {
					comps = append(comps, syn)
				}
				var res *comparison.SorensenDiceResult
				if len(comps) > 1 {
					res = lo.Reduce(comps, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
						if prev.Rating > curr.Rating {
							return prev
						} else {
							return curr
						}
					}, comps[0])
				} else if len(comps) == 1 {
					return comps[0]
				}
				return res
			})

			// Retrieve the best result from all the title variations results
			bestRes := lo.Reduce(comparisons, func(prev *comparison.SorensenDiceResult, curr *comparison.SorensenDiceResult, _ int) *comparison.SorensenDiceResult {
				if prev.Rating > curr.Rating {
					return prev
				} else {
					return curr
				}
			}, comparisons[0])

			return &comparisonRes{
				tmpLocalFile: tmpLf,
				rating:       bestRes.Rating,
			}
		})
	}
	compLfs := p.Wait()

	highestRating := lo.Reduce(compLfs, func(prev float64, curr *comparisonRes, _ int) float64 {
		if prev > curr.rating {
			return prev
		} else {
			return curr.rating
		}
	}, 0.0)

	usedComps := lo.Filter(compLfs, func(item *comparisonRes, index int) bool {
		return item.rating == highestRating || math.Abs(item.rating-highestRating) < 0.2
	})

	usedTmpLfs := lop.Map(usedComps, func(item *comparisonRes, index int) *TmpLocalFile {
		return item.tmpLocalFile
	})

	return usedTmpLfs

}
