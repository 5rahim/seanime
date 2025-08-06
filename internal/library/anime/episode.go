package anime

import (
	"seanime/internal/api/anilist"
	"seanime/internal/api/metadata"
	"strconv"
	"strings"
)

type (
	// Episode represents a single episode of a media entry.
	Episode struct {
		Type                  LocalFileType      `json:"type"`
		DisplayTitle          string             `json:"displayTitle"` // e.g, Show: "Episode 1", Movie: "Violet Evergarden The Movie"
		EpisodeTitle          string             `json:"episodeTitle"` // e.g, "Shibuya Incident - Gate, Open"
		EpisodeNumber         int                `json:"episodeNumber"`
		AniDBEpisode          string             `json:"aniDBEpisode,omitempty"` // AniDB episode number
		AbsoluteEpisodeNumber int                `json:"absoluteEpisodeNumber"`
		ProgressNumber        int                `json:"progressNumber"` // Usually the same as EpisodeNumber, unless there is a discrepancy between AniList and AniDB
		LocalFile             *LocalFile         `json:"localFile"`
		IsDownloaded          bool               `json:"isDownloaded"`            // Is in the local files
		EpisodeMetadata       *EpisodeMetadata   `json:"episodeMetadata"`         // (image, airDate, length, summary, overview)
		FileMetadata          *LocalFileMetadata `json:"fileMetadata"`            // (episode, aniDBEpisode, type...)
		IsInvalid             bool               `json:"isInvalid"`               // No AniDB data
		MetadataIssue         string             `json:"metadataIssue,omitempty"` // Alerts the user that there is a discrepancy between AniList and AniDB
		BaseAnime             *anilist.BaseAnime `json:"baseAnime,omitempty"`
		// IsNakamaEpisode indicates that this episode is from the Nakama host's anime library.
		IsNakamaEpisode bool `json:"_isNakamaEpisode"`
	}

	// EpisodeMetadata represents the metadata of an Episode.
	// Metadata is fetched from Animap (AniDB) and, optionally, AniList (if Animap is not available).
	EpisodeMetadata struct {
		AnidbId  int    `json:"anidbId,omitempty"`
		Image    string `json:"image,omitempty"`
		AirDate  string `json:"airDate,omitempty"`
		Length   int    `json:"length,omitempty"`
		Summary  string `json:"summary,omitempty"`
		Overview string `json:"overview,omitempty"`
		IsFiller bool   `json:"isFiller,omitempty"`
		HasImage bool   `json:"hasImage,omitempty"` // Indicates if the episode has a real image
	}
)

type (
	// NewEpisodeOptions hold data used to create a new Episode.
	NewEpisodeOptions struct {
		LocalFile            *LocalFile
		AnimeMetadata        *metadata.AnimeMetadata // optional
		Media                *anilist.BaseAnime
		OptionalAniDBEpisode string
		// ProgressOffset will offset the ProgressNumber for a specific MAIN file
		// This is used when there is a discrepancy between AniList and AniDB
		// When this is -1, it means that a re-mapping of AniDB Episode is needed
		ProgressOffset   int
		IsDownloaded     bool
		MetadataProvider metadata.Provider // optional
	}

	// NewSimpleEpisodeOptions hold data used to create a new Episode.
	// Unlike NewEpisodeOptions, this struct does not require Animap data. It is used to list episodes without AniDB metadata.
	NewSimpleEpisodeOptions struct {
		LocalFile    *LocalFile
		Media        *anilist.BaseAnime
		IsDownloaded bool
	}
)

