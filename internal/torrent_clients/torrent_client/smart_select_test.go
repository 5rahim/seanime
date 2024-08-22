package torrent_client

//func TestSmartSelect(t *testing.T) {
//	t.Skip("Refactor test")
//	test_utils.InitTestProvider(t, test_utils.TorrentClient())
//
//	_ = t.TempDir()
//
//	anilistClient := anilist.TestGetMockAnilistClient()
//	_ = anilist_platform.NewAnilistPlatform(anilistClient, util.NewLogger())
//
//	// get repo
//
//	tests := []struct {
//		name             string
//		mediaId          int
//		url              string
//		selectedEpisodes []int
//		client           string
//	}{
//		{
//			name:             "Kakegurui xx (Season 2)",
//			mediaId:          100876,
//			url:              "https://nyaa.si/view/1553978", // kakegurui season 1 + season 2
//			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12 in season 2
//			client:           QbittorrentClient,
//		},
//		{
//			name:             "Spy x Family",
//			mediaId:          140960,
//			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
//			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12
//			client:           QbittorrentClient,
//		},
//		{
//			name:             "Spy x Family Part 2",
//			mediaId:          142838,
//			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
//			selectedEpisodes: []int{10, 11, 12, 13},          // should select 22, 23, 24, 25
//			client:           QbittorrentClient,
//		},
//		{
//			name:             "Kakegurui xx (Season 2)",
//			mediaId:          100876,
//			url:              "https://nyaa.si/view/1553978", // kakegurui season 1 + season 2
//			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12 in season 2
//			client:           TransmissionClient,
//		},
//		{
//			name:             "Spy x Family",
//			mediaId:          140960,
//			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
//			selectedEpisodes: []int{10, 11, 12},              // should select 10, 11, 12
//			client:           TransmissionClient,
//		},
//		{
//			name:             "Spy x Family Part 2",
//			mediaId:          142838,
//			url:              "https://nyaa.si/view/1661695", // spy x family (01-25)
//			selectedEpisodes: []int{10, 11, 12, 13},          // should select 22, 23, 24, 25
//			client:           TransmissionClient,
//		},
//	}
//
//	for _, tt := range tests {
//
//		t.Run(tt.name, func(t *testing.T) {
//
//			repo := getTestRepo(t, tt.client)
//
//			ok := repo.Start()
//			if !assert.True(t, ok) {
//				return
//			}
//
//		})
//
//	}
//
//}
