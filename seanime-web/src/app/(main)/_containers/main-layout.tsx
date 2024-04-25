"use client"
import { PlaylistsModal } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner-modal"
import { GlobalSearch } from "@/app/(main)/_containers/global-search"
import { MainSidebar } from "@/app/(main)/_containers/main-sidebar"
import { useAnilistCollectionLoader } from "@/app/(main)/_hooks/anilist-collection.hooks"
import { useLibraryCollectionLoader } from "@/app/(main)/_hooks/anime-library.hooks"
import { useMissingEpisodesLoader } from "@/app/(main)/_hooks/missing-episodes.hooks"
import { useAnilistCollectionListener } from "@/app/(main)/_listeners/anilist-collection.listeners"
import { useAutoDownloaderItemListener } from "@/app/(main)/_listeners/autodownloader.listeners"
import { useMangaListener } from "@/app/(main)/_listeners/manga.listeners"
import { useToastEventListeners } from "@/app/(main)/_listeners/toast-events.listeners"
import { ChapterDownloadsDrawer } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { LibraryWatcher } from "@/components/application/library-watcher"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import React from "react"

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    /**
     * Data loaders
     */
    useLibraryCollectionLoader()
    useAnilistCollectionLoader()
    useMissingEpisodesLoader()

    /**
     * Websocket listeners
     */
    useAutoDownloaderItemListener()
    useAnilistCollectionListener()
    useToastEventListeners()
    useMangaListener()

    return (
        <>
            <GlobalSearch />
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <PlaylistsModal />
            <ChapterDownloadsDrawer />

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
