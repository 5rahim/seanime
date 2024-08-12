"use client"

import { getServerBaseUrl } from "@/api/client/server-url"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { usePlaybackStartManualTracking } from "@/api/hooks/playback_manager.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { EpisodeGridItem } from "@/app/(main)/_features/anime/_components/episode-grid-item"
import { MediaEpisodeInfoModal } from "@/app/(main)/_features/media/_components/media-episode-info-modal"
import { EpisodeListGrid } from "@/app/(main)/entry/_components/episode-list-grid"
import { useMediastreamCurrentFile } from "@/app/(main)/mediastream/_lib/mediastream.atoms"
import { clientIdAtom } from "@/app/websocket-provider"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { logger } from "@/lib/helpers/debug"
import { useAtomValue } from "jotai"
import Link from "next/link"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { toast } from "sonner"

export default function Page() {

    const clientId = useAtomValue(clientIdAtom)
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const { filePath, setFilePath } = useMediastreamCurrentFile()

    const { mutate: startManualTracking, isPending: isStarting } = usePlaybackStartManualTracking()

    const { externalPlayerLink } = useExternalPlayerLink()

    React.useEffect(() => {
        // On mount, when the anime entry is loaded, and the file path is set, play the media file
        if (animeEntry && filePath) {
            // Get the episode
            const episode = animeEntry?.episodes?.find(ep => ep.localFile?.path === filePath)
            logger("MEDIALINKS").info("Filepath", filePath, "Episode", episode)

            if (!episode?.progressNumber) {
                logger("MEDIALINKS").error("Episode progress number is not set.")
                return
            }

            if (episode.type !== "main") {
                logger("MEDIALINKS").warning("Episode is not a main episode. Cannot track progress.")
                return
            }

            if (!externalPlayerLink) {
                logger("MEDIALINKS").error("External player link is not set.")
                toast.warning("External player link is not set.")
                return
            }

            // Send video to external player
            const urlToSend = getServerBaseUrl() + "/api/v1/mediastream/file/" + encodeURIComponent(filePath)
            logger("MEDIALINKS").info("Opening external player", externalPlayerLink, "URL", urlToSend)
            urlToSend.replace("127.0.0.1,", window.location.hostname).replace("localhost", window.location.hostname)

            // e.g. "mpv://http://localhost:3000/api/v1/mediastream/file/..."
            // e.g. "intent://http://localhost:3000/api/v1/mediastream/file/...#Intent;package=org.videolan.vlc;scheme=http;end"
            let url = externalPlayerLink.replace("{url}", urlToSend)

            if (externalPlayerLink.startsWith("intent://")) {
                // e.g. "intent://localhost:3000/api/v1/mediastream/file/...#Intent;package=org.videolan.vlc;scheme=http;end"
                url = url.replace("intent://http://", "intent://").replace("intent://https://", "intent://")
            }

            window.open(url, "_blank")

            // Start manual tracking
            React.startTransition(() => {
                startManualTracking({
                    mediaId: animeEntry.mediaId,
                    episodeNumber: episode?.progressNumber,
                    clientId: clientId || "",
                })
            })

            // Clear the file path
            setFilePath(undefined)
        }
    }, [animeEntry, filePath, externalPlayerLink])

    const mainEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "main") ?? []
    }, [animeEntry?.episodes])

    const specialEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "special") ?? []
    }, [animeEntry?.episodes])

    const ncEpisodes = React.useMemo(() => {
        return animeEntry?.episodes?.filter(ep => ep.type === "nc") ?? []
    }, [animeEntry?.episodes])

    const episodes = React.useMemo(() => {
        return [...mainEpisodes, ...specialEpisodes, ...ncEpisodes]
    }, [mainEpisodes, specialEpisodes, ncEpisodes])

    if (animeEntryLoading) return <LoadingSpinner />

    return (
        <>
            <CustomLibraryBanner discrete />

            <AppLayoutStack className="px-4 lg:px-8 z-[5]">

                <div className="flex flex-col lg:flex-row gap-2 w-full justify-between">
                    <div className="flex gap-4 items-center relative w-full">
                        <Link href={`/entry?id=${animeEntry?.mediaId}`}>
                            <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                        </Link>
                        <h3 className="max-w-full lg:max-w-[50%] text-ellipsis truncate">{animeEntry?.media?.title?.userPreferred}</h3>
                    </div>
                </div>

                <EpisodeListGrid>
                    {episodes.map((episode) => (
                        <EpisodeGridItem
                            key={episode.localFile?.path || ""}
                            id={`episode-${String(episode.episodeNumber)}`}
                            media={episode?.baseAnime as any}
                            title={episode?.displayTitle || episode?.baseAnime?.title?.userPreferred || ""}
                            image={episode?.episodeMetadata?.image || episode?.baseAnime?.coverImage?.large}
                            episodeTitle={episode?.episodeTitle}
                            fileName={episode?.localFile?.parsedInfo?.original}
                            onClick={() => {
                                if (episode.localFile?.path) {
                                    setFilePath(episode.localFile?.path)
                                }
                            }}
                            isWatched={!!animeEntry?.listData?.progress && (animeEntry.listData?.progress >= episode?.progressNumber)}
                            isFiller={episode.episodeMetadata?.isFiller}
                            isSelected={episode.localFile?.path === filePath}
                            length={episode.episodeMetadata?.length}
                            className="flex-none w-full"
                            action={<>
                                <MediaEpisodeInfoModal
                                    title={episode.displayTitle}
                                    image={episode.episodeMetadata?.image}
                                    episodeTitle={episode.episodeTitle}
                                    airDate={episode.episodeMetadata?.airDate}
                                    length={episode.episodeMetadata?.length}
                                    summary={episode.episodeMetadata?.summary}
                                    isInvalid={episode.isInvalid}
                                    filename={episode.localFile?.parsedInfo?.original}
                                />
                            </>}
                        />
                    ))}
                </EpisodeListGrid>

            </AppLayoutStack>
        </>
    )

}
