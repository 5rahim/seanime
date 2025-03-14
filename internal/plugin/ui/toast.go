package plugin_ui

import (
	"seanime/internal/events"

	"github.com/dop251/goja"
)

type ToastManager struct {
	ctx *Context
}

func NewToastManager(ctx *Context) *ToastManager {
	return &ToastManager{
		ctx: ctx,
	}
}

func (t *ToastManager) bind(contextObj *goja.Object) {
	toastObj := t.ctx.vm.NewObject()
	_ = toastObj.Set("success", t.jsToastSuccess)
	_ = toastObj.Set("error", t.jsToastError)
	_ = toastObj.Set("info", t.jsToastInfo)
	_ = toastObj.Set("warning", t.jsToastWarning)

	_ = contextObj.Set("toast", toastObj)
}

func (t *ToastManager) jsToastSuccess(call goja.FunctionCall) goja.Value {
	message, ok := call.Argument(0).Export().(string)
	if !ok {
		t.ctx.handleTypeError("toast: success requires a string message")
		return goja.Undefined()
	}

	t.ctx.wsEventManager.SendEvent(events.SuccessToast, message)
	return goja.Undefined()
}

func (t *ToastManager) jsToastError(call goja.FunctionCall) goja.Value {
	message, ok := call.Argument(0).Export().(string)
	if !ok {
		t.ctx.handleTypeError("toast: error requires a string message")
		return goja.Undefined()
	}

	t.ctx.wsEventManager.SendEvent(events.ErrorToast, message)
	return goja.Undefined()
}

func (t *ToastManager) jsToastInfo(call goja.FunctionCall) goja.Value {
	message, ok := call.Argument(0).Export().(string)
	if !ok {
		t.ctx.handleTypeError("toast: info requires a string message")
		return goja.Undefined()
	}

	t.ctx.wsEventManager.SendEvent(events.InfoToast, message)
	return goja.Undefined()
}

func (t *ToastManager) jsToastWarning(call goja.FunctionCall) goja.Value {
	message, ok := call.Argument(0).Export().(string)
	if !ok {
		t.ctx.handleTypeError("toast: warning requires a string message")
		return goja.Undefined()
	}

	t.ctx.wsEventManager.SendEvent(events.WarningToast, message)
	return goja.Undefined()
}
