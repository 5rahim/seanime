import { Anime_Episode } from "@/api/generated/types"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-editor-manager"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { cn } from "@/components/ui/core/styling"
import { ProgressBar } from "@/components/ui/progress-bar"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/hooks"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { BiAddToQueue } from "react-icons/bi"
import { LuDock, LuEye } from "react-icons/lu"
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
    allowAnimeInfo?: boolean
    forceSingleContainer?: boolean
    additionalContextMenuItems?: React.ReactNode
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
        allowAnimeInfo,
        forceSingleContainer,
        anime,
        episode,
        additionalContextMenuItems,
        ...rest
    } = props

    const router = useRouter()
    const pathname = usePathname()
    const serverStatus = useServerStatus()
    const ts = useThemeSettings()
    const { setPreviewModalMediaId } = useMediaPreviewModal()
    const { selectEpisodeToAddAndOpenEditor } = usePlaylistEditorManager()

    const showAnimeInfo = ts.showEpisodeCardAnimeInfo && !!anime && allowAnimeInfo
    const showTotalEpisodes = React.useMemo(() => !!progressTotal && progressTotal > 1, [progressTotal])
    const offset = React.useMemo(() => hasDiscrepancy ? 1 : 0, [hasDiscrepancy])

    const isSingleContainer = ts.useLegacyEpisodeCard || forceSingleContainer

    const Meta = () => (
        <div data-episode-card-info-container className="relative z-[3] w-full space-y-0">
            <p
                data-episode-card-title
                className={cn(
                    "w-[80%] line-clamp-1 text-md md:text-lg transition-colors duration-200 text-[--foreground] font-semibold",
                    isSingleContainer && "text-sm max-w-[80%] text-white/60",
                )}
            >
                {topTitle?.replaceAll("`", "'")}
            </p>
            <div data-episode-card-info-content className="w-full justify-between flex flex-none items-center">
                <p data-episode-card-subtitle className="line-clamp-1 flex items-center">
                    <span className="flex-none text-base md:text-xl font-medium">{title}{showTotalEpisodes ?
                        <span className="opacity-40">{` / `}{progressTotal! - offset}</span>
                        : ``}</span>
                    <span className="text-[--muted] text-base md:text-lg ml-2 font-normal line-clamp-1">{showAnimeInfo
                        ? "- " + anime.title
                        : ""}</span>
                </p>
                {(!!meta || !!length) && (!isSingleContainer || !minutesRemaining) && (
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
                        {!serverStatus?.isOffline && <ContextMenuItem
                            onClick={() => {
                                setPreviewModalMediaId(anime?.id || 0, "anime")
                            }}
                        >
                            <LuEye /> Preview
                        </ContextMenuItem>}
                        <ContextMenuItem
                            onClick={() => {
                                if (!serverStatus?.isOffline) {
                                    router.push(`/entry?id=${anime?.id}`)
                                } else {
                                    router.push(`/offline/entry/anime?id=${anime?.id}`)
                                }
                            }}
                        >
                            <LuDock /> Open page
                        </ContextMenuItem>
                    </>}
                    {(props.episode && anime?.id && props.episode?.aniDBEpisode) && <ContextMenuItem
                        onClick={() => {
                            selectEpisodeToAddAndOpenEditor(anime.id!, props.episode?.aniDBEpisode!)
                        }}
                    >
                        <BiAddToQueue /> Add to Playlist
                    </ContextMenuItem>}

                    {additionalContextMenuItems}

                    <PluginEpisodeCardContextMenuItems episode={props.episode} />

                </ContextMenuGroup>
            }
        >
            <ContextMenuTrigger>
                <div
                    ref={mRef}
                    className={cn(
                        "rounded-xl overflow-hidden space-y-2 flex-none group/episode-card cursor-pointer",
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
                    <div
                        data-episode-card-image-container
                        className="w-full h-full rounded-xl overflow-hidden z-[1] aspect-[4/2] relative bg-[--background]"
                    >
                        {!!image ? <SeaImage
                            data-episode-card-image
                            src={getImageUrl(image)}
                            alt={""}
                            fill
                            quality={100}
                            placeholder={imageShimmer(700, 475)}
                            sizes="20rem"
                            className={cn(
                                "object-cover rounded-xl object-center transition lg:group-hover/episode-card:scale-[1.02] duration-500",
                                imageClass,
                            )}
                        /> : <div
                            data-episode-card-image-bottom-gradient
                            className="h-full block rounded-xl absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                        ></div>}
                        {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                        <EpisodeItemBottomGradient isSingleContainer={isSingleContainer} className="rounded-b-xl" />

                        {isSingleContainer && (
                            <div className="absolute bottom-0 left-0 w-full h-fit z-[3] p-3">
                                <Meta />
                            </div>
                        )}

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
                                {minutesRemaining && <div
                                    className={cn(
                                        "absolute bottom-2 right-2 text-[--muted]",
                                        isSingleContainer && "right-4 bottom-4 ",
                                    )}
                                >
                                    <span>{minutesRemaining}m left</span>
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
                    {(showAnimeInfo && !isSingleContainer) ? <div data-episode-card-anime-info-container className="flex gap-3 items-center">
                        <div
                            data-episode-card-anime-image-container
                            className="flex-none w-12 aspect-[5/6] rounded-lg overflow-hidden z-[1] relative hidden"
                        >
                            {!!anime?.image && <SeaImage
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
                    </div> : !isSingleContainer ? <Meta /> : null}
                </div>
            </ContextMenuTrigger>
        </SeaContextMenu>
    )
}
