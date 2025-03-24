"use client"
import { useMissingEpisodeCount } from "@/app/(main)/_hooks/missing-episodes-loader"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Badge } from "@/components/ui/badge"
import { NavigationMenu, NavigationMenuProps } from "@/components/ui/navigation-menu"
import { usePathname } from "next/navigation"
import React, { useMemo } from "react"

interface TopMenuProps {
    children?: React.ReactNode
}

export const TopMenu: React.FC<TopMenuProps> = (props) => {

    const { children, ...rest } = props

    const serverStatus = useServerStatus()

    const pathname = usePathname()

    const missingEpisodeCount = useMissingEpisodeCount()

    const navigationItems = useMemo<NavigationMenuProps["items"]>(() => {

        return [
            {
                href: "/",
                // icon: IoLibrary,
                isCurrent: pathname === "/",
                name: "My library",
            },
            {
                href: "/schedule",
                icon: null,
                isCurrent: pathname.startsWith("/schedule"),
                name: "Schedule",
                addon: missingEpisodeCount > 0 ? <Badge
                    className="absolute top-1 right-2 h-2 w-2 p-0" size="sm"
                    intent="alert-solid"
                /> : undefined,
            },
            ...[serverStatus?.settings?.library?.enableManga && {
                href: "/manga",
                icon: null,
                isCurrent: pathname.startsWith("/manga"),
                name: "Manga",
            }].filter(Boolean) as NavigationMenuProps["items"],
            {
                href: "/discover",
                icon: null,
                isCurrent: pathname.startsWith("/discover") || pathname.startsWith("/search"),
                name: "Discover",
            },
            {
                href: "/anilist",
                icon: null,
                isCurrent: pathname.startsWith("/anilist"),
                name: "AniList",
            },
        ].filter(Boolean)
    }, [pathname, missingEpisodeCount, serverStatus?.settings?.library?.enableManga])

    return (
        <NavigationMenu
            className="p-0 hidden lg:inline-block"
            itemClass="text-xl"
            items={navigationItems}
            data-top-menu
        />
    )

}
