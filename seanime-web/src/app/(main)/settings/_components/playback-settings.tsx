import { playbackDownloadedMediaOptions, playbackTorrentStreamingOptions, useCurrentDevicePlaybackSettings } from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Alert } from "@/components/ui/alert"
import { Select } from "@/components/ui/select"
import React from "react"

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

    return (
        <>
            <div>
                <h3>
                    Playback
                </h3>
                <p className="text-[--muted]">
                    Configure how media files are played on this device only.
                </p>
                <p className="text-[--muted]">
                    These settings do not apply to playlists.
                </p>
            </div>

            <Alert
                intent="info" description={<>
                Current client: {serverStatus?.clientDevice || "N/A"}, {serverStatus?.clientPlatform || "N/A"}.
            </>}
            />

            <Select
                label="Downloaded media"
                help="Player to use for downloaded media."
                value={downloadedMediaPlayback}
                onValueChange={v => setDownloadedMediaPlayback(v)}
                disabled={!serverStatus?.mediastreamSettings?.transcodeEnabled}
                options={playbackDownloadedMediaOptions}
            />

            <Select
                name="-"
                label="Torrent streaming"
                help="Player to use for torrent streaming."
                value={torrentStreamingPlayback}
                onValueChange={v => setTorrentStreamingPlayback(v)}
                disabled={!serverStatus?.torrentstreamSettings?.enabled}
                options={playbackTorrentStreamingOptions}
            />

            <br />

            <p className="italic text-sm text-[--muted]">
                Changes are saved automatically.
            </p>

        </>
    )
}
