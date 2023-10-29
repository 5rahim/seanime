package nyaa

import "fmt"

const (
	nyaaBaseURL    = "https://nyaa.si/?page=rss&q=+"
	sukebeiBaseURL = "https://sukebei.nyaa.si/?page=rss&q=+"
	nyaaView       = "https://nyaa.si/view/"
	sukebeiView    = "https://sukebei.nyaa.si/view/"

	sortByComments  = "&s=comments&o=desc"
	sortBySeeders   = "&s=seeders&o=desc"
	sortByLeechers  = "&s=leechers&o=desc"
	sortByDownloads = "&s=downloads&o=desc"
	sortBySize      = "&s=size&o=desc"
	sortByDate      = "&s=id&o=desc"

	filterNoFilter    = "&f=0"
	filterNoRemakes   = "&f=1"
	filterTrustedOnly = "&f=2"

	categoryAll = "&c=0_0"

	categoryAnime       = "&c=1_0"
	categoryAnimeAMV    = "&c=1_1"
	categoryAnimeEng    = "&c=1_2"
	categoryAnimeNonEng = "&c=1_3"
	categoryAnimeRaw    = "&c=1_4"

	categoryAudio         = "&c=2_0"
	categoryAudioLossless = "&c=2_1"
	categoryAudioLossy    = "&c=2_2"

	categoryLiterature       = "&c=3_0"
	categoryLiteratureEng    = "&c=3_1"
	categoryLiteratureNonEng = "&c=3_2"
	categoryLiteratureRaw    = "&c=3_3"

	categoryLiveAction         = "&c=4_0"
	categoryLiveActionRaw      = "&c=4_4"
	categoryLiveActionEng      = "&c=4_1"
	categoryLiveActionNonEng   = "&c=4_3"
	categoryLiveActionIdolProm = "&c=4_2"

	categoryPictures         = "&c=5_0"
	categoryPicturesGraphics = "&c=5_1"
	categoryPicturesPhotos   = "&c=5_2"

	categorySoftware      = "&c=6_0"
	categorySoftwareApps  = "&c=6_1"
	categorySoftwareGames = "&c=6_2"

	categoryArt          = "&c=1_0"
	categoryArtAnime     = "&c=1_1"
	categoryArtDoujinshi = "&c=1_2"
	categoryArtGames     = "&c=1_3"
	categoryArtManga     = "&c=1_4"
	categoryArtPictures  = "&c=1_5"

	categoryRealLife       = "&c=2_0"
	categoryRealLifePhotos = "&c=2_1"
	categoryRealLifeVideos = "&c=2_2"
)

func buildURL(opts SearchOptions) (string, error) {
	var url string

	if opts.Provider == "nyaa" {
		url = nyaaBaseURL
	} else if opts.Provider == "sukebei" {
		url = sukebeiBaseURL
	} else {
		err := fmt.Errorf("provider option could be nyaa or sukebei\nsee docs: https://github.com/irevenko/go-nyaa#provider")
		return "", err
	}

	if opts.Query != "" {
		url += opts.Query
	}

	if opts.Provider == "nyaa" {
		if opts.Category != "" {
			switch opts.Category {
			case "all":
				url += categoryAll
			case "anime":
				url += categoryAnime
			case "anime-amv":
				url += categoryAnimeAMV
			case "anime-eng":
				url += categoryAnimeEng
			case "anime-non-eng":
				url += categoryAnimeNonEng
			case "anime-raw":
				url += categoryAnimeRaw
			case "audio":
				url += categoryAudio
			case "audio-lossless":
				url += categoryAudioLossless
			case "audio-lossy":
				url += categoryAudioLossy
			case "literature":
				url += categoryLiterature
			case "literature-eng":
				url += categoryLiteratureEng
			case "literature-non-eng":
				url += categoryLiteratureNonEng
			case "literature-raw":
				url += categoryLiteratureRaw
			case "live-action":
				url += categoryLiveAction
			case "live-action-raw":
				url += categoryLiveActionRaw
			case "live-action-eng":
				url += categoryLiveActionEng
			case "live-action-non-eng":
				url += categoryLiveActionNonEng
			case "live-action-idol-prom":
				url += categoryLiveActionIdolProm
			case "pictures":
				url += categoryPictures
			case "pictures-graphics":
				url += categoryPicturesGraphics
			case "pictures-photos":
				url += categoryPicturesPhotos
			case "software":
				url += categorySoftware
			case "software-apps":
				url += categorySoftwareApps
			case "software-games":
				url += categorySoftwareGames
			default:
				err := fmt.Errorf("such nyaa category option does not exitst\nsee docs: https://github.com/irevenko/go-nyaa#category")
				return "", err
			}
		}
	}

	if opts.Provider == "sukebei" {
		if opts.Category != "" {
			switch opts.Category {
			case "all":
				url += categoryAll
			case "art":
				url += categoryArt
			case "art-anime":
				url += categoryArtAnime
			case "art-doujinshi":
				url += categoryArtDoujinshi
			case "art-games":
				url += categoryArtGames
			case "art-manga":
				url += categoryArtManga
			case "art-pictures":
				url += categoryArtPictures
			case "real-life":
				url += categoryRealLife
			case "real-life-photos":
				url += categoryRealLifePhotos
			case "real-life-videos":
				url += categoryRealLifeVideos
			default:
				err := fmt.Errorf("such sukebei category option does not exitst\nsee docs: https://github.com/irevenko/go-nyaa#category")
				return "", err
			}
		}
	}

	if opts.SortBy != "" {
		switch opts.SortBy {
		case "downloads":
			url += sortByDownloads
		case "comments":
			url += sortByComments
		case "seeders":
			url += sortBySeeders
		case "leechers":
			url += sortByLeechers
		case "size":
			url += sortBySize
		case "date":
			url += sortByDate
		default:
			err := fmt.Errorf("such sort option does not exitst\nsee docs: https://github.com/irevenko/go-nyaa#sortby")
			return "", err
		}
	}

	if opts.Filter != "" {
		switch opts.Filter {
		case "no-filter":
			url += filterNoFilter
		case "no-remakes":
			url += filterNoRemakes
		case "trusted-only":
			url += filterTrustedOnly
		default:
			err := fmt.Errorf("such filter option does not exitst\nsee docs: https://github.com/irevenko/go-nyaa#filter")
			return "", err
		}
	}

	return url, nil
}
