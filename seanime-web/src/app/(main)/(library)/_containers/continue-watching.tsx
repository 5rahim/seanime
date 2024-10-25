"use client"
import { Anime_Episode, Continuity_WatchHistory } from "@/api/generated/types"
import { getEpisodeMinutesRemaining, getEpisodePercentageComplete, useGetContinuityWatchHistory } from "@/api/hooks/continuity.hooks"
import { __libraryHeaderImageAtom } from "@/app/(main)/(library)/_components/library-header"
import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useDeferredValue } from "react"

export const __libraryHeaderEpisodeAtom = atom<Anime_Episode | null>(null)

export function ContinueWatching({ episodes, isLoading, linkTemplate }: {
    episodes: Anime_Episode[],
    isLoading: boolean
    linkTemplate?: string
}) {

    const ts = useThemeSettings()

    const { data: watchHistory } = useGetContinuityWatchHistory()

    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)
    const [headerEpisode, setHeaderEpisode] = useAtom(__libraryHeaderEpisodeAtom)

    const [episodeRefs, setEpisodeRefs] = React.useState<React.RefObject<any>[]>([])
    const [inViewEpisodes, setInViewEpisodes] = React.useState<any>([])
    const debouncedInViewEpisodes = useDeferredValue(inViewEpisodes)

    const debounceTimeout = React.useRef<NodeJS.Timeout | null>(null)

    // Create refs for each episode
    React.useEffect(() => {
        setEpisodeRefs(episodes.map(() => React.createRef()))
    }, [episodes])

    // Observe each episode
    React.useEffect(() => {
        const observer = new IntersectionObserver((entries) => {
            entries.forEach((entry) => {
                const index = episodeRefs.findIndex(ref => ref.current === entry.target)
                if (entry.isIntersecting) {
                    setInViewEpisodes((prev: any) => [...prev, index])
                } else {
                    setInViewEpisodes((prev: any) => prev.filter((idx: number) => idx !== index))
                }
            })
        })

        episodeRefs.forEach((ref) => {
            if (ref.current) {
                observer.observe(ref.current)
            }
        })

        return () => {
            if (episodeRefs.length > 0) {
                episodeRefs.forEach((ref) => {
                    if (ref.current) {
                        observer.unobserve(ref.current)
                    }
                })
            }
        }
    }, [episodeRefs])

    // Set header image when new episode is in view
    React.useEffect(() => {
        if (debounceTimeout.current) {
            clearTimeout(debounceTimeout.current)
        }

        debounceTimeout.current = setTimeout(() => {
            if (inViewEpisodes.length > 0) {
                const randomIndex = inViewEpisodes[Math.floor(Math.random() * inViewEpisodes.length)]
                const episode = episodes[randomIndex]
                if (episode) {
                    setHeaderImage(episode.baseAnime?.bannerImage || episode.episodeMetadata?.image || null)
                }
            }
        }, 500)
        return () => {
            if (debounceTimeout.current) {
                clearTimeout(debounceTimeout.current)
            }
        }
    }, [debouncedInViewEpisodes, episodes])

    if (episodes.length > 0) return (
        <PageWrapper className="space-y-3 lg:space-y-6 p-4 relative z-[4]">
            <h2>Continue watching</h2>
            {/*<h1 className="w-full lg:max-w-[50%] line-clamp-1 truncate hidden lg:block pb-1">{headerEpisode?.baseAnime?.title?.userPreferred}</h1>*/}
            {(ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && headerEpisode?.baseAnime) && <TextGenerateEffect
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
            if (prev === null) {
                return episode.baseAnime?.bannerImage || episode.episodeMetadata?.image || null
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
            image={episode.episodeMetadata?.image || episode.baseAnime?.bannerImage || episode.baseAnime?.coverImage?.extraLarge}
            topTitle={episode.episodeTitle || episode?.baseAnime?.title?.userPreferred}
            title={episode.displayTitle}
            // meta={episode.episodeMetadata?.airDate ?? undefined}
            isInvalid={episode.isInvalid}
            progressTotal={episode.baseAnime?.episodes}
            progressNumber={episode.progressNumber}
            episodeNumber={episode.episodeNumber}
            length={episode.episodeMetadata?.length}
            hasDiscrepancy={episode.episodeNumber !== episode.progressNumber}
            percentageComplete={getEpisodePercentageComplete(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
            minutesRemaining={getEpisodeMinutesRemaining(watchHistory, episode.baseAnime?.id || 0, episode.episodeNumber)}
            onMouseEnter={() => {
                React.startTransition(() => {
                    setHeaderImage(episode.baseAnime?.bannerImage || episode.episodeMetadata?.image || null)
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

