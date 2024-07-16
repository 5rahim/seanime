import { Anime_AnimeEntry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __anime_torrentStreamingViewActiveAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai/react"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { PiMonitorPlayDuotone } from "react-icons/pi"

type TorrentStreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_AnimeEntry
}

export function TorrentStreamButton(props: TorrentStreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const [torrentStreamingActive, setTorrentStreamingActive] = useAtom(__anime_torrentStreamingViewActiveAtom)

    return (
        <>
            {serverStatus?.torrentstreamSettings?.enabled && (
                <Button
                    intent={torrentStreamingActive ? "alert-subtle" : "white-subtle"}
                    className="w-full"
                    size="md"
                    leftIcon={torrentStreamingActive ? <AiOutlineArrowLeft className="text-xl" /> : <PiMonitorPlayDuotone className="text-2xl" />}
                    onClick={() => setTorrentStreamingActive(p => !p)}
                >
                    {torrentStreamingActive ? "Close torrent streaming" : "Stream"}
                </Button>
            )}
        </>
    )
}
