import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { MediaCardGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import React from "react"

type AnilistMediaEntryListProps = {
    list: AL_AnimeCollection_MediaListCollection_Lists | undefined
}

/**
 * Displays a list of media entry card from an Anilist media list collection.
 */
export function AnilistMediaEntryList(props: AnilistMediaEntryListProps) {

    const {
        list,
        ...rest
    } = props

    return (
        <MediaCardGrid>
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
                    type="anime"
                />
            ))}
        </MediaCardGrid>
    )
}
