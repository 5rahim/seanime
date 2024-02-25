import { cn } from "@/components/ui/core/styling"
import { AiFillWarning } from "react-icons/ai"
import React, { memo } from "react"
import { FcFolder } from "react-icons/fc"

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
    } = props

    return (
        <div
            className={cn(
                "border border-[--border] p-3 pr-12 rounded-lg relative transition hover:bg-gray-900 group/episode-list-item",
                {
                    "border-brand-200 bg-gray-800 hover:bg-gray-800": isSelected,
                    "hover:border-gray-500": !isSelected,
                    "border-red-700": isInvalid,
                    // "opacity-50": isWatched && !isSelected,
                }, className,
            )}
        >
            <div
                className={cn(
                    "flex gap-4 relative",
                    { "cursor-pointer": !!onClick },
                )}
                onClick={onClick}
            >
                <div
                    className={cn(
                        "h-24 w-24 flex-none rounded-md object-cover object-center relative overflow-hidden",
                        "flex items-center justify-center bg-gray-800",
                        "text-xs px-2",
                    )}
                >
                    <p className={cn(
                        "z-[1] font-bold line-clamp-1",
                        {
                            "text-brand-200": releaseGroup.toLowerCase() === "subsplease",
                        },
                    )}>{releaseGroup}</p>
                    {!!image && <img
                        src={image}
                        alt="episode image"
                        className="object-cover object-center absolute w-full h-full blur-xs opacity-20 z-[0] select-none pointer-events-none"
                        data-src={image}
                    />}
                    {isBatch && <FcFolder className="text-7xl absolute opacity-20"/>}
                </div>

                <div className="relative overflow-hidden">
                    {isInvalid && <p className="flex gap-2 text-red-300 items-center"><AiFillWarning
                        className="text-lg text-red-500"/> Unidentified</p>}
                    <h4 className={cn("font-medium transition line-clamp-2")}>{isBatch ? "Batch" : title}</h4>

                    {!!filename && <p className={cn("text-sm text-gray-400 line-clamp-2 mb-2")}>{filename}</p>}

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
