import { Anime_Episode } from "@/api/generated/types"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { cn } from "@/components/ui/core/styling"
import { ProgressBar } from "@/components/ui/progress-bar"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import Image from "next/image"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { PluginEpisodeCardContextMenuItems } from "../../plugin/actions/plugin-actions"

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
    anime?: {
        id?: number
        image?: string
        title?: string
    }
    episode?: Anime_Episode // Optional, used for plugin actions
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
        anime,
        ...rest
    } = props

    const router = useRouter()
    const pathname = usePathname()
    const serverStatus = useServerStatus()
    const ts = useThemeSettings()
    const { setPreviewModalMediaId } = useMediaPreviewModal()

    const showAnimeInfo = ts.showEpisodeCardAnimeInfo && !!anime
    const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
    const offset = React.useMemo(() => hasDiscrepancy ? 1 : 0, [hasDiscrepancy])

    const Meta = () => (
        <div data-episode-card-info-container className="relative z-[3] w-full space-y-0">
            <p
                data-episode-card-title
                className="w-[80%] line-clamp-1 text-md md:text-lg transition-colors duration-200 text-[--foreground] font-semibold"
            >
                {topTitle?.replaceAll("`", "'")}
            </p>
            <div data-episode-card-info-content className="w-full justify-between flex flex-none items-center">
                <p data-episode-card-subtitle className="line-clamp-1 flex items-center">
                    <span className="flex-none text-base md:text-xl font-medium">{title}{showTotalEpisodes ?
                        <span className="opacity-40">{` / `}{progressTotal! - offset}</span>
                        : ``}</span>
                    <span className="text-[--muted] text-base md:text-xl ml-2 font-normal line-clamp-1">{showAnimeInfo
                        ? "- " + anime.title
                        : ""}</span>
                </p>
                {(!!meta || !!length) && (
                    <p data-episode-card-meta-text className="text-[--muted] flex-none ml-2 text-sm md:text-base line-clamp-2 text-right">
                        {meta}{!!meta && !!length && `  • `}{length ? `${length}m` : ""}
                    </p>)}
            </div>
        </div>
    )

    return (
        <SeaContextMenu
            hideMenuIf={!anime?.id}
            content={
                <ContextMenuGroup>
                    <ContextMenuLabel className="text-[--muted] line-clamp-1 py-0 my-2">
                        {anime?.title}
                    </ContextMenuLabel>

                    {pathname !== "/entry" && <>
                        <ContextMenuItem
                            onClick={() => {
                                router.push(`/entry?id=${anime?.id}`)
                            }}
                        >
                            Open page
                        </ContextMenuItem>
                        {!serverStatus?.isOffline && <ContextMenuItem
                            onClick={() => {
                                setPreviewModalMediaId(anime?.id || 0, "anime")
                            }}
                        >
                            Preview
                        </ContextMenuItem>}

                    </>}

                    <PluginEpisodeCardContextMenuItems episode={props.episode} />

                </ContextMenuGroup>
            }
        >
            <ContextMenuTrigger>
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
                    data-episode-card
                    data-episode-number={episodeNumber}
                    data-media-id={anime?.id}
                    data-progress-total={progressTotal}
                    data-progress-number={progressNumber}
                    {...rest}
                >
                    <div data-episode-card-image-container className="w-full h-full rounded-lg overflow-hidden z-[1] aspect-[4/2] relative">
                        {!!image ? <Image
                            data-episode-card-image
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
                            data-episode-card-image-bottom-gradient
                            className="h-full block rounded-lg absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                        ></div>}
                        {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                        <EpisodeItemBottomGradient />

                        {(serverStatus?.settings?.library?.enableWatchContinuity && !!percentageComplete) &&
                            <div
                                data-episode-card-progress-bar-container
                                className="absolute bottom-0 left-0 w-full z-[3]"
                                data-episode-number={episodeNumber}
                                data-media-id={anime?.id}
                                data-progress-total={progressTotal}
                                data-progress-number={progressNumber}
                            >
                                <ProgressBar value={percentageComplete} size="xs" />
                                {minutesRemaining && <div className="absolute bottom-2 right-2">
                                    <p className="text-[--muted] text-sm">{minutesRemaining}m left</p>
                                </div>}
                            </div>}

                        <div
                            data-episode-card-action-icon
                            className={cn(
                                "group-hover/episode-card:opacity-100 text-6xl text-gray-200",
                                "cursor-pointer opacity-0 transition-opacity bg-gray-950 bg-opacity-60 z-[2] absolute w-[105%] h-[105%] items-center justify-center",
                                "hidden md:flex",
                            )}
                        >
                            {actionIcon && actionIcon}
                        </div>

                        {isInvalid &&
                            <p data-episode-card-invalid-metadata className="text-red-300 opacity-50 absolute left-2 bottom-2 z-[2]">No metadata
                                                                                                                                     found</p>}
                    </div>
                    {(showAnimeInfo) ? <div data-episode-card-anime-info-container className="flex gap-3 items-center">
                        <div
                            data-episode-card-anime-image-container
                            className="flex-none w-12 aspect-[5/6] rounded-lg overflow-hidden z-[1] relative"
                        >
                            {!!anime?.image && <Image
                                data-episode-card-anime-image
                                src={getImageUrl(anime.image)}
                                alt={""}
                                fill
                                quality={100}
                                placeholder={imageShimmer(700, 475)}
                                sizes="20rem"
                                className={cn(
                                    "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                                    imageClass,
                                )}
                            />}
                        </div>
                        <Meta />
                    </div> : <Meta />}
                </div>
            </ContextMenuTrigger>
        </SeaContextMenu>
    )
}
