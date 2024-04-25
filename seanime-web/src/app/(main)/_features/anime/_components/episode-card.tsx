import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/image-helpers"
import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"

type EpisodeCardProps = {
    title: React.ReactNode
    actionIcon?: React.ReactElement | null
    image?: string
    onClick?: () => void
    topTitle?: string
    meta?: string
    type?: "carousel" | "grid"
    isInvalid?: boolean
    containerClass?: string
    episodeNumber?: number
    progressNumber?: number
    progressTotal?: number
    mRef?: React.RefObject<HTMLDivElement>
} & Omit<React.ComponentPropsWithoutRef<"div">, "title">

export function EpisodeCard(props: EpisodeCardProps) {

    const {
        children,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-50" /> : undefined,
        image,
        onClick,
        topTitle,
        meta,
        title,
        type = "carousel",
        isInvalid,
        className,
        containerClass,
        mRef,
        episodeNumber,
        progressTotal,
        progressNumber,
        ...rest
    } = props

    const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
    const offset = React.useMemo(() => progressNumber && episodeNumber ? progressNumber - episodeNumber : 0, [progressNumber, episodeNumber])

    return (
        <div
            ref={mRef}
            className={cn(
                "rounded-md border overflow-hidden aspect-[4/2] relative flex items-end flex-none group/episode-card cursor-pointer",
                "user-select-none",
                type === "carousel" && "w-full",
                type === "grid" && "w-72 lg:w-[26rem]",
                className,
                containerClass,
            )}
            onClick={onClick}
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
                {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                <EpisodeItemBottomGradient />
            </div>
            <div
                className={cn(
                    "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
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
                        <span>{title}{showTotalEpisodes ?
                            <span className="opacity-40">{` / `}{progressTotal! - offset}</span>
                            : ``}</span>
                    </p>
                    {(meta) && <p className="text-[--muted] text-sm md:text-base">{meta}</p>}
                </div>
                {isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    )

}
