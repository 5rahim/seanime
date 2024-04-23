import { MediaEntry } from "@/app/(main)/(library)/_lib/anime-library.types"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { AnimeTorrent } from "@/app/(main)/entry/_containers/torrent-search/_lib/torrent.types"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { torrentSearchDrawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { Button, IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useMemo, useState } from "react"
import { BiCollection, BiDownload, BiX } from "react-icons/bi"
import { FcFilmReel, FcFolder } from "react-icons/fc"
import { toast } from "sonner"
import * as upath from "upath"

const isOpenAtom = atom(false)

type TorrentDownloadProps = {
    urls: string[]
    destination: string
    smartSelect: {
        enabled: boolean
        missingEpisodeNumbers: number[]
    }
    media?: BaseMediaFragment
}

type TorrentDownloadFileProps = {
    download_urls: string[]
    destination: string
    media?: BaseMediaFragment
}

export function TorrentConfirmationModal({ onToggleTorrent, media, entry }: {
    onToggleTorrent: (t: AnimeTorrent) => void,
    media: BaseMediaFragment,
    entry: MediaEntry
}) {

    const router = useRouter()
    const serverStatus = useAtomValue(serverStatusAtom)
    const libraryPath = serverStatus?.settings?.library?.libraryPath

    /**
     * Default path for the destination folder
     */
    const defaultPath = useMemo(() => {
        const fPath = entry.localFiles?.findLast(n => n)?.path // file path
        const newPath = libraryPath ? upath.join(libraryPath, sanitizeDirectoryName(media.title?.romaji || "")) : ""
        return fPath ? upath.normalize(upath.dirname(fPath)) : newPath
    }, [libraryPath, entry.localFiles, media.title?.romaji])

    const [destination, setDestination] = useState(defaultPath)

    const [isOpen, setIsOpen] = useAtom(isOpenAtom)
    const setTorrentDrawerIsOpen = useSetAtom(torrentSearchDrawerIsOpenAtom)
    const selectedTorrents = useAtomValue(__torrentSearch_selectedTorrentsAtom)

    /**
     * If the user can auto-select the missing episodes
     */
    const canSmartSelect = useMemo(() => {
        return selectedTorrents.length === 1
            && selectedTorrents[0].isBatch
            && media.format !== "MOVIE"
            && media.status === "FINISHED"
            && !!media.episodes && media.episodes > 1
            && !!entry.downloadInfo?.episodesToDownload && entry.downloadInfo?.episodesToDownload.length > 0
            && entry.downloadInfo?.episodesToDownload.length !== (media.episodes || (media.nextAiringEpisode?.episode! - 1))
    }, [
        selectedTorrents,
        media.format,
        media.status,
        media.episodes,
        entry.downloadInfo?.episodesToDownload,
        media.nextAiringEpisode?.episode,
        serverStatus?.settings?.torrent?.defaultTorrentClient,
    ])


    // download via torrent client
    const { mutate, isPending } = useSeaMutation<boolean, TorrentDownloadProps>({
        endpoint: SeaEndpoints.TORRENT_CLIENT_DOWNLOAD,
        method: "post",
        mutationKey: ["download-torrent"],
        onSuccess: () => {
            toast.success("Download started")
            setIsOpen(false)
            setTorrentDrawerIsOpen(false)
            router.push("/torrent-list")
        },
    })

    // download torrent file
    const { mutate: downloadTorrentFiles, isPending: isDownloadingFiles } = useSeaMutation<boolean, TorrentDownloadFileProps>({
        endpoint: SeaEndpoints.DOWNLOAD_TORRENT_FILE,
        method: "post",
        mutationKey: ["download-torrent-files"],
        onSuccess: () => {
            toast.success("Downloaded torrent files")
            setIsOpen(false)
            setTorrentDrawerIsOpen(false)
        },
    })

    const isDisabled = isPending || isDownloadingFiles

    function handleLaunchDownload(smartSelect: boolean) {
        if (!libraryPath || !destination.toLowerCase().startsWith(libraryPath.slice(0, -1).toLowerCase())) {
            toast.error("Destination folder does not match local library")
            return
        }
        if (smartSelect) {
            mutate({
                urls: selectedTorrents.map(n => n.provider === "seadex" ? n.infoHash : n.link),
                destination,
                smartSelect: {
                    enabled: true,
                    missingEpisodeNumbers: entry.downloadInfo?.episodesToDownload?.map(n => n.episodeNumber) || [],
                },
                media,
            })
        } else {
            mutate({
                urls: selectedTorrents.map(n => n.provider === "seadex" ? n.infoHash : n.link),
                destination,
                smartSelect: {
                    enabled: false,
                    missingEpisodeNumbers: [],
                },
                media,
            })
        }
    }

    function handleDownloadFiles() {
        downloadTorrentFiles({
            download_urls: selectedTorrents.map(n => n.link),
            destination,
            media,
        })
    }

    if (selectedTorrents.length === 0) return null

    return (
        <Modal
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}
            contentClass="max-w-3xl"
            title="Choose the destination"
        >
            <div className="pb-0">
                <DirectorySelector
                    name="destination"
                    label="Destination"
                    leftIcon={<FcFolder />}
                    value={destination}
                    defaultValue={destination}
                    onSelect={setDestination}
                    shouldExist={false}
                />
            </div>

            <div className="space-y-2">
                {selectedTorrents.map(torrent => (
                    <Tooltip
                        key={`${torrent.link}`}
                        trigger={<div
                            className="ml-12 gap-2 p-2 border rounded-md hover:bg-gray-800 relative"
                            key={torrent.name}
                        >
                            <div
                                className="flex flex-none items-center gap-2 w-[90%] cursor-pointer"
                                onClick={() => window.open(torrent.link, "_blank")}
                            >
                                <span className="text-lg">
                                    {(!torrent.isBatch || media.format === "MOVIE") ? <FcFilmReel /> :
                                        <FcFolder className="text-2xl" />} {/*<BsCollectionPlayFill/>*/}
                                </span>
                                <p className="line-clamp-1">{torrent.name}</p>
                            </div>
                            <IconButton
                                icon={<BiX />}
                                className="absolute right-2 top-2 rounded-full"
                                size="xs"
                                intent="gray-outline"
                                onClick={() => {
                                    onToggleTorrent(torrent)
                                }}
                            />
                        </div>}
                    >
                        Open in browser
                    </Tooltip>
                ))}
                <div className="!mt-4 flex w-full justify-between gap-2 items-center">
                    <div>
                        {!selectedTorrents?.some(t => t.provider === "seadex") && <Button
                            leftIcon={<BiDownload />}
                            intent="gray-outline"
                            onClick={() => handleDownloadFiles()}
                            disabled={isDisabled}
                            loading={isDownloadingFiles}
                        >Download torrent files</Button>}
                    </div>

                    <div className="flex w-full justify-end gap-2">

                        {(selectedTorrents.length > 0 && canSmartSelect) && (
                            <Button
                                leftIcon={<BiCollection />}
                                intent="white-outline"
                                onClick={() => handleLaunchDownload(true)}
                                disabled={isDisabled}
                                loading={isPending}
                            >
                                Download missing episodes
                            </Button>
                        )}

                        {selectedTorrents.length > 0 && (
                            <Button
                                leftIcon={<BiDownload />}
                                intent="white"
                                onClick={() => handleLaunchDownload(false)}
                                disabled={isDisabled}
                                loading={isPending}
                            >
                                {canSmartSelect ? "Download all" : "Download"}
                            </Button>
                        )}

                    </div>
                </div>
            </div>
        </Modal>
    )

}


export function TorrentConfirmationContinueButton() {

    const st = useAtomValue(__torrentSearch_selectedTorrentsAtom)
    const setter = useSetAtom(isOpenAtom)

    if (st.length === 0) return null

    return (
        <Button
            intent="primary"
            className=""
            onClick={() => setter(true)}
        >
            Continue ({st.length})
        </Button>
    )

}

function sanitizeDirectoryName(input: string): string {
    const disallowedChars = /[<>:"/\\|?*\x00-\x1F]/g // Pattern for disallowed characters
    // Replace disallowed characters with an underscore
    const sanitized = input.replace(disallowedChars, " ")
    // Remove leading/trailing spaces and dots (periods) which are not allowed
    const trimmed = sanitized.trim().replace(/^\.+|\.+$/g, "").replace(/\s+/g, " ")
    // Ensure the directory name is not empty after sanitization
    return trimmed || "Untitled"
}
