import { useSeaCommand } from "@/app/(main)/_features/sea-command/sea-command.tsx"
import { SeaImage } from "@/components/shared/sea-image"
import { useRouter } from "@/lib/navigation"
import { useAtom } from "jotai"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { useWindowSize } from "react-use"
import { useServerStatus } from "../../_hooks/use-server-status"
import { __settings_tabAtom } from "../../settings/_components/settings-page.atoms"
import { __scanner_modalIsOpen } from "../anime-library/_containers/scanner-modal"
import { tourHelpers, useTour } from "./tour"
import { TourStep } from "./tour"

export const seenChangelogAtom = atomWithStorage<string | null>("sea-seen-changelog", null, undefined, { getOnInit: true })

function useSetupTour(): Record<string, () => TourStep[]> {
    const serverStatus = useServerStatus()
    const router = useRouter()
    const [, openScannerModal] = useAtom(__scanner_modalIsOpen)
    const [, setSettingsTab] = useAtom(__settings_tabAtom)
    const { setSeaCommandOpen, setSeaCommandInput } = useSeaCommand()

    const get3_5_0 = (): TourStep[] => {
        return [
            {
                id: "changelog-1",
                content: (
                    <div>
                        <h4 className="text-xl font-bold text-white">What's New in 3.5.0?</h4>
                        <p>Let's take a look at some of the new features.</p>
                    </div>
                ),
                route: "/",
                nextLabel: "Start",
                ignoreOutsideClick: true,
            },
            {
                id: "scanner",
                target: "[data-home-toolbar-scan-button]",
                title: "New Scanner",
                content: "The scanner's internal logic has been completely overhauled. It now uses a more context-aware algorithm which is more accurate.",
                route: "/",
                advanceOnTargetClick: true,
                ignoreOutsideClick: true,
                condition: () => !!serverStatus?.settings?.library?.libraryPath?.length,
                conditionFailBehavior: "modal",
            },
            {
                id: "scanner-2",
                target: "[data-scanner-modal-content]",
                title: "New Scanner",
                content: "The scanner now supports Anime Offline Database for matching data.",
                route: "/",
                prepare: () => {
                    openScannerModal(true)
                },
                advanceOnTargetClick: true,
                ignoreOutsideClick: true,
                condition: () => !!serverStatus?.settings?.library?.libraryPath?.length,
                conditionFailBehavior: "skip",
            },
            {
                id: "scanner-3",
                target: "[data-settings-anime-library='advanced-accordion-trigger']",
                title: "Scanner Configuration",
                content: "You can now fine-tune the scanner's matching behavior. Check out the documentation for more information.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("library")
                    await tourHelpers.waitForSelector("[data-settings-anime-library='advanced-accordion-trigger']")
                    await tourHelpers.click("[data-settings-anime-library='advanced-accordion-trigger']", 200)
                },
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
            },
            {
                id: "issue-recorder",
                target: "[data-open-issue-recorder-button]",
                title: "Issue Recorder",
                // content: "The issue recorder has been improved and will now record the UI.",
                content: <div>
                    <SeaImage
                        src="https://github.com/5rahim/hibike/blob/main/changelog/3_5-issue-recorder.gif?raw=true"
                        alt="Issue Recorder"
                        width="100%"
                        height="auto"
                        className="rounded-md"
                        allowGif
                    />
                    <p className="mt-2">The issue recorder has improved and can now record the UI, making bug reports more insightful.</p>
                </div>,
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("seanime")
                },
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
                popoverWidth: 500,
            },
            {
                id: "transcode-new-player",
                target: "[data-tab-trigger='mediastream']",
                title: "Transcode Player",
                content: "Transcoding/Direct Play now uses the custom Seanime player used by Seanime Denshi and Online Streaming.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("mediastream")
                },
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
            },
            {
                id: "search",
                target: "[data-vertical-menu-item='Search']",
                title: "Search",
                content: "The search menu item now opens the search page. You can still quickly search from any page by pressing 'S'.",
                route: "/search",
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
            },
            {
                id: "entry",
                title: "New Player Features",
                content: <div>
                    <SeaImage
                        src="https://github.com/5rahim/hibike/blob/main/changelog/3_5-videocore-characters.png?raw=true"
                        alt="Character Lookup"
                        width="100%"
                        height="auto"
                        className="rounded-md"
                    />
                    <p className="mt-2">Press 'H' to quickly look up characters while watching. Press 'Z' to toggle Stats for Nerds.</p>
                </div>,
                route: "/",
                advanceOnTargetClick: false,
                ignoreOutsideClick: false,
                popoverWidth: 500,
            },
        ]
    }

    const get3_7_0 = (): TourStep[] => {
        return [
            {
                id: "changelog-1",
                content: (
                    <div>
                        <h4 className="text-xl font-bold text-white">What's New in 3.7.0?</h4>
                        <p>Let's take a look at some of the new features.</p>
                    </div>
                ),
                route: "/",
                nextLabel: "Start",
                ignoreOutsideClick: true,
            },
            {
                id: "security",
                title: "Security Improvements",
                content: "3.7.0 includes several security improvements, including secure modes. Check out the documentation for more information.",
                route: "/",
                advanceOnTargetClick: true,
                ignoreOutsideClick: true,
            },
            {
                id: "search",
                target: "[data-advanced-search-options-tags='true']",
                title: "Tags",
                content: "The search page now supports searching by tags.",
                route: "/search",
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
            },
            {
                id: "search",
                target: ".sea-command-content",
                title: "Adult Entries in Global Search",
                content: "Global search no longer filters out adult entries if you have adult content enabled. (Reminder: Press 's' to open global search)",
                route: "/search",
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
                prepare: async () => {
                    setSeaCommandOpen(true)
                    setTimeout(() => {
                        setSeaCommandInput("/search ")
                    }, 200)
                    // wait 500ms
                    return new Promise(resolve => setTimeout(resolve, 500))
                },
            },
            {
                id: "changelog-2",
                title: "Bug Fixes",
                content: "Several bugs have been fixed in this release, including some related to Seanime Denshi and plugins. Read the full changelog for more details.",
                route: "/",
                ignoreOutsideClick: true,
            },
        ]
    }

    const get3_8_0 = (): TourStep[] => {
        return [
            {
                id: "changelog-1",
                content: (
                    <div>
                        <h4 className="text-xl font-bold text-white">What's New in 3.8.0?</h4>
                        <p>Let's take a look at the biggest additions in this release.</p>
                    </div>
                ),
                route: "/",
                nextLabel: "Start",
                ignoreOutsideClick: true,
            },
            {
                id: "torrent-search",
                title: "Torrent Search and Downloads",
                content: "Torrent search can now fan out across many providers at once. This release also smooths out a few debrid download edge cases.",
                route: "/",
                ignoreOutsideClick: true,
            },
            {
                id: "subtitle-translation",
                title: "Subtitle Translation",
                content: "Subtitle Translator now supports OpenAI-compatible local LLMs, so tools like LM Studio and Ollama can be used as local translation backends.",
                route: "/",
                ignoreOutsideClick: true,
            },
            {
                id: "external-player-link",
                target: "[data-settings-external-player-link-scheme]",
                title: "Local Subtitle Files",
                content: "Local subtitle files are now picked up automatically from the video folder, and external player links can use the new '{subtitleUrl}' placeholder for those local subtitle files.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("external-player-link")
                    await tourHelpers.waitForSelector("[data-settings-external-player-link-scheme]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "spoilers",
                target: "[data-settings-hide-anime-spoilers]",
                title: "Hide Spoilers",
                content: "You can now hide spoilers across the app, and on anime pages the new '/spoilers' command lets you toggle spoiler hiding for that specific anime.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("seanime")
                    await tourHelpers.waitForSelector("[data-settings-hide-anime-spoilers]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "online-streaming",
                target: "[data-settings-enable-onlinestream]",
                title: "Online Streaming",
                content: "Online streaming now uses a new HTTP/1-based proxy and can automatically cycle through providers until it finds one that works.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("onlinestream")
                    await tourHelpers.waitForSelector("[data-settings-enable-onlinestream]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "default-episode-source",
                target: "[data-settings-default-episode-source]",
                title: "Default Episode Source",
                content: "Choose which episode source Seanime should open by default when you land on an anime page.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("seanime")
                    await tourHelpers.waitForSelector("[data-settings-default-episode-source]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "ui-settings-redesign",
                target: "[data-settings-ui-panel-tabs]",
                title: "Redesigned UI Settings",
                content: "The User Interface settings panel has been redesigned so it is easier to navigate.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("ui")
                    await tourHelpers.waitForSelector("[data-settings-ui-panel-tabs]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "ui-settings-redesign2",
                target: ".settings-ui-navigation-preloading",
                title: "Route Preloading",
                content: "Seanime can now preload routes in the background to make navigation feel instant. You can adjust the preloading behavior in the new UI settings panel.",
                prepare: async () => {
                    window.scrollTo({ top: document.body.scrollHeight, behavior: "smooth" })
                    await new Promise(resolve => setTimeout(resolve, 1000))
                    await tourHelpers.waitForSelector(".settings-ui-navigation-preloading")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "entry-header-redesign",
                target: "[data-media-page-header]",
                title: "UI Updates",
                content: "The media header has been slightly redesigned. There are also new animations and transitions for a snappier experience.",
                prepare: async () => {
                    router.push("/entry?id=21827")
                    await tourHelpers.waitForSelector("[data-media-page-header]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "extensions",
                title: "Extensions",
                content: "Extensions can now be disabled without uninstalling them, and plugins have new APIs for settings, auth, and extension management.",
                route: "/extensions",
                ignoreOutsideClick: true,
            },
            {
                id: "extension-secure-mode",
                target: "[data-settings-enable-extension-secure-mode]",
                title: "Extension Secure Mode",
                content: "Enable Extension Secure Mode to get a confirmation prompt whenever an extension tries to perform a sensitive action.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("seanime")
                    await tourHelpers.waitForSelector("[data-settings-enable-extension-secure-mode]")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
            },
            {
                id: "denshi",
                title: "Denshi Window State",
                content: "Seanime Denshi now remembers its window position and size, so reopening the app brings you back to the same desktop layout.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("denshi")
                },
                condition: () => typeof window !== "undefined" && !!window.electron,
                conditionFailBehavior: "skip",
                ignoreOutsideClick: true,
            },
            // {
            //     id: "denshi",
            //     title: "View Transitions",
            //     content: "Seanime Denshi now uses the View Transitions API for native transitions between different screens.",
            //     route: "/schedule",
            //     prepare: async () => {
            //         setSettingsTab("seanime")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/lists")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/settings")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         // scroll to bottom
            //         window.scrollTo({ top: document.body.scrollHeight, behavior: "smooth" })
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/schedule")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/lists")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/settings")
            //         await new Promise(resolve => setTimeout(resolve, 650))
            //         router.push("/")
            //         await new Promise(resolve => setTimeout(resolve, 500))
            //     },
            //     condition: () => typeof window !== "undefined" && !!window.electron,
            //     conditionFailBehavior: "skip",
            //     ignoreOutsideClick: true,
            // },
            {
                id: "changelog-2",
                title: "Bug Fixes",
                content: "Several bugs have been fixed in this release, including some related to the built-in player. Read the full changelog for more details.",
                route: "/",
                ignoreOutsideClick: true,
            },
        ]
    }

    const get3_9_0 = (): TourStep[] => {
        return [
            {
                id: "changelog-1",
                content: (
                    <div>
                        <h4 className="text-xl font-bold text-white">What's New in 3.9.0?</h4>
                        <p>Let's take a look at the biggest additions in this release.</p>
                    </div>
                ),
                route: "/",
                nextLabel: "Start",
                ignoreOutsideClick: true,
            },
            {
                id: "libmpv-player",
                target: "[data-tab-trigger='playback']",
                title: "New Built-in Player (Denshi)",
                content: "Denshi now features a libmpv-based built-in player. It offers hardware-accelerated rendering directly in the app viewport, flawless codec & subtitle support, and supports mpv.conf options and shaders.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("playback")
                    await tourHelpers.waitForSelector("[data-tab-trigger='playback']")
                },
                ignoreOutsideClick: true,
                popoverWidth: 460,
                condition: () => typeof window !== "undefined" && !!window.electron,
                conditionFailBehavior: "skip",
            },
            {
                id: "torrent-streaming-perf",
                title: "Faster Torrent Streaming",
                content: "Torrent streaming startup is now up to 20% faster depending on seeding, with more accurate download progress reporting and fixed batch selection.",
                route: "/",
                ignoreOutsideClick: true,
            },
            {
                id: "debrid-streaming-perf",
                title: "Faster Debrid Streaming",
                content: "Debrid streaming launch is now up to 5 seconds faster for cached streams.",
                route: "/",
                ignoreOutsideClick: true,
            },
            {
                id: "changelog-2",
                title: "Bug Fixes",
                content: "Several bugs have been fixed in this release, including progress tracking for MPV/IINA, manga image proxy issues, and Seanime Denshi's Electron has been updated to 42.4.0. Read the full changelog for more details.",
                route: "/",
                ignoreOutsideClick: true,
            },
        ]
    }

    return {
        "3.5.0": get3_5_0,
        "3.7.0": get3_7_0,
        "3.8.0": get3_8_0,
        "3.9.0": get3_9_0,
    }
}

export function useChangelogTourListener() {
    const serverStatus = useServerStatus()
    const [seenChangelog, setSeenChangelog] = useAtom(seenChangelogAtom)
    const { start } = useTour()
    const tours = useSetupTour()
    const { width } = useWindowSize()
    const isMobile = width < 768

    const toursRef = React.useRef(tours)
    toursRef.current = tours

    const started = React.useRef(false)
    const timeout = React.useRef<NodeJS.Timeout | null>(null)

    React.useEffect(() => {
        if (!serverStatus?.showChangelogTour) return
        if (serverStatus.isOffline) return
        if (isMobile) return
        if (started.current) return

        if (seenChangelog === serverStatus.showChangelogTour) return

        started.current = true

        const tourId = serverStatus.showChangelogTour

        if (timeout.current) clearTimeout(timeout.current)
        timeout.current = setTimeout(() => {
            const getSteps = toursRef.current[tourId]
            if (getSteps) {
                start(getSteps(), tourId, () => {
                    console.log("tour completed")
                    setSeenChangelog(tourId)
                })
            }
        }, 1000)

        return () => {
            if (timeout.current) clearTimeout(timeout.current)
        }
    }, [serverStatus?.showChangelogTour, serverStatus?.isOffline, seenChangelog, start, setSeenChangelog, isMobile])

    return null
}
