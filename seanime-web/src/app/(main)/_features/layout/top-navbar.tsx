import { useRefreshAnimeCollection } from "@/api/hooks/anilist.hooks"
import { OfflineTopMenu } from "@/app/(main)/(offline)/offline/_components/offline-top-menu"
import { RefreshAnilistButton } from "@/app/(main)/_features/anilist/refresh-anilist-button"
import { LayoutHeaderBackground } from "@/app/(main)/_features/layout/_components/layout-header-background"
import { TopMenu } from "@/app/(main)/_features/navigation/top-menu"
import { ManualProgressTrackingButton } from "@/app/(main)/_features/progress-tracking/manual-progress-tracking"
import { PlaybackManagerProgressTrackingButton } from "@/app/(main)/_features/progress-tracking/playback-manager-progress-tracking"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ChapterDownloadsButton } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-button"
import { __manga_chapterDownloadsDrawerIsOpenAtom } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import { cn } from "@/components/ui/core/styling"
import { Separator } from "@/components/ui/separator/separator"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import React from "react"
import { FaDownload } from "react-icons/fa"
import { IoReload } from "react-icons/io5"
import { PluginSidebarTray } from "../plugin/tray/plugin-sidebar-tray"

type TopNavbarProps = {
    children?: React.ReactNode
}

export function TopNavbar(props: TopNavbarProps) {

    const {
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const isOffline = serverStatus?.isOffline
    const ts = useThemeSettings()

    return (
        <>
            <div
                data-top-navbar
                className={cn(
                    "w-full h-[5rem] relative overflow-hidden flex items-center",
                    (ts.hideTopNavbar || process.env.NEXT_PUBLIC_PLATFORM === "desktop") && "lg:hidden",
                )}
            >
                <div data-top-navbar-content-container className="relative z-10 px-4 w-full flex flex-row md:items-center overflow-x-auto">
                    <div data-top-navbar-content className="flex items-center w-full gap-3">
                        <AppSidebarTrigger />
                        {!isOffline ? <TopMenu /> : <OfflineTopMenu />}
                        <PlaybackManagerProgressTrackingButton />
                        <ManualProgressTrackingButton />
                        <div data-top-navbar-content-separator className="flex flex-1"></div>
                        <PluginSidebarTray place="top" />
                        {!isOffline && <ChapterDownloadsButton />}
                        {!isOffline && <RefreshAnilistButton />}
                    </div>
                </div>
                <LayoutHeaderBackground />
            </div>
        </>
    )
}


type SidebarNavbarProps = {
    isCollapsed: boolean
    handleExpandSidebar: () => void
    handleUnexpandedSidebar: () => void
}

export function SidebarNavbar(props: SidebarNavbarProps) {

    const {
        isCollapsed,
        handleExpandSidebar,
        handleUnexpandedSidebar,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const ts = useThemeSettings()
    const pathname = usePathname()

    const openDownloadQueue = useSetAtom(__manga_chapterDownloadsDrawerIsOpenAtom)
    const isMangaPage = pathname.startsWith("/manga")

    /**
     * @description
     * - Asks the server to fetch an up-to-date version of the user's AniList collection.
     */
    const { mutate: refreshAC, isPending: isRefreshingAC } = useRefreshAnimeCollection()

    if (!ts.hideTopNavbar && process.env.NEXT_PUBLIC_PLATFORM !== "desktop") return null

    return (
        <div data-sidebar-navbar className="flex flex-col gap-1">
            <div data-sidebar-navbar-spacer className="px-4 lg:py-1">
                <Separator className="px-4" />
            </div>
            {!serverStatus?.isOffline && <VerticalMenu
                data-sidebar-navbar-vertical-menu
                className="px-4"
                collapsed={isCollapsed}
                itemClass="relative"
                onMouseEnter={handleExpandSidebar}
                onMouseLeave={handleUnexpandedSidebar}
                items={[
                    {
                        iconType: IoReload,
                        name: "Refresh AniList",
                        onClick: () => {
                            if (isRefreshingAC) return
                            refreshAC()
                        },
                    },
                    ...(isMangaPage ? [
                        {
                            iconType: FaDownload,
                            name: "Manga downloads",
                            onClick: () => {
                                openDownloadQueue(true)
                            },
                        },
                    ] : []),
                ]}
            />}
            <div data-sidebar-navbar-playback-manager-progress-tracking-button className="flex justify-center">
                <PlaybackManagerProgressTrackingButton asSidebarButton />
            </div>
            <div data-sidebar-navbar-manual-progress-tracking-button className="flex justify-center">
                <ManualProgressTrackingButton asSidebarButton />
            </div>
        </div>
    )
}
