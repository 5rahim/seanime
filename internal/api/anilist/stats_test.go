package anilist

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"github.com/seanime-app/seanime/internal/test_utils"
	"github.com/seanime-app/seanime/internal/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStats(t *testing.T) {
	test_utils.InitTestProvider(t, test_utils.Anilist())

	logger := util.NewLogger()

	requestBody, err := json.Marshal(map[string]interface{}{
		"query":     StatsQuery,
		"variables": map[string]interface{}{},
	})
	require.NoError(t, err)

	data, err := customQuery(requestBody, logger, test_utils.ConfigData.Provider.AnilistJwt)
	require.NoError(t, err)

	spew.Dump(data)

}

const StatsQuery = `
query GetStats {
  Viewer {
    statistics {
      anime {
        count
        minutesWatched
        episodesWatched
        meanScore
        formats {
          ...UserFormatStats
        }
        genres {
          ...UserGenreStats
        }
        statuses {
          ...UserStatusStats
        }
        studios {
          ...UserStudioStats
        }
      }
      manga {
        count
        chaptersRead
        meanScore
        formats {
          ...UserFormatStats
        }
        genres {
          ...UserGenreStats
        }
        statuses {
          ...UserStatusStats
        }
        studios {
          ...UserStudioStats
        }
      }
    }
  }
}

fragment UserFormatStats on UserFormatStatistic {
  format
  meanScore
  count
  minutesWatched
  mediaIds
  chaptersRead
}

fragment UserGenreStats on UserGenreStatistic {
  genre
  meanScore
  count
  minutesWatched
  mediaIds
  chaptersRead
}

fragment UserStatusStats on UserStatusStatistic {
  status
  meanScore
  count
  minutesWatched
  mediaIds
  chaptersRead
}

fragment UserStudioStats on UserStudioStatistic {
  studio {
    id
    name
    isAnimationStudio
  }
  meanScore
  count
  minutesWatched
  mediaIds
  chaptersRead
}
`
