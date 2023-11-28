import { keepPreviousData, useQuery } from "@tanstack/react-query"
import { addSeconds, formatDistanceToNow, subDays } from "date-fns"
import React from "react"
import { Slider } from "@/components/shared/slider"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { useRouter } from "next/navigation"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { getRecentMediaAirings } from "@/lib/anilist/queries/recent-airings"

export function RecentReleases() {

    const router = useRouter()

    const { data } = useQuery({
        queryKey: ["recent-released"],
        queryFn: async () => {
            return getRecentMediaAirings({
                page: 1,
                perPage: 50,
                airingAt_lesser: Math.floor(new Date().getTime() / 1000),
                airingAt_greater: Math.floor(subDays(new Date(), 14).getTime() / 1000),
            })
        },
        placeholderData: keepPreviousData,
        gcTime: 1000 * 60 * 10,
    })

    return (
        <AppLayoutStack>
            <h2>Recent releases</h2>
            <Slider>
                {data?.Page?.airingSchedules?.filter(item => item?.media?.isAdult === false
                    && item?.media?.type === "ANIME"
                    && item?.media?.countryOfOrigin === "JP"
                    && item?.media?.format !== "TV_SHORT",
                ).filter(Boolean).map(item => {
                    return (
                        <LargeEpisodeListItem
                            key={item.id}
                            title={`Episode ${item.episode}`}
                            image={item.media?.bannerImage || item.media?.coverImage?.large}
                            topTitle={item.media?.title?.userPreferred}
                            meta={item.airingAt ? formatDistanceToNow(addSeconds(new Date(), item.timeUntilAiring), { addSuffix: true }) : undefined}
                            onClick={() => router.push(`/entry?id=${item.media?.id}`)}
                            actionIcon={null}
                        />
                    )
                })}
            </Slider>
        </AppLayoutStack>
    )
}