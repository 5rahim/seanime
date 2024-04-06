"use client"
import { ProgressTracking } from "@/app/(main)/(library)/_containers/playback-manager/progress-tracking"
import { PlaylistsModal } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scanner/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { useAnilistCollectionListener } from "@/app/(main)/_loaders/anilist-collection"
import { useAnilistUserMediaLoader } from "@/app/(main)/_loaders/anilist-user-media"
import { useLibraryCollectionLoader } from "@/app/(main)/_loaders/library-collection"
import { useListenToAutoDownloaderItems } from "@/app/(main)/auto-downloader/_lib/autodownloader-items"
import { useListenToMissingEpisodes } from "@/atoms/missing-episodes"
import { useWebsocketMessageListener } from "@/atoms/websocket"
import { DynamicHeaderBackground } from "@/components/application/dynamic-header-background"
import { LibraryWatcher } from "@/components/application/library-watcher"
import { MainLayout } from "@/components/application/main-layout"
import { RefreshAnilistButton } from "@/components/application/refresh-anilist-button"
import { TopNavbar } from "@/components/application/top-navbar"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import { WSEvents } from "@/lib/server/endpoints"
import React from "react"
import { toast } from "sonner"

export default function Layout({ children }: { children: React.ReactNode }) {

    useLibraryCollectionLoader()
    useListenToMissingEpisodes()
    useListenToAutoDownloaderItems()
    useAnilistUserMediaLoader()
    useAnilistCollectionListener()

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
        <MainLayout>
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <PlaylistsModal />
            <div className="min-h-screen">
                <div className="w-full h-[5rem] relative overflow-hidden flex items-center">
                    <div className="relative z-10 px-4 w-full flex flex-row justify-between md:items-center">
                        <div className="flex items-center w-full gap-2">
                            <AppSidebarTrigger />
                            <TopNavbar />
                            <ProgressTracking />
                        </div>
                        <div className="flex items-center gap-4">
                            <RefreshAnilistButton />
                        </div>
                    </div>
                    <DynamicHeaderBackground />
                </div>

                <div>
                    {children}
                </div>

            </div>
        </MainLayout>
    )

}
