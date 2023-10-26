package scanner

import "testing"

func TestMatcher_FindBestCorrespondingMedia(t *testing.T) {

	lfs, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("could not get test local files")
	}
	media := MockAllMedia()

	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: *media,
	})

	matcher := NewMatcher(&MatcherOptions{
		localFiles:     lfs,
		mediaContainer: mc,
		baseMediaCache: nil,
	})

	matcher.FindBestCorrespondingMedia(lfs[0])

}
