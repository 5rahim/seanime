package entities

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/anizip"
	"strconv"
	"strings"
)

type (
	MediaEntryEpisode struct {
		Type                  LocalFileType              `json:"type"`
		DisplayTitle          string                     `json:"displayTitle"` // e.g, Show: "Episode 1", Movie: "Violet Evergarden The Movie"
		EpisodeTitle          string                     `json:"episodeTitle"` // e.g, "Shibuya Incident - Gate, Open"
		EpisodeNumber         int                        `json:"episodeNumber"`
		AbsoluteEpisodeNumber int                        `json:"absoluteEpisodeNumber"`
		ProgressNumber        int                        `json:"progressNumber"` // Usually the same as EpisodeNumber, unless episode 0 is included by AniList.
		LocalFile             *LocalFile                 `json:"localFile"`
		IsDownloaded          bool                       `json:"isDownloaded"`            // Is in the local files
		EpisodeMetadata       *MediaEntryEpisodeMetadata `json:"episodeMetadata"`         // (image, airDate, length, summary, overview)
		FileMetadata          *LocalFileMetadata         `json:"fileMetadata"`            // (episode, aniDBEpisode, type...)
		IsInvalid             bool                       `json:"isInvalid"`               // No AniZip data
		MetadataIssue         string                     `json:"metadataIssue,omitempty"` // Alerts the user that there is a discrepancy between AniList and AniDB
		BasicMedia            *anilist.BasicMedia        `json:"basicMedia,omitempty"`
	}

	MediaEntryEpisodeMetadata struct {
		AniDBId  int    `json:"aniDBId,omitempty"`
		Image    string `json:"image,omitempty"`
		AirDate  string `json:"airDate,omitempty"`
		Length   int    `json:"length,omitempty"`
		Summary  string `json:"summary,omitempty"`
		Overview string `json:"overview,omitempty"`
	}

	NewMediaEntryEpisodeOptions struct {
		LocalFile            *LocalFile
		AnizipMedia          *anizip.Media
		Media                *anilist.BaseMedia
		OptionalAniDBEpisode string
		// ProgressOffset will offset the ProgressNumber for a specific MAIN file
		// This is used when there is a discrepancy between AniList and AniDB
		// When this is -1, it means that a re-mapping of AniDB Episode is needed
		ProgressOffset int
		IsDownloaded   bool
	}

	NewSimpleMediaEntryEpisodeOptions struct {
		LocalFile    *LocalFile
		Media        *anilist.BaseMedia
		IsDownloaded bool
	}
)

