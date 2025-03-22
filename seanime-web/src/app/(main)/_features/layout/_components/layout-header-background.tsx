"use client"
import { cn } from "@/components/ui/core/styling"
import { usePathname } from "next/navigation"
import React from "react"
import { SiAnilist } from "react-icons/si"

export function LayoutHeaderBackground() {

    const pathname = usePathname()

    return (
        <>
            {!pathname.startsWith("/entry") && <>
                <div
                    data-layout-header-background
                    className={cn(
                        // "bg-[url(/pattern-3.svg)] bg-[#000] opacity-50 bg-contain bg-center bg-repeat z-[-2] w-full h-[20rem] absolute bottom-0",
                        "bg-[#000] opacity-50 bg-contain bg-center bg-repeat z-[-2] w-full h-[20rem] absolute bottom-0",
                    )}
                >
                </div>
                {pathname.startsWith("/anilist") &&
                    <div
                        data-layout-header-background-anilist-icon-container
                        className="w-full flex items-center justify-center absolute bottom-0 h-[5rem] lg:hidden 2xl:flex"
                    >
                        <SiAnilist className="text-5xl text-white relative z-[2] opacity-40" />
                    </div>}
                <div
                    data-layout-header-background-gradient
                    className="w-full absolute bottom-0 h-[8rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                />
            </>}
        </>
    )
}
