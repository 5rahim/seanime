import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import React from "react"

type MediaEntryProgressBadgeProps = {
    progress?: number
    progressTotal?: number
    forceShowTotal?: boolean
    forceShowProgress?: boolean
    top?: React.ReactNode
}

export const MediaEntryProgressBadge = (props: MediaEntryProgressBadgeProps) => {
    const { progress, progressTotal, forceShowTotal, forceShowProgress, top } = props

    // if (!progress) return null

    return (
        <Badge
            intent="unstyled"
            size="lg"
            className="font-semibold tracking-wide flex-col rounded-[--radius-md] rounded-tl-none rounded-br-none border-0 bg-zinc-950/40 px-1.5 py-0.5 gap-0 !h-auto"
        >
            {top && <span className="block">
                {top}
            </span>}
            {(!!progress || forceShowProgress) && <span className="block">
                {progress || 0}{(!!progressTotal || forceShowTotal) && <span
                className={cn(
                    "text-[--muted]",
                )}
            >/{(!!progressTotal) ? progressTotal : "-"}</span>}
            </span>}
        </Badge>
    )
}
