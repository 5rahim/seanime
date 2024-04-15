"use client"
import { ProgressTracking } from "@/app/(main)/(library)/_containers/playback-manager/progress-tracking"
import { PlaylistsModal } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scanner/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { ChapterDownloadsButton } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-button"
import { ChapterDownloadsDrawer } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-drawer"
import { serverStatusAtom } from "@/atoms/server-status"
import { DynamicHeaderBackground } from "@/components/application/dynamic-header-background"
import { LibraryWatcher } from "@/components/application/library-watcher"
import { MainLayout } from "@/components/application/main-layout"
import { OfflineLayout } from "@/components/application/offline-layout"
import { OfflineTopNavbar } from "@/components/application/offline-top-navbar"
import { RefreshAnilistButton } from "@/components/application/refresh-anilist-button"
import { TopNavbar } from "@/components/application/top-navbar"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import { useAtomValue } from "jotai"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    const serverStatus = useAtomValue(serverStatusAtom)

    if (serverStatus?.isOffline) {
        return (
            <OfflineLayout>
                <div className="min-h-screen">
                    <div className="w-full h-[5rem] relative overflow-hidden flex items-center">
                        <div className="relative z-10 px-4 w-full flex flex-row justify-between md:items-center">
                            <div className="flex items-center w-full gap-2">
                                <AppSidebarTrigger />
                                <OfflineTopNavbar />
                                <ProgressTracking />
                            </div>
                        </div>
                        <DynamicHeaderBackground />
                    </div>

                    <div>
                        {children}
                    </div>
                </div>
            </OfflineLayout>
        )
    }

    return (
        <MainLayout>
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <PlaylistsModal />
            <ChapterDownloadsDrawer />
            <div className="min-h-screen">
                <div className="w-full h-[5rem] relative overflow-hidden flex items-center">
                    <div className="relative z-10 px-4 w-full flex flex-row justify-between md:items-center">
                        <div className="flex items-center w-full gap-2">
                            <AppSidebarTrigger />
                            <TopNavbar />
                            <ProgressTracking />
                        </div>
                        <div className="flex items-center gap-4">
                            <ChapterDownloadsButton />
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


export const dynamic = "force-static"
