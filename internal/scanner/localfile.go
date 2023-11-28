package scanner

import (
	"github.com/rs/zerolog"
	lop "github.com/samber/lo/parallel"
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/filesystem"
)

type AugmentedLocalFile struct {
	LocalFile  *entities.LocalFile
	Media      any
	AnimeTitle string
}

//----------------------------------------------------------------------------------------------------------------------

// LocalFile methods

//----------------------------------------------------------------------------------------------------------------------

// GetLocalFilesFromDir creates a new LocalFile for each video file
func GetLocalFilesFromDir(dirPath string, logger *zerolog.Logger) ([]*entities.LocalFile, error) {
	paths, err := filesystem.GetVideoFilePathsFromDir(dirPath)

	logger.Trace().
		Any("dirPath", dirPath).
		Msg("localfile: Retrieving and creating local files")

	// Concurrently populate localFiles
	localFiles := lop.Map(paths, func(path string, index int) *entities.LocalFile {
		return entities.NewLocalFile(path, dirPath)
	})

	logger.Trace().
		Any("count", len(localFiles)).
		Msg("localfile: Retrieved local files")

	return localFiles, err
}

//----------------------------------------------------------------------------------------------------------------------

//----------------------------------------------------------------------------------------------------------------------
