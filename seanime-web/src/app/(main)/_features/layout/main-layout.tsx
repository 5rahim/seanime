"use client"
import { PlaylistsModal } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner-modal"
import { ErrorExplainer } from "@/app/(main)/_features/error-explainer/error-explainer"
import { GlobalSearch } from "@/app/(main)/_features/global-search/global-search"
import { IssueReport } from "@/app/(main)/_features/issue-report/issue-report"
import { LibraryWatcher } from "@/app/(main)/_features/library-watcher/library-watcher"
import { MediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { MainSidebar } from "@/app/(main)/_features/navigation/main-sidebar"
import { PluginManager } from "@/app/(main)/_features/plugin/plugin-manager"
import { ManualProgressTracking } from "@/app/(main)/_features/progress-tracking/manual-progress-tracking"
import { PlaybackManagerProgressTracking } from "@/app/(main)/_features/progress-tracking/playback-manager-progress-tracking"
import { SeaCommand } from "@/app/(main)/_features/sea-command/sea-command"
import { useAnimeCollectionLoader } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useAnimeLibraryCollectionLoader } from "@/app/(main)/_hooks/anime-library-collection-loader"
import { useMissingEpisodesLoader } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useAnimeCollectionListener } from "@/app/(main)/_listeners/anilist-collection.listeners"
import { useAutoDownloaderItemListener } from "@/app/(main)/_listeners/autodownloader.listeners"
import { useExtensionListener } from "@/app/(main)/_listeners/extensions.listeners"
import { useExternalPlayerLinkListener } from "@/app/(main)/_listeners/external-player-link.listeners"
import { useMangaListener } from "@/app/(main)/_listeners/manga.listeners"
import { useMiscEventListeners } from "@/app/(main)/_listeners/misc-events.listeners"
import { useSyncListener } from "@/app/(main)/_listeners/sync.listeners"
import { DebridStreamOverlay } from "@/app/(main)/entry/_containers/debrid-stream/debrid-stream-overlay"
import { TorrentStreamOverlay } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { ChapterDownloadsDrawer } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import { __isElectronDesktop__ } from "@/types/constants"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { useServerStatus } from "../../_hooks/use-server-status"
import { useInvalidateQueriesListener } from "../../_listeners/invalidate-queries.listeners"
import { Announcements } from "../announcements"
import { NakamaManager } from "../nakama/nakama-manager"
import { NativePlayer } from "../native-player/native-player"
import { TopIndefiniteLoader } from "../top-indefinite-loader"

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    /**
     * Data loaders
     */
    useAnimeLibraryCollectionLoader()
    useAnimeCollectionLoader()
    useMissingEpisodesLoader()

    /**
     * Websocket listeners
     */
    useAutoDownloaderItemListener()
    useAnimeCollectionListener()
    useMiscEventListeners()
    useExtensionListener()
    useMangaListener()
    useExternalPlayerLinkListener()
    useSyncListener()
    useInvalidateQueriesListener()

    const serverStatus = useServerStatus()
    const router = useRouter()
    const pathname = usePathname()

    React.useEffect(() => {
        if (!serverStatus?.isOffline && pathname.startsWith("/offline")) {
            router.push("/")
        }
    }, [serverStatus?.isOffline, pathname])

    if (serverStatus?.isOffline) {
        return <LoadingOverlayWithLogo />
    }

    return (
        <>
            <GlobalSearch />
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <PlaylistsModal />
            <ChapterDownloadsDrawer />
            <TorrentStreamOverlay />
            <DebridStreamOverlay />
            <MediaPreviewModal />
            <PlaybackManagerProgressTracking />
            <ManualProgressTracking />
            <IssueReport />
            <ErrorExplainer />
            <SeaCommand />
            <PluginManager />
            {__isElectronDesktop__ && <NativePlayer />}
            <NakamaManager />
            <TopIndefiniteLoader />
            <Announcements />

            <AppSidebarProvider>
                <AppLayout withSidebar sidebarSize="slim">
                    <AppLayoutSidebar>
                        <MainSidebar />
                    </AppLayoutSidebar>
                    <AppLayout>
                        <AppLayoutContent>
                            {children}
                        </AppLayoutContent>
                    </AppLayout>
                </AppLayout>
            </AppSidebarProvider>
        </>
    )
}
