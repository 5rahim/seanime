package mediastream

import (
	"fmt"
	"github.com/xfrr/goffmpeg/transcoder"
	"os"
	"path/filepath"
	"testing"
)

var dest = "E:\\COLLECTION\\mediastream\\Dungeon Meshi\\04\\master.m3u8"
var videopath = "E:\\COLLECTION\\Dungeon Meshi\\[EMBER] Dungeon Meshi - 04.mkv"

func TestTrans(t *testing.T) {

	trans := new(transcoder.Transcoder)

	err := trans.Initialize(videopath, dest)
	if err != nil {
		panic(err)
	}

	destDir := filepath.Dir(dest)
	os.MkdirAll(destDir, 0755)

	trans.MediaFile().SetHardwareAcceleration("auto")
	//trans.MediaFile().SetSeekTime("00:10:00")
	trans.MediaFile().SetPreset("fast")
	trans.MediaFile().SetVideoCodec("libx264")
	trans.MediaFile().SetHlsPlaylistType("vod")
	trans.MediaFile().SetCRF(32)
	trans.MediaFile().SetHlsMasterPlaylistName("index.m3u8")
	trans.MediaFile().SetHlsSegmentDuration(6)
	trans.MediaFile().SetHlsListSize(0)
	trans.MediaFile().SetPixFmt("yuv420p")
	trans.MediaFile().SetAudioCodec("aac")
	trans.MediaFile().SetAudioChannels(2)

	done := trans.Run(true)
	progress := trans.Output()
	for p := range progress {
		fmt.Println(p)
	}

	fmt.Println(<-done)

}
