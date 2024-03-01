package streamer

import (
	"fmt"
	"github.com/seanime-app/seanime/internal/videofile"
	"math"
	"os"
	"sync"
)

type FileStream struct {
	ready     sync.WaitGroup
	err       error
	Path      string
	Out       string
	Keyframes *Keyframe
	Info      *videofile.MediaInfo
	videos    CMap[Quality, *VideoStream]
	audios    CMap[int32, *AudioStream]
}

func NewFileStream(path string, sha string, route string) *FileStream {
	ret := &FileStream{
		Path:   path,
		Out:    fmt.Sprintf("%s/%s", Settings.Outpath, sha),
		videos: NewCMap[Quality, *VideoStream](),
		audios: NewCMap[int32, *AudioStream](),
	}

	ret.ready.Add(1)
	go func() {
		defer ret.ready.Done()
		info, err := GetInfo(path, sha, route)
		ret.Info = info
		if err != nil {
			ret.err = err
		}
	}()

	ret.ready.Add(1)
	go func() {
		defer ret.ready.Done()
		ret.Keyframes = GetKeyframes(sha, path)
	}()

	return ret
}

func (fs *FileStream) Kill() {
	fs.videos.lock.Lock()
	defer fs.videos.lock.Unlock()
	fs.audios.lock.Lock()
	defer fs.audios.lock.Unlock()

	for _, s := range fs.videos.data {
		s.Kill()
	}
	for _, s := range fs.audios.data {
		s.Kill()
	}
}

func (fs *FileStream) Destroy() {
	fs.Kill()
	_ = os.RemoveAll(fs.Out)
}

func (fs *FileStream) GetMaster() string {
	master := "#EXTM3U\n"
	if fs.Info.Video != nil {
		var transmuxQuality Quality
		for _, quality := range Qualities {
			if quality.Height() >= fs.Info.Video.Quality.Height() || quality.AverageBitrate() >= fs.Info.Video.Bitrate {
				transmuxQuality = quality
				break
			}
		}
		// TODO: also check if the codec is valid in a hls before putting transmux
		if true {
			bitrate := float64(fs.Info.Video.Bitrate)
			master += "#EXT-X-STREAM-INF:"
			master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", int(math.Min(bitrate*0.8, float64(transmuxQuality.AverageBitrate()))))
			master += fmt.Sprintf("BANDWIDTH=%d,", int(math.Min(bitrate, float64(transmuxQuality.MaxBitrate()))))
			master += fmt.Sprintf("RESOLUTION=%dx%d,", fs.Info.Video.Width, fs.Info.Video.Height)
			master += "AUDIO=\"audio\","
			master += "CLOSED-CAPTIONS=NONE\n"
			master += fmt.Sprintf("./%s/index.m3u8\n", Original)
		}
		aspectRatio := float32(fs.Info.Video.Width) / float32(fs.Info.Video.Height)
		for _, quality := range Qualities {
			if quality.Height() < fs.Info.Video.Quality.Height() && quality.AverageBitrate() < fs.Info.Video.Bitrate {
				master += "#EXT-X-STREAM-INF:"
				master += fmt.Sprintf("AVERAGE-BANDWIDTH=%d,", quality.AverageBitrate())
				master += fmt.Sprintf("BANDWIDTH=%d,", quality.MaxBitrate())
				master += fmt.Sprintf("RESOLUTION=%dx%d,", int(aspectRatio*float32(quality.Height())+0.5), quality.Height())
				master += "CODECS=\"avc1.640028\","
				master += "AUDIO=\"audio\","
				master += "CLOSED-CAPTIONS=NONE\n"
				master += fmt.Sprintf("./%s/index.m3u8\n", quality)
			}
		}
	}
	for _, audio := range fs.Info.Audios {
		master += "#EXT-X-MEDIA:TYPE=AUDIO,"
		master += "GROUP-ID=\"audio\","
		if audio.Language != nil {
			master += fmt.Sprintf("LANGUAGE=\"%s\",", *audio.Language)
		}
		if audio.Title != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Title)
		} else if audio.Language != nil {
			master += fmt.Sprintf("NAME=\"%s\",", *audio.Language)
		} else {
			master += fmt.Sprintf("NAME=\"Audio %d\",", audio.Index)
		}
		if audio.IsDefault {
			master += "DEFAULT=YES,"
		}
		master += fmt.Sprintf("URI=\"./audio/%d/index.m3u8\"\n", audio.Index)
	}
	return master
}

func (fs *FileStream) getVideoStream(quality Quality) *VideoStream {
	stream, _ := fs.videos.GetOrCreate(quality, func() *VideoStream {
		return NewVideoStream(fs, quality)
	})
	return stream
}

func (fs *FileStream) GetVideoIndex(quality Quality) (string, error) {
	stream := fs.getVideoStream(quality)
	return stream.GetIndex()
}

func (fs *FileStream) GetVideoSegment(quality Quality, segment int32) (string, error) {
	stream := fs.getVideoStream(quality)
	return stream.GetSegment(segment)
}

func (fs *FileStream) getAudioStream(audio int32) *AudioStream {
	stream, _ := fs.audios.GetOrCreate(audio, func() *AudioStream {
		return NewAudioStream(fs, audio)
	})
	return stream
}

func (fs *FileStream) GetAudioIndex(audio int32) (string, error) {
	stream := fs.getAudioStream(audio)
	return stream.GetIndex()
}

func (fs *FileStream) GetAudioSegment(audio int32, segment int32) (string, error) {
	stream := fs.getAudioStream(audio)
	return stream.GetSegment(segment)
}
