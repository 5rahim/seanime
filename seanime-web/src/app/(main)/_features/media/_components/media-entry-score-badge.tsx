import { cn } from "@/components/ui/core/styling"
import React from "react"
import { BiStar } from "react-icons/bi"

type MediaEntryScoreBadgeProps = {
    score?: number
}

export const MediaEntryScoreBadge = (props: MediaEntryScoreBadgeProps) => {
    const { score } = props

    if (!score) return null

    const scoreColor = score ? (
        score < 5 ? "bg-gray-500" :
            score < 7 ? "bg-gray-500" :
                score < 9 ? "bg-green-500" :
                    "bg-indigo text-white bg-opacity-80"
    ) : ""

    return (
        <div
            className={cn(
                "backdrop-blur-lg inline-flex items-center justify-center gap-1 w-14 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                scoreColor,
            )}
        >
            <BiStar /> {(score === 0) ? "-" : score}
        </div>
    )
}
