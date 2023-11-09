package entities

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/samber/lo"
	"strings"
	"testing"
)

func TestLocalFile_GetTitleVariations(t *testing.T) {

	lfs, ok := MockGetLocalFiles()
	if !ok {
		t.Fatal("failed to get mock local files")
	}

	lf, found := lo.Find(lfs, func(lf *LocalFile) bool {
		return strings.Contains(lf.Name, "Gaiden")
	})
	if !found {
		t.Fatal("failed to find local file")
	}
	lf.Metadata = &LocalFileMetadata{}

	t.Log(spew.Sdump(lf))

	t.Log(spew.Sprint(lf.GetTitleVariations()))

}
