"use client"
import { useAtomValue } from "jotai"
import { serverStatusAtom } from "@/atoms/server-status"

export default function Page() {

    const status = useAtomValue(serverStatusAtom)
    const settings = status?.settings

    if (!settings) return null

    return (
        <>
            <div
                className={"w-[80%] h-[calc(100vh-15rem)] rounded-xl border border-[--border] overflow-hidden mx-auto mt-10 ring-1 ring-[--border] ring-offset-2"}>
                <iframe
                    src={`http://${settings.torrent?.qbittorrentHost}:${String(settings.torrent?.qbittorrentPort)}`}
                    className={"w-full h-full"}
                />
            </div>
        </>
    )
}