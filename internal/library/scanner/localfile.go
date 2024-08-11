package scanner

import (
	"github.com/rs/zerolog"
	lop "github.com/samber/lo/parallel"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
)

// GetLocalFilesFromDir creates a new LocalFile for each video file
func GetLocalFilesFromDir(dirPath string, logger *zerolog.Logger) ([]*anime.LocalFile, error) {
	paths, err := filesystem.GetMediaFilePathsFromDirS(dirPath)

	logger.Trace().
		Any("dirPath", dirPath).
		Msg("localfile: Retrieving and creating local files")

	// Concurrently populate localFiles
	localFiles := lop.Map(paths, func(path string, index int) *anime.LocalFile {
		return anime.NewLocalFile(path, dirPath)
	})

	logger.Trace().
		Any("count", len(localFiles)).
		Msg("localfile: Retrieved local files")

	return localFiles, err
}
