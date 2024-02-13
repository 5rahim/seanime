package mpv_player

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/gen2brain/go-mpv"
	"net/url"
)

var (
	ExitEndOfFile = errors.New("end of file")
	ExitShutdown  = errors.New("shutdown")
)

type (
	MpvPlayer struct {
		Playback *Playback
		ExitCh   chan error
		m        *mpv.Mpv
	}

	Playback struct {
		Filename string
		Paused   bool
		Position float64
		Duration float64
	}
)

func New() *MpvPlayer {
	return &MpvPlayer{
		m:        nil,
		Playback: &Playback{},
		ExitCh:   make(chan error),
	}
}

// OpenAndPlay starts a new video playback
func (p *MpvPlayer) OpenAndPlay(filePath string) error {
	if p.m == nil {
		return errors.New("mpv instance not initialized")
	}

	_ = p.m.RequestLogMessages("info")
	_ = p.m.ObserveProperty(0, "pause", mpv.FormatFlag)
	_ = p.m.ObserveProperty(0, "time-pos", mpv.FormatDouble)
	_ = p.m.ObserveProperty(0, "duration", mpv.FormatDouble)

	_ = p.m.SetPropertyString("input-default-bindings", "yes")
	_ = p.m.SetOptionString("input-vo-keyboard", "yes")
	_ = p.m.SetOption("osc", mpv.FormatFlag, true)

	// Initialize the instance
	err := p.m.Initialize()
	if err != nil {
		close(p.ExitCh)
		return errors.New("failed to initialize mpv")
	}

	filePath, _ = url.PathUnescape(filePath)
	filePath, _ = url.QueryUnescape(filePath)
	spew.Dump(filePath)

	err = p.m.Command([]string{"loadfile", filePath})
	if err != nil {
		close(p.ExitCh)
		return errors.New("failed to load file")
	}

	// Start the playback loop in a new goroutine
	// This will listen for events from the mpv instance
	go func() {
	loop:
		for {

			e := p.m.WaitEvent(2000)

			switch e.EventID {
			// Property
			case mpv.EventPropertyChange:
				prop := e.Property()
				value := prop.Data
				if value != nil {
					switch prop.Name {
					case "pause":
						p.Playback.Paused = map[int]bool{1: true, 0: false}[value.(int)]
					case "time-pos":
						p.Playback.Position = value.(float64)
					case "duration":
						p.Playback.Duration = value.(float64)
					}
				}
			case mpv.EventFileLoaded:
				// File loaded
				p.Playback.Reset()
				pr, err := p.m.GetProperty("media-title", mpv.FormatString)
				if err != nil {
					spew.Dump(err)
					p.ExitCh <- err
					break loop
				}
				if pr != nil {
					p.Playback.Filename = pr.(string)
				}
			//case mpv.EventLogMsg:
			//	msg := e.LogMessage()
			//	fmt.Println("message:", msg.Text)
			//case mpv.EventStart:
			//sf := e.StartFile()
			//fmt.Println("start:", sf.EntryID)
			case mpv.EventEnd:
				ef := e.EndFile()
				//fmt.Println("end:", ef.EntryID, ef.Reason)
				if ef.Reason == mpv.EndFileEOF {
					p.ExitCh <- ExitEndOfFile
					break loop
				} else if ef.Reason == mpv.EndFileError {
				}
				p.Playback.Reset()
			case mpv.EventShutdown:
				p.ExitCh <- ExitShutdown
				p.Playback.Reset()
				break loop
			}

		}
	}()

	// Start the exit listener
	// This will terminate the mpv instance when the exit channel is closed
	go func() {
		_, open := <-p.ExitCh
		if open {
			close(p.ExitCh)
		}
		spew.Dump("Terminating mpv")
		p.m.TerminateDestroy()
	}()

	return nil
}

func (p *MpvPlayer) GetPlaybackStatus() (*Playback, error) {
	if p.Playback == nil {
		return nil, errors.New("no playback status")
	}
	if p.Playback.Filename == "" {
		return nil, errors.New("no media found")
	}
	return p.Playback, nil
}

func (p *MpvPlayer) Start() {
	p.m = mpv.New()
	p.ExitCh = make(chan error)
}

func (p *MpvPlayer) Close() {
	p.m.TerminateDestroy()
	p.ExitCh <- ExitShutdown
}

func (pb *Playback) Reset() {
	pb.Filename = ""
	pb.Paused = false
	pb.Position = 0
	pb.Duration = 0
}
