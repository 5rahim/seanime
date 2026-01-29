import { Anime_Episode } from "@/api/generated/types"
import { useDirectstreamPlayLocalFile } from "@/api/hooks/directstream.hooks"
import { useNakamaPlayVideo } from "@/api/hooks/nakama.hooks"
import { usePlaybackPlayVideo, usePlaybackStartManualTracking } from "@/api/hooks/playback_manager.hooks"
import {
    ElectronPlaybackMethod,
    PlaybackDownloadedMedia,
    useCurrentDevicePlaybackSettings,
    useExternalPlayerLink,
} from "@/app/(main)/_atoms/playback.atoms"
import { useTorrentstreamAutoplay } from "@/app/(main)/_features/autoplay/autoplay"
import { __mpt_currentExternalPlayerLinkAtom } from "@/app/(main)/_features/progress-tracking/manual-progress-tracking"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useMediastreamActiveOnDevice, useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { ExternalPlayerLink } from "@/lib/external-player-link/external-player-link"
import { openTab } from "@/lib/helpers/browser"
import { logger } from "@/lib/helpers/debug"
import { __isElectronDesktop__ } from "@/types/constants"
import { useAtomValue, useSetAtom } from "jotai"
import { useRouter } from "next/navigation"
import React from "react"
import { toast } from "sonner"

