import {
    __onlinestream_autoPlayAtom,
    __onlinestream_qualityAtom,
    __onlinestream_selectedEpisodeNumberAtom,
    __onlinestream_selectedProviderAtom,
    __onlinestream_selectedServerAtom,
    useOnlinestreamEpisodeList,
    useOnlinestreamEpisodeSource,
    useOnlinestreamVideoSource,
} from "@/app/(main)/onlinestream/_lib/episodes"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { logger } from "@/lib/helpers/debug"
import { MediaPlayerInstance } from "@vidstack/react"
import { useAtom, useAtomValue, useSetAtom } from "jotai/react"
import { uniq } from "lodash"
import React from "react"
import { toast } from "sonner"

type OnlinestreamManagerProps = {
    mediaId: string | null
    ref: React.RefObject<MediaPlayerInstance>
}

export function useOnlinestreamManager(props: OnlinestreamManagerProps) {

    const { mediaId, ref: playerRef } = props

    const { episodes, media, isFetching, isLoading, isSuccess } = useOnlinestreamEpisodeList(mediaId)

    const { episodeSource, isLoading: isLoadingEpisodeSource, isFetching: isFetchingEpisodeSource } = useOnlinestreamEpisodeSource(mediaId, isSuccess)

    const setEpisodeNumber = useSetAtom(__onlinestream_selectedEpisodeNumberAtom)
    const setServer = useSetAtom(__onlinestream_selectedServerAtom)
    const setQuality = useSetAtom(__onlinestream_qualityAtom)
    const autoPlay = useAtomValue(__onlinestream_autoPlayAtom)
    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)
    const currentProviderRef = React.useRef<string | null>(null)

    // Get current episode details when [episodes] or [episodeSource] changes
    const episodeDetails = React.useMemo(() => {
        return episodes?.find((episode) => episode.number === episodeSource?.number)
    }, [episodes, episodeSource])

    // Get the list of servers
    const servers = React.useMemo(() => {
        if (!episodeSource) return []
        return uniq(episodeSource.videoSources.map((source) => source.server))
    }, [episodeSource])

    React.useEffect(() => {
        logger("ONLINESTREAM").info("episodeSource", episodeSource)
        if (episodeSource) {
            setEpisodeNumber(episodeSource.number)
        }
    }, [episodeSource])

    // Get the current video source
    // [useOnlinestreamVideoSource] handles selecting the best source
    const { videoSource } = useOnlinestreamVideoSource(episodeSource)

    React.useEffect(() => {
        logger("ONLINESTREAM").info("videoSource", videoSource)
    }, [videoSource])

    const [url, setUrl] = React.useState<string | undefined>(undefined)

    React.useEffect(() => {
        setUrl(undefined)
        if (videoSource?.url) {
            setServer(videoSource.server)
            React.startTransition(() => {
                setUrl(videoSource?.url)
            })
        }
    }, [videoSource?.url])

    React.useEffect(() => {
        logger("ONLINESTREAM").info("provider", provider)
        currentProviderRef.current = provider
    }, [provider])

    React.useEffect(() => {
        logger("ONLINESTREAM").info("url", url)
    }, [url])


    // Handle playback error
    const [erroredServers, setErroredServers] = React.useState<string[]>([])
    React.useEffect(() => {
        setErroredServers([])
    }, [episodeDetails])
    //--
    const onMediaDetached = React.useCallback(() => {
        logger("ONLINESTREAM").error("onMediaDetached", provider == currentProviderRef.current)
        if (provider == currentProviderRef.current) {
            // onFatalError()
        }
    }, [provider])
    //--
    const onFatalError = React.useCallback(() => {
        logger("ONLINESTREAM").error("onFatalError", provider == currentProviderRef.current)
        if (provider == currentProviderRef.current) {
            setUrl(undefined)
            toast.info("Playback error, changing server")
            setTimeout(() => {
                logger("ONLINESTREAM").error("erroredServers", erroredServers)
                if (videoSource?.server) {
                    const otherServers = servers.filter((server) => server !== videoSource?.server && !erroredServers.includes(server))
                    if (otherServers.length > 0) {
                        setErroredServers((prev) => [...prev, videoSource?.server])
                        setServer(otherServers[0])
                    } else {
                        setProvider((prev) => (prev === "gogoanime" ? "zoro" : "gogoanime"))
                    }
                }
            }, 500)
        }
    }, [provider, videoSource])
    //--
    const onProviderSetup = React.useCallback(() => {
        logger("ONLINESTREAM").error("Provider setup", provider == currentProviderRef.current)
        if (provider == currentProviderRef.current) {
            // Restore time if set
            if (previousCurrentTimeRef.current > 0) {
                Object.assign(playerRef.current ?? {}, { currentTime: previousCurrentTimeRef.current })
                setTimeout(() => {
                    if (previousIsPlayingRef.current) {
                        playerRef.current?.play()
                    }
                }, 500)
                previousCurrentTimeRef.current = 0
            }
        }
    }, [provider, videoSource, autoPlay])


    // Quality
    const hasCustomQualities = React.useMemo(() => !!episodeSource?.videoSources?.map(n => n.quality)?.filter(q => q.includes("p"))?.length,
        [episodeSource])
    //--
    const customQualities = React.useMemo(() => uniq(episodeSource?.videoSources?.map(n => n.quality)),
        [episodeSource])
    //--
    const previousCurrentTimeRef = React.useRef(0)
    const previousIsPlayingRef = React.useRef(false)
    const changeQuality = React.useCallback((quality: string) => {
        previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
        previousIsPlayingRef.current = playerRef.current?.paused === false
        setQuality(quality)
    }, [videoSource])

    // Provider
    const changeProvider = React.useCallback((provider: string) => {
        previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
        previousIsPlayingRef.current = playerRef.current?.paused === false
        setProvider(provider)
    }, [videoSource])

    // Server
    const changeServer = React.useCallback((server: string) => {
        previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
        previousIsPlayingRef.current = playerRef.current?.paused === false
        setServer(server)
    }, [videoSource])

    // Episode
    const handleChangeEpisodeNumber = React.useCallback((epNumber: number) => {
        setEpisodeNumber(epNumber)
    }, [episodeSource])

    return {
        currentEpisodeDetails: episodeDetails,
        servers,
        videoSource,
        onMediaDetached,
        onProviderSetup,
        onFatalError,
        url,
        episodes,
        media: media as BaseMediaFragment,
        episodeSource,
        loadPage: !isFetching && !isLoading,
        currentEpisodeNumber: episodeSource?.number ?? 0,
        handleChangeEpisodeNumber,
        episodeLoading: isLoadingEpisodeSource || isFetchingEpisodeSource,
        opts: {
            currentEpisodeDetails: episodeDetails,
            servers,
            videoSource,
            customQualities,
            hasCustomQualities,
            changeQuality,
            changeProvider,
            changeServer,
        },
    }

}

type OnlinestreamManagerOpts = ReturnType<typeof useOnlinestreamManager>

//@ts-ignore
const __OnlinestreamManagerContext = React.createContext<OnlinestreamManagerOpts["opts"]>({})

export function useOnlinestreamManagerContext() {
    return React.useContext(__OnlinestreamManagerContext)
}

export function OnlinestreamManagerProvider(props: { children?: React.ReactNode, opts: OnlinestreamManagerOpts["opts"] }) {
    return (
        <__OnlinestreamManagerContext.Provider
            value={props.opts}
        >
            {props.children}
        </__OnlinestreamManagerContext.Provider>
    )
}
