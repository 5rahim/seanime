# Code

## Dev notes for the plugin UI

Avoid
```go
func (d *DOMManager) getElementChildren(elementID string) []*goja.Object {

	// Listen for changes from the client
	eventListener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(eventListener.ID)
	payload := ClientDOMElementUpdatedEventPayload{}

	doneCh := make(chan []*goja.Object)

	go func(eventListener *EventListener) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "getChildren" && payload.ElementId == elementID {
					if v, ok := payload.Result.([]interface{}); ok {
						arr := make([]*goja.Object, 0, len(v))
						for _, elem := range v {
							if elemData, ok := elem.(map[string]interface{}); ok {
								arr = append(arr, d.createDOMElementObject(elemData))
							}
						}
						doneCh <- arr
						return
					}
				}
			}
		}
	}(eventListener)

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementId: elementID,
		Action:    "getChildren",
		Params:    map[string]interface{}{},
	})
	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return []*goja.Object{}
	case res := <-doneCh:
		return res
	}
}
```

In the above code
```go
arr = append(arr, d.createDOMElementObject(elemData))
```
Uses the VM so it should be scheduled.
```go
d.ctx.ScheduleAsync(func() error {
	arr := make([]*goja.Object, 0, len(v))
	for _, elem := range v {
		if elemData, ok := elem.(map[string]interface{}); ok {
			arr = append(arr, d.createDOMElementObject(elemData))
		}
	}
	return nil
})
```

However, getElementChildren() might be launched in a scheduled task.
```ts
ctx.registerEventHandler("test", () => {
    const el = ctx.dom.queryOne("#test")
    el.getChildren()
})
```

And since getElementChildren() is coded "synchronously" (without promises), it will block the task
until the timeout and won't run its own task.
You'll end up with something like this:

```txt
> event received
> timeout
> processing scheduled task
> sending task
```

Conclusion: Prefer promises when possible. For synchronous functions, avoid scheduling tasks inside them.
