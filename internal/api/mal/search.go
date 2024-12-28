package mal

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"seanime/internal/util/comparison"
	"seanime/internal/util/result"
	"sort"
	"strings"
)

type (
	SearchResultPayload struct {
		MediaType string `json:"media_type"`
		StartYear int    `json:"start_year"`
		Aired     string `json:"aired,omitempty"`
		Score     string `json:"score"`
		Status    string `json:"status"`
	}

	SearchResultAnime struct {
		ID           int                  `json:"id"`
		Type         string               `json:"type"`
		Name         string               `json:"name"`
		URL          string               `json:"url"`
		ImageURL     string               `json:"image_url"`
		ThumbnailURL string               `json:"thumbnail_url"`
		Payload      *SearchResultPayload `json:"payload"`
		ESScore      float64              `json:"es_score"`
	}

	SearchResult struct {
		Categories []*struct {
			Type  string               `json:"type"`
			Items []*SearchResultAnime `json:"items"`
		} `json:"categories"`
	}

	SearchCache struct {
		*result.Cache[int, *SearchResultAnime]
	}
)

//----------------------------------------------------------------------------------------------------------------------

// SearchWithMAL uses MAL's search API to find suggestions that match the title provided.
func SearchWithMAL(title string, slice int) ([]*SearchResultAnime, error) {

	url := "https://myanimelist.net/search/prefix.json?type=anime&v=1&keyword=" + url.QueryEscape(title)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var bodyMap SearchResult
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling error: %v", err)
	}

	if bodyMap.Categories == nil {
		return nil, fmt.Errorf("missing 'categories' in response")
	}

	items := make([]*SearchResultAnime, 0)
	for _, cat := range bodyMap.Categories {
		if cat.Type == "anime" {
			items = append(items, cat.Items...)
		}
	}

	if len(items) > slice {
		return items[:slice], nil
	}
	return items, nil
}

// AdvancedSearchWithMAL is like SearchWithMAL, but it uses additional algorithms to find the best match.
func AdvancedSearchWithMAL(title string) (*SearchResultAnime, error) {

	if len(title) == 0 {
		return nil, fmt.Errorf("title is empty")
	}

	// trim the title
	title = strings.ToLower(strings.TrimSpace(title))

	// MAL typically doesn't use "cour"
	re := regexp.MustCompile(`\bcour\b`)
	title = re.ReplaceAllString(title, "part")

	// fetch suggestions from MAL
	suggestions, err := SearchWithMAL(title, 8)
	if err != nil {
		return nil, err
	}

	// sort the suggestions by score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].ESScore > suggestions[j].ESScore
	})

	// keep anime that have aired
	suggestions = lo.Filter(suggestions, func(n *SearchResultAnime, index int) bool {
		return n.ESScore >= 0.1 && n.Payload.Status != "Not yet aired"
	})
	// reduce score if anime is older than 2006
	suggestions = lo.Map(suggestions, func(n *SearchResultAnime, index int) *SearchResultAnime {
		if n.Payload.StartYear < 2006 {
			n.ESScore -= 0.1
		}
		return n
	})

	tparts := strings.Fields(title)
	tsub := tparts[0]
	if len(tparts) > 1 {
		tsub += " " + tparts[1]
	}
	tsub = strings.TrimSpace(tsub)

	//
	t1, foundT1 := lo.Find(suggestions, func(n *SearchResultAnime) bool {
		nTitle := strings.ToLower(n.Name)

		_tsub := tparts[0]
		if len(tparts) > 1 {
			_tsub += " " + tparts[1]
		}
		_tsub = strings.TrimSpace(_tsub)

		re := regexp.MustCompile(`\b(film|movie|season|part|(s\d{2}e?))\b`)

		return strings.HasPrefix(nTitle, tsub) && n.Payload.MediaType == "TV" && !re.MatchString(nTitle)
	})

	// very generous
	t2, foundT2 := lo.Find(suggestions, func(n *SearchResultAnime) bool {
		nTitle := strings.ToLower(n.Name)

		_tsub := tparts[0]

		re := regexp.MustCompile(`\b(film|movie|season|part|(s\d{2}e?))\b`)

		return strings.HasPrefix(nTitle, _tsub) && n.Payload.MediaType == "TV" && !re.MatchString(nTitle)
	})

	levResult, found := comparison.FindBestMatchWithLevenshtein(&title, lo.Map(suggestions, func(n *SearchResultAnime, index int) *string { return &n.Name }))

	if !found {
		return nil, errors.New("couldn't find a suggestion from levenshtein")
	}

	levSuggestion, found := lo.Find(suggestions, func(n *SearchResultAnime) bool {
		return strings.ToLower(n.Name) == strings.ToLower(*levResult.Value)
	})

	if !found {
		return nil, errors.New("couldn't locate lenshtein result")
	}

	if foundT1 {
		d, found := comparison.FindBestMatchWithLevenshtein(&tsub, []*string{&title, new(string)})
		if found && len(*d.Value) > 0 {
			if d.Distance <= 1 {
				return t1, nil
			}
		}
	}

	// Strong correlation using MAL
	if suggestions[0].ESScore >= 4.5 {
		return suggestions[0], nil
	}

	// Very Likely match using distance
	if levResult.Distance <= 4 {
		return levSuggestion, nil
	}

	if suggestions[0].ESScore < 5 {

		// Likely match using [startsWith]
		if foundT1 {
			dev := math.Abs(t1.ESScore-suggestions[0].ESScore) < 2.0
			if len(tsub) > 6 && dev {
				return t1, nil
			}
		}
		// Likely match using [startsWith]
		if foundT2 {
			dev := math.Abs(t2.ESScore-suggestions[0].ESScore) < 2.0
			if len(tparts[0]) > 6 && dev {
				return t2, nil
			}
		}

		// Likely match using distance
		if levSuggestion.ESScore >= 1 && !(suggestions[0].ESScore > 3) {
			return suggestions[0], nil
		}

		// Less than likely match using MAL
		return suggestions[0], nil

	}

	// Distance above threshold, falling back to first MAL suggestion above
	if levResult.Distance >= 5 && suggestions[0].ESScore >= 1 {
		return suggestions[0], nil
	}

	return nil, nil
}
