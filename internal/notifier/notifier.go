package notifier

import (
	"github.com/samber/mo"
	"seanime/internal/database/models"
	"sync"
)

type (
	Notifier struct {
		dataDir  mo.Option[string]
		settings mo.Option[*models.NotificationSettings]
		mu       sync.Mutex
	}
)

var GlobalNotifier = NewNotifier()

func NewNotifier() *Notifier {
	return &Notifier{
		dataDir:  mo.None[string](),
		settings: mo.None[*models.NotificationSettings](),
		mu:       sync.Mutex{},
	}
}

func (n *Notifier) SetSettings(datadir string, settings *models.NotificationSettings) {
	n.mu.Lock()
	n.dataDir = mo.Some(datadir)
	n.settings = mo.Some(settings)
	n.mu.Unlock()
}
