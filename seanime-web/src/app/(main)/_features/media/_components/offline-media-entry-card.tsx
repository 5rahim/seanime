import { AL_BaseManga, AL_BaseMedia, Offline_AssetMapImageMap, Offline_ListData } from "@/api/generated/types"
import { OfflineAnilistMediaEntryModal } from "@/app/(main)/(offline)/offline/_components/offline-anilist-media-entry-modal"
import { offline_getAssetUrl } from "@/app/(main)/(offline)/offline/_lib/offline-snapshot.utils"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import {
    MediaEntryCardBody,
    MediaEntryCardContainer,
    MediaEntryCardHoverPopup,
    MediaEntryCardHoverPopupBanner,
    MediaEntryCardHoverPopupBody,
    MediaEntryCardHoverPopupFooter,
    MediaEntryCardHoverPopupTitleSection,
    MediaEntryCardNextAiring,
    MediaEntryCardOverlay,
    MediaEntryCardTitleSection,
} from "@/app/(main)/_features/media/_components/media-entry-card-components"
import { MediaEntryAudienceScore } from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { MediaEntryProgressBadge } from "@/app/(main)/_features/media/_components/media-entry-progress-badge"
import { MediaEntryScoreBadge } from "@/app/(main)/_features/media/_components/media-entry-score-badge"
import { Button } from "@/components/ui/button"
import { useAtomValue } from "jotai/react"
import capitalize from "lodash/capitalize"
import Link from "next/link"
import React from "react"
import { BiPlay } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"

type OfflineMediaEntryCardBasedProps = {
    overlay?: React.ReactNode
    withAudienceScore?: boolean
    containerClassName?: string
    listData: Offline_ListData | undefined
    assetMap: Offline_AssetMapImageMap | undefined
}

type OfflineMediaEntryCardProps<T extends "anime" | "manga"> = {
    type: T
    media: T extends "anime" ? AL_BaseMedia : T extends "manga" ? AL_BaseManga : never
} & OfflineMediaEntryCardBasedProps

export function OfflineMediaEntryCard<T extends "anime" | "manga">(props: OfflineMediaEntryCardProps<T>) {

    const serverStatus = useAtomValue(serverStatusAtom)
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
        && type === "anime" ? !!(media as AL_BaseMedia)?.episodes : !!(media as AL_BaseManga)?.chapters
            && listData?.status !== "COMPLETED"
    }, [listData?.progress, media, listData?.status])

    const link = type === "anime" ? `/offline/anime?id=${media.id}` : `/offline/manga?id=${media.id}`

    const progressTotal = type === "anime" ? (media as AL_BaseMedia)?.episodes : (media as AL_BaseManga)?.chapters

    if (!media) return null

    return (
        <MediaEntryCardContainer className={props.containerClassName}>

            <MediaEntryCardOverlay overlay={overlay} />

            {/*ACTION POPUP*/}
            <MediaEntryCardHoverPopup>

                {/*METADATA SECTION*/}
                <MediaEntryCardHoverPopupBody>

                    <MediaEntryCardHoverPopupBanner
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

                    <MediaEntryCardHoverPopupTitleSection
                        title={media.title?.userPreferred || ""}
                        year={media.startDate?.year}
                        season={media.season}
                        format={media.format}
                        link={link}
                    />

                    {type === "anime" && (
                        <MediaEntryCardNextAiring nextAiring={(media as AL_BaseMedia).nextAiringEpisode} />
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

                </MediaEntryCardHoverPopupBody>

                <MediaEntryCardHoverPopupFooter>

                    <OfflineAnilistMediaEntryModal
                        listData={listData}
                        assetMap={assetMap}
                        media={media}
                        type={type}
                    />

                    {withAudienceScore &&
                        <MediaEntryAudienceScore
                            meanScore={media.meanScore}
                            hideAudienceScore={serverStatus?.settings?.anilist?.hideAudienceScore}
                        />}

                </MediaEntryCardHoverPopupFooter>

            </MediaEntryCardHoverPopup>

            <MediaEntryCardBody
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
                    <MediaEntryScoreBadge
                        score={media.meanScore}
                    />
                </div>
                <div className="absolute z-[10] left-1 bottom-1">
                    <MediaEntryProgressBadge
                        progress={listData?.progress}
                        progressTotal={progressTotal}
                    />
                </div>
            </MediaEntryCardBody>

            <MediaEntryCardTitleSection
                title={media.title?.userPreferred || ""}
                year={media.startDate?.year}
                season={media.season}
                format={media.format}
            />

        </MediaEntryCardContainer>
    )
}
