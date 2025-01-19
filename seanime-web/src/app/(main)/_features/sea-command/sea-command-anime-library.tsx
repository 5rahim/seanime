import { usePlayNext } from "@/app/(main)/_atoms/playback.atoms"

import { CommandGroup, CommandItem } from "@/components/ui/command"
import { anilist_animeIsSingleEpisode } from "@/lib/helpers/media"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React from "react"
import { useSeaCommandContext } from "./sea-command"

export function SeaCommandAnimeLibrary() {

    const { params: { page, pageParams }, input, setInput, close } = useSeaCommandContext<"anime-library">()

    const router = useRouter()

    const { setPlayNext } = usePlayNext()

    const items = React.useMemo(() => {
        if (!pageParams?.episodes) return []
        if (!input) return pageParams.episodes
        return pageParams.episodes.filter(episode =>
            episode.baseAnime?.title?.userPreferred?.toLowerCase().includes(input.toLowerCase()) ||
            episode.episodeNumber?.toString().includes(input),
        )
    }, [pageParams?.episodes, input])

    if (!items.length) return null

    return (
        <>
            <CommandGroup heading="Episodes">
                {items.map(episode => (
                    <CommandItem
                        key={episode.baseAnime?.id || ""}
                        onSelect={() => {
                            setPlayNext(episode.baseAnime?.id, () => {
                                router.push(`/entry?id=${episode.baseAnime?.id}`)
                                close()
                            })
                        }}
                        className="flex gap-3 items-center"
                    >
                        <div className="w-12 aspect-[6/5] rounded-md relative overflow-hidden">
                            <Image
                                src={episode.episodeMetadata?.image || ""}
                                alt="episode image"
                                fill
                                className="object-center object-cover"
                            />
                        </div>
                        <div className="flex gap-1 items-center w-full">
                            <p className="max-w-[70%] truncate">{episode.baseAnime?.title?.userPreferred || ""}</p>&nbsp;-&nbsp;
                            {!anilist_animeIsSingleEpisode(episode.baseAnime) && <>
                                <p className="text-[--muted]">Ep</p><span>{episode.episodeNumber}</span>
                            </>}

                        </div>
                    </CommandItem>
                ))}
            </CommandGroup>
        </>
    )
}
