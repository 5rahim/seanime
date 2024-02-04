"use client"
import { __libraryHeaderImageAtom } from "@/app/(main)/(library)/_containers/library-header"
import { LargeEpisodeListItem } from "@/components/shared/large-episode-list-item"
import { Slider } from "@/components/shared/slider"
import { MediaEntryEpisode } from "@/lib/server/types"
import { formatDistanceToNow, isBefore, subYears } from "date-fns"
import { useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import { memo, startTransition, useEffect, useMemo } from "react"

export function ContinueWatching({ list, isLoading }: {
    list: MediaEntryEpisode[],
    isLoading: boolean
}) {

    if (list.length > 0) return (
        <div className="space-y-8 p-4">
            <h2>Continue watching</h2>
            {<Slider>
                {list.map((item, idx) => {
                    return (
                        <EpisodeItem key={item.basicMedia?.id || idx} {...item} />
                    )
                })}
            </Slider>}
        </div>
    )
}

const EpisodeItem = memo((props: MediaEntryEpisode) => {

    const router = useRouter()

    const date = props.episodeMetadata?.airDate ? new Date(props.episodeMetadata.airDate) : undefined
    const setHeaderImage = useSetAtom(__libraryHeaderImageAtom)

    const mediaIsOlder = useMemo(() => date ? isBefore(date, subYears(new Date(), 2)) : undefined, [])

    const offset = props.progressNumber - props.episodeNumber

    useEffect(() => {
        setHeaderImage(prev => {
            if (prev === null) {
                return props.basicMedia?.bannerImage || props.episodeMetadata?.image || null
            }
            return prev
        })
    }, [])

    return (
        <LargeEpisodeListItem
            image={props.episodeMetadata?.image}
            title={<span>{props.displayTitle} {!!props.basicMedia?.episodes &&
                <span className={"opacity-40"}>/{` `}{props.basicMedia.episodes - offset}</span>}</span>}
            topTitle={props.basicMedia?.title?.userPreferred}
            actionIcon={undefined}
            meta={(date) ? (!mediaIsOlder ? `${formatDistanceToNow(date, { addSuffix: true })}` : new Intl.DateTimeFormat(
                "en-US", {
                    day: "2-digit",
                    month: "2-digit",
                    year: "2-digit",
                }).format(date)) : undefined}
            onClick={() => {
                router.push(`/entry?id=${props.basicMedia?.id}&playNext=true`)
            }}
            onMouseEnter={() => {
                startTransition(() => {
                    setHeaderImage(props.basicMedia?.bannerImage || props.episodeMetadata?.image || null)
                })
            }}
            larger
        />
    )
})
