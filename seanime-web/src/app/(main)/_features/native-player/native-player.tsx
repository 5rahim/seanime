import { getServerBaseUrl } from "@/api/client/server-url"
import { NativePlayer_PlaybackInfo, NativePlayer_ServerEvent } from "@/api/generated/types"
import {
    __seaMediaPlayer_autoNextAtom,
    __seaMediaPlayer_autoPlayAtom,
    __seaMediaPlayer_autoSkipIntroOutroAtom,
    __seaMediaPlayer_discreteControlsAtom,
    __seaMediaPlayer_mutedAtom,
    __seaMediaPlayer_volumeAtom,
} from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { submenuClass, VdsSubmenuButton } from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { clientIdAtom } from "@/app/websocket-provider"
import { LuffyError } from "@/components/shared/luffy-error"
import { vidstackLayoutIcons } from "@/components/shared/vidstack"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { __isDesktop__ } from "@/types/constants"
import {
    isHLSProvider,
    MediaCanPlayDetail,
    MediaCanPlayEvent,
    MediaDurationChangeEvent,
    MediaEndedEvent,
    MediaEnterFullscreenRequestEvent,
    MediaErrorDetail,
    MediaErrorEvent,
    MediaFullscreenRequestTarget,
    MediaPauseRequestEvent,
    MediaPlayer,
    MediaPlayerInstance,
    MediaPlayRequestEvent,
    MediaProvider,
    MediaProviderAdapter,
    MediaProviderChangeEvent,
    MediaProviderSetupEvent,
    MediaSeekRequestEvent,
    MediaTimeUpdateEvent,
    MediaTimeUpdateEventDetail,
    Menu,
} from "@vidstack/react"
import { DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import HLS from "hls.js"
import { useAtom, useAtomValue } from "jotai"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { BiExpand } from "react-icons/bi"
import { MdPlaylistPlay } from "react-icons/md"
import { PiSpinnerDuotone } from "react-icons/pi"
import { RxSlider } from "react-icons/rx"
import { useWebsocketMessageListener } from "../../_hooks/handle-websockets"
import { NativePlayerDrawer } from "./native-player-drawer"
import { nativePlayer_stateAtom } from "./native-player.atoms"

const log = logger("NATIVE PLAYER")


export function NativePlayer() {
    const clientId = useAtomValue(clientIdAtom)
    //
    // Player
    //
    // The player reference
    const playerRef = React.useRef<MediaPlayerInstance | null>(null)

    //
    // Control settings
    //
    const autoPlay = useAtomValue(__seaMediaPlayer_autoPlayAtom)
    const autoNext = useAtomValue(__seaMediaPlayer_autoNextAtom)
    const discreteControls = useAtomValue(__seaMediaPlayer_discreteControlsAtom)
    const autoSkipIntroOutro = useAtomValue(__seaMediaPlayer_autoSkipIntroOutroAtom)
    const [volume, setVolume] = useAtom(__seaMediaPlayer_volumeAtom)
    const [muted, setMuted] = useAtom(__seaMediaPlayer_mutedAtom)

    // The state
    const [state, setState] = useAtom(nativePlayer_stateAtom)

    //
    // Start
    //


    // Clean up player when unmounting or changing streams
    React.useEffect(() => {
        return () => {
            if (playerRef.current) {
                log.info("Cleaning up player")
                playerRef.current.destroy()
                playerRef.current = null
            }
        }
    }, [state.playbackInfo?.streamUrl])

    const onProviderSetup = (detail: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) => {
        log.info("Provider setup", detail, nativeEvent)

        // Reset any previous HLS instance
        if (isHLSProvider(detail) && detail.instance) {
            log.info("Destroying previous HLS instance")
            detail.instance.destroy()
            detail.library = HLS
        }
    }

    const onProviderChange = (detail: MediaProviderAdapter | null, nativeEvent: MediaProviderChangeEvent) => {
        log.info("Provider change", detail, nativeEvent)

        // Clean up previous provider if it exists
        if (detail && isHLSProvider(detail) && detail.instance) {
            log.info("Destroying previous HLS provider")
            detail.instance.destroy()
            detail.library = HLS
        }
    }

    const onTimeUpdate = (detail: MediaTimeUpdateEventDetail, e: MediaTimeUpdateEvent) => {
        // log.info("Time update", detail, e)
    }

    const onDurationChange = (detail: number, nativeEvent: MediaDurationChangeEvent) => {
        log.info("Duration change", detail, nativeEvent)
    }

    const onCanPlay = (detail: MediaCanPlayDetail, nativeEvent: MediaCanPlayEvent) => {
        log.info("Can play", detail, nativeEvent)

        log.info("Audio tracks", playerRef.current?.audioTracks?.toArray())
        log.info("Subtitle tracks", playerRef.current?.textTracks?.toArray())
    }

    const onEnded = (nativeEvent: MediaEndedEvent) => {
        log.info("Ended", nativeEvent)
    }

    const onMediaEnterFullscreenRequest = (detail: MediaFullscreenRequestTarget, nativeEvent: MediaEnterFullscreenRequestEvent) => {
        log.info("Media enter fullscreen request", detail, nativeEvent)
    }

    const onMediaPauseRequest = (nativeEvent: MediaPauseRequestEvent) => {
        log.info("Media pause request", nativeEvent)
    }

    const onMediaPlayRequest = (nativeEvent: MediaPlayRequestEvent) => {
        log.info("Media play request", nativeEvent)
    }

    const onMediaSeekRequest = (detail: number, nativeEvent: MediaSeekRequestEvent) => {
        log.info("Media seek request", detail, nativeEvent)
    }

    const onError = (detail: MediaErrorDetail, nativeEvent: MediaErrorEvent) => {
        log.info("Media error", detail, nativeEvent)
    }

    //
    // Server events
    //

    useWebsocketMessageListener({
        type: WSEvents.NATIVE_PLAYER,
        onMessage: ({ type, payload: _payload }: { type: NativePlayer_ServerEvent, payload: any }) => {
            log.info("Server event", type)
            switch (type) {
                // 1. Open and await
                // The server is loading the stream
                case "open-and-await":
                    log.info("Open and await event received")
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = _payload
                        draft.playbackInfo = null
                        draft.playbackError = null
                        return
                    })

                    break
                // 2. Watch
                // We receive the playback info
                case "watch":
                    const payload = _payload as NativePlayer_PlaybackInfo
                    log.info("Watch event received", payload)
                    setState(draft => {
                        draft.playbackInfo = payload
                        draft.loadingState = null
                        draft.playbackError = null
                        return
                    })
                    break
            }
        },
    })

    //
    // Handlers
    //

    function handleTerminateStream() {
        // Clean up player first
        if (playerRef.current) {
            log.info("Destroying player")
            playerRef.current.destroy()
            playerRef.current = null
        }

        setState(draft => {
            draft.active = false
            draft.miniPlayer = false
            draft.playbackInfo = null
            draft.playbackError = null
        })
        // Send terminate stream event
    }



    return (
        <>
            <NativePlayerDrawer
                open={state.active}
                onOpenChange={(v) => {
                    if (!v) {
                        setState(draft => {
                            if (!state.miniPlayer) {
                                draft.miniPlayer = true
                            } else {
                                handleTerminateStream()
                            }
                        })
                    }
                }}
                borderToBorder
                miniPlayer={state.miniPlayer}
                size={state.miniPlayer ? "md" : "full"}
                side={state.miniPlayer ? "right" : "bottom"}
                contentClass={cn(
                    "p-0 m-0",
                    !state.miniPlayer && "h-full",
                    // "h-full p-0 m-0 shadow-none bg-transparent backdrop-blur-sm transition-opacity duration-300",
                    // state.miniPlayer && "pointer-events-none",
                )}
                allowOutsideInteraction={true}
                overlayClass={cn(
                    state.miniPlayer && "hidden",
                )}
                closeClass={cn(
                    "z-[99]",
                    __isDesktop__ && !state.miniPlayer && "top-8",
                )}
                data-native-player-drawer
                // closeButton={
                //     <IconButton
                //         type="button"
                //         intent="gray-basic"
                //         size="sm"
                //         className={cn(
                //             "rounded-full text-2xl flex-none",
                //         )}
                //         icon={<BiX />}
                //     />
                // }
            >
                {state.miniPlayer && (
                    <IconButton
                        type="button"
                        intent="gray-basic"
                        size="sm"
                        className={cn(
                            "rounded-full text-2xl flex-none absolute z-[99] left-4 top-4 pointer-events-auto",
                        )}
                        icon={<BiExpand />}
                        onClick={() => {
                            setState(draft => {
                                draft.miniPlayer = false
                            })
                        }}
                    />
                )}
                <div className="h-full w-full bg-black flex items-center z-[50]" data-native-player-container data-mini-player={state.miniPlayer}>
                    {!!state.playbackError ? (
                        <LuffyError title="Playback Error">
                            <p>
                                {state.playbackError}
                            </p>
                        </LuffyError>
                    ) : (!!state.playbackInfo?.streamUrl && !state.loadingState) ? (
                        <MediaPlayer
                            data-sea-media-player
                            streamType="on-demand"
                            playsInline
                            ref={playerRef}
                            autoPlay={autoPlay}
                            crossOrigin
                            src={{
                                src: state.playbackInfo?.streamUrl?.replace("{{SERVER_URL}}", getServerBaseUrl()) + "?token=" + new Date().getTime(),
                                type: "video/webm",
                            }}
                            aspectRatio={undefined}
                            controlsDelay={discreteControls ? 500 : undefined}
                            className={cn(discreteControls && "discrete-controls")}
                            onProviderSetup={onProviderSetup}
                            onProviderChange={onProviderChange}
                            onMediaEnterFullscreenRequest={onMediaEnterFullscreenRequest}
                            onDurationChange={onDurationChange}
                            onTimeUpdate={onTimeUpdate}
                            onCanPlay={onCanPlay}
                            onEnded={onEnded}
                            onMediaPauseRequest={onMediaPauseRequest}
                            onMediaPlayRequest={onMediaPlayRequest}
                            onMediaSeekRequest={onMediaSeekRequest}
                            onError={onError}
                            volume={volume}
                            onVolumeChange={detail => setVolume(detail.volume)}
                            muted={muted}
                            onMediaMuteRequest={() => setMuted(true)}
                            onMediaUnmuteRequest={() => setMuted(false)}
                            style={{
                                border: "none",
                                width: "100%",
                                height: "100%",
                            }}
                        >
                            <MediaProvider>
                            </MediaProvider>
                            <DefaultVideoLayout
                                icons={vidstackLayoutIcons}
                                slots={{
                                    ...vidstackLayoutIcons,
                                    settingsMenuEndItems: <>
                                        <NativePlayerPlaybackSubmenu />
                                    </>,
                                }}
                            />
                        </MediaPlayer>
                    ) : (
                        <div
                            className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                        >
                            {/* <ParticleBackground className="absolute top-0 left-0 w-full h-full z-[0]" /> */}
                            <LoadingSpinner
                                title={state.loadingState || "Loading..."}
                                spinner={<PiSpinnerDuotone className="size-20 text-white animate-spin" />}
                            />
                        </div>
                    )}
                </div>
            </NativePlayerDrawer>
        </>
    )
}

