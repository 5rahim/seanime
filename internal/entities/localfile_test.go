package entities

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestNewLocalFile(t *testing.T) {

	tests := []struct {
		path string
	}{
		{
			path: "E:\\Anime\\Bungou Stray Dogs 5th Season\\[SubsPlease] Bungou Stray Dogs - 61 (1080p) [F609B947].mkv",
		},
	}

	for _, tt := range tests {

		lf := NewLocalFile(tt.path, "E:\\Anime")

		if lf == nil {
			t.Errorf("NewLocalFile(%v) returned nil", tt.path)
		}

		t.Logf("%s\n%+v\n%+v\n",
			tt.path,
			spew.Sdump(lf.ParsedData),
			spew.Sdump(lf.ParsedFolderData),
		)

	}

}
