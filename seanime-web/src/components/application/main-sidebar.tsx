"use client"
import { useAutoDownloaderQueueCount } from "@/app/(main)/auto-downloader/_lib/auto-downloader-items"
import { useMissingEpisodeCount } from "@/atoms/missing-episodes"
import { serverStatusAtom } from "@/atoms/server-status"
import { useCurrentUser } from "@/atoms/user"
import { __globalSearch_isOpenAtom } from "@/components/application/global-search"
import { UpdateModal } from "@/components/application/update-modal"
import { AppSidebar, useAppSidebarContext } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Modal } from "@/components/ui/modal"
import { VerticalMenu } from "@/components/ui/vertical-menu"
import { useDisclosure } from "@/hooks/use-disclosure"
import { ANILIST_OAUTH_URL } from "@/lib/anilist/config"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { ServerStatus } from "@/lib/server/types"
import { useSetAtom } from "jotai"
import Link from "next/link"
import { usePathname } from "next/navigation"
import React from "react"
import { AiOutlineClockCircle } from "react-icons/ai"
import { BiCollection, BiDownload, BiLogOut } from "react-icons/bi"
import { FiLogIn, FiSearch, FiSettings } from "react-icons/fi"
import { IoLibrary } from "react-icons/io5"
import { MdSyncAlt } from "react-icons/md"
import { PiClockCounterClockwiseFill } from "react-icons/pi"
import { SiMyanimelist } from "react-icons/si"
import { TbWorldDownload } from "react-icons/tb"

export function MainSidebar() {

    const ctx = useAppSidebarContext()
    const isCollapsed = ctx.size === "slim" && !ctx.isBelowBreakpoint

    const { user } = useCurrentUser()
    const pathname = usePathname()
    const setServerStatus = useSetAtom(serverStatusAtom)

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

    return (
        <>
            <AppSidebar
                className="p-4 h-full flex flex-col justify-between"
                // sidebarClass="h-full"
            >
                <div>
                    <div className="mb-4 flex justify-center w-full">
                        <img src="/logo.png" alt="logo" className="w-15 h-10" />
                    </div>
                    <VerticalMenu
                        collapsed={isCollapsed}
                        itemClass="relative"
                        items={[
                            {
                                iconType: IoLibrary,
                                name: "Library",
                                href: "/",
                                isCurrent: pathname === "/",
                            },
                            {
                                iconType: AiOutlineClockCircle,
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
                            {
                                iconType: TbWorldDownload,
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
                        ]}
                        onLinkItemClick={() => ctx.setOpen(false)}
                    />
                </div>
                <div className="flex w-full gap-2 flex-col">
                    <UpdateModal />
                    <div>
                        <VerticalMenu
                            collapsed={isCollapsed}
                            itemClass="relative"
                            items={[
                                {
                                    iconType: FiSettings,
                                    name: "Settings",
                                    href: "/settings",
                                    isCurrent: pathname.includes("/settings"),
                                },
                            ]}
                        />
                    </div>
                    {!user && (
                        <div>
                            <VerticalMenu
                                collapsed={isCollapsed}
                                itemClass="relative"
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
                            trigger={<div className="pt-1 w-full flex justify-center">
                                <Avatar size="sm" className="cursor-pointer" src={user?.avatar?.medium || ""} />
                            </div>}
                        >
                            <Link href="/mal">
                                <DropdownMenuItem>
                                    <SiMyanimelist className="text-lg text-indigo-200" /> MyAnimeList
                                </DropdownMenuItem>
                            </Link>
                            <DropdownMenuItem onClick={() => logout()}>
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
        </>
    )

}
