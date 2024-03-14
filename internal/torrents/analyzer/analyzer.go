package torrent_analyzer

import (
	"github.com/rs/zerolog"
	"github.com/seanime-app/seanime/internal/api/anilist"
	"github.com/seanime-app/seanime/internal/api/anizip"
	"github.com/seanime-app/seanime/internal/library/entities"
	"github.com/seanime-app/seanime/internal/library/scanner"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/seanime-app/seanime/internal/util/limiter"
	"path/filepath"
)

type (
	// Analyzer is a service similar to the scanner, but it is used to analyze torrent files.
	// i.e. torrent files instead of local files.
	Analyzer struct {
		files                []*File
		selectedFiles        []*File // Hydrated after analyzeHydratedFiles is called
		media                *anilist.BaseMedia
		anilistClientWrapper anilist.ClientWrapperInterface
		logger               *zerolog.Logger
	}

	// File represents a torrent file and contains its metadata.
	File struct {
		index     int
		path      string
		localFile *entities.LocalFile
	}
)

type (
	NewAnalyzerOptions struct {
		Logger               *zerolog.Logger
		Filepaths            []string           // Filepath of the torrent files
		Media                *anilist.BaseMedia // The media to compare the files with
		AnilistClientWrapper anilist.ClientWrapperInterface
	}
)

func NewAnalyzer(opts *NewAnalyzerOptions) *Analyzer {
	files := make([]*File, len(opts.Filepaths), len(opts.Filepaths))
	for i, filename := range opts.Filepaths {
		files[i] = newFile(i, filename)
	}
	return &Analyzer{
		files:                files,
		selectedFiles:        make([]*File, 0),
		media:                opts.Media,
		anilistClientWrapper: opts.AnilistClientWrapper,
		logger:               opts.Logger,
	}
}

func (a *Analyzer) Analyze() error {

	if err := a.scanFiles(); err != nil {
		return err
	}

	if err := a.analyzeHydratedFiles(); err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (a *Analyzer) analyzeHydratedFiles() error {

	keptFiles := make([]*File, 0)
	for _, af := range a.files {
		if af.localFile.MediaId == a.media.ID {
			keptFiles = append(keptFiles, af)
		}
	}
	a.selectedFiles = keptFiles

	return nil
}

// scanFiles scans the files and matches them with the media.
func (a *Analyzer) scanFiles() error {

	baseMediaCache := anilist.NewBaseMediaCache()
	anizipCache := anizip.NewCache()
	anilistRateLimiter := limiter.NewAnilistLimiter()

	lfs := a.getLocalFiles() // Extract local files from the Files

	// +---------------------+
	// |   MediaContainer    |
	// +---------------------+

	tree := anilist.NewBaseMediaRelationTree()
	if err := a.media.FetchMediaTree(anilist.FetchMediaTreeAll, a.anilistClientWrapper, anilistRateLimiter, tree, baseMediaCache); err != nil {
		return err
	}

	allMedia := tree.Values()

	mc := scanner.NewMediaContainer(&scanner.MediaContainerOptions{
		AllMedia: allMedia,
	})

	// +---------------------+
	// |      Matcher        |
	// +---------------------+

	matcher := &scanner.Matcher{
		LocalFiles:     lfs,
		MediaContainer: mc,
		BaseMediaCache: baseMediaCache,
		Logger:         util.NewLogger(),
	}

	err := matcher.MatchLocalFilesWithMedia()
	if err != nil {
		return err
	}

	// +---------------------+
	// |    FileHydrator     |
	// +---------------------+

	fh := &scanner.FileHydrator{
		LocalFiles:           lfs,
		AllMedia:             mc.NormalizedMedia,
		BaseMediaCache:       baseMediaCache,
		AnizipCache:          anizipCache,
		AnilistClientWrapper: a.anilistClientWrapper,
		AnilistRateLimiter:   anilistRateLimiter,
		Logger:               a.logger,
	}

	fh.HydrateMetadata()

	for _, af := range a.files {
		for _, lf := range lfs {
			if lf.Path == af.localFile.Path {
				af.localFile = lf // Update the local file in the File
				break
			}
		}
	}

	return nil
}

// newFile creates a new File from a file path.
func newFile(idx int, path string) *File {
	path = filepath.ToSlash(path)

	return &File{
		index:     idx,
		path:      path,
		localFile: entities.NewLocalFile(path, ""),
	}
}

func (a *Analyzer) getLocalFiles() []*entities.LocalFile {
	files := make([]*entities.LocalFile, len(a.files))
	for i, f := range a.files {
		files[i] = f.localFile
	}
	return files
}
