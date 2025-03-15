package plugin_ui

import (
	"seanime/internal/util/result"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// DOMManager handles DOM manipulation requests from plugins
type DOMManager struct {
	ctx              *Context
	elementObservers *result.Map[string, *ElementObserver]
	eventListeners   *result.Map[string, *DOMEventListener]
}

type ElementObserver struct {
	ID       string
	Selector string
	Callback goja.Callable
}

type DOMEventListener struct {
	ID        string
	ElementID string
	EventType string
	Callback  goja.Callable
}

// NewDOMManager creates a new DOM manager
func NewDOMManager(ctx *Context) *DOMManager {
	return &DOMManager{
		ctx:              ctx,
		elementObservers: result.NewResultMap[string, *ElementObserver](),
		eventListeners:   result.NewResultMap[string, *DOMEventListener](),
	}
}

// BindToObj binds DOM manipulation methods to a context object
func (d *DOMManager) BindToObj(vm *goja.Runtime, obj *goja.Object) {
	domObj := vm.NewObject()
	_ = domObj.Set("query", d.jsQuery)
	_ = domObj.Set("queryOne", d.jsQueryOne)
	_ = domObj.Set("observe", d.jsObserve)
	_ = domObj.Set("createElement", d.jsCreateElement)
	_ = domObj.Set("onReady", d.jsOnReady)

	_ = obj.Set("dom", domObj)
}

func (d *DOMManager) jsOnReady(call goja.FunctionCall) goja.Value {

	callback, ok := goja.AssertFunction(call.Argument(0))
	if !ok {
		d.ctx.handleTypeError("onReady requires a callback function")
	}

	// Listen for changes from the client
	listener := d.ctx.RegisterEventListener(ClientDOMReadyEvent)
	defer d.ctx.UnregisterEventListener(listener.ID)

	go func() {
		for event := range listener.Channel {
			if event.Type == ClientDOMReadyEvent {
				d.ctx.scheduler.ScheduleAsync(func() error {
					_, err := callback(goja.Undefined(), d.ctx.vm.ToValue(event.Payload))
					if err != nil {
						d.ctx.handleException(err)
					}
					return nil
				})
			}
		}
	}()

	return d.ctx.vm.ToValue(nil)
}

// jsQuery handles querying for multiple DOM elements
func (d *DOMManager) jsQuery(call goja.FunctionCall) goja.Value {
	selector := call.Argument(0).String()

	promise, resolve, _ := d.ctx.vm.NewPromise()

	// Generate a unique request ID
	requestId := uuid.New().String()

	// Set up a one-time event listener for the response
	listener := d.ctx.RegisterEventListener(ClientDOMQueryResultEvent)

	var payload ClientDOMQueryResultEventPayload
	go func() {
		defer d.ctx.UnregisterEventListener(listener.ID)
		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMQueryResultEvent, &payload) && payload.RequestID == requestId {
				wg := sync.WaitGroup{}
				elemObjs := make([]interface{}, 0, len(payload.Elements))
				wg.Add(1)
				d.ctx.scheduler.ScheduleAsync(func() error {
					for _, elem := range payload.Elements {
						if elemData, ok := elem.(map[string]interface{}); ok {
							elemObjs = append(elemObjs, d.createDOMElementObject(elemData))
						}
					}
					wg.Done()
					return nil
				})
				wg.Wait()
				d.ctx.scheduler.ScheduleAsync(func() error {
					resolve(d.ctx.vm.ToValue(elemObjs))
					return nil
				})
				return
			}
		}
	}()

	go func() {
		// Send the query request to the client
		d.ctx.SendEventToClient(ServerDOMQueryEvent, &ServerDOMQueryEventPayload{
			Selector:  selector,
			RequestID: requestId,
		})
	}()

	return d.ctx.vm.ToValue(promise)
}

