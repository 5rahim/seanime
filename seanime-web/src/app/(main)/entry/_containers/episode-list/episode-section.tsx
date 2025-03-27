"use client"
import { AL_AnimeDetailsById_Media, Anime_Entry } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"

import { useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { EpisodeItem } from "@/app/(main)/entry/_containers/episode-list/episode-item"
import { UndownloadedEpisodeList } from "@/app/(main)/entry/_containers/episode-list/undownloaded-episode-list"
import { useHandleEpisodeSection } from "@/app/(main)/entry/_lib/handle-episode-section"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"
import { IoLibrarySharp } from "react-icons/io5"

type EpisodeSectionProps = {
    entry: Anime_Entry
    details: AL_AnimeDetailsById_Media | undefined
    bottomSection: React.ReactNode
}

export function EpisodeSection({ entry, details, bottomSection }: EpisodeSectionProps) {
    const ts = useThemeSettings()
    const serverStatus = useServerStatus()
    const { currentView } = useAnimeEntryPageView()

    const {
        media,
        hasInvalidEpisodes,
        episodesToWatch,
        mainEpisodes,
        specialEpisodes,
        ncEpisodes,
        playMediaFile,
    } = useHandleEpisodeSection({ entry })

    const { data: watchHistory } = useGetContinuityWatchHistory()

    const { inject, remove } = useSeaCommandInject()

    React.useEffect(() => {
        if (!media) return

        // Combine all episode types
        const allEpisodes = [
            { ...episodesToWatch?.[0], type: "next" as const },
            ...mainEpisodes.map(ep => ({ ...ep, type: "main" as const })),
            ...specialEpisodes.map(ep => ({ ...ep, type: "special" as const })),
            ...ncEpisodes.map(ep => ({ ...ep, type: "other" as const })),
        ]

        inject("library-episodes", {
            items: allEpisodes.filter(n => !!n.episodeTitle).map(episode => ({
                data: episode,
                id: `${episode.type}-${episode.localFile?.path || ""}-${episode.episodeNumber}`,
                value: `${episode.episodeNumber}`,
                heading: episode.type === "next" ? "Next Episode" :
                    episode.type === "special" ? "Specials" :
                        episode.type === "other" ? "Others" : "Episodes",
                priority: episode.type === "next" ? 2 :
                    episode.type === "main" ? 1 : 0,
                render: () => (
                    <div className="flex gap-1 items-center w-full">
                        <p className="max-w-[70%] truncate">{episode.displayTitle}</p>
                        {!!episode.episodeTitle && (
                            <p className="text-[--muted] flex-1 truncate">- {episode.episodeTitle}</p>
                        )}
                    </div>
                ),
                onSelect: () => playMediaFile({
                    path: episode.localFile?.path ?? "",
                    mediaId: entry.mediaId,
                }),
            })),
            filter: ({ item, input }) => {
                if (!input) return true
                return item.value.toLowerCase().includes(input.toLowerCase())
            },
            shouldShow: () => currentView === "library",
            priority: 1,
        })

        return () => remove("library-episodes")
    }, [media, episodesToWatch, mainEpisodes, specialEpisodes, ncEpisodes, currentView])

    if (!media) return null

    if (!!media && (!entry.listData || !entry.libraryData) && !serverStatus?.isOffline) {
        return <div className="space-y-10">
            {media?.status !== "NOT_YET_RELEASED"
                ? <h4 className="text-yellow-50 flex items-center gap-2"><IoLibrarySharp /> Not in your library</h4>
                : <h5 className="text-yellow-50">Not yet released</h5>}
            <div className="overflow-y-auto pt-4 lg:pt-0 space-y-10">
                <UndownloadedEpisodeList
                    downloadInfo={entry.downloadInfo}
                    media={media}
                />
                {bottomSection}
            </div>
        </div>
    }

    return (
        <>
            <AppLayoutStack spacing="lg" data-episode-section-stack>

                {hasInvalidEpisodes && <Alert
                    intent="alert"
                    description="Some episodes are invalid. Update the metadata to fix this."
                />}


                {episodesToWatch.length > 0 && (
                    <>
                        <Carousel
                            className="w-full max-w-full"
                            gap="md"
                            opts={{
                                align: "start",
                            }}
                            data-episode-carousel
                        >
                            <CarouselDotButtons />
                            <CarouselContent>
                                {episodesToWatch.map((episode, idx) => (
                                    <CarouselItem
                                        key={episode?.localFile?.path || idx}
                                        className={episodeCardCarouselItemClass(ts.smallerEpisodeCarouselSize)}
                                    >
                                        <EpisodeCard
                                            key={episode.localFile?.path || ""}
                                            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
                                            topTitle={episode.episodeTitle || episode?.baseAnime?.title?.userPreferred}
                                            title={episode.displayTitle}
                                            isInvalid={episode.isInvalid}
                                            progressTotal={episode.baseAnime?.episodes}
                                            progressNumber={episode.progressNumber}
                                            episodeNumber={episode.episodeNumber}
                                            length={episode.episodeMetadata?.length}
                                            percentageComplete={getEpisodePercentageComplete(watchHistory, entry.mediaId, episode.episodeNumber)}
                                            minutesRemaining={getEpisodeMinutesRemaining(watchHistory, entry.mediaId, episode.episodeNumber)}
                                            onClick={() => playMediaFile({ path: episode.localFile?.path ?? "", mediaId: entry.mediaId })}
                                            anime={{
                                                id: entry.mediaId,
                                                image: episode.baseAnime?.coverImage?.medium,
                                                title: episode?.baseAnime?.title?.userPreferred,
                                            }}
                                        />
                                    </CarouselItem>
                                ))}
                            </CarouselContent>
                        </Carousel>
                    </>
                )}


                <div className="space-y-10" data-episode-list-stack>
                    <EpisodeListGrid data-episode-list-main>
                        {mainEpisodes.map(episode => (
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                media={media}
                                isWatched={!!entry.listData?.progress && entry.listData.progress >= episode.progressNumber}
                                onPlay={playMediaFile}
                                percentageComplete={getEpisodePercentageComplete(watchHistory, entry.mediaId, episode.episodeNumber)}
                                minutesRemaining={getEpisodeMinutesRemaining(watchHistory, entry.mediaId, episode.episodeNumber)}
                            />
                        ))}
                    </EpisodeListGrid>

                    {!serverStatus?.isOffline && <UndownloadedEpisodeList
                        downloadInfo={entry.downloadInfo}
                        media={media}
                    />}

                    {specialEpisodes.length > 0 && <>
                        <h2>Specials</h2>
                        <EpisodeListGrid data-episode-list-specials>
                            {specialEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playMediaFile}
                                />
                            ))}
                        </EpisodeListGrid>
                    </>}

                    {ncEpisodes.length > 0 && <>
                        <h2>Others</h2>
                        <EpisodeListGrid data-episode-list-others>
                            {ncEpisodes.map(episode => (
                                <EpisodeItem
                                    key={episode.localFile?.path || ""}
                                    episode={episode}
                                    media={media}
                                    onPlay={playMediaFile}
                                />
                            ))}
                        </EpisodeListGrid>
                    </>}

                    {bottomSection}

                </div>
            </AppLayoutStack>
        </>
    )

}
