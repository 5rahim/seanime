package videofile

import (
	"testing"
)

var dest = "E:\\ANIME_TEST\\blue_lock\\test.m3u8"

func TestTrans(t *testing.T) {

	//trans := new(transcoder.Transcoder)
	//
	//err := trans.Initialize(videopath, dest)
	//if err != nil {
	//	panic(err)
	//}
	//
	//trans.MediaFile().SetVideoCodec("libx264")
	//trans.MediaFile().SetHlsSegmentDuration(4)
	////trans.MediaFile().SetHlsPlaylistType("event")
	//trans.MediaFile().SetPixFmt("yuv420p")
	//trans.MediaFile().SetAudioCodec("aac")
	//
	//done := trans.Run(true)
	//progress := trans.Output()
	//for p := range progress {
	//	fmt.Println(p)
	//}
	//
	//fmt.Println(<-done)
}
