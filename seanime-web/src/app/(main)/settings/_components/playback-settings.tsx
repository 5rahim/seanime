import {
    PlaybackDownloadedMedia,
    playbackDownloadedMediaOptions,
    PlaybackTorrentStreaming,
    playbackTorrentStreamingOptions,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { __settings_tabAtom } from "@/app/(main)/settings/_components/settings-page.atoms"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { RadioGroup } from "@/components/ui/radio-group"
import { useSetAtom } from "jotai"
import React from "react"
import { MdOutlineDevices } from "react-icons/md"
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
    } = useCurrentDevicePlaybackSettings()

    const { activeOnDevice, setActiveOnDevice } = useMediastreamActiveOnDevice()
    const { externalPlayerLink } = useExternalPlayerLink()
    const setTab = useSetAtom(__settings_tabAtom)


    return (
        <>
            <div>
                <h3>
                    Playback
                </h3>
                <p className="text-[--muted]">
                    Configure how anime is played on this device.
                </p>
                <p className="text-[--muted]">
                    These settings do not apply to playlists.
                </p>
            </div>

            <p className="">
                Current client: {serverStatus?.clientDevice || "N/A"}, {serverStatus?.clientPlatform || "N/A"}.
            </p>

            {(!externalPlayerLink && (downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink || torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink)) && (
                <Alert
                    intent="alert" description={<>
                    External player link is not set. <Button
                    intent="white-link" className="h-5 px-1" onClick={() => {
                    setTab("external-player-link")
                }}
                >Set external player link</Button>
                </>}
                />
            )}

            <SettingsCard title="Downloaded media" description="Player to use for downloaded media.">

                {(downloadedMediaPlayback === PlaybackDownloadedMedia.Default) && (
                    <Alert
                        intent="info" description={<>
                        Using <span className="font-semibold">{(serverStatus?.mediastreamSettings?.transcodeEnabled && activeOnDevice)
                        ? "integrated player (media streaming)"
                        : "desktop media player"}</span> for downloaded media.
                    </>}
                    />
                )}
                {(downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink && !!externalPlayerLink) && (
                    <Alert
                        intent="info" description={<>
                        Using <span className="font-semibold">external player link</span> for downloaded media.
                    </>}
                    />
                )}

                <RadioGroup
                    // label="Downloaded media"
                    // help="Player to use for downloaded media."
                    value={downloadedMediaPlayback}
                    onValueChange={v => {
                        setDownloadedMediaPlayback(v)
                        toast.success("Playback settings updated")
                    }}
                    options={playbackDownloadedMediaOptions}
                    itemContainerClass={cn(
                        "items-start cursor-pointer transition border-transparent rounded-[--radius] p-3 w-full",
                        "bg-transparent dark:hover:bg-gray-900 dark:bg-transparent",
                        "data-[state=checked]:bg-brand-500/5 dark:data-[state=checked]:bg-gray-900",
                        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-transparent transition",
                        "dark:border dark:data-[state=checked]:border-[--border] data-[state=checked]:ring-offset-0",
                    )}
                    itemClass={cn(
                        "absolute top-2 right-2",
                    )}
                    itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer w-full"
                />

                <>
                    {serverStatus?.mediastreamSettings?.transcodeEnabled && <div className="flex gap-4 items-center rounded-[--radius-md]">
                        <MdOutlineDevices className="text-4xl" />
                        <div className="space-y-1">
                            <Checkbox
                                value={activeOnDevice ?? false}
                                onValueChange={v => {
                                    setActiveOnDevice((prev) => typeof v === "boolean" ? v : prev)
                                    if (v) {
                                        toast.success("Media streaming is now active on this device.")
                                    } else {
                                        toast.info("Media streaming is now inactive on this device.")
                                    }
                                }}
                                label="Use media streaming on this device"
                            />
                        </div>
                    </div>}
                </>
            </SettingsCard>

            <SettingsCard title="Torrent/Debrid streaming" description="Player to use for torrent and debrid streaming.">

                <Alert
                    intent="info" description={<>
                    Using <span className="font-semibold">{(torrentStreamingPlayback === PlaybackTorrentStreaming.ExternalPlayerLink)
                    ? "external player link"
                    : "desktop media player"}</span> for torrent/debrid streaming.
                </>}
                />

                <RadioGroup
                    // name="-"
                    // label="Torrent/Debrid streaming"
                    // help="Player to use for torrent or debrid streaming."
                    value={torrentStreamingPlayback}
                    onValueChange={v => {
                        setTorrentStreamingPlayback(v)
                        toast.success("Playback settings updated")
                    }}
                    options={playbackTorrentStreamingOptions}
                    itemContainerClass={cn(
                        "items-start cursor-pointer transition border-transparent rounded-[--radius] p-3 w-full",
                        "bg-transparent dark:hover:bg-gray-900 dark:bg-transparent",
                        "data-[state=checked]:bg-brand-500/5 dark:data-[state=checked]:bg-gray-900",
                        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-transparent transition",
                        "dark:border dark:data-[state=checked]:border-[--border] data-[state=checked]:ring-offset-0",
                    )}
                    itemClass={cn(
                        "absolute top-2 right-2",
                    )}
                    itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer w-full"
                />

            </SettingsCard>

            <p className="italic text-sm text-[--muted]">
                Changes are saved automatically.
            </p>

        </>
    )
}
