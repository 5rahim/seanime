import { MediaEntry, MediaEntryDownloadEpisode } from "@/lib/server/types"
import { Drawer } from "@/components/ui/modal"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { Slider } from "@/components/shared/slider"
import { EpisodeListItem } from "@/components/shared/episode-list-item"
import { Divider } from "@/components/ui/divider"
import { useEffect } from "react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"

export const torrentSearchDrawerIsOpenAtom = atom(false)
export const torrentSearchDrawerEpisodeAtom = atom<number | undefined>(undefined)

export function TorrentSearchDrawer(props: { entry: MediaEntry }) {

    const { entry } = props

    const [isOpen, setter] = useAtom(torrentSearchDrawerIsOpenAtom)
    const searchParams = useSearchParams()
    const router = useRouter()
    const pathname = usePathname()
    const mId = searchParams.get("id")
    const downloadParam = searchParams.get("download")

    useEffect(() => {
        if (!!downloadParam) {
            setter(true)
            router.replace(pathname + `?id=${mId}`)
        }
    }, [downloadParam])

    return (
        <Drawer
            isOpen={isOpen}
            onClose={() => setter(false)}
            isClosable
            size="xl"
            title="Search torrents"
        >
            <EpisodeList episodes={entry.downloadInfo?.episodesToDownload}/>
            <TorrentSearchContainer entry={entry}/>
        </Drawer>
    )

}

function EpisodeList({ episodes }: { episodes: MediaEntryDownloadEpisode[] | undefined }) {

    if (!episodes || !episodes.length) return null

    return (
        <div>
            <div className="space-y-2">
                <h4>Missing episodes:</h4>
                <p>Episode numbers: {episodes.map(n => n.episodeNumber).join(", ")}</p>
                <Slider>
                    {episodes.filter(Boolean).map(item => {
                        return (
                            <EpisodeListItem
                                key={item.episode + item.aniDBEpisode}
                                media={item.episode?.basicMedia as any}
                                title={item.episode?.displayTitle || ""}
                                image={item.episode?.episodeMetadata?.image}
                                episodeTitle={item?.episode?.episodeTitle}
                                description={item.episode?.absoluteEpisodeNumber !== item.episodeNumber ? `(Episode ${item?.episode?.absoluteEpisodeNumber})` : undefined}
                                imageContainerClassName="w-20 h-20"
                                className="flex-none w-80"
                            />
                        )
                    })}
                </Slider>
            </div>
            <Divider className="py-2 mt-4"/>
        </div>
    )

}