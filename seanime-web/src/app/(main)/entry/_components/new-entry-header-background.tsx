"use client"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaEntry } from "@/lib/server/types"
import { motion } from "framer-motion"
import Image from "next/image"
import React from "react"

export function NewEntryHeaderBackground({ entry }: { entry: MediaEntry }) {
    return (
        <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 1, delay: 0.2 }}
            className="__header h-[10rem] md:h-[30rem] "
        >

            <div
                className="h-[35rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] bg-gradient-to-b from-[rgba(0,0,0,0.8)] to-transparent via"
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

                <Image
                    src={"/mask-2.png"}
                    alt="mask"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-left z-[2] transition-opacity duration-1000 opacity-90 hidden lg:block",
                    )}
                />

            </div>
        </motion.div>
    )
}
