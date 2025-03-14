/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/hooks.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />

// @ts-ignore
function init() {
    $ui.register((ctx) => {
        const tray = ctx.newTray({
            tooltipText: "Test Plugin",
            iconUrl: "https://raw.githubusercontent.com/5rahim/hibike/main/icons/seadex.png",
            withContent: true,
        });

        const currentMediaId = ctx.state(0);
        const storageBackgroundImage = ctx.state("");
        const mediaIds = ctx.state([]);

        const customBannerImageRef = ctx.registerFieldRef("customBannerImageRef");

        ctx.screen.loadCurrent()
        tray.updateBadge({ number: 0 })

        const fetchBackgroundImage = () => {
            const backgroundImage = $storage.get<string>("backgroundImages." + currentMediaId.get())
            console.log("backgroundImage", backgroundImage)
            if (backgroundImage) {
                storageBackgroundImage.set(backgroundImage);
                customBannerImageRef.setValue(backgroundImage);
                tray.updateBadge({ number: 1, intent: "info" })
            } else {
                storageBackgroundImage.set("");
                customBannerImageRef.setValue("");
                tray.updateBadge({ number: 0 })
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
                fetchBackgroundImage()
                // tray.open();
            } else {
                currentMediaId.set(0);
                tray.close()
            }

            console.log("updating tray");
        });

        ctx.registerEventHandler("saveBackgroundImage", () => {
            ctx.toast.info("Setting background image to " + customBannerImageRef.current);
            $storage.set('backgroundImages.' + currentMediaId.get(), customBannerImageRef.current);
            ctx.toast.success("Background image saved " + customBannerImageRef.current)
            fetchBackgroundImage();
            $anilist.refreshAnimeCollection();
        });

        // $store.watch("mediaIds", (mId) => {
        // 	mediaIds.set(p => [...p, mId]);
        // });

        ctx.registerEventHandler("button-clicked", () => {
            const previous = $database.localFiles.getAll()
            $database.localFiles.insert([{
                path: "/Volumes/Seagate Portable Drive/ANIME/[SubsPlease] Bocchi the Rock! (01-12) (1080p) [Batch]/[SubsPlease] Bocchi the Rock! - 01v2 (1080p) [ABDDAE16].mkv",
                name: "[SubsPlease] Bocchi the Rock! - 01v2 (1080p) [ABDDAE16].mkv",
                locked: true,
                ignored: false,
                mediaId: 130003,
                metadata: {
                    episode: 1,
                    aniDBEpisode: "1",
                    type: "main",
                },
            }])
            ctx.toast.info("Inserted new local file")

            ctx.setTimeout(() => {
                $database.localFiles.insert(previous)
                ctx.toast.info("Inserted previous local files")
            }, 3000);
        })

        // ctx.registerEventHandler("button-clicked", () => {
        //     console.log("button-clicked");
        //     console.log("navigating to /entry?id=21");
        //     try {
        //         ctx.screen.navigateTo("/entry?id=21");
        //     } catch (e) {
        //         console.error("navigate error", e);
        //     }
        //     ctx.setTimeout(() => {
        //         try {
        //             console.log("navigating to /entry?id=177709");
        //             ctx.screen.navigateTo("/entry?id=177709");
        //         } catch (e) {
        //             console.error("navigate error", e);
        //         }
        //     }, 1000);
        //     ctx.setTimeout(() => {
        //         try {
        //             console.log("opening https://google.com");
        //             const cmd = $os.cmd("open", "https://google.com");
        //             cmd.run();
        //         } catch (e) {
        //             console.error("open error", e);
        //         }
        //     }, 2000);
        // });

        tray.render(() => {
            return tray.stack({
                items: [
                    tray.button("Click me", {onClick: "button-clicked"}),
                    currentMediaId.get() === 0 ? tray.text("Open an anime or manga") : tray.stack({
                        items: [
                            tray.text(`Current media ID: ${currentMediaId.get()}`),
                            tray.input({fieldRef: "customBannerImageRef", value: storageBackgroundImage.get()}),
                            tray.button({label: "Save", onClick: "saveBackgroundImage"}),
                        ],
                    }),
                ],
            });
        });
    })

    $app.onGetAnime((e) => {
        $store.set("mediaIds", e.anime?.id)
        e.next();
    });


    $app.onGetAnimeCollection((e) => {
        console.log("onGetAnimeCollection called")
        const bannerImages = $storage.get('backgroundImages');
        console.log("onGetAnimeCollection bannerImages", bannerImages)
        if (!bannerImages) {
            e.next();
            return;
        }
        if (!!e.animeCollection?.mediaListCollection?.lists?.length) {
            for (let i = 0; i < e.animeCollection?.mediaListCollection?.lists?.length; i++) {
                for (let j = 0; j < e.animeCollection.mediaListCollection.lists[i].entries!.length; j++) {
                    const mediaId = e.animeCollection!.mediaListCollection!.lists[i]!.entries![j]!.media!.id
                    const bannerImage = bannerImages[mediaId.toString()] || ""
                    if (!!bannerImage) {
                        $replace(e.animeCollection!.mediaListCollection!.lists[i]!.entries![j]!.media!.bannerImage, bannerImage)
                    }
                }
            }
        }
        e.next();
    });

    $app.onGetRawAnimeCollection((e) => {
        const bannerImages = $storage.get<Record<string, string>>("backgroundImages")
        if (!bannerImages) {
            e.next();
            return;
        }
        for (let i = 0; i < e.animeCollection!.mediaListCollection!.lists!.length; i++) {
            for (let j = 0; j < e.animeCollection!.mediaListCollection!.lists![i]!.entries!.length; j++) {
                const mediaId = e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.id
                const bannerImage = bannerImages[mediaId.toString()] || "";
                if (!!bannerImage) {
                    $replace(e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.bannerImage, bannerImage)
                }
            }
        }
        e.next();
    });

}
