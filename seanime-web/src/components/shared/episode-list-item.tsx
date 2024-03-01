import { imageShimmer } from "@/components/shared/image-helpers"
import { cn } from "@/components/ui/core/styling"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle, AiFillWarning } from "react-icons/ai"

type EpisodeListItemProps = {
    media: BaseMediaFragment,
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
}

export const EpisodeListItem: React.FC<EpisodeListItemProps & React.ComponentPropsWithoutRef<"div">> = (props) => {

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
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-70 text-4xl" /> : undefined,
        ...rest
    } = props

    return <>
        <div
            className={cn(
                "border p-3 pr-12 rounded-lg relative transition hover:bg-gray-900 group/episode-list-item",
                {
                    "border-brand-200 bg-gray-800 hover:bg-gray-800": isSelected,
                    "border-red-700": isInvalid,
                    // "opacity-50": isWatched && !isSelected,
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
                        "h-24 w-24 flex-none rounded-md object-cover object-center relative overflow-hidden cursor-pointer",
                        "group/ep-item-img-container",
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
                    {(image || media.coverImage?.medium) && <Image
                        src={image || media.coverImage?.medium || ""}
                        alt="episode image"
                        fill
                        quality={60}
                        placeholder={imageShimmer(700, 475)}
                        sizes="10rem"
                        className={cn("object-cover object-center transition", {
                            "opacity-30 group-hover/episode-list-item:opacity-100": isWatched,
                        }, imageClassName)}
                        data-src={image}
                    />}
                </div>
                {(image && unoptimizedImage) && <div
                    className="h-24 w-24 flex-none rounded-md object-cover object-center relative overflow-hidden"
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

                    <h4
                        className={cn("font-medium transition line-clamp-2",
                            { "opacity-50 group-hover/episode-list-item:opacity-100": isWatched })}
                    >{title}</h4>

                    {!!episodeTitle && <p className={cn("text-md text-[--muted] line-clamp-2")}>{episodeTitle}</p>}

                    {!!fileName && <p className="text-sm text-gray-600 line-clamp-1">{fileName}</p>}
                    {!!description && <p className="text-sm text-gray-500 line-clamp-2">{description}</p>}
                    {children && children}
                </div>
            </div>

            {action && <div className="absolute right-1 top-1 flex flex-col items-center">
                {action}
            </div>}
        </div>
    </>

}
