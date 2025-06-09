package manga_providers

import (
	"bytes"
	"image/jpeg"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConvertPDFToImages(t *testing.T) {
	start := time.Now()

	doc, err := fitz.New("")
	require.NoError(t, err)
	defer doc.Close()

	images := make(map[int][]byte, doc.NumPage())

	// Load images into memory
	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			panic(err)
		}

		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			panic(err)
		}

		images[n] = buf.Bytes()
	}

	end := time.Now()

	t.Logf("Converted %d pages in %f seconds", len(images), end.Sub(start).Seconds())

	for n, imgData := range images {
		t.Logf("Page %d: %d bytes", n, len(imgData))
	}

	//tmpDir, err := os.MkdirTemp(os.TempDir(), "manga_test_")
	//require.NoError(t, err)
	//if len(images) > 0 {
	//	// Write the first image to a file for verification
	//	firstImagePath := tmpDir + "/page_0.jpg"
	//	err = os.WriteFile(firstImagePath, images[0], 0644)
	//	require.NoError(t, err)
	//	t.Logf("First image written to: %s", firstImagePath)
	//}
	//
	//time.Sleep(1 * time.Minute)
	//
	//t.Cleanup(func() {
	//	// Clean up the temporary directory
	//	err := os.RemoveAll(tmpDir)
	//	if err != nil {
	//		t.Logf("Failed to remove temp directory: %v", err)
	//	} else {
	//		t.Logf("Temporary directory removed: %s", tmpDir)
	//	}
	//})
}
