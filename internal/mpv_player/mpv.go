package mpv_player

import (
	"fmt"
	"github.com/gen2brain/go-mpv"
)

type (
	MpvPlayer struct {
		paused chan struct{}
	}
)

func NewMpvPlayer() *MpvPlayer {
	return &MpvPlayer{
		paused: make(chan struct{}, 1),
	}
}

// OpenAndPlay starts a new video playback
func (p *MpvPlayer) OpenAndPlay(filePath string) error {

	// Create new MPV instance
	m := mpv.New()
	defer m.TerminateDestroy()

	_ = m.RequestLogMessages("info")
	_ = m.ObserveProperty(0, "pause", mpv.FormatFlag)

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
			value := prop.Data.(int)
			fmt.Println("property:", prop.Name, value)
		case mpv.EventFileLoaded:
			p, err := m.GetProperty("media-title", mpv.FormatString)
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Println("title:", p.(string))
		case mpv.EventLogMsg:
			msg := e.LogMessage()
			fmt.Println("message:", msg.Text)
		case mpv.EventStart:
			sf := e.StartFile()
			fmt.Println("start:", sf.EntryID)
		case mpv.EventEnd:
			ef := e.EndFile()
			fmt.Println("end:", ef.EntryID, ef.Reason)
			if ef.Reason == mpv.EndFileEOF {
				break loop
			} else if ef.Reason == mpv.EndFileError {
				fmt.Println("error:", ef.Error)
			}
		case mpv.EventShutdown:
			fmt.Println("shutdown:", e.EventID)
			break loop
		default:
			fmt.Println("event:", e.EventID)
		}

		if e.Error != nil {
			fmt.Println("error:", e.Error)
		}

	}

	return nil
}
