import { usePatchSetting } from "@/api/hooks/settings.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackDownloadedMedia,
    PlaybackTorrentStreaming,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { mc_settings } from "@/app/(main)/_features/mpv-core/mpv-core.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { __settings_tabAtom } from "@/app/(main)/settings/_components/settings-page.atoms"
import { ExperimentalBadge } from "@/components/shared/beta-badge.tsx"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { RadioGroup } from "@/components/ui/radio-group"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { __isElectronDesktop__ } from "@/types/constants"
import { useAtom, useSetAtom } from "jotai"
import React from "react"
import { BiDesktop } from "react-icons/bi"
import { FaHtml5 } from "react-icons/fa"
import { LuCheck, LuCirclePlay, LuExternalLink, LuLaptop, LuPlay } from "react-icons/lu"
import { MdOutlineBroadcastOnHome } from "react-icons/md"
import { PiVideoDuotone } from "react-icons/pi"
import { RiSettings3Fill } from "react-icons/ri"
import { SiMpv } from "react-icons/si"
import { toast } from "sonner"

type PlaybackChoice = {
    value: string
    title: string
    description: string
    icon: React.ElementType
    preview: React.ReactNode
    badge?: React.ReactNode
    disabled?: boolean
}

type PlaybackChoiceGroupProps = {
    value: string
    options: PlaybackChoice[]
    onValueChange: (value: string) => void
    columns?: "two" | "three"
    disabled?: boolean
}

