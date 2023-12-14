package seanime_parser

import "testing"

func TestTitle(t *testing.T) {

	tests := []struct {
		input    string
		expected string
		debug    bool
	}{
		{"Jujutsu Kaisen Season 01 - 01-13", "Jujutsu Kaisen", false},
		{"Bleach 225", "Bleach", false},
		{"[Conclave-Mendoi]_Mobile_Suit_Gundam_00_S2_-_01v2_[1280x720_H.264_AAC][4863FBE8].mkv", "Mobile Suit Gundam 00", false},
		{"NieR:Automata Ver1.1a - 01", "NieR:Automata Ver1.1a", false},
		{"[SubsPlease] Sousou no Frieren - 14", "Sousou no Frieren", false},
		{"[SubsPlease] Sousou no Frieren - 14 (480p) [6EB72DA5].mkv", "Sousou no Frieren", false},
		{"[SubsPlease] Sousou no Frieren - 14 480p [6EB72DA5].mkv", "Sousou no Frieren", false},
		{"[SubsPlease] Yuzuki-san Chi no Yonkyoudai - 10 (1080p) [6A9D6EE5].mkv", "Yuzuki-san Chi no Yonkyoudai", false},
		{"[chibi-Doki] Seikon no Qwaser - 13v0 (Uncensored Director's Cut) [988DB090].mkv", "Seikon no Qwaser", false},
		{"[Juuni.Kokki]-(Les.12.Royaumes)-[Ep.24]-[x264+OGG]-[JAP+FR+Sub.FR]-[Chap]-[AzF].mkv", "Les 12 Royaumes", false},
		{"[Taka]_Fullmetal_Alchemist_(2009)_04_[720p][40F2A957].mp4", "Fullmetal Alchemist", false},
		{"[FuktLogik][Sayonara_Zetsubou_Sensei][01][DVDRip][x264_AC3].mkv", "Sayonara Zetsubou Sensei", false},
		{"[Mobile Suit Gundam Seed Destiny HD REMASTER][07][Big5][720p][AVC_AAC][encoded by SEED].mp4", "Mobile Suit Gundam Seed Destiny", false},
		{"[52wy][SlamDunk][001][Jpn_Chs_Cht][x264_aac][DVDRip][7FE2C873].mkv", "SlamDunk", false},
		{"[Hatsuyuki] Dragon Ball Kai (2014) - 002 (100) [1280x720][DD66AFB7].mkv", "Dragon Ball Kai", false},
		{"[Coalgirls]_White_Album_1-13_(1280×720_Blu-Ray_FLAC)", "White Album", false},
		{"[Seanime]_One_Piece_800-994_(1280×720_Blu-Ray_FLAC)", "One Piece", false},
		{"Code_Geass_R2_TV_[20_of_25]_[ru_jp]_[HDTV]_[Varies_&_Cuba77_&_AnimeReactor_RU].mkv", "Code Geass R2 TV", false},
		{"[Urusai]_Bokura_Ga_Ita_01_[DVD_h264_AC3]_[BFCE1627][Fixed].mkv", "Bokura Ga Ita", false},
		{"SPY x FAMILY S02E09 The Hand That Connects to the Future 1080p NF WEB-DL AAC2.0 H 264-VARYG", "SPY x FAMILY", false},
		{"[Jumonji-Giri]_[Shinsen-Subs][ASF]_D.C.II_Da_Capo_II_Ep01_(a1fc58a7).mkv", "D.C.II Da Capo II", false},
		{"[Hakugetsu&Speed&MGRT][Dragon_Ball_Z_Battle_of_Gods][BDRIP][BIG5][1280x720].mp4", "Dragon Ball Z Battle of Gods", false},
		{"[Hakugetsu&MGRT][Evangelion 3.0 You Can (Not) Redo][480P][V0].mp4", "Evangelion 3.0 You Can (Not) Redo", false},
		{"Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER", "Violet Evergarden The Movie", false},
		{"【MMZYSUB】★【Golden Time】[24（END）][GB][720P_MP4]", "Golden Time", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()

			assertMetadataExists(t, p, metadataTitle, []string{tt.expected})

			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}
		})
	}

}

