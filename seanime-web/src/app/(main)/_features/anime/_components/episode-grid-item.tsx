import { AL_BaseMedia } from "@/api/generated/types"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { useThemeSettings } from "@/lib/theme/hooks"
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
    className?: string
    disabled?: boolean
    actionIcon?: React.ReactElement | null
    isFiller?: boolean
    length?: string | number | null
    imageClassName?: string
    imageContainerClassName?: string
    episodeTitleClassName?: string
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
        length,
        actionIcon = props.actionIcon !== null ? <AiFillPlayCircle className="opacity-70 text-4xl" /> : undefined,
        episodeTitleClassName,
        ...rest
    } = props

    const ts = useThemeSettings()

    return <>
        <div
            className={cn(
                "max-w-full",
                "rounded-lg relative transition group/episode-list-item select-none",
                !!ts.libraryScreenCustomBackgroundImage && ts.libraryScreenCustomBackgroundOpacity > 5 ? "bg-[--background] p-3" : "py-3",
                "pr-12",
                className,
            )}
            {...rest}
        >

            {isFiller && (
                <Badge
                    className={cn(
                        "font-semibold absolute top-3 left-0 z-[5] text-white bg-orange-800 !bg-opacity-100 rounded-md text-base rounded-bl-none rounded-tr-none",
                        !!ts.libraryScreenCustomBackgroundImage && ts.libraryScreenCustomBackgroundOpacity > 5 && "top-3  left-3",
                    )}
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
                        "w-36 lg:w-40 h-28 flex-none rounded-md object-cover object-center relative overflow-hidden cursor-pointer",
                        "group/ep-item-img-container",

                        {
                            "border-2 border-red-700": isInvalid,
                            "border-2 border-yellow-900": isFiller,
                            "border-2 border-[--brand]": isSelected,
                        },

                        imageContainerClassName,
                    )}
                    onClick={onClick}
                >
                    <div className="bg-[--background] absolute z-[0] rounded-md w-full h-full"></div>
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
                        className={cn("object-cover object-center transition select-none", {
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

                    <p
                        className={cn(
                            !episodeTitle && "text-lg font-semibold",
                            !!episodeTitle && "transition line-clamp-2 text-base text-[--muted]",
                            // { "opacity-50 group-hover/episode-list-item:opacity-100": isWatched },
                        )}
                    >
                        <span
                            className={cn(
                                "font-medium text-white",
                                isSelected && "text-[--brand]",
                            )}
                        >
                            {title?.replaceAll("`", "'")}</span>{(!!episodeTitle && !!length) &&
                        <span className="ml-4">{length}m</span>}
                    </p>

                    {!!episodeTitle &&
                        <p
                            className={cn("text-md font-semibold lg:text-lg text-gray-300 line-clamp-2",
                                episodeTitleClassName)}
                        >{episodeTitle?.replaceAll("`", "'")}</p>}


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
