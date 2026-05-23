package plugin

import (
	databasepkg "seanime/internal/database/db"
	"seanime/internal/database/models"
	"seanime/internal/events"
	"seanime/internal/extension"
	"seanime/internal/extension_repo/prompt"
	"seanime/internal/testutil"
	"seanime/internal/util"
	gojautil "seanime/internal/util/goja"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func newAppSettingsTestContext(t *testing.T, actions SettingsActions) (*AppContextImpl, *databasepkg.Database, *events.MockWSEventManager) {
	t.Helper()

	databasepkg.CurrSettings = nil
	databasepkg.CurrMediastreamSettings = nil
	databasepkg.CurrTorrentstreamSettings = nil
	databasepkg.CurrentDebridSettings = nil
	t.Cleanup(func() {
		databasepkg.CurrSettings = nil
		databasepkg.CurrMediastreamSettings = nil
		databasepkg.CurrTorrentstreamSettings = nil
		databasepkg.CurrentDebridSettings = nil
	})

	logger := util.NewLogger()
	env := testutil.NewTestEnv(t)
	database := env.MustNewDatabase(logger)
	ws := events.NewMockWSEventManager(logger)
	manager := prompt.NewManager(&prompt.NewManagerOptions{Logger: logger, WSEventManager: ws})
	appCtx := NewAppContext().(*AppContextImpl)
	appCtx.SetModulesPartial(AppContextModules{
		Database:      database,
		PromptManager: manager,
		Settings:      actions,
	})

	return appCtx, database, ws
}

func seedAppSettings(t *testing.T, database *databasepkg.Database) {
	t.Helper()

	_, err := database.UpsertSettings(&models.Settings{
		BaseModel: models.BaseModel{ID: 1},
		Library: &models.LibrarySettings{
			LibraryPath: "/library",
		},
		MediaPlayer:   &models.MediaPlayerSettings{},
		Torrent:       &models.TorrentSettings{},
		Anilist:       &models.AnilistSettings{},
		ListSync:      &models.ListSyncSettings{},
		Discord:       &models.DiscordSettings{},
		Manga:         &models.MangaSettings{},
		Notifications: &models.NotificationSettings{},
		Nakama:        &models.NakamaSettings{},
	})
	require.NoError(t, err)

	_, err = database.UpsertMediastreamSettings(&models.MediastreamSettings{
		BaseModel:        models.BaseModel{ID: 1},
		TranscodeEnabled: true,
		TranscodePreset:  "fast",
	})
	require.NoError(t, err)

	_, err = database.UpsertTorrentstreamSettings(&models.TorrentstreamSettings{
		BaseModel:        models.BaseModel{ID: 1},
		Enabled:          true,
		StreamUrlAddress: "http://127.0.0.1:43214",
	})
	require.NoError(t, err)

	_, err = database.UpsertDebridSettings(&models.DebridSettings{
		BaseModel: models.BaseModel{ID: 1},
		Enabled:   true,
		Provider:  "realdebrid",
	})
	require.NoError(t, err)
}

func bindTestAppSettings(t *testing.T, appCtx *AppContextImpl) (*goja.Runtime, *goja.Object, *gojautil.Scheduler) {
	t.Helper()

	vm := goja.New()
	obj := vm.NewObject()
	scheduler := gojautil.NewScheduler()
	logger := util.NewLogger()

	appCtx.BindAppSettingsToContextObj(vm, obj, logger, &extension.Extension{ID: t.Name(), Name: t.Name()}, scheduler)

	t.Cleanup(scheduler.Stop)

	return vm, obj, scheduler
}

func waitForSettingsPromptRequest(t *testing.T, ws *events.MockWSEventManager, index int) prompt.Request {
	t.Helper()

	var request prompt.Request
	require.Eventually(t, func() bool {
		events := ws.Events()
		if len(events) <= index {
			return false
		}
		if events[index].Type != prompt.EventRequest {
			return false
		}
		payload, ok := events[index].Payload.(prompt.Request)
		if !ok {
			return false
		}
		request = payload
		return true
	}, time.Second, 10*time.Millisecond)

	return request
}

func allowSettingsPrompt(ws *events.MockWSEventManager, id string) {
	ws.MockSendClientEvent(&events.WebsocketClientEvent{
		ClientID: "client-1",
		Type:     events.WebsocketClientEventType(prompt.EventResponse),
		Payload:  prompt.Response{ID: id, Allowed: true},
	})
}

func requirePromiseFulfilled(t *testing.T, value goja.Value) *goja.Promise {
	t.Helper()

	promise, ok := value.Export().(*goja.Promise)
	require.True(t, ok, "value should export to a promise")
	require.Eventually(t, func() bool {
		return promise.State() == goja.PromiseStateFulfilled
	}, time.Second, 10*time.Millisecond)

	return promise
}

func TestAppSettingsGetIncludesSecondaryRoots(t *testing.T) {
	appCtx, database, ws := newAppSettingsTestContext(t, SettingsActions{})
	seedAppSettings(t, database)

	vm, obj, _ := bindTestAppSettings(t, appCtx)
	settingsObj := obj.Get("appSettings").ToObject(vm)
	get, ok := goja.AssertFunction(settingsObj.Get("get"))
	require.True(t, ok)

	ret, err := get(settingsObj)
	require.NoError(t, err)

	request := waitForSettingsPromptRequest(t, ws, 0)
	allowSettingsPrompt(ws, request.ID)
	promise := requirePromiseFulfilled(t, ret)

	var settings map[string]interface{}
	require.NoError(t, vm.ExportTo(promise.Result(), &settings))

	mediastream, ok := settings[appSettingsMediastreamRoot].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, mediastream["transcodeEnabled"])
	_, hasMediastreamID := mediastream["id"]
	require.False(t, hasMediastreamID)

	torrentstream, ok := settings[appSettingsTorrentstreamRoot].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, true, torrentstream["enabled"])

	debrid, ok := settings[appSettingsDebridRoot].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "realdebrid", debrid["provider"])
	_, hasDebridID := debrid["id"]
	require.False(t, hasDebridID)
}

