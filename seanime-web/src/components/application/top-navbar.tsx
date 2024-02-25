"use client"
import { useMissingEpisodeCount } from "@/atoms/missing-episodes"
import { Badge } from "@/components/ui/badge"
import { NavigationMenu, NavigationMenuProps } from "@/components/ui/navigation-menu"
import { usePathname } from "next/navigation"
import React, { useMemo } from "react"

interface TopNavbarProps {
    children?: React.ReactNode
}

export const TopNavbar: React.FC<TopNavbarProps> = (props) => {

    const { children, ...rest } = props

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
                    className="absolute top-4 right-2 h-2 w-2 p-0" size="sm"
                    intent="alert-solid"
                /> : undefined,
            },
            {
                href: "/anilist",
                icon: null,
                isCurrent: pathname.startsWith("/anilist"),
                name: "My lists",
            },
            {
                href: "/discover",
                icon: null,
                isCurrent: pathname.startsWith("/discover") || pathname.startsWith("/search"),
                name: "Discover",
            },
        ]
    }, [pathname, missingEpisodeCount])

    return (
        <NavigationMenu
            className="p-0 hidden lg:inline-block"
            itemClass="text-xl"
            items={navigationItems}
        />
    )

}
