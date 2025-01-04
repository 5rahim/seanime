package comparison

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindBestMatchWithLevenstein(t *testing.T) {

	tests := []struct {
		title            string
		comparisonTitles []string
		expectedResult   string
		expectedDistance int
	}{
		{
			title:            "jujutsu kaisen 2",
			comparisonTitles: []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"},
			expectedResult:   "Jujutsu Kaisen 2",
			expectedDistance: 0,
		},
	}

	for _, test := range tests {

		t.Run(test.title, func(t *testing.T) {
			res, ok := FindBestMatchWithLevenshtein(&test.title, lo.ToSlicePtr(test.comparisonTitles))

			if assert.True(t, ok) {
				assert.Equal(t, test.expectedResult, *res.Value, "expected result does not match")
				assert.Equal(t, test.expectedDistance, res.Distance, "expected distance does not match")
				t.Logf("value: %s, distance: %d", *res.Value, res.Distance)
			}

		})

	}

}
func TestFindBestMatchWithDice(t *testing.T) {

	tests := []struct {
		title            string
		comparisonTitles []string
		expectedResult   string
		expectedRating   float64
	}{
		{
			title:            "jujutsu kaisen 2",
			comparisonTitles: []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"},
			expectedResult:   "Jujutsu Kaisen 2",
			expectedRating:   1,
		},
	}

	for _, test := range tests {

		t.Run(test.title, func(t *testing.T) {
			res, ok := FindBestMatchWithSorensenDice(&test.title, lo.ToSlicePtr(test.comparisonTitles))

			if assert.True(t, ok, "expected result, got nil") {
				assert.Equal(t, test.expectedResult, *res.Value, "expected result does not match")
				assert.Equal(t, test.expectedRating, res.Rating, "expected rating does not match")
				t.Logf("value: %s, rating: %f", *res.Value, res.Rating)
			}

		})

	}

}

func TestEliminateLestSimilarValue(t *testing.T) {

	tests := []struct {
		title              string
		comparisonTitles   []string
		expectedEliminated string
	}{
		{
			title:              "jujutsu kaisen 2",
			comparisonTitles:   []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"},
			expectedEliminated: "JJK",
		},
		{
			title:              "One Piece - Film Z",
			comparisonTitles:   []string{"One Piece - Film Z", "One Piece Film Z", "One Piece Gold"},
			expectedEliminated: "One Piece Gold",
		},
		{
			title:              "One Piece - Film Z",
			comparisonTitles:   []string{"One Piece - Film Z", "One Piece Film Z", "One Piece Z"},
			expectedEliminated: "One Piece Z",
		},
		{
			title:              "Mononogatari",
			comparisonTitles:   []string{"Mononogatari", "Mononogatari Cour 2", "Nekomonogatari"},
			expectedEliminated: "Nekomonogatari",
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			res := EliminateLeastSimilarValue(test.comparisonTitles)
			for _, n := range res {
				if n == test.expectedEliminated {
					t.Fatalf("expected \"%s\" to be eliminated from %v", n, res)
				}
			}
		})
	}

}
