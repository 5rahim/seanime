package scanner

import (
	"github.com/seanime-app/seanime-server/internal/util"
	"github.com/sourcegraph/conc/pool"
	"testing"
)

func TestMatcher_MatchLocalFileWithMedia(t *testing.T) {

	logger := util.NewLogger()

	lfs, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("could not get test local files")
	}
	media := MockAllMedia()

	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: *media,
	})

	matcher := &Matcher{
		localFiles:     lfs,
		mediaContainer: mc,
		baseMediaCache: nil,
		logger:         logger,
	}

	p := pool.New()
	for _, lf := range lfs {
		lf := lf
		p.Go(func() {
			matcher.MatchLocalFileWithMedia(lf)
		})
	}
	p.Wait()

}

func TestMatcher_MatchLocalFilesWithMedia(t *testing.T) {

	logger := util.NewLogger()

	lfs, ok := MockGetTestLocalFiles()
	if !ok {
		t.Fatal("could not get test local files")
	}
	media := MockAllMedia()

	mc := NewMediaContainer(&MediaContainerOptions{
		allMedia: *media,
	})

	matcher := &Matcher{
		localFiles:     lfs,
		mediaContainer: mc,
		baseMediaCache: nil,
		logger:         logger,
	}

	if err := matcher.MatchLocalFilesWithMedia(); err != nil {
		t.Fatal(err)
	}

	for _, lf := range lfs {
		t.Logf("local file: %s,\nmedia id: %d\n\n", lf.Name, lf.MediaId)
	}

}
