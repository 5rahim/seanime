import { ContinueWatching } from "@/app/(main)/(library)/_containers/continue-watching"
import { LibraryHeader } from "@/app/(main)/(library)/_containers/library-header"
import { useOfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot-context"
import { OfflineAnimeEntry } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { OfflineMediaListAtom } from "@/components/shared/custom-ui/offline-media-list-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { BasicMediaFragment } from "@/lib/anilist/gql/graphql"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { groupBy } from "lodash"
import React from "react"

export function OfflineAnimeLists() {
    const snapshot = useOfflineSnapshot()
    const { lists, continueWatchingEpisodeList } = useOfflineAnimeLists()
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

export function useOfflineAnimeLists() {
    const snapshot = useOfflineSnapshot()

    const lists = React.useMemo(() => {
        if (!snapshot) return {}

        const grouped = groupBy(snapshot.entries?.animeEntries?.filter(Boolean), n => n?.listData?.status)

        return {
            current: grouped?.CURRENT,
            planning: grouped?.PLANNING,
            completed: grouped?.COMPLETED,
            paused: grouped?.PAUSED,
            dropped: grouped?.DROPPED,
        }
    }, [snapshot?.entries?.animeEntries])

    const continueWatchingEpisodeList = React.useMemo(() => {
        if (!snapshot) return []

        const entries = snapshot.entries?.animeEntries?.filter(Boolean)?.filter(n => n?.listData?.status === "CURRENT")

        return entries?.flatMap(entry => {
            let ep = entry?.episodes?.filter(Boolean)?.find(n => n?.progressNumber == (entry?.listData?.progress || 0) + 1)
            if (!ep) return null
            return {
                ...ep,
                episodeMetadata: {
                    ...ep.episodeMetadata,
                    image: offline_getAssetUrl(ep.episodeMetadata?.image, snapshot.assetMap),
                },
                basicMedia: {
                    ...entry.media,
                    bannerImage: offline_getAssetUrl(entry.media?.bannerImage, snapshot.assetMap),
                    coverImage: {
                        ...entry.media?.coverImage,
                        extraLarge: offline_getAssetUrl(entry.media?.coverImage?.extraLarge, snapshot.assetMap),
                        large: offline_getAssetUrl(entry.media?.coverImage?.large, snapshot.assetMap),
                        medium: offline_getAssetUrl(entry.media?.coverImage?.medium, snapshot.assetMap),
                    },
                } as BasicMediaFragment,
            }
        })?.filter(Boolean) || []
    }, [snapshot?.entries?.animeEntries])


    return {
        lists,
        continueWatchingEpisodeList,
    }
}
