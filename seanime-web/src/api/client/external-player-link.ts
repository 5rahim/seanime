import { __isDesktop__ } from "@/types/constants"

export function getExternalPlayerURL(externalPlayerLink: string, url: string): string {
    if (externalPlayerLink.includes("{scheme}")) {
        return externalPlayerLink.replace("{scheme}", window.location.protocol.replace(":", ""))
    }

    if (__isDesktop__) {
        let retUrl = externalPlayerLink.replace("{url}", url)
        if (externalPlayerLink.startsWith("intent://")) {
            // e.g. "intent://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
            retUrl = retUrl.replace("intent://http://", "intent://").replace("intent://https://", "intent://")
        }
        return retUrl
    }

    // e.g. "mpv://http://localhost:43214/stream/..."
    // e.g. "intent://http://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
    let retUrl = externalPlayerLink.replace("{url}", url)
        .replace("127.0.0.1", window.location.hostname)
        .replace("localhost", window.location.hostname)


    if (externalPlayerLink.startsWith("intent://")) {
        // e.g. "intent://localhost:43214/stream/...#Intent;package=org.videolan.vlc;scheme=http;end"
        retUrl = retUrl.replace("intent://http://", "intent://").replace("intent://https://", "intent://")
    }

    return retUrl
}
