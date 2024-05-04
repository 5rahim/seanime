package transcoder

import (
	"github.com/rs/zerolog"
	"time"
)

type ClientInfo struct {
	client  string
	path    string
	quality *Quality
	audio   int32
	head    int32
}

type Tracker struct {
	// key: client_id
	clients map[string]ClientInfo
	// key: client_id
	visitDate map[string]time.Time
	// key: path
	lastUsage     map[string]time.Time
	transcoder    *Transcoder
	deletedStream chan string
	logger        *zerolog.Logger
}

func NewTracker(t *Transcoder) *Tracker {
	ret := &Tracker{
		clients:    make(map[string]ClientInfo),
		visitDate:  make(map[string]time.Time),
		lastUsage:  make(map[string]time.Time),
		transcoder: t,
		logger:     t.logger,
	}
	go ret.start()
	return ret
}

func Abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func (t *Tracker) start() {
	inactiveTime := 1 * time.Hour
	timer := time.After(inactiveTime)
	for {
		select {
		case info, ok := <-t.transcoder.clientChan:
			if !ok {
				return
			}

			old, ok := t.clients[info.client]
			// First fixup the info. Most routes ruturn partial infos
			if ok && old.path == info.path {
				if info.quality == nil {
					info.quality = old.quality
				}
				if info.audio == -1 {
					info.audio = old.audio
				}
				if info.head == -1 {
					info.head = old.head
				}
			}

			t.clients[info.client] = info
			t.visitDate[info.client] = time.Now()
			t.lastUsage[info.path] = time.Now()

			// now that the new info is stored and fixed, kill old streams
			if ok && old.path == info.path {
				if old.audio != info.audio && old.audio != -1 {
					t.KillAudioIfDead(old.path, old.audio)
				}
				if old.quality != info.quality && old.quality != nil {
					t.KillQualityIfDead(old.path, *old.quality)
				}
				if old.head != -1 && Abs(info.head-old.head) > 100 {
					t.KillOrphanedHeads(old.path, old.quality, old.audio)
				}
			} else if ok {
				t.KillStreamIfDead(old.path)
			}

		case <-timer:
			timer = time.After(inactiveTime)
			// Purge old clients
			for client, date := range t.visitDate {
				if time.Since(date) < inactiveTime {
					continue
				}

				info := t.clients[client]

				if !t.KillStreamIfDead(info.path) {
					audioCleanup := info.audio != -1 && t.KillAudioIfDead(info.path, info.audio)
					videoCleanup := info.quality != nil && t.KillQualityIfDead(info.path, *info.quality)
					if !audioCleanup || !videoCleanup {
						t.KillOrphanedHeads(info.path, info.quality, info.audio)
					}
				}

				delete(t.clients, client)
				delete(t.visitDate, client)
			}
		case path := <-t.deletedStream:
			t.DestroyStreamIfOld(path)
		}
	}
}

func (t *Tracker) KillStreamIfDead(path string) bool {
	for _, stream := range t.clients {
		if stream.path == path {
			return false
		}
	}
	t.logger.Trace().Msgf("Killing stream %s", path)

	stream, ok := t.transcoder.streams.Get(path)
	if !ok {
		return false
	}
	stream.Kill()
	go func() {
		time.Sleep(4 * time.Hour)
		t.deletedStream <- path
	}()
	return true
}

func (t *Tracker) DestroyStreamIfOld(path string) {
	if time.Since(t.lastUsage[path]) < 4*time.Hour {
		return
	}
	stream, ok := t.transcoder.streams.GetAndRemove(path)
	if !ok {
		return
	}
	stream.Destroy()
}

func (t *Tracker) KillAudioIfDead(path string, audio int32) bool {
	for _, stream := range t.clients {
		if stream.path == path && stream.audio == audio {
			return false
		}
	}
	t.logger.Trace().Msgf("Killing audio %d of %s", audio, path)

	stream, ok := t.transcoder.streams.Get(path)
	if !ok {
		return false
	}
	astream, aok := stream.audios.Get(audio)
	if !aok {
		return false
	}
	astream.Kill()
	return true
}

func (t *Tracker) KillQualityIfDead(path string, quality Quality) bool {
	for _, stream := range t.clients {
		if stream.path == path && stream.quality != nil && *stream.quality == quality {
			return false
		}
	}
	t.logger.Trace().Msgf("Killing quality %s of %s", quality, path)

	stream, ok := t.transcoder.streams.Get(path)
	if !ok {
		return false
	}
	vstream, vok := stream.videos.Get(quality)
	if !vok {
		return false
	}
	vstream.Kill()
	return true
}

func (t *Tracker) KillOrphanedHeads(path string, quality *Quality, audio int32) {
	stream, ok := t.transcoder.streams.Get(path)
	if !ok {
		return
	}

	if quality != nil {
		vstream, vok := stream.videos.Get(*quality)
		if vok {
			t.killOrphanedHeads(&vstream.Stream)
		}
	}
	if audio != -1 {
		astream, aok := stream.audios.Get(audio)
		if aok {
			t.killOrphanedHeads(&astream.Stream)
		}
	}
}

func (t *Tracker) killOrphanedHeads(stream *Stream) {
	stream.lock.Lock()
	defer stream.lock.Unlock()

	for encoder_id, head := range stream.heads {
		if head == DeletedHead {
			continue
		}

		distance := int32(99999)
		for _, info := range t.clients {
			if info.head == -1 {
				continue
			}
			distance = min(Abs(info.head-head.segment), distance)
		}
		if distance > 20 {
			t.logger.Trace().Msgf("Killing orphaned head %s %d", stream.file.Path, encoder_id)
			stream.KillHead(encoder_id)
		}
	}
}
