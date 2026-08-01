import { ExtensionRepo_OnlinestreamProviderExtensionItem, Onlinestream_EpisodeListResponse, Onlinestream_EpisodeSource } from "@/api/generated/types"
import { logger, useLatestFunction } from "@/lib/helpers/debug"
import React from "react"
import { toast } from "sonner"
import { getRefreshKey, markSourceRefreshed, orderProviders, shouldRecoverStartup } from "./onlinestream-provider-trial"

type TrialState = {
    providers: string[]
    providerIndex: number
    serverIndex: number
}

type RefreshEpisodeSourceVariables = {
    mediaId: number
    provider: string
    episodeNumber: number
    dubbed: boolean
    refresh?: boolean
}

type UseOnlinestreamAutoProviderCyclerProps = {
    mediaId: number
    provider: string | null
    server: string | undefined
    url: string | null
    providerExtensions: ExtensionRepo_OnlinestreamProviderExtensionItem[]
    dubbed: boolean
    sourceDubbed: boolean
    currentEpisodeNumber: number | null
    episodeListResponse?: Onlinestream_EpisodeListResponse
    episodeListLoading: boolean
    isEpisodeListFetched: boolean
    isEpisodeListError: boolean
    episodeSource?: Onlinestream_EpisodeSource
    episodeSourceLoading: boolean
    isEpisodeSourceError: boolean
    playbackError: string | null
    setProvider: (provider: string | null) => void
    setServer: (server: string | undefined) => void
    setSelectedEpisodeNumber: (episodeNumber: number) => void
    setUrl: (url: string | null) => void
    setPlaybackError: (error: string | null) => void
    refreshEpisodeSource: (variables: RefreshEpisodeSourceVariables) => Promise<Onlinestream_EpisodeSource | undefined>
}

const log = logger("ONLINESTREAM")
const PROVIDER_TIMEOUT_MS = 15_000
const PLAYBACK_TIMEOUT_MS = 20_000

function getServers(episodeSource: Onlinestream_EpisodeSource | undefined) {
    return Array.from(new Set(
        (episodeSource?.videoSources ?? [])
            .map(source => source.server)
            .filter((server): server is string => !!server),
    ))
}

