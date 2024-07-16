import { Offline_MangaEntry } from "@/api/generated/types"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { MediaCardGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { OfflineAnimeEntryCard } from "@/app/(main)/_features/media/_components/offline-media-entry-card"
import { PageWrapper } from "@/components/shared/page-wrapper"
import React from "react"

export function OfflineMangaLists() {
    const { snapshot, mangaLists: lists } = useOfflineSnapshot()

    const Grid = React.useCallback(({ entries }: { entries: Offline_MangaEntry[] }) => {
        return (
            <MediaCardGrid>
                {entries?.map(entry => {
                    if (!entry) return null

                    return <OfflineAnimeEntryCard
                        key={entry.mediaId}
                        media={entry.media!}
                        listData={entry.listData}
                        withAudienceScore={false}
                        assetMap={snapshot?.assetMap}
                        type="manga"
                    />
                })}
            </MediaCardGrid>
        )
    }, [])

    return (
        <>
            <PageWrapper
                className="pt-4 relative space-y-8"
            >
                <div className="space-y-6">
                    <div className="space-y-6">
                        {!!lists.current?.length && (
                            <>
                                <h2>Currently reading</h2>
                                <Grid entries={lists.current} />
                            </>
                        )}
                        {!!lists.paused?.length && (
                            <>
                                <h2>Paused</h2>
                                <Grid entries={lists.paused} />
                            </>
                        )}
                        {!!lists.planning?.length && (
                            <>
                                <h2>Planned</h2>
                                <Grid entries={lists.planning} />
                            </>
                        )}
                        {!!lists.completed?.length && (
                            <>
                                <h2>Completed</h2>
                                <Grid entries={lists.completed} />
                            </>
                        )}
                        {!!lists.dropped?.length && (
                            <>
                                <h2>Dropped</h2>
                                <Grid entries={lists.dropped} />
                            </>
                        )}
                    </div>
                </div>
            </PageWrapper>
        </>
    )
}
