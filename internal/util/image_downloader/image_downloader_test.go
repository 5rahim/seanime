package image_downloader

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestImageDownloader_DownloadImages(t *testing.T) {

	tests := []struct {
		name        string
		urls        []string
		downloadDir string
		expectedNum int
		cancelAfter int
	}{
		{
			name: "test1",
			urls: []string{
				"https://s4.anilist.co/file/anilistcdn/media/anime/banner/153518-7uRvV7SLqmHV.jpg",
				"https://s4.anilist.co/file/anilistcdn/media/anime/banner/153518-7uRvV7SLqmHV.jpg",
				"https://s4.anilist.co/file/anilistcdn/media/anime/cover/medium/bx153518-LEK6pAXtI03D.jpg",
			},
			downloadDir: t.TempDir(),
			expectedNum: 2,
			cancelAfter: 0,
		},
		//{
		//	name:        "test1",
		//	urls:        []string{"https://s4.anilist.co/file/anilistcdn/media/anime/banner/153518-7uRvV7SLqmHVn.jpg"},
		//	downloadDir: t.TempDir(),
		//	cancelAfter: 0,
		//},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			id := NewImageDownloader(tt.downloadDir, util.NewLogger())

			if tt.cancelAfter > 0 {
				go func() {
					time.Sleep(time.Duration(tt.cancelAfter) * time.Second)
					close(id.cancelChannel)
				}()
			}

			fmt.Print(tt.downloadDir)

			if err := id.DownloadImages(tt.urls); err != nil {
				t.Errorf("ImageDownloader.DownloadImages() error = %v", err)
			}

			downloadedImages := make(map[string]string, 0)
			for _, url := range tt.urls {
				imgPath, ok := id.GetImageFilenameByUrl(url)
				downloadedImages[imgPath] = imgPath
				if !ok {
					t.Errorf("ImageDownloader.GetImagePathByUrl() error")
				} else {
					t.Logf("ImageDownloader.GetImagePathByUrl() = %v", imgPath)
				}
			}

			require.Len(t, downloadedImages, tt.expectedNum)
		})

	}

	time.Sleep(1 * time.Second)
}
