import { Anime_Entry } from "@/api/generated/types"
import { serverStatusAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { SeaMediaPlayer } from "@/app/(main)/_features/sea-media-player/sea-media-player"
import { SeaMediaPlayerLayout } from "@/app/(main)/_features/sea-media-player/sea-media-player-layout"
import { SeaMediaPlayerProvider } from "@/app/(main)/_features/sea-media-player/sea-media-player-provider"
import {
    OnlinestreamParametersButton,
    OnlinestreamProviderButton,
    OnlinestreamVideoQualitySubmenu,
    SwitchSubOrDubButton,
} from "@/app/(main)/onlinestream/_components/onlinestream-video-addons"
import { OnlinestreamManualMappingModal } from "@/app/(main)/onlinestream/_containers/onlinestream-manual-matching"
import { useHandleOnlinestream } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import { OnlinestreamManagerProvider } from "@/app/(main)/onlinestream/_lib/onlinestream-manager"
import { LuffyError } from "@/components/shared/luffy-error"
import { IconButton } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { logger } from "@/lib/helpers/debug"
import { isHLSProvider, MediaPlayerInstance, MediaProviderAdapter, MediaProviderChangeEvent, MediaProviderSetupEvent } from "@vidstack/react"
import HLS from "hls.js"
import { atom } from "jotai/index"
import { useAtomValue } from "jotai/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { FaSearch } from "react-icons/fa"
import { useUpdateEffect } from "react-use"
import "@/app/vidstack-theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { PluginEpisodeGridItemMenuItems } from "../../_features/plugin/actions/plugin-actions"

type OnlinestreamPageProps = {
    animeEntry?: Anime_Entry
    animeEntryLoading?: boolean
    hideBackButton?: boolean
}

type ProgressItem = {
    episodeNumber: number
}
const progressItemAtom = atom<ProgressItem | undefined>(undefined)

export function OnlinestreamPage({ animeEntry, animeEntryLoading, hideBackButton }: OnlinestreamPageProps) {
    const serverStatus = useAtomValue(serverStatusAtom)
    const router = useRouter()
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const urlEpNumber = searchParams.get("episode")

    const ref = React.useRef<MediaPlayerInstance>(null)

    const {
        episodes,
        currentEpisodeDetails,
        opts,
        url,
        onMediaDetached,
        onCanPlay: _onCanPlay,
        onFatalError,
        loadPage,
        media,
        episodeSource,
        currentEpisodeNumber,
        handleChangeEpisodeNumber,
        episodeLoading,
        isErrorEpisodeSource,
        isErrorProvider,
        provider,
    } = useHandleOnlinestream({
        mediaId,
        ref,
    })

    const maxEp = media?.nextAiringEpisode?.episode ? (media?.nextAiringEpisode?.episode - 1) : media?.episodes || 0
    const progress = animeEntry?.listData?.progress ?? 0

    /**
     * Set episode number on mount
     */
    const firstRenderRef = React.useRef(true)
    useUpdateEffect(() => {
        if (!!media && firstRenderRef.current && !!episodes) {
            const maxEp = media?.nextAiringEpisode?.episode ? (media?.nextAiringEpisode?.episode - 1) : media?.episodes || 0
            const _urlEpNumber = urlEpNumber ? Number(urlEpNumber) : undefined
            const progress = animeEntry?.listData?.progress ?? 0
            let nextProgressNumber = maxEp ? (progress + 1 < maxEp ? progress + 1 : maxEp) : progress + 1
            if (!episodes.find(e => e.number === nextProgressNumber)) {
                nextProgressNumber = 1
            }
            handleChangeEpisodeNumber(_urlEpNumber || nextProgressNumber || 1)
            logger("ONLINESTREAM").info("Setting episode number to", _urlEpNumber || nextProgressNumber || 1)
            firstRenderRef.current = false
        }
    }, [episodes])

    React.useEffect(() => {
        const t = setTimeout(() => {
            if (urlEpNumber) {
                router.replace(pathname + `?id=${mediaId}`)
            }
        }, 500)

        return () => clearTimeout(t)
    }, [mediaId])

    const episodeTitle = episodes?.find(e => e.number === currentEpisodeNumber)?.title

    function goToNextEpisode() {
        if (currentEpisodeNumber < maxEp) {
            // check if the episode exists
            if (episodes?.find(e => e.number === currentEpisodeNumber + 1)) {
                handleChangeEpisodeNumber(currentEpisodeNumber + 1)
            }
        }
    }

    function goToPreviousEpisode() {
        if (currentEpisodeNumber > 1) {
            // check if the episode exists
            if (episodes?.find(e => e.number === currentEpisodeNumber - 1)) {
                handleChangeEpisodeNumber(currentEpisodeNumber - 1)
            }
        }
    }

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

    React.useEffect(() => {
        const t = setTimeout(() => {
            const element = document.querySelector(".vds-quality-menu")
            if (opts.hasCustomQualities) {
                // Toggle the class
                element?.classList?.add("force-hidden")
            } else {
                // Toggle the class
                element?.classList?.remove("force-hidden")
            }
        }, 1000)
        return () => clearTimeout(t)
    }, [opts.hasCustomQualities, url, episodeLoading])

    if (!loadPage || !media || animeEntryLoading) return <div data-onlinestream-page-loading-container className="space-y-4">
        <div className="flex gap-4 items-center relative">
            <Skeleton className="h-12" />
        </div>
        <div
            className="grid 2xl:grid-cols-[1fr,450px] gap-4 xl:gap-4"
        >
            <div className="w-full min-h-[70dvh] relative">
                <Skeleton className="h-full w-full absolute" />
            </div>

            <Skeleton className="hidden 2xl:block relative h-[78dvh] overflow-y-auto pr-4 pt-0" />

        </div>
    </div>

    return (
        <SeaMediaPlayerProvider
            media={media}
            progress={{
                currentProgress: progress,
                currentEpisodeNumber,
                currentEpisodeTitle: episodeTitle || null,
            }}
        >
            <OnlinestreamManagerProvider opts={opts}>
                <SeaMediaPlayerLayout
                    mediaId={mediaId ? Number(mediaId) : undefined}
                    title={media?.title?.userPreferred}
                    hideBackButton={hideBackButton}
                    episodes={episodes}
                    leftHeaderActions={<>
                        {animeEntry && <OnlinestreamManualMappingModal entry={animeEntry}>
                            <IconButton
                                size="sm"
                                intent="gray-basic"
                                icon={<FaSearch />}
                            />
                        </OnlinestreamManualMappingModal>}
                        <SwitchSubOrDubButton />
                        {!!mediaId && <OnlinestreamParametersButton mediaId={Number(mediaId)} />}
                        <div className="flex flex-1"></div>
                    </>}
                    mediaPlayer={!provider ? (
                        <div className="flex items-center flex-col justify-center w-full h-full">
                            <LuffyError title="No provider selected" />
                            {!!mediaId && <OnlinestreamParametersButton mediaId={Number(mediaId)} />}
                        </div>
                    ) : isErrorProvider ? <LuffyError title="Provider error" /> : (
                        <SeaMediaPlayer
                            url={url}
                            poster={currentEpisodeDetails?.image || media.coverImage?.extraLarge}
                            isLoading={episodeLoading}
                            isPlaybackError={isErrorEpisodeSource}
                            playerRef={ref}
                            onProviderChange={onProviderChange}
                            onProviderSetup={onProviderSetup}
                            onCanPlay={_onCanPlay}
                            onGoToNextEpisode={goToNextEpisode}
                            onGoToPreviousEpisode={goToPreviousEpisode}
                            tracks={episodeSource?.subtitles?.map((sub) => ({
                                id: sub.language,
                                label: sub.language,
                                kind: "subtitles",
                                src: sub.url,
                                language: sub.language,
                                default: sub.language
                                    ? sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us"
                                    : sub.language?.toLowerCase() === "english" || sub.language?.toLowerCase() === "en-us",
                            }))}
                            settingsItems={<>
                                {opts.hasCustomQualities ? (
                                    <OnlinestreamVideoQualitySubmenu />
                                ) : null}
                            </>}
                            videoLayoutSlots={{
                                beforeCaptionButton: <>
                                    <div className="flex items-center">
                                        <OnlinestreamProviderButton />
                                    </div>
                                </>,
                            }}
                        />
                    )}
                    episodeList={<>
                        {(!episodes?.length && !loadPage) && <p>
                            No episodes found
                        </p>}
                        {episodes?.filter(Boolean)?.sort((a, b) => a!.number - b!.number)?.map((episode, idx) => {
                            return (
                                <EpisodeGridItem
                                    key={idx + (episode.title || "") + episode.number}
                                    id={`episode-${String(episode.number)}`}
                                    onClick={() => handleChangeEpisodeNumber(episode.number)}
                                    title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                    episodeTitle={episode.title}
                                    description={episode.description ?? undefined}
                                    image={episode.image}
                                    media={media}
                                    isSelected={episode.number === currentEpisodeNumber}
                                    disabled={episodeLoading}
                                    isWatched={progress ? episode.number <= progress : undefined}
                                    className="flex-none w-full"
                                    isFiller={episode.isFiller}
                                    episodeNumber={episode.number}
                                    progressNumber={episode.number}
                                    action={<>
                                        <MediaEpisodeInfoModal
                                            title={media.format === "MOVIE" ? "Complete movie" : `Episode ${episode.number}`}
                                            image={episode?.image}
                                            episodeTitle={episode.title}
                                            summary={episode?.description}
                                        />

                                        <PluginEpisodeGridItemMenuItems isDropdownMenu={true} type="onlinestream" episode={episode} />
                                    </>}
                                />
                            )
                        })}
                        <p className="text-center text-[--muted] py-2">End</p>
                    </>}
                />
            </OnlinestreamManagerProvider>
        </SeaMediaPlayerProvider>
    )
}
