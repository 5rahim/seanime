package entities

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
	"strconv"
	"strings"
)

type (
	MediaEntryEpisode struct {
		Type LocalFileType `json:"type"`
		// Formatted title
		// e.g, Show: "Episode 1", Movie: "Violet Evergarden The Movie"
		DisplayTitle string `json:"displayTitle"`
		// e.g, "Shibuya Incident - Gate, Open"
		EpisodeTitle string `json:"episodeTitle"`

		EpisodeNumber int `json:"episodeNumber"`
		// ProgressNumber is the source of truth for tracking purposes. It should exactly map to AniList.
		// Usually the same as EpisodeNumber, unless episode 0 is included by AniList.
		// e.g, Movie: 1, Show: 0,1,2...
		ProgressNumber int        `json:"progressNumber"`
		LocalFile      *LocalFile `json:"localFile"`
		IsDownloaded   bool       `json:"isDownloaded"`

		EpisodeMetadata *MediaEntryEpisodeMetadata `json:"episodeMetadata"`
		// Used for settings
		FileMetadata *LocalFileMetadata `json:"fileMetadata"`
		IsInvalid    bool               `json:"isInvalid"`
		// Alerts the user that there is a discrepancy between AniList and AniDB
		MetadataIssue string              `json:"metadataIssue,omitempty"`
		BasicMedia    *anilist.BasicMedia `json:"basicMedia,omitempty"`
	}

	MediaEntryEpisodeMetadata struct {
		Image    string `json:"image,omitempty"`
		AirDate  string `json:"airDate,omitempty"`
		Length   int    `json:"length,omitempty"`
		Summary  string `json:"summary,omitempty"`
		Overview string `json:"overview,omitempty"`
	}

	NewMediaEntryEpisodeOptions struct {
		localFile            *LocalFile
		anizipMedia          *anizip.Media
		media                *anilist.BaseMedia
		optionalAniDBEpisode string
		// progressOffset will offset the ProgressNumber for a specific MAIN file
		// This is used when there is a discrepancy between AniList and AniDB
		// When this is -1, it means that a re-mapping of AniDB Episode is needed
		progressOffset int
		isDownloaded   bool
	}
)

// NewMediaEntryEpisode creates a new episode entity.
//
// It is used to list existing local files as episodes
// OR list non-downloaded episodes by passing the `optionalAniDBEpisode` parameter.
//
// `anizipMedia` should be defined.
// `localFile` is optional.
func NewMediaEntryEpisode(opts *NewMediaEntryEpisodeOptions) *MediaEntryEpisode {
	entryEp := new(MediaEntryEpisode)
	entryEp.BasicMedia = opts.media.ToBasicMedia()
	entryEp.DisplayTitle = ""
	entryEp.EpisodeTitle = ""

	if *opts.media.GetFormat() == anilist.MediaFormatMovie {
		entryEp.DisplayTitle = opts.media.GetPreferredTitle()
		entryEp.EpisodeTitle = "Complete Movie"
	}

	hydrated := false

	// LocalFile exists
	if opts.localFile != nil {

		aniDBEp := opts.localFile.Metadata.AniDBEpisode

		// progressOffset is -1, meaning the hydrator mistakenly set AniDB episode to "S1" (due to torrent name) because the episode number is 0
		// The hydrator ASSUMES that AniDB will not include episode 0 as part of main episodes.
		// We will remap "S1" to "1" and offset other AniDB episodes by 1
		// e.g, ["S1", "1", "2", "3",...,"12"] -> ["1", "2", "3", "4",...,"13"]
		if opts.progressOffset == -1 && opts.localFile.Metadata.Type == LocalFileTypeMain {
			if aniDBEp == "S1" {
				aniDBEp = "1"
				opts.progressOffset = 0
			} else {
				// e.g, "1" -> "2" etc...
				aniDBEp = anizip.OffsetEpisode(aniDBEp, opts.progressOffset)
			}
			entryEp.MetadataIssue = "forced_remapping"
		}

		anizipEpisode, foundAnizipEpisode := opts.anizipMedia.GetEpisode(aniDBEp)

		entryEp.IsDownloaded = true
		entryEp.FileMetadata = opts.localFile.Metadata
		entryEp.Type = opts.localFile.Metadata.Type
		entryEp.LocalFile = opts.localFile

		// Set episode number and progress number
		switch opts.localFile.Metadata.Type {
		case LocalFileTypeMain:
			entryEp.EpisodeNumber = opts.localFile.Metadata.Episode
			entryEp.ProgressNumber = opts.localFile.Metadata.Episode + opts.progressOffset
		case LocalFileTypeSpecial:
			entryEp.EpisodeNumber = opts.localFile.Metadata.Episode
			entryEp.ProgressNumber = 0
		case LocalFileTypeNC:
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0
		}

		// Set titles
		if len(entryEp.DisplayTitle) == 0 {
			switch opts.localFile.Metadata.Type {
			case LocalFileTypeMain:
				if foundAnizipEpisode {
					entryEp.DisplayTitle = "Episode " + strconv.Itoa(opts.localFile.Metadata.Episode)
					entryEp.EpisodeTitle = anizipEpisode.GetTitle()
					hydrated = true // Hydrated
				}
			case LocalFileTypeSpecial:
				if foundAnizipEpisode {
					episodeInt, found := anizip.GetEpisodeInteger(aniDBEp)
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
					entryEp.DisplayTitle = opts.localFile.GetParsedTitle()
					entryEp.EpisodeTitle = ""
					hydrated = true // Hydrated
				}
			}

			entryEp.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode, opts.media)

		}

	}

	// LocalFile does not exist
	if !hydrated && len(opts.optionalAniDBEpisode) > 0 {

		if anizipEpisode, foundAnizipEpisode := opts.anizipMedia.GetEpisode(opts.optionalAniDBEpisode); foundAnizipEpisode {

			entryEp.IsDownloaded = false
			entryEp.Type = LocalFileTypeMain
			if strings.HasPrefix(opts.optionalAniDBEpisode, "S") {
				entryEp.Type = LocalFileTypeSpecial
			} else if strings.HasPrefix(opts.optionalAniDBEpisode, "OP") || strings.HasPrefix(opts.optionalAniDBEpisode, "ED") {
				entryEp.Type = LocalFileTypeNC
			}
			entryEp.EpisodeNumber = 0
			entryEp.ProgressNumber = 0

			if episodeInt, ok := anizip.GetEpisodeInteger(opts.optionalAniDBEpisode); ok {
				entryEp.EpisodeNumber = episodeInt
				switch entryEp.Type {
				case LocalFileTypeMain:
					entryEp.DisplayTitle = "Episode " + strconv.Itoa(episodeInt)
					entryEp.EpisodeTitle = anizipEpisode.GetTitle()
				case LocalFileTypeSpecial:
					entryEp.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					entryEp.EpisodeTitle = anizipEpisode.GetTitle()
				case LocalFileTypeNC:
					entryEp.DisplayTitle = opts.optionalAniDBEpisode
					entryEp.EpisodeTitle = ""
				}
				hydrated = true
			}

			entryEp.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode, opts.media)
		}

	}

	if !hydrated {
		entryEp.IsInvalid = true
		return entryEp
	}

	return entryEp
}

func NewEpisodeMetadata(episode *anizip.Episode, media *anilist.BaseMedia) *MediaEntryEpisodeMetadata {
	if episode == nil {
		return nil
	}

	md := new(MediaEntryEpisodeMetadata)

	md.Image = episode.Image
	if len(episode.Image) == 0 {
		md.Image = *media.GetBannerImage()
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
