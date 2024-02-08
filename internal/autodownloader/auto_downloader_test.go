package autodownloader

import (
	"github.com/seanime-app/seanime/internal/entities"
	"github.com/seanime-app/seanime/internal/models"
	"github.com/seanime-app/seanime/internal/util"
	"testing"
)

var mockSettings = &models.AutoDownloaderSettings{
	Provider:              "nyaa",
	Interval:              10,
	Enabled:               true,
	DownloadAutomatically: true,
}

func TestAutoDownloader_checkForNewEpisodes(t *testing.T) {

	rule := &entities.AutoDownloaderRule{
		DbID:                1,
		Enabled:             true,
		MediaId:             21,
		ReleaseGroups:       []string{"SubsPlease"},
		Resolutions:         []string{"1080p"},
		ComparisonTitle:     "One Piece",
		TitleComparisonType: entities.AutoDownloaderRuleTitleComparisonLikely,
		EpisodeType:         entities.AutoDownloaderRuleEpisodeUnwatched,
		EpisodeNumbers:      []int{},
		Destination:         "E:/ANIME/Test",
	}

	// Create a new AutoDownloader
	ad := NewAutoDownloader(&NewAutoDownloaderOptions{
		Logger:            util.NewLogger(),
		QbittorrentClient: nil,
		WSEventManager:    nil,
		Rules:             []*entities.AutoDownloaderRule{*&rule},
		Database:          nil,
	})
	ad.SetSettings(mockSettings)

	// Check for new episodes
	ad.checkForNewEpisodes()

}