// NewEpisode creates a new episode entity.
//
// It is used to list existing local files as episodes
// OR list non-downloaded episodes by passing the `OptionalAniDBEpisode` parameter.
//
// `AnimeMetadata` should be defined, but this is not always the case.
// `LocalFile` is optional.
func NewEpisode(opts *NewEpisodeOptions) *Episode {
	entryEp := new(Episode)
	entryEp.BaseAnime = opts.Media
	entryEp.DisplayTitle = ""
	entryEp.EpisodeTitle = ""

	hydrated := false

	// LocalFile exists
	if opts.LocalFile != nil {

		aniDBEp := opts.LocalFile.Metadata.AniDBEpisode

		// ProgressOffset is -1, meaning the hydrator mistakenly set AniDB episode to "S1" (due to torrent name) because the episode number is 0
		// The hydrator ASSUMES that AniDB will not include episode 0 as part of main episodes.
		// We will remap "S1" to "1" and offset other AniDB episodes by 1
		// e.g, ["S1", "1", "2", "3",...,"12"] -> ["1", "2", "3", "4",...,"13"]
		if opts.ProgressOffset == -1 && opts.LocalFile.GetType() == LocalFileTypeMain {
			if aniDBEp == "S1" {
				aniDBEp = "1"
				opts.ProgressOffset = 0
			} else {
				// e.g, "1" -> "2" etc...
				aniDBEp = metadata.OffsetAnidbEpisode(aniDBEp, opts.ProgressOffset)
			}
			entryEp.MetadataIssue = "forced_remapping"
		}

		// Get the Animap episode
		foundAnimapEpisode := false
		var episodeMetadata *metadata.EpisodeMetadata
		if opts.AnimeMetadata != nil {
			episodeMetadata, foundAnimapEpisode = opts.AnimeMetadata.FindEpisode(aniDBEp)
		}

		entryEp.IsDownloaded = true
		entryEp.FileMetadata = opts.LocalFile.GetMetadata()
		entryEp.Type = opts.LocalFile.GetType()
		entryEp.LocalFile = opts.LocalFile

		// Set episode number and progress number
		switch opts.LocalFile.Metadata.Type {
		case LocalFileTypeMain:
			entryEp.EpisodeNumber = opts.LocalFile.GetEpisodeNumber()
			entryEp.ProgressNumber = opts.LocalFile.GetEpisodeNumber() + opts.ProgressOffset
			if foundAnimapEpisode {
				entryEp.AniDBEpisode = aniDBEp
				entryEp.AbsoluteEpisodeNumber = entryEp.EpisodeNumber + opts.AnimeMetadata.GetOffset()
			}
		case LocalFileTypeSpecial:
			entryEp.EpisodeNumber = opts.LocalFile.GetEpisodeNumber()
			entryEp.ProgressNumber = 0
		case LocalFileTypeNC:
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0
		}

		// Set titles
		if len(entryEp.DisplayTitle) == 0 {
			switch opts.LocalFile.Metadata.Type {
			case LocalFileTypeMain:
				if foundAnimapEpisode {
					entryEp.AniDBEpisode = aniDBEp
					if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
						entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
						entryEp.EpisodeTitle = "Complete Movie"
					} else {
						entryEp.DisplayTitle = "Episode " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
						entryEp.EpisodeTitle = episodeMetadata.GetTitle()
					}
				} else {
					if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
						entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
						entryEp.EpisodeTitle = "Complete Movie"
					} else {
						entryEp.DisplayTitle = "Episode " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
						entryEp.EpisodeTitle = opts.LocalFile.GetParsedEpisodeTitle()
					}
				}
				hydrated = true // Hydrated
			case LocalFileTypeSpecial:
				if foundAnimapEpisode {
					entryEp.AniDBEpisode = aniDBEp
					episodeInt, found := metadata.ExtractEpisodeInteger(aniDBEp)
					if found {
						entryEp.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					} else {
						entryEp.DisplayTitle = "Special " + aniDBEp
					}
					entryEp.EpisodeTitle = episodeMetadata.GetTitle()
				} else {
					entryEp.DisplayTitle = "Special " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
				}
				hydrated = true // Hydrated
			case LocalFileTypeNC:
				if foundAnimapEpisode {
					entryEp.AniDBEpisode = aniDBEp
					entryEp.DisplayTitle = episodeMetadata.GetTitle()
					entryEp.EpisodeTitle = ""
				} else {
					entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
					entryEp.EpisodeTitle = ""
				}
				hydrated = true // Hydrated
			}
		} else {
			hydrated = true // Hydrated
		}

		// Set episode metadata
		entryEp.EpisodeMetadata = NewEpisodeMetadata(opts.AnimeMetadata, episodeMetadata, opts.Media, opts.MetadataProvider)

	} else if len(opts.OptionalAniDBEpisode) > 0 && opts.AnimeMetadata != nil {
		// No LocalFile, but AniDB episode is provided

		// Get the Animap episode
		if episodeMetadata, foundAnimapEpisode := opts.AnimeMetadata.FindEpisode(opts.OptionalAniDBEpisode); foundAnimapEpisode {

			entryEp.IsDownloaded = false
			entryEp.Type = LocalFileTypeMain
			if strings.HasPrefix(opts.OptionalAniDBEpisode, "S") {
				entryEp.Type = LocalFileTypeSpecial
			} else if strings.HasPrefix(opts.OptionalAniDBEpisode, "OP") || strings.HasPrefix(opts.OptionalAniDBEpisode, "ED") {
				entryEp.Type = LocalFileTypeNC
			}
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0

			if episodeInt, ok := metadata.ExtractEpisodeInteger(opts.OptionalAniDBEpisode); ok {
				entryEp.EpisodeNumber = episodeInt
				entryEp.ProgressNumber = episodeInt + opts.ProgressOffset
				entryEp.AniDBEpisode = opts.OptionalAniDBEpisode
				entryEp.AbsoluteEpisodeNumber = entryEp.EpisodeNumber + opts.AnimeMetadata.GetOffset()
				switch entryEp.Type {
				case LocalFileTypeMain:
					if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
						entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
						entryEp.EpisodeTitle = "Complete Movie"
					} else {
						entryEp.DisplayTitle = "Episode " + strconv.Itoa(episodeInt)
						entryEp.EpisodeTitle = episodeMetadata.GetTitle()
					}
				case LocalFileTypeSpecial:
					entryEp.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					entryEp.EpisodeTitle = episodeMetadata.GetTitle()
				case LocalFileTypeNC:
					entryEp.DisplayTitle = opts.OptionalAniDBEpisode
					entryEp.EpisodeTitle = ""
				}
				hydrated = true
			}

			// Set episode metadata
			entryEp.EpisodeMetadata = NewEpisodeMetadata(opts.AnimeMetadata, episodeMetadata, opts.Media, opts.MetadataProvider)
		} else {
			// No Local file, no Animap data
			// DEVNOTE: Non-downloaded, without any AniDB data. Don't handle this case.
			// Non-downloaded episodes are determined from AniDB data either way.
		}

	}

	// If for some reason the episode is not hydrated, set it as invalid
	if !hydrated {
		if opts.LocalFile != nil {
			entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
		}
		entryEp.EpisodeTitle = ""
		entryEp.IsInvalid = true
		return entryEp
	}

	return entryEp
}

