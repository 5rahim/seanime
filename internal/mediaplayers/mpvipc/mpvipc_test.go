package mpvipc

import (
	"fmt"
	"time"
)

func ExampleConnection_Call() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// toggle play/pause
	_, err = conn.Call("cycle", "pause")
	if err != nil {
		fmt.Print(err)
	}

	// increase volume by 5
	_, err = conn.Call("add", "volume", 5)
	if err != nil {
		fmt.Print(err)
	}

	// decrease volume by 3, showing an osd message and progress bar
	_, err = conn.Call("osd-msg-bar", "add", "volume", -3)
	if err != nil {
		fmt.Print(err)
	}

	// get mpv's version
	version, err := conn.Call("get_version")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("version: %f\n", version.(float64))
}

func ExampleConnection_Set() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// pause playback
	err = conn.Set("pause", true)
	if err != nil {
		fmt.Print(err)
	}

	// seek to the middle of file
	err = conn.Set("percent-pos", 50)
	if err != nil {
		fmt.Print(err)
	}
}

func ExampleConnection_Get() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// see if we're paused
	paused, err := conn.Get("pause")
	if err != nil {
		fmt.Print(err)
	} else if paused.(bool) {
		fmt.Printf("we're paused!\n")
	} else {
		fmt.Printf("we're not paused.\n")
	}

	// see the current position in the file
	elapsed, err := conn.Get("time-pos")
	if err != nil {
		fmt.Print(err)
	} else {
		fmt.Printf("seconds from start of video: %f\n", elapsed.(float64))
	}
}

func ExampleConnection_ListenForEvents() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Call("observe_property", 42, "volume")
	if err != nil {
		fmt.Print(err)
	}

	events := make(chan *Event)
	stop := make(chan struct{})
	go conn.ListenForEvents(events, stop)

	// print all incoming events for 5 seconds, then exit
	go func() {
		time.Sleep(time.Second * 5)
		stop <- struct{}{}
	}()

	for event := range events {
		if event.ID == 42 {
			fmt.Printf("volume now is %f\n", event.Data.(float64))
		} else {
			fmt.Printf("received event: %s\n", event.Name)
		}
	}
}

func ExampleConnection_NewEventListener() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Call("observe_property", 42, "volume")
	if err != nil {
		fmt.Print(err)
	}

	events, stop := conn.NewEventListener()

	// print all incoming events for 5 seconds, then exit
	go func() {
		time.Sleep(time.Second * 5)
		stop <- struct{}{}
	}()

	for event := range events {
		if event.ID == 42 {
			fmt.Printf("volume now is %f\n", event.Data.(float64))
		} else {
			fmt.Printf("received event: %s\n", event.Name)
		}
	}
}

func ExampleConnection_WaitUntilClosed() {
	conn := NewConnection("/tmp/mpv_socket")
	err := conn.Open()
	if err != nil {
		fmt.Print(err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	events, stop := conn.NewEventListener()

	// print events until mpv exits, then exit
	go func() {
		conn.WaitUntilClosed()
		stop <- struct{}{}
	}()

	for event := range events {
		fmt.Printf("received event: %s\n", event.Name)
	}
}
