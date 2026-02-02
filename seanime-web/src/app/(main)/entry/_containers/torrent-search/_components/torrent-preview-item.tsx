import { AL_BaseAnime, Anime_Episode, Habari_Metadata, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentParsedMetadata,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { SeaImage } from "@/components/shared/sea-image"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import { uniqBy } from "lodash"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { BiCalendarAlt, BiLinkExternal } from "react-icons/bi"
import { BsFileEarmarkPlayFill } from "react-icons/bs"
import { FcOpenedFolder } from "react-icons/fc"
import { LuCircleCheckBig, LuGem } from "react-icons/lu"

export const TorrentList = ({ children }: { children?: React.ReactNode }) => {
    return (
        <div className="grid grid-cols-1 gap-3">
            {children}
        </div>
    )
}

export const TorrentListItem = ({ torrent, metadata, debridCached, onClick, isSelected, episode, media, overrideProps, extensionName }: {
    torrent: HibikeTorrent_AnimeTorrent,
    metadata: Habari_Metadata | undefined,
    debridCached: boolean | undefined,
    episode: Anime_Episode | undefined,
    media: AL_BaseAnime | undefined,
    isSelected: boolean
    onClick: () => void
    overrideProps?: Partial<TorrentPreviewItemProps>
    extensionName?: string
}) => {
    return (
        <TorrentPreviewItem
            link={overrideProps?.link ?? torrent?.link}
            confirmed={overrideProps?.confirmed ?? torrent?.confirmed}
            key={torrent.infoHash}
            displayName={overrideProps?.displayName ?? (episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || "")}
            releaseGroup={overrideProps?.releaseGroup ?? (torrent.releaseGroup || "")}
            torrentName={overrideProps?.torrentName ?? torrent.name}
            isBatch={overrideProps?.isBatch ?? torrent.isBatch ?? false}
            isBestRelease={overrideProps?.isBestRelease ?? torrent.isBestRelease}
            image={overrideProps?.image ?? (episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large ||
                (torrent.confirmed ? (media?.coverImage?.large || media?.bannerImage) : null))}
            fallbackImage={overrideProps?.fallbackImage ?? (media?.coverImage?.large || media?.bannerImage)}
            isSelected={overrideProps?.isSelected ?? isSelected}
            metadata={overrideProps?.metadata ?? metadata}
            resolution={overrideProps?.resolution ?? torrent.resolution}
            debridCached={overrideProps?.debridCached ?? debridCached}
            onClick={onClick}
        >
            <div className="flex flex-wrap gap-2 items-center lg:absolute bottom-0 left-0 right-0">
                {torrent.isBestRelease && (
                    <Badge
                        className="rounded-[--radius-md] text-[0.8rem] bg-pink-800 border-transparent border"
                        intent="success-solid"
                        leftIcon={<LuGem className="text-md" />}
                    >
                        Highest quality
                    </Badge>
                )}
                <TorrentSeedersBadge seeders={torrent.seeders} />
                {!!torrent.size && <p className="text-gray-300 text-sm flex items-center gap-1">
                    {torrent.formattedSize}</p>}
                {torrent.date && <p className="text-[--muted] text-sm flex items-center gap-1">
                    <BiCalendarAlt /> {formatDistanceToNowSafe(torrent.date)}
                </p>}
                <div className="flex-1"></div>
                {extensionName && <p className="text-[--muted] font-bold text-sm flex items-center gap-1">{extensionName}</p>}
            </div>
            {metadata && <TorrentParsedMetadata metadata={metadata} />}
        </TorrentPreviewItem>
    )
}

type TorrentPreviewItemProps = {
    link?: string
    isSelected?: boolean
    isInvalid?: boolean
    className?: string
    onClick?: () => void
    releaseGroup: string
    isBatch: boolean
    displayName: string
    torrentName: string
    children?: React.ReactNode
    action?: React.ReactNode
    image?: string | null
    fallbackImage?: string
    isBestRelease?: boolean
    confirmed?: boolean
    addon?: React.ReactNode
    // isBasic?: boolean
    metadata?: Habari_Metadata
    resolution?: string
    debridCached?: boolean
}

const TorrentPreviewItem = memo((props: TorrentPreviewItemProps) => {

    const {
        link,
        // isBasic,
        isSelected,
        isInvalid,
        className,
        onClick,
        releaseGroup,
        isBatch,
        torrentName,
        displayName,
        children,
        action,
        image,
        fallbackImage,
        isBestRelease,
        confirmed,
        addon,
        metadata,
        resolution: _resolution,
        debridCached,
    } = props

    const resolution = _resolution || metadata?.video_resolution

    const mainTitle = React.useMemo(() => {
        const episodeNumbers = metadata?.episode_number
        if (!isBatch) {
            if (!!displayName) return displayName

            if (episodeNumbers?.length === 1) return (
                `Episode ${parseInt(episodeNumbers[0])}`
            )

            if (episodeNumbers?.length === 0) return (
                `Batch`
            )

            if (metadata?.formatted_title) return metadata.formatted_title
            return ""
        }
        let t = ""
        const seasonNumbers = metadata?.season_number
        const partNumbers = metadata?.part_number
        if (partNumbers?.length && partNumbers.length > 1) {
            const s1 = parseInt(partNumbers[0])
            const lastS = parseInt(partNumbers[partNumbers.length - 1])
            if (s1 != lastS) {
                if (uniqBy(partNumbers, n => parseInt(n)).length === 2 && lastS - s1 === 1)
                    t = `Part ${s1} and ${lastS}`
                else
                    t = `Parts ${s1} to ${lastS}`
                return t
            } else {
                return `Part ${s1}`
            }
        }
        if (seasonNumbers?.length && seasonNumbers.length > 1) {
            const s1 = parseInt(seasonNumbers[0])
            const lastS = parseInt(seasonNumbers[seasonNumbers.length - 1])
            if (s1 != lastS) {
                if (uniqBy(seasonNumbers, n => parseInt(n)).length === 2 && lastS - s1 === 1)
                    t = `Season ${s1} and ${lastS}`
                else
                    t = `Seasons ${s1} to ${lastS}`
                return t
            } else {
                return `Season ${s1}`
            }
        }
        if (episodeNumbers?.length && episodeNumbers?.length > 1) {
            t = `Episodes ${parseInt(episodeNumbers[0])} to ${parseInt(episodeNumbers[episodeNumbers.length - 1])}`
            if (seasonNumbers?.length === 1) {
                t += ` (Season ${parseInt(seasonNumbers[0])})`
            }
            return t
        } else if (seasonNumbers?.length && seasonNumbers.length === 1) {
            return `Season ${parseInt(seasonNumbers[0])}`
        }
        return "Batch"
    }, [displayName, metadata])

    return (
        <div
            data-torrent-preview-item
            data-torrent-name={torrentName}
            data-display-name={displayName}
            data-release-group={releaseGroup}
            data-is-batch={isBatch}
            data-is-best-release={isBestRelease}
            data-confirmed={confirmed}
            data-is-invalid={isInvalid}
            data-is-selected={isSelected}
            data-link={link}
            className={cn(
                "border p-3 pr-12 rounded-lg relative transition group/torrent-preview-item overflow-hidden",
                // !__isElectronDesktop__ && "lg:hover:scale-[1.01]",
                "max-w-full bg-[--background]",
                isSelected && "sticky top-2 bottom-2 z-10",
                {
                    "border-brand-200": isSelected,
                    "hover:border-gray-500": !isSelected,
                    "border-red-700": isInvalid,
                    // "opacity-50": isWatched && !isSelected,
                }, className,
            )}
            tabIndex={0}
        >

            {addon}

            {(image || fallbackImage) &&
                <div className="absolute left-0 top-0 w-full h-full max-w-[200px] overflow-hidden" data-torrent-preview-item-image-container>
                    {(image || fallbackImage) && <SeaImage
                        data-torrent-preview-item-image
                        src={image || fallbackImage!}
                        alt="episode image"
                        fill
                        className={cn(
                            "object-cover object-center absolute w-full h-full group-hover/torrent-preview-item:blur-0 transition-opacity opacity-25 group-hover/torrent-preview-item:opacity-60 z-[0] select-none pointer-events-none",
                            (!image && fallbackImage) && "opacity-10 group-hover/torrent-preview-item:opacity-30",
                            isSelected && "opacity-50",
                        )}
                    />}
                    <div
                        data-torrent-preview-item-image-end-gradient
                        className="transition-colors absolute w-full h-full -right-2 bg-gradient-to-l from-[--background] via-[--background] via-30% hover:from-[var(--hover-from-background-color)] to-transparent z-[1] select-none pointer-events-none"
                    ></div>
                </div>}
            {(image && isBatch) &&
                <div className="absolute right-0 top-0 w-full h-full max-w-[200px] overflow-hidden" data-torrent-preview-item-image-container>
                    {(image) && <SeaImage
                        data-torrent-preview-item-image
                        src={image!}
                        alt="episode image"
                        fill
                        className={cn(
                            "object-cover object-center absolute w-full h-full group-hover/torrent-preview-item:blur-0 transition-opacity opacity-25 z-[0] select-none pointer-events-none",
                            (image) && "opacity-10",
                            isSelected && "opacity-10",
                        )}
                    />}
                    <div
                        data-torrent-preview-item-image-end-gradient
                        className="transition-colors absolute w-full h-full -left-2 bg-gradient-to-r from-[--background] via-[--background] via-30% hover:from-[var(--hover-from-background-color)] to-transparent z-[1] select-none pointer-events-none"
                    ></div>
                </div>}

            {/*<div*/}
            {/*    className="absolute w-[calc(100%_-_179px)] h-full bg-[--background] top-0 left-[179px]"*/}
            {/*></div>*/}

            <div
                data-torrent-preview-item-content
                className={cn(
                    "flex gap-4 relative z-[2]",
                    { "cursor-pointer": !!onClick },
                )}
                onClick={onClick}
            >

                {isBatch &&
                    <FcOpenedFolder
                        className={cn(
                            "text-7xl absolute opacity-30 rotate-12 -left-8 -bottom-8 transform-gpu transition-all skew-x-2 group-hover/torrent-preview-item:skew-x-0 group-hover/torrent-preview-item:opacity-60",
                            isSelected && "hover:opacity-80 opacity-80",
                        )}
                    />}

                {debridCached && <div className="absolute -left-1.5 -top-1.5 z-[2]" data-torrent-preview-item-debrid-cached-badge>
                    <TorrentDebridInstantAvailabilityBadge />
                </div>}

                <div
                    data-torrent-preview-item-release-info-container
                    className={cn(
                        "h-24 w-24 lg:w-28 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden",
                        "flex flex-col items-center justify-center",
                        "text-xs",
                        // isBasic && "h-20",
                    )}
                >
                    <p
                        className={cn(
                            "z-[1] font-medium truncate flex items-center max-w-full w-fit px-0 py-1 rounded-[--radius-md] text-[.9rem]",
                            "border-transparent bg-transparent",
                            // "group-hover/torrent-preview-item:bg-gray-950/50 group-hover/torrent-preview-item:text-white",
                        )}
                        data-torrent-preview-item-release-group
                    >
                        <span className="truncate">{releaseGroup}</span>
                    </p>
                    {resolution && <div className="">
                        <TorrentResolutionBadge resolution={resolution} />
                    </div>}
                    {!(image || fallbackImage) && !isBatch && <BsFileEarmarkPlayFill className="text-7xl absolute opacity-10" />}
                </div>

                <div className="relative overflow-hidden space-y-1 w-full" data-torrent-preview-item-metadata>
                    {isInvalid && <p className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"
                    /> Unidentified</p>}

                    {mainTitle && <p
                        className={cn(
                            "font-normal text-[1.1rem] transition line-clamp-1 tracking-wide flex gap-2 items-center max-w-[20rem] 3xl:max-w-[35rem]",
                            // isBasic && "text-sm",
                        )}
                        data-torrent-preview-item-title
                    >{mainTitle} {confirmed && <span className="" data-torrent-preview-item-confirmed-badge>
                        <LuCircleCheckBig
                            className={cn(
                                "text-[--gray] text-sm",
                                isBestRelease ? "text-[--pink] opacity-70" : "opacity-30",
                            )}
                        />
                    </span>}</p>}

                    {!!torrentName && <p
                        className={cn(
                            "text-[.8rem] tracking-wide group-hover/torrent-preview-item:opacity-60 line-clamp-2 break-all",
                            "opacity-30",
                        )}
                        data-torrent-preview-item-subtitle
                    >
                        {torrentName}
                    </p>}

                    <div className="flex flex-col gap-2" data-torrent-preview-item-subcontent>
                        {children && children}
                    </div>
                </div>
            </div>

            <div className="absolute right-1 top-1 flex flex-col items-center" data-torrent-preview-item-actions>
                {link && <Tooltip
                    side="left"
                    trigger={<IconButton
                        data-torrent-preview-item-open-in-browser-button
                        icon={<BiLinkExternal className="text-[--muted]" />}
                        intent="gray-basic"
                        size="sm"
                        onClick={() => openTab(link)}
                    />}
                >Open in browser</Tooltip>}
                {action}
            </div>
        </div>
    )

})
