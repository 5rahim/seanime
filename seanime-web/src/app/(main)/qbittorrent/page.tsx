"use client"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"

export const dynamic = "force-static"

export default function Page() {

    const status = useServerStatus()
    const settings = status?.settings

    if (!settings) return null

    return (
        <>
            <div
                className="w-[80%] h-[calc(100vh-15rem)] rounded-xl border  overflow-hidden mx-auto mt-10 ring-1 ring-[--border] ring-offset-2"
            >
                <iframe
                    src={`http://${settings.torrent?.qbittorrentHost}:${String(settings.torrent?.qbittorrentPort)}`}
                    className="w-full h-full"
                    sandbox="allow-forms allow-fullscreen allow-same-origin allow-scripts allow-popups"
                    referrerPolicy="no-referrer"
                />
            </div>
        </>
    )
}
