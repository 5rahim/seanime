import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { cn } from "@/components/ui/core/styling"
import { MediaEntryEpisode } from "@/lib/server/types"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"

type SliderEpisodeItemProps = {
    episode: MediaEntryEpisode
    onPlay?: ({ path }: { path: string }) => void
} & Omit<React.ComponentPropsWithoutRef<"div">, "onPlay">

export const SliderEpisodeItem = React.forwardRef<HTMLDivElement, SliderEpisodeItemProps>(({ episode, onPlay, ...rest }, ref) => {

    const date = episode.episodeMetadata?.airDate ? new Date(episode.episodeMetadata.airDate) : undefined
    const offset = episode.progressNumber - episode.episodeNumber

    return (
        <div
            ref={ref}
            key={episode.localFile?.path}
            className={cn(
                "rounded-md border overflow-hidden aspect-[4/2] relative flex items-end flex-none group/missed-episode-item cursor-pointer",
                "user-select-none",
                "w-full",
            )}
            onClick={() => onPlay?.({ path: episode.localFile?.path ?? "" })}
            {...rest}
        >
            <div className="absolute w-full h-full overflow-hidden z-[1]">
                {!!episode.episodeMetadata?.image ? <Image
                    src={episode.episodeMetadata?.image}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className="object-cover object-center transition"
                /> : <div
                    className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                <div
                    className="z-[1] absolute bottom-0 w-full h-full md:h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                />
            </div>
            <div
                className={cn(
                    "group-hover/missed-episode-item:opacity-100 text-6xl text-gray-200",
                    "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                    "hidden md:flex",
                )}
            >
                <AiFillPlayCircle className="opacity-50" />
            </div>
            <div className="relative z-[3] w-full p-4 space-y-1">
                <p className="w-[80%] line-clamp-1 text-[--muted] font-semibold">{episode.episodeTitle?.replaceAll("`", "'")}</p>
                <div className="w-full justify-between flex items-center">
                    <p className="text-base md:text-xl lg:text-2xl font-semibold line-clamp-2">
                        <span>{episode.displayTitle} {!!episode.basicMedia?.episodes &&
                            (episode.basicMedia.episodes != 1 &&
                                <span className="opacity-40">/{` `}{episode.basicMedia.episodes - offset}</span>)}
                        </span>
                    </p>
                    <p className="text-[--muted] text-sm md:text-base">{episode.episodeMetadata?.length + "m" || ""}</p>
                </div>
                {episode.isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    )
})

type GenericSliderEpisodeItemProps = {
    title: React.ReactNode
    actionIcon?: React.ReactElement | null
    image?: string | null
    topTitle?: string | null
    meta?: string | null
    larger?: boolean
    isInvalid?: boolean
} & Omit<React.ComponentPropsWithoutRef<"div">, "onPlay">

export const GenericSliderEpisodeItem = React.forwardRef<HTMLDivElement, GenericSliderEpisodeItemProps>((props, ref) => {

    const {
        children,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-50" /> : undefined,
        image,
        topTitle,
        meta,
        title,
        larger = false,
        isInvalid,
        ...rest
    } = props

    return (
        <div
            ref={ref}
            className={cn(
                "rounded-md border overflow-hidden aspect-[4/2] relative flex items-end flex-none group/missed-episode-item cursor-pointer",
                "user-select-none",
                "w-full",
            )}
            {...rest}
        >
            <div className="absolute w-full h-full overflow-hidden z-[1]">
                {!!image ? <Image
                    src={image}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className="object-cover object-center transition"
                /> : <div
                    className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                <div
                    className="z-[1] absolute bottom-0 w-full h-full md:h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                />
            </div>
            <div
                className={cn(
                    "group-hover/missed-episode-item:opacity-100 text-6xl text-gray-200",
                    "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                    "hidden md:flex",
                )}
            >
                {actionIcon && actionIcon}
            </div>
            <div className="relative z-[3] w-full p-4 space-y-1">
                <p className="w-[80%] line-clamp-1 text-[--muted] font-semibold">{topTitle}</p>
                <div className="w-full justify-between flex items-center">
                    <p className="text-base md:text-xl lg:text-2xl font-semibold line-clamp-2">
                        {title}
                    </p>
                    {(meta) && <p className="text-[--muted] text-sm md:text-base">{meta}</p>}
                </div>
                {isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    )
})
