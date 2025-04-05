/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/app.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />

//@ts-ignore
function init() {
    $ui.register((ctx) => {
        // Create the tray icon
        const tray = ctx.newTray({
            iconUrl: "https://seanime.rahim.app/logo_2.png",
            withContent: true,
        })

        // Keep track of the current media ID
        const currentMediaId = ctx.state(0)

        // Create a field ref for the URL input
        const inputRef = ctx.fieldRef()

        // When the plugin loads, fetch the current screen and set the badge to 0
        ctx.screen.loadCurrent() // Triggers onNavigate
        tray.updateBadge({ number: 0 })
        // Also fetch current screen when tray is open
        tray.onOpen(() => {
            ctx.screen.loadCurrent()
        })

        // Updates the field's value and badge based on the current anime page
        function updateState() {
            // Reset the badge and input if the user currently isn't on an anime page
            if (!currentMediaId.get()) {
                inputRef.setValue("")
                tray.updateBadge({ number: 0 })
            }
            // Get the stored background image URL for this anime
            const url = $storage.get<string>("backgroundImages." + currentMediaId.get())
            if (url) {
                // If there's a URL, set the value of the input
                inputRef.setValue(url)
                // Add a badge
                tray.updateBadge({ number: 1, intent: "info" })
            } else {
                inputRef.setValue("")
                tray.updateBadge({ number: 0 })
            }
        }

        // Run the function when the plugin loads
        updateState()

        // Update currentMediaId when the user navigates
        ctx.screen.onNavigate((e) => {
            console.log("onNavigate", e)
            // If the user navigates to an anime page
            if (e.pathname === "/entry" && !!e.searchParams.id) {
                // Get the ID from the URL
                const id = parseInt(e.searchParams.id)
                currentMediaId.set(id)
            } else {
                currentMediaId.set(0)
            }
        })

        // This effect will update the state each time currentMediaId changes
        ctx.effect(() => {
            updateState()
        }, [currentMediaId])

        // Create a handler to store the custom banner image URL
        ctx.registerEventHandler("save", () => {
            if (!!inputRef.current) {
                $storage.set(`backgroundImages.${currentMediaId.get()}`, inputRef.current)
            } else {
                $storage.remove(`backgroundImages.${currentMediaId.get()}`)
            }
            ctx.toast.success("Background image saved")
            updateState() // Update the state

            // Updates the data on the client
            // This is better than calling ctx.screen.reload()
            $anilist.refreshAnimeCollection()

            ctx.notification.send("Background image saved")
        });

        ctx.setTimeout(() => {
            ctx.screen.reload()
        }, 1000)

        // Tray content
        tray.render(() => {
            return tray.stack([
                currentMediaId.get() === 0
                    ? tray.text("Open an anime")
                    : tray.stack([
                        tray.text(`Current media ID: ${currentMediaId.get()}`),
                        tray.input({ fieldRef: inputRef }),
                        tray.button({ label: "Save", onClick: "save", intent: "primary" }),
                    ]),
            ])
        })
    })

    // Register hook handlers to listen and modify the anime collection.

    // Triggers the app loads the user's AniList anime collection
    $app.onGetAnimeCollection((e) => {
        const bannerImages = $storage.get<Record<string, string | undefined>>("backgroundImages")
        if (!bannerImages) {
            e.next()
            return
        }
        if (!e.animeCollection?.mediaListCollection?.lists?.length) {
            e.next()
            return
        }

        for (let i = 0; i < e.animeCollection!.mediaListCollection!.lists!.length; i++) {
            for (let j = 0; j < e.animeCollection!.mediaListCollection!.lists![i]!.entries!.length; j++) {
                const mediaId = e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.id
                const bannerImage = bannerImages[mediaId.toString()]
                if (!!bannerImage) {
                    $replace(e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.bannerImage, bannerImage)
                }
            }
        }

        e.next()
    })

    // Same as onGetAnimeCollection but also includes custom lists.
    $app.onGetRawAnimeCollection((e) => {
        const bannerImages = $storage.get<Record<string, string | undefined>>("backgroundImages")
        if (!bannerImages) {
            e.next()
            return
        }
        if (!e.animeCollection?.mediaListCollection?.lists?.length) {
            e.next()
            return
        }

        for (let i = 0; i < e.animeCollection!.mediaListCollection!.lists!.length; i++) {
            for (let j = 0; j < e.animeCollection!.mediaListCollection!.lists![i]!.entries!.length; j++) {
                const mediaId = e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.id
                const bannerImage = bannerImages[mediaId.toString()]
                if (!!bannerImage) {
                    $replace(e.animeCollection!.mediaListCollection!.lists![i]!.entries![j]!.media!.bannerImage, bannerImage)
                }
            }
        }

        e.next()
    })

    $app.onAnimeEntryRequested((e) => {
        let lfs = $clone(e.localFiles)!

        const toInsert = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15]

        toInsert.forEach(nb => {
            let metadataAniDbEp = nb.toString()
            if (nb > 12) {
                metadataAniDbEp = "S" + (nb - 12)
            }
            lfs.push({
                path: `/Volumes/Seagate Portable Drive/ANIME/Bakemonogatari/Episode ${nb < 10 ? "0" + nb : nb}.mkv`,
                name: `Episode ${nb < 10 ? "0" + nb : nb}.mkv`,
                locked: true,
                ignored: false,
                mediaId: 5081,
                metadata: {
                    episode: nb,
                    aniDBEpisode: metadataAniDbEp,
                    type: "main",
                },
            })
        })

        $replace(e.localFiles, lfs)

        e.next()
    })

    $app.onAnimeEntryRequested((e) => {
        let lfs = $clone(e.localFiles)!

        // const toInsert = [1,2,3,4,5,6,7,8,9,10,11,12,13]
        // const toInsert = [{ ep: 1, aniDbEp: "S1" }, { ep: 2, aniDbEp: "1" }, { ep: 3, aniDbEp: "2" }, { ep: 4, aniDbEp: "3" }, { ep: 5, aniDbEp:
        // "4" }, { ep: 6, aniDbEp: "5" }, { ep: 7, aniDbEp: "6" }, { ep: 8, aniDbEp: "7" }, { ep: 9, aniDbEp: "8" }, { ep: 10, aniDbEp: "9" },  {
        // ep: 11, aniDbEp: "10" }, { ep: 12, aniDbEp: "11" }, { ep: 13, aniDbEp: "12" }]


        // const toInsert = [0,1,2,3,4,5,6,7,8,9,10,11,12]
        const toInsert = [{ ep: 0, aniDbEp: "S1" }, { ep: 1, aniDbEp: "1" }, { ep: 2, aniDbEp: "2" }, { ep: 3, aniDbEp: "3" },
            { ep: 4, aniDbEp: "4" }, { ep: 5, aniDbEp: "5" }, { ep: 6, aniDbEp: "6" }, { ep: 7, aniDbEp: "7" }, { ep: 8, aniDbEp: "8" },
            { ep: 9, aniDbEp: "9" }, { ep: 10, aniDbEp: "10" },
            { ep: 11, aniDbEp: "11" }, { ep: 12, aniDbEp: "12" }]


        // const toInsert = [{ ep: 1, aniDbEp: "1" }, { ep: 2, aniDbEp: "2" }, { ep: 3, aniDbEp: "3" }, { ep: 4, aniDbEp: "4" }, { ep: 5, aniDbEp:
        // "5" }, { ep: 6, aniDbEp: "6" }, { ep: 7, aniDbEp: "7" }, { ep: 8, aniDbEp: "8" }, { ep: 9, aniDbEp: "9" }, { ep: 10, aniDbEp: "10" },  {
        // ep: 11, aniDbEp: "11" }, { ep: 12, aniDbEp: "12" }]

        toInsert.forEach(n => {
            const nb = n.ep
            lfs.push({
                path: `/Volumes/Seagate Portable Drive/ANIME/Mushoku Tensei II: Isekai Ittara Honki Dasu/Episode ${nb < 10 ? "0" + nb : nb}.mkv`,
                name: `Episode ${nb < 10 ? "0" + nb : nb}.mkv`,
                locked: true,
                ignored: false,
                mediaId: 146065,
                parsedInfo: {
                    original: "",
                },
                parsedFolderInfo: [],
                metadata: {
                    episode: nb,
                    aniDBEpisode: n.aniDbEp,
                    type: "main",
                },
            })
        })

        $replace(e.localFiles, lfs)

        e.next()
    })
}
