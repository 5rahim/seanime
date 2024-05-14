import { Anime_MediaEntry } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { __anime_wantStreamingAtom } from "@/app/(main)/entry/_containers/anime-entry-page"
import { Button } from "@/components/ui/button"
import { useAtom } from "jotai/react"
import React from "react"
import { PiMonitorPlayDuotone } from "react-icons/pi"

type TorrentStreamButtonProps = {
    children?: React.ReactNode
    entry: Anime_MediaEntry
}

export function TorrentStreamButton(props: TorrentStreamButtonProps) {

    const {
        children,
        entry,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const [wantStreaming, setWantStreaming] = useAtom(__anime_wantStreamingAtom)

    return (
        <>
            {serverStatus?.torrentstreamSettings?.enabled && (
                <Button
                    intent="white-outline"
                    className="w-full"
                    size="lg"
                    leftIcon={<PiMonitorPlayDuotone className="text-2xl" />}
                    onClick={() => setWantStreaming(p => !p)}
                >
                    {wantStreaming ? "Close" : "Stream"}
                </Button>
            )}
        </>
    )
}