// jsQueryOne handles querying for a single DOM element
func (d *DOMManager) jsQueryOne(call goja.FunctionCall) goja.Value {
	selector := call.Argument(0).String()

	promise, resolve, _ := d.ctx.vm.NewPromise()

	// Generate a unique request ID
	requestId := uuid.New().String()

	// Set up a one-time event listener for the response
	listener := d.ctx.RegisterEventListener(ClientDOMQueryOneResultEvent)

	var payload ClientDOMQueryOneResultEventPayload
	go func() {
		defer d.ctx.UnregisterEventListener(listener.ID)
		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMQueryOneResultEvent, &payload) && payload.RequestID == requestId {
				if payload.Element == nil {
					resolve(goja.Null())
					return
				} else {
					if elemData, ok := payload.Element.(map[string]interface{}); ok {
						d.ctx.scheduler.ScheduleAsync(func() error {
							obj := d.createDOMElementObject(elemData)
							resolve(d.ctx.vm.ToValue(obj))
							return nil
						})
						return
					}
				}
			}
		}
	}()

	// Send the query request to the client
	d.ctx.SendEventToClient(ServerDOMQueryOneEvent, &ServerDOMQueryOneEventPayload{
		Selector:  selector,
		RequestID: requestId,
	})

	return d.ctx.vm.ToValue(promise)
}

// jsObserve starts observing DOM elements matching a selector
func (d *DOMManager) jsObserve(call goja.FunctionCall) goja.Value {
	selector := call.Argument(0).String()
	callback, ok := goja.AssertFunction(call.Argument(1))
	if !ok {
		d.ctx.handleTypeError("observe requires a callback function")
	}

	// Create observer ID
	observerID := uuid.New().String()

	// Store the observer
	observer := &ElementObserver{
		ID:       observerID,
		Selector: selector,
		Callback: callback,
	}

	d.elementObservers.Set(observerID, observer)

	// Send observe request to client
	d.ctx.SendEventToClient(ServerDOMObserveEvent, &ServerDOMObserveEventPayload{
		Selector:   selector,
		ObserverID: observerID,
	})

	// Return a function to stop observing
	return d.ctx.vm.ToValue(func() {
		d.elementObservers.Delete(observerID)

		d.ctx.SendEventToClient(ServerDOMStopObserveEvent, &ServerDOMStopObserveEventPayload{
			ObserverID: observerID,
		})
	})
}

// jsCreateElement creates a new DOM element
func (d *DOMManager) jsCreateElement(call goja.FunctionCall) goja.Value {
	tagName := call.Argument(0).String()

	// Create a promise that will be resolved with the created element
	promise, resolve, _ := d.ctx.vm.NewPromise()

	// Generate a unique request ID
	requestId := uuid.New().String()

	// Set up a one-time event listener for the response
	listener := d.ctx.RegisterEventListener(ClientDOMCreateResultEvent)
	var payload ClientDOMCreateResultEventPayload

	go func() {
		defer d.ctx.UnregisterEventListener(listener.ID)

		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMCreateResultEvent, &payload) && payload.RequestID == requestId {
				if elemData, ok := payload.Element.(map[string]interface{}); ok {
					d.ctx.scheduler.ScheduleAsync(func() error {
						resolve(d.createDOMElementObject(elemData))
						return nil
					})
					return
				}
			}
		}
	}()

	// Send the create request to the client
	d.ctx.SendEventToClient(ServerDOMCreateEvent, &ServerDOMCreateEventPayload{
		TagName:   tagName,
		RequestID: requestId,
	})

	return d.ctx.vm.ToValue(promise)
}

// HandleObserverUpdate processes DOM observer updates from client
func (d *DOMManager) HandleObserverUpdate(observerID string, elements []interface{}) {
	observer, exists := d.elementObservers.Get(observerID)

	if !exists {
		return
	}

	// Convert elements to DOM element objects
	elemObjs := make([]interface{}, 0, len(elements))
	for _, elem := range elements {
		if elemData, ok := elem.(map[string]interface{}); ok {
			elemObjs = append(elemObjs, d.createDOMElementObject(elemData))
		}
	}

	// Schedule callback execution in the VM
	d.ctx.scheduler.ScheduleAsync(func() error {
		_, err := observer.Callback(goja.Undefined(), d.ctx.vm.ToValue(elemObjs))
		if err != nil {
			d.ctx.handleException(err)
		}
		return nil
	})
}