export function useOnlinestreamAutoProviderCycler(props: UseOnlinestreamAutoProviderCyclerProps) {
    const {
        mediaId,
        provider,
        server,
        url,
        providerExtensions,
        dubbed,
        sourceDubbed,
        currentEpisodeNumber,
        episodeListResponse,
        episodeListLoading,
        isEpisodeListFetched,
        isEpisodeListError,
        episodeSource,
        episodeSourceLoading,
        isEpisodeSourceError,
        playbackError,
        setProvider,
        setServer,
        setSelectedEpisodeNumber,
        setUrl,
        setPlaybackError,
        refreshEpisodeSource,
    } = props

    const [trial, setTrial] = React.useState<TrialState | null>(null)
    const [detectedFailure, setDetectedFailure] = React.useState<string | null>(null)
    const [refreshing, setRefreshing] = React.useState(false)
    const trialRef = React.useRef<TrialState | null>(null)
    const playbackTimeRef = React.useRef(0)
    const refreshingRef = React.useRef(false)
    const refreshedSourcesRef = React.useRef(new Set<string>())
    const activeCandidateRef = React.useRef("")
    activeCandidateRef.current = getRefreshKey(provider, server, currentEpisodeNumber)

    const availableProviders = React.useMemo(() => {
        return providerExtensions
            .filter(provider => !dubbed || provider.supportsDub)
            .sort((a, b) => a.name.localeCompare(b.name))
    }, [providerExtensions, dubbed])

    const setTrialState = React.useCallback((nextTrial: TrialState | null) => {
        trialRef.current = nextTrial
        setTrial(nextTrial)
    }, [])

    const stopWithFailure = useLatestFunction((reason: string) => {
        log.warning("No working provider found", reason)
        setTrialState(null)
        setUrl(null)
        setPlaybackError("No working providers found")
        toast.error("No working providers found")
    })

    const goToNextProvider = useLatestFunction((reason: string) => {
        const currentTrial = trialRef.current
        if (!currentTrial) return

        const nextProviderIndex = currentTrial.providerIndex + 1
        if (nextProviderIndex >= currentTrial.providers.length) {
            stopWithFailure(reason)
            return
        }

        log.warning("Trying next provider", reason)
        setUrl(null)
        setPlaybackError(null)
        setTrialState({
            ...currentTrial,
            providerIndex: nextProviderIndex,
            serverIndex: 0,
        })
    })

    const goToNextCandidate = useLatestFunction((reason: string) => {
        const currentTrial = trialRef.current
        if (!currentTrial) {
            setDetectedFailure(reason)
            setPlaybackError(reason)
            return
        }

        const servers = getServers(episodeSource)
        const nextServerIndex = currentTrial.serverIndex + 1
        if (nextServerIndex < servers.length) {
            log.warning("Trying next server", { reason, server: servers[nextServerIndex] })
            setUrl(null)
            setPlaybackError(null)
            setTrialState({ ...currentTrial, serverIndex: nextServerIndex })
            return
        }

        goToNextProvider(reason)
    })

    const tryAllProviders = useLatestFunction(() => {
        if (!mediaId) return
        if (!availableProviders.length) {
            toast.warning(dubbed ? "No dubbed providers available" : "No providers available")
            return
        }

        const providers = orderProviders(availableProviders, provider)
        const currentServerIndex = getServers(episodeSource).findIndex(candidate => candidate === server)
        let providerIndex = 0
        let serverIndex = currentServerIndex >= 0 ? currentServerIndex : 0
        if (detectedFailure && provider && providers[0] === provider) {
            const servers = getServers(episodeSource)
            if (currentServerIndex >= 0 && currentServerIndex + 1 < servers.length) {
                serverIndex = currentServerIndex + 1
            } else if (providers.length > 1) {
                providerIndex = 1
                serverIndex = 0
            }
        }
        const nextTrial = {
            providers,
            providerIndex,
            serverIndex,
        }

        log.info("Trying providers", { providers, dubbed })
        playbackTimeRef.current = 0
        setDetectedFailure(null)
        setTrialState(nextTrial)
        setUrl(null)
        setPlaybackError(null)
        setProvider(providers[providerIndex])
    })

    const onPlaybackError = useLatestFunction((reason: string) => {
        if (!shouldRecoverStartup(playbackTimeRef.current) || refreshingRef.current) return
        if (!provider || currentEpisodeNumber === null) {
            goToNextCandidate(reason)
            return
        }

        const refreshKey = getRefreshKey(provider, server, currentEpisodeNumber)
        if (!markSourceRefreshed(refreshedSourcesRef.current, refreshKey)) {
            goToNextCandidate(reason)
            return
        }

        log.warning("Refreshing episode source", { reason, provider, server })
        refreshingRef.current = true
        setRefreshing(true)
        setUrl(null)
        setPlaybackError(null)

        refreshEpisodeSource({
            mediaId,
            provider,
            episodeNumber: currentEpisodeNumber,
            dubbed: sourceDubbed,
            refresh: true,
        }).then(source => {
            if (activeCandidateRef.current !== refreshKey) return
            if (!source?.videoSources?.length) {
                goToNextCandidate(reason)
                return
            }
            setUrl(null)
            setPlaybackError(null)
        }).catch(() => {
            if (activeCandidateRef.current === refreshKey) {
                goToNextCandidate(reason)
            }
        }).finally(() => {
            refreshingRef.current = false
            setRefreshing(false)
        })
    })

    const onPlaybackStalled = useLatestFunction((reason: string) => {
        onPlaybackError(reason)
    })

    const onLoadedMetadata = useLatestFunction(() => {
        setPlaybackError(null)
    })

    const onTimeUpdate = useLatestFunction((e: React.SyntheticEvent<HTMLVideoElement>) => {
        playbackTimeRef.current = Math.max(playbackTimeRef.current, e.currentTarget.currentTime)
        if (shouldRecoverStartup(playbackTimeRef.current)) return

        if (detectedFailure) {
            setDetectedFailure(null)
        }

        if (!trialRef.current) return

        log.success("Found working provider", { provider, server })
        setTrialState(null)
        setPlaybackError(null)
    })

    const cancel = useLatestFunction(() => {
        if (!trialRef.current) return
        setTrialState(null)
        setDetectedFailure(null)
        toast.info("Stopped trying providers")
    })

    React.useEffect(() => {
        playbackTimeRef.current = 0
        setDetectedFailure(null)
    }, [provider, server, currentEpisodeNumber])

    React.useEffect(() => {
        refreshedSourcesRef.current.clear()
    }, [mediaId, dubbed])

    React.useEffect(() => {
        if (!trial) return

        const targetProvider = trial.providers[trial.providerIndex]
        if (!targetProvider) return

        if (provider !== targetProvider) {
            setUrl(null)
            setPlaybackError(null)
            setServer(undefined)
            setProvider(targetProvider)
        }
    }, [trial, provider, setProvider, setServer, setUrl, setPlaybackError])

    React.useEffect(() => {
        if (!trial) return
        if (provider !== trial.providers[trial.providerIndex]) return
        if (episodeListLoading) return

        if (isEpisodeListError) {
            goToNextProvider("episode list error")
            return
        }

        const episodes = episodeListResponse?.episodes ?? []
        if (isEpisodeListFetched && !episodes.length) {
            goToNextProvider("no episodes")
            return
        }

        if (isEpisodeListFetched && episodes.length && currentEpisodeNumber === null) {
            setSelectedEpisodeNumber(episodes[0].number)
            return
        }

        if (isEpisodeListFetched && currentEpisodeNumber !== null && !episodes.some(episode => episode.number === currentEpisodeNumber)) {
            goToNextProvider("episode not found")
        }
    }, [
        trial,
        provider,
        episodeListResponse,
        episodeListLoading,
        isEpisodeListFetched,
        isEpisodeListError,
        currentEpisodeNumber,
        setSelectedEpisodeNumber,
        goToNextProvider,
    ])

    React.useEffect(() => {
        if (trial || detectedFailure || !episodeListLoading) return

        const timeout = window.setTimeout(() => {
            setDetectedFailure("episode list timeout")
        }, PROVIDER_TIMEOUT_MS)

        return () => window.clearTimeout(timeout)
    }, [trial, detectedFailure, provider, dubbed, episodeListLoading])

    React.useEffect(() => {
        if (!trial) return
        if (provider !== trial.providers[trial.providerIndex]) return
        if (!episodeListLoading) return

        const timeout = window.setTimeout(() => {
            goToNextProvider("episode list timeout")
        }, PROVIDER_TIMEOUT_MS)

        return () => window.clearTimeout(timeout)
    }, [trial, provider, episodeListLoading, goToNextProvider])

    React.useEffect(() => {
        if (!trial || refreshing) return
        if (provider !== trial.providers[trial.providerIndex]) return
        if (!isEpisodeListFetched || episodeListLoading || currentEpisodeNumber === null) return
        if (episodeSourceLoading) return

        if (isEpisodeSourceError) {
            onPlaybackError("episode source error")
            return
        }

        if (!episodeSource) return

        const servers = getServers(episodeSource)
        if (!servers.length) {
            onPlaybackError("no video sources")
            return
        }

        const targetServer = servers[trial.serverIndex]
        if (!targetServer) {
            goToNextProvider("servers exhausted")
            return
        }

        setUrl(null)
        setPlaybackError(null)
        setServer(targetServer)
    }, [
        trial,
        refreshing,
        provider,
        episodeSource,
        episodeSourceLoading,
        isEpisodeSourceError,
        isEpisodeListFetched,
        episodeListLoading,
        currentEpisodeNumber,
        setServer,
        setUrl,
        setPlaybackError,
        onPlaybackError,
        goToNextProvider,
    ])

    React.useEffect(() => {
        if (trial || detectedFailure || refreshing) return
        if (!isEpisodeListFetched || episodeListLoading || currentEpisodeNumber === null) return
        if (episodeSourceLoading) return

        if (isEpisodeSourceError) {
            onPlaybackError("episode source error")
            return
        }

        if (episodeSource && !(episodeSource.videoSources ?? []).length) {
            onPlaybackError("no video sources")
        }
    }, [
        trial,
        detectedFailure,
        refreshing,
        isEpisodeListFetched,
        episodeListLoading,
        currentEpisodeNumber,
        episodeSource,
        episodeSourceLoading,
        isEpisodeSourceError,
        onPlaybackError,
    ])

    React.useEffect(() => {
        if (refreshing || detectedFailure) return
        if (!isEpisodeListFetched || episodeListLoading || currentEpisodeNumber === null) return
        if (!episodeSourceLoading) return

        const timeout = window.setTimeout(() => {
            onPlaybackError("episode source timeout")
        }, PROVIDER_TIMEOUT_MS)

        return () => window.clearTimeout(timeout)
    }, [
        refreshing,
        detectedFailure,
        provider,
        isEpisodeListFetched,
        episodeListLoading,
        currentEpisodeNumber,
        episodeSourceLoading,
        onPlaybackError,
    ])

    React.useEffect(() => {
        if (!trial || refreshing) return
        if (provider !== trial.providers[trial.providerIndex]) return
        if (!episodeSource || episodeSourceLoading || isEpisodeSourceError) return

        const targetServer = getServers(episodeSource)[trial.serverIndex]
        if (!targetServer || server !== targetServer) return

        const timeout = window.setTimeout(() => {
            onPlaybackError("playback timeout")
        }, PLAYBACK_TIMEOUT_MS)

        return () => window.clearTimeout(timeout)
    }, [
        trial,
        refreshing,
        provider,
        server,
        episodeSource,
        episodeSourceLoading,
        isEpisodeSourceError,
        onPlaybackError,
    ])

    React.useEffect(() => {
        if (trial || detectedFailure || refreshing || !url) return
        if (episodeSourceLoading || isEpisodeSourceError) return

        const startedAt = playbackTimeRef.current
        const timeout = window.setTimeout(() => {
            if (playbackTimeRef.current <= startedAt) {
                onPlaybackError("playback timeout")
            }
        }, PLAYBACK_TIMEOUT_MS)

        return () => window.clearTimeout(timeout)
    }, [trial, detectedFailure, refreshing, url, episodeSourceLoading, isEpisodeSourceError, onPlaybackError])

    const hasEpisodeListFailure = isEpisodeListError || (isEpisodeListFetched && !episodeListLoading && !(episodeListResponse?.episodes ?? []).length)
    const hasEpisodeSourceFailure = isEpisodeSourceError || (!!episodeSource && !episodeSourceLoading && !(episodeSource.videoSources ?? []).length)
    const hasFailure = hasEpisodeListFailure || hasEpisodeSourceFailure || !!playbackError || !!detectedFailure
    const canTry = !!mediaId && !!availableProviders.length

    return {
        isTrying: !!trial,
        showButton: canTry && (!!trial || (!refreshing && hasFailure)),
        tryAllProviders,
        cancel,
        onPlaybackError,
        onPlaybackStalled,
        onLoadedMetadata,
        onTimeUpdate,
    }
}
