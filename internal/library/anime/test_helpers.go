package anime

import (
	"strconv"
	"strings"
)

type MockHydratedLocalFileOptions struct {
	filePath             string
	libraryPath          string
	mediaId              int
	metadataEpisode      int
	metadataAniDbEpisode string
	metadataType         LocalFileType
}

func MockHydratedLocalFile(opts MockHydratedLocalFileOptions) *LocalFile {
	lf := NewLocalFile(opts.filePath, opts.libraryPath)
	lf.MediaId = opts.mediaId
	lf.Metadata = &LocalFileMetadata{
		AniDBEpisode: opts.metadataAniDbEpisode,
		Episode:      opts.metadataEpisode,
		Type:         opts.metadataType,
	}
	return lf
}

// MockHydratedLocalFiles creates a slice of LocalFiles based on the provided options
//
// Example:
//
//	MockHydratedLocalFiles(
//		MockHydratedLocalFileOptions{
//			filePath:             "/mnt/anime/One Piece/One Piece - 1070.mkv",
//			libraryPath:          "/mnt/anime/",
//			metadataEpisode:      1070,
//			metadataAniDbEpisode: "1070",
//			metadataType:         LocalFileTypeMain,
//		},
//		MockHydratedLocalFileOptions{
//			...
//		},
//	)
func MockHydratedLocalFiles(opts ...[]MockHydratedLocalFileOptions) []*LocalFile {
	lfs := make([]*LocalFile, 0, len(opts))
	for _, opt := range opts {
		for _, o := range opt {
			lfs = append(lfs, MockHydratedLocalFile(o))
		}
	}
	return lfs
}

type MockHydratedLocalFileWrapperOptionsMetadata struct {
	metadataEpisode      int
	metadataAniDbEpisode string
	metadataType         LocalFileType
}

// MockGenerateHydratedLocalFileGroupOptions generates a slice of MockHydratedLocalFileOptions based on a template string and metadata
//
// Example:
//
//	MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "One Piece/One Piece - %ep.mkv", 21, []MockHydratedLocalFileWrapperOptionsMetadata{
//		{metadataEpisode: 1070, metadataAniDbEpisode: "1070", metadataType: LocalFileTypeMain},
//	})
func MockGenerateHydratedLocalFileGroupOptions(libraryPath string, template string, mId int, m []MockHydratedLocalFileWrapperOptionsMetadata) []MockHydratedLocalFileOptions {
	opts := make([]MockHydratedLocalFileOptions, 0, len(m))
	for _, metadata := range m {
		opts = append(opts, MockHydratedLocalFileOptions{
			filePath:             strings.ReplaceAll(template, "%ep", strconv.Itoa(metadata.metadataEpisode)),
			libraryPath:          libraryPath,
			mediaId:              mId,
			metadataEpisode:      metadata.metadataEpisode,
			metadataAniDbEpisode: metadata.metadataAniDbEpisode,
			metadataType:         metadata.metadataType,
		})
	}
	return opts
}
