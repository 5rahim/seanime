import { Badge } from "@/components/ui/badge"
import React from "react"

type MediaEntryProgressBadgeProps = {
    progress?: number
    progressTotal?: number
}

export const MediaEntryProgressBadge = (props: MediaEntryProgressBadgeProps) => {
    const { progress, progressTotal } = props

    if (!progress) return null

    return (
        <Badge size="lg" className="rounded-md px-1.5">
            {progress}{!!progressTotal ? `/${progressTotal}` : ""}
        </Badge>
    )
}
