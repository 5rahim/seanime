package videofile

import (
	"testing"
)

func TestGetProfileAndLevel(t *testing.T) {

	profile, level, _, err := getProfileAndLevel("E:/COLLECTION/One Piece/[Erai-raws] One Piece - 1072 [1080p][Multiple Subtitle][51CB925F].mkv")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Result: Level: %s  Profile: %s", level, profile)

}
