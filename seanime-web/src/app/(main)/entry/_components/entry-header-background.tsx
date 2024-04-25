"use client"
import { Anime_MediaEntry } from "@/api/generated/types"
import { Skeleton } from "@/components/ui/skeleton"
import { motion } from "framer-motion"
import Image from "next/image"

export function EntryHeaderBackground({ entry }: { entry: Anime_MediaEntry }) {
    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="__header h-[30rem] "
        >
            <div
                className="h-[35rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-50 bg-gradient-to-b from-[#0c0c0c] to-transparent via"
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
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-[--background] via-opacity-50 via-10% to-transparent"
                />

            </div>
        </motion.div>
    )
}
