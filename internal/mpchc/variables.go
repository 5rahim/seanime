package mpchc

import (
	"github.com/PuerkitoBio/goquery"
	"strconv"
	"strings"
)

type Variables struct {
	Version        string  `json:"version"`
	File           string  `json:"file"`
	FilePath       string  `json:"filepath"`
	FileDir        string  `json:"filedir"`
	Size           string  `json:"size"`
	State          int     `json:"state"`
	StateString    string  `json:"statestring"`
	Position       float64 `json:"position"`
	PositionString string  `json:"positionstring"`
	Duration       float64 `json:"duration"`
	DurationString string  `json:"durationstring"`
	VolumeLevel    float64 `json:"volumelevel"`
	Muted          bool    `json:"muted"`
}

func parseVariables(variablePageHtml string) *Variables {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(variablePageHtml))
	if err != nil {
		// Handle error
		return &Variables{}
	}

	fields := make(map[string]string)

	doc.Find("p").Each(func(_ int, s *goquery.Selection) {
		id, exists := s.Attr("id")
		if !exists {
			return
		}
		text := s.Text()
		fields[id] = text
	})

	return &Variables{
		Version:        fields["version"],
		File:           fields["file"],
		FilePath:       fields["filepath"],
		FileDir:        fields["filedir"],
		Size:           fields["size"],
		State:          parseInt(fields["state"]),
		StateString:    fields["statestring"],
		Position:       parseFloat(fields["position"]),
		PositionString: fields["positionstring"],
		Duration:       parseFloat(fields["duration"]),
		DurationString: fields["durationstring"],
		VolumeLevel:    parseFloat(fields["volumelevel"]),
		Muted:          fields["muted"] != "0",
	}
}

func parseInt(value string) int {
	intValue, _ := strconv.Atoi(value)
	return intValue
}

func parseFloat(value string) float64 {
	floatValue, _ := strconv.ParseFloat(value, 64)
	return floatValue
}
