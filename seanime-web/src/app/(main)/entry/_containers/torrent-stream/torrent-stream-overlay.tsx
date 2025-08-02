"use client"
import { Torrentstream_TorrentStatus } from "@/api/generated/types"
import { useTorrentstreamStopStream } from "@/api/hooks/torrentstream.hooks"
import { nativePlayer_stateAtom } from "@/app/(main)/_features/native-player/native-player.atoms"

import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Spinner } from "@/components/ui/loading-spinner"
import { ProgressBar } from "@/components/ui/progress-bar"
import { Tooltip } from "@/components/ui/tooltip"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { Inter } from "next/font/google"
import React, { useState } from "react"
import { BiDownArrow, BiGroup, BiStop, BiUpArrow } from "react-icons/bi"

const inter = Inter({ subsets: ["latin"] })

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

export function TorrentStreamOverlay({ isNativePlayerComponent = false }: { isNativePlayerComponent?: boolean | string }) {

    const [nativePlayerState, setNativePlayerState] = useAtom(nativePlayer_stateAtom)

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
                        setTimeout(() => {
                            setLoadingState("SEARCHING_TORRENTS")
                            setStatus(null)
                            setMediaPlayerStartedPlaying(false)
                        }, 500)
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

    if (isNativePlayerComponent) {
        return (
            <>
                {/* Native player is fullscreen */}
                {/* It's integrated into the media controller */}
                {nativePlayerState.active && !nativePlayerState.miniPlayer && status &&
                    <div
                        className={cn(
                            "absolute left-0 top-8 w-full flex justify-center z-[100] pointer-events-none",
                            isNativePlayerComponent === "info" && "relative justify-left w-fit top-0 items-center text-white/90",
                            isNativePlayerComponent === "control-bar" && "relative justify-left w-fit top-0 h-full flex items-center px-2 truncate",
                        )}
                    >
                        <div
                            className={cn(
                                "flex-wrap w-fit h-14 flex gap-3 items-center text-sm pointer-events-auto",
                                isNativePlayerComponent === "info" && "!font-medium h-auto py-1",
                            )}
                        >

                            <div className="space-x-1"><BiGroup className="inline-block text-lg" />
                                <span>{status.seeders}</span>
                            </div>

                            <div className="space-x-1">
                                <BiDownArrow className="inline-block mr-2" />
                                {status.downloadSpeed !== "" ? status.downloadSpeed : "0 B/s"}
                            </div>

                            <span
                                className={cn("text-[--muted]",
                                    { "text-[--muted] animate-pulse": status.progressPercentage < 5 })}
                            >{status.progressPercentage.toFixed(
                                2)}%</span>

                            <div className="space-x-1">
                                <BiUpArrow className="inline-block mr-2" />
                                {status.uploadSpeed !== "" ? status.uploadSpeed : "0 B/s"}
                            </div>

                            {isNativePlayerComponent !== "control-bar" && isNativePlayerComponent !== "info" && <Tooltip
                                trigger={<IconButton
                                    onClick={() => stop()}
                                    loading={isPending}
                                    intent="alert-basic"
                                    icon={<BiStop />}
                                />}
                            >
                                Stop stream
                            </Tooltip>}
                        </div>
                    </div>}

                {(!!loadingState && loadingState !== "SENDING_STREAM_TO_MEDIA_PLAYER") &&
                    <div className="fixed left-0 top-8 w-full flex justify-center z-[100] pointer-events-none">
                        <div className="lg:max-w-[50%] w-fit h-14 px-6 flex gap-2 items-center text-sm lg:text-base pointer-events-auto">
                            <Spinner className="w-4 h-4" />
                            <div className="truncate max-w-[500px]">
                                {loadingState === "LOADING" ? "Loading..." : ""}
                                {loadingState === "SEARCHING_TORRENTS" ? "Selecting file..." : ""}
                                {loadingState === "ADDING_TORRENT" ? `Adding torrent "${torrentBeingLoaded}"` : ""}
                                {loadingState === "CHECKING_TORRENT" ? `Checking torrent "${torrentBeingLoaded}"` : ""}
                                {loadingState === "SELECTING_FILE" ? `Selecting file...` : ""}
                                {loadingState === "SENDING_STREAM_TO_MEDIA_PLAYER" ? "Getting metadata..." : ""}
                            </div>
                        </div>
                    </div>}

            </>
        )
    }

    if (isLoaded && status) {
        return (
            <>
                {!mediaPlayerStartedPlaying && !nativePlayerState.active && <div className="w-full bg-gray-950 fixed top-0 left-0 z-[100]">
                    <ProgressBar size="xs" isIndeterminate />
                </div>}
                {/* Normal overlay / Native player is not fullscreen */}
                {(!nativePlayerState.active || nativePlayerState.miniPlayer) &&
                    <div className="fixed left-0 top-8 w-full flex justify-center z-[100] pointer-events-none">
                    <div className="bg-gray-950 flex-wrap rounded-full border lg:max-w-[50%] w-fit h-14 px-6 flex gap-3 items-center text-sm lg:text-base pointer-events-auto">

                        <span
                            className={cn("text-green-300",
                                { "text-[--muted] animate-pulse": status.progressPercentage < 70 })}
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
                    </div>}
            </>
        )
    }

    if (loadingState && !nativePlayerState.active) {
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
