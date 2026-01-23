"use client"
import { useRefreshAnimeCollection } from "@/api/hooks/anilist.hooks"
import { useLogout } from "@/api/hooks/auth.hooks"
import { useGetExtensionUpdateData as useGetExtensionUpdateData, usePluginWithIssuesCount } from "@/api/hooks/extensions.hooks"
import { isLoginModalOpenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { useSyncIsActive } from "@/app/(main)/_atoms/sync.atoms"
import { ElectronUpdateModal } from "@/app/(main)/_electron/electron-update-modal"
import { __globalSearch_isOpenAtom } from "@/app/(main)/_features/global-search/global-search"
import { SidebarNavbar } from "@/app/(main)/_features/layout/top-navbar"
import { usePluginSidebarItems } from "@/app/(main)/_features/plugin/webview/plugin-sidebar"
import { useSeaCommand } from "@/app/(main)/_features/sea-command/sea-command"
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
import { defineSchema, Field, Form } from "@/components/ui/form"
import { HoverCard } from "@/components/ui/hover-card"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu, VerticalMenuItem } from "@/components/ui/vertical-menu"
import { openTab } from "@/lib/helpers/browser"
import { ANILIST_OAUTH_URL, ANILIST_PIN_URL } from "@/lib/server/config"
import { TORRENT_CLIENT, TORRENT_PROVIDER } from "@/lib/server/settings"
import { WSEvents } from "@/lib/server/ws-events"
import { useThemeSettings } from "@/lib/theme/hooks"
import { __isDesktop__, __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { useAtom, useSetAtom } from "jotai"
import Link from "next/link"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { BiChevronRight, BiExtension, BiLogIn, BiLogOut } from "react-icons/bi"
import { FiLogIn, FiSearch } from "react-icons/fi"
import { HiOutlineServerStack } from "react-icons/hi2"
import { IoCloudOfflineOutline, IoHomeOutline } from "react-icons/io5"
import { LuBookOpen, LuCalendar, LuCompass, LuRefreshCw, LuRss, LuSettings } from "react-icons/lu"
import { MdOutlineConnectWithoutContact } from "react-icons/md"
import { PiArrowCircleLeftDuotone, PiArrowCircleRightDuotone } from "react-icons/pi"
import { RiListCheck3 } from "react-icons/ri"
import { SiQbittorrent, SiTransmission } from "react-icons/si"
import { TbReportSearch } from "react-icons/tb"
import { nakamaModalOpenAtom, useNakamaStatus } from "../nakama/nakama-manager"
import { PluginSidebarTray } from "../plugin/tray/plugin-sidebar-tray"

export function MainSidebar() {

    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()

    const [expandedSidebar, setExpandSidebar] = React.useState(false)
    const isCollapsed = ts.expandSidebarOnHover ? (!ctx.isBelowBreakpoint && !expandedSidebar) : !ctx.isBelowBreakpoint

    const containerRef = React.useRef<HTMLDivElement>(null)

    // Logout
    const setServerStatus = useSetServerStatus()
    const { mutate: logout, data, isPending } = useLogout()

    React.useEffect(() => {
        if (!isPending) {
            setServerStatus(data)
        }
    }, [isPending, data])


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
                ref={containerRef}
                className={cn(
                    "group/main-sidebar h-full flex flex-col justify-between transition-gpu w-full transition-[width] duration-300 overflow-x-hidden",
                    // Enable scrolling but hide the scrollbar
                    "overflow-y-auto [&::-webkit-scrollbar]:hidden [-ms-overflow-style:'none'] [scrollbar-width:'none']",
                    (!ctx.isBelowBreakpoint && expandedSidebar) && "w-[260px]",
                    (!ctx.isBelowBreakpoint && !ts.disableSidebarTransparency) && "bg-transparent",
                    (!ctx.isBelowBreakpoint && !ts.disableSidebarTransparency && ts.expandSidebarOnHover && expandedSidebar) && "bg-[--background] rounded-tr-xl rounded-br-xl border-[--border]",
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

                <SidebarNavigation
                    isCollapsed={isCollapsed}
                    containerRef={containerRef}
                />

                <div className="flex w-full gap-2 flex-col px-4 shrink-0 pb-2">
                    <SidebarUpdates isCollapsed={isCollapsed} />
                    <SidebarFooter isCollapsed={isCollapsed} onLogout={logout} />
                    <SidebarUser expandedSidebar={expandedSidebar} onLogout={logout} isCollapsed={isCollapsed} />
                </div>
            </AppSidebar>
        </>
    )

}


function SidebarNavigation({ isCollapsed, containerRef }: { isCollapsed: boolean, containerRef: React.RefObject<HTMLDivElement> }) {
    const ctx = useAppSidebarContext()
    const ts = useThemeSettings()
    const router = useRouter()
    const pathname = usePathname()
    const serverStatus = useServerStatus()

    // Commands
    const { setSeaCommandOpen } = useSeaCommand()
    const setGlobalSearchIsOpen = useSetAtom(__globalSearch_isOpenAtom)

    // Data
    const missingEpisodeCount = useMissingEpisodeCount()
    const autoDownloaderQueueCount = useAutoDownloaderQueueCount()

    // Torrents
    const [activeTorrentCount, setActiveTorrentCount] = React.useState({ downloading: 0, paused: 0, seeding: 0 })
    useWebsocketMessageListener<{ downloading: number, paused: number, seeding: number }>({
        type: WSEvents.ACTIVE_TORRENT_COUNT_UPDATED,
        onMessage: data => {
            setActiveTorrentCount(data)
        },
    })

    // Refresh AniList
    const { mutate: refreshAC, isPending: isRefreshingAC } = useRefreshAnimeCollection()

    // Items
    const items = React.useMemo(() => [
        {
            id: "home",
            iconType: IoHomeOutline,
            name: "Home",
            href: "/",
            isCurrent: pathname === "/",
        },
        // ...(process.env.NODE_ENV === "development" ? [{
        //     id: "test",
        //     iconType: GrTest,
        //     name: "Test",
        //     href: "/test",
        //     isCurrent: pathname === "/test",
        // }] : []),
        {
            id: "schedule",
            iconType: LuCalendar,
            name: "Schedule",
            href: "/schedule",
            isCurrent: pathname === "/schedule",
            addon: missingEpisodeCount > 0 ? <Badge
                className="absolute right-0 top-0" size="sm"
                intent="alert-solid"
            >{missingEpisodeCount}</Badge> : undefined,
        },
        ...serverStatus?.settings?.library?.enableManga ? [{
            id: "manga",
            iconType: LuBookOpen,
            name: "Manga",
            href: "/manga",
            isCurrent: pathname.startsWith("/manga"),
        }] : [],
        {
            id: "lists",
            iconType: RiListCheck3,
            name: "My lists",
            href: "/lists",
            isCurrent: pathname === "/lists",
        },
        {
            id: "discover",
            iconType: LuCompass,
            name: "Discover",
            href: "/discover",
            isCurrent: pathname === "/discover",
        },
        ...(
            serverStatus?.settings?.library?.torrentProvider !== TORRENT_PROVIDER.NONE
            && serverStatus?.settings?.torrent?.defaultTorrentClient !== TORRENT_CLIENT.NONE)
            ? [{
                id: "torrent-list",
                iconType: serverStatus?.settings?.torrent?.defaultTorrentClient === TORRENT_CLIENT.QBITTORRENT ? SiQbittorrent : SiTransmission,
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
            }] : [],
        ...(serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider) ? [{
            id: "debrid",
            iconType: HiOutlineServerStack,
            name: "Debrid",
            href: "/debrid",
            isCurrent: pathname === "/debrid",
        }] : [],
        ...(!!serverStatus?.settings?.library?.libraryPath) ? [{
            id: "scan-summaries",
            iconType: TbReportSearch,
            name: "Scan summaries",
            href: "/scan-summaries",
            isCurrent: pathname === "/scan-summaries",
        }] : [],
        ...(serverStatus?.settings?.library?.torrentProvider !== TORRENT_PROVIDER.NONE && !!serverStatus?.settings?.library?.libraryPath) ? [{
            id: "auto-downloader",
            iconType: LuRss,
            name: "Auto Downloader",
            href: "/auto-downloader",
            isCurrent: pathname === "/auto-downloader",
            addon: autoDownloaderQueueCount > 0 ? <Badge
                className="absolute right-0 top-0" size="sm"
                intent="alert-solid"
            >{autoDownloaderQueueCount}</Badge> : undefined,
        }] : [],
        {
            id: "search",
            iconType: FiSearch,
            name: "Search",
            onClick: () => {
                ctx.setOpen(false)
                setGlobalSearchIsOpen(true)
            },
        },
    ], [
        pathname,
        missingEpisodeCount,
        serverStatus?.settings?.library?.enableManga,
        serverStatus?.settings?.library?.torrentProvider,
        serverStatus?.settings?.torrent?.defaultTorrentClient,
        serverStatus?.settings?.torrent?.showActiveTorrentCount,
        serverStatus?.debridSettings?.enabled,
        serverStatus?.debridSettings?.provider,
        serverStatus?.settings?.library?.libraryPath,
        activeTorrentCount.seeding,
        activeTorrentCount.downloading,
        activeTorrentCount.paused,
        autoDownloaderQueueCount,
    ])

    // Plugins
    const pluginWebviewItems = usePluginSidebarItems()

    // Overflow logic
    const [autoUnpinnedIds, setAutoUnpinnedIds] = React.useState<string[]>([])
    const overflowCheckTimeoutRef = React.useRef<NodeJS.Timeout>()

    React.useEffect(() => {
        const handleResize = () => setAutoUnpinnedIds([])
        window.addEventListener("resize", handleResize)
        return () => window.removeEventListener("resize", handleResize)
    }, [])

    const allPinnedItems = React.useMemo(() => {
        return items.filter(item => !ts.unpinnedMenuItems?.includes(item.id))
    }, [items, ts.unpinnedMenuItems])

    const displayedPinnedItems = React.useMemo(() => {
        return allPinnedItems.filter(item => !autoUnpinnedIds.includes(item.id))
    }, [allPinnedItems, autoUnpinnedIds])

    const displayedPluginItems = React.useMemo(() => {
        return pluginWebviewItems.filter((item: any) => !autoUnpinnedIds.includes(item.id))
    }, [pluginWebviewItems, autoUnpinnedIds])

    const checkOverflow = React.useCallback(() => {
        if (!containerRef.current) return

        const { scrollHeight, clientHeight } = containerRef.current
        if (scrollHeight > clientHeight + 2) {
            if (displayedPluginItems.length > 0) {
                const lastPlugin = displayedPluginItems[displayedPluginItems.length - 1] as any
                if (lastPlugin?.id) {
                    setAutoUnpinnedIds(prev => {
                        if (prev.includes(lastPlugin.id)) return prev
                        return [...prev, lastPlugin.id]
                    })
                    return
                }
            }

            if (displayedPinnedItems.length > 1) {
                const lastItem = displayedPinnedItems[displayedPinnedItems.length - 1]
                setAutoUnpinnedIds(prev => {
                    if (prev.includes(lastItem.id)) return prev
                    return [...prev, lastItem.id]
                })
            }
        }
    }, [displayedPinnedItems, displayedPluginItems])

    React.useEffect(() => {
        if (!containerRef.current) return

        const observer = new ResizeObserver(() => {
            if (overflowCheckTimeoutRef.current) {
                clearTimeout(overflowCheckTimeoutRef.current)
            }
            overflowCheckTimeoutRef.current = setTimeout(() => {
                checkOverflow()
            }, 16)
        })

        observer.observe(containerRef.current)
        checkOverflow()

        return () => {
            observer.disconnect()
            if (overflowCheckTimeoutRef.current) {
                clearTimeout(overflowCheckTimeoutRef.current)
            }
        }
    }, [checkOverflow])

    const unpinnedMenuItems = React.useMemo(() => {
        const manuallyUnpinned = items.filter(item => ts.unpinnedMenuItems?.includes(item.id))
        const forcedUnpinned = items.filter(item => autoUnpinnedIds.includes(item.id))
        const forcedUnpinnedPlugins = pluginWebviewItems.filter(item => autoUnpinnedIds.includes(item.id))

        const allHidden = [...manuallyUnpinned, ...forcedUnpinnedPlugins, ...forcedUnpinned]

        if (allHidden.length === 0) return []

        return [
            {
                iconType: BiChevronRight,
                name: "More",
                subContent: <VerticalMenu
                    items={allHidden}
                    isSidebar
                />,
            } as VerticalMenuItem,
        ]
    }, [items, ts.unpinnedMenuItems, autoUnpinnedIds, pluginWebviewItems])

    return (
        <div>
            <div
                className={cn(
                    "mb-4 p-4 pb-0 flex justify-center w-full",
                    __isDesktop__ && "mt-2",
                )}
            >
                <img
                    src="/seanime-logo.png"
                    alt="logo"
                    className="w-15 h-10 transition-all duration-300"
                />
            </div>
            <VerticalMenu
                className="px-4"
                collapsed={isCollapsed}
                itemClass="relative"
                itemChevronClass="hidden"
                itemIconClass="transition-transform group-data-[state=open]/verticalMenu_parentItem:rotate-90"
                items={[
                    ...displayedPinnedItems,
                    ...displayedPluginItems,
                    ...unpinnedMenuItems,
                    {
                        iconType: LuRefreshCw,
                        name: "Refresh AniList",
                        onClick: () => {
                            ctx.setOpen(false)
                            if (isRefreshingAC) return
                            refreshAC()
                        },
                    },
                ]}
                subContentClass={cn((ts.hideTopNavbar || __isDesktop__) && "border-transparent !border-b-0")}
                onLinkItemClick={() => ctx.setOpen(false)}
                isSidebar
            />

            <SidebarNavbar
                isCollapsed={isCollapsed}
                handleExpandSidebar={() => { }}
                handleUnexpandedSidebar={() => { }}
            />
            {__isDesktop__ && <div className="w-full flex justify-center px-4">
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
    )
}

function SidebarUpdates({ isCollapsed }: { isCollapsed: boolean }) {
    return (
        !__isDesktop__ ? <UpdateModal collapsed={isCollapsed} /> :
            __isTauriDesktop__ ? <TauriUpdateModal collapsed={isCollapsed} /> :
                __isElectronDesktop__ ? <ElectronUpdateModal collapsed={isCollapsed} /> :
                    null
    )
}

function SidebarFooter({ isCollapsed, onLogout }: { isCollapsed: boolean, onLogout: () => void }) {
    const ctx = useAppSidebarContext()
    const pathname = usePathname()
    const serverStatus = useServerStatus()
    const user = useCurrentUser()

    // Extensions
    const { data: updateData } = useGetExtensionUpdateData()
    const pluginWithIssuesCount = usePluginWithIssuesCount()

    // Sync
    const { syncIsActive } = useSyncIsActive()

    // Nakama
    const [nakamaModalOpen, setNakamaModalOpen] = useAtom(nakamaModalOpenAtom)
    const nakamaStatus = useNakamaStatus()

    // Sign out
    const confirmSignOut = useConfirmationDialog({
        title: "Sign out",
        description: "Are you sure you want to sign out?",
        onConfirm: () => {
            onLogout()
        },
    })
    // Login
    const [loginModal, setLoginModal] = useAtom(isLoginModalOpenAtom)


    return (
        <div>
            <VerticalMenu
                collapsed={isCollapsed}
                itemClass="relative"
                onMouseEnter={() => { }}
                onMouseLeave={() => { }}
                onLinkItemClick={() => ctx.setOpen(false)}
                isSidebar
                items={[
                    // {
                    //     iconType: RiSlashCommands2,
                    //     name: "Command palette",
                    //     onClick: () => {
                    //         setSeaCommandOpen(true)
                    //     }
                    // },
                    ...serverStatus?.settings?.nakama?.enabled ? [{
                        iconType: MdOutlineConnectWithoutContact,
                        iconClass: "size-6",
                        name: "Nakama",
                        isCurrent: nakamaModalOpen,
                        onClick: () => {
                            ctx.setOpen(false)
                            setNakamaModalOpen(true)
                        },
                        addon: <>
                            {nakamaStatus?.isHost && !!nakamaStatus?.connectedPeers?.length && <Badge
                                className="absolute right-0 top-0" size="sm"
                                intent="info"
                            >{nakamaStatus?.connectedPeers?.length}</Badge>}

                            {nakamaStatus?.isConnectedToHost && <div
                                className="absolute right-2 top-2 animate-pulse size-2 bg-green-500 rounded-full"
                            ></div>}
                        </>,
                    }] : [],
                    {
                        iconType: BiExtension,
                        name: "Extensions",
                        href: "/extensions",
                        isCurrent: pathname.includes("/extensions"),
                        addon: (!!updateData?.length || !!pluginWithIssuesCount)
                            ? <Badge
                                className="absolute right-0 top-0 bg-red-500 animate-pulse" size="sm"
                                intent="alert-solid"
                            >
                                {updateData?.length || pluginWithIssuesCount || 1}
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
                        iconType: LuSettings,
                        name: "Settings",
                        href: "/settings",
                        isCurrent: pathname === ("/settings"),
                    },
                    ...(ctx.isBelowBreakpoint ? [
                        {
                            iconType: user?.isSimulated ? FiLogIn : BiLogOut,
                            name: user?.isSimulated ? "Sign in" : "Sign out",
                            onClick: user?.isSimulated ? () => setLoginModal(true) : confirmSignOut.open,
                        },
                    ] : []),
                ]}
            />
            <ConfirmationDialog {...confirmSignOut} />
        </div>
    )
}

function SidebarUser({ isCollapsed, expandedSidebar, onLogout }: { isCollapsed: boolean, expandedSidebar: boolean, onLogout: () => void }) {
    const ctx = useAppSidebarContext()
    const user = useCurrentUser()
    const router = useRouter()

    const [dropdownOpen, setDropdownOpen] = React.useState(false)
    const [loginModal, setLoginModal] = useAtom(isLoginModalOpenAtom)
    const [loggingIn, setLoggingIn] = React.useState(false)

    // Sign out
    const confirmSignOut = useConfirmationDialog({
        title: "Sign out",
        description: "Are you sure you want to sign out?",
        onConfirm: () => {
            onLogout()
        },
    })

    return (
        <>
            {!user && (
                <div>
                    <VerticalMenu
                        collapsed={isCollapsed}
                        itemClass="relative"
                        onLinkItemClick={() => ctx.setOpen(false)}
                        isSidebar
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
                            "w-full flex p-2 pt-1 items-center space-x-3",
                            { "hidden": ctx.isBelowBreakpoint },
                        )}
                    >
                        <Avatar size="sm" className="cursor-pointer" src={user?.viewer?.avatar?.medium || undefined} />
                        {expandedSidebar && <p className="truncate text-sm text-[--muted]">{user?.viewer?.name}</p>}
                    </div>}
                    open={dropdownOpen}
                    onOpenChange={setDropdownOpen}
                >
                    {!user.isSimulated ? <DropdownMenuItem onClick={confirmSignOut.open}>
                        <BiLogOut /> Sign out
                    </DropdownMenuItem> : <DropdownMenuItem onClick={() => setLoginModal(true)}>
                        <BiLogIn /> Log in with AniList
                    </DropdownMenuItem>}
                </DropdownMenu>
            </div>}

            <Modal
                title="Log in with AniList"
                description="Using an AniList account is recommended."
                open={loginModal && user?.isSimulated}
                onOpenChange={(v) => setLoginModal(v)}
                overlayClass="bg-opacity-95 bg-gray-950"
                contentClass="border"
            >
                <div className="mt-5 text-center space-y-4">

                    <Link
                        href={ANILIST_PIN_URL}
                        target="_blank"
                    >
                        <Button
                            leftIcon={<svg
                                xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="24" height="24"
                                viewBox="0 0 24 24" role="img"
                            >
                                <path
                                    d="M6.361 2.943 0 21.056h4.942l1.077-3.133H11.4l1.052 3.133H22.9c.71 0 1.1-.392 1.1-1.101V17.53c0-.71-.39-1.101-1.1-1.101h-6.483V4.045c0-.71-.392-1.102-1.101-1.102h-2.422c-.71 0-1.101.392-1.101 1.102v1.064l-.758-2.166zm2.324 5.948 1.688 5.018H7.144z"
                                />
                            </svg>}
                            intent="white"
                            size="md"
                        >Get AniList token</Button>
                    </Link>

                    <Form
                        schema={defineSchema(({ z }) => z.object({
                            token: z.string().min(1, "Token is required"),
                        }))}
                        onSubmit={data => {
                            setLoggingIn(true)
                            router.push("/auth/callback#access_token=" + data.token.trim())
                            setLoginModal(false)
                            setLoggingIn(false)
                        }}
                    >
                        <Field.Textarea
                            name="token"
                            label="Enter the token"
                            fieldClass="px-4"
                        />
                        <Field.Submit showLoadingOverlayOnSuccess loading={loggingIn}>Continue</Field.Submit>
                    </Form>

                </div>
            </Modal>

            <ConfirmationDialog {...confirmSignOut} />
        </>
    )
}
