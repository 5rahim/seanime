import { cn } from "@/components/ui/core/styling"
import React from "react"

type EpisodeListGridProps = {
    children?: React.ReactNode
}

export function EpisodeListGrid(props: EpisodeListGridProps) {

    const {
        children,
        ...rest
    } = props


    return (
        <div
            className={cn(
                "grid grid-cols-1 lg:grid-cols-2 2xl:grid-cols-3 min-[2000px]:grid-cols-4",
                "gap-4",
            )}
            {...rest}
            data-episode-list-grid
        >
            {children}
        </div>
    )
}
