package extension_repo

import (
	"os"
	"seanime/internal/events"
	"seanime/internal/extension"
	hibikemanga "seanime/internal/extension/hibike/manga"
	hibikeonlinestream "seanime/internal/extension/hibike/onlinestream"
	hibiketorrent "seanime/internal/extension/hibike/torrent"
	"seanime/internal/goja/goja_runtime"
	"seanime/internal/hook"
	"seanime/internal/plugin"
	"seanime/internal/util/filecache"
	"seanime/internal/util/result"

	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type (
	// Repository manages all extensions
	Repository struct {
		logger         *zerolog.Logger
		fileCacher     *filecache.Cacher
		wsEventManager events.WSEventManagerInterface
		// Absolute path to the directory containing all extensions
		extensionDir string
		// Store all active Goja VMs
		// - When reloading extensions, all VMs are interrupted
		gojaExtensions *result.Map[string, GojaExtension]

		gojaRuntimeManager *goja_runtime.Manager
		// Extension bank
		// - When reloading extensions, external extensions are removed & re-added
		extensionBank *extension.UnifiedBank

		invalidExtensions *result.Map[string, *extension.InvalidExtension]

		hookManager hook.Manager

		appContext *plugin.AppContext
	}

	AllExtensions struct {
		Extensions        []*extension.Extension        `json:"extensions"`
		InvalidExtensions []*extension.InvalidExtension `json:"invalidExtensions"`
		// List of extensions with invalid user config extensions, these extensions are still loaded
		InvalidUserConfigExtensions []*extension.InvalidExtension `json:"invalidUserConfigExtensions"`
		// List of extension IDs that have an update available
		// This is only populated when the user clicks on "Check for updates"
		HasUpdate []UpdateData `json:"hasUpdate"`
	}

	UpdateData struct {
		ExtensionID string `json:"extensionID"`
		ManifestURI string `json:"manifestURI"`
		Version     string `json:"version"`
	}

	MangaProviderExtensionItem struct {
		ID       string               `json:"id"`
		Name     string               `json:"name"`
		Lang     string               `json:"lang"` // ISO 639-1 language code
		Settings hibikemanga.Settings `json:"settings"`
	}

	OnlinestreamProviderExtensionItem struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		Lang           string   `json:"lang"` // ISO 639-1 language code
		EpisodeServers []string `json:"episodeServers"`
		SupportsDub    bool     `json:"supportsDub"`
	}

	AnimeTorrentProviderExtensionItem struct {
		ID       string                              `json:"id"`
		Name     string                              `json:"name"`
		Lang     string                              `json:"lang"` // ISO 639-1 language code
		Settings hibiketorrent.AnimeProviderSettings `json:"settings"`
	}

	TrayPluginExtensionItem struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		IsPinned bool   `json:"isPinned"`
	}
)

type NewRepositoryOptions struct {
	Logger         *zerolog.Logger
	ExtensionDir   string
	WSEventManager events.WSEventManagerInterface
	FileCacher     *filecache.Cacher
	HookManager    hook.Manager
}

func NewRepository(opts *NewRepositoryOptions) *Repository {

	// Make sure the extension directory exists
	_ = os.MkdirAll(opts.ExtensionDir, os.ModePerm)

	ret := &Repository{
		logger:             opts.Logger,
		extensionDir:       opts.ExtensionDir,
		wsEventManager:     opts.WSEventManager,
		gojaExtensions:     result.NewResultMap[string, GojaExtension](),
		gojaRuntimeManager: goja_runtime.NewManager(opts.Logger, 20),
		extensionBank:      extension.NewUnifiedBank(),
		invalidExtensions:  result.NewResultMap[string, *extension.InvalidExtension](),
		fileCacher:         opts.FileCacher,
		hookManager:        opts.HookManager,
	}

	clientEventSubscriber := ret.wsEventManager.SubscribeToClientEvents("extension-repository")

	go func() {
		for event := range clientEventSubscriber.Channel {
			switch event.Type {
			case "tray:list":
				ret.wsEventManager.SendEvent("tray:list", ret.ListTrayPluginExtensions())
			}
		}
	}()

	return ret
}

