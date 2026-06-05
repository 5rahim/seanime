//go:build !windows && !android && !ios

package notifier

import (
	"github.com/gen2brain/beeep"
)

func defaultPush(title, message, icon string) error {
	return beeep.Notify(title, message, icon)
}
