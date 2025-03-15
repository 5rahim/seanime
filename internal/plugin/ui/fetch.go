package plugin_ui

import (
	"seanime/internal/goja/goja_bindings"

	"github.com/dop251/goja"
)

func (c *Context) bindFetch(obj *goja.Object) {
	f := goja_bindings.NewFetch(c.vm)

	_ = obj.Set("fetch", f.Fetch)

	go func() {
		for fn := range f.ResponseChannel() {
			c.scheduler.ScheduleAsync(func() error {
				fn()
				return nil
			})
		}
	}()

	c.registerOnCleanup(func() {
		c.logger.Debug().Msg("plugin: Terminating fetch")
		f.Close()
	})
}
