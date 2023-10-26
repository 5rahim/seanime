package comparison

import (
	"github.com/samber/lo"
	"testing"
)

func TestFindBestMatchWithLevenstein(t *testing.T) {

	title := "jujutsu kaisen 2"
	titles := []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"}
	expected := "Jujutsu Kaisen 2"

	if res, ok := FindBestMatchWithLevenstein(&title, lo.ToSlicePtr(titles)); ok {
		if *res.Value != expected {
			t.Errorf("expected %s for %s, got %s", expected, title, res.Value)
		}
		if res.Distance != 0 {
			t.Errorf("expected a distance of 0, got %d", res.Distance)
		}
		t.Logf("value: %s, distance: %d", res.Value, res.Distance)
	} else {
		t.Error("expected result, got nil")
	}

}

func TestEliminateLestSimilarValue(t *testing.T) {

	titles := []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"}

	res := EliminateLestSimilarValue(titles)

	for _, n := range res {
		if n == "JJK" {
			t.Fatalf("expected \"%s\" to be eliminated from %v", n, res)
		}
	}

	titles = []string{"One Piece - Film Z", "One Piece Film Z", "One Piece Gold"}

	res = EliminateLestSimilarValue(titles)

	for _, n := range res {
		if n == "One Piece Gold" {
			t.Fatalf("expected \"%s\" to be eliminated from %v", n, res)
		}
	}

	titles = []string{"One Piece - Film Z", "One Piece Film Z", "One Piece Z"}

	res = EliminateLestSimilarValue(titles)

	for _, n := range res {
		if n == "One Piece Z" {
			t.Fatalf("expected \"%s\" to be eliminated from %v", n, res)
		}
	}

	titles = []string{"Mononogatari", "Mononogatari Cour 2", "Nekomonogatari"}

	res = EliminateLestSimilarValue(titles)

	for _, n := range res {
		if n == "Nekomonogatari" {
			t.Fatalf("expected \"%s\" to be eliminated from %v", n, res)
		}
	}

}
