package plugin_ui

import (
	"net/url"
	"strings"
	"sync"

	"github.com/dop251/goja"
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

				parsedQuery, _ := url.ParseQuery(strings.TrimPrefix(payload.Query, "?"))
				queryMap := make(map[string]string)
				for key, value := range parsedQuery {
					queryMap[key] = strings.Join(value, ",")
				}

				ret := map[string]interface{}{
					"pathname":     payload.Pathname,
					"searchParams": queryMap,
				}

				_, err := callback(goja.Undefined(), s.ctx.vm.ToValue(ret))
				return err
			})
		}
	})

	return goja.Undefined()
}
