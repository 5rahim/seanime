package seanime_parser

import (
	"testing"
)

func TestEpisodes(t *testing.T) {

	tests := []struct {
		input            string
		expectedTknValue []string
		debug            bool
	}{
		{"Jujutsu Kaisen Season 01 - 01-13", []string{"01", "13"}, false},
		{"Bleach 225", []string{"225"}, false},
		{"[Conclave-Mendoi]_Mobile_Suit_Gundam_00_S2_-_01v2_[1280x720_H.264_AAC][4863FBE8].mkv", []string{"01v2"}, false},
		{"NieR:Automata Ver1.1a - 01", []string{"01"}, false},
		{"NieR:Automata Ver1.1a - Ep01", []string{"01"}, false},
		{"[SubsPlease] Sousou no Frieren - 14", []string{"14"}, false},
		{"[SubsPlease] Sousou no Frieren - 14 (480p) [6EB72DA5].mkv", []string{"14"}, false},
		{"[SubsPlease] Sousou no Frieren - 14 480p 10 [6EB72DA5].mkv", []string{"14"}, false},
		{"[SubsPlease] Yuzuki-san Chi no Yonkyoudai - 10 (1080p) [6A9D6EE5].mkv", []string{"10"}, false},
		{"[chibi-Doki] Seikon no Qwaser - 13v0 (Uncensored Director's Cut) [988DB090].mkv", []string{"13v0"}, false},
		{"[Juuni.Kokki]-(Les.12.Royaumes)-[Ep.24]-[x264+OGG]-[JAP+FR+Sub.FR]-[Chap]-[AzF].mkv", []string{"24"}, false},
		{"[Taka]_Fullmetal_Alchemist_(2009)_04_[720p][40F2A957].mp4", []string{"04"}, false},
		{"[FuktLogik][Sayonara_Zetsubou_Sensei][01][DVDRip][x264_AC3].mkv", []string{"01"}, false},
		{"[Mobile Suit Gundam Seed Destiny HD REMASTER][07][Big5][720p][AVC_AAC][encoded by SEED].mp4", []string{"07"}, false},
		{"[Mobile Suit Gundam Seed Destiny HD REMASTER][12][Big5][720p][AVC_AAC][encoded by SEED].mp4", []string{"12"}, false},
		{"[52wy][SlamDunk][001][Jpn_Chs_Cht][x264_aac][DVDRip][7FE2C873].mkv", []string{"001"}, false},
		{"[Hatsuyuki] Dragon Ball Kai (2014) - 002 (100) [1280x720][DD66AFB7].mkv", []string{"002"}, false},
		{"[Coalgirls]_White_Album_1-13_(1280×720_Blu-Ray_FLAC)", []string{"1", "13"}, false},
		{"[Seanime]_One_Piece_800-994_(1280×720_Blu-Ray_FLAC)", []string{"800", "994"}, false},
		{"Code_Geass_R2_TV_[20_of_25]_[ru_jp]_[HDTV]_[Varies_&_Cuba77_&_AnimeReactor_RU].mkv", []string{"20", "25"}, false},
		{"【MMZYSUB】★【Golden Time】[24（END）][GB][720P_MP4]", []string{"24"}, false},
		{"[Trix] Shingeki no Kyojin - S04E29-31 (Part 3) [Multi Subs] (1080p AV1 E-AC3)", []string{"29", "31"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()

			assertMetadataExists(t, p, metadataEpisodeNumber, tt.expectedTknValue)

			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}
		})
	}

}

func TestOtherEpisodes(t *testing.T) {

	tests := []struct {
		input            string
		expectedTknValue []string
		debug            bool
	}{
		{"[Seanime] Jujutsu Kaisen SP1.mkv", []string{"1"}, false},
		{"[Seanime] Jujutsu Kaisen SP 1.5.mkv", []string{"1.5"}, false},
		{"[Seanime] Jujutsu Kaisen SP1.5.mkv", []string{"1.5"}, false},
		{"[Seanime] Jujutsu Kaisen SP 1.mkv", []string{"1"}, false},
		{"[Seanime] Jujutsu Kaisen OVA 01.mkv", []string{"01"}, false},
		{"[Seanime] Jujutsu Kaisen OVA1.mkv", []string{"1"}, false},
		{"[Seanime] Jujutsu Kaisen NCED1.mkv", []string{"1"}, false},
		{"[Seanime] Jujutsu Kaisen Movie 1.mkv", []string{"1"}, false},
		{"[Seanime] Jujutsu Kaisen Movies 1 ~ 3.mkv", []string{"1", "3"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := newParser(tt.input)
			p.parse()

			assertMetadataExists(t, p, metadataOtherEpisodeNumber, tt.expectedTknValue)

			if tt.debug {
				t.Log(p.tokenManager.tokens.Sdump())
			}
		})
	}

}
