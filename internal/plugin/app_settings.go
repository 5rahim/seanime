package plugin

import (
	"errors"
	"reflect"
	"sort"
	"strings"
	"time"

	"seanime/internal/database/models"
	"seanime/internal/extension"
	"seanime/internal/extension_repo/prompt"
	"seanime/internal/goja/goja_bindings"
	gojautil "seanime/internal/util/goja"

	"github.com/dop251/goja"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
)

const (
	appSettingsMediastreamRoot   = "mediastream"
	appSettingsTorrentstreamRoot = "torrentstream"
	appSettingsDebridRoot        = "debrid"
)

type appSettingsBundle struct {
	Settings      *models.Settings
	Mediastream   *models.MediastreamSettings
	Torrentstream *models.TorrentstreamSettings
	Debrid        *models.DebridSettings
}

type appSettingsChanges struct {
	base          bool
	mediastream   bool
	torrentstream bool
	debrid        bool
}

func (o appSettingsChanges) hasChanges() bool {
	return o.base || o.mediastream || o.torrentstream || o.debrid
}

func (o *appSettingsChanges) markKey(key string) {
	switch strings.TrimSpace(key) {
	case appSettingsMediastreamRoot:
		o.mediastream = true
	case appSettingsTorrentstreamRoot:
		o.torrentstream = true
	case appSettingsDebridRoot:
		o.debrid = true
	case "":
		return
	default:
		o.base = true
	}
}

func changesForPath(path string) (ret appSettingsChanges) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return ret
	}
	ret.markKey(parts[0])
	return ret
}

func changesForMap(in map[string]interface{}) (ret appSettingsChanges) {
	for key := range in {
		ret.markKey(key)
	}
	return ret
}

func (a *AppContextImpl) bindSettingsObj(vm *goja.Runtime, ext *extension.Extension, scheduler *gojautil.Scheduler) *goja.Object {
	settingsObj := vm.NewObject()
	cache := newPromptCache()

	_ = settingsObj.Set("get", func(call goja.FunctionCall) goja.Value {
		path := ""
		if len(call.Arguments) > 0 && !goja.IsUndefined(call.Argument(0)) && !goja.IsNull(call.Argument(0)) {
			path = call.Argument(0).String()
		}

		hasFallback := false
		var fallback interface{}
		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) {
			hasFallback = true
			fallback = call.Argument(1).Export()
		}

		detail := "All settings"
		path = strings.TrimSpace(path)
		message := "Allow \"" + ext.Name + "\" to view your app settings?"
		if path != "" {
			detail = "Setting: \"" + path + "\""
			message = "Allow \"" + ext.Name + "\" to view \"" + path + "\"?"
		}

		return a.settingsAction(vm, scheduler, ext, prompt.Options{
			Kind:     "settings",
			Action:   "view \"" + detail + "\"",
			Resource: detail,
			Message:  message,
			Details:  []string{path},
			Cache:    cache,
			CacheKey: settingsCacheKey("view", path),
		}, func() (interface{}, error) {
			base, err := a.getSettingsMap()
			if err != nil {
				return nil, err
			}
			if path == "" {
				return base, nil
			}
			value, found := getPath(base, path)
			if !found && hasFallback {
				return fallback, nil
			}
			return value, nil
		})
	})

	_ = settingsObj.Set("set", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 || goja.IsUndefined(call.Argument(0)) || goja.IsNull(call.Argument(0)) {
			return rejectNow(vm, errors.New("settings value is required"))
		}

		if len(call.Arguments) > 1 && !goja.IsUndefined(call.Argument(1)) && !goja.IsNull(call.Argument(1)) {
			path := call.Argument(0).String()
			value := call.Argument(1).Export()
			if strings.TrimSpace(path) == "" {
				return rejectNow(vm, errors.New("settings path is empty"))
			}

			return a.settingsAction(vm, scheduler, ext, prompt.Options{
				Kind:       "settings",
				Action:     "edit \"" + path + "\"",
				Resource:   "Setting: \"" + path + "\"",
				Message:    "Allow \"" + ext.Name + "\" to edit \"" + path + "\"?",
				Details:    []string{path},
				AllowLabel: "Allow",
				DenyLabel:  "Don't Allow",
				Cache:      cache,
				CacheKey:   settingsCacheKey("edit", path),
			}, func() (interface{}, error) {
				bundle, base, err := a.getSettingsBundleAndMap()
				if err != nil {
					return nil, err
				}
				if err := setPath(base, path, value); err != nil {
					return nil, err
				}
				changes := changesForPath(path)
				next, err := mapToSettingsBundle(base, bundle, changes)
				if err != nil {
					return nil, err
				}
				return a.saveSettings(next, changes)
			})
		}

		currentBundle, currentMap, err := a.getSettingsBundleAndMap()
		if err != nil {
			return rejectNow(vm, err)
		}

		nextInput := make(map[string]interface{})
		if err := decodeValue(call.Argument(0), &nextInput); err != nil {
			return rejectNow(vm, err)
		}

		changes := changesForMap(nextInput)
		nextBundle, err := mapToSettingsBundle(nextInput, currentBundle, changes)
		if err != nil {
			return rejectNow(vm, err)
		}
		nextMap, err := settingsBundleToMap(nextBundle)
		if err != nil {
			return rejectNow(vm, err)
		}

		details := []string{"all settings"}
		details = diffAppSettingsPaths(currentMap, nextMap)
		if len(details) == 0 {
			details = []string{"no setting changes"}
		}

		return a.settingsAction(vm, scheduler, ext, prompt.Options{
			Kind:     "settings",
			Action:   "edit app settings",
			Resource: "App settings",
			Message:  "Allow \"" + ext.Name + "\" to edit these settings?",
			Details:  details,
			Cache:    cache,
			CacheKey: settingsCacheKey("edit", details...),
		}, func() (interface{}, error) {
			return a.saveSettings(nextBundle, changes)
		})
	})

	_ = settingsObj.Set("patch", func(patch map[string]interface{}) goja.Value {
		details := settingPaths(patch)
		if len(details) == 0 {
			details = []string{"app settings"}
		}

		return a.settingsAction(vm, scheduler, ext, prompt.Options{
			Kind:     "settings",
			Action:   "edit app settings",
			Resource: "App settings",
			Message:  "Allow \"" + ext.Name + "\" to edit these settings?",
			Details:  details,
			Cache:    cache,
			CacheKey: settingsCacheKey("edit", details...),
		}, func() (interface{}, error) {
			bundle, base, err := a.getSettingsBundleAndMap()
			if err != nil {
				return nil, err
			}

			merge(base, patch)
			changes := changesForMap(patch)
			next, err := mapToSettingsBundle(base, bundle, changes)
			if err != nil {
				return nil, err
			}
			return a.saveSettings(next, changes)
		})
	})

	return settingsObj
}