// HandleDOMEvent processes DOM events from client
func (d *DOMManager) HandleDOMEvent(elementID string, eventType string, eventData map[string]interface{}) {
	// Find all event listeners for this element and event type
	d.eventListeners.Range(func(key string, listener *DOMEventListener) bool {
		if listener.ElementID == elementID && listener.EventType == eventType {
			// Schedule callback execution in the VM
			d.ctx.scheduler.ScheduleAsync(func() error {
				_, err := listener.Callback(goja.Undefined(), d.ctx.vm.ToValue(eventData))
				if err != nil {
					d.ctx.handleException(err)
				}
				return nil
			})
		}
		return true
	})
}

// createDOMElementObject creates a JavaScript object representing a DOM element
func (d *DOMManager) createDOMElementObject(elemData map[string]interface{}) *goja.Object {
	elementObj := d.ctx.vm.NewObject()

	// Set basic properties
	elementID, _ := elemData["id"].(string)
	_ = elementObj.Set("id", elementID)

	if tagName, ok := elemData["tagName"].(string); ok {
		_ = elementObj.Set("tagName", tagName)
	}

	if text, ok := elemData["text"].(string); ok {
		_ = elementObj.Set("text", text)
	}

	if attributes, ok := elemData["attributes"].(map[string]interface{}); ok {
		for key, value := range attributes {
			_ = elementObj.Set(key, value)
		}
	}

	if style, ok := elemData["style"].(map[string]interface{}); ok {
		for key, value := range style {
			_ = elementObj.Set(key, value)
		}
	}

	if className, ok := elemData["className"].(string); ok {
		_ = elementObj.Set("className", className)
	}

	if classList, ok := elemData["classList"].([]string); ok {
		_ = elementObj.Set("classList", classList)
	}

	if children, ok := elemData["children"].([]interface{}); ok {
		childrenObjs := make([]*goja.Object, 0, len(children))
		for _, child := range children {
			if childData, ok := child.(map[string]interface{}); ok {
				childrenObjs = append(childrenObjs, d.createDOMElementObject(childData))
			}
		}
		_ = elementObj.Set("children", childrenObjs)
	}

	if parent, ok := elemData["parent"].(map[string]interface{}); ok {
		elementObj.Set("parent", d.createDOMElementObject(parent))
	}

	// Define methods
	_ = elementObj.Set("getText", func() string {
		return d.getElementText(elementID)
	})

	_ = elementObj.Set("setText", func(text string) {
		d.setElementText(elementID, text)
	})

	_ = elementObj.Set("getAttribute", func(name string) interface{} {
		return d.getElementAttribute(elementID, name)
	})

	_ = elementObj.Set("setAttribute", func(name, value string) {
		d.setElementAttribute(elementID, name, value)
	})

	_ = elementObj.Set("removeAttribute", func(name string) {
		d.removeElementAttribute(elementID, name)
	})

	_ = elementObj.Set("addClass", func(className string) {
		d.addElementClass(elementID, className)
	})

	_ = elementObj.Set("removeClass", func(className string) {
		d.removeElementClass(elementID, className)
	})

	_ = elementObj.Set("hasClass", func(className string) bool {
		return d.hasElementClass(elementID, className)
	})

	_ = elementObj.Set("setStyle", func(property, value string) {
		d.setElementStyle(elementID, property, value)
	})

	_ = elementObj.Set("getStyle", func(property string) string {
		return d.getElementStyle(elementID, property)
	})

	_ = elementObj.Set("getComputedStyle", func(property string) string {
		return d.getElementComputedStyle(elementID, property)
	})

	_ = elementObj.Set("append", func(child *goja.Object) {
		childID := child.Get("id").String()
		d.appendElement(elementID, childID)
	})

	_ = elementObj.Set("before", func(sibling *goja.Object) {
		siblingID := sibling.Get("id").String()
		d.insertElementBefore(elementID, siblingID)
	})

	_ = elementObj.Set("after", func(sibling *goja.Object) {
		siblingID := sibling.Get("id").String()
		d.insertElementAfter(elementID, siblingID)
	})

	_ = elementObj.Set("remove", func() {
		d.removeElement(elementID)
	})

	_ = elementObj.Set("getParent", func() goja.Value {
		return d.getElementParent(elementID)
	})

	_ = elementObj.Set("getChildren", func() goja.Value {
		return d.getElementChildren(elementID)
	})

	_ = elementObj.Set("addEventListener", func(event string, callback goja.Callable) func() {
		return d.addElementEventListener(elementID, event, callback)
	})

	return elementObj
}

