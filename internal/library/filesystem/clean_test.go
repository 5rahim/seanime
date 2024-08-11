package filesystem

import (
	"seanime/internal/util"
	"testing"
)

func TestDeleteEmptyDirectories(t *testing.T) {

	path := "E:/ANIME_TEST"

	RemoveEmptyDirectories(path, util.NewLogger())

	t.Log("All empty directories removed successfully.")

}
