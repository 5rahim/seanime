import { cn } from "@/components/ui/core/styling"
import { getScoreColor } from "@/lib/helpers/score"
import React from "react"
import { BiSolidStar, BiStar } from "react-icons/bi"

type MediaEntryScoreBadgeProps = {
    isMediaCard?: boolean
    score?: number // 0-100
}

export const MediaEntryScoreBadge = (props: MediaEntryScoreBadgeProps) => {
    const { score, isMediaCard } = props

    if (!score) return null
    return (
        <div
            data-media-entry-score-badge
            className={cn(
                "backdrop-blur-lg inline-flex items-center justify-center border gap-1 w-14 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                isMediaCard && "rounded-none rounded-tl-lg border-none",
                getScoreColor(score, "user"),
            )}
        >
            {score >= 90 ? <BiSolidStar className="text-sm" /> : <BiStar className="text-sm" />} {(score === 0) ? "-" : score / 10}
        </div>
    )
}
