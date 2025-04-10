import { AL_AnimeDetailsById_Media, AL_BaseAnime, AL_MangaDetailsById_Media, Anime_Entry, Manga_Entry, Nullish } from "@/api/generated/types"
import { useGetAnilistAnimeDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { useGetMangaEntry, useGetMangaEntryDetails } from "@/api/hooks/manga.hooks"
import { TrailerModal } from "@/app/(main)/_features/anime/_components/trailer-modal"
import { AnimeEntryStudio } from "@/app/(main)/_features/media/_components/anime-entry-studio"
import {
    AnimeEntryRankings,
    MediaEntryAudienceScore,
    MediaEntryGenresList,
} from "@/app/(main)/_features/media/_components/media-entry-metadata-components"
import { MediaPageHeaderEntryDetails } from "@/app/(main)/_features/media/_components/media-page-header-components"
import { useHasDebridService, useHasTorrentProvider, useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_components/relations-recommendations-section"
import { TorrentSearchButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-button"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import {
    __torrentSearch_drawerEpisodeAtom,
    __torrentSearch_drawerIsOpenAtom,
    TorrentSearchDrawer,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { MangaRecommendations } from "@/app/(main)/manga/_components/manga-recommendations"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Skeleton } from "@/components/ui/skeleton"
import { getImageUrl } from "@/lib/server/assets"
import { TORRENT_CLIENT } from "@/lib/server/settings"
import { ThemeMediaPageBannerSize, ThemeMediaPageInfoBoxSize, useThemeSettings } from "@/lib/theme/hooks"
import { usePrevious } from "@uidotdev/usehooks"
import { atom } from "jotai"
import { ScopeProvider } from "jotai-scope"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import { usePathname } from "next/navigation"
import React from "react"
import { BiX } from "react-icons/bi"
import { GoArrowLeft } from "react-icons/go"
import { SiAnilist } from "react-icons/si"


// unused

type AnimePreviewModalProps = {
    children?: React.ReactNode
}

const __mediaPreview_mediaIdAtom = atom<{ mediaId: number, type: "anime" | "manga" } | undefined>(undefined)

export function useMediaPreviewModal() {
    const setInfo = useSetAtom(__mediaPreview_mediaIdAtom)
    return {
        setPreviewModalMediaId: (mediaId: number, type: "anime" | "manga") => {
            setInfo({ mediaId, type })
        },
    }
}

export function MediaPreviewModal(props: AnimePreviewModalProps) {

    const {
        children,
        ...rest
    } = props

    const [info, setInfo] = useAtom(__mediaPreview_mediaIdAtom)
    const previousInfo = usePrevious(info)

    const pathname = usePathname()

    React.useEffect(() => {
        setInfo(undefined)
    }, [pathname])

    return (
        <>
            <Modal
                open={!!info}
                onOpenChange={v => setInfo(prev => v ? prev : undefined)}
                contentClass="max-w-7xl relative"
                hideCloseButton
                {...rest}
            >

                {info && <div className="z-[12] absolute right-2 top-2 flex gap-2 items-center">
                    {(!!previousInfo && previousInfo.mediaId !== info.mediaId) && <IconButton
                        intent="white-subtle" size="sm" className="rounded-full" icon={<GoArrowLeft />}
                        onClick={() => {
                            setInfo(previousInfo)
                        }}
                    />}
                    <IconButton
                        intent="alert" size="sm" className="rounded-full" icon={<BiX />}
                        onClick={() => {
                            setInfo(undefined)
                        }}
                    />
                </div>}

                {info?.type === "anime" && <Anime mediaId={info.mediaId} />}
                {info?.type === "manga" && <Manga mediaId={info.mediaId} />}


            </Modal>
        </>
    )
}

function Anime({ mediaId }: { mediaId: number }) {
    const { data: entry, isLoading: entryLoading } = useGetAnimeEntry(mediaId)
    const { data: details, isLoading: detailsLoading } = useGetAnilistAnimeDetails(mediaId)

    return <Content entry={entry} details={details} entryLoading={entryLoading} detailsLoading={detailsLoading} type="anime" />
}

function Manga({ mediaId }: { mediaId: number }) {
    const { data: entry, isLoading: entryLoading } = useGetMangaEntry(mediaId)
    const { data: details, isLoading: detailsLoading } = useGetMangaEntryDetails(mediaId)

    return <Content entry={entry} details={details} entryLoading={entryLoading} detailsLoading={detailsLoading} type="manga" />
}

function Content({ entry, entryLoading, detailsLoading, details, type }: {
    entry: Nullish<Anime_Entry | Manga_Entry>,
    entryLoading: boolean,
    detailsLoading: boolean,
    details: Nullish<AL_AnimeDetailsById_Media | AL_MangaDetailsById_Media>
    type: "anime" | "manga"
}) {

    const serverStatus = useServerStatus()

    const ts = useThemeSettings()
    const media = entry?.media
    const bannerImage = media?.bannerImage || media?.coverImage?.extraLarge

    const { hasTorrentProvider } = useHasTorrentProvider()
    const { hasDebridService } = useHasDebridService()

    return (
        <ScopeProvider atoms={[__torrentSearch_drawerIsOpenAtom, __torrentSearch_drawerEpisodeAtom, __torrentSearch_selectedTorrentsAtom]}>
            <div
                className={cn(
                    "absolute z-[0] opacity-30 w-full rounded-t-[--radius] overflow-hidden",
                    "w-full flex-none object-cover object-center z-[3] bg-[--background] h-[12rem]",
                    ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small ? "lg:h-[23rem]" : "h-[12rem] lg:h-[22rem] 2xl:h-[30rem]",
                )}
            >

                {/*BOTTOM OVERFLOW FADE*/}
                <div
                    className={cn(
                        "w-full z-[2] absolute bottom-[-5rem] h-[5rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                    )}
                />

                <div
                    className={cn(
                        "absolute top-0 left-0 w-full h-full",
                    )}
                >
                    {(!!bannerImage) && <Image
                        src={getImageUrl(bannerImage || "")}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className={cn(
                            "object-cover object-center z-[1]",
                        )}
                    />}

                    {/*LEFT MASK*/}
                    <div
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[100rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background]  transition-opacity to-transparent",
                            "opacity-85 duration-1000",
                            // y > 300 && "opacity-70",
                        )}
                    />
                    <div
                        className={cn(
                            "hidden lg:block max-w-[60rem] xl:max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-25% transition-opacity to-transparent",
                            "opacity-50 duration-500",
                        )}
                    />
                </div>

                {/*BOTTOM FADE*/}
                <div
                    className={cn(
                        "w-full z-[3] absolute bottom-0 h-[50%] bg-gradient-to-t from-[--background] via-transparent via-100% to-transparent",
                    )}
                />

                <div
                    className={cn(
                        "absolute h-full w-full block lg:hidden bg-[--background] opacity-70 z-[2]",
                    )}
                />

            </div>

            {entryLoading && <div className="space-y-4 relative z-[5]">
                <Skeleton
                    className={cn(
                        "h-[12rem]",
                        ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small ? "lg:h-[23rem]" : "h-[12rem] lg:h-[22rem] 2xl:h-[30rem]",
                    )}
                />
                {/*<LoadingSpinner />*/}
            </div>}

            {(!entryLoading && entry) && <>

                <div className="z-[5] relative">
                    <MediaPageHeaderEntryDetails
                        coverImage={entry.media?.coverImage?.extraLarge || entry.media?.coverImage?.large}
                        title={entry.media?.title?.userPreferred}
                        color={entry.media?.coverImage?.color}
                        englishTitle={entry.media?.title?.english}
                        romajiTitle={entry.media?.title?.romaji}
                        startDate={entry.media?.startDate}
                        season={entry.media?.season}
                        progressTotal={(entry.media as AL_BaseAnime)?.episodes}
                        status={entry.media?.status}
                        description={entry.media?.description}
                        listData={entry.listData}
                        media={entry.media!}
                        smallerTitle
                        type="anime"
                    >
                        <div
                            className={cn(
                                "flex gap-3 flex-wrap items-center relative z-[10]",
                                ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid && "justify-center lg:justify-start lg:max-w-[65vw]",
                            )}
                        >
                            <MediaEntryAudienceScore meanScore={entry?.media?.meanScore} badgeClass="bg-transparent" />

                            {(details as AL_AnimeDetailsById_Media)?.studios &&
                                <AnimeEntryStudio studios={(details as AL_AnimeDetailsById_Media)?.studios} />}

                            <MediaEntryGenresList genres={details?.genres} />

                            <div
                                className={cn(
                                    ts.mediaPageBannerInfoBoxSize === ThemeMediaPageInfoBoxSize.Fluid ? "w-full" : "contents",
                                )}
                            >
                                <AnimeEntryRankings rankings={details?.rankings} />
                            </div>
                        </div>
                    </MediaPageHeaderEntryDetails>

                    <div className="mt-6 flex gap-3 items-center">

                        <SeaLink href={type === "anime" ? `/entry?id=${media?.id}` : `/manga/entry?id=${media?.id}`}>
                            <Button className="px-0" intent="gray-link">
                                Open page
                            </Button>
                        </SeaLink>

                        {type === "anime" && !!(entry?.media as AL_BaseAnime)?.trailer?.id && <TrailerModal
                            trailerId={(entry?.media as AL_BaseAnime)?.trailer?.id} trigger={
                            <Button intent="gray-link" className="px-0">
                                Trailer
                            </Button>}
                        />}

                        <SeaLink href={`https://anilist.co/${type}/${entry.mediaId}`} target="_blank">
                            <IconButton intent="gray-link" className="px-0" icon={<SiAnilist className="text-lg" />} />
                        </SeaLink>

                        {(
                            type === "anime" &&
                            entry?.media?.status !== "NOT_YET_RELEASED"
                            && hasTorrentProvider
                            && (
                                serverStatus?.settings?.torrent?.defaultTorrentClient !== TORRENT_CLIENT.NONE
                                || hasDebridService
                            )
                        ) && (
                            <TorrentSearchButton
                                entry={entry as Anime_Entry}
                            />
                        )}
                    </div>

                    {detailsLoading ? <LoadingSpinner /> : <div className="space-y-6 pt-6">
                        {type === "anime" && <RelationsRecommendationsSection entry={entry as Anime_Entry} details={details} />}
                        {type === "manga" && <MangaRecommendations entry={entry as Manga_Entry} details={details} />}
                    </div>}
                </div>

                {/*<div className="absolute top-0 left-0 w-full h-full z-[0] bg-[--background] rounded-xl">*/}
                {/*    <Image*/}
                {/*        src={media?.bannerImage || ""}*/}
                {/*        alt={""}*/}
                {/*        fill*/}
                {/*        quality={100}*/}
                {/*        sizes="20rem"*/}
                {/*        className="object-cover object-center transition opacity-15"*/}
                {/*    />*/}

                {/*    <div*/}
                {/*        className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "*/}
                {/*    ></div>*/}
                {/*</div>*/}

            </>}

            {(type === "anime" && !!entry) && <TorrentSearchDrawer entry={entry as Anime_Entry} />}
        </ScopeProvider>
    )
}
