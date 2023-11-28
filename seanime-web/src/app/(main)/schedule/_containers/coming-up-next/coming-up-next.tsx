import { Slider } from "@/components/shared/slider"
import React, { useEffect, useMemo } from "react"
import { addSeconds, formatDistanceToNow } from "date-fns"
import Image from "next/image"
import { AppLayoutStack } from "@/components/ui/app-layout"
import Link from "next/link"
import { useQueryClient } from "@tanstack/react-query"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { useAnilistCollection } from "@/lib/server/hooks/media"

export function ComingUpNext() {

    const queryClient = useQueryClient()

    const { anilistLists } = useAnilistCollection()
    const _media = useMemo(() => {
        const collectionEntries = anilistLists?.map(n => n?.entries).flat() ?? []
        return collectionEntries.filter(Boolean).map(entry => entry.media) as BaseMediaFragment[]
    }, [anilistLists])

    const media = _media.filter(item => !!item.nextAiringEpisode?.episode).sort((a, b) => a.nextAiringEpisode!.timeUntilAiring - b.nextAiringEpisode!.timeUntilAiring)

    useEffect(() => {
        (async () => {
            console.log(await queryClient.fetchQuery({ queryKey: ["get-anilist-collection"] }))
        })()
    }, [])

    if (media.length === 0) return null

    return (
        <AppLayoutStack>
            <h2>Coming up next</h2>
            <p className={"text-[--muted]"}>Based on your anime list</p>
            <Slider>
                {media.map(item => {
                    return (
                        <div
                            key={item.id}
                            className={"rounded-md border border-gray-800 overflow-hidden aspect-[4/2] w-96 relative flex items-end flex-none group/missed-episode-item"}
                        >
                            <div
                                className={"absolute w-full h-full rounded-md rounded-b-none overflow-hidden z-[1]"}>
                                {!!item.bannerImage ? <Image
                                    src={item.bannerImage}
                                    alt={""}
                                    fill
                                    quality={100}
                                    sizes="20rem"
                                    className="object-cover object-center transition opacity-20"
                                /> : <div
                                    className={"h-full block absolute w-full bg-gradient-to-t from-gray-800 to-transparent z-[2]"}></div>}
                                <div
                                    className={"z-[1] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background-color] to-transparent"}
                                />
                            </div>
                            <div className={"relative z-[3] w-full p-4 space-y-1"}>
                                <Link
                                    href={`/entry?id=${item.id}`}
                                    className={"w-[80%] line-clamp-1 text-[--muted] font-semibold cursor-pointer"}
                                >
                                    {item.title?.userPreferred}
                                </Link>
                                <div className={"w-full justify-between flex items-center"}>
                                    <p className={"text-xl font-semibold"}>Episode {item.nextAiringEpisode?.episode}</p>
                                    {item.nextAiringEpisode?.timeUntilAiring &&
                                        <p className={"text-[--muted]"}>{formatDistanceToNow(addSeconds(new Date(), item.nextAiringEpisode?.timeUntilAiring), { addSuffix: true })}</p>}
                                </div>
                            </div>
                        </div>
                    )
                })}
            </Slider>
        </AppLayoutStack>
    )
}