// Element manipulation methods
// These send events to the client and handle responses

func (d *DOMManager) getElementText(elementID string) string {

	// Set up a one-time event listener for the response
	listener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(listener.ID)

	var payload ClientDOMElementUpdatedEventPayload
	doneCh := make(chan string)

	go func() {
		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "getText" && payload.ElementID == elementID {
					doneCh <- payload.Result.(string)
					return
				}
			}
		}
	}()

	// Send the request to the client
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getText",
		Params:    map[string]interface{}{},
	})

	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return ""
	case res := <-doneCh:
		return res
	}
}

func (d *DOMManager) setElementText(elementID, text string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "setText",
		Params: map[string]interface{}{
			"text": text,
		},
	})
}

func (d *DOMManager) getElementAttribute(elementID, name string) interface{} {
	listener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(listener.ID)

	var payload ClientDOMElementUpdatedEventPayload
	doneCh := make(chan interface{})

	go func() {
		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "getAttribute" && payload.ElementID == elementID {
					doneCh <- payload.Result
				}
			}
		}
	}()

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getAttribute",
		Params: map[string]interface{}{
			"name": name,
		},
	})

	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return nil
	case res := <-doneCh:
		return res
	}
}

func (d *DOMManager) setElementAttribute(elementID, name, value string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "setAttribute",
		Params: map[string]interface{}{
			"name":  name,
			"value": value,
		},
	})
}

func (d *DOMManager) removeElementAttribute(elementID, name string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "removeAttribute",
		Params: map[string]interface{}{
			"name": name,
		},
	})
}

func (d *DOMManager) addElementClass(elementID, className string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "addClass",
		Params: map[string]interface{}{
			"className": className,
		},
	})
}

func (d *DOMManager) removeElementClass(elementID, className string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "removeClass",
		Params: map[string]interface{}{
			"className": className,
		},
	})
}

func (d *DOMManager) hasElementClass(elementID, className string) bool {
	listener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(listener.ID)

	var payload ClientDOMElementUpdatedEventPayload
	doneCh := make(chan bool)

	go func() {
		for event := range listener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "hasClass" && payload.ElementID == elementID {
					if v, ok := payload.Result.(bool); ok {
						doneCh <- v
					} else {
						doneCh <- false
					}
				}
			}
		}
	}()

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "hasClass",
		Params: map[string]interface{}{
			"className": className,
		},
	})

	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return false
	case res := <-doneCh:
		return res
	}
}

func (d *DOMManager) setElementStyle(elementID, property, value string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "setStyle",
		Params: map[string]interface{}{
			"property": property,
			"value":    value,
		},
	})
}

func (d *DOMManager) getElementStyle(elementID, property string) string {

	// Listen for changes from the client
	eventListener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(eventListener.ID)
	payload := ClientDOMElementUpdatedEventPayload{}

	doneCh := make(chan string)

	go func(eventListener *EventListener) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) && payload.ElementID == elementID {
				if payload.Action == "getStyle" {
					if v, ok := payload.Result.(string); ok {
						doneCh <- v
					} else {
						doneCh <- ""
					}
				}
			}
		}
	}(eventListener)

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getStyle",
		Params: map[string]interface{}{
			"property": property,
		},
	})

	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return ""
	case res := <-doneCh:
		return res
	}
}

func (d *DOMManager) getElementComputedStyle(elementID, property string) string {
	// Listen for changes from the client
	eventListener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(eventListener.ID)
	payload := ClientDOMElementUpdatedEventPayload{}

	doneCh := make(chan string)

	go func(eventListener *EventListener) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) && payload.ElementID == elementID {
				if payload.Action == "getComputedStyle" {
					if v, ok := payload.Result.(string); ok {
						doneCh <- v
					} else {
						doneCh <- ""
					}
				}
			}
		}
	}(eventListener)

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getComputedStyle",
		Params: map[string]interface{}{
			"property": property,
		},
	})

	timeout := time.After(4 * time.Second)

	select {
	case <-timeout:
		return ""
	case res := <-doneCh:
		return res
	}
}

