import { useGetAnilistCollection } from "@/api/hooks/anilist.hooks"
import { AnimeListItemBottomGradient } from "@/app/(main)/_features/custom-ui/item-bottom-gradients"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { MonthCalendar } from "@/app/(main)/schedule/_components/month-calendar"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { addSeconds, formatDistanceToNow } from "date-fns"
import Image from "next/image"
import Link from "next/link"
import React from "react"

/**
 * @description
 * Displays a carousel of upcoming episodes based on the user's anime list.
 */
export function ComingUpNext() {
    const serverStatus = useServerStatus()
    const { data: anilistCollection } = useGetAnilistCollection()
    const _media = React.useMemo(() => {
        const collectionEntries = anilistCollection?.MediaListCollection?.lists?.map(n => n?.entries).flat() ?? []
        return collectionEntries?.map(entry => entry?.media)?.filter(Boolean)
    }, [anilistCollection])

    const media = React.useMemo(() => {
        let ret = _media.filter(item => !!item.nextAiringEpisode?.episode)
            .sort((a, b) => a.nextAiringEpisode!.timeUntilAiring - b.nextAiringEpisode!.timeUntilAiring)
        if (serverStatus?.settings?.anilist?.enableAdultContent) {
            return ret
        } else {
            return ret.filter(item => !item.isAdult)
        }
    }, [_media])

    if (media.length === 0) return null

    return (
        <AppLayoutStack>
            <h2>Release schedule</h2>
            <p className="text-[--muted]">Based on your anime list</p>

            <MonthCalendar />

            <h2>Coming up next</h2>

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
                                <div
                                    className="rounded-md border bg-[--background] border-gray-800 overflow-hidden aspect-[4/2] relative flex items-end flex-none group/upcoming-episode-item"
                                >
                                    <div
                                        className="absolute w-full h-full rounded-md rounded-b-none overflow-hidden z-[1]"
                                    >
                                        {(!!item.bannerImage || !!item.coverImage?.large) ? <Image
                                            src={item.bannerImage || item.coverImage?.large || ""}
                                            alt={""}
                                            fill
                                            quality={100}
                                            sizes="20rem"
                                            className="object-cover object-top transition-opacity opacity-20 group-hover/upcoming-episode-item:opacity-30"
                                        /> : <div
                                            className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                                        ></div>}
                                        <AnimeListItemBottomGradient />
                                    </div>
                                    <div className="relative z-[3] w-full p-4 space-y-1">
                                        <Link
                                            href={`/entry?id=${item.id}`}
                                            className="w-[80%] line-clamp-1 text-[--muted] font-semibold cursor-pointer"
                                        >
                                            {item.title?.userPreferred}
                                        </Link>
                                        <div className="w-full justify-between flex items-center">
                                            <p className="text-xl font-semibold">Episode {item.nextAiringEpisode?.episode}</p>
                                            {item.nextAiringEpisode?.timeUntilAiring &&
                                                <p className="text-[--muted]">{formatDistanceToNow(addSeconds(new Date(),
                                                    item.nextAiringEpisode?.timeUntilAiring), { addSuffix: true })}</p>}
                                        </div>
                                    </div>
                                </div>
                            </CarouselItem>
                        )
                    })}
                </CarouselContent>
            </Carousel>
        </AppLayoutStack>
    )
}
