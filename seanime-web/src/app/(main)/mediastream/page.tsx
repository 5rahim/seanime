"use client"

import { useHandleMediastream } from "@/app/(main)/mediastream/_lib/handle-mediastream"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Skeleton } from "@/components/ui/skeleton"
import { getDirectPlayProfiles } from "@/lib/utils/playback-profiles/directplay-profile"
import { getCodecProfiles } from "@/lib/utils/playback-profiles/helpers/codec-profiles"
import { hasH264Support } from "@/lib/utils/playback-profiles/helpers/mp4-video-formats"
import { canPlayNativeHls, hasMkvSupport } from "@/lib/utils/playback-profiles/helpers/transcoding-formats"
import { MediaPlayer, MediaPlayerInstance, MediaProvider, Track } from "@vidstack/react"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import { CaptionsFileFormat } from "media-captions"
import React from "react"


export default function Page() {

    const playerRef = React.useRef<MediaPlayerInstance>(null)

    const oldHls = React.useRef<string | null>(null)

    const {
        url,
        isError,
        isMediaContainerLoading,
        streamType,
        subtitles,
        subtitleEndpointUri,
        onProviderChange,
        onProviderSetup,
    } = useHandleMediastream({ playerRef })

    const videoRef = React.useRef<HTMLVideoElement>(null)
    React.useEffect(() => {
        if (videoRef.current) {
            console.log(hasMkvSupport(videoRef.current!))
            console.log(canPlayNativeHls(videoRef.current!))
            console.log(hasH264Support(videoRef.current!))
            console.log(getCodecProfiles(videoRef.current!))
            console.log(getDirectPlayProfiles(videoRef.current!))
        }
    }, [])

    return (
        <AppLayoutStack className="p-8">
            <h3>Streaming</h3>

            <video ref={videoRef} />

            <div
                className={cn(
                    "aspect-video relative lg:max-w-[70%]",
                )}
            >
                {isError ?
                    <LuffyError title="Failed to load media" /> :
                    (!!url && !isMediaContainerLoading) ? <MediaPlayer
                        ref={playerRef}
                        crossOrigin
                        src={url}
                        // poster={currentEpisodeDetails?.image || media.coverImage?.extraLarge || ""}
                        onProviderChange={onProviderChange}
                        onProviderSetup={onProviderSetup}
                        onTimeUpdate={(e) => {

                        }}
                        onEnded={(e) => {

                        }}
                        onCanPlay={(e) => {

                        }}
                    >
                        <MediaProvider>
                            {subtitles?.map((sub) => (
                                <Track
                                    key={String(sub.index)}
                                    src={subtitleEndpointUri + sub.link}
                                    label={sub.title || sub.language}
                                    lang={sub.language}
                                    type={(sub.extension?.replace(".", "") || "ass") as CaptionsFileFormat}
                                    kind="subtitles"
                                    default={sub.isDefault}
                                />
                            ))}
                        </MediaProvider>
                        {/*<div className="absolute bottom-24 px-4 w-full justify-between flex items-center">*/}
                        {/*    <div>*/}
                        {/*        {(showSkipIntroButton) && (*/}
                        {/*            <Button intent="white" onClick={() => seekTo(aniSkipData?.op?.interval?.endTime || 0)}>Skip*/}
                        {/*                                                                                                   intro</Button>*/}
                        {/*        )}*/}
                        {/*    </div>*/}
                        {/*    <div>*/}
                        {/*        {(showSkipEndingButton) && (*/}
                        {/*            <Button intent="white" onClick={() => seekTo(aniSkipData?.ed?.interval?.endTime || 0)}>Skip*/}
                        {/*                                                                                                   ending</Button>*/}
                        {/*        )}*/}
                        {/*    </div>*/}
                        {/*</div>*/}
                        <DefaultVideoLayout
                            icons={defaultLayoutIcons}
                        />
                    </MediaPlayer> : (
                        <Skeleton className="h-full w-full absolute">
                            <LoadingSpinner containerClass="h-full absolute" />
                        </Skeleton>
                    )}
            </div>

            {/*{(mediaContainer && url) && <MediaPlayer*/}
            {/*    ref={playerRef}*/}
            {/*    crossOrigin*/}
            {/*    src={url}*/}
            {/*    onProviderChange={onProviderChange}*/}
            {/*    onProviderSetup={onProviderSetup}*/}
            {/*    onCanPlay={onCanPlay}*/}
            {/*>*/}
            {/*    <MediaProvider>*/}
            {/*        {mediaContainer?.mediaInfo?.subtitles?.map((sub) => (*/}
            {/*            <Track*/}
            {/*                key={String(sub.index)}*/}
            {/*                src={`http://192.168.1.151:${__DEV_SERVER_PORT}/api/v1/mediastream/transcode-subs` + sub.link}*/}
            {/*                label={sub.title}*/}
            {/*                lang={sub.language}*/}
            {/*                type={"ass"}*/}
            {/*                kind="subtitles"*/}
            {/*                default={sub.isDefault}*/}
            {/*            />*/}
            {/*        ))}*/}
            {/*    </MediaProvider>*/}
            {/*    <DefaultVideoLayout*/}
            {/*        // thumbnails="https://image.mux.com/VZtzUzGRv02OhRnZCxcNg49OilvolTqdnFLEqBsTwaxU/storyboard.vtt"*/}
            {/*        icons={defaultLayoutIcons}*/}
            {/*    />*/}
            {/*</MediaPlayer>}*/}
        </AppLayoutStack>
    )

}
