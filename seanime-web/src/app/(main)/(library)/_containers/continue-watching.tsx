"use client"
import { Anime_Episode, Continuity_WatchHistory } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { __libraryHeaderImageAtom } from "@/app/(main)/(library)/_components/library-header"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { useSeaCommandInject } from "@/app/(main)/_features/sea-command/use-inject"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { useDebounce } from "@/hooks/use-debounce"
import { anilist_animeIsMovie } from "@/lib/helpers/media"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { useWindowSize } from "@uidotdev/usehooks"
import { atom } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React from "react"
import { seaCommand_compareMediaTitles } from "../../_features/sea-command/utils"

export const __libraryHeaderEpisodeAtom = atom<Anime_Episode | null>(null)

export function ContinueWatching({ episodes, isLoading, linkTemplate }: {
    episodes: Anime_Episode[],
    isLoading: boolean
    linkTemplate?: string
}) {

    const router = useRouter()
    const ts = useThemeSettings()

    const { data: watchHistory } = useGetContinuityWatchHistory()

    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)
    const [headerEpisode, setHeaderEpisode] = useAtom(__libraryHeaderEpisodeAtom)

    const [episodeRefs, setEpisodeRefs] = React.useState<React.RefObject<any>[]>([])
    const [inViewEpisodes, setInViewEpisodes] = React.useState<number[]>([])
    const debouncedInViewEpisodes = useDebounce(inViewEpisodes, 500)

    const { width } = useWindowSize()

    // Create refs for each episode
    React.useEffect(() => {
        setEpisodeRefs(episodes.map(() => React.createRef()))
    }, [episodes])

    // Observe each episode
    React.useEffect(() => {
        const observer = new IntersectionObserver((entries) => {
            entries.forEach((entry) => {
                const index = episodeRefs.findIndex(ref => ref.current === entry.target)
                if (index !== -1) {
                    if (entry.isIntersecting) {
                        setInViewEpisodes(prev => prev.includes(index) ? prev : [...prev, index])
                    } else {
                        setInViewEpisodes(prev => prev.filter((idx) => idx !== index))
                    }
                }
            })
        }, { threshold: 0.5 })  // Trigger callback when 50% of the element is visible

        episodeRefs.forEach((ref) => {
            if (ref.current) {
                observer.observe(ref.current)
            }
        })

        return () => {
            episodeRefs.forEach((ref) => {
                if (ref.current) {
                    observer.unobserve(ref.current)
                }
            })
        }
    }, [episodeRefs, width])

    const prevSelectedEpisodeRef = React.useRef<Anime_Episode | null>(null)

    // Set header image when new episode is in view
    React.useEffect(() => {
        if (debouncedInViewEpisodes.length === 0) return

        let candidateIndices = debouncedInViewEpisodes
        // Exclude previously selected episode if possible
        if (prevSelectedEpisodeRef.current && debouncedInViewEpisodes.length > 1) {
            candidateIndices = debouncedInViewEpisodes.filter(idx => episodes[idx]?.baseAnime?.id !== prevSelectedEpisodeRef.current?.baseAnime?.id)
        }
        if (candidateIndices.length === 0) candidateIndices = debouncedInViewEpisodes

        const randomCandidateIdx = candidateIndices[Math.floor(Math.random() * candidateIndices.length)]
        const selectedEpisode = episodes[randomCandidateIdx]

        if (selectedEpisode) {
            setHeaderImage({
                bannerImage: selectedEpisode.baseAnime?.bannerImage || null,
                episodeImage: selectedEpisode.baseAnime?.bannerImage || selectedEpisode.baseAnime?.coverImage?.extraLarge || null,
            })
            prevSelectedEpisodeRef.current = selectedEpisode
        }
    }, [debouncedInViewEpisodes, episodes, setHeaderImage])

    const { setPlayNext } = usePlayNext()

    const { inject, remove } = useSeaCommandInject()

    React.useEffect(() => {
        inject("continue-watching", {
            items: episodes.map(episode => ({
                data: episode,
                id: `${episode.localFile?.path || episode.baseAnime?.title?.userPreferred || ""}-${episode.episodeNumber || 1}`,
                value: `${episode.episodeNumber || 1}`,
                heading: "Continue Watching",
                priority: 100,
                render: () => (
                    <>
                        <div className="w-12 aspect-[6/5] flex-none rounded-[--radius-md] relative overflow-hidden">
                            <Image
                                src={episode.episodeMetadata?.image || ""}
                                alt="episode image"
                                fill
                                className="object-center object-cover"
                            />
                        </div>
                        <div className="flex gap-1 items-center w-full">
                            <p className="max-w-[70%] truncate">{episode.baseAnime?.title?.userPreferred || ""}</p>&nbsp;-&nbsp;
                            {!anilist_animeIsMovie(episode.baseAnime) ? <>
                                <p className="text-[--muted]">Ep</p><span>{episode.episodeNumber}</span>
                            </> : <>
                                <p className="text-[--muted]">Movie</p>
                            </>}

                        </div>
                    </>
                ),
                onSelect: () => setPlayNext(episode.baseAnime?.id, () => {
                    router.push(`/entry?id=${episode.baseAnime?.id}`)
                }),
            })),
            filter: ({ item, input }) => {
                if (!input) return true
                return item.value.toLowerCase().includes(input.toLowerCase()) ||
                    seaCommand_compareMediaTitles(item.data.baseAnime?.title, input)
            },
            priority: 100,
        })

        return () => remove("continue-watching")
    }, [episodes, inject, remove, router, setPlayNext])

    if (episodes.length > 0) return (
        <PageWrapper className="space-y-3 lg:space-y-6 p-4 relative z-[4]" data-continue-watching-container>
            <h2 data-continue-watching-title>Continue watching</h2>
            {(ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && headerEpisode?.baseAnime) && <TextGenerateEffect
                data-continue-watching-media-title
                words={headerEpisode?.baseAnime?.title?.userPreferred || ""}
                className="w-full text-xl lg:text-5xl lg:max-w-[50%] h-[3.2rem] !mt-1 line-clamp-1 truncate text-ellipsis hidden lg:block pb-1"
            />}
            <Carousel
                className="w-full max-w-full"
                gap="md"
                opts={{
                    align: "start",
                }}
                autoScroll
                autoScrollDelay={8000}
            >
                <CarouselDotButtons />
                <CarouselContent>
                    {episodes.map((episode, idx) => (
                        <CarouselItem
                            key={episode?.localFile?.path || idx}
                            className={episodeCardCarouselItemClass(ts.smallerEpisodeCarouselSize)}
                        >
                            <_EpisodeCard
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                mRef={episodeRefs[idx]}
                                overrideLink={linkTemplate}
                                watchHistory={watchHistory}
                            />
                        </CarouselItem>
                    ))}
                </CarouselContent>
            </Carousel>
        </PageWrapper>
    )
    return null
}

