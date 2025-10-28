import { Anime_Entry, Anime_EntryDownloadEpisode } from "@/api/generated/types"
import { usePlaylistManager } from "@/app/(main)/_features/playlists/_containers/global-playlist-manager"
import { useTorrentSearchSelection } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { TorrentConfirmationContinueButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-download-modal"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { GlowingEffect } from "@/components/shared/glowing-effect"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { cn } from "@/components/ui/core/styling"
import { Vaul, VaulContent } from "@/components/vaul"
import { useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const __torrentSearch_selectionAtom = atom<TorrentSelectionType | undefined>(undefined)
export const __torrentSearch_selectionEpisodeAtom = atom<number | undefined>(undefined)

export type TorrentSelectionType =
    "torrentstream-select"
    | "torrentstream-select-file"
    | "debridstream-select"
    | "debridstream-select-file"
    | "download"

export function TorrentSearchDrawer(props: { entry: Anime_Entry, isPlaylistDrawer?: boolean }) {

    const { entry, isPlaylistDrawer } = props
    const ts = useThemeSettings()

    const [selectionType, setSelection] = useAtom(__torrentSearch_selectionAtom)
    const searchParams = useSearchParams()
    const router = useRouter()
    const pathname = usePathname()
    const mId = searchParams.get("id")
    const downloadParam = searchParams.get("download")

    const { currentPlaylist } = usePlaylistManager()

    useEffect(() => {
        if (!!downloadParam) {
            setSelection("download")
            router.replace(pathname + `?id=${mId}`)
        }
    }, [downloadParam])

    const { onTorrentValidated } = useTorrentSearchSelection({ entry, type: selectionType })

    if (currentPlaylist && !isPlaylistDrawer) return null

    // if (layoutType === "modal") return (
    //     <Modal
    //         open={selectionType !== undefined}
    //         onOpenChange={() => setSelection(undefined)}
    //         // size="xl"
    //         contentClass="max-w-5xl bg-gray-950 bg-opacity-75 firefox:bg-opacity-100 sm:rounded-xl"
    //         title={`${entry?.media?.title?.userPreferred || "Anime"}`}
    //         titleClass="max-w-[500px] text-ellipsis truncate"
    //         data-torrent-search-drawer
    //         overlayClass="bg-gray-950/70 backdrop-blur-sm"
    //         onInteractOutside={e => {if (isPlaylistDrawer) e.preventDefault()}}
    //     >
    //
    //         <AppLayoutStack className="relative z-[1]" data-torrent-search-drawer-content>
    //             {selectionType === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
    //             {!!selectionType && <TorrentSearchContainer type={selectionType} entry={entry} />}
    //         </AppLayoutStack>
    //
    //         <TorrentConfirmationContinueButton type={selectionType || "download"} onTorrentValidated={onTorrentValidated} />
    //     </Modal>
    // )

    return (
        <Vaul
            open={selectionType !== undefined}
            onOpenChange={() => setSelection(undefined)}
        >

            <VaulContent
                className={cn(
                    "bg-gray-950 h-[90%] lg:h-[80%] bg-opacity-95 6xl:max-w-[1900px] firefox:bg-opacity-100 mx-4 lg:mx-8 6xl:mx-auto overflow-hidden",
                    selectionType === "download" && "lg:h-[92.5%] xl:mx-[10rem] 2xl:mx-[20rem]",
                    selectionType === undefined && "lg:h-[80%] xl:mx-[10rem] 2xl:mx-[20rem]",
                )}
            >
                <GlowingEffect
                    spread={40}
                    // blur={1}
                    glow={true}
                    disabled={false}
                    proximity={100}
                    inactiveZone={0.01}
                    className="opacity-30"
                />
                <div className="p-4 lg:p-8 flex-1 overflow-y-auto">
                    <AppLayoutStack className="relative z-[1]" data-torrent-search-drawer-content>
                        {selectionType === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
                        {!!selectionType && <TorrentSearchContainer type={selectionType} entry={entry} />}
                    </AppLayoutStack>

                    <TorrentConfirmationContinueButton type={selectionType || "download"} onTorrentValidated={onTorrentValidated} />
                    <TorrentConfirmationContinueButton type={selectionType || "download"} onTorrentValidated={onTorrentValidated} />
                </div>
            </VaulContent>
        </Vaul>
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
