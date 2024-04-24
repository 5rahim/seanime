import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { MediaEntryCard } from "@/app/(main)/_components/features/media/media-entry-card"
import React from "react"

type AnilistEntryListProps = {
    list: AL_AnimeCollection_MediaListCollection_Lists | undefined
}

export function AnilistEntryList(props: AnilistEntryListProps) {

    const {
        list,
        ...rest
    } = props

    return (
        <div
            className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
        >
            {list?.entries?.filter(Boolean)?.map((entry) => (
                <MediaEntryCard
                    key={`${entry.media?.id}`}
                    listData={{
                        progress: entry.progress!,
                        score: entry.score!,
                        status: entry.status!,
                        startedAt: entry.startedAt?.year ? new Date(entry.startedAt.year,
                            (entry.startedAt.month || 1) - 1,
                            entry.startedAt.day || 1).toISOString() : undefined,
                        completedAt: entry.completedAt?.year ? new Date(entry.completedAt.year,
                            (entry.completedAt.month || 1) - 1,
                            entry.completedAt.day || 1).toISOString() : undefined,
                    }}
                    showLibraryBadge
                    media={entry.media!}
                    showListDataButton
                />
            ))}
        </div>
    )
}
