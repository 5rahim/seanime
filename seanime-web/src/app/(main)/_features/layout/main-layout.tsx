"use client"
import { PlaylistsModal } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner-modal"
import { GlobalSearch } from "@/app/(main)/_features/global-search/global-search"
import { LibraryWatcher } from "@/app/(main)/_features/library-watcher/library-watcher"
import { MainSidebar } from "@/app/(main)/_features/navigation/main-sidebar"
import { useAnimeCollectionLoader } from "@/app/(main)/_hooks/anilist-collection-loader"
import { useAnimeLibraryCollectionLoader } from "@/app/(main)/_hooks/anime-library-collection-loader"
import { useMissingEpisodesLoader } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useAnimeCollectionListener } from "@/app/(main)/_listeners/anilist-collection.listeners"
import { useAutoDownloaderItemListener } from "@/app/(main)/_listeners/autodownloader.listeners"
import { useExtensionListener } from "@/app/(main)/_listeners/extensions.listeners"
import { useMangaListener } from "@/app/(main)/_listeners/manga.listeners"
import { useToastEventListeners } from "@/app/(main)/_listeners/toast-events.listeners"
import { TorrentStreamOverlay } from "@/app/(main)/entry/_containers/torrent-stream/torrent-stream-overlay"
import { ChapterDownloadsDrawer } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import React from "react"

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
    useToastEventListeners()
    useExtensionListener()
    useMangaListener()

    return (
        <>
            <GlobalSearch />
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <PlaylistsModal />
            <ChapterDownloadsDrawer />
            <TorrentStreamOverlay />

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
