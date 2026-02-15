import { AL_BaseAnime } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import { ProgressBar } from "@/components/ui/progress-bar"
import { getImageUrl } from "@/lib/server/assets"
import { useThemeSettings } from "@/lib/theme/theme-hooks"
import React from "react"
import { AiFillWarning } from "react-icons/ai"
import { FaCirclePlay } from "react-icons/fa6"

type EpisodeGridItemProps = {
    media: AL_BaseAnime,
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
    percentageComplete?: number
    minutesRemaining?: number
    episodeNumber?: number
    progressNumber?: number
}

export const EpisodeGridItem = React.memo((props: EpisodeGridItemProps & React.ComponentPropsWithoutRef<"div">) => {

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
        actionIcon = props.actionIcon !== null ? <FaCirclePlay data-episode-grid-item-action-icon className="opacity-70 text-4xl" /> : undefined,
        episodeTitleClassName,
        percentageComplete,
        minutesRemaining,
        episodeNumber,
        progressNumber,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const ts = useThemeSettings()

    // const missingImage = [media?.bannerImage, media?.coverImage?.large, media?.coverImage?.extraLarge]?.filter(Boolean)?.includes(image || "")
    const missingImage = false

    return <>
        <div
            data-episode-grid-item
            data-media-id={media.id}
            data-media-type={media.type}
            data-filename={fileName}
            data-episode-number={episodeNumber}
            data-progress-number={progressNumber}
            data-is-watched={isWatched}
            data-description={description}
            data-episode-title={episodeTitle}
            data-title={title}
            data-file-name={fileName}
            data-is-invalid={isInvalid}
            data-is-filler={isFiller}
            className={cn(
                "max-w-full",
                "rounded-lg relative transition group/episode-list-item select-none",
                !!ts.libraryScreenCustomBackgroundImage && ts.libraryScreenCustomBackgroundOpacity > 5 ? "bg-[--background] p-3" : "py-3",
                "pr-12",
                disabled && "cursor-not-allowed opacity-50 pointer-events-none",
                className,
            )}
            {...rest}
        >

            {isFiller && (
                <Badge
                    data-episode-grid-item-filler-badge
                    className={cn(
                        "font-semibold absolute top-3 left-0 z-[5] text-white bg-orange-800 !bg-opacity-100 rounded-[--radius-md] text-base rounded-bl-none rounded-tr-none",
                        "lg:group-hover/episode-list-item:scale-105 lg:group-hover/episode-list-item:-translate-x-0.5 lg:group-hover/episode-list-item:-translate-y-0.5 transition-transform",
                        !!ts.libraryScreenCustomBackgroundImage && ts.libraryScreenCustomBackgroundOpacity > 5 && "top-3  left-3",
                    )}
                    intent="gray"
                    size="lg"
                >Filler</Badge>
            )}

            <div
                data-episode-grid-item-container
                className={cn(
                    "flex gap-4 relative",
                )}
            >
                <div
                    data-episode-grid-item-image-container
                    className={cn(
                        "w-36 h-28 lg:w-44 lg:h-32",
                        (ts.hideEpisodeCardDescription) && "w-36 h-28 lg:w-40 lg:h-28",
                        "flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden",
                        "group/ep-item-img-container",
                        "lg:group-hover/episode-list-item:scale-105 transition-transform",
                        // onClick && "hover:ring-1 ring-[rgba(255,255,255,0.1)] ring-offset-transparent ring-offset-2",
                        onClick && "cursor-pointer",
                        {
                            "border-2 border-red-700": isInvalid,
                            "border-2 border-yellow-900": isFiller,
                            "border-2 border-[--brand]": isSelected,
                        },

                        imageContainerClassName,
                    )}
                    onClick={onClick}
                >
                    <div data-episode-grid-item-image-overlay className="absolute z-[1] rounded-[--radius-md] w-full h-full"></div>
                    <div
                        data-episode-grid-item-image-background
                        className="bg-[--background] absolute z-[0] rounded-[--radius-md] w-full h-full"
                    ></div>
                    {!!onClick && <div
                        data-episode-grid-item-action-overlay
                        className={cn(
                            "absolute inset-0 bg-gray-950 bg-opacity-60 z-[2] flex items-center justify-center",
                            "transition-opacity opacity-0 group-hover/ep-item-img-container:opacity-100",
                        )}
                    >
                        {actionIcon && actionIcon}
                    </div>}

                    {missingImage && !title?.toLowerCase?.()?.includes?.("movie") && <div
                        data-episode-card-action-icon
                        className={cn(
                            "px-6 text-gray-200",
                            "cursor-pointer bg-gray-900/50 z-[1] absolute w-[105%] h-[105%] items-center justify-center",
                            "hidden md:flex flex-col gap-1",
                        )}
                    >
                        <div className="bg-gray-900/70 px-3 py-2 rounded-lg text-center">
                            {/*{topTitle !== title && <p className="line-clamp-1 text-[--muted]">{topTitle}</p>}*/}
                            <p className="text-md tracking-wide line-clamp-1">{title}</p>
                        </div>
                    </div>}

                    {(image || media.coverImage?.large) && <SeaImage
                        data-episode-grid-item-image
                        src={getImageUrl(image || media.coverImage?.large || "")}
                        alt="episode image"
                        fill
                        quality={60}
                        placeholder={imageShimmer(700, 475)}
                        sizes="10rem"
                        className={cn("object-cover object-center transition select-none", {
                                "opacity-25 lg:group-hover/episode-list-item:opacity-100": isWatched && !isSelected,
                            },
                            // missingImage && "opacity-50",
                            "", imageClassName)}
                        data-src={image}
                    />}

                    {(serverStatus?.settings?.library?.enableWatchContinuity && !!percentageComplete && !isWatched) &&
                        <div data-episode-grid-item-progress-bar-container className="absolute bottom-0 left-0 w-full z-[3]">
                            <ProgressBar value={percentageComplete} size="xs" />
                        </div>}
                </div>

                <div data-episode-grid-item-content className="relative overflow-hidden">
                    {isInvalid && <p data-episode-grid-item-invalid-metadata className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"
                    /> Unidentified</p>}
                    {/*{isInvalid &&*/}
                    {/*    <p data-episode-grid-item-invalid-metadata className="flex gap-2 text-red-200 text-sm items-center">No metadata found</p>}*/}

                    <p
                        className={cn(
                            !episodeTitle && "text-lg font-semibold",
                            !!episodeTitle && "transition line-clamp-2 text-base text-[--muted]",
                            // { "opacity-50 group-hover/episode-list-item:opacity-100": isWatched },
                        )}
                        data-episode-grid-item-title
                    >
                        <span
                            className={cn(
                                "font-medium text-[--foreground]",
                                isSelected && "text-[--brand]",
                            )}
                        >
                            {title?.replaceAll("`", "'")}</span>{(!!episodeTitle && !!length) &&
                        <span className="ml-4">{length}m</span>}
                    </p>

                    {!!episodeTitle && episodeTitle !== title &&
                        <p
                            data-episode-grid-item-episode-title
                            className={cn("text-md font-medium lg:text-lg text-gray-300 line-clamp-2 lg:!leading-6",
                                episodeTitleClassName)}
                        >{episodeTitle?.replaceAll("`", "'")}</p>}


                    {!!description && !ts.hideEpisodeCardDescription &&
                        <p data-episode-grid-item-episode-description className="text-sm text-[--muted] line-clamp-2">{description.replaceAll("`",
                            "'")}</p>}
                    {!!fileName && !ts.hideDownloadedEpisodeCardFilename &&
                        <p data-episode-grid-item-filename className="text-xs tracking-wider opacity-75 line-clamp-1 mt-1">{fileName}</p>}
                    {children && children}
                </div>
            </div>

            {action && <div data-episode-grid-item-action className="absolute right-1 top-1 flex flex-col items-center">
                {action}
            </div>}
        </div>
    </>

})
