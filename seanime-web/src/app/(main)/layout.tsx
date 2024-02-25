"use client"
import { ScanProgressBar } from "@/app/(main)/(library)/_containers/scanner/scan-progress-bar"
import { ScannerModal } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { useAnilistCollectionListener } from "@/app/(main)/_loaders/anilist-collection"
import { useAnilistUserMediaLoader } from "@/app/(main)/_loaders/anilist-user-media"
import { useLibraryCollectionLoader } from "@/app/(main)/_loaders/library-collection"
import { useListenToAutoDownloaderItems } from "@/app/(main)/auto-downloader/_lib/auto-downloader-items"
import { useListenToMissingEpisodes } from "@/atoms/missing-episodes"
import { DynamicHeaderBackground } from "@/components/application/dynamic-header-background"
import { LibraryWatcher } from "@/components/application/library-watcher"
import { MainLayout } from "@/components/application/main-layout"
import { RefreshAnilistButton } from "@/components/application/refresh-anilist-button"
import { TopNavbar } from "@/components/application/top-navbar"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    useLibraryCollectionLoader()
    useListenToMissingEpisodes()
    useListenToAutoDownloaderItems()
    useAnilistUserMediaLoader()
    useAnilistCollectionListener()

    return (
        <MainLayout>
            <ScanProgressBar />
            <LibraryWatcher />
            <ScannerModal />
            <div className="min-h-screen">
                <div className="w-full md:h-[8rem] relative overflow-hidden pt-[--titlebar-h]">
                    <div
                        className="relative z-10 px-4 w-full flex flex-col md:flex-row justify-between md:items-center">
                        <div className="flex items-center w-full gap-2">
                            <AppSidebarTrigger/>
                            <TopNavbar/>
                        </div>
                        <div className="flex items-center gap-4">
                            <RefreshAnilistButton/>
                        </div>
                    </div>
                    <DynamicHeaderBackground/>
                </div>

                <div>
                    {children}
                </div>

            </div>
        </MainLayout>
    )

}
