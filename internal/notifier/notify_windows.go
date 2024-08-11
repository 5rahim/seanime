//go:build windows

package notifier

import (
	"github.com/go-toast/toast"
	"seanime/internal/util"
)

// Notify sends a notification to the user.
// This is run in a goroutine.
func (n *Notifier) Notify(id Notification, message string) {
	go func() {
		defer util.HandlePanicInModuleThen("notifier/Notify", func() {})

		n.mu.Lock()
		defer n.mu.Unlock()

		if !n.canProceed(id) {
			return
		}

		notification := toast.Notification{
			AppID:   "Seanime",
			Title:   string(id),
			Message: message,
			Icon:    n.logoPath,
		}

		err := notification.Push()
		if err != nil {
			if n.logger.IsPresent() {
				n.logger.MustGet().Trace().Msgf("notifier: Failed to push notification: %v", err)
			}
		}
		if n.logger.IsPresent() {
			n.logger.MustGet().Trace().Msgf("notifier: Pushed notification: %v", id)
		}
	}()
}