func TestAppSettingsSetPathSavesSecondaryRoot(t *testing.T) {
	baseSaved := 0
	mediastreamSaved := 0
	appCtx, database, ws := newAppSettingsTestContext(t, SettingsActions{
		OnSaved: func(_ *models.Settings) {
			baseSaved++
		},
		OnMediastreamSaved: func(_ *models.MediastreamSettings) {
			mediastreamSaved++
		},
	})
	seedAppSettings(t, database)

	vm, obj, _ := bindTestAppSettings(t, appCtx)
	settingsObj := obj.Get("appSettings").ToObject(vm)
	set, ok := goja.AssertFunction(settingsObj.Get("set"))
	require.True(t, ok)

	ret, err := set(settingsObj, vm.ToValue("mediastream.transcodeEnabled"), vm.ToValue(false))
	require.NoError(t, err)

	request := waitForSettingsPromptRequest(t, ws, 0)
	allowSettingsPrompt(ws, request.ID)
	_ = requirePromiseFulfilled(t, ret)

	settings, found := database.GetMediastreamSettings()
	require.True(t, found)
	require.False(t, settings.TranscodeEnabled)
	require.Equal(t, 0, baseSaved)
	require.Equal(t, 1, mediastreamSaved)
}

func TestAppSettingsPatchSavesSecondaryRoots(t *testing.T) {
	baseSaved := 0
	torrentstreamSaved := 0
	debridSaved := 0
	appCtx, database, ws := newAppSettingsTestContext(t, SettingsActions{
		OnSaved: func(_ *models.Settings) {
			baseSaved++
		},
		OnTorrentstreamSaved: func(_ *models.TorrentstreamSettings) {
			torrentstreamSaved++
		},
		OnDebridSaved: func(_ *models.DebridSettings) {
			debridSaved++
		},
	})
	seedAppSettings(t, database)

	vm, obj, _ := bindTestAppSettings(t, appCtx)
	settingsObj := obj.Get("appSettings").ToObject(vm)
	patch, ok := goja.AssertFunction(settingsObj.Get("patch"))
	require.True(t, ok)

	ret, err := patch(settingsObj, vm.ToValue(map[string]interface{}{
		appSettingsTorrentstreamRoot: map[string]interface{}{
			"enabled":          false,
			"streamUrlAddress": "http://localhost:9999",
		},
		appSettingsDebridRoot: map[string]interface{}{
			"provider": "torbox",
		},
	}))
	require.NoError(t, err)

	request := waitForSettingsPromptRequest(t, ws, 0)
	allowSettingsPrompt(ws, request.ID)
	_ = requirePromiseFulfilled(t, ret)

	torrentstream, found := database.GetTorrentstreamSettings()
	require.True(t, found)
	require.False(t, torrentstream.Enabled)
	require.Equal(t, "http://localhost:9999", torrentstream.StreamUrlAddress)

	debrid, found := database.GetDebridSettings()
	require.True(t, found)
	require.Equal(t, "torbox", debrid.Provider)
	require.True(t, debrid.Enabled)

	require.Equal(t, 0, baseSaved)
	require.Equal(t, 1, torrentstreamSaved)
	require.Equal(t, 1, debridSaved)
}