func (d *DOMManager) appendElement(parentID, childID string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: parentID,
		Action:    "append",
		Params: map[string]interface{}{
			"childID": childID,
		},
	})
}

func (d *DOMManager) insertElementBefore(elementID, siblingID string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "before",
		Params: map[string]interface{}{
			"siblingID": siblingID,
		},
	})
}

func (d *DOMManager) insertElementAfter(elementID, siblingID string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "after",
		Params: map[string]interface{}{
			"siblingID": siblingID,
		},
	})
}

func (d *DOMManager) removeElement(elementID string) {
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "remove",
		Params:    map[string]interface{}{},
	})
}

func (d *DOMManager) getElementParent(elementID string) goja.Value {
	promise, resolve, _ := d.ctx.vm.NewPromise()
	// Listen for changes from the client
	eventListener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(eventListener.ID)
	payload := ClientDOMElementUpdatedEventPayload{}

	go func(eventListener *EventListener) {

		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "getParent" && payload.ElementID == elementID {
					if v, ok := payload.Result.(map[string]interface{}); ok {
						resolve(d.createDOMElementObject(v))
						return
					} else {
						resolve(goja.Null())
					}
				}
			}
		}
	}(eventListener)

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getParent",
		Params:    map[string]interface{}{},
	})

	return d.ctx.vm.ToValue(promise)
}

func (d *DOMManager) getElementChildren(elementID string) goja.Value {
	promise, resolve, _ := d.ctx.vm.NewPromise()

	// Listen for changes from the client
	eventListener := d.ctx.RegisterEventListener(ClientDOMElementUpdatedEvent)
	defer d.ctx.UnregisterEventListener(eventListener.ID)
	payload := ClientDOMElementUpdatedEventPayload{}

	go func(eventListener *EventListener) {
		for event := range eventListener.Channel {
			if event.ParsePayloadAs(ClientDOMElementUpdatedEvent, &payload) {
				if payload.Action == "getChildren" && payload.ElementID == elementID {
					if v, ok := payload.Result.([]interface{}); ok {
						d.ctx.scheduler.ScheduleAsync(func() error {
							arr := make([]*goja.Object, 0, len(v))
							for _, elem := range v {
								if elemData, ok := elem.(map[string]interface{}); ok {
									arr = append(arr, d.createDOMElementObject(elemData))
								}
							}
							resolve(d.ctx.vm.ToValue(arr))
							return nil
						})
						return
					} else {
						d.ctx.scheduler.ScheduleAsync(func() error {
							resolve(goja.Null())
							return nil
						})
					}
				}
			}
		}
	}(eventListener)

	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "getChildren",
		Params:    map[string]interface{}{},
	})

	return d.ctx.vm.ToValue(promise)
}

func (d *DOMManager) addElementEventListener(elementID, event string, callback goja.Callable) func() {
	// Create a unique ID for this event listener
	listenerID := uuid.New().String()

	// Store the event listener
	listener := &DOMEventListener{
		ID:        listenerID,
		ElementID: elementID,
		EventType: event,
		Callback:  callback,
	}

	d.eventListeners.Set(listenerID, listener)

	// Send the request to add the event listener to the client
	d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
		ElementID: elementID,
		Action:    "addEventListener",
		Params: map[string]interface{}{
			"event":      event,
			"listenerID": listenerID,
		},
	})

	// Return a function to remove the event listener
	return func() {
		d.eventListeners.Delete(listenerID)

		// Send the request to remove the event listener from the client
		d.ctx.SendEventToClient(ServerDOMManipulateEvent, &ServerDOMManipulateEventPayload{
			ElementID: elementID,
			Action:    "removeEventListener",
			Params: map[string]interface{}{
				"event":      event,
				"listenerID": listenerID,
			},
		})
	}
}
