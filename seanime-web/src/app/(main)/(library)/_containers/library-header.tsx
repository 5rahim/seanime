"use client"
import { cn } from "@/components/ui/core/styling"
import { Transition } from "@headlessui/react"
import { atom, useAtomValue } from "jotai"
import Image from "next/image"
import React, { useEffect, useState } from "react"
import { useWindowScroll } from "react-use"

export const __libraryHeaderImageAtom = atom<string | null>(null)

// ugly but works
export function LibraryHeader() {

    const image = useAtomValue(__libraryHeaderImageAtom)
    const [actualImage, setActualImage] = useState<string | null>(null)
    const [prevImage, setPrevImage] = useState<string | null>(null)
    const [dimmed, setDimmed] = useState(false)

    useEffect(() => {
        if (actualImage === null) {
            setActualImage(image)
        } else {
            setActualImage(null)
        }
        const t = setTimeout(() => {
            setActualImage(image)
        }, 500)

        return () => {
            clearTimeout(t)
        }
    }, [image])

    useEffect(() => {
        if (actualImage)
            setPrevImage(actualImage)
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
        <div className="__header h-[18rem] z-[-1] top-0 w-full lg:w-[calc(100%-5rem)] fixed group/library-header hidden md:block">
            <div
                className="h-[25rem] z-[0] w-full flex-none object-cover object-center absolute top-0 overflow-hidden">
                <div
                    className="w-full absolute z-[2] top-0 h-[10rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
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
                            { "opacity-20": dimmed },
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
                    className="w-full z-[2] absolute bottom-0 h-[40rem] bg-gradient-to-t from-[--background] via-opacity-50 via-10% to-transparent"
                />
                <div
                    className="w-[4rem] z-[2] absolute top-0 right-0 h-[40rem] bg-gradient-to-l from-[--background] via-opacity-50 via-10% to-transparent"
                />
            </div>
        </div>
    )

}
