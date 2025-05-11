"use client"
import { Torrentstream_TorrentStatus } from "@/api/generated/types"
import { useTorrentstreamStopStream } from "@/api/hooks/torrentstream.hooks"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Spinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { Tooltip } from "@/components/ui/tooltip"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { BiDownArrow, BiGroup, BiStop, BiUpArrow } from "react-icons/bi"

const enum TorrentStreamEvents {
    TorrentLoading = "loading",
    TorrentLoadingFailed = "loading-failed",
    TorrentLoadingStatus = "loading-status",
    TorrentLoaded = "loaded",
    TorrentStartedPlaying = "started-playing",
    TorrentStatus = "status",
    TorrentStopped = "stopped",
}

export const __torrentstream__loadingStateAtom = atom<string | null>(null)
export const __torrentstream__isLoadedAtom = atom<boolean>(false)

// uncomment for testing
// export const __torrentstream__loadingStateAtom = atom<Torrentstream_TorrentLoadingStatusState | null>("SEARCHING_TORRENTS")
// export const __torrentstream__stateAtom = atom<TorrentStreamState>(TorrentStreamState.Loaded)

export function TorrentStreamOverlay() {

    const [loadingState, setLoadingState] = useAtom(__torrentstream__loadingStateAtom)
    const [isLoaded, setIsLoaded] = useAtom(__torrentstream__isLoadedAtom)

    const [status, setStatus] = useState<Torrentstream_TorrentStatus | null>(null)
    const [torrentBeingLoaded, setTorrentBeingLoaded] = useState<string | null>(null)
    const [mediaPlayerStartedPlaying, setMediaPlayerStartedPlaying] = useState<boolean>(false)

    const { mutate: stop, isPending } = useTorrentstreamStopStream()

    useWebsocketMessageListener({
        type: WSEvents.TORRENTSTREAM_STATE,
        onMessage: ({ state, data }: { state: string, data: any }) => {
            switch (state) {
                case TorrentStreamEvents.TorrentLoading:
                    if (!data) {
                        setLoadingState("SEARCHING_TORRENTS")
                        setStatus(null)
                        setMediaPlayerStartedPlaying(false)
                    } else {
                        setLoadingState(data.state)
                        setTorrentBeingLoaded(data.torrentBeingLoaded)
                        setMediaPlayerStartedPlaying(false)
                    }
                    break
                case TorrentStreamEvents.TorrentLoadingFailed:
                    setLoadingState(null)
                    setStatus(null)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentLoaded:
                    setLoadingState("SENDING_STREAM_TO_MEDIA_PLAYER")
                    setIsLoaded(true)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentStartedPlaying:
                    setLoadingState(null)
                    setIsLoaded(true)
                    setMediaPlayerStartedPlaying(true)
                    break
                case TorrentStreamEvents.TorrentStopped:
                    setLoadingState(null)
                    setIsLoaded(false)
                    setStatus(null)
                    setMediaPlayerStartedPlaying(false)
                    break
                case TorrentStreamEvents.TorrentStatus:
                    setIsLoaded(true)
                    setStatus(data)
                    break
            }
        },
    })

    if (isLoaded && status) {
        return (
            <>
                {!mediaPlayerStartedPlaying && <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                    <ProgressBar size="xs" isIndeterminate />
                </div>}
                <div className="fixed left-0 top-8 w-full flex justify-center z-[100] pointer-events-none">
                    <div className="bg-gray-950 flex-wrap rounded-full border lg:max-w-[50%] w-fit h-14 px-6 flex gap-3 items-center text-sm lg:text-base pointer-events-auto">

                        <span
                            className={cn("text-green-300",
                                { "text-[--muted] animate-pulse": status.progressPercentage < 5 })}
                        >{status.progressPercentage.toFixed(
                            2)}%</span>

                        <div className="space-x-1"><BiGroup className="inline-block text-lg" />
                            <span>{status.seeders}</span>
                        </div>

                        <div className="space-x-1">
                            <BiDownArrow className="inline-block mr-2" />
                            {status.downloadSpeed !== "" ? status.downloadSpeed : "0 B/s"}
                        </div>

                        <div className="space-x-1">
                            <BiUpArrow className="inline-block mr-2" />
                            {status.uploadSpeed !== "" ? status.uploadSpeed : "0 B/s"}
                        </div>

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
            <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                <ProgressBar size="xs" isIndeterminate />
            </div>
            <div className="fixed left-0 top-8 w-full flex justify-center z-[100] pointer-events-none">
                <div className="bg-gray-950 rounded-full border lg:max-w-[50%] w-fit h-14 px-6 flex gap-2 items-center text-sm lg:text-base pointer-events-auto">
                    <Spinner className="w-4 h-4" />
                    <div className="truncate max-w-[500px]">
                        {loadingState === "LOADING" ? "Loading..." : ""}
                        {loadingState === "SEARCHING_TORRENTS" ? "Selecting file..." : ""}
                        {loadingState === "ADDING_TORRENT" ? `Adding torrent "${torrentBeingLoaded}"` : ""}
                        {loadingState === "CHECKING_TORRENT" ? `Checking torrent "${torrentBeingLoaded}"` : ""}
                        {loadingState === "SELECTING_FILE" ? `Selecting file...` : ""}
                        {loadingState === "SENDING_STREAM_TO_MEDIA_PLAYER" ? "Sending stream to media player" : ""}
                    </div>
                </div>
            </div>
        </>
    }

    return null

}
