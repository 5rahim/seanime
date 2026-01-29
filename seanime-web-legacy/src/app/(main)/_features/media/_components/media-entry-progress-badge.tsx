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
            data-media-entry-progress-badge
        >
            {top && <span data-media-entry-progress-badge-top className="block">
                {top}
            </span>}
            {(!!progress || forceShowProgress) && <span
                data-media-entry-progress-badge-progress
                className="block"
                data-progress={progress}
                data-progress-total={progressTotal}
                data-force-show-total={forceShowTotal}
            >
                {progress || 0}{(!!progressTotal || forceShowTotal) && <span
                data-media-entry-progress-badge-progress-total
                className={cn(
                    "text-[--muted]",
                )}
            >/{(!!progressTotal) ? progressTotal : "-"}</span>}
            </span>}
        </Badge>
    )
}
