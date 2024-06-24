package videofile

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

func TestFfprobeGetInfo_1(t *testing.T) {
	t.Skip()

	testFilePath := ""
	hash, err := GetHashFromPath(testFilePath)
	if err != nil {
		t.Fatalf("Error getting hash from path: %v", err)
	}

	mi, err := FfprobeGetInfo("", testFilePath, hash)
	if err != nil {
		t.Fatalf("Error getting media info: %v", err)
	}

	spew.Dump(mi)

}
