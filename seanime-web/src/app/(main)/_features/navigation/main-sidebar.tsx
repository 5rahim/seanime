"use client"
import { useLogout } from "@/api/hooks/auth.hooks"
import { useGetExtensionUpdateData as useGetExtensionUpdateData } from "@/api/hooks/extensions.hooks"
import { useSyncIsActive } from "@/app/(main)/_atoms/sync.atoms"
import { __globalSearch_isOpenAtom } from "@/app/(main)/_features/global-search/global-search"
import { SidebarNavbar } from "@/app/(main)/_features/layout/top-navbar"
import { useOpenSeaCommand } from "@/app/(main)/_features/sea-command/sea-command"
import { UpdateModal } from "@/app/(main)/_features/update/update-modal"
import { useAutoDownloaderQueueCount } from "@/app/(main)/_hooks/autodownloader-queue-count"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useMissingEpisodeCount } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useCurrentUser, useServerStatus, useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { TauriUpdateModal } from "@/app/(main)/_tauri/tauri-update-modal"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { AppSidebar, useAppSidebarContext } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { HoverCard } from "@/components/ui/hover-card"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useDisclosure } from "@/hooks/use-disclosure"
import { openTab } from "@/lib/helpers/browser"
import { ANILIST_OAUTH_URL } from "@/lib/server/config"
import { TORRENT_CLIENT, TORRENT_PROVIDER } from "@/lib/server/settings"
import { WSEvents } from "@/lib/server/ws-events"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useSetAtom } from "jotai"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { BiCalendarAlt, BiDownload, BiExtension, BiLogOut, BiNews } from "react-icons/bi"
import { FaBookReader } from "react-icons/fa"
import { FiLogIn, FiSearch, FiSettings } from "react-icons/fi"
import { HiOutlineServerStack } from "react-icons/hi2"
import { IoCloudOfflineOutline, IoLibrary } from "react-icons/io5"
import { PiArrowCircleLeftDuotone, PiArrowCircleRightDuotone, PiClockCounterClockwiseFill } from "react-icons/pi"
import { SiAnilist } from "react-icons/si"
import { TbWorldDownload } from "react-icons/tb"
import { PluginSidebarTray } from "../plugin/tray/plugin-sidebar-tray"

/**
 * @description
 * - Displays navigation items
 * - Button to logout
 * - Shows count of missing episodes and auto downloader queue
 */
