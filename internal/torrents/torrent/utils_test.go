package torrent

import (
	"io"
	"net/http"
	"testing"
)

func TestFileToMagnetLink(t *testing.T) {
	t.Skip()

	tests := []struct {
		name string
		url  string
	}{
		{
			name: "1",
			url:  "",
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
