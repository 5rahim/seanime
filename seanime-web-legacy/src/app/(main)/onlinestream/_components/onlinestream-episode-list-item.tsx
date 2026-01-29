import { AL_BaseAnime } from "@/api/generated/types"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { cn } from "@/components/ui/core/styling"
import React from "react"
import { AiFillPlayCircle, AiFillWarning } from "react-icons/ai"

type EpisodeListItemProps = {
    media: AL_BaseAnime,
    children?: React.ReactNode
    action?: React.ReactNode
    image?: string | null
    onClick?: () => void
    title: string,
    episodeTitle?: string | null
    description?: string | null
    fileName?: string
    isSelected?: boolean
    isWatched?: boolean
    unoptimizedImage?: boolean
    isInvalid?: boolean
    imageClassName?: string
    imageContainerClassName?: string
    className?: string
    actionIcon?: React.ReactElement | null
    disabled?: boolean
}

export const OnlinestreamEpisodeListItem: React.FC<EpisodeListItemProps & React.ComponentPropsWithoutRef<"div">> = (props) => {

    const {
        children,
        action,
        image,
        onClick,
        episodeTitle,
        description,
        title,
        fileName,
        isSelected,
        media,
        isWatched,
        unoptimizedImage,
        isInvalid,
        imageClassName,
        imageContainerClassName,
        className,
        disabled,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-70 text-4xl" /> : undefined,
        ...rest
    } = props

    return <>
        <div
            className={cn(
                "border p-3 pr-12 rounded-lg relative transition hover:bg-gray-900 group/episode-list-item bg-[--background]",
                {
                    "border-zinc-500 bg-gray-900 hover:bg-gray-900": isSelected,
                    "border-red-700": isInvalid,
                    "opacity-50 pointer-events-none": disabled,
                    "opacity-50": isWatched && !isSelected,
                }, className,
            )}
            {...rest}
        >
            {/*{isCompleted && <div className="absolute top-1 left-1 w-full h-1 bg-brand rounded-full"/>}*/}

            <div
                className={cn(
                    "flex gap-4 relative",
                )}
            >
                <div
                    className={cn(
                        "size-20 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden",
                        "group/ep-item-img-container",
                        !disabled && "cursor-pointer",
                        disabled && "pointer-events-none",
                        imageContainerClassName,
                    )}
                    onClick={onClick}
                >
                    {!!onClick && <div
                        className={cn(
                            "absolute inset-0 bg-gray-950 bg-opacity-60 z-[1] flex items-center justify-center",
                            "transition-opacity opacity-0 group-hover/ep-item-img-container:opacity-100",
                        )}
                    >
                        {actionIcon && actionIcon}
                    </div>}
                    {(image || media.coverImage?.medium) && <SeaImage
                        src={image || media.coverImage?.medium || ""}
                        alt="episode image"
                        fill
                        quality={60}
                        placeholder={imageShimmer(700, 475)}
                        sizes="10rem"
                        className={cn("object-cover object-center transition", {
                            "opacity-25 group-hover/episode-list-item:opacity-100": isWatched,
                        }, imageClassName)}
                        data-src={image}
                    />}
                </div>
                {(image && unoptimizedImage) && <div
                    className="h-24 w-24 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                >
                    <img
                        src={image}
                        alt="episode image"
                        className="object-cover object-center absolute w-full h-full"
                        data-src={image}
                    />
                </div>}

                <div className="relative overflow-hidden">
                    {isInvalid && <p className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"
                    /> Unidentified</p>}
                    {isInvalid && <p className="flex gap-2 text-red-200 text-sm items-center">No metadata found</p>}

                    <p
                        className={cn(
                            "font-medium transition text-lg line-clamp-2",
                            { "text-[--muted]": !isSelected },
                            // { "opacity-50 group-hover/episode-list-item:opacity-100": isWatched },
                        )}
                    >{!!title ? title?.replaceAll("`", "'") : "No title"}</p>

                    {!!episodeTitle && <p className={cn("text-sm text-[--muted] line-clamp-2")}>{episodeTitle?.replaceAll("`", "'")}</p>}

                    {!!fileName && <p className="text-sm text-gray-600 line-clamp-1">{fileName}</p>}
                    {!!description && <p className="text-sm text-[--muted] line-clamp-1 italic">{description}</p>}
                    {children && children}
                </div>
            </div>

            {action && <div className="absolute right-1 top-1 flex flex-col items-center">
                {action}
            </div>}
        </div>
    </>

}