export function MainSidebar() {

    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()

    const [expandedSidebar, setExpandSidebar] = React.useState(false)
    const [dropdownOpen, setDropdownOpen] = React.useState(false)
    // const isCollapsed = !ctx.isBelowBreakpoint && !expandedSidebar
    const isCollapsed = ts.expandSidebarOnHover ? (!ctx.isBelowBreakpoint && !expandedSidebar) : !ctx.isBelowBreakpoint

    const router = useRouter()
    const pathname = usePathname()
    const serverStatus = useServerStatus()
    const setServerStatus = useSetServerStatus()
    const user = useCurrentUser()

    const { setSeaCommandOpen } = useOpenSeaCommand()

    const missingEpisodeCount = useMissingEpisodeCount()
    const autoDownloaderQueueCount = useAutoDownloaderQueueCount()

    // Logout
    const { mutate: logout, data, isPending } = useLogout()

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

    const [activeTorrentCount, setActiveTorrentCount] = React.useState({ downloading: 0, paused: 0, seeding: 0 })
    useWebsocketMessageListener<{ downloading: number, paused: number, seeding: number }>({
        type: WSEvents.ACTIVE_TORRENT_COUNT_UPDATED,
        onMessage: data => {
            setActiveTorrentCount(data)
        },
    })

    const { syncIsActive } = useSyncIsActive()

    const { data: updateData } = useGetExtensionUpdateData()

    return (
        <>
            <AppSidebar
                className={cn(
                    "group/main-sidebar h-full flex flex-col justify-between transition-gpu w-full transition-[width] duration-300",
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
                            // ...[process.env.NEXT_PUBLIC_PLATFORM === "desktop" && {
                            //     iconType: AiOutlineArrowLeft,
                            //     name: "Back",
                            //     onClick: () => {
                            //         router.back()
                            //     },
                            // }],
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
                            ...[serverStatus?.settings?.library?.enableManga && {
                                iconType: FaBookReader,
                                name: "Manga",
                                href: "/manga",
                                isCurrent: pathname.startsWith("/manga"),
                            }],
                            {
                                iconType: BiNews,
                                name: "Discover",
                                href: "/discover",
                                isCurrent: pathname === "/discover",
                            },
                            {
                                iconType: SiAnilist,
                                name: "AniList",
                                href: "/anilist",
                                isCurrent: pathname === "/anilist",
                            },
                            ...[serverStatus?.settings?.library?.torrentProvider !== TORRENT_PROVIDER.NONE && {
                                iconType: TbWorldDownload,
                                name: "Auto Downloader",
                                href: "/auto-downloader",
                                isCurrent: pathname === "/auto-downloader",
                                addon: autoDownloaderQueueCount > 0 ? <Badge
                                    className="absolute right-0 top-0" size="sm"
                                    intent="alert-solid"
                                >{autoDownloaderQueueCount}</Badge> : undefined,
                            }],
                            ...[(
                                serverStatus?.settings?.library?.torrentProvider !== TORRENT_PROVIDER.NONE
                                && !serverStatus?.settings?.torrent?.hideTorrentList
                                && serverStatus?.settings?.torrent?.defaultTorrentClient !== TORRENT_CLIENT.NONE)
                            && {
                                iconType: BiDownload,
                                name: (activeTorrentCount.seeding === 0 || !serverStatus?.settings?.torrent?.showActiveTorrentCount)
                                    ? "Torrent list"
                                    : `Torrent list (${activeTorrentCount.seeding} seeding)`,
                                href: "/torrent-list",
                                isCurrent: pathname === "/torrent-list",
                                addon: ((activeTorrentCount.downloading + activeTorrentCount.paused) > 0 && serverStatus?.settings?.torrent?.showActiveTorrentCount)
                                    ? <Badge
                                        className="absolute right-0 top-0 bg-green-500" size="sm"
                                        intent="alert-solid"
                                    >{activeTorrentCount.downloading + activeTorrentCount.paused}</Badge>
                                    : undefined,
                            }],
                            ...[(serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider) && {
                                iconType: HiOutlineServerStack,
                                name: "Debrid",
                                href: "/debrid",
                                isCurrent: pathname === "/debrid",
                            }],
                            {
                                iconType: PiClockCounterClockwiseFill,
                                name: "Scan summaries",
                                href: "/scan-summaries",
                                isCurrent: pathname === "/scan-summaries",
                            },
                            {
                                iconType: FiSearch,
                                name: "Search",
                                onClick: () => {
                                    ctx.setOpen(false)
                                    setGlobalSearchIsOpen(true)
                                },
                            },
                        ].filter(Boolean)}
                        onLinkItemClick={() => ctx.setOpen(false)}
                    />

                    <SidebarNavbar
                        isCollapsed={isCollapsed}
                        handleExpandSidebar={() => {}}
                        handleUnexpandedSidebar={() => {}}
                    />
                    {process.env.NEXT_PUBLIC_PLATFORM === "desktop" && <div className="w-full flex justify-center px-4">
                        <HoverCard
                            side="right"
                            sideOffset={-8}
                            className="bg-transparent border-none"
                            trigger={<IconButton
                                intent="gray-basic"
                                className="!text-[--muted] hover:!text-[--foreground]"
                                icon={<PiArrowCircleLeftDuotone />}
                                onClick={() => {
                                    router.back()
                                }}
                            />}
                        >
                            <IconButton
                                icon={<PiArrowCircleRightDuotone />}
                                intent="gray-subtle"
                                className="opacity-50 hover:opacity-100"
                                onClick={() => {
                                    router.forward()
                                }}
                            />
                        </HoverCard>
                    </div>}

                    <PluginSidebarTray place="sidebar" />

                </div>
                <div className="flex w-full gap-2 flex-col px-4">
                    {process.env.NEXT_PUBLIC_PLATFORM !== "desktop" ? <UpdateModal collapsed={isCollapsed} /> :
                        <TauriUpdateModal collapsed={isCollapsed} />}
                    <div>
                        <VerticalMenu
                            collapsed={isCollapsed}
                            itemClass="relative"
                            onMouseEnter={() => {}}
                            onMouseLeave={() => {}}
                            onLinkItemClick={() => ctx.setOpen(false)}
                            items={[
                                // {
                                //     iconType: RiSlashCommands2,
                                //     name: "Command palette",
                                //     onClick: () => {
                                //         setSeaCommandOpen(true)
                                //     }
                                // },
                                {
                                    iconType: BiExtension,
                                    name: "Extensions",
                                    href: "/extensions",
                                    isCurrent: pathname.includes("/extensions"),
                                    addon: !!updateData?.length
                                        ? <Badge
                                            className="absolute right-0 top-0 bg-red-500 animate-pulse" size="sm"
                                            intent="alert-solid"
                                        >
                                            {updateData?.length || 1}
                                        </Badge>
                                        : undefined,
                                },
                                {
                                    iconType: IoCloudOfflineOutline,
                                    name: "Offline",
                                    href: "/sync",
                                    isCurrent: pathname.includes("/sync"),
                                    addon: (syncIsActive)
                                        ? <Badge
                                            className="absolute right-0 top-0 bg-blue-500" size="sm"
                                            intent="alert-solid"
                                        >
                                            1
                                        </Badge>
                                        : undefined,
                                },
                                {
                                    iconType: FiSettings,
                                    name: "Settings",
                                    href: "/settings",
                                    isCurrent: pathname === ("/settings"),
                                },
                                ...(ctx.isBelowBreakpoint ? [
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
                                        onClick: () => openTab(ANILIST_OAUTH_URL),
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
                            openTab(ANILIST_OAUTH_URL)
                        }} intent="primary-outline"
                    >Login with AniList</Button>
                </div>
            </Modal>

            <ConfirmationDialog {...confirmSignOut} />
        </>
    )

}
