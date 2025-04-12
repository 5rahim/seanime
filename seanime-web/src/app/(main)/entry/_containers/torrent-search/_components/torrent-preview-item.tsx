import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import Image from "next/image"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { BiLinkExternal } from "react-icons/bi"
import { BsFileEarmarkPlayFill } from "react-icons/bs"
import { FcFolder } from "react-icons/fc"
import { MdVerified } from "react-icons/md"

type TorrentPreviewItemProps = {
    link?: string
    isSelected?: boolean
    isInvalid?: boolean
    className?: string
    onClick?: () => void
    releaseGroup: string
    isBatch: boolean
    subtitle: string
    title: string
    children?: React.ReactNode
    action?: React.ReactNode
    image?: string | null
    fallbackImage?: string
    isBestRelease?: boolean
    confirmed?: boolean
    addon?: React.ReactNode
    isBasic?: boolean
}

export const TorrentPreviewItem = memo((props: TorrentPreviewItemProps) => {

    const {
        link,
        isBasic,
        isSelected,
        isInvalid,
        className,
        onClick,
        releaseGroup,
        isBatch,
        title,
        subtitle,
        children,
        action,
        image,
        fallbackImage,
        isBestRelease,
        confirmed,
        addon,
    } = props

    const _title = isBatch ? "" : title

    return (
        <div
            data-torrent-preview-item
            data-title={title}
            data-subtitle={subtitle}
            data-release-group={releaseGroup}
            data-is-batch={isBatch}
            data-is-best-release={isBestRelease}
            data-confirmed={confirmed}
            data-is-invalid={isInvalid}
            data-is-selected={isSelected}
            data-link={link}
            className={cn(
                "border p-3 pr-12 rounded-lg relative transition lg:hover:scale-[1.01] group/torrent-preview-item overflow-hidden",
                "max-w-full bg-[--background]",
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

            {confirmed && <div className="absolute left-2 top-2" data-torrent-preview-item-confirmed-badge>
                <MdVerified
                    className={cn(
                        "text-[--gray] text-lg",
                        isBestRelease ? "text-[--pink]" : "opacity-30",
                    )}
                />
            </div>}

            <div className="absolute left-0 top-0 w-full h-full max-w-[180px]" data-torrent-preview-item-image-container>
                {(confirmed ? !!image : !!fallbackImage) && <Image
                    data-torrent-preview-item-image
                    src={confirmed ? image! : fallbackImage!}
                    alt="episode image"
                    fill
                    className={cn(
                        "object-cover object-center absolute w-full h-full group-hover/torrent-preview-item:blur-0 transition-opacity opacity-25 group-hover/torrent-preview-item:opacity-60 z-[0] select-none pointer-events-none",
                        isSelected && "opacity-50",
                    )}
                />}
                <div
                    data-torrent-preview-item-image-bottom-gradient
                    className="transition-colors absolute w-full h-full bg-gradient-to-l from-[--background] hover:from-[var(--hover-from-background-color)] to-transparent z-[1] select-none pointer-events-none"
                ></div>
            </div>

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


                <div
                    data-torrent-preview-item-release-info-container
                    className={cn(
                        "h-24 w-24 lg:w-28 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden",
                        "flex items-center justify-center",
                        "text-xs px-2",
                        isBasic && "h-20",
                    )}
                >
                    <p
                        className={cn(
                            "z-[1] font-bold truncate flex items-center max-w-full w-fit px-2 py-1 rounded-[--radius-md]",
                            "border-transparent bg-transparent",
                            // "group-hover/torrent-preview-item:bg-gray-950/50 group-hover/torrent-preview-item:text-white",
                        )}
                        data-torrent-preview-item-release-group
                    >
                        <span className="truncate">{releaseGroup}</span>
                    </p>
                    {isBatch && <FcFolder className="text-7xl absolute opacity-20 group-hover/torrent-preview-item:opacity-30" />}
                    {!(image || fallbackImage) && !isBatch && <BsFileEarmarkPlayFill className="text-7xl absolute opacity-10" />}
                </div>

                <div className="relative overflow-hidden space-y-1" data-torrent-preview-item-metadata>
                    {isInvalid && <p className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"
                    /> Unidentified</p>}
                    <p
                        className={cn(
                            "font-medium text-base transition line-clamp-2 tracking-wider",
                            isBasic && "text-sm",
                        )}
                        data-torrent-preview-item-title
                    >{_title}</p>

                    {!!subtitle && <p
                        className={cn(
                            "text-sm tracking-wide group-hover/torrent-preview-item:text-gray-200 line-clamp-2 break-all",
                            !(_title) ? "font-medium transition tracking-wider" : "text-[--muted]",
                        )}
                        data-torrent-preview-item-subtitle
                    >
                        {subtitle}
                    </p>}

                    <div className="flex items-center gap-2" data-torrent-preview-item-subcontent>
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
