package mediacore

import (
	"fmt"
	"seanime/internal/util/limiter"
	"seanime/internal/util/result"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog"
)

type InSight struct {
	logger                *zerolog.Logger
	rateLimiter           *limiter.Limiter
	characterDetailsCache *result.BoundedCache[int, *InSightCharacterDetails]
}

type InSightCharacter struct {
	MalID  int    `json:"mal_id"`
	URL    string `json:"url"`
	Images struct {
		Jpg struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
		Webp struct {
			ImageURL      string `json:"image_url"`
			SmallImageURL string `json:"small_image_url"`
		} `json:"webp"`
	} `json:"images"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Favorites int    `json:"favorites"`
}

type InSightCharacterDetails struct {
	MalID  int    `json:"mal_id"`
	URL    string `json:"url"`
	Images struct {
		Jpg struct {
			ImageURL string `json:"image_url"`
		} `json:"jpg"`
		Webp struct {
			ImageURL      string `json:"image_url"`
			SmallImageURL string `json:"small_image_url"`
		} `json:"webp"`
	} `json:"images"`
	Name      string   `json:"name"`
	NameKanji string   `json:"name_kanji"`
	Nicknames []string `json:"nicknames"`
	Favorites int      `json:"favorites"`
	About     string   `json:"about"`
}

type jikanCharactersResponse struct {
	Data []*struct {
		Character InSightCharacter `json:"character"`
		Role      string           `json:"role"`
		Favorites int              `json:"favorites"`
	} `json:"data"`
}

type jikanCharacterResponse struct {
	Data InSightCharacterDetails `json:"data"`
}

func NewInSight(logger *zerolog.Logger) *InSight {
	return &InSight{
		logger:                logger,
		rateLimiter:           limiter.NewLimiter(time.Second, 3),
		characterDetailsCache: result.NewBoundedCache[int, *InSightCharacterDetails](100),
	}
}

func (is *InSight) FetchCharacters(malID int) ([]*InSightCharacter, error) {
	if malID <= 0 {
		return nil, fmt.Errorf("invalid malID: %d", malID)
	}
	is.rateLimiter.Wait()
	resp, err := req.C().R().Get(fmt.Sprintf("https://api.jikan.moe/v4/anime/%d/characters", malID))
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("failed to fetch characters from Jikan: %s", resp.Status)
	}
	var response jikanCharactersResponse
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}
	characters := make([]*InSightCharacter, 0, len(response.Data))
	for _, value := range response.Data {
		character := value.Character
		character.Role = value.Role
		character.Favorites = value.Favorites
		characters = append(characters, &character)
	}
	return characters, nil
}

func (is *InSight) GetCharacterInfo(malID int) (*InSightCharacterDetails, error) {
	if cached, ok := is.characterDetailsCache.Get(malID); ok {
		return cached, nil
	}
	is.rateLimiter.Wait()
	resp, err := req.C().R().Get(fmt.Sprintf("https://api.jikan.moe/v4/characters/%d/full", malID))
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("failed to fetch character info: %s", resp.Status)
	}
	var response jikanCharacterResponse
	if err := resp.UnmarshalJson(&response); err != nil {
		return nil, err
	}
	is.characterDetailsCache.Set(malID, &response.Data)
	return &response.Data, nil
}

func (is *InSight) Clear() {
	is.characterDetailsCache.Clear()
}
