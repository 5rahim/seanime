import { useGetAnimeCollection } from "@/api/hooks/anilist.hooks"
import { EpisodeCard } from "@/app/(main)/_features/anime/_components/episode-card"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { addSeconds, formatDistanceToNow } from "date-fns"
import { useRouter } from "next/navigation"
import React from "react"

/**
 * @description
 * Displays a carousel of upcoming episodes based on the user's anime list.
 */
export function ComingUpNext() {
    const serverStatus = useServerStatus()
    const router = useRouter()

    const { data: animeCollection } = useGetAnimeCollection()

    const media = React.useMemo(() => {
        // get all media
        const _media = (animeCollection?.MediaListCollection?.lists?.filter(n => n.status !== "DROPPED")
            .map(n => n?.entries)
            .flat() ?? []).map(entry => entry?.media)?.filter(Boolean)
        // keep media with next airing episodes
        let ret = _media.filter(item => !!item.nextAiringEpisode?.episode)
            .sort((a, b) => a.nextAiringEpisode!.timeUntilAiring - b.nextAiringEpisode!.timeUntilAiring)
        if (serverStatus?.settings?.anilist?.enableAdultContent) {
            return ret
        } else {
            // remove adult media
            return ret.filter(item => !item.isAdult)
        }
    }, [animeCollection])

    if (!media?.length) return null

    return (
        <AppLayoutStack>
            {media.length > 0 && (
                <>
                    <div>
                        <h2>Upcoming episodes</h2>
                        <p className="text-[--muted]">Based on your anime list</p>
                    </div>

                    <Carousel
                        className="w-full max-w-full"
                        gap="md"
                        opts={{
                            align: "start",
                        }}
                        autoScroll
                    >
                        <CarouselDotButtons />
                        <CarouselContent>
                            {media.map(item => {
                                return (
                                    <CarouselItem
                                        key={item.id}
                                        className="md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5"
                                    >
                                        <EpisodeCard
                                            key={item.id}
                                            image={item.bannerImage || item.coverImage?.large}
                                            topTitle={item.title?.userPreferred}
                                            title={`Episode ${item.nextAiringEpisode?.episode}`}
                                            meta={formatDistanceToNow(addSeconds(new Date(), item.nextAiringEpisode?.timeUntilAiring!),
                                                { addSuffix: true })}
                                            imageClass="opacity-50"
                                            actionIcon={null}
                                            onClick={() => {
                                                router.push(`/entry?id=${item.id}`)
                                            }}
                                            anime={{
                                                id: item.id,
                                                image: item.coverImage?.large,
                                                title: item.title?.userPreferred,
                                            }}
                                        />
                                    </CarouselItem>
                                )
                            })}
                        </CarouselContent>
                    </Carousel>
                </>
            )}
        </AppLayoutStack>
    )
}
