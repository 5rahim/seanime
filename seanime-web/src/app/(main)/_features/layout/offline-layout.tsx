import { ErrorExplainer } from "@/app/(main)/_features/error-explainer/error-explainer"
import { IssueReport } from "@/app/(main)/_features/issue-report/issue-report"
import { OfflineSidebar } from "@/app/(main)/_features/navigation/offline-sidebar"
import { PluginManager } from "@/app/(main)/_features/plugin/plugin-manager"
import { ManualProgressTracking } from "@/app/(main)/_features/progress-tracking/manual-progress-tracking"
import { PlaybackManagerProgressTracking } from "@/app/(main)/_features/progress-tracking/playback-manager-progress-tracking"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useInvalidateQueriesListener } from "@/app/(main)/_listeners/invalidate-queries.listeners"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { SeaCommand } from "../sea-command/sea-command"

type OfflineLayoutProps = {
    children?: React.ReactNode
}

export function OfflineLayout(props: OfflineLayoutProps) {

    const {
        children,
        ...rest
    } = props


    const serverStatus = useServerStatus()
    const pathname = usePathname()
    const router = useRouter()

    useInvalidateQueriesListener()

    const [cont, setContinue] = React.useState(false)

    React.useEffect(() => {
        if (!serverStatus?.isOffline) {
            setContinue(false)
            return
        }

        if (
            pathname.startsWith("/offline") ||
            pathname.startsWith("/settings") ||
            pathname.startsWith("/mediastream") ||
            pathname.startsWith("/medialinks")
        ) {
            setContinue(true)
            return
        }

        router.push("/offline")
    }, [pathname])

    if (!cont) return <LoadingOverlayWithLogo />

    return (
        <>
            <PlaybackManagerProgressTracking />
            <ManualProgressTracking />
            <IssueReport />
            <ErrorExplainer />
            <SeaCommand />
            <PluginManager />

            <AppSidebarProvider>
                <AppLayout withSidebar sidebarSize="slim">
                    <AppLayoutSidebar>
                        <OfflineSidebar />
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
