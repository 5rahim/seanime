import { Button, IconButton } from "@/components/ui/button"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import {
    __torrentSearch_selectedTorrentsAtom,
} from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { atom } from "jotai"
import { Modal } from "@/components/ui/modal"
import { MediaEntry, SearchTorrent } from "@/lib/server/types"
import { Tooltip } from "@/components/ui/tooltip"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { FcFilmReel } from "@react-icons/all-files/fc/FcFilmReel"
import { FcFolder } from "@react-icons/all-files/fc/FcFolder"
import { BiX } from "@react-icons/all-files/bi/BiX"
import { DirectorySelector } from "@/components/shared/directory-selector"
import React, { useMemo, useState } from "react"
import * as upath from "upath"
import { serverStatusAtom } from "@/atoms/server-status"
import { BiCollection } from "@react-icons/all-files/bi/BiCollection"
import { BiDownload } from "@react-icons/all-files/bi/BiDownload"
import toast from "react-hot-toast"
import { useSeaMutation } from "@/lib/server/queries/utils"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useRouter } from "next/navigation"
import { torrentSearchDrawerIsOpenAtom } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-drawer"

const isOpenAtom = atom(false)

type TorrentDownloadProps = {
    urls: string[]
    destination: string
    smartSelect: {
        enabled: boolean
        missingEpisodeNumbers: number[]
        absoluteOffset: number
    }
    media?: BaseMediaFragment
}

export function TorrentConfirmationModal({ onToggleTorrent, media, entry }: {
    onToggleTorrent: (t: SearchTorrent) => void,
    media: BaseMediaFragment,
    entry: MediaEntry
}) {

    const router = useRouter()
    const serverStatus = useAtomValue(serverStatusAtom)
    const libraryPath = serverStatus?.settings?.library?.libraryPath

    const defaultPath = useMemo(() => {
        const fPath = entry.localFiles?.findLast(n => n)?.path // file path
        const newPath = libraryPath ? upath.join(libraryPath, sanitizeDirectoryName(media.title?.romaji || "")) : ""
        return fPath ? upath.normalize(upath.dirname(fPath)) : newPath
    }, [libraryPath, entry.localFiles, media.title?.romaji])

    const [destination, setDestination] = useState(defaultPath)

    const [isOpen, setIsOpen] = useAtom(isOpenAtom)
    const setTorrentDrawerIsOpen = useSetAtom(torrentSearchDrawerIsOpenAtom)
    const selectedTorrents = useAtomValue(__torrentSearch_selectedTorrentsAtom)


    const canSmartSelect = useMemo(() => {
        return selectedTorrents.length === 1
            && selectedTorrents[0].isBatch
            && media.format !== "MOVIE"
            && media.status === "FINISHED"
            && !!media.episodes && media.episodes > 1
            && !!entry.downloadInfo?.episodesToDownload && entry.downloadInfo?.episodesToDownload.length > 0
            && entry.downloadInfo?.episodesToDownload.length !== (media.episodes || (media.nextAiringEpisode?.episode! - 1))
    }, [selectedTorrents, media.format, media.status, media.episodes, entry.downloadInfo?.episodesToDownload, media.nextAiringEpisode?.episode])


    // mutation
    const { mutate, data, isPending } = useSeaMutation<boolean, TorrentDownloadProps>({
        endpoint: SeaEndpoints.DOWNLOAD,
        method: "post",
        mutationKey: ["download-torrent"],
        onSuccess: () => {
            toast.success("Download started")
            setIsOpen(false)
            setTorrentDrawerIsOpen(false)
            router.push("/torrent-list")
        },
    })

    function handleLaunchDownload(smartSelect: boolean) {
        if (!libraryPath || !destination.toLowerCase().startsWith(libraryPath.slice(0, -1).toLowerCase())) {
            toast.error("Destination folder does not match local library")
            return
        }
        if (smartSelect) {
            mutate({
                urls: selectedTorrents.map(n => n.guid),
                destination,
                smartSelect: {
                    enabled: true,
                    missingEpisodeNumbers: entry.downloadInfo?.episodesToDownload?.map(n => n.episodeNumber) || [],
                    absoluteOffset: entry.downloadInfo?.absoluteOffset || 0,
                },
                media,
            })
        } else {
            mutate({
                urls: selectedTorrents.map(n => n.guid),
                destination,
                smartSelect: {
                    enabled: false,
                    missingEpisodeNumbers: [],
                    absoluteOffset: 0,
                },
                media,
            })
        }
    }

    if (selectedTorrents.length === 0) return null

    return (
        <Modal
            isOpen={isOpen}
            onClose={() => setIsOpen(false)}
            size={"2xl"} isClosable
            title={"Choose the destination"}
        >
            <div className="pb-4">
                <DirectorySelector
                    name="destination"
                    label="Destination"
                    leftIcon={<FcFolder/>}
                    defaultValue={destination}
                    onSelect={setDestination}
                    shouldExist={false}
                />
            </div>

            <div className={"space-y-2"}>
                {selectedTorrents.map(torrent => (
                    <Tooltip
                        key={`${torrent.guid}`}
                        trigger={<div
                            className={"ml-12 gap-2 p-2 border border-[--border] rounded-md hover:bg-gray-800 relative"}
                            key={torrent.name}
                        >
                            <div
                                className={"flex flex-none items-center gap-2 w-[90%] cursor-pointer"}
                                onClick={() => window.open(torrent.guid, "_blank")}
                            >
                                <span className={"text-lg"}>
                                    {(!torrent.isBatch || media.format === "MOVIE") ? <FcFilmReel/> :
                                        <FcFolder className={"text-2xl"}/>} {/*<BsCollectionPlayFill/>*/}
                                </span>
                                <p className={"truncate text-ellipsis"}>{torrent.name}</p>
                            </div>
                            <IconButton
                                icon={<BiX/>}
                                className={"absolute right-2 top-2 rounded-full"}
                                size={"xs"}
                                intent={"gray-outline"}
                                onClick={() => {
                                    onToggleTorrent(torrent)
                                }}
                            />
                        </div>}>
                        Open on NYAA
                    </Tooltip>
                ))}
                <div className={"mt-4 flex w-full justify-end gap-2"}>
                    {(selectedTorrents.length > 0 && canSmartSelect) && <Button
                        leftIcon={<BiCollection/>}
                        intent={"white-outline"}
                        onClick={() => handleLaunchDownload(true)}
                        isLoading={isPending}
                    >Download missing only</Button>}
                    {selectedTorrents.length > 0 && <Button
                        leftIcon={<BiDownload/>}
                        intent={"white"}
                        onClick={() => handleLaunchDownload(false)}
                        isLoading={isPending}
                    >{canSmartSelect ? "Download all" : "Download"}</Button>}
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
            className="animate-pulse"
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