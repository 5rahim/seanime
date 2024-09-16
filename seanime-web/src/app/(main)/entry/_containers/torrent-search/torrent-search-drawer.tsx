import { Anime_Entry, Anime_EntryDownloadEpisode } from "@/api/generated/types"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { Modal } from "@/components/ui/modal"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const __torrentSearch_drawerIsOpenAtom = atom<TorrentSelectionType | undefined>(undefined)
export const __torrentSearch_drawerEpisodeAtom = atom<number | undefined>(undefined)

export type TorrentSelectionType = "select" | "select-file" | "download"

export function TorrentSearchDrawer(props: { entry: Anime_Entry }) {

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
        <Modal
            open={type !== undefined}
            onOpenChange={() => setter(undefined)}
            // size="xl"
            contentClass="max-w-5xl"
            title="Search torrents"
        >
            <div className="">
                <div className="relative z-[1]">
                    {type === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
                    {!!type && <TorrentSearchContainer type={type} entry={entry} />}
                </div>
            </div>
        </Modal>
    )

}


function EpisodeList({ episodes }: { episodes: Anime_EntryDownloadEpisode[] | undefined }) {

    if (!episodes || !episodes.length) return null

    return (
        <div className="space-y-2 mt-4">
            <h4>Missing episodes:</h4>
            <p>Episode {episodes.slice(0, 5).map(n => n.episodeNumber).join(", ")}{episodes.length > 5
                ? `, ..., ${episodes[episodes.length - 1].episodeNumber}`
                : ""}
            </p>
        </div>
    )

}
