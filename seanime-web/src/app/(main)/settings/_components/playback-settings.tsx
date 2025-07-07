import {
    ElectronPlaybackMethod,
    PlaybackDownloadedMedia,
    PlaybackTorrentStreaming,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { __settings_tabAtom } from "@/app/(main)/settings/_components/settings-page.atoms"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Switch } from "@/components/ui/switch"
import { __isElectronDesktop__ } from "@/types/constants"
import { useSetAtom } from "jotai"
import React from "react"
import { BiDesktop, BiPlay } from "react-icons/bi"
import { IoPlayBackCircleSharp } from "react-icons/io5"
import { LuClapperboard, LuExternalLink, LuLaptop } from "react-icons/lu"
import { MdOutlineBroadcastOnHome } from "react-icons/md"
import { RiSettings3Fill } from "react-icons/ri"
import { toast } from "sonner"

type PlaybackSettingsProps = {
    children?: React.ReactNode
}

export function PlaybackSettings(props: PlaybackSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const {
        downloadedMediaPlayback,
        setDownloadedMediaPlayback,
        torrentStreamingPlayback,
        setTorrentStreamingPlayback,
        electronPlaybackMethod,
        setElectronPlaybackMethod,
    } = useCurrentDevicePlaybackSettings()

    const { activeOnDevice, setActiveOnDevice } = useMediastreamActiveOnDevice()
    const { externalPlayerLink } = useExternalPlayerLink()
    const setTab = useSetAtom(__settings_tabAtom)

    const usingNativePlayer = __isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer

    return (
        <>
            <div className="space-y-4">
                <SettingsPageHeader
                    title="Video playback"
                    description="Choose how anime is played on this device"
                    icon={IoPlayBackCircleSharp}
                />

                <div className="flex items-center gap-2 text-sm bg-gray-50 dark:bg-gray-900/50 rounded-lg p-3 border border-gray-200 dark:border-gray-800">
                    <BiDesktop className="text-lg text-gray-500" />
                    <span className="text-gray-600 dark:text-gray-400">Device:</span>
                    <span className="font-medium">{serverStatus?.clientDevice || "-"}</span>
                    <span className="text-gray-400">•</span>
                    <span className="font-medium">{serverStatus?.clientPlatform || "-"}</span>
                </div>
            </div>

            {(!externalPlayerLink && (downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink || torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink)) && (
                <Alert
                    intent="alert-basic"
                    description={
                        <div className="flex items-center justify-between gap-3">
                            <span>No external player custom scheme has been set</span>
                            <Button
                                intent="gray-outline"
                                size="sm"
                                onClick={() => setTab("external-player-link")}
                            >
                                Add
                            </Button>
                        </div>
                    }
                />
            )}

            {__isElectronDesktop__ && (
                <SettingsCard
                    title="Seanime Denshi"
                    className="border-2 border-dashed dark:border-gray-700 bg-gradient-to-r from-indigo-50/50 to-pink-50/50 dark:from-gray-900/20 dark:to-gray-900/20"
                >
                    <div className="space-y-4">

                        <div className="flex items-center gap-4">
                            <div className="p-3 rounded-lg bg-gradient-to-br from-indigo-500/20 to-indigo-500/20 border border-indigo-500/20">
                                <LuClapperboard className="text-2xl text-indigo-600 dark:text-indigo-400" />
                            </div>
                            <div className="flex-1">
                                <Switch
                                    label="Use built-in player"
                                    help="When enabled, all media will use the built-in player (overrides settings below)"
                                    value={electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer}
                                    onValueChange={v => {
                                        setElectronPlaybackMethod(v ? ElectronPlaybackMethod.NativePlayer : ElectronPlaybackMethod.Default)
                                        toast.success("Playback settings updated")
                                    }}
                                />
                            </div>
                        </div>
                    </div>
                </SettingsCard>
            )}

            <SettingsCard
                title="Downloaded Media"
                description="Choose how to play anime files stored on your device"
                className={cn(
                    "transition-all duration-200",
                    usingNativePlayer && "opacity-50 pointer-events-none",
                )}
            >
                <div className="space-y-4">

                    {/* Option Comparison */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {/* Desktop Player Option */}
                        <div
                            className={cn(
                                "p-4 rounded-lg border cursor-pointer transition-all",
                                downloadedMediaPlayback === PlaybackDownloadedMedia.Default && !activeOnDevice
                                    ? "border-[--brand] bg-brand-900/10"
                                    : "border-gray-700 hover:border-gray-600",
                            )}
                            onClick={() => {
                                setDownloadedMediaPlayback(PlaybackDownloadedMedia.Default)
                                setActiveOnDevice(false)
                                toast.success("Playback settings updated")
                            }}
                        >
                            <div className="flex items-start gap-3">
                                <LuLaptop className="text-xl text-brand-600 dark:text-brand-400 mt-1" />
                                <div className="flex-1 space-y-2">
                                    <div>
                                        <p className="font-medium">Desktop Media Player</p>
                                        <p className="text-xs text-gray-600 dark:text-gray-400">Opens files in your system player with automatic
                                                                                                tracking</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* Web Player Option */}
                        <div
                            className={cn(
                                "p-4 rounded-lg border cursor-pointer transition-all",
                                downloadedMediaPlayback === PlaybackDownloadedMedia.Default && activeOnDevice
                                    ? "border-[--brand] bg-brand-900/10"
                                    : "border-gray-700 hover:border-gray-600",
                                !serverStatus?.mediastreamSettings?.transcodeEnabled && "opacity-50",
                            )}
                            onClick={() => {
                                if (serverStatus?.mediastreamSettings?.transcodeEnabled) {
                                    setDownloadedMediaPlayback(PlaybackDownloadedMedia.Default)
                                    setActiveOnDevice(true)
                                    toast.success("Playback settings updated")
                                }
                            }}
                        >
                            <div className="flex items-start gap-3">
                                <MdOutlineBroadcastOnHome className="text-xl text-brand-600 dark:text-brand-400 mt-1" />
                                <div className="flex-1 space-y-2">
                                    <div>
                                        <p className="font-medium">Transcoding / Direct Play</p>
                                        <p className="text-xs text-gray-600 dark:text-gray-400">
                                            {serverStatus?.mediastreamSettings?.transcodeEnabled
                                                ? "Plays in browser with transcoding"
                                                : "Transcoding not enabled"
                                            }
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* External Player Option */}
                        <div
                            className={cn(
                                "p-4 rounded-lg border cursor-pointer transition-all",
                                downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink
                                    ? "border-[--brand] bg-brand-900/10"
                                    : "border-gray-700 hover:border-gray-600",
                            )}
                            onClick={() => {
                                setDownloadedMediaPlayback(PlaybackDownloadedMedia.ExternalPlayerLink)
                                toast.success("Playback settings updated")
                            }}
                        >
                            <div className="flex items-start gap-3">
                                <LuExternalLink className="text-xl text-brand-600 dark:text-brand-400 mt-1" />
                                <div className="flex-1 space-y-2">
                                    <div>
                                        <p className="font-medium">External Player Link</p>
                                        <p className="text-xs text-gray-600 dark:text-gray-400">Send stream URL to another application</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </SettingsCard>

            <SettingsCard
                title="Torrent & Debrid Streaming"
                description="Choose how to play streamed content from torrents and debrid services"
                className={cn(
                    "transition-all duration-200",
                    usingNativePlayer && "opacity-50 pointer-events-none",
                )}
            >
                <div className="space-y-4">

                    {/* Option Comparison */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {/* Desktop Player Option */}
                        <div
                            className={cn(
                                "p-4 rounded-lg border cursor-pointer transition-all",
                                torrentStreamingPlayback === PlaybackTorrentStreaming.Default
                                    ? "border-[--brand] bg-brand-900/10"
                                    : "border-gray-700 hover:border-gray-600",
                            )}
                            onClick={() => {
                                setTorrentStreamingPlayback(PlaybackTorrentStreaming.Default)
                                toast.success("Playback settings updated")
                            }}
                        >
                            <div className="flex items-start gap-3">
                                <LuLaptop className="text-xl text-brand-600 dark:text-brand-400 mt-1" />
                                <div className="flex-1 space-y-2">
                                    <div>
                                        <p className="font-medium">Desktop Media Player</p>
                                        <p className="text-xs text-gray-600 dark:text-gray-400">Opens streams in your system player with automatic
                                                                                                tracking</p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        {/* External Player Option */}
                        <div
                            className={cn(
                                "p-4 rounded-lg border cursor-pointer transition-all",
                                torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink
                                    ? "border-[--brand] bg-brand-900/10"
                                    : "border-gray-700 hover:border-gray-600",
                            )}
                            onClick={() => {
                                setTorrentStreamingPlayback(PlaybackTorrentStreaming.ExternalPlayerLink)
                                toast.success("Playback settings updated")
                            }}
                        >
                            <div className="flex items-start gap-3">
                                <LuExternalLink className="text-xl text-brand-600 dark:text-brand-400 mt-1" />
                                <div className="flex-1 space-y-2">
                                    <div>
                                        <p className="font-medium">External Player Link</p>
                                        <p className="text-xs text-gray-600 dark:text-gray-400">Send stream URL to another application</p>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </SettingsCard>

            <div className="flex items-center gap-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-200 dark:border-gray-800 border-dashed">
                <RiSettings3Fill className="text-base" />
                <span>Settings are saved automatically</span>
            </div>

            {usingNativePlayer && (
                <div className="text-center">
                    <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-indigo-100 dark:bg-indigo-900/30 border border-indigo-200 dark:border-indigo-700">
                        <BiPlay className="text-indigo-600 dark:text-indigo-400" />
                        <span className="text-sm text-indigo-600 dark:text-indigo-400 font-medium">
                            Native player is active - other settings are disabled
                        </span>
                    </div>
                </div>
            )}
        </>
    )
}
