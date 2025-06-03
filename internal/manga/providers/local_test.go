package manga_providers

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanChapterFilename(t *testing.T) {
	tests := []struct {
		filename             string
		expectedChapter      []string
		expectedMangaTitle   string
		expectedChapterTitle string
		expectedVolume       []string
	}{
		{
			filename:             "1.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "2.5.pdf",
			expectedChapter:      []string{"2.5"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Chapter 5.5.pdf",
			expectedChapter:      []string{"5.5"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "ch 1.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "ch 1.5-2.cbz",
			expectedChapter:      []string{"1.5", "2"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Some title Chapter 1.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Some title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Chapter 23 The Fanatics.pdf",
			expectedChapter:      []string{"23"},
			expectedMangaTitle:   "The Fanatics",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "chapter_1.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "1 - Some title.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "Some title",
			expectedVolume:       []string{},
		},
		{
			filename:             "30 - Some title.cbz",
			expectedChapter:      []string{"30"},
			expectedMangaTitle:   "",
			expectedChapterTitle: "Some title",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title - c001 [123456].cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title - c12.5 [654321].cbz",
			expectedChapter:      []string{"12.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title 05 - ch10.cbz",
			expectedChapter:      []string{"10"},
			expectedMangaTitle:   "Manga Title 05",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title - ch10.cbz",
			expectedChapter:      []string{"10"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title - ch_11.cbz",
			expectedChapter:      []string{"11"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "[Group] Manga Title - ch-12.cbz",
			expectedChapter:      []string{"12"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title v01 c001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{"01"},
		},
		{
			filename:             "Manga Title v01 c001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{"01"},
		},
		{
			filename:             "Manga Title - 003.cbz",
			expectedChapter:      []string{"003"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 003.5.cbz",
			expectedChapter:      []string{"003.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 3.5 (Digital).cbz",
			expectedChapter:      []string{"3.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 10 (Digital) [Group].cbz",
			expectedChapter:      []string{"10"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp_15.cbz",
			expectedChapter:      []string{"15"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp-16.cbz",
			expectedChapter:      []string{"16"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp17.cbz",
			expectedChapter:      []string{"17"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp 18.cbz",
			expectedChapter:      []string{"18"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001 (v2).cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001v2.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{"2"},
		},
		{
			filename:             "Manga Title - 001 [v2].cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001 [Digital] [v2].cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001-002.cbz",
			expectedChapter:      []string{"001", "002"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001-001.5.cbz",
			expectedChapter:      []string{"001", "001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1-2.cbz",
			expectedChapter:      []string{"1", "2"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1.5-2.cbz",
			expectedChapter:      []string{"1.5", "2"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1 (Sample).cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1 (Preview).cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1 (Special Edition).cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1 (Digital) (Official).cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1.cbz",
			expectedChapter:      []string{"1"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 1.0.cbz",
			expectedChapter:      []string{"1.0"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 01.cbz",
			expectedChapter:      []string{"01"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 01.5.cbz",
			expectedChapter:      []string{"01.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - 001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch_001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch_001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch-001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - ch-001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp_001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp_001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp-001.cbz",
			expectedChapter:      []string{"001"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
		{
			filename:             "Manga Title - chp-001.5.cbz",
			expectedChapter:      []string{"001.5"},
			expectedMangaTitle:   "Manga Title",
			expectedChapterTitle: "",
			expectedVolume:       []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			res, ok := scanChapterFilename(tt.filename)
			if !ok {
				t.Errorf("Failed to scan chapter filename: %s", tt.filename)
			}
			require.Equalf(t, tt.expectedChapter, res.Chapter, "Expected chapter '%v' for '%s' but got '%v'", tt.expectedChapter, tt.filename, res.Chapter)
			require.Equalf(t, tt.expectedMangaTitle, res.MangaTitle, "Expected manga title '%v' for '%s' but got '%v'", tt.expectedMangaTitle, tt.filename, res.MangaTitle)
			require.Equalf(t, tt.expectedChapterTitle, res.ChapterTitle, "Expected chapter title '%v' for '%s' but got '%v'", tt.expectedChapterTitle, tt.filename, res.ChapterTitle)
			require.Equalf(t, tt.expectedVolume, res.Volume, "Expected volume '%v' for '%s' but got '%v'", tt.expectedVolume, tt.filename, res.Volume)
		})
	}
}

func TestPageSorting(t *testing.T) {
	tests := []struct {
		expectedOrder []string
	}{
		{
			expectedOrder: []string{"1149-000.jpg", "1149-001.jpg", "1149-002.jpg", "1149-019.jpg", "1149-019b.jpg", "1149-020.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.expectedOrder), func(t *testing.T) {
			newSlice := tt.expectedOrder
			slices.SortFunc(newSlice, func(a, b string) int {
				return strings.Compare(a, b)
			})
			for i, filename := range tt.expectedOrder {
				require.Equalf(t, filename, newSlice[i], "Expected order '%v' for '%s' but got '%v'", tt.expectedOrder, tt.expectedOrder[i], filename)
			}
		})
	}
}

func TestParsePageFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected float64
	}{
		{
			filename: "1.jpg",
			expected: 1,
		},
		{
			filename: "1.5.jpg",
			expected: 1.5,
		},
		{
			filename: "Page 001.jpg",
			expected: 1,
		},
		{
			filename: "1.55.jpg",
			expected: 1.55,
		},
		{
			filename: "2.5 -.jpg",
			expected: 2.5,
		},
		{
			filename: "page_27.jpg",
			expected: 27,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			res, ok := parsePageFilename(tt.filename)
			if !ok {
				t.Errorf("Failed to parse page filename: %s", tt.filename)
			}
			require.Equalf(t, tt.expected, res.Number, "Expected number '%v' for '%s' but got '%v'", tt.expected, tt.filename, res.Number)
		})
	}
}
