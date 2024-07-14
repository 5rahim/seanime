import { AL_BaseAnime, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { useGetOfflineSnapshot } from "@/api/hooks/offline.hooks"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { groupBy } from "lodash"
import React from "react"

export function useHandleOfflineSnapshot() {

    const { data: snapshot, isLoading } = useGetOfflineSnapshot()

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
                baseAnime: {
                    ...entry.media,
                    bannerImage: offline_getAssetUrl(entry.media?.bannerImage, snapshot.assetMap),
                    coverImage: {
                        ...entry.media?.coverImage,
                        extraLarge: offline_getAssetUrl(entry.media?.coverImage?.extraLarge, snapshot.assetMap),
                        large: offline_getAssetUrl(entry.media?.coverImage?.large, snapshot.assetMap),
                        medium: offline_getAssetUrl(entry.media?.coverImage?.medium, snapshot.assetMap),
                    },
                } as AL_BaseAnime,
            }
        })?.filter(Boolean) || [] as Anime_MediaEntryEpisode[]
    }, [snapshot?.entries?.animeEntries])

    return {
        snapshot,
        animeLists,
        mangaLists,
        continueWatchingEpisodeList,
        isLoading,
    }
}
