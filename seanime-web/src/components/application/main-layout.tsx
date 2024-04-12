"use client"
import { useAnilistCollectionListener } from "@/app/(main)/_loaders/anilist-collection"
import { useAnilistUserMediaLoader } from "@/app/(main)/_loaders/anilist-user-media"
import { useLibraryCollectionLoader } from "@/app/(main)/_loaders/library-collection"
import { useMangaListener } from "@/app/(main)/_loaders/manga.listeners"
import { useListenToAutoDownloaderItems } from "@/app/(main)/auto-downloader/_lib/autodownloader-items"
import { useListenToMissingEpisodes } from "@/atoms/missing-episodes"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { MainSidebar } from "@/components/application/main-sidebar"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import { WSEvents } from "@/lib/server/endpoints"
import dynamic from "next/dynamic"
import React from "react"
import { toast } from "sonner"

const GlobalSearch = dynamic(() => import("@/components/application/global-search").then((mod) => mod.GlobalSearch))

export const MainLayout = ({ children }: { children: React.ReactNode }) => {

    useLibraryCollectionLoader()
    useListenToMissingEpisodes()
    useListenToAutoDownloaderItems()
    useAnilistUserMediaLoader()
    useAnilistCollectionListener()
    useMangaListener()

    useWebsocketMessageListener<string>({
        type: WSEvents.INFO_TOAST, onMessage: data => {
            if (!!data) {
                toast.info(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.SUCCESS_TOAST, onMessage: data => {
            if (!!data) {
                toast.success(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.WARNING_TOAST, onMessage: data => {
            if (!!data) {
                toast.warning(data)
            }
        },
    })

    useWebsocketMessageListener<string>({
        type: WSEvents.ERROR_TOAST, onMessage: data => {
            if (!!data) {
                toast.error(data)
            }
        },
    })

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
