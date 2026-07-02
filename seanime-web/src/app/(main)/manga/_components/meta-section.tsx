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
import { PluginMangaPageButtons, PluginMangaPageDropdownItems } from "@/app/(main)/_features/plugin/actions/plugin-actions"
import { PluginWebviewSlot } from "@/app/(main)/_features/plugin/webview/plugin-webviews"
import { SeaLink } from "@/components/shared/sea-link"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Skeleton } from "@/components/ui/skeleton"
import { Tooltip } from "@/components/ui/tooltip"
import { copyToClipboard, openTab } from "@/lib/helpers/browser"
import { getCustomSourceExtensionId, getCustomSourceMediaSiteUrl, isCustomSource } from "@/lib/server/utils"
import { ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/theme-hooks"
import React from "react"
import { BiDotsVerticalRounded, BiExtension } from "react-icons/bi"
import { LuExternalLink } from "react-icons/lu"
import { SiAnilist } from "react-icons/si"


export function MetaSection(props: { entry: Manga_Entry | undefined, details: AL_MangaDetailsById_Media | undefined, detailsLoading?: boolean }) {

    const { entry, details, detailsLoading } = props
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
                <MediaEntryAudienceScore meanScore={entry.media?.meanScore} badgeClass="bg-transparent border-transparent px-0" />

                {(detailsLoading && !details) ? <Skeleton className="h-6 w-52 rounded-full opacity-60" /> :
                    <MediaEntryGenresList genres={details?.genres} type="manga" />}
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
                    after={<div className="w-full flex flex-wrap gap-4 items-center" data-manga-meta-section-buttons-container>

                        {isCustomSource(entry.mediaId) && (
                            <Tooltip
                                trigger={<div>
                                    <SeaLink href={`/custom-sources?provider=${getCustomSourceExtensionId(entry.media)}`}>
                                        <IconButton size="sm" intent="gray-link" className="px-0" icon={<BiExtension className="text-lg" />} />
                                    </SeaLink>
                                </div>}
                            >
                                Custom source
                            </Tooltip>
                        )}

                        {!isCustomSource(entry.mediaId) && <SeaLink href={`https://anilist.co/manga/${entry.mediaId}`} target="_blank">
                            <IconButton size="sm" intent="gray-link" className="px-0" icon={<SiAnilist className="text-lg" />} />
                        </SeaLink>}

                        {isCustomSource(entry.mediaId) && !!getCustomSourceMediaSiteUrl(entry.media) && <Tooltip
                            trigger={<div>
                                <SeaLink href={getCustomSourceMediaSiteUrl(entry.media)!} target="_blank">
                                    <IconButton size="sm" intent="gray-link" className="px-0" icon={<LuExternalLink className="text-lg" />} />
                                </SeaLink>
                            </div>}
                        >
                            Open in website
                        </Tooltip>}
                    </div>}
                >
                    {ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && <Details />}
                </MediaPageHeaderEntryDetails>

                <PluginWebviewSlot slot="after-media-entry-details" />

                {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <Details />}


                <div className="w-full flex flex-wrap gap-2 items-center" data-manga-meta-section-buttons-row>

                    {ts.mediaPageBannerInfoBoxSize !== ThemeMediaPageInfoBoxSize.Fluid && <div className="flex-1 hidden lg:flex"></div>}

                    <DropdownMenu
                        data-manga-entry-dropdown-menu
                        trigger={<IconButton
                            data-manga-entry-dropdown-menu-trigger
                            icon={<BiDotsVerticalRounded />}
                            intent="gray-subtle"
                            size="md"
                        />}
                    >
                        {!isCustomSource(entry.mediaId) && <DropdownMenuItem
                            onClick={() => openTab(`https://anilist.co/manga/${entry.mediaId}`)}
                        >
                            <SiAnilist /> Open on AniList
                        </DropdownMenuItem>}
                        {isCustomSource(entry.mediaId) && !!getCustomSourceMediaSiteUrl(entry.media) && <DropdownMenuItem
                            onClick={() => openTab(getCustomSourceMediaSiteUrl(entry.media)!)}
                        >
                            <LuExternalLink /> Open in website
                        </DropdownMenuItem>}
                        {isCustomSource(entry.mediaId) && <DropdownMenuItem
                            onClick={() => copyToClipboard(entry.mediaId.toString())}
                        >
                            Copy ID
                        </DropdownMenuItem>}
                        <PluginMangaPageDropdownItems media={entry.media} />
                    </DropdownMenu>

                    <MediaSyncTrackButton mediaId={entry.mediaId} type="manga" size="md" />

                    <PluginMangaPageButtons media={entry.media} />
                </div>

            </MediaPageHeaderDetailsContainer>
        </MediaPageHeader>

    )

}
