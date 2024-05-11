package core

import (
	"flag"
	"fmt"
	"strings"
)

type (
	SeanimeFlags struct {
		DataDir string
		TrueWd  bool
	}
)

func GetSeanimeFlags() SeanimeFlags {
	// Help flag
	flag.Usage = func() {
		fmt.Printf("Self-hosted, user-friendly, media server for anime and manga enthusiasts.\n\n")
		fmt.Printf("Usage:\n  seanime [flags]\n\n")
		fmt.Printf("Flags:\n")
		fmt.Printf("  -datadir, --datadir string")
		fmt.Printf("   directory that contains all Seanime data\n")
		fmt.Printf("  -h                           show this help message\n")
	}
	// Parse flags
	var dataDir string
	flag.StringVar(&dataDir, "datadir", "", "Directory that contains all Seanime data")
	var truewd bool
	flag.BoolVar(&truewd, "truewd", false, "Force Seanime to use the binary's directory as the working directory")
	flag.Parse()

	return SeanimeFlags{
		DataDir: strings.TrimSpace(dataDir),
		TrueWd:  truewd,
	}
}
