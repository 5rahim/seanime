package animeofflinedb

import (
	"testing"
)

func TestGetAnilistID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sources []string
		want    int
	}{
		{
			name: "Death Note",
			sources: []string{
				"https://anidb.net/anime/4563",
				"https://anilist.co/anime/1535",
				"https://anime-planet.com/anime/death-note",
				"https://myanimelist.net/anime/1535",
			},
			want: 1535,
		},
		{
			name: "No AniList source",
			sources: []string{
				"https://anidb.net/anime/4563",
				"https://myanimelist.net/anime/1535",
			},
			want: 0,
		},
		{
			name:    "Empty sources",
			sources: []string{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AnimeEntry{Sources: tt.sources}
			got := e.GetAnilistID()
			if got != tt.want {
				t.Errorf("GetAnilistID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMALID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		sources []string
		want    int
	}{
		{
			name: "Death Note",
			sources: []string{
				"https://anidb.net/anime/4563",
				"https://anilist.co/anime/1535",
				"https://myanimelist.net/anime/1535",
			},
			want: 1535,
		},
		{
			name: "No MAL source",
			sources: []string{
				"https://anidb.net/anime/4563",
				"https://anilist.co/anime/1535",
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AnimeEntry{Sources: tt.sources}
			got := e.GetMALID()
			if got != tt.want {
				t.Errorf("GetMALID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToNormalizedMedia(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entry      AnimeEntry
		wantNil    bool
		wantID     int
		wantTitle  string
		wantEps    int
		wantFormat string
	}{
		{
			name: "Valid entry",
			entry: AnimeEntry{
				Sources: []string{
					"https://anilist.co/anime/1535",
					"https://myanimelist.net/anime/1535",
				},
				Title:    "Death Note",
				Type:     "TV",
				Episodes: 37,
				Status:   "FINISHED",
				AnimeSeason: AnimeSeason{
					Season: "FALL",
					Year:   2006,
				},
				Synonyms: []string{"DN", "デスノート"},
			},
			wantNil:    false,
			wantID:     1535,
			wantTitle:  "Death Note",
			wantEps:    37,
			wantFormat: "TV",
		},
		{
			name: "No AniList ID",
			entry: AnimeEntry{
				Sources: []string{
					"https://myanimelist.net/anime/1535",
				},
				Title: "Death Note",
			},
			wantNil: true,
		},
		{
			name: "Movie format",
			entry: AnimeEntry{
				Sources: []string{
					"https://anilist.co/anime/199",
				},
				Title:    "Spirited Away",
				Type:     "MOVIE",
				Episodes: 1,
			},
			wantNil:    false,
			wantID:     199,
			wantTitle:  "Spirited Away",
			wantEps:    1,
			wantFormat: "MOVIE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.entry.ToNormalizedMedia()

			if tt.wantNil {
				if got != nil {
					t.Errorf("ToNormalizedMedia() expected nil, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("ToNormalizedMedia() returned nil, expected non-nil")
			}

			if got.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", got.ID, tt.wantID)
			}

			if got.GetTitleSafe() != tt.wantTitle {
				t.Errorf("Title = %v, want %v", got.GetTitleSafe(), tt.wantTitle)
			}

			if tt.wantEps > 0 {
				if got.Episodes == nil || *got.Episodes != tt.wantEps {
					t.Errorf("Episodes = %v, want %v", got.Episodes, tt.wantEps)
				}
			}
		})
	}
}

func TestConvertToNormalizedMedia(t *testing.T) {
	t.Parallel()

	db := &DatabaseRoot{
		Data: []AnimeEntry{
			{
				Sources:  []string{"https://anilist.co/anime/1"},
				Title:    "Test Anime 1",
				Type:     "TV",
				Episodes: 12,
			},
			{
				Sources:  []string{"https://anilist.co/anime/2"},
				Title:    "Test Anime 2",
				Type:     "MOVIE",
				Episodes: 1,
			},
			{
				Sources:  []string{"https://myanimelist.net/anime/3"}, // No AniList ID
				Title:    "Test Anime 3",
				Type:     "TV",
				Episodes: 24,
			},
		},
	}

	// Exclude anime with ID 1
	existing := map[int]bool{1: true}

	result := ConvertToNormalizedMedia(db, existing)

	// Should only include anime 2 (anime 1 is excluded, anime 3 has no AniList ID)
	if len(result) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result))
	}

	if len(result) > 0 && result[0].ID != 2 {
		t.Errorf("Expected anime ID 2, got %d", result[0].ID)
	}
}

// TestFetchDatabase tests fetching the actual database (integration)
func TestFetchDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Clear any cached data first
	ClearCache()

	db, err := FetchDatabase()
	if err != nil {
		t.Fatalf("FetchDatabase() error = %v", err)
	}

	if db == nil {
		t.Fatal("FetchDatabase() returned nil")
	}

	if len(db.Data) == 0 {
		t.Error("FetchDatabase() returned empty data")
	}

	// Check that we can find a known anime (Death Note)
	var foundDeathNote bool
	for _, entry := range db.Data {
		if entry.GetAnilistID() == 1535 {
			foundDeathNote = true
			if entry.Title != "Death Note" {
				t.Errorf("Expected 'Death Note', got '%s'", entry.Title)
			}
			break
		}
	}

	if !foundDeathNote {
		t.Error("Could not find Death Note (AniList ID 1535) in database")
	}

	t.Logf("Fetched %d anime entries from database", len(db.Data))

	ClearCache()
}
