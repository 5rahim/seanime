import { cn } from "@/components/ui/core/styling"
import { getScoreColor } from "@/lib/helpers/score"
import React from "react"
import { BiSolidStar, BiStar } from "react-icons/bi"

type MediaEntryScoreBadgeProps = {
    score?: number // 0-100
}

export const MediaEntryScoreBadge = (props: MediaEntryScoreBadgeProps) => {
    const { score } = props

    if (!score) return null
    return (
        <div
            className={cn(
                "backdrop-blur-lg inline-flex items-center justify-center border gap-1 w-14 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                getScoreColor(score, "user"),
            )}
        >
            {score >= 90 ? <BiSolidStar className="text-xs" /> : <BiStar className="text-xs" />} {(score === 0) ? "-" : score / 10}
        </div>
    )
}
