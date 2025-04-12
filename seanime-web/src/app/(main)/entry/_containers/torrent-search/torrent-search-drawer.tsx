import { Anime_Entry, Anime_EntryDownloadEpisode } from "@/api/generated/types"
import { useHandleTorrentSelection } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { TorrentConfirmationContinueButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Modal } from "@/components/ui/modal"
import { getImageUrl } from "@/lib/server/assets"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const __torrentSearch_drawerIsOpenAtom = atom<TorrentSelectionType | undefined>(undefined)
export const __torrentSearch_drawerEpisodeAtom = atom<number | undefined>(undefined)

export type TorrentSelectionType =
    "select" // torrent streaming, torrent selection
    | "select-file" // torrent streaming, torrent & file selection
    | "debrid-stream-select" // debrid streaming, torrent selection only
    | "debrid-stream-select-file"  // debrid streaming, torrent & file selection
    | "download"

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

    const { onTorrentValidated } = useHandleTorrentSelection({ entry, type })

    return (
        <Modal
            open={type !== undefined}
            onOpenChange={() => setter(undefined)}
            // size="xl"
            contentClass="max-w-5xl"
            title={`${entry?.media?.title?.userPreferred || "Anime"}`}
            titleClass="max-w-[500px] text-ellipsis truncate"
            data-torrent-search-drawer
        >

            {entry?.media?.bannerImage && <div
                data-torrent-search-drawer-banner-image-container
                className="Sea-TorrentSearchDrawer__bannerImage h-36 w-full flex-none object-cover object-center overflow-hidden rounded-t-xl absolute left-0 top-0 z-[-1]"
            >
                <Image
                    data-torrent-search-drawer-banner-image
                    src={getImageUrl(entry?.media?.bannerImage!)}
                    alt="banner"
                    fill
                    quality={80}
                    priority
                    sizes="20rem"
                    className="object-cover object-center opacity-10"
                />
                <div
                    data-torrent-search-drawer-banner-image-bottom-gradient
                    className="Sea-TorrentSearchDrawer__bannerImage-bottomGradient z-[5] absolute bottom-0 w-full h-[70%] bg-gradient-to-t from-[--background] to-transparent"
                />
            </div>}

            <AppLayoutStack className="relative z-[1]" data-torrent-search-drawer-content>
                {type === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
                {!!type && <TorrentSearchContainer type={type} entry={entry} />}
            </AppLayoutStack>

            <TorrentConfirmationContinueButton type={type || "download"} onTorrentValidated={onTorrentValidated} />
        </Modal>
    )

}


function EpisodeList({ episodes }: { episodes: Anime_EntryDownloadEpisode[] | undefined }) {

    if (!episodes || !episodes.length) return null

    const missingEpisodes = episodes.sort((a, b) => a.episodeNumber - b.episodeNumber)

    return (
        <div className="space-y-2" data-torrent-search-drawer-episode-list>
            <p><span className="font-semibold">Missing episode{missingEpisodes.length > 1 ? "s" : ""}</span>: {missingEpisodes.slice(0, 5)
                .map(n => n.episodeNumber)
                .join(", ")}{missingEpisodes.length > 5
                ? `, ..., ${missingEpisodes[missingEpisodes.length - 1].episodeNumber}`
                : ""}
            </p>
        </div>
    )

}
