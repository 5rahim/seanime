package mappings

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetReducedAnimeLists(t *testing.T) {

	res, err := GetReducedAnimeLists()
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEmpty(t, res, "response should not be empty")

	t.Logf("Anime list count: %d", res.Count)

	tests := []struct {
		name           string
		anidbID        int
		expectedTvdbID int
	}{
		{
			name:           "Anime 1",
			anidbID:        18132,
			expectedTvdbID: 420280,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			tvdbID, ok := res.FindTvdbIDFromAnidbID(test.anidbID)
			if !ok {
				t.Fatalf("tvdbID not found")
			}

			assert.Equal(t, test.expectedTvdbID, tvdbID, "tvdbID should match expected value")

		})

	}

}
