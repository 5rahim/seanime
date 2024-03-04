"use client"
import { __libraryHeaderEpisodeAtom } from "@/app/(main)/(library)/_containers/continue-watching"
import { cn } from "@/components/ui/core/styling"
import { MediaEntryEpisode } from "@/lib/server/types"
import { useThemeSettings } from "@/lib/theme/hooks"
import { Transition } from "@headlessui/react"
import { motion } from "framer-motion"
import { atom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import Image from "next/image"
import React, { useEffect, useState } from "react"
import { useWindowScroll } from "react-use"

export const __libraryHeaderImageAtom = atom<string | null>(null)

export function LibraryHeader({ list }: { list: MediaEntryEpisode[] }) {

    const ts = useThemeSettings()

    const image = useAtomValue(__libraryHeaderImageAtom)
    const [actualImage, setActualImage] = useState<string | null>(null)
    const [prevImage, setPrevImage] = useState<string | null>(null)
    const [dimmed, setDimmed] = useState(false)

    const setHeaderEpisode = useSetAtom(__libraryHeaderEpisodeAtom)

    useEffect(() => {
        if (image != actualImage) {
            if (actualImage === null) {
                setActualImage(image)
            } else {
                setActualImage(null)
            }
        }
    }, [image])

    React.useLayoutEffect(() => {
        const t = setTimeout(() => {
            if (image != actualImage) {
                setActualImage(image)
                setHeaderEpisode(list.find(ep => ep.basicMedia?.bannerImage === image || ep.episodeMetadata?.image === image) || null)
            }
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
        <>
            {!!ts.libraryScreenCustomBackgroundImage && (
                <div
                    className="LIB_HEADER_FADE_BG w-full absolute z-[1] top-0 h-[40rem] opacity-100 bg-gradient-to-b from-[#0c0c0c] via-[#0c0c0c] to-transparent via"
                />
            )}
            <div
                className={cn(
                    "LIB_HEADER_CONTAINER __header h-[20rem] z-[1] top-0 w-full absolute group/library-header",
                    // Make it not fixed when the user scrolls down if a background image is set
                    !ts.libraryScreenCustomBackgroundImage && "fixed",
                )}
            >
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.2 }}
                    className="LIB_HEADER_INNER_CONTAINER h-full z-[0] w-full flex-none object-cover object-center absolute top-0 overflow-hidden"
                >

                    <div
                        className="LIB_HEADER_TOP_FADE w-full absolute z-[2] top-0 h-[10rem] opacity-20 bg-gradient-to-b from-[#0c0c0c] to-transparent via"
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
                                "object-cover object-center z-[1] opacity-80 transition-all duration-700",
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
                        className="LIB_HEADER_IMG_BOTTOM_FADE w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-opacity-50 via-10% to-transparent"
                    />
                    <div className="h-full absolute w-full xl-right-48">
                        <Image
                            src={"/mask-2.png"}
                            alt="mask"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            className={cn(
                                "object-cover object-left z-[2] transition-opacity duration-1000 opacity-5",
                            )}
                        />
                    </div>
                    <div className="h-full absolute w-full xl:-right-48">
                        <Image
                            src={"/mask.png"}
                            alt="mask"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            className={cn(
                                "object-cover object-right z-[2] transition-opacity duration-1000 opacity-5",
                            )}
                        />
                    </div>
                </motion.div>
            </div>
        </>
    )

}
