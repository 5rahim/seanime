import { OfflineSnapshotProvider } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { OfflineSidebar } from "@/app/(main)/_features/navigation/offline-sidebar"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import { usePathname, useRouter } from "next/navigation"
import React from "react"

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

    const [cont, setContinue] = React.useState(false)

    React.useEffect(() => {
        if (!serverStatus?.isOffline) {
            setContinue(false)
            return
        }
        if (pathname.startsWith("/offline") && !pathname.startsWith("/offline-mode")) {
            setContinue(true)
            return
        }
        if (pathname.startsWith("/settings")) {
            setContinue(true)
            return
        }

        router.push("/offline")
    }, [pathname])

    if (!cont) return <LoadingOverlayWithLogo />

    return (
        <OfflineSnapshotProvider>
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
        </OfflineSnapshotProvider>
    )
}
