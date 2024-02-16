"use client"
import { useAutoDownloaderQueueCount } from "@/atoms/auto-downloader-items"
import { useMissingEpisodeCount } from "@/atoms/missing-episodes"
import { serverStatusAtom } from "@/atoms/server-status"
import { useCurrentUser } from "@/atoms/user"
import { __globalSearch_isOpenAtom } from "@/components/application/global-search"
import { UpdateModal } from "@/components/application/update-modal"
import { AppSidebar } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem, DropdownMenuLink } from "@/components/ui/dropdown-menu"
import { Modal } from "@/components/ui/modal"
import { VerticalNav } from "@/components/ui/vertical-nav"
import { useDisclosure } from "@/hooks/use-disclosure"
import { ANILIST_OAUTH_URL } from "@/lib/anilist/config"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { ServerStatus } from "@/lib/server/types"
import { AiOutlineClockCircle } from "@react-icons/all-files/ai/AiOutlineClockCircle"
import { BiCollection } from "@react-icons/all-files/bi/BiCollection"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { FiLogIn } from "@react-icons/all-files/fi/FiLogIn"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { FiSettings } from "@react-icons/all-files/fi/FiSettings"
import { IoLibrary } from "@react-icons/all-files/io5/IoLibrary"
import { useSetAtom } from "jotai"
import { usePathname } from "next/navigation"
import React, { useEffect, useMemo } from "react"
import { BiLogOut } from "react-icons/bi"
import { MdSyncAlt } from "react-icons/md"
import { PiClockCounterClockwiseFill } from "react-icons/pi"
import { SiMyanimelist } from "react-icons/si"
import { TbWorldDownload } from "react-icons/tb"

export function MainSidebar() {

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

    useEffect(() => {
        if (!isPending) {
            setServerStatus(data)
        }
    }, [isPending, data])

    const setGlobalSearchIsOpen = useSetAtom(__globalSearch_isOpenAtom)

    const loginModal = useDisclosure(false)

    const watchListItem = useMemo(() => !!user ? [
        {
            icon: BiCollection,
            name: "My lists",
            href: "/anilist",
            isCurrent: pathname === "/anilist",
        },
        {
            icon: TbWorldDownload,
            name: "Auto downloader",
            href: "/auto-downloader",
            isCurrent: pathname === "/auto-downloader",
            addon: autoDownloaderQueueCount > 0 ? <Badge
                className={"absolute right-0 top-0"} size={"sm"}
                intent={"alert-solid"}
            >{autoDownloaderQueueCount}</Badge> : undefined,
        },
        {
            icon: BiDownload,
            name: "Torrent list",
            href: "/torrent-list",
            isCurrent: pathname === "/torrent-list",
        },
        {
            icon: PiClockCounterClockwiseFill,
            name: "Scan summaries",
            href: "/scan-summaries",
            isCurrent: pathname === "/scan-summaries",
        },
        {
            icon: MdSyncAlt,
            name: "List sync",
            href: "/list-sync",
            isCurrent: pathname === "/list-sync",
        },
        {
            icon: FiSearch,
            name: "Search",
            onClick: () => setGlobalSearchIsOpen(true),
        },
    ] : [], [user, pathname, autoDownloaderQueueCount])

    return (
        <>
            <AppSidebar className={"p-4 h-full flex flex-col justify-between"} sidebarClassName="h-full">
                <div>
                    <div className={"mb-4 flex justify-center w-full"}>
                        <img src="/logo.png" alt="logo" className={"w-15 h-10"} />
                    </div>
                    <VerticalNav
                        itemClassName={"relative"}
                        items={[
                            {
                                icon: IoLibrary,
                                name: "Library",
                                href: "/",
                                isCurrent: pathname === "/",
                            },
                            {
                                icon: AiOutlineClockCircle,
                                name: "Schedule",
                                href: "/schedule",
                                isCurrent: pathname === "/schedule",
                                addon: missingEpisodeCount > 0 ? <Badge
                                    className={"absolute right-0 top-0"} size={"sm"}
                                    intent={"alert-solid"}
                                >{missingEpisodeCount}</Badge> : undefined,
                            },
                            ...watchListItem,
                        ]}
                    />
                </div>
                <div className={"flex w-full gap-2 flex-col"}>
                    <UpdateModal />
                    <div>
                        <VerticalNav
                            items={[
                                {
                                    icon: FiSettings,
                                    name: "Settings",
                                    href: "/settings",
                                    isCurrent: pathname.includes("/settings"),
                                },
                            ]}
                        />
                    </div>
                    {!user && (
                        <div>
                            <VerticalNav
                                items={[
                                    {
                                        icon: FiLogIn,
                                        name: "Login",
                                        onClick: () => window.open(ANILIST_OAUTH_URL, "_self"),
                                    },
                                ]}
                            />
                        </div>
                    )}
                    {!!user && <div className={"flex w-full gap-2 flex-col"}>
                        <DropdownMenu
                            trigger={<div className={"pt-1 w-full flex justify-center"}>
                                <Avatar size={"sm"} className={"cursor-pointer"} src={user?.avatar?.medium || ""} />
                            </div>}
                        >
                            <DropdownMenuLink href="/mal">
                                <SiMyanimelist className="text-lg text-indigo-200" /> MyAnimeList
                            </DropdownMenuLink>
                            <DropdownMenuItem onClick={() => logout()}>
                                <BiLogOut /> Sign out
                            </DropdownMenuItem>
                        </DropdownMenu>
                    </div>}
                </div>
            </AppSidebar>

            <Modal title={"Login"} isOpen={loginModal.isOpen} onClose={loginModal.close} isClosable>
                <div className={"mt-5 text-center space-y-4"}>
                    <Button
                        onClick={() => {
                            window.open(ANILIST_OAUTH_URL)
                        }} intent={"primary-outline"}
                    >Login with AniList</Button>
                </div>
            </Modal>
        </>
    )

}
