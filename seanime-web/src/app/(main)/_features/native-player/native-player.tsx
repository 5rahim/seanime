import {
    __seaMediaPlayer_autoNextAtom,
    __seaMediaPlayer_autoPlayAtom,
    __seaMediaPlayer_autoSkipIntroOutroAtom,
    __seaMediaPlayer_discreteControlsAtom,
    __seaMediaPlayer_mutedAtom,
    __seaMediaPlayer_volumeAtom,
} from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import { submenuClass, VdsSubmenuButton } from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { LuffyError } from "@/components/shared/luffy-error"
import { vidstackLayoutIcons } from "@/components/shared/vidstack"
import { cn } from "@/components/ui/core/styling"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Switch } from "@/components/ui/switch"
import { logger } from "@/lib/helpers/debug"
import {
    MediaCanPlayDetail,
    MediaCanPlayEvent,
    MediaDurationChangeEvent,
    MediaEndedEvent,
    MediaEnterFullscreenRequestEvent,
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
import { useAtom, useAtomValue } from "jotai"
import Image from "next/image"
import React from "react"
import { AiFillPlayCircle } from "react-icons/ai"
import { MdPlaylistPlay } from "react-icons/md"
import { RxSlider } from "react-icons/rx"
import { nativePlayer_openAtom } from "./native-player.atoms"

const log = logger("NATIVE PLAYER")

export function NativePlayer() {
    const [open, setOpen] = useAtom(nativePlayer_openAtom)

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

    // The playback error state
    const [playbackError, setPlaybackError] = React.useState<string | null>(null)
    // The url of the media to be played
    const [url, setUrl] = React.useState<string | null>(null)
    // The loading state
    const [loadingState, setLoadingState] = React.useState<string | null>(null)

    //
    // Callbacks
    //

    const onProviderSetup = (detail: MediaProviderAdapter, nativeEvent: MediaProviderSetupEvent) => {
        log.info("Provider setup", detail, nativeEvent)
    }

    const onProviderChange = (detail: MediaProviderAdapter | null, nativeEvent: MediaProviderChangeEvent) => {
        log.info("Provider change", detail, nativeEvent)
    }

    const onTimeUpdate = (detail: MediaTimeUpdateEventDetail, e: MediaTimeUpdateEvent) => {
        // log.info("Time update", detail, e)
    }

    const onDurationChange = (detail: number, nativeEvent: MediaDurationChangeEvent) => {
        log.info("Duration change", detail, nativeEvent)
    }

    const onCanPlay = (detail: MediaCanPlayDetail, nativeEvent: MediaCanPlayEvent) => {
        log.info("Can play", detail, nativeEvent)
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

    return (
        <Drawer
            open={open}
            onOpenChange={setOpen}
            borderToBorder
            size="full"
            side="bottom"
            contentClass="h-full p-0 m-0 shadow-none bg-transparent backdrop-blur-sm"
            // hideCloseButton
            data-native-player-drawer
        >
            <div className="h-full w-full" data-native-player-container>
                {!!playbackError ? (
                    <LuffyError title="Playback Error">
                        <p>
                            {playbackError}
                        </p>
                    </LuffyError>
                ) : (!!url && !loadingState) ? (
                    <MediaPlayer
                        data-sea-media-player
                        streamType="on-demand"
                        playsInline
                        ref={playerRef}
                        autoPlay={autoPlay}
                        crossOrigin
                        src={url}
                        aspectRatio="16/9"
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
                        // onFullscreenChange={(isFullscreen: boolean, event: MediaFullscreenChangeEvent) => {
                        //     if (isFullscreen) {
                        //         // Store the currently focused element
                        //         lastFocusedElementRef.current = document.activeElement as HTMLElement
                        //     } else {
                        //         // Restore focus
                        //         setTimeout(() => {
                        //             lastFocusedElementRef.current?.focus()
                        //         }, 100)
                        //     }
                        // }}
                        volume={volume}
                        onVolumeChange={detail => setVolume(detail.volume)}
                        muted={muted}
                        onMediaMuteRequest={() => setMuted(true)}
                        onMediaUnmuteRequest={() => setMuted(false)}
                    >
                        <MediaProvider>
                            {/*{tracks.map((track, index) => (*/}
                            {/*    <Track key={`track-${index}`} {...track} />*/}
                            {/*))}*/}
                            {/*{chapters?.length > 0 ? chapters.map((chapter, index) => (*/}
                            {/*    <Track kind="chapters" key={`chapter-${index}`} {...chapter} />*/}
                            {/*)) : cues.length > 0 ? cues.map((cue, index) => (*/}
                            {/*    <Track kind="chapters" key={`cue-${index}`} {...cue} />*/}
                            {/*)) : null}*/}
                        </MediaProvider>
                        {/*<div*/}
                        {/*    data-sea-media-player-skip-intro-outro-container*/}
                        {/*    className="absolute bottom-24 px-4 w-full justify-between flex items-center"*/}
                        {/*>*/}
                        {/*    <div>*/}
                        {/*        {showSkipIntroButton && (*/}
                        {/*            <Button intent="white" size="sm" onClick={onSkipIntro} loading={autoSkipIntroOutro}>*/}
                        {/*                Skip opening*/}
                        {/*            </Button>*/}
                        {/*        )}*/}
                        {/*    </div>*/}
                        {/*    <div>*/}
                        {/*        {showSkipEndingButton && (*/}
                        {/*            <Button intent="white" size="sm" onClick={onSkipOutro} loading={autoSkipIntroOutro}>*/}
                        {/*                Skip ending*/}
                        {/*            </Button>*/}
                        {/*        )}*/}
                        {/*    </div>*/}
                        {/*</div>*/}
                        <DefaultVideoLayout
                            icons={vidstackLayoutIcons}
                            slots={{
                                ...vidstackLayoutIcons,
                                settingsMenuEndItems: <>
                                    {/* {settingsItems} */}
                                    <NativePlayerPlaybackSubmenu />
                                </>,
                                // centerControlsGroupStart: <div>
                                //     {onGoToPreviousEpisode && (
                                //         <IconButton
                                //             intent="white-basic"
                                //             size="lg"
                                //             onClick={onGoToPreviousEpisode}
                                //             aria-label="Previous Episode"
                                //             icon={<LuArrowLeft className="size-12" />}
                                //         />
                                //     )}
                                // </div>,
                                // centerControlsGroupEnd: <div className="flex items-center justify-center gap-2">
                                //     {onGoToNextEpisode && (
                                //         <IconButton
                                //             intent="white-basic"
                                //             size="lg"
                                //             onClick={onGoToNextEpisode}
                                //             aria-label="Next Episode"
                                //             icon={<LuArrowRight className="size-12" />}
                                //         />
                                //     )}
                                // </div>
                            }}
                        />
                    </MediaPlayer>
                ) : (
                    <div
                        className="w-full h-full absolute flex justify-center items-center flex-col space-y-4 bg-black rounded-md"
                    >
                        <LoadingSpinner
                            spinner={
                                <div className="w-16 h-16 lg:w-[100px] lg:h-[100px] relative">
                                    <Image
                                        src="/logo_2.png"
                                        alt="Loading..."
                                        priority
                                        fill
                                        className="animate-pulse"
                                    />
                                </div>
                            }
                        />
                        <div className="text-center text-xs lg:text-sm">
                            {!!loadingState && <>
                                <p>{loadingState}</p>
                            </>}
                        </div>
                    </div>
                )}
            </div>
        </Drawer>
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
