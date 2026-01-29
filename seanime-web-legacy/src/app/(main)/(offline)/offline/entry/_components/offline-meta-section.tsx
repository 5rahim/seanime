"use client"
import { AL_BaseAnime, AL_BaseManga, Anime_Entry, Manga_Entry } from "@/api/generated/types"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import {
    MediaPageHeader,
    MediaPageHeaderDetailsContainer,
    MediaPageHeaderEntryDetails,
} from "@/app/(main)/_features/media/_components/media-page-header-components"
import React from "react"

type OfflineMetaSectionProps<T extends "anime" | "manga"> = {
    type: T,
    entry: T extends "anime" ? Anime_Entry : Manga_Entry
}

export function OfflineMetaSection<T extends "anime" | "manga">(props: OfflineMetaSectionProps<T>) {

    const { type, entry } = props

    if (!entry?.media) return null

    return (
        <MediaPageHeader
            backgroundImage={entry.media?.bannerImage}
            coverImage={entry.media?.coverImage?.extraLarge}
        >

            <MediaPageHeaderDetailsContainer>

                <MediaPageHeaderEntryDetails
                    coverImage={entry.media?.coverImage?.extraLarge || entry.media?.coverImage?.large}
                    title={entry.media?.title?.userPreferred}
                    color={entry.media?.coverImage?.color}
                    englishTitle={entry.media?.title?.english}
                    romajiTitle={entry.media?.title?.romaji}
                    startDate={entry.media?.startDate}
                    season={entry.media?.season}
                    progressTotal={type === "anime" ? (entry.media as AL_BaseAnime)?.episodes : (entry.media as AL_BaseManga)?.chapters}
                    status={entry.media?.status}
                    description={entry.media?.description}
                    listData={entry.listData}
                    media={entry.media}
                    type="anime"
                />


                <div className="flex gap-2 items-center">
                    <MediaEntryAudienceScore meanScore={entry.media?.meanScore} badgeClass="bg-transparent" />
                </div>
            </MediaPageHeaderDetailsContainer>
        </MediaPageHeader>
    )

}
