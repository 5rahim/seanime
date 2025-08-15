"use client"

import { getExternalPlayerURL } from "@/api/client/external-player-link"
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
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { openTab } from "@/lib/helpers/browser"
import { logger } from "@/lib/helpers/debug"
import { useAtomValue } from "jotai"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { toast } from "sonner"
import { PluginEpisodeGridItemMenuItems } from "../_features/plugin/actions/plugin-actions"
import { useServerHMACAuth } from "../_hooks/use-server-status"

export default function Page() {

    const clientId = useAtomValue(clientIdAtom)
    const router = useRouter()
    const searchParams = useSearchParams()
    const mediaId = searchParams.get("id")
    const { data: animeEntry, isLoading: animeEntryLoading } = useGetAnimeEntry(mediaId)
    const { filePath, setFilePath } = useMediastreamCurrentFile()
    const { getHMACTokenQueryParam } = useServerHMACAuth()

    const { mutate: startManualTracking, isPending: isStarting } = usePlaybackStartManualTracking()

    const { externalPlayerLink, encodePath } = useExternalPlayerLink()

    function encodeFilePath(filePath: string) {
        if (encodePath) {
            return Buffer.from(filePath).toString("base64")
        }
        return encodeURIComponent(filePath)
    }

    React.useEffect(() => {
        // On mount, when the anime entry is loaded, and the file path is set, play the media file
        if (animeEntry && filePath) {
            const handleMediaPlay = async () => {
                // Get the episode
                const episode = animeEntry?.episodes?.find(ep => ep.localFile?.path === filePath)
                logger("MEDIALINKS").info("Filepath", filePath, "Episode", episode)

                if (!episode) {
                    logger("MEDIALINKS").error("Episode not found.")
                    toast.error("Episode not found.")
                    return
                }

                if (episode.type !== "main") {
                    logger("MEDIALINKS").warning("Episode is not a main episode. Cannot track progress.")
                }

                if (!externalPlayerLink) {
                    logger("MEDIALINKS").error("External player link is not set.")
                    toast.warning("External player link is not set.")
                    return
                }

                const endpoint = "/api/v1/mediastream/file?path=" + encodeFilePath(filePath)
                const tokenQueryParam = await getHMACTokenQueryParam("/api/v1/mediastream/file", "&")

                // Send video to external player
                let urlToSend = getServerBaseUrl() + endpoint + tokenQueryParam
                logger("MEDIALINKS").info("Opening external player", externalPlayerLink, "URL", urlToSend)

                // If the external player link includes a query parameter, we need to encode the URL to prevent query parameter conflicts
                if (externalPlayerLink.includes("?")) {
                    urlToSend = encodeURIComponent(urlToSend)
                }

                openTab(getExternalPlayerURL(externalPlayerLink, urlToSend))

                if (episode?.progressNumber && episode.type === "main") {
                    logger("MEDIALINKS").error("Starting manual tracking")
                    // Start manual tracking
                    React.startTransition(() => {
                        startManualTracking({
                            mediaId: animeEntry.mediaId,
                            episodeNumber: episode?.progressNumber,
                            clientId: clientId || "",
                        })
                    })
                } else {
                    logger("MEDIALINKS").warning("No manual tracking, progress number is not set.")
                }

                // Clear the file path
                setFilePath(undefined)
            }

            handleMediaPlay()
        }
    }, [animeEntry, filePath, externalPlayerLink, getHMACTokenQueryParam])

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
                        <SeaLink href={`/entry?id=${animeEntry?.mediaId}`}>
                            <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="md" />
                        </SeaLink>
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
                            description={episode?.episodeMetadata?.summary || episode?.episodeMetadata?.overview}
                            isWatched={!!animeEntry?.listData?.progress && (animeEntry.listData?.progress >= episode?.progressNumber)}
                            isFiller={episode.episodeMetadata?.isFiller}
                            isSelected={episode.localFile?.path === filePath}
                            length={episode.episodeMetadata?.length}
                            className="flex-none w-full"
                            episodeNumber={episode.episodeNumber}
                            progressNumber={episode.progressNumber}
                            action={<>
                                <MediaEpisodeInfoModal
                                    title={episode.displayTitle}
                                    image={episode.episodeMetadata?.image}
                                    episodeTitle={episode.episodeTitle}
                                    airDate={episode.episodeMetadata?.airDate}
                                    length={episode.episodeMetadata?.length}
                                    summary={episode.episodeMetadata?.summary || episode.episodeMetadata?.overview}
                                    isInvalid={episode.isInvalid}
                                    filename={episode.localFile?.parsedInfo?.original}
                                />

                                <PluginEpisodeGridItemMenuItems isDropdownMenu={true} type="medialinks" episode={episode} />
                            </>}
                        />
                    ))}
                </EpisodeListGrid>

            </AppLayoutStack>
        </>
    )

}
