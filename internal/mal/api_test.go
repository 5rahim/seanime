package mal

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestGetAnimeDetails(t *testing.T) {

	info := MockJWTs()

	res, err := GetAnimeDetails(info.MALJwt, 51179)

	spew.Dump(res)

	if err != nil {
		t.Fatalf("error while fetching media, %v", err)
	}

	t.Log(res.Title)
}

func TestUpdateAnimeListStatus(t *testing.T) {

	info := MockJWTs()

	mId := 51179
	progress := 2
	status := MediaListStatusWatching

	err := UpdateAnimeListStatus(info.MALJwt, &AnimeListStatusParams{
		Status:             &status,
		NumWatchedEpisodes: &progress,
	}, mId)

	if err != nil {
		t.Fatalf("error while fetching media, %v", err)
	}

}
