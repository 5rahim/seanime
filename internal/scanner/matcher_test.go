package scanner

import (
	"github.com/sourcegraph/conc/pool"
	"testing"
)

func TestMatcher_MatchLocalFileWithMedia(t *testing.T) {

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

	p := pool.New()
	for _, lf := range lfs {
		lf := lf
		p.Go(func() {
			matcher.MatchLocalFileWithMedia(lf)
		})
	}
	p.Wait()

}
