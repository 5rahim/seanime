import { AL_AnimeCollection_MediaListCollection_Lists } from "@/api/generated/types"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import React from "react"


type AnilistAnimeEntryListProps = {
    list: AL_AnimeCollection_MediaListCollection_Lists | undefined
    type: "anime" | "manga"
}

/**
 * Displays a list of media entry card from an Anilist media list collection.
 */
export function AnilistAnimeEntryList(props: AnilistAnimeEntryListProps) {

    const {
        list,
        type,
        ...rest
    } = props

    return (
        <MediaCardLazyGrid itemCount={list?.entries?.filter(Boolean)?.length || 0} data-anilist-anime-entry-list>
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
                    type={type}
                />
            ))}
        </MediaCardLazyGrid>
    )
}
