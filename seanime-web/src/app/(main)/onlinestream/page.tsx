"use client"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import {
    __onlinestream_selectedServerAtom,
    useOnlinestreamEpisodeList,
    useOnlinestreamEpisodeSources,
    useOnlinestreamVideoSource,
} from "@/app/(main)/onlinestream/_lib/episodes"
import { IconButton } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import {
    isHLSProvider,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    Track,
} from "@vidstack/react"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS, { LoadPolicy } from "hls.js"
import { useAtom } from "jotai/react"
import { uniq } from "lodash"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { toast } from "sonner"


let hls: HLS | null = null

export default function Page() {

    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")

    const ref = React.useRef<MediaPlayerInstance>(null)

    const [erroredServers, setErroredServers] = React.useState<string[]>([])

    const [episodeNumber, setEpisodeNumber] = React.useState(searchParams.get("episode") ? Number(searchParams.get("episode")) : 1)

    const [selectedServer, setSelectedServer] = useAtom(__onlinestream_selectedServerAtom)

    const { episodes, media } = useOnlinestreamEpisodeList(mediaId)

    const { episodeSource } = useOnlinestreamEpisodeSources(mediaId)
    const episodeDetails = React.useMemo(() => {
        return episodes?.find((episode) => episode.number === episodeSource?.number)
    }, [episodes, episodeSource])

    const servers = React.useMemo(() => {
        if (!episodeSource) return []
        return uniq(episodeSource.videoSources.map((source) => source.server))
    }, [episodeSource])

    const { videoSource } = useOnlinestreamVideoSource(episodeSource)


    React.useEffect(() => {
        console.log("videoSource:", videoSource)
    }, [videoSource])

    function onProviderChange(
        provider: MediaProviderAdapter | null,
        nativeEvent: MediaProviderChangeEvent,
    ) {
        const loadPolicy: LoadPolicy = {
            default: {
                maxTimeToFirstByteMs: Infinity,
                maxLoadTimeMs: 60_000,
                timeoutRetry: {
                    maxNumRetry: 2,
                    retryDelayMs: 0,
                    maxRetryDelayMs: 0,
                },
                errorRetry: {
                    maxNumRetry: 5,
                    retryDelayMs: 0,
                    maxRetryDelayMs: 0,
                },
            },
        }
        if (isHLSProvider(provider)) {
            provider.library = HLS
            provider.config = {
                autoStartLoad: true,
                startLevel: Infinity,
                abrEwmaDefaultEstimate: 35_000_000,
                abrEwmaDefaultEstimateMax: 50_000_000,
                debug: true,
                lowLatencyMode: false,
                fragLoadPolicy: {
                    default: {
                        maxTimeToFirstByteMs: Infinity,
                        maxLoadTimeMs: 60_000,
                        timeoutRetry: {
                            maxNumRetry: 5,
                            retryDelayMs: 100,
                            maxRetryDelayMs: 0,
                        },
                        errorRetry: {
                            maxNumRetry: 5,
                            retryDelayMs: 0,
                            maxRetryDelayMs: 100,
                        },
                    },
                },
                keyLoadPolicy: loadPolicy,
                certLoadPolicy: loadPolicy,
                playlistLoadPolicy: loadPolicy,
                manifestLoadPolicy: loadPolicy,
                steeringManifestLoadPolicy: loadPolicy,
            }
        }
    }

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        if (isHLSProvider(provider)) {
            if (HLS.isSupported()) {
                provider.instance?.attachMedia(provider.video)
                provider.instance?.startLoad(0)
            }
            // } else if (provider.video.canPlayType("application/vnd.apple.mpegurl")) {
            //     provider.video.src = "http://192.168.1.151:43211/api/v1/stream3/master.m3u8"
            //
            // }
        }
    }

    const handlePlaybackError = React.useCallback(() => {
        setTimeout(() => {
            // if (videoSource?.server) {
            //     setErroredServers((prev) => [...prev, videoSource.server])
            //     const otherServers = servers.filter((server) => server !== videoSource?.server && !erroredServers.includes(server))
            //     if (otherServers.length > 0) {
            //         setSelectedServer(otherServers[0])
            //     }
            // }
            toast.error("Playback error, please choose another server")
        }, 2000)
    }, [servers, videoSource?.server, erroredServers])

    React.useEffect(() => {
        if (videoSource) {
            ref.current?.provider?.loadSource({
                src: videoSource.url,
                type: "application/x-mpegurl",
            })
            ref.current?.startLoading()
        }
    }, [videoSource])

    if (!media) return null

    return (
        <div>
            <div className="col-span-1 2xl:col-span-full flex gap-4 items-center relative">
                <Link href={`/entry?id=${media?.id}`}>
                    <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                </Link>
                <h3>{media.title?.userPreferred}</h3>
                <pre>
                {JSON.stringify(servers, null, 2)}
            </pre>
            </div>
            <div
                className="col-span-1 2xl:col-span-5 h-[fit-content] 2xl:sticky top-[5rem] space-y-4"
            >
                <div>
                    <Select
                        value={videoSource?.server || ""}
                        options={servers.map((server) => ({ label: server, value: server }))}
                        onValueChange={(v) => {
                            setSelectedServer(v)
                        }}
                    />
                    <MediaPlayer
                        ref={ref}
                        crossOrigin="anonymous"
                        poster={episodeDetails?.image || media.coverImage?.extraLarge || ""}
                        src={{
                            src: videoSource?.url || "",
                            type: "application/x-mpegurl",
                        }}
                        onProviderChange={onProviderChange}
                        onProviderSetup={onProviderSetup}
                        onHlsDestroying={(e) => {
                            console.log("onHlsDestroying", e)
                        }}
                    >
                        <MediaProvider>
                            {episodeSource?.subtitles?.map((sub) => {
                                return <Track
                                    key={sub.url}
                                    {...{
                                        id: sub.language,
                                        label: sub.language,
                                        kind: "subtitles",
                                        src: sub.url,
                                        language: sub.language,
                                        default: sub.language
                                            ? sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us"
                                            : sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us",
                                    }}
                                />
                            })}
                        </MediaProvider>
                        <DefaultVideoLayout icons={defaultLayoutIcons} />
                    </MediaPlayer>
                </div>
            </div>
        </div>

    )

}
