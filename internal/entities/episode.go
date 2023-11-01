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

		// Used for settings
		FileMetadata    *LocalFileMetadata         `json:"fileMetadata"`
		EpisodeMetadata *MediaEntryEpisodeMetadata `json:"episodeMetadata"`
		IsInvalid       bool                       `json:"isInvalid"`
		// Alerts the user that there is a discrepancy between AniList and AniDB
		MetadataIssue string `json:"metadataIssue,omitempty"`
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
	ep := new(MediaEntryEpisode)

	ep.DisplayTitle = ""
	ep.EpisodeTitle = ""

	if *opts.media.GetFormat() == anilist.MediaFormatMovie {
		ep.DisplayTitle = opts.media.GetPreferredTitle()
		ep.EpisodeTitle = "Complete Movie"
	}

	hydrated := false

	// LocalFile exists
	if opts.localFile != nil {

		aniDBEpisode := opts.localFile.Metadata.AniDBEpisode

		// progressOffset is -1, meaning the hydrator mistakenly set AniDB episode to "S1" (due to torrent name) because the episode number is 0
		// The hydrator ASSUMES that AniDB will not include episode 0 as part of main episodes.
		// We will remap "S1" to "1" and offset other AniDB episodes by 1
		// e.g, ["S1", "1", "2", "3",...,"12"] -> ["1", "2", "3", "4",...,"13"]
		if opts.progressOffset == -1 && opts.localFile.Metadata.Type == LocalFileTypeMain {
			if aniDBEpisode == "S1" {
				aniDBEpisode = "1"
				opts.progressOffset = 0
			} else {
				// e.g, "1" -> "2" etc...
				aniDBEpisode = anizip.OffsetEpisode(aniDBEpisode, opts.progressOffset)
			}
		}

		anizipEpisode, foundAnizipEpisode := opts.anizipMedia.GetEpisode(aniDBEpisode)

		ep.IsDownloaded = true
		ep.FileMetadata = opts.localFile.Metadata
		ep.Type = opts.localFile.Metadata.Type
		ep.LocalFile = opts.localFile

		// Set episode number and progress number
		switch opts.localFile.Metadata.Type {
		case LocalFileTypeMain:
			ep.EpisodeNumber = opts.localFile.Metadata.Episode
			ep.ProgressNumber = opts.localFile.Metadata.Episode + opts.progressOffset
		case LocalFileTypeSpecial:
			ep.EpisodeNumber = opts.localFile.Metadata.Episode
			ep.ProgressNumber = 0
		case LocalFileTypeNC:
			ep.EpisodeNumber = 0
			ep.ProgressNumber = 0
		}

		// Set titles if it has not been set
		if len(ep.DisplayTitle) == 0 && foundAnizipEpisode {
			switch opts.localFile.Metadata.Type {
			case LocalFileTypeMain:
				ep.DisplayTitle = "Episode " + strconv.Itoa(opts.localFile.Metadata.Episode)
				ep.EpisodeTitle = anizipEpisode.Title["en"]
			case LocalFileTypeSpecial:
				episodeInt, found := anizip.GetEpisodeInteger(aniDBEpisode)
				if found {
					ep.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
				} else {
					ep.DisplayTitle = "Special " + aniDBEpisode
				}
				ep.EpisodeTitle = anizipEpisode.Title["en"]
			case LocalFileTypeNC:
				ep.DisplayTitle = opts.localFile.GetParsedTitle()
				ep.EpisodeTitle = ""
			}

			ep.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode)

			hydrated = true
		}

	}

	// LocalFile does not exist
	if !hydrated && len(opts.optionalAniDBEpisode) > 0 {

		anizipEpisode, foundAnizipEpisode := opts.anizipMedia.GetEpisode(opts.optionalAniDBEpisode)

		if foundAnizipEpisode {
			ep.IsDownloaded = false
			ep.Type = LocalFileTypeMain
			if strings.HasPrefix(opts.optionalAniDBEpisode, "S") {
				ep.Type = LocalFileTypeSpecial
			} else if strings.HasPrefix(opts.optionalAniDBEpisode, "OP") || strings.HasPrefix(opts.optionalAniDBEpisode, "ED") {
				ep.Type = LocalFileTypeNC
			}
			ep.EpisodeNumber = 0
			ep.ProgressNumber = 0

			episodeInt, ok := anizip.GetEpisodeInteger(opts.optionalAniDBEpisode)
			if ok {
				ep.EpisodeNumber = episodeInt
				switch ep.Type {
				case LocalFileTypeMain:
					ep.DisplayTitle = "Episode " + strconv.Itoa(episodeInt)
					ep.EpisodeTitle = anizipEpisode.Title["en"]
				case LocalFileTypeSpecial:
					ep.DisplayTitle = "Special " + strconv.Itoa(episodeInt)
					ep.EpisodeTitle = anizipEpisode.Title["en"]
				case LocalFileTypeNC:
					ep.DisplayTitle = opts.optionalAniDBEpisode
					ep.EpisodeTitle = ""
				}
				hydrated = true
			}
			ep.EpisodeMetadata = NewEpisodeMetadata(anizipEpisode)
		}
	}

	if !hydrated {
		ep.IsInvalid = true
		return ep
	}

	return ep
}

func NewEpisodeMetadata(episode *anizip.Episode) *MediaEntryEpisodeMetadata {
	if episode == nil {
		return nil
	}

	md := new(MediaEntryEpisodeMetadata)

	md.Image = episode.Image
	md.AirDate = episode.Airdate
	md.Length = episode.Length
	if episode.Runtime > 0 {
		md.Length = episode.Runtime
	}
	md.Summary = episode.Summary
	md.Overview = episode.Overview

	return md
}
