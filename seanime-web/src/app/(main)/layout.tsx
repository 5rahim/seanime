"use client"
import { OfflineTopNavbar } from "@/app/(main)/(offline)/offline/_components/offline-top-navbar"
import { LayoutHeaderBackground } from "@/app/(main)/_features/layout/_components/layout-header-background"
import { MainLayout } from "@/app/(main)/_features/layout/main-layout"
import { OfflineLayout } from "@/app/(main)/_features/layout/offline-layout"
import { TopNavbar } from "@/app/(main)/_features/layout/top-navbar"
import { PlaybackManagerProgressTracking } from "@/app/(main)/_features/progress-tracking/playback-manager-progress-tracking"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ServerDataWrapper } from "@/app/(main)/server-data-wrapper"
import { AppSidebarTrigger } from "@/components/ui/app-layout"
import React from "react"

export default function Layout({ children }: { children: React.ReactNode }) {

    const serverStatus = useServerStatus()

    const [host, setHost] = React.useState<string>("")

    React.useEffect(() => {
        setHost(window?.location?.host || "")
    }, [])

    if (serverStatus?.isOffline) {
        return (
            <ServerDataWrapper host={host}>
                <OfflineLayout>
                    <div className="h-auto">
                        <div className="w-full h-[5rem] relative overflow-hidden flex items-center">
                            <div className="relative z-10 px-4 w-full flex flex-row justify-between md:items-center">
                                <div className="flex items-center w-full gap-2">
                                    <AppSidebarTrigger />
                                    <OfflineTopNavbar />
                                    <PlaybackManagerProgressTracking />
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
        <ServerDataWrapper host={host}>
            <MainLayout>
                <div className="h-auto">
                    <TopNavbar />

                    <div>
                        {children}
                    </div>

                </div>
            </MainLayout>
        </ServerDataWrapper>
    )

}


export const dynamic = "force-static"
