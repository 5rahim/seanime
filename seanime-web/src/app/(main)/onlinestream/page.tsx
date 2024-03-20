"use client"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { OnlinestreamProviderButton, OnlinestreamServerButton } from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { __onlinestream_selectedServerAtom } from "@/app/(main)/onlinestream/_lib/episodes"
import { OnlinestreamManagerProvider, useOnlinestreamManager } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Skeleton } from "@/components/ui/skeleton"
import {
    isHLSProvider,
    MediaPlayer,
    MediaPlayerInstance,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    Poster,
    Track,
} from "@vidstack/react"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS from "hls.js"
import { useAtom } from "jotai/react"
import Link from "next/link"
import { useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"

export default function Page() {

    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")

    const ref = React.useRef<MediaPlayerInstance>(null)


    const [episodeNumber, setEpisodeNumber] = React.useState(searchParams.get("episode") ? Number(searchParams.get("episode")) : 1)

    const [selectedServer, setSelectedServer] = useAtom(__onlinestream_selectedServerAtom)

    const {
        videoSource,
        currentEpisodeDetails,
        opts,
        url,
        onMediaDetached,
        onFatalError,
        loadPage,
        media,
        episodeSource,
    } = useOnlinestreamManager({
        mediaId,
    })

    function onProviderChange(
        provider: MediaProviderAdapter | null,
        nativeEvent: MediaProviderChangeEvent,
    ) {
        if (isHLSProvider(provider)) {
            provider.library = HLS
            provider.config = {
                // debug: true,
            }
        }
    }

    function onProviderSetup(provider: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) {
        console.table("onProviderSetup")
        if (isHLSProvider(provider)) {
            if (HLS.isSupported()) {
                provider.instance?.on(HLS.Events.MEDIA_DETACHED, (event) => {
                    onMediaDetached()
                })
                provider.instance?.on(HLS.Events.ERROR, (event, data) => {
                    if (data.fatal) {
                        onFatalError()
                    }
                })
            } else if (provider.video.canPlayType("application/vnd.apple.mpegurl")) {
                provider.video.src = url || ""
            }
        }
    }

    if (!loadPage) return null

    return (
        <PageWrapper className="p-4 sm:p-8 space-y-4">
            <OnlinestreamManagerProvider
                opts={opts}
            >

                <div className="col-span-1 2xl:col-span-full flex gap-4 items-center relative">
                    <Link href={`/entry?id=${media?.id}`}>
                        <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                    </Link>
                    <h3>{media.title?.userPreferred}</h3>
                </div>
                <div
                    className="space-y-4 grid lg:grid-cols-[1fr,500px] gap-4 lg:gap-8"
                >
                    <div>

                        <div className="aspect-video relative">
                            {!!url ? <MediaPlayer
                                ref={ref}
                                crossOrigin="anonymous"
                                src={{
                                    src: url || "",
                                    type: "application/x-mpegurl",
                                }}
                                onProviderChange={onProviderChange}
                                onProviderSetup={onProviderSetup}
                                className="w-full h-full absolute"
                            >
                                <MediaProvider>
                                    <Poster
                                        src={currentEpisodeDetails?.image || media.coverImage?.extraLarge || ""}
                                        alt="Episode"
                                    />
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
                                <DefaultVideoLayout
                                    icons={defaultLayoutIcons}
                                    slots={{
                                        beforeCaptionButton: (
                                            <div className="flex items-center">
                                                <OnlinestreamProviderButton />
                                                <OnlinestreamServerButton />
                                            </div>
                                        ),
                                    }}
                                />
                            </MediaPlayer> : (
                                <Skeleton className="h-full w-full absolute">
                                    <LoadingSpinner containerClass="h-full absolute" />
                                </Skeleton>
                            )}
                        </div>
                    </div>

                    <div>

                    </div>
                </div>
            </OnlinestreamManagerProvider>

        </PageWrapper>

    )

}
