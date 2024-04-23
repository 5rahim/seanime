import { LibraryHeader } from "@/app/(main)/(library)/_components/library-header"
import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { OfflineAnimeEntry } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { OfflineMediaListAtom } from "@/components/shared/custom-ui/offline-media-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function OfflineAnimeLists() {
    const { snapshot, animeLists: lists, continueWatchingEpisodeList } = useOfflineSnapshot()
    const ts = useThemeSettings()

    const Grid = React.useCallback(({ entries }: { entries: OfflineAnimeEntry[] }) => {
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
                    />
                })}
            </div>
        )
    }, [])

    return (
        <>
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <LibraryHeader list={continueWatchingEpisodeList} />}
            <PageWrapper
                className="pt-4 relative space-y-8"
            >
                <div className="space-y-6">
                    <ContinueWatching
                        list={continueWatchingEpisodeList}
                        isLoading={false}
                        linkTemplate={"/offline/anime?id={id}&playNext=true"}
                    />
                    <div className="p-4 space-y-6">
                        {!!lists.current?.length && (
                            <>
                                <h2>Currently watching</h2>
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
