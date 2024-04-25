"use client"
import { ProgressTracking } from "@/app/(main)/(library)/_containers/playback-manager/progress-tracking"
import { OfflineTopNavbar } from "@/app/(main)/(offline)/offline/_components/offline-top-navbar"
import { RefreshAnilistButton } from "@/app/(main)/_containers/refresh-anilist-button"
import { LayoutHeaderBackground } from "@/app/(main)/_features/layout/_components/layout-header-background"
import { MainLayout } from "@/app/(main)/_features/layout/main-layout"
import { OfflineLayout } from "@/app/(main)/_features/layout/offline-layout"
import { TopMenu } from "@/app/(main)/_features/navigation/top-menu"
import { useServerStatus } from "@/app/(main)/_hooks/server-status.hooks"
import { ChapterDownloadsButton } from "@/app/(main)/manga/_containers/chapter-downloads/chapter-downloads-button"
import { ServerDataWrapper } from "@/app/(main)/server-data-wrapper"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    const serverStatus = useServerStatus()

    if (serverStatus?.isOffline) {
        return (
            <ServerDataWrapper>
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
                            <LayoutHeaderBackground />
                        </div>

                        <div>
                            {children}
                        </div>
                    </div>
                </OfflineLayout>
            </ServerDataWrapper>
        )
    }

    return (
        <ServerDataWrapper>
            <MainLayout>
                <div className="min-h-screen">
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

                    <div>
                        {children}
                    </div>

                </div>
            </MainLayout>
        </ServerDataWrapper>
    )

}


export const dynamic = "force-static"