export function NativePlayerPlaybackSubmenu() {

    const [autoPlay, setAutoPlay] = useAtom(__seaMediaPlayer_autoPlayAtom)
    const [autoNext, setAutoNext] = useAtom(__seaMediaPlayer_autoNextAtom)
    const [autoSkipIntroOutro, setAutoSkipIntroOutro] = useAtom(__seaMediaPlayer_autoSkipIntroOutroAtom)
    const [discreteControls, setDiscreteControls] = useAtom(__seaMediaPlayer_discreteControlsAtom)

    return (
        <>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Auto Play`}
                    hint={autoPlay ? "On" : "Off"}
                    disabled={false}
                    icon={AiFillPlayCircle}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Auto play"
                        fieldClass="py-2 px-2"
                        value={autoPlay}
                        onValueChange={setAutoPlay}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Auto Play Next Episode`}
                    hint={autoNext ? "On" : "Off"}
                    disabled={false}
                    icon={MdPlaylistPlay}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Auto play next episode"
                        fieldClass="py-2 px-2"
                        value={autoNext}
                        onValueChange={setAutoNext}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Skip Intro/Outro`}
                    hint={autoSkipIntroOutro ? "On" : "Off"}
                    disabled={false}
                    icon={MdPlaylistPlay}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Skip intro/outro"
                        fieldClass="py-2 px-2"
                        value={autoSkipIntroOutro}
                        onValueChange={setAutoSkipIntroOutro}
                    />
                </Menu.Content>
            </Menu.Root>
            <Menu.Root>
                <VdsSubmenuButton
                    label={`Discrete Controls`}
                    hint={discreteControls ? "On" : "Off"}
                    disabled={false}
                    icon={RxSlider}
                />
                <Menu.Content className={submenuClass}>
                    <Switch
                        label="Discrete controls"
                        help="Only show the controls when the mouse is over the bottom part. (Large screens only)"
                        fieldClass="py-2 px-2"
                        value={discreteControls}
                        onValueChange={setDiscreteControls}
                        fieldHelpTextClass="max-w-xs"
                    />
                </Menu.Content>
            </Menu.Root>
        </>
    )
}
