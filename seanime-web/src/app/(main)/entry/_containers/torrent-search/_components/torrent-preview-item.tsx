import { cn } from "@/components/ui/core/styling"
import Image from "next/image"
import React, { memo } from "react"
import { AiFillWarning } from "react-icons/ai"
import { FcFolder } from "react-icons/fc"
import { MdVerified } from "react-icons/md"

type TorrentPreviewItemProps = {
    isSelected?: boolean
    isInvalid?: boolean
    className?: string
    onClick?: () => void
    releaseGroup: string
    isBatch: boolean
    filename: string
    title: string
    children?: React.ReactNode
    action?: React.ReactNode
    image?: string | null
    fallbackImage?: string
    isBestRelease?: boolean
    confirmed?: boolean
}

export const TorrentPreviewItem = memo((props: TorrentPreviewItemProps) => {

    const {
        isSelected,
        isInvalid,
        className,
        onClick,
        releaseGroup,
        isBatch,
        title,
        filename,
        children,
        action,
        image,
        fallbackImage,
        isBestRelease,
        confirmed,
    } = props

    const _title = isBatch ? "" : title

    return (
        <div
            className={cn(
                "border p-3 pr-12 rounded-lg relative transition lg:hover:scale-[1.01] group/torrent-preview-item overflow-hidden",
                "max-w-full",
                {
                    "border-brand-200": isSelected,
                    "hover:border-gray-500": !isSelected,
                    "border-red-700": isInvalid,
                    // "opacity-50": isWatched && !isSelected,
                }, className,
            )}
        >

            {confirmed && <div className="absolute left-2 top-2">
                <MdVerified
                    className={cn(
                        "text-[--green] text-xl",
                        isBestRelease && "text-[--pink]",
                    )}
                />
            </div>}

            <div className="absolute left-0 top-0 w-full h-full max-w-[180px]">
                {(confirmed ? !!image : !!fallbackImage) && <Image
                    src={confirmed ? image! : fallbackImage!}
                    alt="episode image"
                    fill
                    className={cn(
                        "object-cover object-center absolute w-full h-full  group-hover/torrent-preview-item:blur-0 transition-opacity opacity-20 group-hover/torrent-preview-item:opacity-40 z-[0] select-none pointer-events-none",
                        isSelected && "opacity-50",
                    )}
                />}
                <div
                    className="transition-colors absolute w-full h-full bg-gradient-to-l from-[--background] hover:from-[var(--hover-from-background-color)] to-transparent z-[1] select-none pointer-events-none"
                ></div>
            </div>

            {/*<div*/}
            {/*    className="absolute w-[calc(100%_-_179px)] h-full bg-[--background] top-0 left-[179px]"*/}
            {/*></div>*/}

            <div
                className={cn(
                    "flex gap-4 relative z-[2]",
                    { "cursor-pointer": !!onClick },
                )}
                onClick={onClick}
            >


                <div
                    className={cn(
                        "h-24 w-24 lg:w-28 flex-none rounded-md object-cover object-center relative overflow-hidden",
                        "flex items-center justify-center",
                        "text-xs px-2",
                    )}
                >
                    <p
                        className={cn(
                            "z-[1] font-bold truncate flex items-center max-w-full w-fit px-2 py-1 rounded-md",
                            "border-transparent bg-transparent",
                            // "group-hover/torrent-preview-item:bg-gray-950/50 group-hover/torrent-preview-item:text-white",
                        )}
                    >
                        <span className="truncate">{releaseGroup}</span>
                    </p>
                    {isBatch && <FcFolder className="text-7xl absolute opacity-30" />}
                </div>

                <div className="relative overflow-hidden">
                    {isInvalid && <p className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"
                    /> Unidentified</p>}
                    <h4 className={cn("font-medium text-base transition line-clamp-2")}>{_title}</h4>

                    {!!filename && <p
                        className={cn(
                            "text-sm group-hover/torrent-preview-item:text-gray-200 line-clamp-2 mb-2 break-all",
                            !(_title) ? "text-gray-200 text-base" : "text-[--muted]",
                        )}
                    >
                        {filename}
                    </p>}

                    <div className="flex items-center gap-2">
                        {children && children}
                    </div>
                </div>
            </div>

            {action && <div className="absolute right-1 top-1 flex flex-col items-center">
                {action}
            </div>}
        </div>
    )

})
