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
        <Badge size="lg" className="rounded-md px-1.5 gap-0">
            {progress}{!!progressTotal && <span
            className={cn(
                progress != progressTotal && "text-[--muted]",
            )}
        >/{progressTotal}</span>}
        </Badge>
    )
}
