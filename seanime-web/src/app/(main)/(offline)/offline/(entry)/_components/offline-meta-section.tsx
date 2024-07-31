"use client"
import { AL_BaseAnime, AL_BaseManga, Offline_AnimeEntry, Offline_AssetMapImageMap, Offline_MangaEntry } from "@/api/generated/types"
import { OfflineAnilistMediaEntryModal } from "@/app/(main)/(offline)/offline/_containers/offline-anilist-media-entry-modal"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import {
    MediaPageHeader,
    MediaPageHeaderDetailsContainer,
    MediaPageHeaderEntryDetails,
} from "@/app/(main)/_features/media/_components/media-page-header-components"
import React from "react"

type OfflineMetaSectionProps<T extends "anime" | "manga"> = {
    type: T,
    entry: T extends "anime" ? Offline_AnimeEntry : Offline_MangaEntry
    assetMap: Offline_AssetMapImageMap | undefined
}

export function OfflineMetaSection<T extends "anime" | "manga">(props: OfflineMetaSectionProps<T>) {

    const { type, entry, assetMap } = props

    if (!entry?.media) return null

    return (
        <MediaPageHeader
            size="smaller"
            backgroundImage={offline_getAssetUrl(entry.media?.bannerImage, assetMap)
                || offline_getAssetUrl(entry.media.coverImage?.extraLarge, assetMap)}
        >

            <MediaPageHeaderDetailsContainer>

                <MediaPageHeaderEntryDetails
                    coverImage={offline_getAssetUrl(entry.media.coverImage?.extraLarge, assetMap)
                        || offline_getAssetUrl(entry.media.coverImage?.extraLarge, assetMap)}
                    color={entry.media?.coverImage?.color}
                    title={entry.media?.title?.userPreferred}
                    englishTitle={entry.media?.title?.english}
                    romajiTitle={entry.media?.title?.romaji}
                    startDate={entry.media?.startDate}
                    season={entry.media?.season}
                    progressTotal={type === "anime" ? (entry.media as AL_BaseAnime)?.episodes : (entry.media as AL_BaseManga)?.chapters}
                    status={entry.media?.status}
                    description={entry.media?.description}
                    listData={entry.listData}
                    media={entry.media}
                    type={type}
                    offlineAnilistAnimeEntryModal={<OfflineAnilistMediaEntryModal
                        media={entry.media}
                        assetMap={assetMap}
                        type={type}
                        listData={entry.listData}
                    />}
                />


                <div className="flex gap-2 items-center">
                    <MediaEntryAudienceScore meanScore={entry.media?.meanScore} />
                </div>
            </MediaPageHeaderDetailsContainer>
        </MediaPageHeader>
    )

}
