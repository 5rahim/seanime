package plugin_ui

import (
	"seanime/internal/notifier"

	"github.com/dop251/goja"
)

type NotificationManager struct {
	ctx *Context
}

func NewNotificationManager(ctx *Context) *NotificationManager {
	return &NotificationManager{
		ctx: ctx,
	}
}

func (n *NotificationManager) bind(contextObj *goja.Object) {
	notificationObj := n.ctx.vm.NewObject()
	_ = notificationObj.Set("send", n.jsNotify)

	_ = contextObj.Set("notification", notificationObj)
}

func (n *NotificationManager) jsNotify(call goja.FunctionCall) goja.Value {
	message, ok := call.Argument(0).Export().(string)
	if !ok {
		n.ctx.handleTypeError("notification: notify requires a string message")
		return goja.Undefined()
	}

	notifier.GlobalNotifier.Notify(notifier.Notification(n.ctx.ext.Name), message)

	return goja.Undefined()
}
