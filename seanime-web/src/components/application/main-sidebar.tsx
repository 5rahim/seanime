"use client"
import React, { useEffect, useMemo } from "react"
import { VerticalNav } from "@/components/ui/vertical-nav"
import { AppSidebar } from "@/components/ui/app-layout"
import { Avatar } from "@/components/ui/avatar"
import { ANILIST_OAUTH_URL } from "@/lib/anilist/config"
import { FiLogIn } from "@react-icons/all-files/fi/FiLogIn"
import { useDisclosure } from "@/hooks/use-disclosure"
import { Modal } from "@/components/ui/modal"
import { Button } from "@/components/ui/button"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useCurrentUser } from "@/atoms/user"
import { usePathname } from "next/navigation"
import { FiSettings } from "@react-icons/all-files/fi/FiSettings"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import { useSetAtom } from "jotai"
import { __globalSearch_isOpenAtom } from "@/components/application/global-search"
import { IoLibrary } from "@react-icons/all-files/io5/IoLibrary"
import { AiOutlineClockCircle } from "@react-icons/all-files/ai/AiOutlineClockCircle"
import { BiCollection } from "@react-icons/all-files/bi/BiCollection"
import { serverStatusAtom } from "@/atoms/server-status"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { ServerStatus } from "@/lib/server/types"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useMissingEpisodeCount } from "@/atoms/missing-episodes"
import { Badge } from "@/components/ui/badge"

export function MainSidebar() {

    const { user } = useCurrentUser()
    const pathname = usePathname()
    const setServerStatus = useSetAtom(serverStatusAtom)

    const missingEpisodeCount = useMissingEpisodeCount()

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
    ] : [], [user, pathname])

    return (
        <>
            <AppSidebar className={"p-4 h-full flex flex-col justify-between"} sidebarClassName="h-full">
                <div>
                    <div className={"mb-4 flex justify-center w-full"}>
                        <img src="/logo.png" alt="logo" className={"w-15 h-10"}/>
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
                                addon: missingEpisodeCount > 0 ? <Badge className={"absolute right-0 top-0"} size={"sm"}
                                                                        intent={"alert-solid"}>{missingEpisodeCount}</Badge> : undefined,
                            },
                            ...watchListItem,
                            {
                                icon: FiSearch,
                                name: "Search",
                                onClick: () => setGlobalSearchIsOpen(true),
                            },
                            {
                                icon: BiDownload,
                                name: "Torrent list",
                                href: "/torrent-list",
                                isCurrent: pathname === "/torrent-list",
                            },
                        ]}/>
                </div>
                <div className={"flex w-full gap-2 flex-col"}>
                    <div>
                        <VerticalNav items={[
                            {
                                icon: FiSettings,
                                name: "Settings",
                                href: "/settings",
                                isCurrent: pathname.includes("/settings"),
                            },
                        ]}/>
                    </div>
                    {!user && (
                        <div>
                            <VerticalNav items={[
                                {
                                    icon: FiLogIn,
                                    name: "Login",
                                    onClick: () => window.open(ANILIST_OAUTH_URL, "_self"),
                                },
                            ]}/>
                        </div>
                    )}
                    {!!user && <div className={"flex w-full gap-2 flex-col"}>
                        <DropdownMenu trigger={<div className={"pt-1 w-full flex justify-center"}>
                            <Avatar size={"sm"} className={"cursor-pointer"} src={user?.avatar?.medium || ""}/>
                        </div>}>
                            <DropdownMenuItem onClick={() => logout()}>
                                Sign out
                            </DropdownMenuItem>
                        </DropdownMenu>
                    </div>}
                </div>
            </AppSidebar>

            <Modal title={"Login"} isOpen={loginModal.isOpen} onClose={loginModal.close} isClosable>
                <div className={"mt-5 text-center space-y-4"}>
                    <Button onClick={() => {
                        window.open(ANILIST_OAUTH_URL)
                    }} intent={"primary-outline"}>Login with AniList</Button>
                </div>
            </Modal>
        </>
    )

}