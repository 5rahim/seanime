import { AL_BaseAnime, AL_BaseManga, Offline_AssetMapImageMap, Offline_ListData } from "@/api/generated/types"
import { OfflineAnilistAnimeEntryModal } from "@/app/(main)/(offline)/offline/_containers/offline-anilist-media-entry-modal"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import {
    AnimeEntryCardBody,
    AnimeEntryCardContainer,
    AnimeEntryCardHoverPopup,
    AnimeEntryCardHoverPopupBanner,
    AnimeEntryCardHoverPopupBody,
    AnimeEntryCardHoverPopupFooter,
    AnimeEntryCardHoverPopupTitleSection,
    AnimeEntryCardNextAiring,
    AnimeEntryCardOverlay,
    AnimeEntryCardTitleSection,
} from "@/app/(main)/_features/media/_components/media-entry-card-components"
import { AnimeEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { AnimeEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { AnimeEntryScoreBadge } from "@/app/(main)/_features/media/_components/media-entry-score-badge"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button } from "@/components/ui/button"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import React from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"

type OfflineAnimeEntryCardBasedProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    listData: Offline_ListData | undefined
    assetMap: Offline_AssetMapImageMap | undefined
}

type OfflineAnimeEntryCardProps<T extends "anime" | "manga"> = {
    type: T
    media: T extends "anime" ? AL_BaseAnime : T extends "manga" ? AL_BaseManga : never
} & OfflineAnimeEntryCardBasedProps

export function OfflineAnimeEntryCard<T extends "anime" | "manga">(props: OfflineAnimeEntryCardProps<T>) {

    const serverStatus = useServerStatus()
    const {
        media,
        listData,
        overlay,
        type,
        withAudienceScore = true,
        assetMap,
    } = props

    const showProgressBar = React.useMemo(() => {
        return !!listData?.progress
        && type === "anime" ? !!(media as AL_BaseAnime)?.episodes : !!(media as AL_BaseManga)?.chapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, media, listData?.status])

    const link = type === "anime" ? `/offline/anime?id=${media.id}` : `/offline/manga?id=${media.id}`

    const progressTotal = type === "anime" ? (media as AL_BaseAnime)?.episodes : (media as AL_BaseManga)?.chapters

    if (!media) return null

    return (
        <AnimeEntryCardContainer className={props.containerClassName}>

            <AnimeEntryCardOverlay overlay={overlay} />

            {/*ACTION POPUP*/}
            <AnimeEntryCardHoverPopup>

                {/*METADATA SECTION*/}
                <AnimeEntryCardHoverPopupBody>

                    <AnimeEntryCardHoverPopupBanner
                        trailerId={(media as any)?.trailer?.id}
                        showProgressBar={showProgressBar}
                        mediaId={media.id}
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                        showTrailer={false}
                        disableAnimeCardTrailers={serverStatus?.settings?.library?.disableAnimeCardTrailers}
                        bannerImage={offline_getAssetUrl(media.bannerImage, assetMap) || offline_getAssetUrl(media.coverImage?.large, assetMap) || ""}
                        isAdult={media.isAdult}
                        blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
                        link={link}
                        listStatus={listData?.status}
                    />

                    <AnimeEntryCardHoverPopupTitleSection
                        title={media.title?.userPreferred || ""}
                        year={media.startDate?.year}
                        season={media.season}
                        format={media.format}
                        link={link}
                    />

                    {type === "anime" && (
                        <AnimeEntryCardNextAiring nextAiring={(media as AL_BaseAnime).nextAiringEpisode} />
                    )}

                    {type === "anime" && <div className="py-1">
                        <Link
                            href={`/offline/anime?id=${media.id}${(!!listData?.progress && (listData?.status !== "COMPLETED"))
                                ? "&playNext=true"
                                : ""}`}
                        >
                            <Button
                                leftIcon={<BiPlay className="text-2xl" />}
                                intent="white"
                                size="md"
                                className="w-full text-md"
                            >
                                {!!listData?.progress && (listData?.status === "CURRENT" || listData?.status === "PAUSED")
                                    ? "Continue watching"
                                    : "Watch"}
                            </Button>
                        </Link>
                    </div>}

                    {type === "manga" && <Link
                        href={`/offline/manga?id=${props.media.id}`}
                    >
                        <Button
                            leftIcon={<IoLibrarySharp />}
                            intent="white"
                            size="md"
                            className="w-full text-md mt-2"
                        >
                            Read
                        </Button>
                    </Link>}

                    {(listData?.status) &&
                        <p className="text-center">
                            {listData?.status === "CURRENT" ? type === "anime" ? "Watching" : "Reading"
                                : capitalize(listData?.status ?? "")}
                        </p>}

                </AnimeEntryCardHoverPopupBody>

                <AnimeEntryCardHoverPopupFooter>

                    <OfflineAnilistAnimeEntryModal
                        listData={listData}
                        assetMap={assetMap}
                        media={media}
                        type={type}
                    />

                    {withAudienceScore &&
                        <AnimeEntryAudienceScore
                            meanScore={media.meanScore}
                        />}

                </AnimeEntryCardHoverPopupFooter>

            </AnimeEntryCardHoverPopup>

            <AnimeEntryCardBody
                link={link}
                type={type}
                title={media.title?.userPreferred || ""}
                season={media.season}
                listStatus={listData?.status}
                status={media.status}
                showProgressBar={showProgressBar}
                progress={listData?.progress}
                progressTotal={progressTotal}
                startDate={media.startDate}
                bannerImage={offline_getAssetUrl(media.coverImage?.extraLarge, assetMap) || ""}
                isAdult={media.isAdult}
                showLibraryBadge={false}
                blurAdultContent={serverStatus?.settings?.anilist?.blurAdultContent}
            >
                <div className="absolute z-[10] right-1 bottom-1">
                    <AnimeEntryScoreBadge
                        score={listData?.score}
                    />
                </div>
                <div className="absolute z-[10] left-1 bottom-1">
                    <AnimeEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                    />
                </div>
            </AnimeEntryCardBody>

            <AnimeEntryCardTitleSection
                title={media.title?.userPreferred || ""}
                year={media.startDate?.year}
                season={media.season}
                format={media.format}
            />

        </AnimeEntryCardContainer>
    )
}