func (r *Repository) GetAllExtensions(withUpdates bool) (ret *AllExtensions) {
	invalidExtensions := r.ListInvalidExtensions()

	fatalInvalidExtensions := lo.Filter(invalidExtensions, func(ext *extension.InvalidExtension, _ int) bool {
		return ext.Code != extension.InvalidExtensionUserConfigError
	})

	userConfigInvalidExtensions := lo.Filter(invalidExtensions, func(ext *extension.InvalidExtension, _ int) bool {
		return ext.Code == extension.InvalidExtensionUserConfigError
	})

	ret = &AllExtensions{
		Extensions:                  r.ListExtensionData(),
		InvalidExtensions:           fatalInvalidExtensions,
		InvalidUserConfigExtensions: userConfigInvalidExtensions,
	}
	if withUpdates {
		ret.HasUpdate = r.checkForUpdates()
	}
	return
}

func (r *Repository) ListExtensionData() (ret []*extension.Extension) {
	r.extensionBank.Range(func(key string, ext extension.BaseExtension) bool {
		ret = append(ret, extension.ToExtensionData(ext))
		return true
	})

	return ret
}

func (r *Repository) ListInvalidExtensions() (ret []*extension.InvalidExtension) {
	r.invalidExtensions.Range(func(key string, ext *extension.InvalidExtension) bool {
		//ext.Extension.Payload = "" // Remove the payload so the client knows the extension is installed
		ret = append(ret, ext)
		return true
	})

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Lists
// - Lists are used to display available options to the user based on the extensions installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) ListTrayPluginExtensions() []*TrayPluginExtensionItem {
	ret := make([]*TrayPluginExtensionItem, 0)

	// TODO: Figure out which plugin is a tray plugin and if it's pinned
	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.PluginExtension) bool {
		ret = append(ret, &TrayPluginExtensionItem{
			ID:       ext.GetID(),
			Name:     ext.GetName(),
			Icon:     ext.GetIcon(),
			IsPinned: true,
		})
		return true
	})

	return ret
}

func (r *Repository) ListMangaProviderExtensions() []*MangaProviderExtensionItem {
	ret := make([]*MangaProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.MangaProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &MangaProviderExtensionItem{
			ID:       ext.GetID(),
			Name:     ext.GetName(),
			Lang:     extension.GetExtensionLang(ext.GetLang()),
			Settings: settings,
		})
		return true
	})

	return ret
}

func (r *Repository) ListOnlinestreamProviderExtensions() []*OnlinestreamProviderExtensionItem {
	ret := make([]*OnlinestreamProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.OnlinestreamProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &OnlinestreamProviderExtensionItem{
			ID:             ext.GetID(),
			Name:           ext.GetName(),
			Lang:           extension.GetExtensionLang(ext.GetLang()),
			EpisodeServers: settings.EpisodeServers,
			SupportsDub:    settings.SupportsDub,
		})
		return true
	})

	return ret
}

func (r *Repository) ListAnimeTorrentProviderExtensions() []*AnimeTorrentProviderExtensionItem {
	ret := make([]*AnimeTorrentProviderExtensionItem, 0)

	extension.RangeExtensions(r.extensionBank, func(key string, ext extension.AnimeTorrentProviderExtension) bool {
		settings := ext.GetProvider().GetSettings()
		ret = append(ret, &AnimeTorrentProviderExtensionItem{
			ID:   ext.GetID(),
			Name: ext.GetName(),
			Lang: extension.GetExtensionLang(ext.GetLang()),
			Settings: hibiketorrent.AnimeProviderSettings{
				Type:           settings.Type,
				CanSmartSearch: settings.CanSmartSearch,
				SupportsAdult:  settings.SupportsAdult,
				SmartSearchFilters: lo.Map(settings.SmartSearchFilters, func(value hibiketorrent.AnimeProviderSmartSearchFilter, _ int) hibiketorrent.AnimeProviderSmartSearchFilter {
					return value
				}),
			},
		})

		return true
	})

	return ret
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetLoadedExtension returns the loaded extension by ID.
// It returns an extension.BaseExtension interface, so it can be used to get the extension's details.
func (r *Repository) GetLoadedExtension(id string) (extension.BaseExtension, bool) {
	var ext extension.BaseExtension
	ext, found := r.extensionBank.Get(id)
	if found {
		return ext, true
	}

	return nil, false
}

func (r *Repository) GetExtensionBank() *extension.UnifiedBank {
	return r.extensionBank
}

func (r *Repository) GetMangaProviderExtensionByID(id string) (extension.MangaProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.MangaProviderExtension](r.extensionBank, id)
	return ext, found
}

