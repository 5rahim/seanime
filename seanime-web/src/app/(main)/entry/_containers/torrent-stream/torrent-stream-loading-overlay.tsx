"use client"
import { Torrentstream_TorrentLoadingStatus, Torrentstream_TorrentLoadingStatusState, Torrentstream_TorrentStatus } from "@/api/generated/types"
import { useTorrentstreamStopStream } from "@/api/hooks/torrentstream.hooks"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Spinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { Tooltip } from "@/components/ui/tooltip"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { BiDownArrow, BiGroup, BiStop, BiUpArrow } from "react-icons/bi"

const enum TorrentStreamLoadingEvents {
    TorrentLoading = "torrentstream-torrent-loading",
    TorrentLoadingStatus = "torrentstream-torrent-loading-status",
    TorrentLoaded = "torrentstream-torrent-loaded",
    TorrentStartedPlaying = "torrentstream-torrent-started-playing",
    TorrentStatus = "torrentstream-torrent-status",
    TorrentStopped = "torrentstream-torrent-stopped",
}

export const enum TorrentStreamState {
    Loaded = "loaded",
    Stopped = "stopped",
}

export const __torrentstream__loadingStateAtom = atom<Torrentstream_TorrentLoadingStatusState | null>(null)
export const __torrentstream__stateAtom = atom<TorrentStreamState>(TorrentStreamState.Stopped)

export function TorrentStreamLoadingOverlay() {

    const [loadingState, setLoadingState] = useAtom(__torrentstream__loadingStateAtom)
    const [state, setState] = useAtom(__torrentstream__stateAtom)

    const [status, setStatus] = useState<Torrentstream_TorrentStatus | null>(null)
    const [torrentBeingLoaded, setTorrentBeingLoaded] = useState<string | null>(null)
    const [mediaPlayerStartedPlaying, setMediaPlayerStartedPlaying] = useState<boolean>(false)

    const { mutate: stop, isPending } = useTorrentstreamStopStream()

    /**
     * Received when the torrent is first being loaded, this is the first message received
     */
    useWebsocketMessageListener({
        type: TorrentStreamLoadingEvents.TorrentLoading,
        onMessage: _ => {
            setLoadingState("SEARCHING_TORRENTS")
            setStatus(null)
            setMediaPlayerStartedPlaying(false)
        },
    })

    /**
     * Received while the torrent is being loaded, checked, etc.
     */
    useWebsocketMessageListener<Torrentstream_TorrentLoadingStatus>({
        type: TorrentStreamLoadingEvents.TorrentLoadingStatus,
        onMessage: data => {
            setLoadingState(data.state)
            setTorrentBeingLoaded(data.torrentBeingChecked)
            setMediaPlayerStartedPlaying(false)
        },
    })

    /**
     * Received when the torrent is loaded and sent to the media player
     */
    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentLoaded,
        onMessage: _ => {
            // The StartStream function returned
            setLoadingState("SENDING_STREAM_TO_MEDIA_PLAYER")
            setState(TorrentStreamState.Loaded)
            setMediaPlayerStartedPlaying(false)
        },
    })

    /**
     * Received when the media player loads the total duration of the video
     */
    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentStartedPlaying,
        onMessage: _ => {
            setLoadingState(null)
            setState(TorrentStreamState.Loaded)
            setMediaPlayerStartedPlaying(true)
        },
    })

    /**
     * Received anytime the torrent streaming process is stopped
     */
    useWebsocketMessageListener<void>({
        type: TorrentStreamLoadingEvents.TorrentStopped,
        onMessage: _ => {
            setLoadingState(null)
            setState(TorrentStreamState.Stopped)
            setStatus(null)
            setMediaPlayerStartedPlaying(false)
        },
    })

    /**
     * Received when the torrent status (downloading, uploading, etc.) changes
     */
    useWebsocketMessageListener<Torrentstream_TorrentStatus>({
        type: TorrentStreamLoadingEvents.TorrentStatus,
        onMessage: data => {
            setState(TorrentStreamState.Loaded)
            setStatus(data)
        },
    })

    if (state === TorrentStreamState.Loaded && status) {
        return (
            <>
                {!mediaPlayerStartedPlaying && <div className="w-full bg-gray-900 fixed top-0 left-0 z-[100]">
                    <ProgressBar isIndeterminate />
                </div>}
                <div className="fixed left-0 top-8 w-full flex justify-center z-[100]">
                    <div className="bg-gray-900 rounded-full border lg:max-w-[50%] w-fit py-3 px-6 flex gap-2 items-center">
                        <BiGroup className="inline-block text-2xl" />
                        <span>{status.seeders}</span>
                        {`  `}
                        <span className={cn({ "text-green-300": status.downloadProgress > 0 })}>{status.progressPercentage.toFixed(2)}%</span>
                        {` `}
                        <BiDownArrow className="inline-block ml-2" />
                        {status.downloadSpeed !== "" ? status.downloadSpeed : "0 B/s"}
                        {` `}
                        <BiUpArrow className="inline-block ml-2" />
                        {status.uploadSpeed !== "" ? status.uploadSpeed : "0 B/s"}
                        {` `}
                        <Tooltip
                            trigger={<IconButton
                                onClick={() => stop()}
                                loading={isPending}
                                intent="alert-basic"
                                icon={<BiStop />}
                            />}
                        >
                            Stop stream
                        </Tooltip>
                    </div>

                </div>
            </>
        )
    }

    if (loadingState) {
        return <>
            <div className="w-full bg-gray-900 fixed top-0 left-0 z-[100]">
                <ProgressBar isIndeterminate />
            </div>
            <div className="fixed left-0 top-8 w-full flex justify-center z-[100]">
                <div className="bg-gray-900 rounded-full border lg:max-w-[50%] w-fit py-3 px-6 flex gap-2 items-center">
                    <Spinner className="w-4 h-4" />
                    <div className="truncate">
                        {loadingState === "SEARCHING_TORRENTS" ? "Searching for torrents..." : ""}
                        {loadingState === "ADDING_TORRENT" ? `Adding torrent "${torrentBeingLoaded}"` : ""}
                        {loadingState === "CHECKING_TORRENT" ? `Checking torrent "${torrentBeingLoaded}"` : ""}
                        {loadingState === "SELECTING_FILE" ? `Selecting file` : ""}
                        {loadingState === "SENDING_STREAM_TO_MEDIA_PLAYER" ? "Sending stream to media player" : ""}
                    </div>
                </div>
            </div>
        </>
    }

    return null

}
