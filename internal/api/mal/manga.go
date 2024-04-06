package mal

import (
	"fmt"
	"net/url"
)

const (
	BaseMangaFields string = "id,title,main_picture,alternative_titles,start_date,end_date,nsfw,synopsis,num_episodes,mean,rank,popularity,media_type,status"
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
		NumEpisodes int         `json:"num_episodes"`
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
	reqUrl := fmt.Sprintf("%s/manga/%d?fields=%s", ApiBaseURL, mId, BaseMangaFields)

	if w.AccessToken == "" {
		return nil, fmt.Errorf("access token is empty")
	}

	var manga BasicManga
	err := w.doQuery("GET", reqUrl, nil, "application/json", &manga)
	if err != nil {
		return nil, err
	}

	return &manga, nil
}

func (w *Wrapper) GetMangaCollection() ([]*MangaListEntry, error) {

	reqUrl := fmt.Sprintf("%s/users/@me/mangalist?fields=list_status&limit=1000", ApiBaseURL)

	type response struct {
		Data []*MangaListEntry `json:"data"`
	}

	var data response
	err := w.doQuery("GET", reqUrl, nil, "application/json", &data)
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

type MangaListProgressParams struct {
	NumChaptersRead *int
}

func (w *Wrapper) UpdateMangaProgress(opts *MangaListProgressParams, mId int) error {
	// Get manga details
	manga, err := w.GetMangaDetails(mId)
	if err != nil {
		return err
	}

	status := MediaListStatusWatching
	if manga.Status == MediaStatusFinishedAiring && manga.NumEpisodes == *opts.NumChaptersRead {
		status = MediaListStatusCompleted
	}

	// Update MAL list entry
	err = w.UpdateMangaListStatus(&MangaListStatusParams{
		Status:          &status,
		NumChaptersRead: opts.NumChaptersRead,
	}, mId)

	return err
}

type MangaListStatusParams struct {
	Status          *MediaListStatus
	IsRereading     *bool
	NumChaptersRead *int
	Score           *int
}

func (w *Wrapper) UpdateMangaListStatus(opts *MangaListStatusParams, mId int) error {

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
		return err
	}
	return nil
}

func (w *Wrapper) DeleteMangaListItem(mId int) error {

	reqUrl := fmt.Sprintf("%s/manga/%d/my_list_status", ApiBaseURL, mId)

	err := w.doMutation("DELETE", reqUrl, "")
	if err != nil {
		return err
	}

	return nil
}