export function useHandlePlayMedia() {
    const router = useRouter()
    const serverStatus = useServerStatus()
    const clientId = useAtomValue(clientIdAtom)

    const { activeOnDevice: mediastreamActiveOnDevice } = useMediastreamActiveOnDevice()
    const { setFilePath: setMediastreamFilePath } = useMediastreamCurrentFile()

    const { mutate: startManualTracking, isPending: isStarting } = usePlaybackStartManualTracking()
    const setCurrentExternalPlayerLink = useSetAtom(__mpt_currentExternalPlayerLinkAtom)

    const { downloadedMediaPlayback, electronPlaybackMethod } = useCurrentDevicePlaybackSettings()
    const { externalPlayerLink } = useExternalPlayerLink()

    // Play using desktop external player
    const { mutate: playVideo } = usePlaybackPlayVideo()
    const { mutate: playNakamaVideo } = useNakamaPlayVideo()

    const { mutate: directstreamPlayLocalFile } = useDirectstreamPlayLocalFile()

    const { setTorrentstreamAutoplayInfo } = useTorrentstreamAutoplay()

    const { getForcePlaybackMethod, resetForcePlaybackMethod } = useForcePlaybackMethod()

    function playMediaFile({
        path,
        mediaId,
        episode,
    }: {
        path: string,
        mediaId: number,
        episode: Anime_Episode
    }) {
        const anidbEpisode = episode.localFile?.metadata?.aniDBEpisode ?? ""

        const forcePlaybackMethod = getForcePlaybackMethod()
        resetForcePlaybackMethod()

        setTorrentstreamAutoplayInfo(null)

        if (episode._isNakamaEpisode) {
            // If external player link is set, open the media file in the external player
            if ((!forcePlaybackMethod && downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink) ||
                (forcePlaybackMethod && forcePlaybackMethod === "externalPlayerLink")
            ) {
                const link = new ExternalPlayerLink(externalPlayerLink)
                link.setEpisodeNumber(episode.progressNumber)
                link.setMediaTitle(episode.baseAnime?.title?.userPreferred)
                link.to({
                    endpoint: "/api/v1/nakama/stream?type=file&path=" + Buffer.from(path).toString("base64"),
                }).then()
                openTab(link.getFullUrl())
                setCurrentExternalPlayerLink(link.getFullUrl())

                if (episode?.progressNumber && episode.type === "main") {
                    logger("PLAY MEDIA").error("Starting manual tracking for nakama file")
                    // Start manual tracking
                    React.startTransition(() => {
                        startManualTracking({
                            mediaId: mediaId,
                            episodeNumber: episode?.progressNumber,
                            clientId: clientId || "",
                        })
                    })
                } else {
                    logger("PLAY MEDIA").warning("No manual tracking, progress number is not set for nakama file")
                }
                return
            }
            return playNakamaVideo({
                path,
                mediaId,
                anidbEpisode,
                clientId: clientId ?? "",
                forcePlaybackMethod: forcePlaybackMethod || undefined,
            })
        }

        logger("PLAY MEDIA").info("Playing media file", path)

        //
        // Electron native player
        //
        if (__isElectronDesktop__ && (
            (!forcePlaybackMethod && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer) ||
            (forcePlaybackMethod && forcePlaybackMethod === "nativeplayer")
        )) {
            directstreamPlayLocalFile({ path, clientId: clientId ?? "" })
            return
        }

        // If external player link is set, open the media file in the external player
        if ((!forcePlaybackMethod && downloadedMediaPlayback === PlaybackDownloadedMedia.ExternalPlayerLink) ||
            (forcePlaybackMethod && forcePlaybackMethod === "externalPlayerLink")
        ) {
            if (!externalPlayerLink) {
                toast.error("External player link is not set.")
                return
            }

            logger("PLAY MEDIA").info("Opening media file in external player", externalPlayerLink, path)

            setMediastreamFilePath(path)
            React.startTransition(() => {
                router.push(`/medialinks?id=${mediaId}`)
            })
            return
        }

        // Handle media streaming
        if (serverStatus?.mediastreamSettings?.transcodeEnabled && mediastreamActiveOnDevice) {
            setMediastreamFilePath(path)
            React.startTransition(() => {
                router.push(`/mediastream?id=${mediaId}`)
            })
            return
        }

        return playVideo({ path })
    }

    return {
        isUsingNativePlayer: __isElectronDesktop__ && electronPlaybackMethod === ElectronPlaybackMethod.NativePlayer,
        playMediaFile,
    }
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type ForcePlaybackMethod = "playbackmanager" | "nativeplayer" | "externalPlayerLink"

// maintain value outside react
const __forcePlaybackMethodStore = (() => {
    let current: ForcePlaybackMethod | undefined = undefined
    const listeners = new Set<() => void>()
    return {
        get: () => current,
        set: (val: ForcePlaybackMethod | undefined) => {
            current = val
            listeners.forEach(l => l())
        },
        subscribe: (l: () => void) => {
            listeners.add(l)
            return () => listeners.delete(l)
        },
    }
})()

// Returns the forced playback method, if any
export function useForcePlaybackMethod() {
    const queueRef = React.useRef<Array<{ method: ForcePlaybackMethod, cb?: () => void }>>([])
    const processingRef = React.useRef(false)

    const processQueue = React.useCallback(() => {
        if (processingRef.current) return
        if (queueRef.current.length === 0) return
        processingRef.current = true
        const { method, cb } = queueRef.current[0]
        __forcePlaybackMethodStore.set(method)
        Promise.resolve().then(() => {
            cb?.()
            // devnote: don't, this resets playback method before user selects a torrent
            // __forcePlaybackMethodStore.set(undefined)
            queueRef.current.shift()
            processingRef.current = false
            processQueue()
        })
    }, [])

    const forcePlaybackMethodFn = React.useCallback((method: ForcePlaybackMethod | undefined, cb?: () => void) => {
        if (!method) {
            cb?.()
            return
        }
        queueRef.current.push({ method, cb })
        processQueue()
    }, [processQueue])

    const getForcePlaybackMethod = React.useCallback(() => __forcePlaybackMethodStore.get(), [])

    const resetForcePlaybackMethod = React.useCallback(() => {
        queueRef.current = []
        processingRef.current = false
        __forcePlaybackMethodStore.set(undefined)
    }, [])

    return { forcePlaybackMethodFn, resetForcePlaybackMethod, getForcePlaybackMethod }
}

