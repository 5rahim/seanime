import { Anime_Episode } from "@/api/generated/types"
import { EpisodeItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { cn } from "@/components/ui/core/styling"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"

type SliderEpisodeItemProps = {
    episode: Anime_Episode
    onPlay?: ({ path }: { path: string }) => void
} & Omit<React.ComponentPropsWithoutRef<"div">, "onPlay">

export const SliderEpisodeItem = React.forwardRef<HTMLDivElement, SliderEpisodeItemProps>(({ episode, onPlay, ...rest }, ref) => {

    // const date = episode.episodeMetadata?.airDate ? new Date(episode.episodeMetadata.airDate) : undefined
    const offset = episode.progressNumber - episode.episodeNumber

    return (
        <div
            ref={ref}
            key={episode.localFile?.path}
            className={cn(
                "rounded-[--radius-md] border overflow-hidden aspect-[4/2] relative flex items-end flex-none group/missed-episode-item cursor-pointer",
                "select-none",
                "focus-visible:ring-2 ring-[--brand]",
                "w-full",
            )}
            onClick={() => onPlay?.({ path: episode.localFile?.path ?? "" })}
            tabIndex={0}
            {...rest}
        >
            <div className="absolute w-full h-full overflow-hidden z-[1]">
                {!!episode.episodeMetadata?.image ? <SeaImage
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
                {/*[CUSTOM UI] BOTTOM GRADIENT*/}
                <EpisodeItemBottomGradient />
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
                        <span>{episode.displayTitle} {!!episode.baseAnime?.episodes &&
                            (episode.baseAnime.episodes != 1 &&
                                <span className="opacity-40">/{` `}{episode.baseAnime.episodes - offset}</span>)}
                        </span>
                    </p>
                    <div className="flex flex-1"></div>
                    {!!episode.episodeMetadata?.length &&
                        <p className="text-[--muted] text-sm md:text-base">{episode.episodeMetadata?.length + "m" || ""}</p>}
                </div>
                {episode.isInvalid && <p className="text-red-300">No metadata found</p>}
            </div>
        </div>
    )
})
