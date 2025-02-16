package extension_playground

import (
	"os"
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"seanime/internal/extension"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGojaAnimeTorrentProvider(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t, test_utils.Anilist())

	logger := util.NewLogger()

	anilistClient := anilist.TestGetMockAnilistClient()
	platform := anilist_platform.NewAnilistPlatform(anilistClient, logger)
	metadataProvider := metadata.GetMockProvider(t)

	repo := NewPlaygroundRepository(logger, platform, metadataProvider)

	// Get the script
	filepath := "../extension_repo/goja_torrent_test/my-torrent-provider.ts"
	fileB, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatal(err)
	}

	params := RunPlaygroundCodeParams{
		Type:     extension.TypeAnimeTorrentProvider,
		Language: extension.LanguageTypescript,
		Code:     string(fileB),
		Inputs:   nil,
		Function: "",
	}

	tests := []struct {
		name     string
		inputs   map[string]interface{}
		function string
	}{
		{
			name:     "Search",
			function: "search",
			inputs: map[string]interface{}{
				"query":   "One Piece",
				"mediaId": 21,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params.Function = tt.function
			params.Inputs = tt.inputs

			resp, err := repo.RunPlaygroundCode(&params)
			require.NoError(t, err)

			t.Log("Logs:")

			t.Log(resp.Logs)

			t.Log("\n\nValue:")

			t.Log(resp.Value)
		})
	}

}
