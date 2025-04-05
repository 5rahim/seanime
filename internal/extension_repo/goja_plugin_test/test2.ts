/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/app.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />
/// <reference path="../../goja/goja_bindings/js/core.d.ts" />

//@ts-ignore
function init() {
    $app.onGetAnimeCollection((e) => {
        const animeDataSet: Record<string, { bannerImage: string, title: string }> = {}
        for (const list of e.animeCollection?.mediaListCollection?.lists || []) {
            for (const entry of list.entries || []) {
                if (!entry?.media) continue

                if (animeDataSet[String(entry.media.id)]) {
                    continue
                }

                animeDataSet[String(entry.media.id)] = {
                    bannerImage: entry.media.bannerImage || entry.media.coverImage?.extraLarge || "",
                    title: entry.media.title?.userPreferred || "",
                }
            }
        }
        $storage.set("animeDataSet", animeDataSet)

        e.next()
    })

    $ui.register((ctx) => {
        // Create the tray icon
        const tray = ctx.newTray({
            iconUrl: "https://cdn-icons-png.flaticon.com/512/3686/3686669.png",
            withContent: true,
        })

        $anilist.getAnimeCollection(false)

        const [, refetchEpisodeCard] = ctx.dom.observe("[data-episode-card]", async (episodeCards) => {
            try {
                const hideThumbnails = $storage.get("params.hideThumbnails") || false
                const hideTitles = $storage.get("params.hideTitles") || false
                const skipNextEpisode = $storage.get("params.skipNextEpisode") || false
                const animeDataSet = $storage.get<Record<string, { bannerImage: string, title: string }>>("animeDataSet") || {}

                const listDataElement = await ctx.dom.queryOne("[data-anime-entry-list-data]")
                if (!listDataElement) {
                    for (const episodeCard of episodeCards) {
                        const animeIdStr = episodeCard.attributes["data-media-id"]
                        const animeId = Number(animeIdStr)

                        if (!isNaN(animeId) && !!animeDataSet[String(animeId)]) {

                            const $ = LoadDoc(episodeCard.innerHTML!)
                            const imageSelection = $("[data-episode-card-image]")
                            if (imageSelection.length() === 0 || !imageSelection.attr("id")) {
                                continue
                            }

                            const image = ctx.dom.asElement(imageSelection.attr("id")!)

                            const previous = JSON.parse(imageSelection.data("original") || "{}")

                            if (hideThumbnails && !skipNextEpisode) {
                                image.setProperty("src", animeDataSet[String(animeId)].bannerImage)
                            } else if (previous.property?.src) {
                                image.setProperty("src", previous.property.src)
                            }

                            const titleSelection = $("[data-episode-card-title]")
                            if (titleSelection.length() === 0 || !titleSelection.attr("id")) {
                                continue
                            }

                            const title = ctx.dom.asElement(titleSelection.attr("id")!)
                            const titlePrevious = JSON.parse(titleSelection.data("original") || "{}")

                            if (hideTitles && !skipNextEpisode) {
                                title.setText(animeDataSet[String(animeId)].title)
                            } else if (titlePrevious.text?.textContent) {
                                title.setText(titlePrevious.text.textContent)
                            }

                        }
                    }

                    return
                }
                const listDataStr = await listDataElement.getAttribute("data-anime-entry-list-data")
                const listData = JSON.parse(listDataStr || "{}") as Record<string, any>


                let progress = Number(listData?.progress || 0)
                if (skipNextEpisode) {
                    progress = progress + 1
                }


                for (const episodeCard of episodeCards) {
                    const episodeNumberStr = episodeCard.attributes["data-episode-number"]
                    const episodeNumber = Number(episodeNumberStr)
                    if (!isNaN(episodeNumber)) {
                        const $ = LoadDoc(episodeCard.innerHTML!)
                        const imageSelection = $("[data-episode-card-image]")
                        if (imageSelection.length() === 0 || !imageSelection.attr("id")) {
                            continue
                        }

                        const image = ctx.dom.asElement(imageSelection.attr("id")!)
                        if (hideThumbnails && episodeNumber > progress) {
                            image.setStyle("filter", "blur(24px)")
                        } else {
                            image.removeStyle("filter")
                        }


                        const titleSelection = $("[data-episode-card-title]")
                        if (titleSelection.length() === 0 || !titleSelection.attr("id")) {
                            continue
                        }

                        const title = ctx.dom.asElement(titleSelection.attr("id")!)
                        if (hideTitles && episodeNumber > progress) {
                            title.setStyle("filter", "blur(4px)")
                        } else {
                            title.removeStyle("filter")
                        }
                    }
                }
            }
            catch (e) {
                console.error("Error processing episodeCard", e)
            }
        }, { withInnerHTML: true, identifyChildren: true })


        const [, refetchEpisodeGridItem] = ctx.dom.observe("[data-episode-grid-item]", async (episodeGridItems) => {
            try {
                const hideThumbnails = $storage.get("params.hideThumbnails") || false
                const hideTitles = $storage.get("params.hideTitles") || false
                const hideDescriptions = $storage.get("params.hideDescriptions") || false

                const listDataElement = await ctx.dom.queryOne("[data-anime-entry-list-data]")
                if (!listDataElement) {
                    console.error("listDataElement not found")
                    return
                }
                const listDataStr = await listDataElement.getAttribute("data-anime-entry-list-data")
                const listData = JSON.parse(listDataStr || "{}") as Record<string, any>

                const skipNextEpisode = $storage.get("params.skipNextEpisode") || false

                let progress = Number(listData?.progress || 0)
                if (skipNextEpisode) {
                    progress = progress + 1
                }


                for (const episodeGridItem of episodeGridItems) {
                    const episodeNumberStr = episodeGridItem.attributes["data-episode-number"]
                    const episodeNumber = Number(episodeNumberStr)
                    if (!isNaN(episodeNumber)) {
                        const $ = LoadDoc(episodeGridItem.innerHTML!)

                        const imageSelection = $("[data-episode-grid-item-image]")
                        if (imageSelection.length() === 0 || !imageSelection.attr("id")) {
                            continue
                        }

                        const image = ctx.dom.asElement(imageSelection.attr("id")!)

                        if (hideThumbnails && episodeNumber > progress) {
                            image.setStyle("filter", "blur(24px)")
                        } else {
                            image.removeStyle("filter")
                        }

                        try {
                            const titleSelection = $("[data-episode-grid-item-episode-title]")
                            if (titleSelection.length() === 0 || !titleSelection.attr("id")) {
                                continue
                            }

                            const title = ctx.dom.asElement(titleSelection.attr("id")!)

                            if (hideTitles && episodeNumber > progress) {
                                title.setStyle("filter", "blur(4px)")
                            } else {
                                title.removeStyle("filter")
                            }
                        }
                        catch (e) {
                            console.error("Error processing title", e)
                        }

                        try {
                            const descriptionSelection = $("[data-episode-grid-item-episode-description]")
                            if (descriptionSelection.length() === 0 || !descriptionSelection.attr("id")) {
                                continue
                            }

                            const description = ctx.dom.asElement(descriptionSelection.attr("id")!)
                            if (hideDescriptions && episodeNumber > progress) {
                                description.setStyle("filter", "blur(4px)")
                            } else {
                                description.removeStyle("filter")
                            }
                        }
                        catch (e) {
                            console.error("Error processing description", e)
                        }
                    }
                }
            }
            catch (e) {
                console.error("Error processing episodeGridItem", e)
            }
        }, { withInnerHTML: true, identifyChildren: true })


        const hideThumbnailsRef = ctx.fieldRef<boolean>()
        const hideTitlesRef = ctx.fieldRef<boolean>()
        const hideDescriptionsRef = ctx.fieldRef<boolean>()
        const skipNextEpisodeRef = ctx.fieldRef<boolean>()
        function updateForm() {
            const params = $storage.get("params") || {}
            hideThumbnailsRef.setValue(params.hideThumbnails || false)
            hideTitlesRef.setValue(params.hideTitles || false)
            hideDescriptionsRef.setValue(params.hideDescriptions || false)
            skipNextEpisodeRef.setValue(params.skipNextEpisode || false)
        }

        tray.onOpen(() => {
            updateForm()
        })

        ctx.registerEventHandler("save", () => {
            $storage.set("params", {
                hideThumbnails: hideThumbnailsRef.current,
                hideTitles: hideTitlesRef.current,
                hideDescriptions: hideDescriptionsRef.current,
                skipNextEpisode: skipNextEpisodeRef.current,
            })
            updateForm()
            refetchEpisodeCard()
            refetchEpisodeGridItem()
            ctx.toast.success("Settings saved")
        })

        tray.render(() => tray.stack([
            tray.text("Hide potential spoilers"),
            tray.stack([
                tray.switch("Hide thumbnails", { fieldRef: hideThumbnailsRef }),
                tray.switch("Hide titles", { fieldRef: hideTitlesRef }),
                tray.switch("Hide descriptions", { fieldRef: hideDescriptionsRef }),
            ], { gap: 0 }),
            tray.checkbox("Skip next episode", { fieldRef: skipNextEpisodeRef }),
            tray.button("Save", { onClick: "save", intent: "primary" }),
        ]))

        ctx.dom.onReady(() => {
            refetchEpisodeCard()
            refetchEpisodeGridItem()
        })
    });
}
