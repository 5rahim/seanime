package mediastream

import (
	"fmt"
	"github.com/xfrr/goffmpeg/transcoder"
	"os"
	"path/filepath"
	"testing"
)

func TestTrans(t *testing.T) {
	t.Skip("Do not run")
	var dest = "E:\\TRANSCODING_TEMP\\id\\index.m3u8"
	var videopath = "E:\\ANIME\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 15.mkv"
	_ = os.MkdirAll(filepath.Dir(dest), 0755)

	trans := new(transcoder.Transcoder)

	err := trans.Initialize(videopath, dest)
	if err != nil {
		panic(err)
	}

	trans.MediaFile().SetHardwareAcceleration("auto")
	//trans.MediaFile().SetSeekTime("00:10:00")
	trans.MediaFile().SetPreset("veryfast")
	trans.MediaFile().SetVideoCodec("libx264")
	trans.MediaFile().SetHlsPlaylistType("vod")
	trans.MediaFile().SetCRF(32)
	trans.MediaFile().SetHlsMasterPlaylistName("index.m3u8")
	trans.MediaFile().SetHlsSegmentDuration(4)
	trans.MediaFile().SetHlsSegmentFilename("segment-%03d.ts")
	//trans.MediaFile().SetHlsListSize(0)
	trans.MediaFile().SetPixFmt("yuv420p")
	trans.MediaFile().SetAudioCodec("aac")
	trans.MediaFile().SetTags(map[string]string{"-map": "0:v:0 0:a:0"})

	done := trans.Run(true)
	progress := trans.Output()
	for p := range progress {
		fmt.Println(p)
	}

	fmt.Println(<-done)

}
