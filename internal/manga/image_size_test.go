package manga

import (
	"github.com/davecgh/go-spew/spew"
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"testing"
)

func TestGetImageNaturalSize(t *testing.T) {
	// Test the function
	width, height, err := getImageNaturalSize("https://meo.comick.pictures/2-G67V7CCdKhluM.png?width=3180")
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(width, height)
}
