"use client"
import { cn } from "@/components/ui/core/styling"
import { usePathname } from "next/navigation"
import React from "react"

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
                <div
                    data-layout-header-background-gradient
                    className="w-full absolute bottom-0 h-[8rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                />
            </>}
        </>
    )
}
