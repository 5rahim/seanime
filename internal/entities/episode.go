package entities

import (
	"github.com/seanime-app/seanime-server/internal/anilist"
	"github.com/seanime-app/seanime-server/internal/anizip"
)

type (
	MediaEntryEpisode struct {
		// Formatted title
		// e.g, Show: "Episode 1", Movie: "Violet Evergarden The Movie"
		DisplayTitle string `json:"displayTitle"`
		// e.g, "Shibuya Incident - Gate, Open"
		EpisodeTitle string `json:"episodeTitle"`

		EpisodeNumber int `json:"episodeNumber,omitempty"`
		// ProgressNumber is the source of truth for tracking purposes. It should exactly map to AniList.
		// Usually the same as EpisodeNumber, unless episode 0 is included by AniList.
		// e.g, Movie: 1, Show: 0,1,2...
		ProgressNumber int        `json:"progressNumber,omitempty"`
		LocalFile      *LocalFile `json:"localFile"`
		IsDownloaded   bool       `json:"isDownloaded"`
	}

	NewMediaEntryEpisodeOptions struct {
		localFile   *LocalFile
		anizipMedia *anizip.Media
		media       *anilist.BaseMedia
		// progressOffset will offset the ProgressNumber for a specific MAIN file
		// This is used when there is a discrepancy between AniList and AniDB
		progressOffset int
		isDownloaded   bool
	}
)

func NewMediaEntryEpisode(opts *NewMediaEntryEpisodeOptions) *MediaEntryEpisode {
	ep := new(MediaEntryEpisode)

	ep.DisplayTitle = ""
	ep.EpisodeTitle = ""

	// Format DisplayTitle
	if *opts.media.GetFormat() == anilist.MediaFormatMovie {
		ep.DisplayTitle = opts.media.GetPreferredTitle()
	}

	//anizipEpisode, found := opts.anizipMedia.GetEpisode(opts.localFile.Metadata.AniDBEpisode)

	// LocalFile exists
	if opts.localFile != nil {
		ep.IsDownloaded = true

		switch opts.localFile.Metadata.Type {
		case LocalFileTypeMain:
			ep.EpisodeNumber = opts.localFile.Metadata.Episode
			ep.ProgressNumber = opts.localFile.Metadata.Episode + opts.progressOffset
		case LocalFileTypeSpecial:
			ep.EpisodeNumber = opts.localFile.Metadata.Episode
			ep.ProgressNumber = 0
		}
	}

	return ep
}
