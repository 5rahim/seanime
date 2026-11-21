package handlers

import "testing"

func TestSuwayomiSourceName(t *testing.T) {
	for _, tt := range []struct{ lang, want string }{
		{"en", "Suwayomi: Example (EN)"},
		{"ja", "Suwayomi: Example (JP)"},
		{"es", "Suwayomi: Example (ES)"},
		{"all", "Suwayomi: Example (ALL)"},
		{"", "Suwayomi: Example"},
	} {
		if got := suwayomiSourceName("Example", tt.lang); got != tt.want {
			t.Errorf("language %q: got %q, want %q", tt.lang, got, tt.want)
		}
	}
}

func TestValidateSuwayomiImageURL(t *testing.T) {
	base := "http://127.0.0.1:4567/prefix"
	for _, image := range []string{
		base + "/api/v1/manga/7/thumbnail",
		base + "/api/v1/manga/7/chapter/2/page/0",
	} {
		if _, err := validateSuwayomiImageURL(base, image); err != nil {
			t.Fatalf("valid image rejected: %v", err)
		}
	}
	for _, image := range []string{
		"http://localhost:4567/prefix/api/v1/manga/7/thumbnail",
		base + "/api/v1/manga/7/chapter/2/page/0?token=secret",
		base + "/api/v1/manga/7/../settings",
		base + "/api/v1/manga/7/chapter/2/page/0/extra",
		"http://user@127.0.0.1:4567/prefix/api/v1/manga/7/thumbnail",
	} {
		if _, err := validateSuwayomiImageURL(base, image); err == nil {
			t.Fatalf("accepted invalid image: %s", image)
		}
	}
}
