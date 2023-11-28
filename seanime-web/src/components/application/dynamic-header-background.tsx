"use client"
import Image from "next/image"
import React from "react"
import { usePathname } from "next/navigation"

export function DynamicHeaderBackground() {

    const pathname = usePathname()

    return (
        <>
            {!pathname.startsWith("/entry") && <>
                {!pathname.startsWith("/anilist") && <Image
                    src={"/landscape-beach.jpg"}
                    alt={"tenki no ko"}
                    fill
                    priority
                    className={"object-cover object-center z-[-2]"}
                />}
                {pathname.startsWith("/anilist") && <Image
                    src={"/landscape-tenki-no-ko.jpg"}
                    alt={"tenki no ko"}
                    fill
                    priority
                    className={"object-cover z-[-2]"}
                />}
                <div
                    className={"w-full absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background-color] to-transparent z-[-2]"}
                />
            </>}
        </>
    )
}