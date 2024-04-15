"use client"
import { useAnilistUserMediaLoader } from "@/app/(main)/_lib/anilist-user-media"
import { useLibraryCollectionLoader } from "@/app/(main)/_lib/anime-library-collection"
import { useAnilistCollectionListener } from "@/app/(main)/_listeners/anilist-collection.listeners"
import { useMangaListener } from "@/app/(main)/_listeners/manga.listeners"
import { useToastEventListeners } from "@/app/(main)/_listeners/toast-events.listeners"
import { useAutoDownloaderItemListener } from "@/app/(main)/auto-downloader/_lib/autodownloader-items"
import { useMissingEpisodeListener } from "@/atoms/missing-episodes"
import { GlobalSearch } from "@/components/application/global-search"
import { MainSidebar } from "@/components/application/main-sidebar"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import React from "react"

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    useLibraryCollectionLoader()
    useAnilistUserMediaLoader()

    useMissingEpisodeListener()
    useAutoDownloaderItemListener()
    useAnilistCollectionListener()
    useMangaListener()
    useToastEventListeners()

    return (
        <>
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
            <GlobalSearch />
        </>
    )
}
