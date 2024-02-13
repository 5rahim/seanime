package mpv_player

import (
	"fmt"
	"github.com/gen2brain/go-mpv"
)

type (
	MpvPlayer struct {
		Playback Playback
		exit     chan struct{}
	}

	Playback struct {
		Filename string
		Paused   bool
		Position float64
		Duration int
	}
)

func NewMpvPlayer() *MpvPlayer {
	return &MpvPlayer{
		exit: make(chan struct{}, 1),
	}
}

// OpenAndPlay starts a new video playback
func (p *MpvPlayer) OpenAndPlay(filePath string) error {

	// Create new MPV instance
	m := mpv.New()
	defer m.TerminateDestroy()

	_ = m.RequestLogMessages("info")
	_ = m.ObserveProperty(0, "pause", mpv.FormatFlag)
	_ = m.ObserveProperty(0, "time-pos", mpv.FormatDouble)
	_ = m.ObserveProperty(0, "duration", mpv.FormatDouble)

	_ = m.SetPropertyString("input-default-bindings", "yes")
	_ = m.SetOptionString("input-vo-keyboard", "yes")
	_ = m.SetOption("osc", mpv.FormatFlag, true)

	// Initialize the instance
	err := m.Initialize()
	if err != nil {
		return err
	}

	err = m.Command([]string{"loadfile", filePath})
	if err != nil {
		return err
	}

loop:
	for {

		e := m.WaitEvent(2000)

		switch e.EventID {
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
					p.Playback.Duration = int(value.(float64) * 1000)
				}
			}
		case mpv.EventFileLoaded:
			pr, err := m.GetProperty("media-title", mpv.FormatString)
			if err != nil {
				fmt.Println("error:", err)
			}
			p.Playback = Playback{
				Filename: pr.(string),
			}
		//case mpv.EventLogMsg:
		//	msg := e.LogMessage()
		//	fmt.Println("message:", msg.Text)
		case mpv.EventStart:
			sf := e.StartFile()
			fmt.Println("start:", sf.EntryID)
		case mpv.EventEnd:
			ef := e.EndFile()
			fmt.Println("end:", ef.EntryID, ef.Reason)
			if ef.Reason == mpv.EndFileEOF {
				p.exit <- struct{}{}
				break loop
			} else if ef.Reason == mpv.EndFileError {
				fmt.Println("error:", ef.Error)
			}
		case mpv.EventShutdown:
			fmt.Println("shutdown:", e.EventID)
			p.exit <- struct{}{}
			break loop
		default:
			//fmt.Println("event:", e.EventID)
		}

		if e.Error != nil {
			fmt.Println("error:", e.Error)
		}

	}

	return nil
}
