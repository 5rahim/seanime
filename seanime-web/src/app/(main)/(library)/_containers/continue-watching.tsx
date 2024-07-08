"use client"
import { Anime_MediaEntryEpisode } from "@/api/generated/types"
import { __libraryHeaderImageAtom } from "@/app/(main)/(library)/_components/library-header"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { episodeCardCarouselItemClass } from "@/components/shared/classnames"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/text-generate-effect"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useDeferredValue, useEffect } from "react"

export const __libraryHeaderEpisodeAtom = atom<Anime_MediaEntryEpisode | null>(null)

export function ContinueWatching({ episodes, isLoading, linkTemplate }: {
    episodes: Anime_MediaEntryEpisode[],
    isLoading: boolean
    linkTemplate?: string
}) {

    const ts = useThemeSettings()

    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)
    const [headerEpisode, setHeaderEpisode] = useAtom(__libraryHeaderEpisodeAtom)

    const [episodeRefs, setEpisodeRefs] = React.useState<React.RefObject<any>[]>([])
    const [inViewEpisodes, setInViewEpisodes] = React.useState<any>([])
    const debouncedInViewEpisodes = useDeferredValue(inViewEpisodes)

    const debounceTimeout = React.useRef<NodeJS.Timeout | null>(null)

    // Create refs for each episode
    useEffect(() => {
        setEpisodeRefs(episodes.map(() => React.createRef()))
    }, [episodes])

    // Observe each episode
    useEffect(() => {
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
    useEffect(() => {
        if (debounceTimeout.current) {
            clearTimeout(debounceTimeout.current)
        }

        debounceTimeout.current = setTimeout(() => {
            if (inViewEpisodes.length > 0) {
                const randomIndex = inViewEpisodes[Math.floor(Math.random() * inViewEpisodes.length)]
                const episode = episodes[randomIndex]
                if (episode) {
                    setHeaderImage(episode.basicMedia?.bannerImage || episode.episodeMetadata?.image || null)
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
            {/*<h1 className="w-full lg:max-w-[50%] line-clamp-1 truncate hidden lg:block pb-1">{headerEpisode?.basicMedia?.title?.userPreferred}</h1>*/}
            {(ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && headerEpisode?.basicMedia) && <TextGenerateEffect
                words={headerEpisode?.basicMedia?.title?.userPreferred || ""}
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
                            />
                        </CarouselItem>
                    ))}
                </CarouselContent>
            </Carousel>
        </PageWrapper>
    )
}

const _EpisodeCard = React.memo(({ episode, mRef, overrideLink }: {
    episode: Anime_MediaEntryEpisode,
    mRef: React.RefObject<any>,
    overrideLink?: string
}) => {
    const router = useRouter()
    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)
    const setHeaderEpisode = useSetAtom(__libraryHeaderEpisodeAtom)

    useEffect(() => {
        setHeaderImage(prev => {
            if (prev === null) {
                return episode.basicMedia?.bannerImage || episode.episodeMetadata?.image || null
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

    return (
        <EpisodeCard
            key={episode.localFile?.path || ""}
            image={episode.episodeMetadata?.image || episode.basicMedia?.bannerImage || episode.basicMedia?.coverImage?.extraLarge}
            topTitle={episode.episodeTitle || episode?.basicMedia?.title?.userPreferred}
            title={episode.displayTitle}
            // meta={episode.episodeMetadata?.airDate ?? undefined}
            isInvalid={episode.isInvalid}
            progressTotal={episode.basicMedia?.episodes}
            progressNumber={episode.progressNumber}
            episodeNumber={episode.episodeNumber}
            length={episode.episodeMetadata?.length}
            onMouseEnter={() => {
                React.startTransition(() => {
                    setHeaderImage(episode.basicMedia?.bannerImage || episode.episodeMetadata?.image || null)
                })
            }}
            mRef={mRef}
            onClick={() => {
                if (!overrideLink) {
                    router.push(`/entry?id=${episode.basicMedia?.id}&playNext=true`)
                } else {
                    router.push(overrideLink.replace("{id}", episode.basicMedia?.id ? String(episode.basicMedia.id) : ""))
                }
            }}
        />
    )
})

