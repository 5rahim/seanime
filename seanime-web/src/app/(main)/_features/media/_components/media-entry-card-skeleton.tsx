import { Skeleton } from "@/components/ui/skeleton"
import React from "react"

export const MediaEntryCardSkeleton = () => {
    return (
        <>
            <Skeleton
                data-media-entry-card-skeleton
                className="min-w-[250px] basis-[250px] max-w-[250px] h-[350px] bg-gray-900 rounded-[--radius-md] mt-8 mx-2"
            />
        </>
    )
}
