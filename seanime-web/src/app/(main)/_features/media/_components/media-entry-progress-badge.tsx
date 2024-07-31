import { Badge } from "@/components/ui/badge"
import { cn } from "@/components/ui/core/styling"
import React from "react"

type MediaEntryProgressBadgeProps = {
    progress?: number
    progressTotal?: number
}

export const MediaEntryProgressBadge = (props: MediaEntryProgressBadgeProps) => {
    const { progress, progressTotal } = props

    if (!progress) return null

    return (
        <Badge
            intent="unstyled"
            size="lg"
            className="font-medium tracking-wide rounded-md rounded-tl-none rounded-br-none border-0 bg-zinc-950/40 px-1.5 gap-0"
        >
            {progress}{!!progressTotal && <span
            className={cn(
                "text-[--muted]",
            )}
        >/{progressTotal}</span>}
        </Badge>
    )
}
