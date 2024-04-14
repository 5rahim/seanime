"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { NavigationMenu, NavigationMenuProps } from "@/components/ui/navigation-menu"
import { useAtomValue } from "jotai/react"
import { usePathname } from "next/navigation"
import React, { useMemo } from "react"

interface OfflineTopNavbarProps {
    children?: React.ReactNode
}

export const OfflineTopNavbar: React.FC<OfflineTopNavbarProps> = (props) => {

    const { children, ...rest } = props

    const serverStatus = useAtomValue(serverStatusAtom)

    const pathname = usePathname()


    const navigationItems = useMemo<NavigationMenuProps["items"]>(() => {

        return [
            {
                href: "/offline",
                // icon: IoLibrary,
                isCurrent: pathname === "/offline",
                name: "My library",
            },
            ...[serverStatus?.settings?.library?.enableManga && {
                href: "/offline#manga",
                icon: null,
                isCurrent: pathname.includes("/offline#manga"),
                name: "Manga",
            }].filter(Boolean) as NavigationMenuProps["items"],
        ].filter(Boolean)
    }, [pathname, serverStatus?.settings?.library?.enableManga])

    return (
        <NavigationMenu
            className="p-0 hidden lg:inline-block"
            itemClass="text-xl"
            items={navigationItems}
        />
    )

}
