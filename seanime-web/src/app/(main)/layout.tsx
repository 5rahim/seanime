"use client"
import { MainLayout } from "@/components/application/main-layout"
import React from "react"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import { TopNavbar } from "@/components/application/top-navbar"
import { RefreshAnilistButton } from "@/components/application/refresh-anilist-button"
import { DynamicHeaderBackground } from "@/components/application/dynamic-header-background"
import { useAtomicLibraryCollectionLoader } from "@/atoms/collection"
import { useAnilistCollectionListener } from "@/lib/server/hooks/media"
import { useListenToMissingEpisodes } from "@/atoms/missing-episodes"

export default function Layout({ children }: { children: React.ReactNode }) {

    useAtomicLibraryCollectionLoader()
    useListenToMissingEpisodes()
    // Listen to refresh events
    useAnilistCollectionListener()

    return (
        <MainLayout>
            <div className="min-h-screen">
                <div className={"w-full md:h-[8rem] relative overflow-hidden pt-[--titlebar-h]"}>
                    <div
                        className="relative z-10 px-4 w-full flex flex-col md:flex-row justify-between md:items-center">
                        <div className={"flex items-center w-full gap-2"}>
                            <AppSidebarTrigger/>
                            <TopNavbar/>
                        </div>
                        <div className={"flex items-center gap-4"}>
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