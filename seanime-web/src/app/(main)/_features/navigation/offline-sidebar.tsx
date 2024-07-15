"use client"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AppSidebar, useAppSidebarContext } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { cn } from "@/components/ui/core/styling"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useThemeSettings } from "@/lib/theme/hooks"
import { usePathname } from "next/navigation"
import React from "react"
import { FaBookReader } from "react-icons/fa"
import { FiSettings } from "react-icons/fi"
import { IoLibrary } from "react-icons/io5"


export function OfflineSidebar() {
    const serverStatus = useServerStatus()
    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()
    const { snapshot } = useOfflineSnapshot()

    const [expandedSidebar, setExpandSidebar] = React.useState(false)
    const isCollapsed = ts.expandSidebarOnHover ? (!ctx.isBelowBreakpoint && !expandedSidebar) : !ctx.isBelowBreakpoint

    const pathname = usePathname()


    const handleExpandSidebar = () => {
        if (!ctx.isBelowBreakpoint && ts.expandSidebarOnHover) {
            setExpandSidebar(true)
        }
    }
    const handleUnexpandedSidebar = () => {
        if (expandedSidebar && ts.expandSidebarOnHover) {
            setExpandSidebar(false)
        }
    }

    return (
        <>
            <AppSidebar
                className={cn(
                    "h-full flex flex-col justify-between transition-gpu w-full transition-[width]",
                    { "w-[400px]": !ctx.isBelowBreakpoint && expandedSidebar },
                )}
                // sidebarClass="h-full"
            >
                <div>
                    <div className="mb-4 p-4 pb-0 flex justify-center w-full">
                        <img src="/logo.png" alt="logo" className="w-15 h-10" />
                    </div>
                    <VerticalMenu
                        className="px-4"
                        collapsed={isCollapsed}
                        itemClass="relative"
                        onMouseEnter={handleExpandSidebar}
                        onMouseLeave={handleUnexpandedSidebar}
                        items={[
                            {
                                iconType: IoLibrary,
                                name: "Library",
                                href: "/offline",
                                isCurrent: pathname === "/offline",
                            },
                            ...[serverStatus?.settings?.library?.enableManga && {
                                iconType: FaBookReader,
                                name: "Manga",
                                href: "/offline#manga",
                                isCurrent: pathname.startsWith("/offline#manga"),
                            }].filter(Boolean) as any,
                        ].filter(Boolean)}
                        onLinkItemClick={() => ctx.setOpen(false)}
                    />
                </div>
                <div className="flex w-full gap-2 flex-col px-4">
                    <div>
                        <VerticalMenu
                            collapsed={isCollapsed}
                            itemClass="relative"
                            onMouseEnter={handleExpandSidebar}
                            onMouseLeave={handleUnexpandedSidebar}
                            onLinkItemClick={() => ctx.setOpen(false)}
                            items={[
                                {
                                    iconType: FiSettings,
                                    name: "Settings",
                                    href: "/settings",
                                    isCurrent: pathname === ("/settings"),
                                },
                            ]}
                        />
                    </div>
                    <div className="flex w-full gap-2 flex-col">
                        <div
                            className={cn(
                                "w-full flex p-2.5 pt-1 items-center space-x-2",
                                { "hidden": ctx.isBelowBreakpoint },
                            )}
                        >
                            <Avatar
                                size="sm"
                                className="cursor-pointer"
                                src={offline_getAssetUrl(snapshot?.user?.viewer?.avatar?.medium, snapshot?.assetMap) || ""}
                            />
                            {expandedSidebar && <p className="truncate">{snapshot?.user?.viewer?.name}</p>}
                        </div>
                    </div>
                </div>
            </AppSidebar>
        </>
    )

}
