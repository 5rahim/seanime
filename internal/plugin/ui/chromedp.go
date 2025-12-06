package plugin_ui

import (
	"seanime/internal/goja/goja_bindings"

	"github.com/dop251/goja"
)

func (c *Context) bindChromeDP(obj *goja.Object) {
	cdp := goja_bindings.NewChromeDP(c.vm)

	_ = obj.Set("chromeDP", cdp)

	c.registerOnCleanup(func() {
		c.logger.Debug().Msg("plugin: Terminating fetch")
		cdp.Close()
	})
}
