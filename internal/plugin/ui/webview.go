package plugin_ui

type WebviewManager struct {
	ctx *Context
}

func NewWebviewManager(ctx *Context) *WebviewManager {
	return &WebviewManager{
		ctx: ctx,
	}
}
