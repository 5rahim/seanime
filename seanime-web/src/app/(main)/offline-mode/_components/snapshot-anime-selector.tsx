import { libraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { cn } from "@/components/ui/core/styling"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React from "react"

type SnapshotAnimeSelectorProps = {
    children?: React.ReactNode
    animeMediaIds: number[]
    setAnimeMediaIds: React.Dispatch<React.SetStateAction<number[]>>
}

export function SnapshotAnimeSelector(props: SnapshotAnimeSelectorProps) {

    const {
        animeMediaIds,
        setAnimeMediaIds,
        children,
        ...rest
    } = props

    const libraryCollection = useAtomValue(libraryCollectionAtom)

    return (
        <>
            <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                {libraryCollection?.lists?.filter(n => n.type === "PLANNING" || n.type === "PAUSED" || n.type === "CURRENT")
                    ?.flatMap(n => n.entries)?.filter(Boolean)
                    ?.map(entry => {
                        return (
                            <div
                                key={entry.mediaId}
                                className={cn(
                                    "col-span-1 aspect-[6/7] rounded-md border overflow-hidden relative bg-[var(--background)] cursor-pointer transition-opacity",
                                    !animeMediaIds.includes(entry.mediaId) && "opacity-80",
                                )}
                                onClick={() => {
                                    setAnimeMediaIds(prev => {
                                        if (prev.includes(entry.mediaId)) {
                                            return prev.filter(n => n !== entry.mediaId)
                                        } else {
                                            return [...prev, entry.mediaId]
                                        }
                                    })
                                }}
                            >
                                <Image
                                    src={entry.media?.coverImage?.large || entry.media?.bannerImage || ""}
                                    placeholder={imageShimmer(700, 475)}
                                    sizes="10rem"
                                    fill
                                    alt=""
                                    className={cn(
                                        "object-center object-cover transition-opacity",
                                        animeMediaIds.includes(entry.mediaId) ? "opacity-100" : "opacity-60",
                                    )}
                                />
                                <p className="line-clamp-2 text-sm absolute m-2 bottom-0 font-semibold z-[10]">
                                    {entry.media?.title?.userPreferred || entry.media?.title?.romaji}
                                </p>
                                <div
                                    className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                                />
                                <div
                                    className={cn(
                                        "z-[5] absolute top-0 w-full h-[80%] bg-gradient-to-b from-[--background] to-transparent transition-opacity",
                                        animeMediaIds.includes(entry.mediaId) ? "opacity-0" : "opacity-100 hover:opacity-80",
                                    )}
                                />

                                {/*<div className="absolute top-0 p-2 z-[6]">*/}
                                {/*    <Checkbox*/}
                                {/*        size="lg"*/}
                                {/*        value={animeMediaIds.includes(entry.mediaId)}*/}
                                {/*        onValueChange={(v) => {*/}
                                {/*            if (v) {*/}
                                {/*                setAnimeMediaIds(prev => [...prev, entry.mediaId])*/}
                                {/*            } else {*/}
                                {/*                setAnimeMediaIds(prev => prev.filter(n => n !== entry.mediaId))*/}
                                {/*            }*/}
                                {/*        }}*/}
                                {/*    />*/}
                                {/*</div>*/}
                            </div>
                        )
                    })}
            </div>
        </>
    )
}
