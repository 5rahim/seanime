import { Anime_AnimeEntry, Anime_AnimeEntryDownloadEpisode } from "@/api/generated/types"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { Drawer } from "@/components/ui/drawer"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { Separator } from "@/components/ui/separator"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const __torrentSearch_drawerIsOpenAtom = atom<TorrentSelectionType | undefined>(undefined)
export const __torrentSearch_drawerEpisodeAtom = atom<number | undefined>(undefined)

export type TorrentSelectionType = "select" | "select-file" | "download"

export function TorrentSearchDrawer(props: { entry: Anime_AnimeEntry }) {

    const { entry } = props

    const [type, setter] = useAtom(__torrentSearch_drawerIsOpenAtom)
    const searchParams = useSearchParams()
    const router = useRouter()
    const pathname = usePathname()
    const mId = searchParams.get("id")
    const downloadParam = searchParams.get("download")

    useEffect(() => {
        if (!!downloadParam) {
            setter("download")
            router.replace(pathname + `?id=${mId}`)
        }
    }, [downloadParam])

    return (
        <Drawer
            open={type !== undefined}
            onOpenChange={() => setter(undefined)}
            size="xl"
            title="Search torrents"
        >
            <div
                className="bg-[url(/pattern-2.svg)] z-[0] w-full h-[10rem] absolute opacity-50 top-[-5rem] left-0 bg-no-repeat bg-right bg-contain"
            >
                <div
                    className="w-full absolute bottom-0 h-[10rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                />
            </div>
            <div className="relative z-[1]">
                {type === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
                {!!type && <TorrentSearchContainer type={type} entry={entry} />}
            </div>
        </Drawer>
    )

}

function EpisodeList({ episodes }: { episodes: Anime_AnimeEntryDownloadEpisode[] | undefined }) {

    if (!episodes || !episodes.length) return null

    return (
        <div className="space-y-2 mt-4">
            <h4>Missing episodes:</h4>
            <p>Episode numbers: {episodes.slice(0, 5).map(n => n.episodeNumber).join(", ")}{episodes.length > 5 ? ", ..." : ""}</p>
            <HorizontalDraggableScroll>
                {episodes.filter(Boolean).slice(0, 10).map(item => {
                    return (
                        <EpisodeGridItem
                            key={item.episode + item.aniDBEpisode}
                            media={item.episode?.baseAnime as any}
                            title={item.episode?.displayTitle || item.episode?.baseAnime?.title?.userPreferred || ""}
                            image={item.episode?.episodeMetadata?.image || item.episode?.baseAnime?.coverImage?.large}
                            episodeTitle={item?.episode?.episodeTitle}
                            description={item.episode?.absoluteEpisodeNumber !== item.episodeNumber
                                ? `(Episode ${item?.episode?.absoluteEpisodeNumber})`
                                : undefined}
                            imageContainerClassName="size-20 lg:size-20"
                            className="flex-none w-72"
                            episodeTitleClassName="text-sm lg:text-sm line-clamp-1"
                        />
                    )
                })}
            </HorizontalDraggableScroll>
            <Separator className="!mt-4 mb-4" />
        </div>
    )

}
