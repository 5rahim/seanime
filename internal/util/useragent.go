package util

import "github.com/mileusna/useragent"

const (
	PlatformAndroid  = "android"
	PlatformIOS      = "ios"
	PlatformLinux    = "linux"
	PlatformMac      = "mac"
	PlatformWindows  = "windows"
	PlatformChromeOS = "chromeos"
)

const (
	DeviceDesktop = "desktop"
	DeviceMobile  = "mobile"
	DeviceTablet  = "tablet"
)

type ClientInfo struct {
	Device   string
	Platform string
}

func GetClientInfo(userAgent string) ClientInfo {
	ua := useragent.Parse(userAgent)

	var device string
	var platform string

	if ua.Mobile {
		device = DeviceMobile
	} else if ua.Tablet {
		device = DeviceTablet
	} else {
		device = DeviceDesktop
	}

	platform = ua.OS
	if platform == "" {
		platform = "-"
	}

	return ClientInfo{
		Device:   device,
		Platform: platform,
	}
}
