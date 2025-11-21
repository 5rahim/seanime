package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"seanime/internal/matroska"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <mkv-file>\n", os.Args)
		return
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	r, err := matroska.NewDemuxer(f)
	if err != nil {
		log.Fatalf("failed to create matroska demuxer: %v", err)
	}

	fmt.Println("--- Seeking by timecode ---")
	targetTime := 37*time.Minute + 33*time.Second + 689*time.Millisecond
	r.Seek(uint64(targetTime.Nanoseconds()), 0)

	p, err := r.ReadPacket()
	if err != nil {
		log.Fatalf("failed to read packet after seek: %v", err)
	}
	fmt.Printf(
		"Found packet at: %s, Track: %d, Keyframe: %v, Size: %d\n",
		time.Duration(p.StartTime),
		p.Track,
		(p.Flags&matroska.KF) != 0,
		len(p.Data),
	)

	chapters := r.GetChapters()
	if len(chapters) > 0 {
		fmt.Println("\n--- Seeking by chapter ---")
		fmt.Println("Chapters found:")
		for i, ch := range chapters {
			var chapterName string
			if len(ch.Display) > 0 {
				chapterName = ch.Display[0].String
			} else {
				chapterName = "Unnamed Chapter"
			}
			fmt.Printf("  %d: %s (%s)\n", i, chapterName, time.Duration(ch.Start))
		}

		for i, ch := range chapters {
			var chapterName string
			if len(ch.Display) > 0 {
				chapterName = ch.Display[0].String
			} else {
				chapterName = "Unnamed Chapter"
			}
			fmt.Printf("\n--- Seeking to chapter %d: %s (%s) ---\n", i, chapterName, time.Duration(ch.Start))
			r.Seek(ch.Start, 0)

			pReadPacket, errReadPacket := r.ReadPacket()
			if errReadPacket != nil {
				log.Printf("failed to read packet after seek for chapter %d: %v", i, errReadPacket)
				continue
			}

			fmt.Printf(
				"Found packet at: %s, Track: %d, Keyframe: %v, Size: %d\n",
				time.Duration(pReadPacket.StartTime),
				pReadPacket.Track,
				(pReadPacket.Flags&matroska.KF) != 0,
				len(pReadPacket.Data),
			)
		}
	} else {
		fmt.Println("\nNo chapters found in the file.")
	}
}
