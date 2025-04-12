package scanner

import (
	"seanime/internal/api/anilist"
	"seanime/internal/events"
	"seanime/internal/library/anime"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

//----------------------------------------------------------------------------------------------------------------------

func TestScanner_Scan(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	anilistClient := anilist.TestGetMockAnilistClient()
	logger := util.NewLogger()
	anilistPlatform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	wsEventManager := events.NewMockWSEventManager(util.NewLogger())
	dir := "E:/Anime"

	tests := []struct {
		name  string
		paths []string
	}{
		{
			name: "Scan",
			paths: []string{
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 20v2 (1080p) [30072859].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 21v2 (1080p) [4B1616A5].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 22v2 (1080p) [58BF43B4].mkv",
				"E:/Anime/[SubsPlease] 86 - Eighty Six (01-23) (1080p) [Batch]/[SubsPlease] 86 - Eighty Six - 23v2 (1080p) [D94B4894].mkv",
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			existingLfs := make([]*anime.LocalFile, 0)
			for _, path := range tt.paths {
				lf := anime.NewLocalFile(path, dir)
				existingLfs = append(existingLfs, lf)
			}

			// +---------------------+
			// |        Scan         |
			// +---------------------+

			scanner := &Scanner{
				DirPath:            dir,
				Enhanced:           false,
				Platform:           anilistPlatform,
				Logger:             util.NewLogger(),
				WSEventManager:     wsEventManager,
				ExistingLocalFiles: existingLfs,
				SkipLockedFiles:    false,
				SkipIgnoredFiles:   false,
				ScanLogger:         nil,
				ScanSummaryLogger:  nil,
			}

			lfs, err := scanner.Scan()
			if err != nil {
				t.Fatal("expected result, got error:", err.Error())
			}

			for _, lf := range lfs {
				t.Log(lf.Name)
			}

		})

	}

}
