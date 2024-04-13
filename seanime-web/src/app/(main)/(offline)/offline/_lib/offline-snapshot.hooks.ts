import { OfflineSnapshot } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.types"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { BasicMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import { groupBy } from "lodash"
import React from "react"

export function useGetOfflineSnapshot() {
    const qc = useQueryClient()

    const { data: snapshot, isLoading } = useSeaQuery<OfflineSnapshot>({
        endpoint: SeaEndpoints.OFFLINE_SNAPSHOT,
        queryKey: ["get-offline-snapshot"],
    })

    const animeLists = React.useMemo(() => {
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

    const mangaLists = React.useMemo(() => {
        if (!snapshot) return {}

        const grouped = groupBy(snapshot.entries?.mangaEntries?.filter(Boolean), n => n?.listData?.status)

        return {
            current: grouped?.CURRENT,
            planning: grouped?.PLANNING,
            completed: grouped?.COMPLETED,
            paused: grouped?.PAUSED,
            dropped: grouped?.DROPPED,
        }
    }, [snapshot?.entries?.mangaEntries])

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
        snapshot,
        animeLists,
        mangaLists,
        continueWatchingEpisodeList,
        isLoading,
    }
}