// NewEpisodeMetadata creates a new EpisodeMetadata from an Animap episode and AniList media.
// If the Animap episode is nil, it will just set the image from the media.
func NewEpisodeMetadata(
	animeMetadata *metadata.AnimeMetadata,
	episode *metadata.EpisodeMetadata,
	media *anilist.BaseAnime,
	metadataProvider metadata.Provider,
) *EpisodeMetadata {
	md := new(EpisodeMetadata)

	// No Animap data
	if episode == nil {
		md.Image = media.GetCoverImageSafe()
		return md
	}
	epInt, err := strconv.Atoi(episode.Episode)

	if err == nil {
		aw := metadataProvider.GetAnimeMetadataWrapper(media, animeMetadata)
		epMetadata := aw.GetEpisodeMetadata(epInt)
		md.AnidbId = epMetadata.AnidbId
		md.Image = epMetadata.Image
		md.AirDate = epMetadata.AirDate
		md.Length = epMetadata.Length
		md.Summary = epMetadata.Summary
		md.Overview = epMetadata.Overview
		md.HasImage = epMetadata.HasImage
		md.IsFiller = false
	} else {
		md.Image = media.GetBannerImageSafe()
	}

	return md
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// NewSimpleEpisode creates a Episode without AniDB metadata.
func NewSimpleEpisode(opts *NewSimpleEpisodeOptions) *Episode {
	entryEp := new(Episode)
	entryEp.BaseAnime = opts.Media
	entryEp.DisplayTitle = ""
	entryEp.EpisodeTitle = ""
	entryEp.EpisodeMetadata = new(EpisodeMetadata)

	hydrated := false

	// LocalFile exists
	if opts.LocalFile != nil {

		entryEp.IsDownloaded = true
		entryEp.FileMetadata = opts.LocalFile.GetMetadata()
		entryEp.Type = opts.LocalFile.GetType()
		entryEp.LocalFile = opts.LocalFile

		// Set episode number and progress number
		switch opts.LocalFile.Metadata.Type {
		case LocalFileTypeMain:
			entryEp.EpisodeNumber = opts.LocalFile.GetEpisodeNumber()
			entryEp.ProgressNumber = opts.LocalFile.GetEpisodeNumber()
			hydrated = true // Hydrated
		case LocalFileTypeSpecial:
			entryEp.EpisodeNumber = opts.LocalFile.GetEpisodeNumber()
			entryEp.ProgressNumber = 0
			hydrated = true // Hydrated
		case LocalFileTypeNC:
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0
			hydrated = true // Hydrated
		}

		// Set titles
		if len(entryEp.DisplayTitle) == 0 {
			switch opts.LocalFile.Metadata.Type {
			case LocalFileTypeMain:
				if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
					entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
					entryEp.EpisodeTitle = "Complete Movie"
				} else {
					entryEp.DisplayTitle = "Episode " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
					entryEp.EpisodeTitle = opts.LocalFile.GetParsedEpisodeTitle()
				}

				hydrated = true // Hydrated
			case LocalFileTypeSpecial:
				entryEp.DisplayTitle = "Special " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
				hydrated = true // Hydrated
			case LocalFileTypeNC:
				entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
				entryEp.EpisodeTitle = ""
				hydrated = true // Hydrated
			}
		}

		entryEp.EpisodeMetadata.Image = opts.Media.GetCoverImageSafe()

	}

	if !hydrated {
		if opts.LocalFile != nil {
			entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
		}
		entryEp.EpisodeTitle = ""
		entryEp.IsInvalid = true
		entryEp.MetadataIssue = "no_anidb_data"
		return entryEp
	}

	entryEp.MetadataIssue = "no_anidb_data"
	return entryEp
}
