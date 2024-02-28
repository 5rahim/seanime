"use client"
import { __libraryHeaderEpisodeAtom } from "@/app/(main)/(library)/_containers/continue-watching"
import { cn } from "@/components/ui/core/styling"
import { MediaEntryEpisode } from "@/lib/server/types"
import { Transition } from "@headlessui/react"
import { motion } from "framer-motion"
import { atom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import Image from "next/image"
import React, { useEffect, useState } from "react"
import { useWindowScroll } from "react-use"

export const __libraryHeaderImageAtom = atom<string | null>(null)

// ugly but works
export function LibraryHeader({ list }: { list: MediaEntryEpisode[] }) {

    const image = useAtomValue(__libraryHeaderImageAtom)
    const [actualImage, setActualImage] = useState<string | null>(null)
    const [prevImage, setPrevImage] = useState<string | null>(null)
    const [dimmed, setDimmed] = useState(false)

    const setHeaderEpisode = useSetAtom(__libraryHeaderEpisodeAtom)

    useEffect(() => {
        if (actualImage === null) {
            setActualImage(image)
        } else {
            setActualImage(null)
        }
        const t = setTimeout(() => {
            setActualImage(image)
            setHeaderEpisode(list.find(ep => ep.basicMedia?.bannerImage === image || ep.episodeMetadata?.image === image) || null)
        }, 600)

        return () => {
            clearTimeout(t)
        }
    }, [image])

    useEffect(() => {
        if (actualImage) {
            setPrevImage(actualImage)
            setHeaderEpisode(list.find(ep => ep.basicMedia?.bannerImage === actualImage || ep.episodeMetadata?.image === actualImage) || null)
        }
    }, [actualImage])

    const { y } = useWindowScroll()

    useEffect(() => {
        if (y > 100)
            setDimmed(true)
        else
            setDimmed(false)
    }, [(y > 100)])

    if (!image) return null

    return (
        <div className="__header h-[20rem] z-[-1] top-0 w-full fixed group/library-header">
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                transition={{ duration: 1, delay: 0.2 }}
                className="h-[30rem] z-[0] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[10rem] opacity-20 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <Transition
                    show={!!actualImage}
                    enter="transition-opacity duration-500"
                    enterFrom="opacity-0"
                    enterTo="opacity-100"
                    leave="transition-opacity duration-500"
                    leaveFrom="opacity-100"
                    leaveTo="opacity-0"
                >
                    {(actualImage || prevImage) && <Image
                        src={actualImage || prevImage!}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className={cn(
                            "object-cover object-center z-[1] opacity-100 transition-all duration-700",
                            // "group-hover/library-header:opacity-100",
                            { "opacity-10": dimmed },
                        )}
                    />}
                </Transition>
                {prevImage && <Image
                    src={prevImage}
                    alt="banner image"
                    fill
                    quality={100}
                    priority
                    sizes="100vw"
                    className={cn(
                        "object-cover object-center z-[1] opacity-50 transition-all",
                        { "opacity-10": dimmed },
                    )}
                />}
                <div
                    className="w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-opacity-50 via-10% to-transparent"
                />
                <div className="h-full absolute w-full xl:-left-28">
                    <Image
                        src={"/mask-2.png"}
                        alt="mask"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className={cn(
                            "object-cover object-left z-[2] transition-opacity duration-1000 opacity-70",
                        )}
                    />
                </div>
            </motion.div>
        </div>
    )

}