// NewMediaEntryEpisode creates a new episode entity.
//
// It is used to list existing local files as episodes
// OR list non-downloaded episodes by passing the `OptionalAniDBEpisode` parameter.
//
// `AnizipMedia` should be defined.
// `LocalFile` is optional.
func NewMediaEntryEpisode(opts *NewMediaEntryEpisodeOptions) *MediaEntryEpisode {
	entryEp := new(MediaEntryEpisode)
	entryEp.BasicMedia = opts.Media.ToBasicMedia()
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
				aniDBEp = anizip.OffsetEpisode(aniDBEp, opts.ProgressOffset)
			}
			entryEp.MetadataIssue = "forced_remapping"
		}

		anizipEpisode, foundAnizipEpisode := opts.AnizipMedia.FindEpisode(aniDBEp)

		entryEp.IsDownloaded = true
		entryEp.FileMetadata = opts.LocalFile.GetMetadata()
		entryEp.Type = opts.LocalFile.GetType()
		entryEp.LocalFile = opts.LocalFile

		// Set episode number and progress number
		switch opts.LocalFile.Metadata.Type {
		case LocalFileTypeMain:
			entryEp.EpisodeNumber = opts.LocalFile.GetEpisodeNumber()
			entryEp.ProgressNumber = opts.LocalFile.GetEpisodeNumber() + opts.ProgressOffset
			if foundAnizipEpisode {
				entryEp.AbsoluteEpisodeNumber = entryEp.EpisodeNumber + opts.AnizipMedia.GetOffset()
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
				if foundAnizipEpisode {
					if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
						entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
						entryEp.EpisodeTitle = "Complete Movie"
					} else {
						entryEp.DisplayTitle = "Episode " + strconv.Itoa(opts.LocalFile.GetEpisodeNumber())
						entryEp.EpisodeTitle = anizipEpisode.GetTitle()
					}

					hydrated = true // Hydrated
				}
			case LocalFileTypeSpecial:
				if foundAnizipEpisode {
					episodeInt, found := anizip.ExtractEpisodeInteger(aniDBEp)
					if found {
						entryEp.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					} else {
						entryEp.DisplayTitle = "Special " + aniDBEp
					}
					entryEp.EpisodeTitle = anizipEpisode.GetTitle()
					hydrated = true // Hydrated
				}
			case LocalFileTypeNC:
				if foundAnizipEpisode {
					entryEp.DisplayTitle = anizipEpisode.GetTitle()
					entryEp.EpisodeTitle = ""
					hydrated = true // Hydrated
				} else {
					entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
					entryEp.EpisodeTitle = ""
					hydrated = true // Hydrated
				}
			}
		} else {
			hydrated = true // Hydrated
		}
		entryEp.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode, opts.Media)

	}

	// LocalFile does not exist
	if !hydrated && len(opts.OptionalAniDBEpisode) > 0 {

		// Get the AniZip episode
		if anizipEpisode, foundAnizipEpisode := opts.AnizipMedia.FindEpisode(opts.OptionalAniDBEpisode); foundAnizipEpisode {

			entryEp.IsDownloaded = false
			entryEp.Type = LocalFileTypeMain
			if strings.HasPrefix(opts.OptionalAniDBEpisode, "S") {
				entryEp.Type = LocalFileTypeSpecial
			} else if strings.HasPrefix(opts.OptionalAniDBEpisode, "OP") || strings.HasPrefix(opts.OptionalAniDBEpisode, "ED") {
				entryEp.Type = LocalFileTypeNC
			}
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0

			if episodeInt, ok := anizip.ExtractEpisodeInteger(opts.OptionalAniDBEpisode); ok {
				entryEp.EpisodeNumber = episodeInt
				if foundAnizipEpisode {
					entryEp.AbsoluteEpisodeNumber = entryEp.EpisodeNumber + opts.AnizipMedia.GetOffset()
				}
				switch entryEp.Type {
				case LocalFileTypeMain:
					if *opts.Media.GetFormat() == anilist.MediaFormatMovie {
						entryEp.DisplayTitle = opts.Media.GetPreferredTitle()
						entryEp.EpisodeTitle = "Complete Movie"
					} else {
						entryEp.DisplayTitle = "Episode " + strconv.Itoa(episodeInt)
						entryEp.EpisodeTitle = anizipEpisode.GetTitle()
					}
				case LocalFileTypeSpecial:
					entryEp.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					entryEp.EpisodeTitle = anizipEpisode.GetTitle()
				case LocalFileTypeNC:
					entryEp.DisplayTitle = opts.OptionalAniDBEpisode
					entryEp.EpisodeTitle = ""
				}
				hydrated = true
			}

			entryEp.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode, opts.Media)
		}

	}

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

func NewEpisodeMetadata(episode *anizip.Episode, media *anilist.BaseMedia) *MediaEntryEpisodeMetadata {
	md := new(MediaEntryEpisodeMetadata)

	// No AniZip data
	if episode == nil {
		md.Image = *media.GetCoverImage().GetLarge()
		return md
	}

	md.AniDBId = episode.AnidbEid

	md.Image = episode.Image
	if len(episode.Image) == 0 && media.GetBannerImage() != nil {
		md.Image = *media.GetBannerImage()
	}
	if len(md.Image) == 0 && media.GetCoverImage().GetLarge() != nil {
		md.Image = *media.GetCoverImage().GetLarge()
	}
	md.AirDate = episode.Airdate
	md.Length = episode.Length
	if episode.Runtime > 0 {
		md.Length = episode.Runtime
	}
	md.Summary = episode.Summary
	md.Overview = episode.Overview

	return md
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func NewSimpleMediaEntryEpisode(opts *NewSimpleMediaEntryEpisodeOptions) *MediaEntryEpisode {
	entryEp := new(MediaEntryEpisode)
	entryEp.BasicMedia = opts.Media.ToBasicMedia()
	entryEp.DisplayTitle = ""
	entryEp.EpisodeTitle = ""
	entryEp.EpisodeMetadata = new(MediaEntryEpisodeMetadata)

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
					entryEp.EpisodeTitle = opts.LocalFile.ParsedData.EpisodeTitle
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

		entryEp.EpisodeMetadata.Image = *opts.Media.GetCoverImage().GetLarge()

	}

	if !hydrated {
		if opts.LocalFile != nil {
			entryEp.DisplayTitle = opts.LocalFile.GetParsedTitle()
		}
		entryEp.EpisodeTitle = ""
		entryEp.IsInvalid = true
		entryEp.MetadataIssue = "no_anizip_data"
		return entryEp
	}

	entryEp.MetadataIssue = "no_anizip_data"
	return entryEp
}