func settingsCacheKey(action string, parts ...string) string {
	key := strings.Join(parts, "|")
	if key == "" {
		key = "all"
	}
	return promptKey("settings", action, key)
}

func (a *AppContextImpl) BindAppSettingsToContextObj(vm *goja.Runtime, obj *goja.Object, logger *zerolog.Logger, ext *extension.Extension, scheduler *gojautil.Scheduler) {
	_ = logger
	_ = obj.Set("appSettings", a.bindSettingsObj(vm, ext, scheduler))
}

func (a *AppContextImpl) settingsAction(vm *goja.Runtime, scheduler *gojautil.Scheduler, ext *extension.Extension, opts prompt.Options, run func() (interface{}, error)) goja.Value {
	promise, resolve, reject := vm.NewPromise()

	go func() {
		ret, err := interface{}(nil), a.ask(ext, opts)
		if err == nil {
			ret, err = run()
		}

		scheduler.ScheduleAsync(func() error {
			if err != nil {
				reject(goja_bindings.NewErrorString(vm, err.Error()))
				return nil
			}
			resolve(vm.ToValue(ret))
			return nil
		})
	}()

	return vm.ToValue(promise)
}

func (a *AppContextImpl) saveSettings(bundle *appSettingsBundle, changes appSettingsChanges) (map[string]interface{}, error) {
	if bundle == nil {
		return nil, errors.New("settings is nil")
	}

	database, ok := a.database.Get()
	if !ok {
		return nil, errors.New("database not set")
	}

	now := time.Now()

	var (
		savedSettings      *models.Settings
		savedMediastream   *models.MediastreamSettings
		savedTorrentstream *models.TorrentstreamSettings
		savedDebrid        *models.DebridSettings
		err                error
	)

	if changes.base {
		if bundle.Settings == nil {
			return nil, errors.New("settings is nil")
		}
		bundle.Settings.BaseModel = models.BaseModel{ID: 1, UpdatedAt: now}
		savedSettings, err = database.UpsertSettings(bundle.Settings)
		if err != nil {
			return nil, err
		}
	}

	if changes.mediastream {
		if bundle.Mediastream == nil {
			bundle.Mediastream = &models.MediastreamSettings{}
		}
		bundle.Mediastream.BaseModel = models.BaseModel{ID: 1, UpdatedAt: now}
		savedMediastream, err = database.UpsertMediastreamSettings(bundle.Mediastream)
		if err != nil {
			return nil, err
		}
	}

	if changes.torrentstream {
		if bundle.Torrentstream == nil {
			bundle.Torrentstream = &models.TorrentstreamSettings{}
		}
		bundle.Torrentstream.BaseModel = models.BaseModel{ID: 1, UpdatedAt: now}
		savedTorrentstream, err = database.UpsertTorrentstreamSettings(bundle.Torrentstream)
		if err != nil {
			return nil, err
		}
	}

	if changes.debrid {
		if bundle.Debrid == nil {
			bundle.Debrid = &models.DebridSettings{}
		}
		bundle.Debrid.BaseModel = models.BaseModel{ID: 1, UpdatedAt: now}
		savedDebrid, err = database.UpsertDebridSettings(bundle.Debrid)
		if err != nil {
			return nil, err
		}
	}

	if savedSettings != nil && a.settings.OnSaved != nil {
		a.settings.OnSaved(savedSettings)
	}
	if savedMediastream != nil && a.settings.OnMediastreamSaved != nil {
		a.settings.OnMediastreamSaved(savedMediastream)
	}
	if savedTorrentstream != nil && a.settings.OnTorrentstreamSaved != nil {
		a.settings.OnTorrentstreamSaved(savedTorrentstream)
	}
	if savedDebrid != nil && a.settings.OnDebridSaved != nil {
		a.settings.OnDebridSaved(savedDebrid)
	}

	if !changes.hasChanges() {
		return settingsBundleToMap(bundle)
	}

	return a.getSettingsMap()
}

