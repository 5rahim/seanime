package anime

import (
	"strconv"
	"strings"
)

type MockHydratedLocalFileOptions struct {
	FilePath             string
	LibraryPath          string
	MediaId              int
	MetadataEpisode      int
	MetadataAniDbEpisode string
	MetadataType         LocalFileType
}

func MockHydratedLocalFile(opts MockHydratedLocalFileOptions) *LocalFile {
	lf := NewLocalFile(opts.FilePath, opts.LibraryPath)
	lf.MediaId = opts.MediaId
	lf.Metadata = &LocalFileMetadata{
		AniDBEpisode: opts.MetadataAniDbEpisode,
		Episode:      opts.MetadataEpisode,
		Type:         opts.MetadataType,
	}
	return lf
}

// MockHydratedLocalFiles creates a slice of LocalFiles based on the provided options
//
// Example:
//
//	MockHydratedLocalFiles(
//		MockHydratedLocalFileOptions{
//			FilePath:             "/mnt/anime/One Piece/One Piece - 1070.mkv",
//			LibraryPath:          "/mnt/anime/",
//			MetadataEpisode:      1070,
//			MetadataAniDbEpisode: "1070",
//			MetadataType:         LocalFileTypeMain,
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
	MetadataEpisode      int
	MetadataAniDbEpisode string
	MetadataType         LocalFileType
}

// MockGenerateHydratedLocalFileGroupOptions generates a slice of MockHydratedLocalFileOptions based on a template string and metadata
//
// Example:
//
//	MockGenerateHydratedLocalFileGroupOptions("/mnt/anime/", "One Piece/One Piece - %ep.mkv", 21, []MockHydratedLocalFileWrapperOptionsMetadata{
//		{MetadataEpisode: 1070, MetadataAniDbEpisode: "1070", MetadataType: LocalFileTypeMain},
//	})
func MockGenerateHydratedLocalFileGroupOptions(libraryPath string, template string, mId int, m []MockHydratedLocalFileWrapperOptionsMetadata) []MockHydratedLocalFileOptions {
	opts := make([]MockHydratedLocalFileOptions, 0, len(m))
	for _, metadata := range m {
		opts = append(opts, MockHydratedLocalFileOptions{
			FilePath:             strings.ReplaceAll(template, "%ep", strconv.Itoa(metadata.MetadataEpisode)),
			LibraryPath:          libraryPath,
			MediaId:              mId,
			MetadataEpisode:      metadata.MetadataEpisode,
			MetadataAniDbEpisode: metadata.MetadataAniDbEpisode,
			MetadataType:         metadata.MetadataType,
		})
	}
	return opts
}
