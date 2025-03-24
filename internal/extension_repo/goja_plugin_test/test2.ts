/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/hooks.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />

function init() {
    $ui.register((ctx) => {
        // Create the tray icon
        const tray = ctx.newTray({
            tooltipText: "Hide spoilers",
            iconUrl: "https://seanime.rahim.app/logo_2.png",
            withContent: true,
        })

        const [, refetchEpisodeCard] = ctx.dom.observe("[data-episode-card]", async (episodeCards) => {
            const blurThumbnails = $storage.get("params.blurThumbnails") || false
            const blurTitles = $storage.get("params.blurTitles") || false
            const blurDescriptions = $storage.get("params.blurDescriptions") || false

            const listDataElement = await ctx.dom.queryOne("[data-anime-entry-list-data]")
            if (!listDataElement) {
                console.error("listDataElement not found")
                return
            }
            const listDataStr = await listDataElement.getAttribute("data-anime-entry-list-data")
            const listData = JSON.parse(listDataStr || "{}") as Record<string, any>

            await Promise.all(episodeCards.map(async (episodeCard) => {
                const episodeNumberStr = await episodeCard.getAttribute("data-episode-number")

                const episodeNumber = Number(episodeNumberStr)
                if (!isNaN(episodeNumber)) {
                    const image = await episodeCard.queryOne("[data-episode-card-image]")
                    if (image) {
                        if (blurThumbnails) {
                            if (episodeNumber > Number(listData.progress)) {
                                image.setStyle("filter", "blur(24px)")
                            } else {
                                image.removeStyle("filter")
                            }
                        } else {
                            image.removeStyle("filter")
                        }
                    }
                    const title = await episodeCard.queryOne("[data-episode-card-title]")
                    if (title) {
                        if (blurTitles) {
                            if (episodeNumber > Number(listData.progress)) {
                                title.setStyle("filter", "blur(4px)")
                            } else {
                                title.removeStyle("filter")
                            }
                        } else {
                            title.removeStyle("filter")
                        }
                    }
                }
            }))
        })

        const [, refetchEpisodeGridItem] = ctx.dom.observe("[data-episode-grid-item]", async (episodeGridItems) => {
            try {
                const blurThumbnails = $storage.get("params.blurThumbnails") || false
                const blurTitles = $storage.get("params.blurTitles") || false
                const blurDescriptions = $storage.get("params.blurDescriptions") || false

                const listDataElement = await ctx.dom.queryOne("[data-anime-entry-list-data]")
                if (!listDataElement) {
                    console.error("listDataElement not found")
                    return
                }
                const listDataStr = await listDataElement.getAttribute("data-anime-entry-list-data")
                const listData = JSON.parse(listDataStr || "{}") as Record<string, any>

                console.log("observe episodeGridItem", episodeGridItems.length)

                for (const episodeGridItem of episodeGridItems) {
                    const episodeNumberStr = await episodeGridItem.getAttribute("data-episode-number")
                    const episodeNumber = Number(episodeNumberStr)
                    if (!isNaN(episodeNumber)) {
                        const image = await episodeGridItem.queryOne("[data-episode-grid-item-image]")
                        if (image) {
                            if (blurThumbnails) {
                                if (episodeNumber > Number(listData.progress)) {
                                    image.setStyle("filter", "blur(24px)")
                                } else {
                                    image.removeStyle("filter")
                                }
                            } else {
                                image.removeStyle("filter")
                            }
                        } else {
                            console.error("image not found")
                        }
                        const title = await episodeGridItem.queryOne("[data-episode-grid-item-episode-title]")
                        if (title) {
                            if (blurTitles) {
                                if (episodeNumber > Number(listData.progress)) {
                                    title.setStyle("filter", "blur(4px)")
                                } else {
                                    title.removeStyle("filter")
                                }
                            } else {
                                title.removeStyle("filter")
                            }
                        } else {
                            console.error("title not found")
                        }
                        const description = await episodeGridItem.queryOne("[data-episode-grid-item-episode-description]")
                        if (description) {
                            if (blurDescriptions) {
                                if (episodeNumber > Number(listData.progress)) {
                                    description.setStyle("filter", "blur(4px)")
                                } else {
                                    description.removeStyle("filter")
                                }
                            } else {
                                description.removeStyle("filter")
                            }
                        }
                    }
                }
            }
            catch (e) {
                console.error("Error processing episodeGridItem", e)
            }
        })


        const blurThumbnailsRef = ctx.registerFieldRef<boolean>("blurThumbnailsRef")
        const blurTitlesRef = ctx.registerFieldRef<boolean>("blurTitlesRef")
        const blurDescriptionsRef = ctx.registerFieldRef<boolean>("blurDescriptionsRef")

        function updateForm() {
            const params = $storage.get("params") || {}
            blurThumbnailsRef.setValue(params.blurThumbnails || false)
            blurTitlesRef.setValue(params.blurTitles || false)
            blurDescriptionsRef.setValue(params.blurDescriptions || false)
            tray.update()
        }

        tray.onOpen(() => {
            updateForm()
        })

        ctx.registerEventHandler("save", () => {
            $storage.set("params", {
                blurThumbnails: blurThumbnailsRef.current,
                blurTitles: blurTitlesRef.current,
                blurDescriptions: blurDescriptionsRef.current,
            })
            updateForm()
            refetchEpisodeCard()
            refetchEpisodeGridItem()
            ctx.toast.success("Settings saved")
        })

        tray.render(() => tray.stack([
            tray.text("Hide potential spoilers"),
            tray.switch("Blur thumbnails", { fieldRef: "blurThumbnailsRef", value: blurThumbnailsRef.current }),
            tray.switch("Blur titles", { fieldRef: "blurTitlesRef", value: blurTitlesRef.current }),
            tray.switch("Blur descriptions", { fieldRef: "blurDescriptionsRef", value: blurDescriptionsRef.current }),
            tray.button("Save", { onClick: "save" }),
        ]))

        ctx.dom.onReady(() => {
            console.log("ready")
        })
    });
}
