package summary

import (
	"github.com/seanime-app/seanime/internal/anilist"
	"github.com/seanime-app/seanime/internal/entities"
)

type (
	ScanSummary struct {
		Groups []*ScanSummaryGroup `json:"groups"`
		Files  []*ScanSummaryFile  `json:"files"`
	}

	ScanSummaryFile struct {
		FilePath string   `json:"filePath"`
		Errors   []string `json:"errors"`
		Warnings []string `json:"warnings"`
		Logs     []string `json:"logs"`
	}

	ScanSummaryGroup struct {
		LocalFiles          []*entities.LocalFile `json:"localFiles"`
		MediaId             int                   `json:"mediaId"`
		Media               *anilist.BasicMedia   `json:"media,omitempty"`
		MediaIsInCollection bool                  `json:"mediaIsInCollection"` // Whether the media is in the user's AniList collection
	}
)
