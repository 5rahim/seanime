package torrent

import (
	"io"
	"net/http"
	"testing"
)

func TestFileToMagnetLink(t *testing.T) {

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "1",
			url:  "https://animetosho.org/storage/torrent/da9aad67b6f8bb82757bb3ef95235b42624c34f7/%5BSubsPlease%5D%20Make%20Heroine%20ga%20Oosugiru%21%20-%2011%20%281080p%29%20%5B58B3496A%5D.torrent",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := http.Client{}
			resp, err := client.Get(test.url)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Error: %v", err)
			}

			magnet, err := StrDataToMagnetLink(string(data))
			if err != nil {
				t.Fatalf("Error: %v", err)
			}
			t.Log(magnet)
		})
	}

}
