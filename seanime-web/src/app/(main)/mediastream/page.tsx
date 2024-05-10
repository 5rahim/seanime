"use client"

import { useGetAnimeEntry, useUpdateAnimeEntryProgress } from "@/api/hooks/anime_entries.hooks"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { __mediastream_progressItemAtom, useHandleMediastream } from "@/app/(main)/mediastream/_lib/handle-mediastream"
import { useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Skeleton } from "@/components/ui/skeleton"
import { MediaPlayer, MediaPlayerInstance, MediaProvider, Track } from "@vidstack/react"
import "@vidstack/react/player/styles/default/theme.css"
import "@vidstack/react/player/styles/default/layouts/video.css"
import { defaultLayoutIcons, DefaultVideoLayout } from "@vidstack/react/player/layouts/default"
import { useAtom } from "jotai/react"
import { CaptionsFileFormat } from "media-captions"
import Image from "next/image"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import React, { useMemo } from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"


export default function Page() {

    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: mediaEntry, isLoading: mediaEntryLoading } = useGetAnimeEntry(mediaId)
    const playerRef = React.useRef<MediaPlayerInstance>(null)
    const { filePath } = useMediastreamCurrentFile()

    const mainEpisodes = useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "main") ?? []
    }, [mediaEntry?.episodes])

    const specialEpisodes = useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "special") ?? []
    }, [mediaEntry?.episodes])

    const ncEpisodes = useMemo(() => {
        return mediaEntry?.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [mediaEntry?.episodes])

    const episodes = React.useMemo(() => {
        return [...mainEpisodes, ...specialEpisodes, ...ncEpisodes]
    }, [mainEpisodes, specialEpisodes, ncEpisodes])

    const {
        url,
        isError,
        isMediaContainerLoading,
        streamType,
        mediaContainer,
        subtitles,
        subtitleEndpointUri,
        onProviderChange,
        onProviderSetup,
        onTimeUpdate,
        onCanPlay,
        onEnded,
        onPlayFile,
    } = useHandleMediastream({ playerRef, episodes })

    const { mutate: updateProgress, isPending: isUpdatingProgress, isSuccess: hasUpdatedProgress } = useUpdateAnimeEntryProgress(mediaId)

    const [progressItem, setProgressItem] = useAtom(__mediastream_progressItemAtom)

    const [currentProgress, setCurrentProgress] = React.useState(mediaEntry?.listData?.progress ?? 0)

    React.useEffect(() => {
        if (!mediaId || (!mediaEntryLoading && !mediaEntry) || (!mediaEntryLoading && !!mediaEntry && !filePath)) {
            router.push("/")
        }
        if (mediaEntry) {
            setCurrentProgress(mediaEntry.listData?.progress ?? 0)
        }
    }, [mediaId, mediaEntry, mediaEntryLoading, filePath])

    if (mediaEntryLoading) return <LoadingSpinner />

    return (
        <AppLayoutStack className="p-8">

            <div className="flex w-full justify-between">
                <div className="flex gap-4 items-center relative w-full">
                    <Link href={`/entry?id=${mediaEntry?.mediaId}`}>
                        <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                    </Link>
                    <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">Streaming: {mediaEntry?.media?.title?.userPreferred}</h3>
                </div>

                <div className="flex gap-2 items-center">
                    {(!!progressItem && !progressItem.updated && mediaEntry?.media && progressItem.episodeNumber > currentProgress) && <Button
                        className="animate-pulse"
                        loading={isUpdatingProgress}
                        disabled={hasUpdatedProgress}
                        onClick={() => {
                            updateProgress({
                                episodeNumber: progressItem.episodeNumber,
                                mediaId: mediaEntry.media!.id,
                                totalEpisodes: mediaEntry.media!.episodes || 0,
                                malId: mediaEntry.media!.idMal || undefined,
                            }, {
                                onSuccess: () => setProgressItem(prev => !!prev ? ({
                                    ...prev,
                                    updated: true,
                                }) : undefined),
                            })
                            setCurrentProgress(progressItem.episodeNumber)
                        }}
                    >Update progress</Button>}
                </div>
            </div>

            <div
                className={cn(
                    "grid gap-4 xl:gap-4 w-full",
                    "xl:grid-cols-[1fr,500px]",
                )}
            >

                <div
                    className={cn(
                        "aspect-video relative w-full",
                    )}
                >
                    {isError ?
                        <LuffyError title="Playback Error" /> :
                        (!!url && !isMediaContainerLoading) ? <MediaPlayer
                            ref={playerRef}
                            crossOrigin
                            src={url}
                            poster={mediaEntry?.media?.bannerImage || mediaEntry?.media?.coverImage?.extraLarge || ""}
                            onProviderChange={onProviderChange}
                            onProviderSetup={onProviderSetup}
                            onTimeUpdate={onTimeUpdate}
                            onCanPlay={onCanPlay}
                            onEnded={onEnded}
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
                                slots={{
                                    // beforeSettingsMenu: (
                                    //     <MediastreamAudioSubmenu />
                                    // )
                                }}
                            />
                        </MediaPlayer> : (
                            <Skeleton className="h-full w-full absolute flex justify-center items-center flex-col space-y-4">
                                <LoadingSpinner
                                    containerClass=""
                                    spinner={<Image
                                        src="/logo_2.png"
                                        alt="Loading..."
                                        priority
                                        width={100}
                                        height={100}
                                        className="animate-pulse"
                                    />}
                                />
                                <p>
                                    Extracting video metadata...
                                </p>
                                <p>
                                    This might take a while.
                                </p>
                            </Skeleton>
                        )}
                </div>

                <ScrollArea className="relative xl:sticky h-[75dvh] overflow-y-auto pr-4 pt-0">
                    <div className="space-y-4">
                        {episodes.map((episode) => (
                            <EpisodeGridItem
                                key={episode.localFile?.path || ""}
                                media={episode?.basicMedia as any}
                                title={episode?.displayTitle || episode?.basicMedia?.title?.userPreferred || ""}
                                image={episode?.episodeMetadata?.image || episode?.basicMedia?.coverImage?.large}
                                episodeTitle={episode?.episodeTitle}
                                description={episode?.episodeMetadata?.overview}
                                onClick={() => {
                                    if (episode.localFile?.path) {
                                        onPlayFile(episode.localFile?.path || "")
                                    }
                                }}
                                // description={episode?.absoluteEpisodeNumber !== episodeNumber
                                //     ? `(Episode ${episode?.absoluteEpisodeNumber})`
                                //     : undefined}
                                isWatched={!!currentProgress && currentProgress >= episode?.progressNumber}
                                isSelected={episode.localFile?.path === filePath}
                                imageContainerClassName="w-20 h-20"
                                className="flex-none w-full"
                            />
                        ))}
                    </div>
                </ScrollArea>

            </div>


        </AppLayoutStack>
    )

}
