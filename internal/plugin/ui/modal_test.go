package plugin_ui

import (
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/util"
	gojautil "seanime/internal/util/goja"
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestModalOpenUpdateAndClose(t *testing.T) {
	logger := util.NewLogger()
	scheduler := gojautil.NewScheduler()
	ws := events.NewMockWSEventManager(logger)
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	ui := NewUI(NewUIOptions{
		Logger:    logger,
		VM:        vm,
		WSManager: ws,
		Scheduler: scheduler,
		Extension: &extension.Extension{ID: "modal-test", Plugin: &extension.PluginManifest{}},
	})
	t.Cleanup(func() {
		ui.context.Stop()
	})
	ctx := ui.context
	run := func(script string) {
		t.Helper()
		require.NoError(t, scheduler.Schedule(func() error {
			_, err := ui.vm.RunString(script)
			return err
		}))
		require.NoError(t, scheduler.Schedule(func() error { return nil }))
		ctx.flushEventBatch()
	}
	require.NoError(t, scheduler.Schedule(func() error {
		if err := ui.vm.Set("newModal", ctx.modalManager.jsNewModal); err != nil {
			return err
		}
		button := &AnimePageButton{BaseAction: BaseAction{BaseActionProps: BaseActionProps{ID: "open-details", Label: "Details"}}}
		return ui.vm.Set("button", button.CreateObject(ctx.actionManager))
	}))
	run(`
        let label = "first";
        const modal = newModal({ title: "Details" });
        modal.render(() => [modal.text(label)]);
        button.onClick(() => modal.open());
    `)
	require.Empty(t, ws.Events())
	listeners, ok := ctx.eventBus.Get(ClientActionClickedEvent)
	require.True(t, ok)
	listeners.Range(func(_ string, listener *EventListener) bool {
		listener.callback(&ClientPluginEvent{
			Type:    ClientActionClickedEvent,
			Payload: ClientActionClickedEventPayload{ActionID: "open-details"},
		})
		return true
	})
	run(``)
	require.Len(t, ws.Events(), 1)
	open := ws.Events()[0].Payload.(*ServerPluginEvent)
	require.Equal(t, "modal-test", open.ExtensionID)
	payload := open.Payload.(ServerModalUpdatedEventPayload)
	require.Equal(t, "open", payload.Action)
	require.Equal(t, "Details", payload.Title)
	require.Len(t, payload.Components, 1)

	run(`label = "updated"`)
	ctx.triggerUIUpdate()
	run(``)
	require.Len(t, ws.Events(), 2)
	update := ws.Events()[1].Payload.(*ServerPluginEvent).Payload.(ServerModalUpdatedEventPayload)
	require.Equal(t, "update", update.Action)
	components := update.Components.([]interface{})
	props := components[0].(map[string]interface{})["props"].(map[string]interface{})
	require.Equal(t, "updated", props["text"])

	run(`modal.close()`)
	require.Len(t, ws.Events(), 3)
	close := ws.Events()[2].Payload.(*ServerPluginEvent).Payload.(ServerModalUpdatedEventPayload)
	require.Equal(t, "close", close.Action)
	run(`modal.update()`)
	require.Len(t, ws.Events(), 3)

	run(`modal.open()`)
	require.Len(t, ws.Events(), 4)
	listeners, ok = ctx.eventBus.Get(ClientModalClosedEvent)
	require.True(t, ok)
	listeners.Range(func(_ string, listener *EventListener) bool {
		listener.callback(&ClientPluginEvent{Type: ClientModalClosedEvent})
		return true
	})
	run(`modal.update()`)
	require.Len(t, ws.Events(), 4)
}
