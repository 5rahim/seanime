import {
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

    const { episodes, media, isFetching, isLoading } = useOnlinestreamEpisodeList(mediaId)

    const { episodeSource, isLoading: isLoadingEpisodeSource, isFetching: isFetchingEpisodeSource } = useOnlinestreamEpisodeSource(mediaId)

    const [episodeNumber, setEpisodeNumber] = useAtom(__onlinestream_selectedEpisodeNumberAtom)
    const [selectedServer, setServer] = useAtom(__onlinestream_selectedServerAtom)
    const setQuality = useSetAtom(__onlinestream_qualityAtom)
    const provider = useAtomValue(__onlinestream_selectedProviderAtom)
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
                        setServer(otherServers[0])
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
                previousCurrentTimeRef.current = 0
            }
        }
    }, [provider, videoSource])


    // Quality
    const hasCustomQualities = React.useMemo(() => !!episodeSource?.videoSources?.map(n => n.quality)?.filter(q => q.includes("p"))?.length,
        [episodeSource])
    //--
    const customQualities = React.useMemo(() => uniq(episodeSource?.videoSources?.map(n => n.quality)),
        [episodeSource])
    //--
    const previousCurrentTimeRef = React.useRef(0)
    const changeQuality = React.useCallback((quality: string) => {
        previousCurrentTimeRef.current = playerRef.current?.currentTime ?? 0
        setQuality(quality)
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
        episodeNumber: episodeSource?.number ?? 0,
        handleChangeEpisodeNumber,
        episodeLoading: isLoadingEpisodeSource || isFetchingEpisodeSource,
        opts: {
            currentEpisodeDetails: episodeDetails,
            servers,
            videoSource,
            customQualities,
            hasCustomQualities,
            changeQuality,
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
