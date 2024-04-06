package mal

import (
	"fmt"
	"net/url"
)

const (
	BaseMangaFields string = "id,title,main_picture,alternative_titles,start_date,end_date,nsfw,synopsis,num_volumes,num_chapters,mean,rank,popularity,media_type,status"
)

type (
	BasicManga struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		MainPicture struct {
			Medium string `json:"medium"`
			Large  string `json:"large"`
		} `json:"main_picture"`
		AlternativeTitles struct {
			Synonyms []string `json:"synonyms"`
			En       string   `json:"en"`
			Ja       string   `json:"ja"`
		} `json:"alternative_titles"`
		StartDate   string      `json:"start_date"`
		EndDate     string      `json:"end_date"`
		Synopsis    string      `json:"synopsis"`
		NSFW        string      `json:"nsfw"`
		NumVolumes  int         `json:"num_volumes"`
		NumChapters int         `json:"num_chapters"`
		Mean        float32     `json:"mean"`
		Rank        int         `json:"rank"`
		Popularity  int         `json:"popularity"`
		MediaType   MediaType   `json:"media_type"`
		Status      MediaStatus `json:"status"`
	}

	MangaListEntry struct {
		Node struct {
			ID          int    `json:"id"`
			Title       string `json:"title"`
			MainPicture struct {
				Medium string `json:"medium"`
				Large  string `json:"large"`
			} `json:"main_picture"`
		} `json:"node"`
		ListStatus struct {
			Status          MediaListStatus `json:"status"`
			IsRereading     bool            `json:"is_rereading"`
			NumVolumesRead  int             `json:"num_volumes_read"`
			NumChaptersRead int             `json:"num_chapters_read"`
			Score           int             `json:"score"`
			UpdatedAt       string          `json:"updated_at"`
		} `json:"list_status"`
	}
)

func (w *Wrapper) GetMangaDetails(mId int) (*BasicManga, error) {
	w.logger.Debug().Int("mId", mId).Msg("mal: Getting manga details")

	reqUrl := fmt.Sprintf("%s/manga/%d?fields=%s", ApiBaseURL, mId, BaseMangaFields)

	if w.AccessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}

	var manga BasicManga
	err := w.doQuery("GET", reqUrl, nil, "application/json", &manga)
	if err != nil {
		w.logger.Error().Err(err).Msg("mal: Failed to get manga details")
		return nil, err
	}

	w.logger.Info().Int("mId", mId).Msg("mal: Fetched manga details")

	return &manga, nil
}

func (w *Wrapper) GetMangaCollection() ([]*MangaListEntry, error) {
	w.logger.Debug().Msg("mal: Getting manga collection")

	reqUrl := fmt.Sprintf("%s/users/@me/mangalist?fields=list_status&limit=1000", ApiBaseURL)

	type response struct {
		Data []*MangaListEntry `json:"data"`
	}

	var data response
	err := w.doQuery("GET", reqUrl, nil, "application/json", &data)
	if err != nil {
		w.logger.Error().Err(err).Msg("mal: Failed to get manga collection")
		return nil, err
	}

	w.logger.Info().Msg("mal: Fetched manga collection")

	return data.Data, nil
}

type MangaListProgressParams struct {
	NumChaptersRead *int
}

func (w *Wrapper) UpdateMangaProgress(opts *MangaListProgressParams, mId int) error {
	w.logger.Debug().Int("mId", mId).Msg("mal: Updating manga progress")

	// Get manga details
	manga, err := w.GetMangaDetails(mId)
	if err != nil {
		return err
	}

	status := MediaListStatusReading
	if manga.Status == MediaStatusFinished && manga.NumChapters > 0 && manga.NumChapters <= *opts.NumChaptersRead {
		status = MediaListStatusCompleted
	}

	if manga.NumChapters > 0 && *opts.NumChaptersRead > manga.NumChapters {
		*opts.NumChaptersRead = manga.NumChapters
	}

	// Update MAL list entry
	err = w.UpdateMangaListStatus(&MangaListStatusParams{
		Status:          &status,
		NumChaptersRead: opts.NumChaptersRead,
	}, mId)

	if err == nil {
		w.logger.Info().Int("mId", mId).Msg("mal: Updated manga progress")
	}

	return err
}

type MangaListStatusParams struct {
	Status          *MediaListStatus
	IsRereading     *bool
	NumChaptersRead *int
	Score           *int
}

func (w *Wrapper) UpdateMangaListStatus(opts *MangaListStatusParams, mId int) error {
	w.logger.Debug().Int("mId", mId).Msg("mal: Updating manga list status")

	reqUrl := fmt.Sprintf("%s/manga/%d/my_list_status", ApiBaseURL, mId)

	// Build URL
	urlData := url.Values{}
	if opts.Status != nil {
		urlData.Set("status", string(*opts.Status))
	}
	if opts.IsRereading != nil {
		urlData.Set("is_rereading", fmt.Sprintf("%t", *opts.IsRereading))
	}
	if opts.NumChaptersRead != nil {
		urlData.Set("num_chapters_read", fmt.Sprintf("%d", *opts.NumChaptersRead))
	}
	if opts.Score != nil {
		urlData.Set("score", fmt.Sprintf("%d", *opts.Score))
	}
	encodedData := urlData.Encode()

	err := w.doMutation("PATCH", reqUrl, encodedData)
	if err != nil {
		w.logger.Error().Err(err).Msg("mal: Failed to update manga list status")
		return err
	}
	return nil
}

func (w *Wrapper) DeleteMangaListItem(mId int) error {
	w.logger.Debug().Int("mId", mId).Msg("mal: Deleting manga list item")

	reqUrl := fmt.Sprintf("%s/manga/%d/my_list_status", ApiBaseURL, mId)

	err := w.doMutation("DELETE", reqUrl, "")
	if err != nil {
		w.logger.Error().Err(err).Msg("mal: Failed to delete manga list item")
		return err
	}

	w.logger.Info().Int("mId", mId).Msg("mal: Deleted manga list item")

	return nil
}
