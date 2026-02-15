import { SeaImage } from "@/components/shared/sea-image"
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
    const [, openScannerModal] = useAtom(__scanner_modalIsOpen)
    const [settingsTab, setSettingsTab] = useAtom(__settings_tabAtom)

    const get3_5_0 = (): TourStep[] => {
        return [
            {
                id: "changelog-1",
                content: (
                    <div>
                        <h4 className="text-xl font-bold text-white">3.5.0</h4>
                        <p>Let's take a look at some of the new features in 3.5.0.</p>
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
                content: "You can now fine-tune the scanner's matching and hydration behavior. Check out the documentation for more information.",
                route: "/settings",
                prepare: async () => {
                    setSettingsTab("library")
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
                        src="https://i.postimg.cc/7Z13W8HN/2026-02-15-10-39-43.gif"
                        alt="Issue Recorder"
                        width="100%"
                        height="auto"
                        className="rounded-md"
                        allowGif
                    />
                    <p className="mt-2">The issue recorder has been improved and can now record the UI, making bug reports more insightful.</p>
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
                content: "Transcoding/Direct Play now uses the default Seanime player used by Seanime Denshi and Online Streaming.",
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
                content: "The search menu item now opens the search page. You can still quickly search from any page by using the 's' keyboard shortcut.",
                route: "/search",
                advanceOnTargetClick: false,
                ignoreOutsideClick: true,
            },
            {
                id: "entry",
                title: "New Player Features",
                // content: "Use the 'H' keybind to quickly look up characters in the player. Use 'Z' to toggle Stats for Nerds.",
                content: <div>
                    <SeaImage
                        src="https://i.postimg.cc/W4FkcBjf/img-2026-02-07-13-31-33.png"
                        alt="Character Lookup"
                        width="100%"
                        height="auto"
                        className="rounded-md"
                    />
                    <p className="mt-2">Use the 'H' keybind to quickly look up characters in the player. Use 'Z' to toggle Stats for Nerds.</p>
                </div>,
                route: "/",
                advanceOnTargetClick: false,
                ignoreOutsideClick: false,
                popoverWidth: 500,
            },
        ]
    }

    return {
        "3.5.0": get3_5_0,
    }
}

export function useChangelogTourListener() {
    const serverStatus = useServerStatus()
    const [seenChangelog, setSeenChangelog] = useAtom(seenChangelogAtom)
    const { start } = useTour()
    const tours = useSetupTour()
    const { width } = useWindowSize()
    const isMobile = width < 1024

    const started = React.useRef(false)
    const timeout = React.useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (started.current) return
        if (isMobile) return
        if (timeout.current) clearTimeout(timeout.current)
        timeout.current = setTimeout(() => {
            if (!serverStatus?.isOffline && !!serverStatus?.showChangelogTour) {
                const tour = tours[serverStatus.showChangelogTour]
                const seen = serverStatus.showChangelogTour === seenChangelog
                started.current = true
                if (tour && !seen) {
                    start(tour(), serverStatus.showChangelogTour, () => {
                        console.log("tour completed")
                        setSeenChangelog(serverStatus.showChangelogTour)
                    })
                }
            }
        }, 1000)
    }, [serverStatus, setSeenChangelog, serverStatus?.showChangelogTour, tours, isMobile])

    return null
}