func (a *AppContextImpl) getSettingsBundle() (*appSettingsBundle, error) {
	database, ok := a.database.Get()
	if !ok {
		return nil, errors.New("database not set")
	}

	settings, err := database.GetSettings()
	if err != nil {
		return nil, err
	}

	ret := &appSettingsBundle{Settings: settings}
	if mediastream, found := database.GetMediastreamSettings(); found {
		ret.Mediastream = mediastream
	}
	if torrentstream, found := database.GetTorrentstreamSettings(); found {
		ret.Torrentstream = torrentstream
	}
	if debrid, found := database.GetDebridSettings(); found {
		ret.Debrid = debrid
	}

	return ret, nil
}

func (a *AppContextImpl) getSettingsBundleAndMap() (*appSettingsBundle, map[string]interface{}, error) {
	bundle, err := a.getSettingsBundle()
	if err != nil {
		return nil, nil, err
	}

	base, err := settingsBundleToMap(bundle)
	if err != nil {
		return nil, nil, err
	}

	return bundle, base, nil

}

func (a *AppContextImpl) getSettingsMap() (map[string]interface{}, error) {
	_, base, err := a.getSettingsBundleAndMap()
	if err != nil {
		return nil, err
	}
	return base, nil
}

func toMap(in interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var ret map[string]interface{}
	if err := json.Unmarshal(bytes, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func merge(dst map[string]interface{}, src map[string]interface{}) {
	for key, value := range src {
		srcMap, srcOk := value.(map[string]interface{})
		dstMap, dstOk := dst[key].(map[string]interface{})
		if srcOk && dstOk {
			merge(dstMap, srcMap)
			continue
		}
		dst[key] = value
	}
}

func getPath(settings map[string]interface{}, path string) (interface{}, bool) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return settings, true
	}

	var curr interface{} = settings
	for _, part := range parts {
		currMap, ok := curr.(map[string]interface{})
		if !ok {
			return nil, false
		}
		curr, ok = currMap[part]
		if !ok {
			return nil, false
		}
	}

	return curr, true
}

func setPath(settings map[string]interface{}, path string, value interface{}) error {
	parts := splitPath(path)
	if len(parts) == 0 {
		return errors.New("settings path is empty")
	}

	curr := settings
	for _, part := range parts[:len(parts)-1] {
		next, ok := curr[part].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			curr[part] = next
		}
		curr = next
	}
	curr[parts[len(parts)-1]] = value
	return nil
}

