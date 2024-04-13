import { OfflineSnapshotProvider } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { serverStatusAtom } from "@/atoms/server-status"
import { OfflineSidebar } from "@/components/application/offline-sidebar"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { AppLayout, AppLayoutContent, AppLayoutSidebar, AppSidebarProvider } from "@/components/ui/app-layout"
import { useAtomValue } from "jotai"
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


    const serverStatus = useAtomValue(serverStatusAtom)
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
        <>
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
        </>
    )
}
