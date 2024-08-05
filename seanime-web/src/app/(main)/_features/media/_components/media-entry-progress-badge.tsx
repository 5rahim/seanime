import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import React from "react"

type MediaEntryProgressBadgeProps = {
    progress?: number
    progressTotal?: number
    forceShowTotal?: boolean
}

export const MediaEntryProgressBadge = (props: MediaEntryProgressBadgeProps) => {
    const { progress, progressTotal, forceShowTotal } = props

    if (!progress) return null

    return (
        <Badge
            intent="unstyled"
            size="lg"
            className="font-semibold tracking-wide rounded-md rounded-tl-none rounded-br-none border-0 bg-zinc-950/40 px-1.5 gap-0"
        >
            {progress}{(!!progressTotal || forceShowTotal) && <span
            className={cn(
                "text-[--muted]",
            )}
        >/{(!!progressTotal) ? progressTotal : "-"}</span>}
        </Badge>
    )
}
