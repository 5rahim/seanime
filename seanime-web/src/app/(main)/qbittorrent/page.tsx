import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Button } from "@/components/ui/button"
import { __isElectronDesktop__ } from "@/types/constants"
import React from "react"
import { LuExternalLink } from "react-icons/lu"

export default function Page() {

    const status = useServerStatus()
    const settings = status?.settings

    const qbittorrentUrl = settings?.torrent
        ? `http://${settings.torrent.qbittorrentHost}:${String(settings.torrent.qbittorrentPort)}`
        : ""

    const [isAllowed, setIsAllowed] = React.useState(false)

    React.useEffect(() => {
        if (__isElectronDesktop__ && qbittorrentUrl) {
            window.electron?.localServer?.allowWebviewOrigin?.(qbittorrentUrl)
                .then(() => setIsAllowed(true))
                .catch(() => setIsAllowed(true))
        } else {
            setIsAllowed(true)
        }
    }, [qbittorrentUrl])

    if (!settings) return null

    return (
        <PageWrapper className="p-4 sm:p-6 lg:p-8 space-y-4">
            <header className="flex items-center justify-between">
                <div>
                    <h2>qBittorrent</h2>
                    <p className="text-[--muted]">Access the embedded qBittorrent client Web UI.</p>
                </div>
                <div className="flex items-center gap-2">
                    <a
                        href={qbittorrentUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        <Button
                            intent="gray-outline"
                            leftIcon={<LuExternalLink />}
                        >
                            Open in browser
                        </Button>
                    </a>
                </div>
            </header>

            <div
                className="w-full h-[calc(100vh-16rem)] rounded-xl border border-[--border] overflow-hidden ring-1 ring-[--border] ring-offset-2 ring-offset-[--background]"
            >
                {isAllowed && (
                    __isElectronDesktop__ ? (
                        <webview
                            src={qbittorrentUrl}
                            style={{ width: "100%", height: "100%" }}
                            {...({ allowpopups: "true" } as any)}
                        />
                    ) : (
                        <iframe
                            src={qbittorrentUrl}
                            className="w-full h-full"
                            sandbox="allow-forms allow-fullscreen allow-same-origin allow-scripts allow-popups"
                            referrerPolicy="no-referrer"
                            {...({ credentialless: "true" } as any)}
                        />
                    )
                )}
            </div>
        </PageWrapper>
    )
}