func mapToSettings(in map[string]interface{}) (*models.Settings, error) {
	bytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	var ret models.Settings
	if err := json.Unmarshal(bytes, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func settingsBundleToMap(bundle *appSettingsBundle) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	if bundle != nil && bundle.Settings != nil {
		base, err := toMap(bundle.Settings)
		if err != nil {
			return nil, err
		}
		if base != nil {
			ret = base
		}
	}

	if bundle == nil {
		return ret, nil
	}

	if err := setSettingsSection(ret, appSettingsMediastreamRoot, bundle.Mediastream); err != nil {
		return nil, err
	}
	if err := setSettingsSection(ret, appSettingsTorrentstreamRoot, bundle.Torrentstream); err != nil {
		return nil, err
	}
	if err := setSettingsSection(ret, appSettingsDebridRoot, bundle.Debrid); err != nil {
		return nil, err
	}

	return ret, nil
}

func setSettingsSection(dst map[string]interface{}, key string, value interface{}) error {
	if value == nil {
		return nil
	}

	section, err := toMap(value)
	if err != nil {
		return err
	}
	stripSettingMetaKeys(section)
	dst[key] = section
	return nil
}

func stripSettingMetaKeys(in map[string]interface{}) {
	delete(in, "id")
	delete(in, "createdAt")
	delete(in, "updatedAt")
}

func mapToSettingsBundle(in map[string]interface{}, prev *appSettingsBundle, changes appSettingsChanges) (*appSettingsBundle, error) {
	base := make(map[string]interface{}, len(in))
	for key, value := range in {
		switch key {
		case appSettingsMediastreamRoot, appSettingsTorrentstreamRoot, appSettingsDebridRoot:
			continue
		default:
			base[key] = value
		}
	}

	settings, err := mapToSettings(base)
	if err != nil {
		return nil, err
	}

	var prevMediastream *models.MediastreamSettings
	var prevTorrentstream *models.TorrentstreamSettings
	var prevDebrid *models.DebridSettings
	if prev != nil {
		prevMediastream = prev.Mediastream
		prevTorrentstream = prev.Torrentstream
		prevDebrid = prev.Debrid
	}

	mediastream, err := decodeSettingsSection[models.MediastreamSettings](appSettingsMediastreamRoot, in[appSettingsMediastreamRoot], prevMediastream, changes.mediastream)
	if err != nil {
		return nil, err
	}
	torrentstream, err := decodeSettingsSection[models.TorrentstreamSettings](appSettingsTorrentstreamRoot, in[appSettingsTorrentstreamRoot], prevTorrentstream, changes.torrentstream)
	if err != nil {
		return nil, err
	}
	debrid, err := decodeSettingsSection[models.DebridSettings](appSettingsDebridRoot, in[appSettingsDebridRoot], prevDebrid, changes.debrid)
	if err != nil {
		return nil, err
	}

	return &appSettingsBundle{
		Settings:      settings,
		Mediastream:   mediastream,
		Torrentstream: torrentstream,
		Debrid:        debrid,
	}, nil
}

func decodeSettingsSection[T any](key string, raw interface{}, prev *T, changed bool) (*T, error) {
	if !changed {
		return prev, nil
	}
	if raw == nil {
		return new(T), nil
	}

	section, ok := raw.(map[string]interface{})
	if !ok {
		return nil, errors.New("settings section \"" + key + "\" should be an object")
	}

	bytes, err := json.Marshal(section)
	if err != nil {
		return nil, err
	}

	ret := new(T)
	if err := json.Unmarshal(bytes, ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func decodeValue(value goja.Value, out interface{}) error {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return errors.New("value is empty")
	}

	bytes, err := json.Marshal(value.Export())
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, out)
}

func rejectNow(vm *goja.Runtime, err error) goja.Value {
	promise, _, reject := vm.NewPromise()
	reject(goja_bindings.NewErrorString(vm, err.Error()))
	return vm.ToValue(promise)
}

func settingPaths(patch map[string]interface{}) []string {
	ret := make([]string, 0)
	collectSettingPaths("", patch, &ret)
	sort.Strings(ret)
	return ret
}

func collectSettingPaths(prefix string, in map[string]interface{}, ret *[]string) {
	for key, value := range in {
		if prefix == "" && isSettingMetaKey(key) {
			continue
		}

		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		valueMap, ok := value.(map[string]interface{})
		if ok && len(valueMap) > 0 {
			collectSettingPaths(path, valueMap, ret)
			continue
		}

		*ret = append(*ret, path)
	}
}

func diffAppSettingsPaths(prev map[string]interface{}, next map[string]interface{}) []string {
	ret := make([]string, 0)
	diffMapPaths("", prev, next, &ret)
	sort.Strings(ret)
	return ret
}

func diffMapPaths(prefix string, prev map[string]interface{}, next map[string]interface{}, ret *[]string) {
	seen := make(map[string]struct{}, len(prev)+len(next))
	for key := range prev {
		seen[key] = struct{}{}
	}
	for key := range next {
		seen[key] = struct{}{}
	}

	for key := range seen {
		if prefix == "" && isSettingMetaKey(key) {
			continue
		}

		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		prevVal, prevFound := prev[key]
		nextVal, nextFound := next[key]
		prevMap, prevIsMap := prevVal.(map[string]interface{})
		nextMap, nextIsMap := nextVal.(map[string]interface{})
		if prevFound && nextFound && prevIsMap && nextIsMap {
			diffMapPaths(path, prevMap, nextMap, ret)
			continue
		}

		if !prevFound || !nextFound || !reflect.DeepEqual(prevVal, nextVal) {
			*ret = append(*ret, path)
		}
	}
}

func isSettingMetaKey(key string) bool {
	switch key {
	case "id", "createdAt", "updatedAt":
		return true
	default:
		return false
	}
}

func splitPath(path string) []string {
	parts := strings.Split(path, ".")
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ret = append(ret, part)
		}
	}
	return ret
}
