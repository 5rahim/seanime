"use client"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaEntry } from "@/lib/server/types"
import Image from "next/image"

export function EntryHeaderBackground({ entry }: { entry: MediaEntry }) {
    return (
        <div className="__header h-[30rem] ">
            <div
                className="h-[35rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                {(!!entry.media?.bannerImage || !!entry.media?.coverImage?.extraLarge) && <Image
                    src={entry.media?.bannerImage || entry.media?.coverImage?.extraLarge || ""}
                    alt="banner image"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className="object-cover [object-position:50%_25%] z-[1]"
                />}
                {entry.media?.bannerImage && <Skeleton className="z-0 h-full absolute w-full" />}
                <div
                    className={"w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"}
                />

            </div>
        </div>
    )
}
