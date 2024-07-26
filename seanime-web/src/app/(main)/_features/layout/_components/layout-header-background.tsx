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
                {/*{!pathname.startsWith("/anilist") && <div*/}
                {/*    className={cn(*/}
                {/*        "bg-[url(/pattern-2.svg)] bg-[--background] opacity-60 bg-cover bg-center bg-repeat z-[-2] w-full h-[15rem] absolute bottom-0",*/}
                {/*    )}*/}
                {/*/>}*/}
                <div
                    className={cn(
                        // "bg-[url(/pattern-3.svg)] bg-[#000] opacity-50 bg-contain bg-center bg-repeat z-[-2] w-full h-[20rem] absolute bottom-0",
                        "bg-[#000] opacity-50 bg-contain bg-center bg-repeat z-[-2] w-full h-[20rem] absolute bottom-0",
                    )}
                >
                    {pathname.startsWith("/anilist") && <div className="w-full flex items-center justify-center absolute bottom-0 h-[5rem]">
                        <SiAnilist className="text-5xl text-white relative z-[1]" />
                    </div>}
                </div>
                <div
                    className="w-full absolute bottom-0 h-[8rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                />
            </>}
        </>
    )
}
