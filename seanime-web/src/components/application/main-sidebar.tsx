"use client"
import { useAutoDownloaderQueueCount } from "@/app/(main)/auto-downloader/_lib/autodownloader-items"
import { useMissingEpisodeCount } from "@/atoms/missing-episodes"
import { serverStatusAtom } from "@/atoms/server-status"
import { useCurrentUser } from "@/atoms/user"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/application/confirmation-dialog"
import { __globalSearch_isOpenAtom } from "@/components/application/global-search"
import { UpdateModal } from "@/components/application/update-modal"
import { AppSidebar, useAppSidebarContext } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useDisclosure } from "@/hooks/use-disclosure"
import { ANILIST_OAUTH_URL } from "@/lib/anilist/config"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useThemeSettings } from "@/lib/theme/hooks"
import { ServerStatus } from "@/lib/types/server-status.types"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React from "react"
import { BiCalendarAlt, BiChart, BiCollection, BiDownload, BiLogOut } from "react-icons/bi"
import { FaBookReader, FaRssSquare } from "react-icons/fa"
import { FiLogIn, FiSearch, FiSettings } from "react-icons/fi"
import { IoLibrary } from "react-icons/io5"
import { LuLayoutDashboard } from "react-icons/lu"
import { MdSyncAlt } from "react-icons/md"
import { PiClockCounterClockwiseFill } from "react-icons/pi"
import { SiMyanimelist } from "react-icons/si"


