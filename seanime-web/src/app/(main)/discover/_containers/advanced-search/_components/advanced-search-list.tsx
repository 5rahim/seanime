import { useAnilistAdvancedSearch } from "@/app/(main)/discover/_containers/advanced-search/_lib/queries"
import { AnimeListItem } from "@/components/shared/anime-list-item"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"
import { AiOutlinePlusCircle } from "react-icons/ai"

export function AdvancedSearchList() {

    const { isLoading, data, fetchNextPage, hasNextPage } = useAnilistAdvancedSearch()

    return <>
        {!isLoading && <div
            className="px-4 grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6 min-[2000px]:grid-cols-8 gap-4"
        >
            {data?.pages.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean).filter(media => !!media.startDate?.year).map(media => (
                <AnimeListItem
                    key={`${media.id}`}
                    media={media}
                    showLibraryBadge={true}
                />
            ))}
            {((data?.pages.filter(Boolean).flatMap(n => n.Page?.media).filter(Boolean) || []).length > 0 && hasNextPage) &&
                <div
                    className={cn(
                        "h-full col-span-1 group/anime-list-item relative flex flex-col place-content-stretch rounded-md animate-none min-h-[348px]",
                        "cursor-pointer border border-dashed  border-none text-[--muted] hover:text-white pt-24 items-center gap-2 transition",
                    )}
                    onClick={() => fetchNextPage()}
                >
                    <AiOutlinePlusCircle className="text-4xl" />
                    <p className="text-lg font-medium">Load more</p>
                </div>}
        </div>}
        {isLoading && <LoadingSpinner/>}
    </>
}
