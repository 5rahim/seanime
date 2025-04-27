import { AL_BaseAnime, Anime_Entry, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import { useDebridAddTorrents } from "@/api/hooks/debrid.hooks"
import { useDownloadTorrentFile } from "@/api/hooks/download.hooks"
import { useTorrentClientDownload } from "@/api/hooks/torrent_client.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __torrentSearch_selectedTorrentsAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { __torrentSearch_drawerIsOpenAtom, TorrentSelectionType } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"
import { DirectorySelector } from "@/components/shared/directory-selector"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { Switch } from "@/components/ui/switch"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { upath } from "@/lib/helpers/upath"
import { TORRENT_CLIENT } from "@/lib/server/settings"
import { atom } from "jotai"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useMemo, useState } from "react"
import { AiOutlineCloudServer } from "react-icons/ai"
import { BiCollection, BiDownload, BiX } from "react-icons/bi"
import { FcFilmReel, FcFolder } from "react-icons/fc"
import { LuDownload, LuPlay } from "react-icons/lu"

const isOpenAtom = atom(false)

export function TorrentConfirmationModal({ onToggleTorrent, media, entry }: {
    onToggleTorrent: (t: HibikeTorrent_AnimeTorrent) => void,
    media: AL_BaseAnime,
    entry: Anime_Entry
}) {

    const router = useRouter()
    const serverStatus = useServerStatus()
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
    const setTorrentDrawerIsOpen = useSetAtom(__torrentSearch_drawerIsOpenAtom)
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
    const { mutate, isPending } = useTorrentClientDownload(() => {
        setIsOpen(false)
        setTorrentDrawerIsOpen(undefined)
        router.push("/torrent-list")
    })

    // download torrent file
    const { mutate: downloadTorrentFiles, isPending: isDownloadingFiles } = useDownloadTorrentFile(() => {
        setIsOpen(false)
        setTorrentDrawerIsOpen(undefined)
    })

    // download via debrid service
    const { mutate: debridAddTorrents, isPending: isDownloadingDebrid } = useDebridAddTorrents(() => {
        setIsOpen(false)
        setTorrentDrawerIsOpen(undefined)
        router.push("/debrid")
    })

    const isDisabled = isPending || isDownloadingFiles || isDownloadingDebrid

    function handleLaunchDownload(smartSelect: boolean) {
        if (smartSelect) {
            mutate({
                torrents: selectedTorrents,
                destination,
                smartSelect: {
                    enabled: true,
                    missingEpisodeNumbers: entry.downloadInfo?.episodesToDownload?.map(n => n.episodeNumber) || [],
                },
                media,
            })
        } else {
            mutate({
                torrents: selectedTorrents,
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
            download_urls: selectedTorrents.map(n => n.downloadUrl),
            destination,
            media,
        })
    }

    function handleDebridAddTorrents() {
        debridAddTorrents({
            torrents: selectedTorrents,
            destination,
            media,
        })
    }

    const debridActive = serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider
    const [isDebrid, setIsDebrid] = useState(debridActive)

    if (selectedTorrents.length === 0) return null

    return (
        <Modal
            open={isOpen}
            onOpenChange={() => setIsOpen(false)}
            contentClass="max-w-3xl"
            title="Choose the destination"
            data-torrent-confirmation-modal
        >

            {debridActive && (
                <Switch
                    label="Download with Debrid service"
                    value={isDebrid}
                    onValueChange={v => setIsDebrid(v)}
                />
            )}

            <DirectorySelector
                name="destination"
                label="Destination"
                leftIcon={<FcFolder />}
                value={destination}
                defaultValue={destination}
                onSelect={setDestination}
                shouldExist={false}
            />

            {selectedTorrents.map(torrent => (
                <Tooltip
                    data-torrent-confirmation-modal-tooltip
                    key={`${torrent.link}`}
                    trigger={<div
                        className={cn(
                            "ml-12 gap-2 p-2 border rounded-[--radius-md] hover:bg-gray-800 relative",
                        )}
                        key={torrent.name}
                        data-torrent-confirmation-modal-torrent-item
                    >
                        <div
                            data-torrent-confirmation-modal-torrent-item-content
                            className="flex flex-none items-center gap-2 w-[90%] cursor-pointer"
                            onClick={() => openTab(torrent.link)}
                        >
                            <span className="text-lg" data-torrent-confirmation-modal-torrent-item-icon>
                                {(!torrent.isBatch || media.format === "MOVIE") ? <FcFilmReel /> :
                                    <FcFolder className="text-2xl" />} {/*<BsCollectionPlayFill/>*/}
                            </span>
                            <p className="line-clamp-1" data-torrent-confirmation-modal-torrent-item-name>
                                {torrent.name}
                            </p>
                        </div>
                        <IconButton
                            icon={<BiX />}
                            className="absolute right-2 top-2 rounded-full"
                            size="xs"
                            intent="gray-outline"
                            onClick={() => {
                                onToggleTorrent(torrent)
                            }}
                            data-torrent-confirmation-modal-torrent-item-close-button
                        />
                    </div>}
                >
                    Open in browser
                </Tooltip>
            ))}

            {isDebrid ? (
                <>
                    {(serverStatus?.debridSettings?.enabled && !!serverStatus?.debridSettings?.provider) && (
                        <Button
                            data-torrent-confirmation-modal-debrid-button
                            leftIcon={<AiOutlineCloudServer className="text-2xl" />}
                            intent="white"
                            onClick={() => handleDebridAddTorrents()}
                            disabled={isDisabled}
                            loading={isDownloadingDebrid}
                            className="w-full"
                        >
                            Download with Debrid service
                        </Button>
                    )}
                </>
            ) : (
                <>
                    <div className="space-y-2" data-torrent-confirmation-modal-download-buttons>

                        <div className="flex w-full gap-2" data-torrent-confirmation-modal-download-buttons-left>
                            {!!selectedTorrents?.every(t => t.downloadUrl) && <Button
                                data-torrent-confirmation-modal-download-files-button
                                leftIcon={<BiDownload />}
                                intent="gray-outline"
                                onClick={() => handleDownloadFiles()}
                                disabled={isDisabled}
                                loading={isDownloadingFiles}
                                className="w-full"
                            >Download '.torrent' files</Button>}

                            {selectedTorrents.length > 0 && (
                                <Button
                                    data-torrent-confirmation-modal-download-button
                                    leftIcon={<BiDownload />}
                                    intent="white"
                                    onClick={() => handleLaunchDownload(false)}
                                    disabled={isDisabled || serverStatus?.settings?.torrent?.defaultTorrentClient === TORRENT_CLIENT.NONE}
                                    loading={isPending}
                                    className="w-full"
                                >
                                    {!serverStatus?.debridSettings?.enabled
                                        ? (canSmartSelect ? "Download all" : "Download")
                                        : "Download with torrent client"}
                                </Button>
                            )}
                        </div>

                        {(selectedTorrents.length > 0 && canSmartSelect) && (
                            <Button
                                data-torrent-confirmation-modal-download-missing-episodes-button
                                leftIcon={<BiCollection />}
                                intent="gray-outline"
                                onClick={() => handleLaunchDownload(true)}
                                disabled={isDisabled}
                                loading={isPending}
                                className="w-full"
                            >
                                Download missing episodes
                            </Button>
                        )}

                    </div>
                </>
            )}
        </Modal>
    )

}


export function TorrentConfirmationContinueButton({ type, onTorrentValidated }: { type: TorrentSelectionType, onTorrentValidated: () => void }) {

    const st = useAtomValue(__torrentSearch_selectedTorrentsAtom)
    const setter = useSetAtom(isOpenAtom)

    if (st.length === 0) return null

    return (
        <Button
            data-torrent-search-confirmation-continue-button
            intent="primary"
            className="Sea-TorrentSearchConfirmationContinueButton fixed z-[9999] left-0 right-0 bottom-4 rounded-full max-w-lg mx-auto halo"
            size="lg"
            onClick={() => {
                if (type === "download") {
                    setter(true)
                } else {
                    onTorrentValidated()
                }
            }}
            leftIcon={type === "download" ? <LuDownload /> : <LuPlay />}
        >
            {type === "download" ? "Download" : "Stream"}
            {type === "download" ? ` (${st.length})` : ""}
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
