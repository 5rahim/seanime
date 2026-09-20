package plugin_ui

import (
	"fmt"
	"seanime/internal/extension_repo/prompt"
	"seanime/internal/goja/goja_bindings"
	"seanime/internal/plugin"
	"seanime/internal/security"
	"seanime/internal/util/result"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

type MarketplaceManager struct {
	ctx   *Context
	cache *result.Cache[string, bool]
}

func NewMarketplaceManager(ctx *Context) *MarketplaceManager {
	return &MarketplaceManager{
		ctx:   ctx,
		cache: result.NewCache[string, bool](),
	}
}

func (m *MarketplaceManager) bindExtensions(extensionsObj *goja.Object) {
	_ = extensionsObj.Set("getMarketplaceUrl", m.jsGetMarketplaceUrl)
	_ = extensionsObj.Set("setMarketplaceUrl", m.jsSetMarketplaceUrl)
}

// jsGetMarketplaceUrl asks the client for the current marketplace URL.
//
//	Example:
//	const url = await ctx.extensions.getMarketplaceUrl()
func (m *MarketplaceManager) jsGetMarketplaceUrl() goja.Value {
	vm := m.ctx.vm
	promise, resolve, reject := vm.NewPromise()
	requestId := uuid.New().String()

	go func() {
		err := plugin.GlobalAppContext.Ask(m.ctx.ext, prompt.Options{
			Kind:     "extensions",
			Action:   "view the marketplace URL",
			Resource: "marketplace URL",
			Message:  fmt.Sprintf("Allow %q to view the marketplace URL?", m.ctx.ext.Name),
			Details:  []string{},
			Cache:    m.cache,
			CacheKey: "extensions:marketplace:view",
		})
		if err != nil {
			m.rejectAsync(reject, err)
			return
		}

		listener := m.ctx.RegisterEventListener(ClientMarketplaceGetURLResultEvent)
		listener.SetCallback(func(event *ClientPluginEvent) {
			var payload ClientMarketplaceGetURLResultEventPayload
			if !event.ParsePayloadAs(ClientMarketplaceGetURLResultEvent, &payload) || payload.RequestID != requestId {
				return
			}
			m.ctx.UnregisterEventListener(listener.ID)
			m.ctx.scheduler.ScheduleAsync(func() error {
				resolve(vm.ToValue(payload.URL))
				return nil
			})
		})

		m.ctx.SendEventToClient(ServerMarketplaceGetURLEvent, &ServerMarketplaceGetURLEventPayload{
			RequestID: requestId,
		})
	}()

	return vm.ToValue(promise)
}

// jsSetMarketplaceUrl asks the client to change the marketplace URL.
// An empty URL resets it to the default marketplace.
//
//	Example:
//	await ctx.extensions.setMarketplaceUrl("https://example.com/marketplace.json")
func (m *MarketplaceManager) jsSetMarketplaceUrl(url string) goja.Value {
	vm := m.ctx.vm
	promise, resolve, reject := vm.NewPromise()

	go func() {
		if url != "" {
			if err := security.ValidateOutboundUrl(url); err != nil {
				m.rejectAsync(reject, err)
				return
			}
		}

		err := plugin.GlobalAppContext.Ask(m.ctx.ext, prompt.Options{
			Kind:     "extensions",
			Action:   "change the marketplace URL",
			Resource: url,
			Message:  fmt.Sprintf("Allow %q to change the marketplace URL to %q?", m.ctx.ext.Name, url),
			Details:  []string{url},
			Cache:    m.cache,
			CacheKey: "extensions:marketplace:set:" + url,
		})
		if err != nil {
			m.rejectAsync(reject, err)
			return
		}

		m.ctx.SendEventToClient(ServerMarketplaceSetURLEvent, &ServerMarketplaceSetURLEventPayload{
			URL: url,
		})

		m.ctx.scheduler.ScheduleAsync(func() error {
			resolve(vm.ToValue(true))
			return nil
		})
	}()

	return vm.ToValue(promise)
}

func (m *MarketplaceManager) rejectAsync(reject func(reason interface{}) error, err error) {
	m.ctx.scheduler.ScheduleAsync(func() error {
		reject(goja_bindings.NewErrorString(m.ctx.vm, err.Error()))
		return nil
	})
}