export function PlaybackSettings() {

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
    const { mutate: patchSetting, isPending: isPatching } = usePatchSetting()

    const [mpvSettings, setMpvSettings] = useAtom(mc_settings)
    const [isExportingMpvLogs, setIsExportingMpvLogs] = React.useState(false)

    const usingNativePlayer = __isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer
    const usingMpvPlayer = usingNativePlayer && serverStatus?.settings?.mediaPlayer?.mpvPrismEnabled
    const isMediastreamEnabled = !!serverStatus?.mediastreamSettings?.transcodeEnabled

    const downloadedMethod = downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink
        ? PlaybackDownloadedMedia.ExternalPlayerLink
        : activeOnDevice ? "mediastream" : PlaybackDownloadedMedia.Default

    const engineMethod = serverStatus?.settings?.mediaPlayer?.mpvPrismEnabled ? "mpvcore" : "videocore"

    function notifyUpdated() {
        toast.success("Playback settings updated")
    }

    function handleDenshiMethodChange(value: string) {
        setElectronPlaybackMethod(value as ElectronPlaybackMethod)
        notifyUpdated()
    }

    function handleDenshiEngineChange(value: string) {
        const enabled = value === "mpvcore"
        if (enabled === !!serverStatus?.settings?.mediaPlayer?.mpvPrismEnabled) return

        patchSetting({
            path: "mediaPlayer.mpvPrismEnabled",
            value: enabled,
        })
    }

    function handleDownloadedMethodChange(value: string) {
        switch (value) {
            case PlaybackDownloadedMedia.Default:
                setDownloadedMediaPlayback(PlaybackDownloadedMedia.Default)
                setActiveOnDevice(false)
                notifyUpdated()
                return
            case "mediastream":
                if (!isMediastreamEnabled) return
                setDownloadedMediaPlayback(PlaybackDownloadedMedia.Default)
                setActiveOnDevice(true)
                notifyUpdated()
                return
            case PlaybackDownloadedMedia.ExternalPlayerLink:
                setDownloadedMediaPlayback(PlaybackDownloadedMedia.ExternalPlayerLink)
                notifyUpdated()
                return
        }
    }

    function handleTorrentMethodChange(value: string) {
        setTorrentStreamingPlayback(value)
        notifyUpdated()
    }

    async function handleExportMpvLogs() {
        try {
            setIsExportingMpvLogs(true)
            if (!window.electron?.mpvCore) throw new Error("MpvCore is not available")
            await window.electron.mpvCore.exportLogs()
            toast.success("MpvCore logs exported")
        }
        catch (error) {
            let msg = error instanceof Error ? error.message : "Failed to export MpvCore logs"
            msg = msg.replace(/^Error:\s*/i, "").replace(/Error invoking remote method '.*?':\s*/i, "")
            toast.error(msg)
        }
        finally {
            setIsExportingMpvLogs(false)
        }
    }

    return (
        <>
            <div className="space-y-4">
                <SettingsPageHeader
                    title="Video playback"
                    description="Choose how anime is played on this device"
                    icon={LuCirclePlay}
                />

                <div className="flex flex-wrap items-center gap-2 text-sm bg-[--paper] rounded-lg p-3 border border-[--border]">
                    <BiDesktop className="text-lg text-[--muted]" />
                    <span className="text-[--muted]">Device:</span>
                    <span className="font-medium">{serverStatus?.clientDevice || "-"}</span>
                    <span className="text-[--muted]">/</span>
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
                    className="border-[--border] bg-[--paper]"
                >
                    <div className="space-y-5">
                        <div className="flex items-center gap-4">
                            <div className="p-3 rounded-lg border border-[--border] bg-[--subtle]">
                                <PiVideoDuotone className="text-2xl text-[--brand]" />
                            </div>
                            <div className="flex-1">
                                <Switch
                                    label="Use built-in player"
                                    help="When enabled, all media playback will use the built-in player (overrides settings below)"
                                    value={electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer}
                                    onValueChange={v => {
                                        setElectronPlaybackMethod(v ? ElectronPlaybackMethod.NativePlayer : ElectronPlaybackMethod.Default)
                                        notifyUpdated()
                                    }}
                                />
                            </div>
                        </div>

                        {usingNativePlayer && (
                            <div className="space-y-4 border-t border-[--border] pt-5">
                                <div className="flex flex-wrap items-center justify-between gap-3">
                                    <div>
                                        <p className="font-semibold">Built-in player engine</p>
                                        <p className="text-sm text-[--muted]">Choose the renderer Denshi should use for integrated playback.</p>
                                    </div>
                                </div>

                                <PlaybackChoiceGroup
                                    columns="two"
                                    value={engineMethod}
                                    onValueChange={handleDenshiEngineChange}
                                    disabled={isPatching}
                                    options={[
                                        {
                                            value: "videocore",
                                            title: "VideoCore",
                                            description: "HTML5 player powered by Chromium's video handling.",
                                            icon: FaHtml5,
                                            preview: <VideoCorePreview />,
                                        },
                                        {
                                            value: "mpvcore",
                                            title: "MpvCore",
                                            description: "Native player powered by libmpv, with broader codec support.",
                                            icon: SiMpv,
                                            badge: <ExperimentalBadge />,
                                            preview: <MpvCorePreview />,
                                        },
                                    ]}
                                />

                                {usingMpvPlayer && (
                                    <div className="space-y-1 pl-4 border-l border-[--border] ml-2">
                                        <div className="flex flex-wrap items-center gap-3">
                                            <div className="min-w-0 flex-1">
                                                <Switch
                                                    label="Enable logging"
                                                    side="right"
                                                    help="If enabled, debug logs will be written to the Denshi data directory."
                                                    value={serverStatus?.settings?.mediaPlayer?.mpvPrismLogging ?? false}
                                                    onValueChange={v => {
                                                        patchSetting({
                                                            path: "mediaPlayer.mpvPrismLogging",
                                                            value: v,
                                                        })
                                                    }}
                                                    disabled={isPatching}
                                                />
                                            </div>
                                        </div>
                                        {serverStatus?.settings?.mediaPlayer?.mpvPrismLogging && <div className="py-2">
                                            <Button
                                                intent="white"
                                                size="sm"
                                                loading={isExportingMpvLogs}
                                                onClick={handleExportMpvLogs}
                                            >
                                                Export logs
                                            </Button>
                                        </div>}
                                        <div className="space-y-2 pt-4 border-t border-[--border] mt-4">
                                            <div className="flex justify-between items-center">
                                                <label className="text-sm font-semibold">Custom MPV Options</label>
                                            </div>
                                            <p className="text-xs text-muted-foreground">
                                                Add custom <code>mpv.conf</code> options.
                                            </p>
                                            <Textarea
                                                value={mpvSettings.customMpvConfig || ""}
                                                onValueChange={value => {
                                                    setMpvSettings({
                                                        ...mpvSettings,
                                                        customMpvConfig: value,
                                                    })
                                                }}
                                                placeholder="# Add custom settings here"
                                                className="font-mono text-sm mt-1 h-[300px]"
                                                fieldClass=""
                                                size="sm"
                                            />
                                        </div>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>
                </SettingsCard>
            )}

            <SettingsCard
                title="Downloaded Media"
                description="Choose how to play anime files stored on your device."
                className={cn(
                    "transition-all duration-200",
                    usingNativePlayer && "opacity-60",
                )}
            >
                <div className="space-y-4">
                    {usingNativePlayer && <OverrideNotice />}
                    <PlaybackChoiceGroup
                        value={downloadedMethod}
                        onValueChange={handleDownloadedMethodChange}
                        options={[
                            {
                                value: PlaybackDownloadedMedia.Default,
                                title: "Desktop media player",
                                description: "Open the stream in your configured player with automatic tracking.",
                                icon: LuLaptop,
                                preview: <DesktopPlayerPreview />,
                            },
                            {
                                value: "mediastream",
                                title: "Transcoding / Direct Play",
                                description: isMediastreamEnabled
                                    ? "Play local files through an HTML5 video player, available on web."
                                    : "Enable transcoding first to use the browser player.",
                                icon: MdOutlineBroadcastOnHome,
                                preview: <MediastreamPreview disabled={!isMediastreamEnabled} />,
                                badge: !isMediastreamEnabled ? <Badge intent="warning" size="sm">Disabled</Badge> : undefined,
                                disabled: !isMediastreamEnabled,
                            },
                            {
                                value: PlaybackDownloadedMedia.ExternalPlayerLink,
                                title: "External player link",
                                description: "Send the stream URL to another app using your custom scheme.",
                                icon: LuExternalLink,
                                preview: <ExternalLinkPreview />,
                            },
                        ]}
                    />
                </div>
            </SettingsCard>

            <SettingsCard
                title="Torrent & Debrid Streaming"
                description="Choose how to play streamed content from torrents and debrid services."
                className={cn(
                    "transition-all duration-200",
                    usingNativePlayer && "opacity-60",
                )}
            >
                <div className="space-y-4">
                    {usingNativePlayer && <OverrideNotice />}
                    <PlaybackChoiceGroup
                        columns="two"
                        value={torrentStreamingPlayback}
                        onValueChange={handleTorrentMethodChange}
                        options={[
                            {
                                value: PlaybackTorrentStreaming.Default,
                                title: "Desktop media player",
                                description: "Open the stream in your configured player with automatic tracking.",
                                icon: LuLaptop,
                                preview: <TorrentDesktopPreview />,
                            },
                            {
                                value: PlaybackTorrentStreaming.ExternalPlayerLink,
                                title: "External player link",
                                description: "Send the stream URL to another app using your custom scheme.",
                                icon: LuExternalLink,
                                preview: <TorrentExternalPreview />,
                            },
                        ]}
                    />
                </div>
            </SettingsCard>

            <div className="flex items-center gap-2 text-sm text-[--muted] bg-[--paper] rounded-lg p-3 border border-[--border] border-dashed">
                <RiSettings3Fill className="text-base" />
                <span>Settings are saved automatically</span>
            </div>
        </>
    )
}

function PlaybackChoiceGroup(props: PlaybackChoiceGroupProps) {
    const { value, options, onValueChange, columns = "three", disabled } = props

    return (
        <RadioGroup
            value={value}
            onValueChange={onValueChange}
            disabled={disabled}
            options={options.map(option => ({
                value: option.value,
                disabled: option.disabled,
                label: <PlaybackChoiceLabel choice={option} selected={value === option.value} />,
            }))}
            className={cn(
                "w-full",
                columns === "two" ? "max-w-2xl" : "max-w-5xl",
            )}
            stackClass={cn(
                "grid grid-cols-1 gap-4",
                columns === "two" ? "lg:grid-cols-2" : "xl:grid-cols-3",
            )}
            itemContainerClass={cn(
                "group/playback-choice block relative min-w-0 max-w-sm overflow-hidden rounded-xl border bg-white/70 p-0 transition-colors",
                "dark:bg-gray-950/30 border-gray-200 dark:border-gray-800",
                "hover:border-gray-300 dark:hover:border-gray-700 hover:bg-[--subtle]",
                "data-[state=checked]:border-[--border] data-[state=checked]:ring-1 data-[state=checked]:ring-white/30",
                "data-[state=checked]:bg-brand-50/40 dark:data-[state=checked]:bg-gray-800",
                "data-[disabled=true]:cursor-not-allowed data-[disabled=true]:opacity-55",
            )}
            itemIndicatorClass="hidden"
            itemCheckIconClass="hidden"
            itemClass="hidden"
            itemLabelClass="block w-full cursor-pointer"
        />
    )
}

function PlaybackChoiceLabel({ choice, selected }: { choice: PlaybackChoice, selected: boolean }) {
    const Icon = choice.icon

    return (
        <div className="flex flex-col h-full w-full">
            {choice.preview}
            <div className="flex items-start gap-3 p-4 flex-1">
                <div
                    className={cn(
                        "mt-0.5 flex h-9 w-9 flex-none items-center justify-center rounded-lg border bg-[--paper] text-[--muted]",
                        selected && "bg-brand-500/10 text-[--brand]",
                    )}
                >
                    <Icon className="text-lg" />
                </div>
                <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-0">
                        <p className="font-semibold leading-snug text-sm">{choice.title}</p>
                        {choice.badge}
                    </div>
                    <p className="mt-1 text-xs leading-normal text-[--muted]">{choice.description}</p>
                </div>
                {selected && (
                    <div className="mt-1 flex h-5 w-5 flex-none items-center justify-center rounded-full bg-brand text-white">
                        <LuCheck className="text-xs" />
                    </div>
                )}
            </div>
        </div>
    )
}

function OverrideNotice() {
    return null
}

function PreviewFrame({ children, className }: { children: React.ReactNode, className?: string }) {
    return (
        <div className={cn("relative h-40 aspect-[16/10]] overflow-hidden bg-gray-950 border-b border-gray-200 dark:border-gray-800/80", className)}>
            {children}
        </div>
    )
}

function MockLine({ w = "w-full", className }: { w?: string, className?: string }) {
    return <div className={cn("h-[4px] rounded-full bg-white/[0.12]", w, className)} />
}

function MockBlock({ className }: { className?: string }) {
    return <div className={cn("rounded-[3px] bg-white/[0.08]", className)} />
}

function DenshiBuiltInPreview({ activeEngine }: { activeEngine: string }) {
    return (
        <PreviewFrame>
            <div className="flex h-5 items-center gap-1.5 bg-gray-900 border-b border-white/5 px-2.5">
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <MockLine w="w-10" className="ml-1" />
            </div>
            <div className="flex h-[calc(100%-1.25rem)]">
                <div className="w-12 shrink-0 border-r border-white/5 bg-gray-900/50 p-2 space-y-2">
                    <MockBlock className="h-3.5 w-full" />
                    <MockLine w="w-full" />
                    <MockLine w="w-4/5" />
                    <MockLine w="w-3/5" />
                </div>
                <div className="flex-1 relative bg-gray-950">
                    <div className="absolute inset-0 flex items-center justify-center">
                        <div className="size-8 rounded-full bg-white/10 flex items-center justify-center">
                            <LuPlay className="text-[10px] text-white/50" />
                        </div>
                    </div>
                    <div className="absolute inset-x-3 bottom-3 space-y-1.5">
                        <div className="h-[3px] rounded-full bg-white/10 overflow-hidden">
                            <div className="h-full w-3/5 bg-white/25 rounded-full" />
                        </div>
                        <div className="flex justify-between">
                            <MockLine w="w-6" />
                            <MockLine w="w-6" />
                        </div>
                    </div>
                </div>
            </div>
        </PreviewFrame>
    )
}

function ContentRoutingPreview() {
    return (
        <PreviewFrame>
            <div className="flex h-5 items-center bg-gray-900 border-b border-white/5 px-2.5">
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <MockLine w="w-16" className="ml-1" />
            </div>
            <div className="p-3 space-y-2.5 h-[calc(100%-1.25rem)] bg-gray-950/30">
                {[["w-2/3", "w-1/2"], ["w-1/2", "w-1/3"], ["w-3/5", "w-2/5"]].map(([t, s], i) => (
                    <div key={i} className="flex items-center justify-between border-b border-white/5 pb-2 last:border-0 last:pb-0">
                        <div className="flex items-center gap-2">
                            <MockBlock className="size-4 shrink-0" />
                            <div className="space-y-1">
                                <MockLine w={t} />
                                <MockLine w={s} className="opacity-50" />
                            </div>
                        </div>
                        <div className={cn("h-3.5 w-7 rounded-full bg-white/10 p-0.5 transition-colors", i === 0 && "bg-white/25")} />
                    </div>
                ))}
            </div>
        </PreviewFrame>
    )
}

function VideoCorePreview() {
    return (
        <PreviewFrame className="bg-gray-950">
            <div className="flex h-5 items-center gap-1.5 bg-gray-900 border-b border-white/5 px-2.5">
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                {/*<MockLine w="w-12" className="ml-1" />*/}
            </div>
            <div className="relative h-[calc(100%-1.25rem)]">
                <div className="absolute inset-0 flex items-center justify-center text-[22px] font-semibold tracking-wider text-white/[0.1] select-none pointer-events-none font-mono">
                    HTML5
                </div>
                {/* <div className="absolute inset-0 flex items-center justify-center">
                 <div className="size-8 rounded-full bg-white/10 flex items-center justify-center">
                 <LuPlay className="text-[10px] text-white/50" />
                 </div>
                 </div> */}
                <div className="absolute inset-x-3 bottom-3 space-y-1.5">
                    <div className="h-[3px] rounded-full bg-white/10 overflow-hidden">
                        <div className="h-full w-2/5 bg-white/25 rounded-full" />
                    </div>
                    <div className="flex justify-between">
                        <MockLine w="w-6" />
                        <MockLine w="w-6" />
                    </div>
                </div>
            </div>
        </PreviewFrame>
    )
}

function MpvCorePreview() {
    return (
        <PreviewFrame className="bg-gray-950">
            <div className="flex h-5 items-center gap-1.5 bg-gray-900 border-b border-white/5 px-2.5">
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <span className="size-[4px] rounded-full bg-white/20" />
                <MockLine w="w-12" className="ml-1" />
            </div>
            <div className="relative h-[calc(100%-1.25rem)]">
                <div className="absolute inset-0 flex items-center justify-center text-[22px] font-semibold tracking-wider text-white/[0.1] select-none pointer-events-none font-mono">
                    MPV
                </div>
                {/* <div className="absolute right-2.5 top-2.5 w-20 rounded border border-white/10 bg-gray-900/95 p-1.5 space-y-1.5 shadow-lg z-10">
                 <MockLine w="w-10" className="bg-white/20" />
                 <div className="h-px bg-white/5" />
                 <div className="flex items-center gap-1">
                 <span className="size-1 rounded-full bg-white/60" />
                 <MockLine w="w-10" />
                 </div>
                 <div className="flex items-center gap-1">
                 <span className="size-1 rounded-full bg-white/0" />
                 <MockLine w="w-8" className="opacity-50" />
                 </div>
                 </div>
                 <div className="absolute inset-0 flex items-center justify-center">
                 <div className="size-8 rounded-full bg-white/10 flex items-center justify-center">
                 <LuPlay className="text-[10px] text-white/50" />
                 </div>
                 </div> */}
                <div className="absolute inset-x-3 bottom-3 space-y-1.5">
                    <div className="h-[3px] rounded-full bg-white/10 overflow-hidden">
                        <div className="h-full w-2/5 bg-white/25 rounded-full" />
                    </div>
                    <div className="flex justify-between">
                        <MockLine w="w-6" />
                        <MockLine w="w-6" />
                    </div>
                </div>
            </div>
        </PreviewFrame>
    )
}

function DesktopPlayerPreview() {
    return (
        <PreviewFrame className="bg-gray-900 p-3 flex items-center justify-center w-full">
            <PreviewFrame className="bg-gray-950 w-full">
                <div className="flex h-5 items-center gap-1.5 bg-gray-900 border-b border-white/5 px-2.5">
                    <span className="size-[4px] rounded-full bg-white/20" />
                    <span className="size-[4px] rounded-full bg-white/20" />
                    <span className="size-[4px] rounded-full bg-white/20" />
                </div>
                <div className="relative h-[calc(100%-1.25rem)]">
                </div>
            </PreviewFrame>
            <div className="absolute w-[58%] aspect-video bg-gray-950 rounded border border-white/10 shadow-2xl overflow-hidden">
                <div className="flex h-4 items-center gap-1 bg-gray-900 border-b border-white/5 px-1.5">
                    <span className="size-[3px] rounded-full bg-white/20" />
                    <span className="size-[3px] rounded-full bg-white/20" />
                    <span className="size-[3px] rounded-full bg-white/20" />
                    <MockLine w="w-10" className="ml-1" />
                </div>
                <div className="relative h-[calc(100%-1rem)]">
                    <div className="absolute inset-0 flex items-center justify-center">
                        <div className="size-6 rounded-full bg-white/10 flex items-center justify-center">
                            <LuPlay className="text-[8px] text-white/50" />
                        </div>
                    </div>
                    <div className="absolute inset-x-2 bottom-1.5">
                        <div className="h-[2px] rounded-full bg-white/10 overflow-hidden">
                            <div className="h-full w-1/2 bg-white/25 rounded-full" />
                        </div>
                    </div>
                </div>
            </div>
        </PreviewFrame>
    )
}

function MediastreamPreview({ disabled }: { disabled: boolean }) {
    return (
        <PreviewFrame className={disabled ? "opacity-40" : undefined}>
            <div className="flex h-4 items-end gap-0.5 bg-gray-900 border-b border-white/5 px-2">
                <div className="h-3.5 w-12 rounded-t bg-gray-950 border-t border-x border-white/5 flex items-center justify-center" />
                <div className="h-3 w-8 rounded-t bg-white/5" />
            </div>
            <div className="flex h-5 items-center px-2 bg-gray-950 border-b border-white/5">
                <MockBlock className="h-3.5 flex-1 rounded" />
            </div>
            <div className="relative h-[calc(100%-2.25rem)] bg-gray-950">
                <div className="absolute inset-0 flex items-center justify-center">
                    <div className="size-8 rounded-full bg-white/10 flex items-center justify-center">
                        <LuPlay className="text-[10px] text-white/50" />
                    </div>
                </div>
                <div className="absolute inset-x-3 bottom-3 h-[3px] bg-white/10">
                    <div className="h-full w-1/3 bg-white/25" />
                </div>
            </div>
        </PreviewFrame>
    )
}

function ExternalLinkPreview() {
    return (
        <PreviewFrame className="bg-gray-950 flex items-center justify-center p-3">
            <div className="flex items-center justify-between w-full px-2">
                <div className="w-16 rounded border border-white/10 bg-gray-900 overflow-hidden shadow-md">
                    <div className="flex h-3 items-center gap-0.5 bg-gray-800 px-1 border-b border-white/5">
                        <span className="size-[2px] rounded-full bg-white/20" />
                        <span className="size-[2px] rounded-full bg-white/20" />
                        <span className="size-[2px] rounded-full bg-white/20" />
                    </div>
                    <div className="p-1 space-y-1 bg-gray-900/50">
                        <MockBlock className="h-2 w-full" />
                        <MockLine w="w-10" />
                        <MockLine w="w-6" className="opacity-50" />
                    </div>
                </div>

                <div className="flex-1 flex items-center justify-center px-1 relative">
                    <div className="w-full border-t border-dashed border-white/15" />
                    <div className="absolute size-6 rounded-full bg-gray-900 border border-white/10 flex items-center justify-center shadow">
                        <LuExternalLink className="text-[10px] text-white/60" />
                    </div>
                </div>

                <div className="w-16 rounded border border-white/10 bg-gray-900 overflow-hidden shadow-md">
                    <div className="flex h-3 items-center gap-0.5 bg-gray-800 px-1 border-b border-white/5">
                        <span className="size-[2px] rounded-full bg-white/20" />
                        <span className="size-[2px] rounded-full bg-white/20" />
                        <span className="size-[2px] rounded-full bg-white/20" />
                    </div>
                    <div className="p-1 bg-gray-950">
                        <div className="h-6 bg-gray-900 rounded border border-white/5 flex items-center justify-center">
                            <div className="size-3.5 rounded-full bg-white/10 flex items-center justify-center">
                                <LuPlay className="text-[5px] text-white/40" />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </PreviewFrame>
    )
}

function TorrentDesktopPreview() {
    return <DesktopPlayerPreview />
}

function TorrentExternalPreview() {
    return <ExternalLinkPreview />
}
