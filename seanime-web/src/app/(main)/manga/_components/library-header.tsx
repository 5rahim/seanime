"use client"
import { AL_BaseManga } from "@/api/generated/types"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { cn } from "@/components/ui/core/styling"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { atom, useAtomValue } from "jotai"
import { useSetAtom } from "jotai/react"
import { AnimatePresence, motion } from "motion/react"
import Image from "next/image"
import React, { useEffect, useState } from "react"
import { useWindowScroll } from "react-use"

export const __mangaLibraryHeaderImageAtom = atom<string | null>(null)
export const __mangaLibraryHeaderMangaAtom = atom<AL_BaseManga | null>(null)

const MotionImage = motion(Image)

export function LibraryHeader({ manga }: { manga: AL_BaseManga[] }) {

    const ts = useThemeSettings()

    const image = useAtomValue(__mangaLibraryHeaderImageAtom)
    const [actualImage, setActualImage] = useState<string | null>(null)
    const [prevImage, setPrevImage] = useState<string | null>(null)
    const [dimmed, setDimmed] = useState(false)

    const setHeaderManga = useSetAtom(__mangaLibraryHeaderMangaAtom)

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
                setHeaderManga(manga.find(ep => ep?.bannerImage === image) || null)
            }
        }, 600)

        return () => {
            clearTimeout(t)
        }
    }, [image])

    useEffect(() => {
        if (actualImage) {
            setPrevImage(actualImage)
            setHeaderManga(manga.find(ep => ep?.bannerImage === actualImage) || null)
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
            <div
                data-library-header-container
                className={cn(
                    "LIB_HEADER_CONTAINER __header h-[25rem] z-[1] top-0 w-full absolute group/library-header",
                    // Make it not fixed when the user scrolls down if a background image is set
                    !ts.libraryScreenCustomBackgroundImage && "fixed",
                )}
            >
                <div
                    data-library-header-banner-top-gradient
                    className={cn(
                        "w-full z-[3] absolute bottom-[-10rem] h-[10rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                        !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    )}
                />

                <motion.div
                    data-library-header-inner-container
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.2 }}
                    className={cn(
                        "LIB_HEADER_INNER_CONTAINER h-full z-[0] w-full flex-none object-cover object-center absolute top-0 overflow-hidden bg-[--background]",
                        !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    )}
                >

                    {!ts.disableSidebarTransparency && <div
                        data-library-header-banner-top-gradient
                        className="hidden lg:block h-full absolute z-[2] w-[20%] opacity-70 left-0 top-0 bg-gradient bg-gradient-to-r from-[var(--background)] to-transparent"
                    />}

                    <div
                        data-library-header-banner-bottom-gradient
                        className="w-full z-[3] opacity-70 lg:opacity-50 absolute top-0 h-[5rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent"
                    />

                    {/*<div*/}
                    {/*    className="LIB_HEADER_TOP_FADE w-full absolute z-[2] top-0 h-[10rem] opacity-20 bg-gradient-to-b from-[var(--background)] to-transparent via"*/}
                    {/*/>*/}
                    <AnimatePresence>
                        {!!actualImage && (
                            <motion.div
                                key="manga-library-header-banner-image-container"
                                data-library-header-image-container
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                exit={{ opacity: 0.4 }}
                                transition={{ duration: 0.5 }}
                            >
                                <MotionImage
                                    data-library-header-banner-image
                                    src={getImageUrl(actualImage || prevImage!)}
                                    alt="banner image"
                                    fill
                                    quality={100}
                                    priority
                                    sizes="100vw"
                                    className={cn(
                                        "object-cover object-center z-[1] opacity-100 transition-opacity duration-700 scroll-locked-offset",
                                        { "opacity-5": dimmed },
                                    )}
                                    initial={{ scale: 1.01, y: 0 }}
                                    animate={{
                                        scale: Math.min(1 + y * 0.0002, 1.03),
                                        // y: Math.max(y * -0.9, -10)
                                    }}
                                    exit={{ scale: 1.01, y: 0 }}
                                    transition={{ duration: 0.6, ease: "easeOut" }}
                                />
                            </motion.div>
                        )}
                    </AnimatePresence>
                    {/* {prevImage && <MotionImage
                     data-library-header-banner-previous-image
                     src={getImageUrl(prevImage)}
                     alt="banner image"
                     fill
                     quality={100}
                     priority
                     sizes="100vw"
                     className={cn(
                     "object-cover object-center z-[1] opacity-50 transition-opacity scroll-locked-offset",
                     { "opacity-5": dimmed },
                     )}
                     initial={{ scale: 1, y: 0 }}
                     animate={{ scale: Math.min(1 + y * 0.0002, 1.03), y: Math.max(y * -0.9, -10) }}
                     exit={{ scale: 1, y: 0 }}
                     transition={{ duration: 0.6, ease: "easeOut" }}
                     />} */}
                    <div
                        data-library-header-banner-bottom-fade
                        className="LIB_HEADER_IMG_BOTTOM_FADE w-full z-[2] absolute bottom-0 h-[20rem] lg:h-[15rem] bg-gradient-to-t from-[--background] lg:via-opacity-50 lg:via-10% to-transparent"
                    />
                </motion.div>
            </div>
        </>
    )

}
