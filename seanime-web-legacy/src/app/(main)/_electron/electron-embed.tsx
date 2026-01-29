import { __isElectronDesktop__ } from "@/types/constants"
import React from "react"

export function useElectronYoutubeEmbed() {
    const [localServerPort, setLocalServerPort] = React.useState<number | null>(null)

    React.useEffect(() => {
        (async () => {
            if (__isElectronDesktop__) {
                setLocalServerPort(await window.electron?.localServer?.getPort?.() ?? null)
            }
        })()
    }, [])

    return {
        electronEmbedAddress: localServerPort ? `http://localhost:${localServerPort}/player/` : null,
    }
}

export function ElectronYoutubeEmbed({ trailerId, isCompact, isBanner, ...props }: {
    isCompact?: boolean,
    isBanner?: boolean,
    trailerId: string | null | undefined
} & React.HTMLAttributes<HTMLWebViewElement>) {
    const { electronEmbedAddress } = useElectronYoutubeEmbed()

    const webviewRef = React.useRef<HTMLWebViewElement | null>(null)

    React.useEffect(() => {
        const webview = webviewRef.current
        if (!webview) return

        const handleFinishLoad = (e: any) => {
            props.onLoad?.(e)
        }

        const handleFailLoad = (e: any) => {
            props.onError?.(e)
        }

        // const handleDomReady = () => {
        //     console.log("dom ready")
        //     // we can inject JS if needed:
        //     // webview.executeJavaScript("console.log('Hello from inside webview')")
        // }

        webview.addEventListener("did-finish-load", handleFinishLoad)
        webview.addEventListener("did-fail-load", handleFailLoad)
        // webview.addEventListener("dom-ready", handleDomReady)

        return () => {
            webview.removeEventListener("did-finish-load", handleFinishLoad)
            webview.removeEventListener("did-fail-load", handleFailLoad)
            // webview.removeEventListener("dom-ready", handleDomReady)
        }
    }, [electronEmbedAddress, trailerId])

    if (!electronEmbedAddress) return null

    return <webview
        ref={webviewRef}
        src={`${electronEmbedAddress}${isCompact ? "compact_" : isBanner ? "banner_" : ""}${trailerId}`}
        style={(!isCompact || isBanner) ? { width: "100%", height: "100%" } : {
            width: "180%",
            height: "180%",
            translate: "-25% -20%",
        }}
        allowFullScreen
    />
}
