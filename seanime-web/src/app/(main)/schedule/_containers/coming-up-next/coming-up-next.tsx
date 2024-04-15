import { useAnilistCollection } from "@/app/(main)/_lib/anilist-anime-collection"
import { serverStatusAtom } from "@/atoms/server-status"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Carousel, CarouselContent, CarouselDotButtons, CarouselItem } from "@/components/ui/carousel"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { addSeconds, formatDistanceToNow } from "date-fns"
import { useAtomValue } from "jotai/index"
import Image from "next/image"
import Link from "next/link"
import React from "react"

export function ComingUpNext() {
    const serverStatus = useAtomValue(serverStatusAtom)
    const { anilistLists } = useAnilistCollection()
    const _media = React.useMemo(() => {
        const collectionEntries = anilistLists?.map(n => n?.entries).flat() ?? []
        return collectionEntries.filter(Boolean).map(entry => entry.media) as BaseMediaFragment[]
    }, [anilistLists])

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
            <h2>Coming up next</h2>
            <p className="text-[--muted]">Based on your anime list</p>
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
                                    className="rounded-md border border-gray-800 overflow-hidden aspect-[4/2] relative flex items-end flex-none group/missed-episode-item"
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
                                            className="object-cover object-top transition opacity-20"
                                        /> : <div
                                            className="h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"
                                        ></div>}
                                        <div
                                            className="z-[1] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                                        />
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
