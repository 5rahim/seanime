"use client"

import { AppLayoutStack } from "@/components/ui/app-layout"
import { __DEV_SERVER_PORT } from "@/lib/server/config"
import {
    isHLSProvider,
    LibASSTextRenderer,
    MediaCanPlayDetail,
    MediaCanPlayEvent,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
} from "@vidstack/react"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS, { LoadPolicy } from "hls.js"
import React from "react"

let hls: HLS | null = null

function uuidv4(): string {
    // @ts-ignore
    return ([1e7] + -1e3 + -4e3 + -8e3 + -1e11).replace(/[018]/g, (c) =>
        (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16),
    )
}

let client_id = typeof window === "undefined" ? "ssr" : uuidv4()


export default function Page() {

    const ref = React.useRef<MediaPlayerInstance>(null)

    const oldHls = React.useRef<string | null>(null)


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
                    maxNumRetry: 1,
                    retryDelayMs: 0,
                    maxRetryDelayMs: 0,
                },
            },
        }
        if (isHLSProvider(provider)) {
            // Static import
            // provider.library = HLS
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

    // React.useEffect(() => {
    //     ref.current?.provider?.play()
    //     // ref.current?.startLoading()
    // }, [])

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {

        if (isHLSProvider(provider)) {
            if (HLS.isSupported()) {
                provider.instance?.loadSource("http://192.168.1.151:43211/api/v1/stream/master.m3u8")
                provider.instance?.attachMedia(provider.video)
                provider.instance?.startLoad(0)
            }
            // } else if (provider.video.canPlayType("application/vnd.apple.mpegurl")) {
            //     provider.video.src = "http://192.168.1.151:43211/api/v1/stream3/master.m3u8"
            //
            // }
        }
    }

    // We can listen for the `can-play` event to be notified when the player is ready.
    function onCanPlay(detail: MediaCanPlayDetail, nativeEvent: MediaCanPlayEvent) {
        console.log("Can play", detail)
    }

    React.useEffect(() => {
        // @ts-ignore
        const renderer = new LibASSTextRenderer(() => import("jassub"), {
            workerUrl: "/jassub/jassub-worker.js",
            legacyWorkerUrl: "/jassub/jassub-worker-legacy.js",
        })

        ref.current!.textRenderers.add(renderer)
    }, [])

    return (
        <AppLayoutStack className="p-8">
            <h3>Streaming</h3>

            {/*<ArtPlayer url="http://192.168.1.151:43211/api/v1/stream2/test.m3u8" />*/}

            <MediaPlayer
                ref={ref}
                crossOrigin
                src={`http://192.168.1.151:${__DEV_SERVER_PORT}/api/v1/stream/master.m3u8`}
                onProviderChange={onProviderChange}
                onProviderSetup={onProviderSetup}
                onCanPlay={onCanPlay}
            >
                <MediaProvider>
                    {/*<Track*/}
                    {/*    src="http://192.168.1.151:43211/api/v1/stream2/english.ass"*/}
                    {/*    kind="subtitles"*/}
                    {/*    type="ass"*/}
                    {/*    label="English"*/}
                    {/*    lang="en-US"*/}
                    {/*    default*/}
                    {/*/>*/}
                </MediaProvider>
                <DefaultVideoLayout
                    // thumbnails="https://image.mux.com/VZtzUzGRv02OhRnZCxcNg49OilvolTqdnFLEqBsTwaxU/storyboard.vtt"
                    icons={defaultLayoutIcons}
                />
            </MediaPlayer>
        </AppLayoutStack>
    )

}