func TestReleaseGroup(t *testing.T) {

	tests := []struct {
		input    string
		expected string
		debug    bool
	}{
		{"Jujutsu Kaisen Season 01 - 01-13", "", false},
		{"Bleach 225", "", false},
		{"[Conclave-Mendoi]_Mobile_Suit_Gundam_00_S2_-_01v2_[1280x720_H.264_AAC][4863FBE8].mkv", "Conclave-Mendoi", false},
		{"NieR:Automata Ver1.1a - 01", "", false},
		{"[SubsPlease] Sousou no Frieren - 14", "SubsPlease", false},
		{"[SubsPlease] Sousou no Frieren - 14 (480p) [6EB72DA5].mkv", "SubsPlease", false},
		{"[SubsPlease] Yuzuki-san Chi no Yonkyoudai - 10 (1080p) [6A9D6EE5].mkv", "SubsPlease", false},
		{"[chibi-Doki] Seikon no Qwaser - 13v0 (Uncensored Director's Cut) [988DB090].mkv", "chibi-Doki", false},
		{"[Juuni.Kokki]-(Les.12.Royaumes)-[Ep.24]-[x264+OGG]-[JAP+FR+Sub.FR]-[Chap]-[AzF].mkv", "Juuni Kokki", false},
		{"[Taka]_Fullmetal_Alchemist_(2009)_04_[720p][40F2A957].mp4", "Taka", false},
		{"[FuktLogik][Sayonara_Zetsubou_Sensei][01][DVDRip][x264_AC3].mkv", "FuktLogik", false},
		{"[Mobile Suit Gundam Seed Destiny HD REMASTER][07][Big5][720p][AVC_AAC][encoded by SEED].mp4", "encoded by SEED", false},
		{"[52wy][SlamDunk][001][Jpn_Chs_Cht][x264_aac][DVDRip][7FE2C873].mkv", "52wy", false},
		{"[Hatsuyuki] Dragon Ball Kai (2014) - 002 (100) [1280x720][DD66AFB7].mkv", "Hatsuyuki", false},
		{"[Coalgirls]_White_Album_1-13_(1280×720_Blu-Ray_FLAC)", "Coalgirls", false},
		{"[Seanime]_One_Piece_800-994_(1280×720_Blu-Ray_FLAC)", "Seanime", false},
		{"Code_Geass_R2_TV_[20_of_25]_[ru_jp]_[HDTV]_[Varies_&_Cuba77_&_AnimeReactor_RU].mkv", "Varies & Cuba77 & AnimeReactor RU", false},
		{"[Urusai]_Bokura_Ga_Ita_01_[DVD_h264_AC3]_[BFCE1627][Fixed].mkv", "Urusai", false},
		{"SPY x FAMILY S02E09 The Hand That Connects to the Future 1080p NF WEB-DL AAC2.0 H 264-VARYG", "VARYG", false},
		{"[Jumonji-Giri]_[Shinsen-Subs][ASF]_D.C.II_Da_Capo_II_Ep01_(a1fc58a7).mkv", "Jumonji-Giri Shinsen-Subs", false},
		{"[Hakugetsu&Speed&MGRT][Dragon_Ball_Z_Battle_of_Gods][BDRIP][BIG5][1280x720].mp4", "Hakugetsu&Speed&MGRT", false},
		{"[Hakugetsu&MGRT][Evangelion 3.0 You Can (Not) Redo][480P][V0].mp4", "Hakugetsu&MGRT", false},
		{"Violet.Evergarden.The.Movie.1080p.Dual.Audio.BDRip.10.bits.DD.x265-EMBER", "EMBER", false},
		{"【MMZYSUB】★【Golden Time】[24（END）][GB][720P_MP4]", "MMZYSUB", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()

			assertMetadataExists(t, p, metadataReleaseGroup, []string{tt.expected})

			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}
		})
	}

}
