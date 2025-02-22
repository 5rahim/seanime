package plugin_ui

type WebviewManager struct {
	context *Context
}

func NewWebviewManager(context *Context) *WebviewManager {
	return &WebviewManager{
		context: context,
	}
}
