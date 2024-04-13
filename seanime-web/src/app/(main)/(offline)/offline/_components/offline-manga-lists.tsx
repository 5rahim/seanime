import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { OfflineMangaEntry } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { OfflineMediaListAtom } from "@/components/shared/custom-ui/offline-media-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import React from "react"

export function OfflineMangaLists() {
    const { snapshot, mangaLists: lists } = useOfflineSnapshot()

    const Grid = React.useCallback(({ entries }: { entries: OfflineMangaEntry[] }) => {
        return (
            <div
                className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-7 min-[2000px]:grid-cols-8 gap-4"
            >
                {entries?.map(entry => {
                    if (!entry) return null

                    return <OfflineMediaListAtom
                        key={entry.mediaId}
                        media={entry.media!}
                        listData={entry.listData}
                        withAudienceScore={false}
                        assetMap={snapshot?.assetMap}
                        isManga
                    />
                })}
            </div>
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
