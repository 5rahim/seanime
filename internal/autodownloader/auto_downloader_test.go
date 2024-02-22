package autodownloader

import (
	"github.com/seanime-app/seanime/internal/models"
	"testing"
)

var mockSettings = &models.AutoDownloaderSettings{
	Provider:              "animetosho",
	Interval:              10,
	Enabled:               true,
	DownloadAutomatically: true,
}

func TestAutoDownloader_checkForNewEpisodes(t *testing.T) {

}