func (r *Repository) GetOnlinestreamProviderExtensionByID(id string) (extension.OnlinestreamProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.OnlinestreamProviderExtension](r.extensionBank, id)
	return ext, found
}

func (r *Repository) GetAnimeTorrentProviderExtensionByID(id string) (extension.AnimeTorrentProviderExtension, bool) {
	ext, found := extension.GetExtension[extension.AnimeTorrentProviderExtension](r.extensionBank, id)
	return ext, found
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Built-in extensions
// - Built-in extensions are loaded once, on application startup
// - The "manifestURI" field is set to "builtin" to indicate that the extension is not external
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) LoadBuiltInMangaProviderExtension(info extension.Extension, provider hibikemanga.Provider) {
	r.extensionBank.Set(info.ID, extension.NewMangaProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in manga provider extension")
}

func (r *Repository) LoadBuiltInAnimeTorrentProviderExtension(info extension.Extension, provider hibiketorrent.AnimeProvider) {
	r.extensionBank.Set(info.ID, extension.NewAnimeTorrentProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in anime torrent provider extension")
}

func (r *Repository) LoadBuiltInOnlinestreamProviderExtension(info extension.Extension, provider hibikeonlinestream.Provider) {
	r.extensionBank.Set(info.ID, extension.NewOnlinestreamProviderExtension(&info, provider))
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

func (r *Repository) LoadBuiltInOnlinestreamProviderExtensionJS(info extension.Extension) {
	err := r.loadExternalOnlinestreamExtensionJS(&info, info.Language)
	if err != nil {
		r.logger.Error().Err(err).Str("id", info.ID).Msg("extensions: Failed to load built-in JS onlinestream provider extension")
		return
	}
	r.logger.Debug().Str("id", info.ID).Msg("extensions: Loaded built-in onlinestream provider extension")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (r *Repository) loadPlugins() {

	testExt := &extension.Extension{
		ID:       "test-plugin",
		Language: extension.LanguageTypescript,
		Plugin: &extension.PluginManifest{
			Permissions: []extension.PluginPermission{extension.PluginPermissionStorage, extension.PluginPermissionAnilist, extension.PluginPermissionOS, extension.PluginPermissionDatabase},
		},
		Payload: `
			function init() {
				$ui.register((ctx) => {
					const tray = ctx.newTray();

					const currentMediaId = ctx.state(0);
					const storageBackgroundImage = ctx.state("");
					const mediaIds = ctx.state([]);
					
					const customBannerImageRef = ctx.registerFieldRef("customBannerImageRef");
					
					const fetchBackgroundImage = () => {
						const backgroundImage = $storage.get('backgroundImages.' + currentMediaId.get());
						if (backgroundImage) {
							storageBackgroundImage.set(backgroundImage);
							customBannerImageRef.setValue(backgroundImage);
						} else {
							storageBackgroundImage.set("");
							customBannerImageRef.setValue("");
						}
					}

					ctx.effect(() => {
						console.log("media ID changed, fetching background image and updating tray");
						fetchBackgroundImage();

						console.log("updating tray");
					}, [currentMediaId]);


					fetchBackgroundImage()

					ctx.screen.onNavigate((e) => {
						console.log("screen navigated", e);
						if (e.pathname === "/entry" && !!e.query) {
							const id = parseInt(e.query.replace("?id=", ""));
							currentMediaId.set(id);
						} else {
							currentMediaId.set(0);
						}

						console.log("updating tray");
					});

					ctx.registerEventHandler("saveBackgroundImage", () => {
						ctx.toast.info("Setting background image to " + customBannerImageRef.current);
						$storage.set('backgroundImages.' + currentMediaId.get(), customBannerImageRef.current);
						ctx.toast.success("Background image saved");
						fetchBackgroundImage();
						$anilist.refreshAnimeCollection();
					});

					// $store.watch("mediaIds", (mId) => {
					// 	mediaIds.set(p => [...p, mId]);
					// });

					ctx.registerEventHandler("button-clicked", () => {
						console.log("button-clicked");
						console.log("navigating to /entry?id=21");
						try {
							ctx.screen.navigateTo("/entry?id=21");
						} catch (e) {
							console.error("navigate error", e);
						}
						ctx.setTimeout(() => {
							try {
								console.log("navigating to /entry?id=177709");
								ctx.screen.navigateTo("/entry?id=177709");
							} catch (e) {
								console.error("navigate error", e);
							}
						}, 1000);
						ctx.setTimeout(() => {
							try {
								console.log("opening https://google.com");
								const cmd = $os.cmd("open", "https://google.com");
								cmd.run();
							} catch (e) {
								console.error("open error", e);
							}
						}, 2000);
					});

					tray.render(() => {
						return tray.stack({
							items: [
								tray.button("Click me", { onClick: "button-clicked" }),
								currentMediaId.get() === 0 ? tray.text("Open an anime or manga") : tray.stack({
									items: [
										tray.text("Current media ID: " + currentMediaId.get()),
										tray.input({ fieldRef: "customBannerImageRef", value: storageBackgroundImage.get() }),
										tray.button({ label: "Save", onClick: "saveBackgroundImage" }),
									],
								}),
							],
						});
					});
				})

				$app.onGetAnime((e) => {
					$store.set("mediaIds", e.anime.id);
					e.next();
				});

				// $app.onAnimeEntry((e) => {
				// 	const mediaId = e.entry.mediaId;
				// 	const bannerImage = $storage.get('backgroundImages')[mediaId.toString()] || "";
				// 	console.log("onAnimeEntry", e.entry.mediaId, bannerImage);
				// 	console.log("storage", $storage.get('backgroundImages'));
				// 	if (!!bannerImage) {
				// 		$replace(e.entry.media.bannerImage, bannerImage);
				// 		console.log("replaced banner image", e.entry.media.bannerImage);
				// 	}
					
				// });

				$app.onGetAnimeCollection((e) => {
					const bannerImages = $storage.get('backgroundImages');
					for (let i = 0; i < e.animeCollection.mediaListCollection.lists.length; i++) {
						for (let j = 0; j < e.animeCollection.mediaListCollection.lists[i].entries.length; j++) {
							const mediaId = e.animeCollection.mediaListCollection.lists[i].entries[j].media.id;
							const bannerImage = bannerImages[mediaId.toString()] || "";
							if (!!bannerImage) {
								$replace(e.animeCollection.mediaListCollection.lists[i].entries[j].media.bannerImage, bannerImage);
							}
						}
					}
					e.next();
				});
				
				$app.onGetRawAnimeCollection((e) => {
					const bannerImages = $storage.get('backgroundImages');
					//console.log("onGetRawAnimeCollection fired", e.animeCollection.mediaListCollection)
					for (let i = 0; i < e.animeCollection.mediaListCollection.lists.length; i++) {
						for (let j = 0; j < e.animeCollection.mediaListCollection.lists[i].entries.length; j++) {
							const mediaId = e.animeCollection.mediaListCollection.lists[i].entries[j].media.id;
							const bannerImage = bannerImages[mediaId.toString()] || "";
							if (!!bannerImage) {
								$replace(e.animeCollection.mediaListCollection.lists[i].entries[j].media.bannerImage, bannerImage);
							}
						}
					}
					e.next();
				});

				// $app.onAnimeEntryLibraryDataRequest((e) => {
				// 	$replace(e.options.entryLocalFiles, [])
				// 	e.next();
				// });
				// $app.OnAnimeEntryRequest((e) => {
				// 	e.mediaId = 21;
				// 	e.next();
				// });
				
			}
			`,
		Name:        "Test Plugin",
		Version:     "1.0.0",
		ManifestURI: "https://raw.githubusercontent.com/5rahim/seanime-extensions/refs/heads/main/anime-torrent-providers/basic-nyaa/basic-nyaa.json",
		Type:        extension.TypePlugin,
		Description: "Test Plugin",
		Author:      "Test Author",
		PayloadURI:  "",
	}

	err := r.loadPluginExtension(testExt)
	if err != nil {
		r.logger.Error().Err(err).Msg("extensions: Failed to load test extension")
	}

}
