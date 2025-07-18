package local

import "fmt"

// FormatAssetUrl formats the asset URL for the given mediaId and filename.
//
//	FormatAssetUrl(123, "cover.jpg") -> "{{LOCAL_ASSETS}}/123/cover.jpg"
func FormatAssetUrl(mediaId int, filename string) *string {
	// {{LOCAL_ASSETS}} should be replaced in the client with the actual URL
	// e.g. http://<hostname>/local_assets/123/cover.jpg
	a := fmt.Sprintf("{{LOCAL_ASSETS}}/%d/%s", mediaId, filename)
	return &a
}
