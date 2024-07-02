import { AL_BaseMedia } from "@/api/generated/types"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle, AiFillWarning } from "react-icons/ai"

type EpisodeGridItemProps = {
    media: AL_BaseMedia,
    children?: React.ReactNode
    action?: React.ReactNode
    image?: string | null
    onClick?: () => void
    title: string,
    episodeTitle?: string | null
    description?: string | null
    fileName?: string
    isWatched?: boolean
    isSelected?: boolean
    unoptimizedImage?: boolean
    isInvalid?: boolean
    imageClassName?: string
    imageContainerClassName?: string
    className?: string
    disabled?: boolean
    actionIcon?: React.ReactElement | null
    isFiller?: boolean
}

export const EpisodeGridItem: React.FC<EpisodeGridItemProps & React.ComponentPropsWithoutRef<"div">> = (props) => {

    const {
        children,
        action,
        image,
        onClick,
        episodeTitle,
        description,
        title,
        fileName,
        media,
        isWatched,
        isSelected,
        unoptimizedImage,
        isInvalid,
        imageClassName,
        imageContainerClassName,
        className,
        disabled,
        isFiller,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-70 text-4xl" /> : undefined,
        ...rest
    } = props

    return <>
        <div
            className={cn(
                "bg-[--background] hover:bg-[var(--hover-from-background-color)] max-w-full",
                "border p-3 pr-12 rounded-lg relative transition group/episode-list-item",
                {
                    // "opacity-50": isWatched && !isSelected,
                    "border-zinc-500 bg-gray-900 hover:bg-gray-900": isSelected,
                    "border-red-700": isInvalid,
                    "border-yellow-900": isFiller,
                    // "opacity-50 pointer-events-none": disabled,
                    // "opacity-50": isWatched && !isSelected,
                }, className,
            )}
            {...rest}
        >
            {/*{isCompleted && <div className="absolute top-1 left-1 w-full h-1 bg-brand rounded-full"/>}*/}

            {isFiller && (
                <Badge
                    className="font-semibold absolute top-0 left-0 z-[5] text-white bg-yellow-900 !bg-opacity-100 rounded-md text-base rounded-bl-none rounded-tr-none"
                    intent="gray"
                    size="lg"
                >Filler</Badge>
            )}

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
                            "opacity-25 group-hover/episode-list-item:opacity-100": isWatched,
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
                        className={cn(
                            "font-medium transition line-clamp-2",
                            // { "opacity-50 group-hover/episode-list-item:opacity-100": isWatched },
                        )}
                    >{title?.replaceAll("`", "'")}</h4>

                    {!!episodeTitle && <p className={cn("text-sm lg:text-md text-gray-300 line-clamp-2")}>{episodeTitle?.replaceAll("`", "'")}</p>}

                    {!!fileName && <p className="text-sm text-[--muted] line-clamp-1">{fileName}</p>}
                    {!!description && <p className="text-sm text-[--muted] line-clamp-2">{description.replaceAll("`", "'")}</p>}
                    {children && children}
                </div>
            </div>

            {action && <div className="absolute right-1 top-1 flex flex-col items-center">
                {action}
            </div>}
        </div>
    </>

}
