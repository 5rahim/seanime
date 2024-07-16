import { cn } from "@/components/ui/core/styling"
import { getScoreColor } from "@/lib/helpers/score"
import React from "react"
import { BiStar } from "react-icons/bi"

type AnimeEntryScoreBadgeProps = {
    score?: number // 0-100
}

export const AnimeEntryScoreBadge = (props: AnimeEntryScoreBadgeProps) => {
    const { score } = props

    if (!score) return null
    return (
        <div
            className={cn(
                "backdrop-blur-lg inline-flex items-center justify-center border gap-1 w-14 h-7 rounded-full font-bold bg-opacity-70 drop-shadow-sm shadow-lg",
                getScoreColor(score, "user"),
            )}
        >
            <BiStar /> {(score === 0) ? "-" : score / 10}
        </div>
    )
}
