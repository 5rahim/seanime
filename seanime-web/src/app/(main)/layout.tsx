"use client"
import { MainLayout } from "@/app/(main)/_features/layout/main-layout"
import { OfflineLayout } from "@/app/(main)/_features/layout/offline-layout"
import { TopNavbar } from "@/app/(main)/_features/layout/top-navbar"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ServerDataWrapper } from "@/app/(main)/server-data-wrapper"
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
                    <div data-offline-layout-container className="h-auto">
                        <TopNavbar />
                        <div data-offline-layout-content>
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
                <div data-main-layout-container className="h-auto">
                    <TopNavbar />
                    <div data-main-layout-content>
                        {children}
                    </div>
                </div>
            </MainLayout>
        </ServerDataWrapper>
    )

}


export const dynamic = "force-static"
