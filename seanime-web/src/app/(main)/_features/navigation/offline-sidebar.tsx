"use client"
import { useSetOfflineMode } from "@/api/hooks/local.hooks"
import { SidebarNavbar } from "@/app/(main)/_features/layout/top-navbar"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppSidebar, useAppSidebarContext } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { cn } from "@/components/ui/core/styling"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useThemeSettings } from "@/lib/theme/hooks"
import { usePathname } from "next/navigation"
import React from "react"
import { FaBookReader } from "react-icons/fa"
import { FiSettings } from "react-icons/fi"
import { IoCloudyOutline, IoLibrary } from "react-icons/io5"
import { PluginSidebarTray } from "../plugin/tray/plugin-sidebar-tray"


export function OfflineSidebar() {
    const serverStatus = useServerStatus()
    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()

    const [expandedSidebar, setExpandSidebar] = React.useState(false)
    const isCollapsed = ts.expandSidebarOnHover ? (!ctx.isBelowBreakpoint && !expandedSidebar) : !ctx.isBelowBreakpoint

    const { mutate: setOfflineMode, isPending: isSettingOfflineMode } = useSetOfflineMode()

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

    const confirmDialog = useConfirmationDialog({
        title: "Disable offline mode",
        description: "Are you sure you want to disable offline mode?",
        actionText: "Yes",
        actionIntent: "primary",
        onConfirm: () => {
            setOfflineMode({ enabled: false })
        },
    })

    return (
        <>
            <AppSidebar
                className={cn(
                    "h-full flex flex-col justify-between transition-gpu w-full transition-[width]",
                    (!ctx.isBelowBreakpoint && expandedSidebar) && "w-[260px]",
                    (!ctx.isBelowBreakpoint && !ts.disableSidebarTransparency) && "bg-transparent",
                    (!ctx.isBelowBreakpoint && !ts.disableSidebarTransparency && ts.expandSidebarOnHover) && "hover:bg-[--background]",
                )}
                onMouseEnter={handleExpandSidebar}
                onMouseLeave={handleUnexpandedSidebar}
            >
                {(!ctx.isBelowBreakpoint && ts.expandSidebarOnHover && ts.disableSidebarTransparency) && <div
                    className={cn(
                        "fixed h-full translate-x-0 w-[50px] bg-gradient bg-gradient-to-r via-[--background] from-[--background] to-transparent",
                        "group-hover/main-sidebar:translate-x-[250px] transition opacity-0 duration-300 group-hover/main-sidebar:opacity-100",
                    )}
                ></div>}


                <div>
                    <div className="mb-4 p-4 pb-0 flex justify-center w-full">
                        <img src="/logo.png" alt="logo" className="w-15 h-10" />
                    </div>
                    <VerticalMenu
                        className="px-4"
                        collapsed={isCollapsed}
                        itemClass="relative"
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
                                href: "/offline/manga",
                                isCurrent: pathname.startsWith("/offline/manga"),
                            }].filter(Boolean) as any,
                        ].filter(Boolean)}
                        onLinkItemClick={() => ctx.setOpen(false)}
                    />

                    <SidebarNavbar
                        isCollapsed={isCollapsed}
                        handleExpandSidebar={() => { }}
                        handleUnexpandedSidebar={() => { }}
                    />

                    <PluginSidebarTray place="sidebar" />
                </div>
                <div className="flex w-full gap-2 flex-col px-4">
                    <div>
                        <VerticalMenu
                            collapsed={isCollapsed}
                            itemClass="relative"
                            onLinkItemClick={() => ctx.setOpen(false)}
                            items={[
                                {
                                    iconType: IoCloudyOutline,
                                    name: "Disable offline mode",
                                    onClick: () => {
                                        confirmDialog.open()
                                    },
                                },
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
                            />
                            {expandedSidebar && <p className="truncate">Offline</p>}
                        </div>
                    </div>
                </div>
            </AppSidebar>
            <ConfirmationDialog
                {...confirmDialog}
            />
        </>
    )

}
