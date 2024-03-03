"use client"
import { __libraryHeaderImageAtom } from "@/app/(main)/(library)/_containers/library-header"
import { SliderEpisodeItem } from "@/components/shared/slider-episode-item"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { TextGenerateEffect } from "@/components/shared/styling/text-generate-effect"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { cn } from "@/components/ui/core/styling"
import { MediaEntryEpisode } from "@/lib/server/types"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai/index"
import { useAtom, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useDeferredValue, useEffect } from "react"

export const __libraryHeaderEpisodeAtom = atom<MediaEntryEpisode | null>(null)

export function ContinueWatching({ list, isLoading }: {
    list: MediaEntryEpisode[],
    isLoading: boolean
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
        setEpisodeRefs(list.map(() => React.createRef()))
    }, [list])

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
                const episode = list[randomIndex]
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
    }, [debouncedInViewEpisodes, list])

    if (list.length > 0) return (
        <PageWrapper className="space-y-3 lg:space-y-6 p-4 relative z-[4]">
            <h2>Continue watching</h2>
            {/*<h1 className="w-full lg:max-w-[50%] line-clamp-1 truncate hidden lg:block pb-1">{headerEpisode?.basicMedia?.title?.userPreferred}</h1>*/}
            {ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && <TextGenerateEffect
                words={headerEpisode?.basicMedia?.title?.userPreferred || ""}
                className="w-full text-xl lg:text-5xl lg:max-w-[50%] !mt-1 line-clamp-1 truncate hidden lg:block pb-1"
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
                    {list.map((episode, idx) => (
                        <CarouselItem
                            key={episode?.localFile?.path || idx}
                            className={cn(
                                !ts.smallerEpisodeCarouselSize && "md:basis-1/2 lg:basis-1/2 2xl:basis-1/3 min-[2000px]:basis-1/4",
                                ts.smallerEpisodeCarouselSize && "md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5",
                            )}
                        >
                            <EpisodeItem
                                key={episode.localFile?.path || ""}
                                episode={episode}
                                mRef={episodeRefs[idx]}
                            />
                        </CarouselItem>
                    ))}
                </CarouselContent>
            </Carousel>
        </PageWrapper>
    )
}

const EpisodeItem = React.memo(({ episode, mRef }: { episode: MediaEntryEpisode, mRef: React.RefObject<any> }) => {
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
        <SliderEpisodeItem
            key={episode.localFile?.path || ""}
            episode={episode}
            onMouseEnter={() => {
                React.startTransition(() => {
                    setHeaderImage(episode.basicMedia?.bannerImage || episode.episodeMetadata?.image || null)
                })
            }}
            ref={mRef}
            onClick={() => {
                router.push(`/entry?id=${episode.basicMedia?.id}&playNext=true`)
            }}
        />
    )
})