const _EpisodeCard = React.memo(({ episode, mRef, overrideLink, watchHistory }: {
    episode: Anime_Episode,
    mRef: React.RefObject<any>,
    overrideLink?: string
    watchHistory: Continuity_WatchHistory | undefined
}) => {
    const serverStatus = useServerStatus()
    const router = useRouter()
    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)
    const setHeaderEpisode = useSetAtom(__libraryHeaderEpisodeAtom)

    React.useEffect(() => {
        setHeaderImage(prev => {
            if (prev?.episodeImage === null) {
                return {
                    bannerImage: episode.baseAnime?.bannerImage || null,
                    episodeImage: episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge || null,
                }
            }
            return prev
        })
        setHeaderEpisode(prev => {
            if (prev === null) {
                return episode
            }
            return prev
        })
    }, [])

    const { setPlayNext } = usePlayNext()

    return (
        <EpisodeCard
            key={episode.localFile?.path || ""}
            episode={episode}
            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
            topTitle={episode.episodeTitle || episode?.baseAnime?.title?.userPreferred}
            title={episode.displayTitle}
            isInvalid={episode.isInvalid}
            progressTotal={episode.baseAnime?.episodes}
            progressNumber={episode.progressNumber}
            episodeNumber={episode.episodeNumber}
            length={episode.episodeMetadata?.length}
            hasDiscrepancy={episode.episodeNumber !== episode.progressNumber}
            percentageComplete={getEpisodePercentageComplete(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
            minutesRemaining={getEpisodeMinutesRemaining(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
            anime={{
                id: episode?.baseAnime?.id || 0,
                image: episode?.baseAnime?.coverImage?.medium,
                title: episode?.baseAnime?.title?.userPreferred,
            }}
            onMouseEnter={() => {
                React.startTransition(() => {
                    setHeaderImage({
                        bannerImage: episode.baseAnime?.bannerImage || null,
                        episodeImage: episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge || null,
                    })
                })
            }}
            mRef={mRef}
            onClick={() => {
                if (!overrideLink) {
                    setPlayNext(episode.baseAnime?.id, () => {
                        if (!serverStatus?.isOffline) {
                            router.push(`/entry?id=${episode.baseAnime?.id}`)
                        } else {
                            router.push(`/offline/entry/anime?id=${episode.baseAnime?.id}`)
                        }
                    })
                } else {
                    setPlayNext(episode.baseAnime?.id, () => {
                        router.push(overrideLink.replace("{id}", episode.baseAnime?.id ? String(episode.baseAnime.id) : ""))
                    })
                }
            }}
        />
    )
})

