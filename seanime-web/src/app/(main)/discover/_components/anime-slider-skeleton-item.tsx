import { Skeleton } from "@/components/ui/skeleton"
import React from "react"

export const AnimeSliderSkeletonItem = () => {
    return (
        <>
            <Skeleton
                className="min-w-[250px] max-w-[250px] h-[350px] bg-gray-700 rounded-md mt-8"
            />
        </>
    )
}
