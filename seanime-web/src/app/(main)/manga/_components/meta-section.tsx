"use client"
import { AL_MangaDetailsById_Media, Manga_Entry } from "@/api/generated/types"
import {
    AnimeEntryRankings,
    MediaEntryAudienceScore,
    MediaEntryGenresList,
} from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import {
    MediaPageHeader,
    MediaPageHeaderDetailsContainer,
    MediaPageHeaderEntryDetails,
} from "@/app/(main)/_features/media/_components/media-page-header-components"
import { MediaSyncTrackButton } from "@/app/(main)/_features/media/_containers/media-sync-track-button"
import { SeaLink } from "@/components/shared/sea-link"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { SiAnilist } from "react-icons/si"
import { PluginMangaPageButtons } from "../../_features/plugin/actions/plugin-actions"


export function MetaSection(props: { entry: Manga_Entry | undefined, details: AL_MangaDetailsById_Media | undefined }) {

    const { entry, details } = props
    const ts = useThemeSettings()

    if (!entry?.media) return null

    const Details = () => (
        <>
            <div
                className={cn(
                    "flex gap-2 flex-wrap items-center",
                    ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "justify-center lg:justify-start lg:max-w-[65vw]",
                )}
            >
                <MediaEntryAudienceScore meanScore={entry.media?.meanScore} badgeClass="bg-transparent" />

                <MediaEntryGenresList genres={details?.genres} type="manga" />
            </div>

            <AnimeEntryRankings rankings={details?.rankings} />
        </>
    )

    return (
        <MediaPageHeader
            backgroundImage={entry.media?.bannerImage}
            coverImage={entry.media?.coverImage?.extraLarge}
        >

            <MediaPageHeaderDetailsContainer>

                <MediaPageHeaderEntryDetails
                    coverImage={entry.media?.coverImage?.extraLarge || entry.media?.coverImage?.large}
                    title={entry.media?.title?.userPreferred}
                    englishTitle={entry.media?.title?.english}
                    romajiTitle={entry.media?.title?.romaji}
                    startDate={entry.media?.startDate}
                    season={entry.media?.season}
                    color={entry.media?.coverImage?.color}
                    progressTotal={entry.media?.chapters}
                    status={entry.media?.status}
                    description={entry.media?.description}
                    listData={entry.listData}
                    media={entry.media}
                    type="manga"
                >
                    {ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && <Details />}
                </MediaPageHeaderEntryDetails>

                {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <Details />}


                <div className="w-full flex flex-wrap gap-4 items-center" data-manga-meta-section-buttons-container>

                    <SeaLink href={`https://anilist.co/manga/${entry.mediaId}`} target="_blank">
                        <IconButton intent="gray-link" className="px-0" icon={<SiAnilist className="text-lg" />} />
                    </SeaLink>

                    {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <div className="flex-1 hidden lg:flex"></div>}

                    <MediaSyncTrackButton mediaId={entry.mediaId} type="manga" size="md" />

                    <PluginMangaPageButtons media={entry.media} />
                </div>

            </MediaPageHeaderDetailsContainer>
        </MediaPageHeader>

    )

}
