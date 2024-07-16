import { Skeleton } from "@/components/ui/skeleton"
import React from "react"

export const AnimeEntryCardSkeleton = () => {
    return (
        <>
            <Skeleton
                className="min-w-[250px] basis-[250px] max-w-[250px] h-[350px] bg-gray-900 rounded-md mt-8 mx-2"
            />
        </>
    )
}
