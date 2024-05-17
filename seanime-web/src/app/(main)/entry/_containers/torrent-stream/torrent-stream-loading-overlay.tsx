"use client"
import { Torrentstream_TorrentLoadingStatus, Torrentstream_TorrentLoadingStatusState } from "@/api/generated/types"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { Spinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useState } from "react"

const enum TorrentStreamLoadingEvents {
    TorrentLoading = "torrentstream-torrent-loading",
    TorrentLoadingStatus = "torrentstream-torrent-loading-status",
    TorrentLoaded = "torrentstream-torrent-loaded",
    TorrentStartedPlaying = "torrentstream-torrent-started-playing",
    TorrentStatus = "torrentstream-torrent-status",
}

export const __torrentstream__loadingStateAtom = atom<Torrentstream_TorrentLoadingStatusState | null>(null)

export function TorrentStreamLoadingOverlay() {

    const [state, setState] = useAtom(__torrentstream__loadingStateAtom)

    const [progress, setProgress] = useState(0)
    const [status, setStatus] = useState("Scanning...")
    const [torrentBeingLoaded, setTorrentBeingLoaded] = useState<string | null>(null)

    useWebsocketMessageListener({
        type: TorrentStreamLoadingEvents.TorrentLoading,
        onMessage: _ => {
            setState("SEARCHING_TORRENTS")
        },
    })

    useWebsocketMessageListener<Torrentstream_TorrentLoadingStatus>({
        type: TorrentStreamLoadingEvents.TorrentLoadingStatus,
        onMessage: data => {
            setState(data.state)
            setTorrentBeingLoaded(data.torrentBeingChecked)
        },
    })

    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentLoaded,
        onMessage: _ => {
            // The StartStream function returned
        },
    })

    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentStartedPlaying,
        onMessage: _ => {
            setState(null)
        },
    })

    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentStatus,
        onMessage: _ => {

        },
    })

    if (!state) return null

    return (
        <>
            <div className="w-full bg-gray-900 fixed top-0 left-0 z-[100]">
                <ProgressBar isIndeterminate />
            </div>
            <div className="fixed left-0 top-8 w-full flex justify-center z-[100]">
                <div className="bg-gray-900 rounded-full border lg:max-w-[50%] w-fit py-3 px-6 flex gap-2 items-center">
                    <Spinner className="w-4 h-4" />
                    <div className="truncate">
                        {state === "SEARCHING_TORRENTS" ? "Searching for torrents..." : ""}
                        {state === "ADDING_TORRENT" ? `Adding torrent "${torrentBeingLoaded}"` : ""}
                        {state === "CHECKING_TORRENT" ? `Checking torrent "${torrentBeingLoaded}"` : ""}
                        {state === "SELECTING_FILE" ? `Selecting file` : ""}
                        {state === "SENDING_STREAM_TO_MEDIA_PLAYER" ? "Sending streaming to media player" : ""}
                        {/*{state === "SENDING_STREAM_TO_MEDIA_PLAYER" ? "Sending streaming to media player" : ""}*/}

                    </div>
                </div>
            </div>
        </>
    )

}
