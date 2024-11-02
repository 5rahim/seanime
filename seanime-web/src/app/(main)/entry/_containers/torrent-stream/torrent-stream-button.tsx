import { Anime_Entry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useAnimeEntryPageView } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { PiMonitorPlayDuotone } from "react-icons/pi"

type TorrentStreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_Entry
}

export function TorrentStreamButton(props: TorrentStreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const { isLibraryView, isTorrentStreamingView, toggleTorrentStreamingView } = useAnimeEntryPageView()

    if (
        !entry ||
        entry.media?.status === "NOT_YET_RELEASED" ||
        !serverStatus?.torrentstreamSettings?.enabled
    ) return null

    if (!isLibraryView && !isTorrentStreamingView) return null

    return (
        <>
            <Button
                intent={isTorrentStreamingView ? "alert-subtle" : "white-subtle"}
                className="w-full"
                size="md"
                leftIcon={isTorrentStreamingView ? <AiOutlineArrowLeft className="text-xl" /> : <PiMonitorPlayDuotone className="text-2xl" />}
                onClick={() => toggleTorrentStreamingView()}
            >
                {isTorrentStreamingView ? "Close torrent streaming" : "Stream"}
            </Button>
        </>
    )
}
