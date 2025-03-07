package plugin_ui

import (
	"seanime/internal/util/result"

	"github.com/dop251/goja"
)

type CommandPalette struct {
	ctx *Context

	commands *result.Map[string, *Command]
}

type Command struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Value   string `json:"value"`
	OnClick string `json:"onClick"`
	Show    bool   `json:"show"`
}

func NewCommandPalette(ctx *Context) *CommandPalette {
	return &CommandPalette{
		ctx:      ctx,
		commands: result.NewResultMap[string, *Command](),
	}
}

func (c *CommandPalette) bind(ctxObj *goja.Object) {
	commandPaletteObj := c.ctx.vm.NewObject()
	_ = ctxObj.Set("commandPalette", commandPaletteObj)
}
