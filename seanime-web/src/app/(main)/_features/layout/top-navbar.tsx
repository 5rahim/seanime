import { useRefreshAnimeCollection } from "@/api/hooks/anilist.hooks"
import { ProgressTracking } from "@/app/(main)/(library)/_containers/playback-manager/progress-tracking"
import { RefreshAnilistButton } from "@/app/(main)/_features/anilist/refresh-anilist-button"
import { LayoutHeaderBackground } from "@/app/(main)/_features/layout/_components/layout-header-background"
import { TopMenu } from "@/app/(main)/_features/navigation/top-menu"
import { ChapterDownloadsButton } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-button"
import { __manga__chapterDownloadsDrawerIsOpenAtom } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import { Separator } from "@/components/ui/separator/separator"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai/react"
import { usePathname } from "next/navigation"
import React from "react"
import { FaDownload } from "react-icons/fa"
import { IoReload } from "react-icons/io5"

type TopNavbarProps = {
    children?: React.ReactNode
}

export function TopNavbar(props: TopNavbarProps) {

    const {
        children,
        ...rest
    } = props

    const ts = useThemeSettings()
    const pathname = usePathname()

    if (ts.hideTopNavbar) return null

    return (
        <>
            <div className="w-full h-[5rem] relative overflow-hidden flex items-center">
                <div className="relative z-10 px-4 w-full flex flex-row md:items-center overflow-x-auto">
                    <div className="flex items-center w-full gap-3">
                        <AppSidebarTrigger />
                        <TopMenu />
                        <ProgressTracking />
                        <div className="flex flex-1"></div>
                        <ChapterDownloadsButton />
                        <RefreshAnilistButton />
                    </div>
                </div>
                <LayoutHeaderBackground />
            </div>
        </>
    )
}


type SidebarNavbarProps = {
    children?: React.ReactNode
    isCollapsed: boolean
    handleExpandSidebar: () => void
    handleUnexpandedSidebar: () => void
}

export function SidebarNavbar(props: SidebarNavbarProps) {

    const {
        children,
        isCollapsed,
        handleExpandSidebar,
        handleUnexpandedSidebar,
        ...rest
    } = props

    const ts = useThemeSettings()
    const pathname = usePathname()

    const openDownloadQueue = useSetAtom(__manga__chapterDownloadsDrawerIsOpenAtom)
    const isMangaPage = pathname.startsWith("/manga")

    /**
     * @description
     * - Asks the server to fetch an up-to-date version of the user's AniList collection.
     */
    const { mutate: refreshAC, isPending: isRefreshingAC } = useRefreshAnimeCollection()

    if (!ts.hideTopNavbar) return null

    return (
        <div className="flex flex-col gap-1">
            <div className="px-4 lg:py-1">
                <Separator className="px-4" />
            </div>
            <VerticalMenu
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
            />
            <div className="flex justify-center">
                <ProgressTracking asSidebarButton />
            </div>
        </div>
    )
}
