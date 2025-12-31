package plugin_ui

import (
	"net/url"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type ScreenManager struct {
	ctx *Context
	mu  sync.RWMutex
}

func NewScreenManager(ctx *Context) *ScreenManager {
	return &ScreenManager{
		ctx: ctx,
	}
}

// bind binds 'screen' to the ctx object
//
//	Example:
//	ctx.screen.navigateTo("/entry?id=21");
func (s *ScreenManager) bind(ctxObj *goja.Object) {
	screenObj := s.ctx.vm.NewObject()
	_ = screenObj.Set("onNavigate", s.jsOnNavigate)
	_ = screenObj.Set("navigateTo", s.jsNavigateTo)
	_ = screenObj.Set("reload", s.jsReload)
	_ = screenObj.Set("loadCurrent", s.jsLoadCurrent)
	_ = screenObj.Set("state", s.jsState)

	_ = ctxObj.Set("screen", screenObj)
}

// jsNavigateTo navigates to a new screen
//
//	Example:
//	ctx.screen.navigateTo("/entry?id=21");
func (s *ScreenManager) jsNavigateTo(path string, searchParams map[string]string) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	queryString := ""
	if len(searchParams) > 0 {
		query := url.Values{}
		for key, value := range searchParams {
			query.Add(key, value)
		}
		queryString = "?" + query.Encode()
	}

	finalPath := path + queryString

	s.ctx.SendEventToClient(ServerScreenNavigateToEvent, ServerScreenNavigateToEventPayload{
		Path: finalPath,
	})
}

// jsReload reloads the current screen
func (s *ScreenManager) jsReload() {
	s.ctx.SendEventToClient(ServerScreenReloadEvent, ServerScreenReloadEventPayload{})
}

// jsLoadCurrent calls onNavigate with the current screen data
func (s *ScreenManager) jsLoadCurrent() {
	s.ctx.SendEventToClient(ServerScreenGetCurrentEvent, ServerScreenGetCurrentEventPayload{})
}

// parse calls onNavigate with the current screen data
func (s *ScreenManager) parse(pathname, query string) map[string]interface{} {
	parsedQuery, _ := url.ParseQuery(strings.TrimPrefix(query, "?"))
	queryMap := make(map[string]string)
	for key, value := range parsedQuery {
		queryMap[key] = strings.Join(value, ",")
	}

	return map[string]interface{}{
		"pathname":     pathname,
		"searchParams": queryMap,
	}
}

// jsOnNavigate registers a callback to be called when the current screen changes
//
//	Example:
//	const onNavigate = (event) => {
//		console.log(event.screen);
//	};
//	ctx.screen.onNavigate(onNavigate);
func (s *ScreenManager) jsOnNavigate(callback goja.Callable) goja.Value {
	eventListener := s.ctx.RegisterEventListener(ClientScreenChangedEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientScreenChangedEventPayload
		if event.ParsePayloadAs(ClientScreenChangedEvent, &payload) {
			s.ctx.scheduler.ScheduleAsync(func() error {
				ret := s.parse(payload.Pathname, payload.Query)
				_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(ret))
				return err
			})
		}
	})

	return goja.Undefined()
}

// jsState returns a new state object
//
//	Example:
//	const screen = ctx.screen.state()
//	screen.get().pathname
func (s *ScreenManager) jsState(call goja.FunctionCall) goja.Value {
	id := uuid.New().String()
	initial := s.ctx.vm.ToValue(map[string]interface{}{
		"pathname":     "",
		"searchParams": map[string]string{},
	})

	state := &State{
		ID:    id,
		Value: initial,
	}

	// Store the initial state
	s.ctx.states.Set(id, state)

	jsGetState := func(call goja.FunctionCall) goja.Value {
		res, _ := s.ctx.states.Get(id)
		return res.Value
	}

	eventListener := s.ctx.RegisterEventListener(ClientScreenChangedEvent)

	eventListener.SetCallback(func(event *ClientPluginEvent) {
		var payload ClientScreenChangedEventPayload
		if event.ParsePayloadAs(ClientScreenChangedEvent, &payload) {
			s.ctx.scheduler.ScheduleAsync(func() error {
				ret := s.parse(payload.Pathname, payload.Query)
				s.ctx.states.Set(id, &State{
					ID:    id,
					Value: s.ctx.vm.ToValue(ret),
				})
				s.ctx.queueStateUpdate(id)
				return nil
			})
		}
	})

	s.jsLoadCurrent()

	return s.ctx.createStateObject(id, jsGetState, nil)
}