export function MainSidebar() {

    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()

    const [expandedSidebar, setExpandSidebar] = React.useState(false)
    const [dropdownOpen, setDropdownOpen] = React.useState(false)
    // const isCollapsed = !ctx.isBelowBreakpoint && !expandedSidebar
    const isCollapsed = ts.expandSidebarOnHover ? (!ctx.isBelowBreakpoint && !expandedSidebar) : !ctx.isBelowBreakpoint

    const { user } = useCurrentUser()
    const pathname = usePathname()
    const [serverStatus, setServerStatus] = useAtom(serverStatusAtom)

    const missingEpisodeCount = useMissingEpisodeCount()
    const autoDownloaderQueueCount = useAutoDownloaderQueueCount()

    // Logout
    const { mutate: logout, data, isPending } = useSeaMutation<ServerStatus>({
        endpoint: SeaEndpoints.LOGOUT,
        mutationKey: ["logout"],
    })

    React.useEffect(() => {
        if (!isPending) {
            setServerStatus(data)
        }
    }, [isPending, data])

    const setGlobalSearchIsOpen = useSetAtom(__globalSearch_isOpenAtom)

    const loginModal = useDisclosure(false)

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

    const confirmSignOut = useConfirmationDialog({
        title: "Sign out",
        description: "Are you sure you want to sign out?",
        onConfirm: () => {
            logout()
        },
    })

    // shelved
    // React.useEffect(() => {
    //     let r = document.querySelector(":root") as any
    //
    //     r.style.setProperty("--background", ts.backgroundColor)
    //     r.style.setProperty("--paper", ts.backgroundColor !== "#0c0c0c" ? "rgba(0,0,0,0.2)" : "#101010")
    // }, [ts.backgroundColor])

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
                                href: "/",
                                isCurrent: pathname === "/",
                            },
                            {
                                iconType: BiCalendarAlt,
                                name: "Schedule",
                                href: "/schedule",
                                isCurrent: pathname === "/schedule",
                                addon: missingEpisodeCount > 0 ? <Badge
                                    className="absolute right-0 top-0" size="sm"
                                    intent="alert-solid"
                                >{missingEpisodeCount}</Badge> : undefined,
                            },
                            {
                                iconType: BiCollection,
                                name: "My lists",
                                href: "/anilist",
                                isCurrent: pathname === "/anilist",
                            },
                            ...[serverStatus?.settings?.library?.enableManga && {
                                iconType: FaBookReader,
                                name: "Manga",
                                href: "/manga",
                                isCurrent: pathname.startsWith("/manga"),
                            }].filter(Boolean) as any,
                            {
                                iconType: BiChart,
                                name: "Discover",
                                href: "/discover",
                                isCurrent: pathname === "/discover",
                            },
                            {
                                iconType: FaRssSquare,
                                name: "Auto downloader",
                                href: "/auto-downloader",
                                isCurrent: pathname === "/auto-downloader",
                                addon: autoDownloaderQueueCount > 0 ? <Badge
                                    className="absolute right-0 top-0" size="sm"
                                    intent="alert-solid"
                                >{autoDownloaderQueueCount}</Badge> : undefined,
                            },
                            {
                                iconType: BiDownload,
                                name: "Torrent list",
                                href: "/torrent-list",
                                isCurrent: pathname === "/torrent-list",
                            },
                            {
                                iconType: PiClockCounterClockwiseFill,
                                name: "Scan summaries",
                                href: "/scan-summaries",
                                isCurrent: pathname === "/scan-summaries",
                            },
                            {
                                iconType: MdSyncAlt,
                                name: "List sync",
                                href: "/list-sync",
                                isCurrent: pathname === "/list-sync",
                            },
                            {
                                iconType: FiSearch,
                                name: "Search",
                                onClick: () => setGlobalSearchIsOpen(true),
                            },
                        ].filter(Boolean)}
                        onLinkItemClick={() => ctx.setOpen(false)}
                    />
                </div>
                <div className="flex w-full gap-2 flex-col px-4">
                    <UpdateModal collapsed={isCollapsed} />
                    <div>
                        <VerticalMenu
                            collapsed={isCollapsed}
                            itemClass="relative"
                            onMouseEnter={handleExpandSidebar}
                            onMouseLeave={handleUnexpandedSidebar}
                            onLinkItemClick={() => ctx.setOpen(false)}
                            items={[
                                {
                                    iconType: LuLayoutDashboard,
                                    name: "UI Settings",
                                    href: "/settings/ui",
                                    isCurrent: pathname.includes("/settings/ui"),
                                },
                                {
                                    iconType: FiSettings,
                                    name: "Settings",
                                    href: "/settings",
                                    isCurrent: pathname === ("/settings"),
                                },
                                ...(ctx.isBelowBreakpoint ? [
                                    {
                                        iconType: SiMyanimelist,
                                        name: "MyAnimeList",
                                        href: "/mal",
                                    },
                                    {
                                        iconType: BiLogOut,
                                        name: "Sign out",
                                        onClick: confirmSignOut.open,
                                    },
                                ] : []),
                            ]}
                        />
                    </div>
                    {!user && (
                        <div>
                            <VerticalMenu
                                collapsed={isCollapsed}
                                itemClass="relative"
                                onMouseEnter={handleExpandSidebar}
                                onMouseLeave={handleUnexpandedSidebar}
                                onLinkItemClick={() => ctx.setOpen(false)}
                                items={[
                                    {
                                        iconType: FiLogIn,
                                        name: "Login",
                                        onClick: () => window.open(ANILIST_OAUTH_URL, "_self"),
                                    },
                                ]}
                            />
                        </div>
                    )}
                    {!!user && <div className="flex w-full gap-2 flex-col">
                        <DropdownMenu
                            trigger={<div
                                className={cn(
                                    "w-full flex p-2.5 pt-1 items-center space-x-2",
                                    { "hidden": ctx.isBelowBreakpoint },
                                )}
                            >
                                <Avatar size="sm" className="cursor-pointer" src={user?.avatar?.medium || ""} />
                                {expandedSidebar && <p className="truncate">{user?.name}</p>}
                            </div>}
                            open={dropdownOpen}
                            onOpenChange={setDropdownOpen}
                        >
                            <Link href="/mal">
                                <DropdownMenuItem>
                                    <SiMyanimelist className="text-lg text-indigo-200" /> MyAnimeList
                                </DropdownMenuItem>
                            </Link>
                            <DropdownMenuItem onClick={confirmSignOut.open}>
                                <BiLogOut /> Sign out
                            </DropdownMenuItem>
                        </DropdownMenu>
                    </div>}
                </div>
            </AppSidebar>

            <Modal
                title="Login"
                open={loginModal.isOpen}
                onOpenChange={loginModal.close}
            >
                <div className="mt-5 text-center space-y-4">
                    <Button
                        onClick={() => {
                            window.open(ANILIST_OAUTH_URL)
                        }} intent="primary-outline"
                    >Login with AniList</Button>
                </div>
            </Modal>

            <ConfirmationDialog {...confirmSignOut} />
        </>
    )

}
