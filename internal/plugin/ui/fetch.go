package plugin_ui

import (
	"seanime/internal/goja/goja_bindings"

	"github.com/dop251/goja"
)

func (c *Context) bindFetch(obj *goja.Object, allowedDomains []string, anilistToken string) {
	f := goja_bindings.NewFetch(c.vm, allowedDomains)
	f.SetAnilistToken(anilistToken)

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

func (c *Context) bindAbortContext() {
	goja_bindings.BindAbortContext(c.vm, c.scheduler)
}
