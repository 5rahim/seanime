import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { cn } from "@/components/ui/core/styling"
import { ProgressBar } from "@/components/ui/progress-bar"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
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
    hasDiscrepancy?: boolean
    length?: string | number | null
    imageClass?: string
    badge?: React.ReactNode
    percentageComplete?: number
    minutesRemaining?: number
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
        hasDiscrepancy,
        length,
        imageClass,
        badge,
        percentageComplete,
        minutesRemaining,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const ts = useThemeSettings()

    const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
    const offset = React.useMemo(() => hasDiscrepancy ? 1 : 0, [hasDiscrepancy])

    if (ts.useLegacyEpisodeCard) {
        return (
            <div
                ref={mRef}
                className={cn(
                    "rounded-lg overflow-hidden aspect-[4/2] relative flex items-end flex-none group/episode-card cursor-pointer",
                    "select-none",
                    type === "carousel" && "w-full",
                    type === "grid" && "w-72 lg:w-[26rem]",
                    className,
                    containerClass,
                )}
                onClick={onClick}
                {...rest}
            >
                <div className="absolute w-full h-full rounded-lg overflow-hidden z-[1]">
                    {!!image ? <Image
                        src={getImageUrl(image)}
                        alt={""}
                        fill
                        quality={100}
                        placeholder={imageShimmer(700, 475)}
                        sizes="20rem"
                        className={cn(
                            "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                            imageClass,
                        )}
                    /> : <div
                        className="h-full block rounded-lg absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                    ></div>}
                    {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                    <EpisodeItemBottomGradient />

                    {(serverStatus?.settings?.library?.enableWatchContinuity && !!percentageComplete) &&
                        <div className="absolute bottom-0 left-0 w-full z-[3]">
                            <ProgressBar value={percentageComplete} size="sm" />
                        </div>}
                </div>
                <div
                    className={cn(
                        "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
                        "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                        "hidden md:flex")}
                > {actionIcon && actionIcon} </div>
                <div className="relative z-[3] w-full p-4 space-y-0"><p
                    className="w-[80%] line-clamp-1 text-md md:text-lg transition-colors duration-200 text-[--foreground] font-semibold"
                >{topTitle?.replaceAll("`", "'")}</p>
                    <div className="w-full justify-between flex flex-none items-center">
                        <p
                            className="text-base md:text-lg font-medium line-clamp-1"
                        >
                            <span>{title}{showTotalEpisodes ? <span className="opacity-40">{` / `}{progressTotal! - offset}</span> :
                                ``}</span>
                        </p> {(!!meta || !!length) &&
                        <p className="text-[--muted] flex-none ml-2 text-sm md:text-base line-clamp-2 text-right"> {meta}{!!meta && !!length && `  • `}{length
                            ? `${length}m`
                            : ""} </p>}
                    </div>
                    {isInvalid && <p className="text-red-300">No metadata found</p>} </div>
            </div>
        )
    }

    return (
        <div
            ref={mRef}
            className={cn(
                "rounded-lg overflow-hidden space-y-2 flex-none group/episode-card cursor-pointer",
                "select-none",
                type === "carousel" && "w-full",
                type === "grid" && "aspect-[4/2] w-72 lg:w-[26rem]",
                className,
                containerClass,
            )}
            onClick={onClick}
            {...rest}
        >
            <div className="w-full h-full rounded-lg overflow-hidden z-[1] aspect-[4/2] relative">
                {!!image ? <Image
                    src={getImageUrl(image)}
                    alt={""}
                    fill
                    quality={100}
                    placeholder={imageShimmer(700, 475)}
                    sizes="20rem"
                    className={cn(
                        "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                        imageClass,
                    )}
                /> : <div
                    className="h-full block rounded-lg absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                ></div>}
                {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                <EpisodeItemBottomGradient />

                {(serverStatus?.settings?.library?.enableWatchContinuity && !!percentageComplete) &&
                    <div className="absolute bottom-0 left-0 w-full z-[3]">
                        <ProgressBar value={percentageComplete} size="xs" />
                        {minutesRemaining && <div className="absolute bottom-2 right-2">
                            <p className="text-[--muted] text-sm">{minutesRemaining}m left</p>
                        </div>}
                    </div>}

                <div
                    className={cn(
                        "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
                        "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                        "hidden md:flex",
                    )}
                >
                    {actionIcon && actionIcon}
                </div>

                {isInvalid && <p className="text-red-300 opacity-50 absolute left-2 bottom-2 z-[2]">No metadata found</p>}
            </div>
            <div className="relative z-[3] w-full space-y-0">
                <p className="w-[80%] line-clamp-1 text-md md:text-lg transition-colors duration-200 text-[--foreground] font-semibold">{topTitle?.replaceAll(
                    "`",
                    "'")}</p>
                <div className="w-full justify-between flex flex-none items-center">
                    <p className="text-base md:text-xl font-medium line-clamp-1">
                        <span>{title}{showTotalEpisodes ?
                            <span className="opacity-40">{` / `}{progressTotal! - offset}</span>
                            : ``}</span>
                    </p>
                    {(!!meta || !!length) && <p className="text-[--muted] flex-none ml-2 text-sm md:text-base line-clamp-2 text-right">
                        {meta}{!!meta && !!length && `  • `}{length ? `${length}m` : ""}
                    </p>}
                </div>
            </div>
        </div>
    )

}
