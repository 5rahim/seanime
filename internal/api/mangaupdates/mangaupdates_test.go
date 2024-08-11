package mangaupdates

import (
	"bytes"
	"github.com/davecgh/go-spew/spew"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestApi(t *testing.T) {

	tests := []struct {
		title     string
		startDate string
	}{
		{
			title:     "Dandadan",
			startDate: "2021-04-06",
		},
	}

	type searchReleaseBody struct {
		Search    string `json:"search"`
		StartDate string `json:"start_date,omitempty"`
	}

	var apiUrl = "https://api.mangaupdates.com/v1/releases/search"

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {

			client := http.Client{Timeout: 10 * time.Second}

			body := searchReleaseBody{
				Search:    strings.ToLower(test.title),
				StartDate: test.startDate,
			}

			bodyB, err := json.Marshal(body)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(bodyB))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			var result interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			spew.Dump(result)

		})
	}

}
