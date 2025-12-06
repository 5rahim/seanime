package core

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
)

type (
	SeanimeFlags struct {
		DataDir          string
		Host             string
		Port             int
		Update           bool
		IsDesktopSidecar bool
		DisableFeatures  []FeatureKey
		Password         string
		DisablePassword  bool
		LockDown         bool
	}
)

func GetSeanimeFlags() SeanimeFlags {
	flags := SeanimeFlags{}
	var disableFeaturesStr string

	flag.Usage = func() {
		fmt.Printf("The Anime and Manga media server.\n\n")
		if runtime.GOOS == "windows" {
			fmt.Printf("Usage: seanime.exe [flags]\n\n")
		} else {
			fmt.Printf("Usage: seanime [flags]\n\n")
		}
		fmt.Printf("Flags:\n")
		fmt.Printf("  --datadir string              directory that contains all Seanime data\n")
		fmt.Printf("  --host string                 host address to bind to (default: 127.0.0.1)\n")
		fmt.Printf("  --port int                    port to bind to (default: 43211)\n")
		fmt.Printf("  --update                      update the application\n")
		fmt.Printf("  --desktop-sidecar             run as the desktop sidecar\n")
		fmt.Printf("  --disable-features string     comma-separated list of features to disable\n")
		fmt.Printf("  --disable-all-features        disable all features that can be disabled\n")
		fmt.Printf("  --password string             password to use for the instance\n")
		fmt.Printf("  --disable-password            disable password protection\n")
		fmt.Printf("  -h                           show this help message\n")
	}

	flag.StringVar(&flags.DataDir, "datadir", "", "Directory that contains all Seanime data")
	flag.StringVar(&flags.Host, "host", "", "Host address to bind to")
	flag.IntVar(&flags.Port, "port", 0, "Port to bind to")
	flag.BoolVar(&flags.Update, "update", false, "Update the application")
	flag.BoolVar(&flags.IsDesktopSidecar, "desktop-sidecar", false, "Run as the desktop sidecar")
	flag.StringVar(&disableFeaturesStr, "disable-features", "", "Comma-separated list of features to disable")
	flag.BoolVar(&flags.LockDown, "disable-all-features", false, "Disables all features that can be disabled")
	flag.StringVar(&flags.Password, "password", "", "Password to use for the instance")
	flag.BoolVar(&flags.DisablePassword, "disable-password", false, "Disable password protection")

	flag.Parse()

	flags.DataDir = strings.TrimSpace(flags.DataDir)
	flags.Host = strings.TrimSpace(flags.Host)

	if disableFeaturesStr != "" {
		features := strings.Split(disableFeaturesStr, ",")
		for _, feature := range features {
			if trimmed := strings.TrimSpace(feature); trimmed != "" {
				flags.DisableFeatures = append(flags.DisableFeatures, FeatureKey(trimmed))
			}
		}
	}

	return flags
}
