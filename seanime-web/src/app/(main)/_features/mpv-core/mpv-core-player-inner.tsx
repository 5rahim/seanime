import { API_ENDPOINTS } from "@/api/generated/endpoints"
import type { Player_PlaybackInfo, Player_SkipData, Player_SubtitleTrack } from "@/api/generated/types"
import { useVideoCoreSaveScreenshot } from "@/api/hooks/videocore.hooks"
import {
    MediaCoreControlBarView,
    MediaCoreControlButtonIcon,
    MediaCoreFullscreenButton,
    MediaCoreNextButton,
    MediaCorePipButton,
    MediaCorePlayButton,
    MediaCorePreviousButton,
    MediaCoreTimestamp,
    MediaCoreVolumeButton,
} from "@/app/(main)/_features/media-core/media-core-control-bar"

import { MediaCoreDrawer } from "@/app/(main)/_features/media-core/media-core-drawer"
import { MediaCoreMenu, MediaCoreMenuBody, MediaCoreMenuTitle, MediaCoreSettingSelect } from "@/app/(main)/_features/media-core/media-core-menu"
import {
    MediaCoreBufferingOverlay,
    MediaCoreErrorOverlay,
    MediaCoreFeedbackOverlay,
    MediaCoreLoadingOverlay,
} from "@/app/(main)/_features/media-core/media-core-overlays"
import { MediaCoreTopSectionView } from "@/app/(main)/_features/media-core/media-core-playback-info"
import { mediaCorePreferencesAtom } from "@/app/(main)/_features/media-core/media-core-preferences"
import { startVideoCoreMiniPlayerTransition } from "@/app/(main)/_features/video-core/video-core"
import { useVideoCoreInSight, vc_inSight_open, VideoCoreInSight } from "@/app/(main)/_features/video-core/video-core-in-sight"

import { useVideoCorePlaylist, useVideoCorePlaylistSetup } from "@/app/(main)/_features/video-core/video-core-playlist"
import { vc_formatTime } from "@/app/(main)/_features/video-core/video-core.utils"
import { useWebsocketMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { PlaybackPlayPill } from "@/app/(main)/entry/_containers/torrent-stream/playback-play-pill"
import { clientIdAtom } from "@/app/websocket-provider"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { WSEvents } from "@/lib/server/ws-events"
import { __isDesktop__ } from "@/types/constants"
import type { MpvPrismMpvInitOptions, MpvPrismTrack, MpvPrismTrackSelection } from "@mpv-prism/core"

import { MpvPrismVideo, useMpvPrismEvent, useMpvPrismPlayer } from "@mpv-prism/react"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useAtomValue, useSetAtom } from "jotai"
import React from "react"
import { LuCaptions, LuFilm, LuHeadphones, LuPaintbrush } from "react-icons/lu"
import { RemoveScrollBar } from "react-remove-scroll-bar"
import { toast } from "sonner"

import {
    applyMpvSubtitleSettings,
    createMpvChapterCues,
    createSkipChapterCues,
    isEditableKeyboardTarget,
    mc_cacheBufferedSeconds,
    mc_formatAudioTrack,
    mc_formatSubtitleTrack,
    mc_parseCustomMpvConfig,
    mc_resolveAnime4KProfile,
    mc_resolveSource,
    mc_selectPreferredTrack,
    mc_trackKind,
    mc_trackLabel,
    type MpvCoreEnvelope,
    type MpvCoreNativeChapter,
    normalizeMpvChapterList,
} from "./mpv-core"
import { MpvCoreCastButton } from "./mpv-core-cast-button"
import { MpvCoreFloatingButtons } from "./mpv-core-floating-buttons"

import { MpvCorePreferencesModal, mpvCorePreferencesModalAtom } from "./mpv-core-preferences"
import { MpvCoreScreenshotDirPrompt } from "./mpv-core-screenshot-prompt"
import { MpvCoreSettingsMenu } from "./mpv-core-settings-menu"
import { MpvCoreStats } from "./mpv-core-stats"
import { MpvCoreTimeRange } from "./mpv-core-time-range"
import { MpvCoreTopPlaybackInfo } from "./mpv-core-top-section"
import { MpvCoreWatchPartyChat } from "./mpv-core-watch-party-chat"

import {
    mc_autoNext,
    mc_autoPlay,
    mc_autoSkip,
    mc_buffered,
    mc_buffering,
    mc_cacheState,
    mc_currentTime,
    mc_duration,
    mc_frameDrops,
    mc_highlightOPEDChapters,
    mc_isFullscreen,
    mc_isPip,
    mc_keybindingsAtom,
    mc_overlayFeedback,
    mc_paused,
    mc_pendingScreenshotAtom,
    mc_screenshotPromptOpenAtom,
    mc_settings,
    mc_shaderSettings,
    mc_showChapterMarkers,
    mc_showStats,
    mc_skipData,
    mc_storedMuted,
    mc_storedSpeed,
    mc_storedVolume,
    mc_tracks,
    mpvCore_stateAtom,
} from "./mpv-core.atoms"

type DocumentPictureInPictureApi = {
    requestWindow(options?: { width?: number; height?: number }): Promise<Window>
}

const subtitleExts = ["srt", "ass", "ssa", "vtt", "ttml", "stl", "txt"]

type MpvCorePlayerContentProps = {
    activeMpvConfig: string
    customMpvConfigPath: string | null
    initialDeband: boolean
}

export function MpvCorePlayerInner() {
    const [mpvSettings] = useAtom(mc_settings)
    const [activeMpvConfig] = React.useState(mpvSettings.customMpvConfig)
    const [initialDeband] = React.useState(mpvSettings.deband)
    const [configState, setConfigState] = React.useState({
        ready: !activeMpvConfig.trim(),
        path: null as string | null,
    })

    React.useEffect(() => {
        let cancelled = false

        async function writeConfig() {
            if (!activeMpvConfig.trim()) {
                setConfigState({ ready: true, path: null })
                return
            }

            const writeConfigFile = window.electron?.mpvCore?.writeConfigFile
            if (!writeConfigFile) {
                toast.error("MPV config files are unavailable in this Denshi build")
                setConfigState({ ready: true, path: null })
                return
            }

            setConfigState({ ready: false, path: null })
            try {
                const path = await writeConfigFile(activeMpvConfig)
                if (!cancelled) {
                    setConfigState({ ready: true, path })
                }
            }
            catch (error) {
                if (!cancelled) {
                    toast.error(error instanceof Error ? error.message : "Failed to write MPV config")
                    setConfigState({ ready: true, path: null })
                }
            }
        }

        writeConfig()
        return () => {
            cancelled = true
        }
    }, [activeMpvConfig])

    if (!configState.ready) return null

    return (
        <MpvCorePlayerContent
            activeMpvConfig={activeMpvConfig}
            customMpvConfigPath={configState.path}
            initialDeband={initialDeband}
        />
    )
}

function MpvCorePlayerContent(props: MpvCorePlayerContentProps) {
    const { activeMpvConfig, customMpvConfigPath, initialDeband } = props
    const [state, setState] = useAtom(mpvCore_stateAtom)
    const serverStatus = useServerStatus()
    const qc = useQueryClient()
    const [mpvSettings, setMpvSettings] = useAtom(mc_settings)

    const [playerGeneration, setPlayerGeneration] = React.useState(0)
    const mpvOptions = React.useMemo<MpvPrismMpvInitOptions>(() => {
        const { parsed } = mc_parseCustomMpvConfig(activeMpvConfig)
        const options: NonNullable<MpvPrismMpvInitOptions["options"]> = {
            "keep-open": "yes",
        }
        if (!("hwdec" in parsed)) {
            options["hwdec"] = "auto-safe"
        }
        if (!("deband" in parsed)) {
            options["deband"] = initialDeband ? "yes" : "no"
        }

        const result: MpvPrismMpvInitOptions = {
            options,
            observe: [
                "estimated-vf-fps",
                "display-fps",
                "video-bitrate",
                "audio-bitrate",
                "file-format",
                "hwdec-current",
                "chapter-list",
                "vo-passes",
            ],
        }
        if (customMpvConfigPath) {
            result.config = { files: [customMpvConfigPath] }
        }
        return result
    }, [activeMpvConfig, customMpvConfigPath, initialDeband])
    const expectedPlayerId = `seanime-mpv-core-active-${playerGeneration}`
    const createdPlayer = useMpvPrismPlayer({
        playerId: expectedPlayerId,
        mpv: mpvOptions,
    })
    const player = createdPlayer?.id === expectedPlayerId ? createdPlayer : null

    // Setup playlist hooks
    useVideoCorePlaylistSetup(state as any)
    const { playEpisode, hasNextEpisode, hasPreviousEpisode } = useVideoCorePlaylist()
    const clientId = useAtomValue(clientIdAtom) ?? ""
    const { sendMessage } = useWebsocketSender()
    const [paused, setPaused] = useAtom(mc_paused)
    const [currentTime, setCurrentTime] = useAtom(mc_currentTime)
    const [duration, setDuration] = useAtom(mc_duration)
    const [buffered, setBuffered] = useAtom(mc_buffered)
    const [buffering, setBuffering] = useAtom(mc_buffering)
    const [tracks, setTracks] = useAtom(mc_tracks)
    const [skipData, setSkipData] = useAtom(mc_skipData)
    const [overlayFeedback, setOverlayFeedback] = useAtom(mc_overlayFeedback)
    const [isFullscreen, setIsFullscreen] = useAtom(mc_isFullscreen)
    const [isPip, setIsPip] = useAtom(mc_isPip)
    const [volume, setVolume] = useAtom(mc_storedVolume)
    const [muted, setMuted] = useAtom(mc_storedMuted)
    const [speed, setSpeed] = useAtom(mc_storedSpeed)
    const [autoPlay, setAutoPlay] = useAtom(mc_autoPlay)
    const [autoNext, setAutoNext] = useAtom(mc_autoNext)
    const [autoSkip, setAutoSkip] = useAtom(mc_autoSkip)
    const [showChapterMarkers, setChapterMarkers] = useAtom(mc_showChapterMarkers)
    const [highlightOPEDChapters, setHighlightOPEDChapters] = useAtom(mc_highlightOPEDChapters)
    const [keybindings] = useAtom(mc_keybindingsAtom)
    const [shaderSettings, setShaderSettings] = useAtom(mc_shaderSettings)
    const [showStats, setShowStats] = useAtom(mc_showStats)
    const { toggleOpen: toggleInSight, setData: setInSightData } = useVideoCoreInSight()
    const inSightWasPlayingRef = React.useRef(false)
    const inSightOpen = useAtomValue(vc_inSight_open)
    const cacheState = useAtomValue(mc_cacheState)
    const frameDrops = useAtomValue(mc_frameDrops)
    const setFrameDrops = useSetAtom(mc_frameDrops)
    const setCacheState = useSetAtom(mc_cacheState)
    const preferencesOpen = useAtomValue(mpvCorePreferencesModalAtom)
    const setPreferencesOpen = useSetAtom(mpvCorePreferencesModalAtom)

    // Shared preferences
    const [preferences, setPreferences] = useAtom(mediaCorePreferencesAtom)
    const timestampMode = preferences.timestampMode
    const setTimestampMode = React.useCallback((mode: "elapsed" | "remaining") => {
        setPreferences(prev => ({ ...prev, timestampMode: mode }))
    }, [setPreferences])

    const infoRef = React.useRef<Player_PlaybackInfo | null>(null)
    const sessionTokenRef = React.useRef(0)
    const suppressEndRef = React.useRef(false)
    const completedRef = React.useRef(false)
    const metadataReadyRef = React.useRef(false)
    const canPlayRef = React.useRef(false)
    const currentTimeRef = React.useRef(0)
    const durationRef = React.useRef(0)
    const pausedRef = React.useRef(true)
    const lastSeekEventRef = React.useRef(0)
    const castWasPausedRef = React.useRef(true)
    const lastClickTimeRef = React.useRef(0)
    const miniPlayerEnteredAtRef = React.useRef(0)
    const startupRetryCountRef = React.useRef(0)
    const startupRetryPlaybackIdRef = React.useRef<string | null>(null)
    const terminatingRef = React.useRef(false)
    const closeTimerRef = React.useRef<number | null>(null)
    const resetMiniPlayerTimerRef = React.useRef<number | null>(null)
    const isPipRef = React.useRef(isPip)
    const [containerElement, setContainerElement] = React.useState<HTMLDivElement | null>(null)
    const [isTerminateConfirmOpen, setTerminateConfirmOpen] = React.useState(false)
    const [anime4kDirectory, setAnime4kDirectory] = React.useState<MpvCoreAnime4KDirectory | null>(null)
    const [anime4kError, setAnime4kError] = React.useState<string | null>(null)
    const [diagnostics, setDiagnostics] = React.useState<Record<string, unknown>>({})
    const [nativeChapters, setNativeChapters] = React.useState<MpvCoreNativeChapter[]>([])

    const setPromptOpen = useSetAtom(mc_screenshotPromptOpenAtom)
    const setPendingScreenshot = useSetAtom(mc_pendingScreenshotAtom)
    const { mutateAsync: saveScreenshotMutation } = useVideoCoreSaveScreenshot()

    const overlayFeedbackTimeoutRef = React.useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        return () => {
            if (overlayFeedbackTimeoutRef.current) clearTimeout(overlayFeedbackTimeoutRef.current)
        }
    }, [])

    const showMessage = React.useCallback((message: string, type: "message" | "time" | "icon" = "message", durationMs?: number) => {
        if (overlayFeedbackTimeoutRef.current) {
            clearTimeout(overlayFeedbackTimeoutRef.current)
        }
        setOverlayFeedback({ message, type })
        const duration = durationMs ?? (type === "icon" ? 200 : (pausedRef.current ? 1000 : 500))
        overlayFeedbackTimeoutRef.current = setTimeout(() => {
            setOverlayFeedback(null)
            overlayFeedbackTimeoutRef.current = null
        }, duration)
    }, [setOverlayFeedback])


    const [busy, setBusy] = React.useState(true)
    const [hoveringControlBar, setHoveringControlBar] = React.useState(false)
    const [openMenu, setOpenMenu] = React.useState<string | null>(null)
    const [openSection, setOpenSection] = React.useState<string | null>(null)
    const [openSubSection, setOpenSubSection] = React.useState<string | null>(null)

    const [debouncedMenuOpen, setDebouncedMenuOpen] = React.useState(false)
    React.useEffect(() => {
        if (!!openMenu) {
            setDebouncedMenuOpen(true)
            return
        }
        let t = setTimeout(() => {
            setDebouncedMenuOpen(false)
        }, 800)
        return () => {
            clearTimeout(t)
        }
    }, [openMenu])

    const cursorBusy = hoveringControlBar || !!openMenu

    const setNotBusyTimeout = React.useRef<NodeJS.Timeout | null>(null)
    const lastPointerPosition = React.useRef({ x: 0, y: 0 })
    const busyRef = React.useRef(busy)
    const cursorBusyRef = React.useRef(cursorBusy)

    React.useEffect(() => {
        busyRef.current = busy
    }, [busy])

    React.useEffect(() => {
        cursorBusyRef.current = cursorBusy
    }, [cursorBusy])

    React.useEffect(() => {
        return () => {
            if (setNotBusyTimeout.current) clearTimeout(setNotBusyTimeout.current)
            if (closeTimerRef.current !== null) window.clearTimeout(closeTimerRef.current)
            if (resetMiniPlayerTimerRef.current !== null) window.clearTimeout(resetMiniPlayerTimerRef.current)
        }
    }, [])

    const handleContainerPointerMove = React.useCallback((e: React.PointerEvent<HTMLDivElement>) => {
        const { clientX: x, clientY: y } = e
        const dx = x - lastPointerPosition.current.x
        const dy = y - lastPointerPosition.current.y
        if (Math.abs(dx) < 15 && Math.abs(dy) < 15) return
        if (setNotBusyTimeout.current) clearTimeout(setNotBusyTimeout.current)
        if (!busyRef.current) {
            busyRef.current = true
            setBusy(true)
        }
        setNotBusyTimeout.current = setTimeout(() => {
            if (!cursorBusyRef.current) {
                busyRef.current = false
                setBusy(false)
            }
        }, 1000)
        lastPointerPosition.current = { x, y }
    }, [])

    React.useEffect(() => {
        infoRef.current = state.playbackInfo
    }, [state.playbackInfo])
    React.useEffect(() => {
        currentTimeRef.current = currentTime
    }, [currentTime])
    React.useEffect(() => {
        durationRef.current = duration
    }, [duration])
    React.useEffect(() => {
        pausedRef.current = paused
    }, [paused])
    React.useEffect(() => {
        isPipRef.current = isPip
    }, [isPip])

    // Media Session
    React.useEffect(() => {
        if (!("mediaSession" in navigator) || !state.active || !state.playbackInfo) {
            return
        }

        const info = state.playbackInfo
        const episode = info.episode
        const anime = info.media

        const title = episode?.displayTitle || info.localFile?.name || "Seanime"
        const artist = anime?.title?.userPreferred || anime?.title?.romaji || anime?.title?.english || "Anime"

        const artwork: MediaImage[] = []
        const imageUrl = episode?.episodeMetadata?.image || anime?.coverImage?.large || anime?.coverImage?.medium

        if (imageUrl) {
            artwork.push({ src: imageUrl, sizes: "512x512", type: "image/webp" })
        }

        navigator.mediaSession.metadata = new MediaMetadata({ title, artist, artwork })

        return () => {
            navigator.mediaSession.metadata = null
        }
    }, [state.active, state.playbackInfo])

    React.useEffect(() => {
        if (!("mediaSession" in navigator) || !state.active) {
            return
        }
        navigator.mediaSession.playbackState = paused ? "paused" : "playing"
    }, [state.active, paused])

    React.useEffect(() => {
        if (!("mediaSession" in navigator) || !state.active || !("setPositionState" in navigator.mediaSession)) {
            return
        }
        try {
            navigator.mediaSession.setPositionState({
                duration: duration || 0,
                playbackRate: speed || 1,
                position: currentTime || 0,
            })
        } catch {
            // ignore
        }
    }, [state.active, currentTime, duration, speed])

    React.useEffect(() => {
        if (!("mediaSession" in navigator) || !state.active || !player) {
            return
        }

        const handlePlay = () => {
            player.setPaused(false).catch(() => undefined)
        }
        const handlePause = () => {
            player.setPaused(true).catch(() => undefined)
        }
        const handleSeekForward = () => {
            player.seek(10, "relative+exact").catch(() => undefined)
        }
        const handleSeekBackward = () => {
            player.seek(-10, "relative+exact").catch(() => undefined)
        }
        const handleSeekTo = (details: MediaSessionActionDetails) => {
            if (details.seekTime !== undefined) {
                player.seek(details.seekTime, "absolute+exact").catch(() => undefined)
            }
        }

        navigator.mediaSession.setActionHandler("play", handlePlay)
        navigator.mediaSession.setActionHandler("pause", handlePause)
        navigator.mediaSession.setActionHandler("seekforward", handleSeekForward)
        navigator.mediaSession.setActionHandler("seekbackward", handleSeekBackward)
        try {
            navigator.mediaSession.setActionHandler("seekto", handleSeekTo)
        } catch {
        }

        return () => {
            navigator.mediaSession.setActionHandler("play", null)
            navigator.mediaSession.setActionHandler("pause", null)
            navigator.mediaSession.setActionHandler("seekforward", null)
            navigator.mediaSession.setActionHandler("seekbackward", null)
            try {
                navigator.mediaSession.setActionHandler("seekto", null)
            } catch {
            }
        }
    }, [state.active, player])

    const chapterCues = React.useMemo(() => {
        if (!duration || duration <= 1) return []
        const mpvChapters = createMpvChapterCues(nativeChapters, duration)
        return mpvChapters.length ? mpvChapters : createSkipChapterCues(skipData, duration)
    }, [duration, nativeChapters, skipData])
    const audioTracks = React.useMemo(() => tracks.filter(track => mc_trackKind(track) === "audio"), [tracks])
    const subtitleTracks = React.useMemo(() => tracks.filter(track => mc_trackKind(track) === "subtitle"), [tracks])

    const refreshAnime4KDirectory = React.useCallback(async (requestedDirectory?: string) => {
        if (!window.electron?.mpvCore) return null
        try {
            const result = requestedDirectory
                ? await window.electron.mpvCore.scanAnime4KDirectory(requestedDirectory)
                : await window.electron.mpvCore.getAnime4KDirectory()
            setAnime4kDirectory(result)
            setAnime4kError(null)
            if (result.directory !== shaderSettings.directory) {
                setShaderSettings(current => ({ ...current, directory: result.directory }))
            }
            return result
        }
        catch (error) {
            const message = error instanceof Error ? error.message : String(error)
            setAnime4kDirectory(null)
            setAnime4kError(message)
            return null
        }
    }, [shaderSettings.directory, setShaderSettings])

    React.useEffect(() => {
        refreshAnime4KDirectory(shaderSettings.directory || undefined)
    }, [shaderSettings.directory, refreshAnime4KDirectory])

    const applyShaderSettings = React.useCallback(async (p: typeof player) => {
        if (!p) return
        try {
            if (shaderSettings.mode === "off") {
                await p.clearShaders()
                setAnime4kError(null)
                return
            }
            const directory = anime4kDirectory
                ?? await refreshAnime4KDirectory(shaderSettings.directory || undefined)

            if (shaderSettings.mode === "custom") {
                const selectedPaths = (shaderSettings.customShaders || []).map(name => {
                    const match = directory?.shaders.find(s => s.name === name)
                    return match ? match.path : null
                }).filter((p): p is string => !!p)

                if (selectedPaths.length === 0) {
                    await p.clearShaders()
                    setAnime4kError(null)
                    return
                }
                await p.setShaders(selectedPaths)
                setAnime4kError(null)
                return
            }

            const profile = mc_resolveAnime4KProfile(directory, shaderSettings.anime4kMode, shaderSettings.anime4kQuality)
            if (profile.missing.length) {
                await p.clearShaders()
                setAnime4kError(`Missing ${profile.missing.join(", ")}`)
                return
            }
            await p.setShaders(profile.paths)
            setAnime4kError(null)
        }
        catch (error) {
            setAnime4kError(error instanceof Error ? error.message : String(error))
        }
    }, [
        anime4kDirectory,
        shaderSettings.directory,
        shaderSettings.mode,
        shaderSettings.anime4kMode,
        shaderSettings.anime4kQuality,
        shaderSettings.customShaders,
        refreshAnime4KDirectory,
    ])

    const applyShaderSettingsRef = React.useRef(applyShaderSettings)
    React.useEffect(() => {
        applyShaderSettingsRef.current = applyShaderSettings
    }, [applyShaderSettings])

    React.useEffect(() => {
        applyShaderSettings(player)
    }, [player, applyShaderSettings])

    const sendEvent = React.useCallback((type: string, payload: unknown = {}) => {
        sendMessage({
            type: WSEvents.MPVCORE,
            payload: { clientId, type, payload },
        })
    }, [clientId, sendMessage])

    const statusPayload = React.useCallback(() => ({
        id: infoRef.current?.id ?? "",
        clientId,
        currentTime: currentTimeRef.current,
        duration: durationRef.current,
        paused: pausedRef.current,
    }), [clientId])


    const addSubtitle = React.useCallback(async (track: Player_SubtitleTrack) => {
        if (!player) return
        let path = mc_resolveSource(track.uri || track.sourceUrl)
        if (!path && track.content && window.electron?.mpvCore) {
            path = await window.electron.mpvCore.createTempSubtitle(track.label || `subtitle-${track.index}.srt`, track.content)
        }
        if (!path) return
        await player.runCommand("sub-add", path, "auto", track.label || "External subtitle", track.language || "und")
        setTracks(await player.getTracks().catch(() => []))
    }, [player, setTracks])

    const terminate = React.useCallback(async (reason = "") => {
        if (terminatingRef.current) return
        terminatingRef.current = true
        const info = infoRef.current
        sessionTokenRef.current += 1
        suppressEndRef.current = true
        setTerminateConfirmOpen(false)
        setBuffering(false)
        closePipWindow()
        const finalStatus = statusPayload()
        player?.stop().catch(() => undefined)
        sendEvent("status", finalStatus)
        sendEvent("terminated", {
            id: info?.id ?? "",
            clientId,
            playbackType: info?.playbackType ?? "",
            reason,
        })
        infoRef.current = null
        setState(draft => {
            // if we clear the size and active state at the same time, vaul animates from the wrong rectangle
            // so we keep the mini size set until the close animation finishes.
            draft.miniPlayer = true
            draft.loadingState = "Ending stream..."
            draft.playbackInfo = null
            draft.playbackError = null
        })
        setCurrentTime(0)
        setDuration(0)
        setTracks([])
        setNativeChapters([])

        if (closeTimerRef.current !== null) window.clearTimeout(closeTimerRef.current)
        if (resetMiniPlayerTimerRef.current !== null) window.clearTimeout(resetMiniPlayerTimerRef.current)
        closeTimerRef.current = window.setTimeout(() => {
            setState(draft => {
                draft.active = false
                draft.loadingState = null
            })
            closeTimerRef.current = null

            // reset only after the drawer's close animation has finished
            resetMiniPlayerTimerRef.current = window.setTimeout(() => {
                setState(draft => {
                    draft.miniPlayer = false
                })
                terminatingRef.current = false
                resetMiniPlayerTimerRef.current = null
            }, 550)
        }, 700)
    }, [clientId, player, sendEvent, setBuffering, setCurrentTime, setDuration, setState, setTracks])

    useWebsocketMessageListener<MpvCoreEnvelope>({
        type: WSEvents.MPVCORE,
        onMessage: ({ type, payload }) => {
            switch (type) {
                case "open-and-await":
                    player?.stop().catch(() => undefined)
                    if (closeTimerRef.current !== null) {
                        window.clearTimeout(closeTimerRef.current)
                        closeTimerRef.current = null
                    }
                    if (resetMiniPlayerTimerRef.current !== null) {
                        window.clearTimeout(resetMiniPlayerTimerRef.current)
                        resetMiniPlayerTimerRef.current = null
                    }
                    terminatingRef.current = false
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = String(payload || "Preparing stream...")
                        draft.playbackInfo = null
                        draft.playbackError = null
                    })
                    break
                case "abort-open":
                    if (payload) {
                        setBuffering(false)
                        setState(draft => {
                            draft.playbackError = String(payload)
                            draft.loadingState = null
                        })
                    } else {
                        terminate("open aborted")
                    }
                    break
                case "watch":
                    if (closeTimerRef.current !== null) {
                        window.clearTimeout(closeTimerRef.current)
                        closeTimerRef.current = null
                    }
                    if (resetMiniPlayerTimerRef.current !== null) {
                        window.clearTimeout(resetMiniPlayerTimerRef.current)
                        resetMiniPlayerTimerRef.current = null
                    }
                    terminatingRef.current = false
                    setState(draft => {
                        draft.active = true
                        draft.miniPlayer = false
                        draft.loadingState = "Loading..."
                        draft.playbackInfo = payload as Player_PlaybackInfo
                        draft.playbackError = null
                    })
                    break
                case "stream-error":
                    setBuffering(false)
                    setState(draft => {
                        draft.playbackError = (payload as { error?: string })?.error || "Stream error"
                        draft.loadingState = null
                    })
                    break
                case "pause":
                    player?.pause()
                    break
                case "resume":
                    player?.play()
                    break
                case "seek":
                    player?.seek(Number(payload), "relative+exact")
                    break
                case "seek-to":
                    player?.seek(Number(payload), "absolute+exact")
                    break
                case "terminate":
                    terminate("server termination")
                    break
                case "set-fullscreen":
                    toggleFullscreen(Boolean(payload))
                    break
                case "set-pip":
                    togglePip(Boolean(payload))
                    break
                case "set-audio-track":
                    player?.selectTrack("audio", payload as MpvPrismTrackSelection)
                    break
                case "set-subtitle-track":
                    player?.selectTrack("subtitle", payload as MpvPrismTrackSelection)
                    break
                case "add-subtitle-track":
                    addSubtitle(payload as Player_SubtitleTrack)
                    break
                case "show-message": {
                    const value = payload as { message: string, duration?: number }
                    showMessage(value.message, "message", value.duration ?? 2200)
                    break
                }
                case "get-status":
                    sendEvent("status", statusPayload())
                    break
                case "get-playlist":
                    sendEvent("playlist-state", { playlist: null })
                    break
                case "get-skip-data":
                    sendEvent("skip-data", { skipData })
                    break
                case "set-skip-data":
                    setSkipData(payload as Player_SkipData | null)
                    break
                case "play-playlist-episode":
                    playEpisode(payload as string)
                    break
                case "in-sight-data":
                    setInSightData((payload ?? null) as any)
                    break
            }
        },
        deps: [player, skipData, playEpisode],
    })

    React.useEffect(() => {
        const info = state.playbackInfo
        if (!player || !info || !state.active) return
        const token = ++sessionTokenRef.current
        completedRef.current = false
        if (startupRetryPlaybackIdRef.current !== info.id) {
            startupRetryPlaybackIdRef.current = info.id
            startupRetryCountRef.current = 0
        }
        metadataReadyRef.current = false
        canPlayRef.current = false
        suppressEndRef.current = true
        setCurrentTime(0)
        setDuration(0)
        setTracks([])
        setNativeChapters([])
        setBuffering(true)
        setDiagnostics({})
        setFrameDrops({})
        setCacheState(null);

        (async () => {
            try {
                await player.awaitPresentationReady()
                if (token !== sessionTokenRef.current) return
                await player.stop().catch(() => undefined)
                if (token !== sessionTokenRef.current) return
                suppressEndRef.current = false
                await player.load(mc_resolveSource(info.playbackUri))
                if (token !== sessionTokenRef.current) return
                sendEvent("playback-loaded", { id: info.id, clientId })
                await Promise.all([
                    player.setVolume(volume * 100),
                    player.setMute(muted),
                    player.setSpeed(speed),
                    applyMpvSubtitleSettings(player, mpvSettings),
                    applyShaderSettingsRef.current(player).catch(() => undefined),
                ])
                if (!autoPlay || info.initialState?.paused) {
                    await player.pause()
                }
            }
            catch (error) {
                if (token !== sessionTokenRef.current) return
                const message = error instanceof Error ? error.message : String(error)
                const startupError = (
                    message.includes("player is not registered for this renderer") ||
                    message.includes("native player must be attached before playback commands") ||
                    message.includes("video presenter did not attach before playback commands")
                )
                if (startupError && startupRetryCountRef.current < 2) {
                    startupRetryCountRef.current += 1
                    setBuffering(true)
                    setPlayerGeneration(current => current + 1)
                    return
                }
                setBuffering(false)
                setState(draft => {
                    draft.playbackError = message
                    draft.loadingState = null
                })
                sendEvent("player-error", { error: message })
                suppressEndRef.current = true
                await player.stop().catch(() => undefined)
            }
        })()
    }, [player, state.playbackInfo?.id])

    useMpvPrismEvent(player, "position", event => {
        const value = event.position ?? 0
        setCurrentTime(value)
        if (autoSkip && skipData) {
            for (const entry of [skipData.op, skipData.ed]) {
                if (entry && value >= entry.interval.startTime && value < entry.interval.endTime) {
                    player?.seek(entry.interval.endTime, "absolute+exact")
                    break
                }
            }
        }
        const total = durationRef.current
        if (!completedRef.current && total > 0 && value / total >= 0.8) {
            completedRef.current = true
            sendEvent("completed", { ...statusPayload(), currentTime: value })
        }
    })
    useMpvPrismEvent(player, "duration", event => setDuration(event.duration ?? 0))
    useMpvPrismEvent(player, "paused", event => {
        setPaused(event.paused)
        if (!metadataReadyRef.current) return
        sendEvent(event.paused ? "paused" : "resumed", { ...statusPayload(), paused: event.paused })
    })
    useMpvPrismEvent(player, "speed", event => {
        if (!metadataReadyRef.current) return
        if (event.speed != null) setSpeed(event.speed)
    })
    useMpvPrismEvent(player, "volume", event => {
        if (!metadataReadyRef.current) return
        if (event.volume != null) setVolume(Math.max(0, Math.min(1, event.volume / 100)))
    })
    useMpvPrismEvent(player, "mute", event => {
        if (!metadataReadyRef.current) return
        setMuted(event.muted)
    })
    useMpvPrismEvent(player, "tracks", event => setTracks(event.tracks))
    useMpvPrismEvent(player, "trackSelection", event => {
        if (event.kind === "audio") sendEvent("audio-track-changed", { trackId: event.id })
        if (event.kind === "subtitle") sendEvent("subtitle-track-changed", { trackId: event.id })
    })
    useMpvPrismEvent(player, "cache", event => {
        const value = event.state as Record<string, unknown> | number | null
        const isBuffering = typeof value === "number"
            ? value > 0
            : Boolean(value && (value["underrun"] || Number(value["cache-buffering-state"]) > 0))
        setBuffering(isBuffering)
        setBuffered(mc_cacheBufferedSeconds(event.state, durationRef.current, currentTimeRef.current))
    })
    useMpvPrismEvent(player, "cache", event => setCacheState(event.state))
    useMpvPrismEvent(player, "frameDrops", event => {
        setFrameDrops(current => ({ ...current, [event.name]: event.value ?? 0 }))
    })
    useMpvPrismEvent(player, "property", event => {
        if (event.name === "chapter-list") {
            setNativeChapters(normalizeMpvChapterList(event.value))
            return
        }
        if (!diagnosticsProperties.has(event.name)) return
        setDiagnostics(current => ({ ...current, [event.name]: event.value }))
    })
    useMpvPrismEvent(player, "state", event => {
        if (event.state === "file-loaded") {
            const token = sessionTokenRef.current;
            (async () => {
                const info = infoRef.current
                if (!player || !info) return
                const [nextDuration, nextPosition, nextTracks, nextChapters] = await Promise.all([
                    player.getProperty<number>("duration").catch(() => 0),
                    player.getProperty<number>("time-pos").catch(() => 0),
                    player.getTracks().catch(() => []),
                    player.getProperty<unknown>("chapter-list").catch(() => []),
                ])
                if (token !== sessionTokenRef.current) return
                setDuration(Number(nextDuration) || 0)
                setCurrentTime(Number(nextPosition) || 0)
                setNativeChapters(normalizeMpvChapterList(nextChapters))
                const finalTracks = nextTracks
                setTracks(finalTracks)
                const preferredAudio = mc_selectPreferredTrack(
                    finalTracks,
                    "audio",
                    mpvSettings.preferredAudioLanguage,
                )
                const preferredSubtitle = mc_selectPreferredTrack(
                    finalTracks,
                    "subtitle",
                    mpvSettings.preferredSubtitleLanguage,
                    mpvSettings.preferredSubtitleBlacklist,
                )
                const restoreTime = info.initialState?.currentTime
                await Promise.all([
                    preferredAudio?.id != null ? player.selectTrack("audio", preferredAudio.id).catch(() => undefined) : Promise.resolve(),
                    preferredSubtitle?.id != null ? player.selectTrack("subtitle", preferredSubtitle.id).catch(() => undefined) : Promise.resolve(),
                    typeof restoreTime === "number" && restoreTime > 0 ? player.seek(restoreTime, "absolute+exact").catch(() => undefined) : Promise.resolve(),
                    player.setVolume(volume * 100).catch(() => undefined),
                    player.setMute(muted).catch(() => undefined),
                    player.setSpeed(speed).catch(() => undefined),
                    applyMpvSubtitleSettings(player, mpvSettings).catch(() => undefined),
                    applyShaderSettingsRef.current(player).catch(() => undefined),
                ])
                metadataReadyRef.current = true
                setState(draft => {
                    draft.loadingState = null
                })
                sendEvent("loaded-metadata", {
                    id: info.id,
                    clientId,
                    currentTime: restoreTime ?? Number(nextPosition) ?? 0,
                    duration: Number(nextDuration) || 0,
                    paused: player.paused,
                })
            })()
        }
        if (event.state === "playback-restart") {
            setBuffering(false)
            if (!canPlayRef.current && infoRef.current) {
                canPlayRef.current = true
                sendEvent("can-play", statusPayload())
            }
            const now = Date.now()
            if (now - lastSeekEventRef.current > 250 && metadataReadyRef.current) {
                lastSeekEventRef.current = now
                sendEvent("seeked", statusPayload())
            }
        }
    })
    useMpvPrismEvent(player, "ended", event => {
        if (suppressEndRef.current) {
            suppressEndRef.current = false
            return
        }
        if (event.error) {
            setBuffering(false)
            sendEvent("player-error", { error: event.error })
            return
        }
        if ((event.reason ?? "").toLowerCase() !== "eof") return
        sendEvent("ended", { autoNext })
    })
    useMpvPrismEvent(player, "error", event => {
        setBuffering(false)
        setState(draft => {
            draft.playbackError = event.message
            draft.loadingState = null
        })
        sendEvent("player-error", { error: event.message })
    })

    useMpvPrismEvent(player, "pip", event => {
        setIsPip(event.enabled)
        isPipRef.current = event.enabled
        sendEvent("pip-changed", { pip: event.enabled })
    })

    React.useEffect(() => {
        if (!state.active || !state.playbackInfo) return
        const interval = window.setInterval(() => sendEvent("status", statusPayload()), 1000)
        return () => window.clearInterval(interval)
    }, [state.active, state.playbackInfo?.id, sendEvent, statusPayload])

    React.useEffect(() => {
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistory.key] })
        qc.invalidateQueries({ queryKey: [API_ENDPOINTS.CONTINUITY.GetContinuityWatchHistoryItem.key] })
    }, [state.playbackInfo?.id])

    React.useEffect(() => {
        if (!player || !state.active || !metadataReadyRef.current) return
        applyMpvSubtitleSettings(player, mpvSettings).catch(error => {
            toast.error(error instanceof Error ? error.message : "Failed to apply subtitle settings")
        })
    }, [player, state.active, mpvSettings])

    React.useEffect(() => {
        if (!player || !state.active) return
        const { parsed } = mc_parseCustomMpvConfig(activeMpvConfig)
        if ("deband" in parsed) {
            return
        }
        player.setProperty("deband", mpvSettings.deband ? "yes" : "no").catch(() => undefined)
    }, [player, state.active, mpvSettings.deband, activeMpvConfig])

    React.useEffect(() => {
        if (!player || !state.active) return
        if (inSightOpen) {
            inSightWasPlayingRef.current = !pausedRef.current
            if (inSightWasPlayingRef.current) player.setPaused(true)
        } else {
            if (inSightWasPlayingRef.current) player.setPaused(false)
        }
    }, [inSightOpen, player, state.active])

    React.useEffect(() => {
        const remove = window.electron?.on?.("window:fullscreen", (value: boolean) => {
            setIsFullscreen(value)
            sendEvent("fullscreen-changed", { fullscreen: value })
        })
        window.electron?.window.isFullscreen().then(setIsFullscreen).catch(() => undefined)
        return remove
    }, [sendEvent, setIsFullscreen])

    React.useEffect(() => {
        if (!state.active) closePipWindow()
    }, [state.active])

    React.useEffect(() => {
        if (!state.active) return
        const onKeyDown = async (event: KeyboardEvent) => {
            if (
                event.defaultPrevented ||
                preferencesOpen ||
                isEditableKeyboardTarget(event.target) ||
                isEditableKeyboardTarget(document.activeElement) ||
                event.ctrlKey ||
                event.shiftKey ||
                event.altKey ||
                event.metaKey
            ) return
            if (!player) return

            if (event.code === "Escape" && isFullscreen) {
                event.preventDefault()
                event.stopPropagation()
                await toggleFullscreen(false)
                return
            }

            if (state.miniPlayer) {
                if (event.code === keybindings.fullscreen.key) {
                    event.preventDefault()
                    startVideoCoreMiniPlayerTransition(() => {
                        setState(draft => {
                            draft.miniPlayer = false
                        })
                    })
                    window.setTimeout(() => toggleFullscreen(true), 0)
                }
                return
            }

            if (event.code === "Space" || event.code === "Enter") {
                event.preventDefault()
                const nextPaused = !pausedRef.current
                await player.setPaused(nextPaused)
                showMessage(nextPaused ? "PAUSE" : "PLAY", "icon")
                return
            }
            if (event.code === "Home" || event.code === "End") {
                event.preventDefault()
                const target = event.code === "Home" ? 0 : durationRef.current
                await player.seek(target, "absolute+exact")
                showMessage(event.code === "Home" ? "Beginning" : "End")
                return
            }
            if (/^Digit[0-9]$/.test(event.code)) {
                event.preventDefault()
                const percentage = Number(event.code.slice(-1)) * 10
                const target = durationRef.current * percentage / 100
                await player.seek(target, "absolute+exact")
                showMessage(`${vc_formatTime(target)} / ${vc_formatTime(durationRef.current)}`)
                return
            }
            if (event.code === "Comma" || event.code === "Period") {
                event.preventDefault()
                await player.seek(event.code === "Comma" ? -1 / 24 : 1 / 24, "relative+exact")
                showMessage(event.code === "Comma" ? "Previous Frame" : "Next Frame")
                return
            }

            const seekRelative = async (value: number) => {
                await player.seek(value, "relative+exact")
                const target = Math.min(durationRef.current, Math.max(0, currentTimeRef.current + value))
                showMessage(`${vc_formatTime(target)} / ${vc_formatTime(durationRef.current)}`)
            }
            const cycleTrack = async (kind: "audio" | "subtitle", candidates: MpvPrismTrack[]) => {
                if (!candidates.length) return
                const selectedIndex = candidates.findIndex(track => track.selected)
                if (kind === "subtitle" && selectedIndex === candidates.length - 1) {
                    await player.selectTrack("subtitle", "no")
                    showMessage("Subtitles: Off")
                    return
                }
                const next = candidates[selectedIndex < 0 ? 0 : (selectedIndex + 1) % candidates.length]
                await player.selectTrack(kind, next.id)
                showMessage(`${kind === "audio" ? "Audio" : "Subtitles"}: ${mc_trackLabel(next)}`)
            }

            if (event.code === keybindings.seekForward.key) {
                event.preventDefault()
                const interval = [skipData?.op?.interval, skipData?.ed?.interval]
                    .find(value => value && currentTimeRef.current >= value.startTime && currentTimeRef.current < value.endTime)
                if (interval) {
                    await player.seek(interval.endTime, "absolute+exact")
                    showMessage("Skipped chapter")
                } else {
                    await seekRelative(keybindings.seekForward.value)
                }
            } else if (event.code === keybindings.seekBackward.key) {
                event.preventDefault()
                await seekRelative(-keybindings.seekBackward.value)
            } else if (event.code === keybindings.seekForwardFine.key) {
                event.preventDefault()
                await seekRelative(keybindings.seekForwardFine.value)
            } else if (event.code === keybindings.seekBackwardFine.key) {
                event.preventDefault()
                await seekRelative(-keybindings.seekBackwardFine.value)
            } else if (event.code === keybindings.nextChapter.key) {
                event.preventDefault()
                const next = chapterCues.find(chapter => chapter.startTime > currentTimeRef.current + 1)
                if (next) {
                    await player.seek(next.startTime, "absolute+exact")
                    showMessage(next.text ? `Chapter: ${next.text}` : "Next chapter")
                } else {
                    await player.seek(durationRef.current, "absolute+exact")
                    showMessage("End of chapters")
                }
            } else if (event.code === keybindings.previousChapter.key) {
                event.preventDefault()
                const currentIndex = chapterCues.findIndex((chapter, index) => (
                    chapter.startTime <= currentTimeRef.current &&
                    (!chapterCues[index + 1] || currentTimeRef.current < chapterCues[index + 1].startTime)
                ))
                const previous = currentIndex > 0 ? chapterCues[currentIndex - 1] : chapterCues[0]
                await player.seek(previous?.startTime ?? 0, "absolute+exact")
                showMessage(previous?.text ? `Chapter: ${previous.text}` : "Beginning")
            } else if (event.code === keybindings.volumeUp.key) {
                event.preventDefault()
                const next = Math.min(1, volume + keybindings.volumeUp.value / 100)
                setVolume(next)
                await player.setVolume(next * 100)
            } else if (event.code === keybindings.volumeDown.key) {
                event.preventDefault()
                const next = Math.max(0, volume - keybindings.volumeDown.value / 100)
                setVolume(next)
                await player.setVolume(next * 100)
            } else if (event.code === keybindings.mute.key) {
                event.preventDefault()
                setMuted(!muted)
                await player.setMute(!muted)
            } else if (event.code === keybindings.cycleSubtitles.key) {
                event.preventDefault()
                await cycleTrack("subtitle", subtitleTracks)
            } else if (event.code === keybindings.cycleAudio.key) {
                event.preventDefault()
                await cycleTrack("audio", audioTracks)
            } else if (event.code === keybindings.nextEpisode.key && hasNextEpisode) {
                event.preventDefault()
                playEpisode("next")
            } else if (event.code === keybindings.previousEpisode.key && hasPreviousEpisode) {
                event.preventDefault()
                playEpisode("previous")
            } else if (event.code === keybindings.fullscreen.key) {
                event.preventDefault()
                toggleFullscreen()
            } else if (event.code === keybindings.pictureInPicture.key) {
                event.preventDefault()
                togglePip()
            } else if (event.code === keybindings.takeScreenshot.key) {
                event.preventDefault()
                takeScreenshot()
            } else if (event.code === keybindings.statsForNerds.key) {
                event.preventDefault()
                setShowStats(value => !value)
            } else if (event.code === keybindings.increaseSpeed.key) {
                event.preventDefault()
                changeSpeed(Math.min(8, speed + keybindings.increaseSpeed.value))
            } else if (event.code === keybindings.decreaseSpeed.key) {
                event.preventDefault()
                changeSpeed(Math.max(0.2, speed - keybindings.decreaseSpeed.value))
            } else if (event.code === keybindings.openInSight.key) {
                event.preventDefault()
                toggleInSight()
            }
        }
        window.addEventListener("keydown", onKeyDown, true)
        return () => window.removeEventListener("keydown", onKeyDown, true)
    }, [
        state.active,
        state.miniPlayer,
        player,
        muted,
        speed,
        volume,
        keybindings,
        chapterCues,
        skipData,
        audioTracks,
        subtitleTracks,
        isFullscreen,
        preferencesOpen,
        hasNextEpisode,
        hasPreviousEpisode,
        playEpisode,
        setMuted,
        setShowStats,
        setState,
        setVolume,
        showMessage,
        toggleInSight,
    ])

    const toggleFullscreen = React.useCallback(async (force?: boolean) => {
        const next = force ?? !isFullscreen
        window.electron?.window.setFullscreen(next)
    }, [isFullscreen])

    const handlePlayerSurfaceClick = React.useCallback(() => {
        const now = Date.now()
        if (!debouncedMenuOpen) {
            const nextPaused = !pausedRef.current
            player?.setPaused(nextPaused)
            showMessage(nextPaused ? "PAUSE" : "PLAY", "icon")
        }
        if (lastClickTimeRef.current && now - lastClickTimeRef.current < 300) {
            toggleFullscreen()
        } else {
            window.setTimeout(() => setBusy(false), 100)
        }
        lastClickTimeRef.current = now
    }, [debouncedMenuOpen, player, showMessage, toggleFullscreen])

    async function togglePip(force?: boolean) {
        if (!player) return
        const next = force ?? !player.isPip
        try {
            if (next) {
                await player.enterPip()
            } else {
                await player.exitPip()
            }
        } catch (error) {
            console.error("Failed to toggle PiP:", error)
            toast.error(error instanceof Error ? error.message : "Failed to toggle PiP")
        }
    }

    async function closePipWindow() {
        if (!player) return
        try {
            if (player.isPip) {
                await player.exitPip()
            }
        } catch (error) {
            console.error("Failed to close PiP:", error)
        }
    }

    async function takeScreenshot() {
        if (!player || !window.electron?.mpvCore) return
        try {
            const videoCanvas = document.querySelector<HTMLCanvasElement>("[data-mpv-prism-video-canvas]")
            const video = document.querySelector<HTMLVideoElement>("[data-mpv-prism-video]")
            const targetElement = videoCanvas || video

            if (!targetElement) {
                toast.error("No video element found to capture")
                return
            }

            const canvas = document.createElement("canvas")
            let width = 0
            let height = 0

            if (videoCanvas) {
                width = videoCanvas.width
                height = videoCanvas.height
            } else if (video) {
                width = video.videoWidth || 1920
                height = video.videoHeight || 1080
            }

            canvas.width = width
            canvas.height = height

            const ctx = canvas.getContext("2d")
            if (!ctx) {
                toast.error("Failed to create canvas context")
                return
            }

            // Detect if the video/canvas is flipped vertically (common in mpv-prism shared texture mode)
            const style = window.getComputedStyle(targetElement)
            const transform = style.transform || ""
            const isFlipped = transform.includes("matrix")
                ? transform.split(",")[3]?.trim().startsWith("-")
                : transform.includes("scaleY(-1)")

            if (isFlipped) {
                ctx.translate(0, height)
                ctx.scale(1, -1)
            }

            ctx.drawImage(targetElement, 0, 0, width, height)

            const dataUrl = canvas.toDataURL("image/png")
            const base64Data = dataUrl.replace(/^data:image\/png;base64,/, "")

            const screenshotDir = serverStatus?.settings?.mediaPlayer?.screenshotDir

            if (!screenshotDir) {
                setPendingScreenshot({ base64Data })
                setPromptOpen(true)
                return
            }

            const filename = `seanime_screenshot_${new Date().getTime()}.png`
            await saveScreenshotMutation({
                dir: screenshotDir,
                filename,
                base64Data,
            })

            showMessage(`Screenshot saved to ${screenshotDir}`, "message", 4000)
        } catch (error) {
            console.error("Screenshot capture failed:", error)
            toast.error(error instanceof Error ? error.message : "Failed to capture screenshot")
        }
    }

    async function changeSpeed(value: number) {
        setSpeed(value)
        await player?.setSpeed(value)
        showMessage(`Speed: ${value.toFixed(2)}x`)
    }

    async function addSubtitleFile(file: File) {
        const extension = file.name.split(".").pop()?.toLowerCase() ?? ""
        if (!subtitleExts.includes(extension)) {
            toast.error("Unsupported subtitle format")
            return
        }
        const content = await file.text()
        if (!window.electron?.mpvCore) return
        const path = await window.electron.mpvCore.createTempSubtitle(file.name, content)
        await player?.runCommand("sub-add", path, "select", file.name)
        setTracks(await player?.getTracks().catch(() => []) ?? [])
        showMessage(`Loaded subtitle ${file.name}`)
    }

    function handleDrop(event: React.DragEvent) {
        event.preventDefault()
        const file = event.dataTransfer.files[0]
        if (file) addSubtitleFile(file)
    }

    function handlePaste(event: React.ClipboardEvent) {
        const file = event.clipboardData.files[0]
        if (file) {
            addSubtitleFile(file)
            return
        }
        const content = event.clipboardData.getData("text/plain")
        if (!content || (!content.includes("-->") && !content.startsWith("WEBVTT"))) return
        const fileName = content.startsWith("WEBVTT") ? "pasted-subtitle.vtt" : "pasted-subtitle.srt"
        addSubtitleFile(new File([content], fileName, { type: "text/plain" }))
    }

    const onVolumeChange = React.useCallback((vol: number) => {
        setVolume(vol)
        player?.setVolume(vol * 100)
        if (muted && vol > 0) {
            setMuted(false)
            player?.setMute(false)
        }
    }, [player, muted, setVolume, setMuted])

    const onMuteToggle = React.useCallback(() => {
        const nextMuted = !muted
        setMuted(nextMuted)
        player?.setMute(nextMuted)
    }, [player, muted, setMuted])

    const selectedAudio = audioTracks.find(track => track.selected)?.id ?? ""
    const selectedSubtitle = subtitleTracks.find(track => track.selected)?.id ?? "no"
    const setMiniPlayer = React.useCallback((value: boolean) => {
        if (value) miniPlayerEnteredAtRef.current = Date.now()
        setTerminateConfirmOpen(false)
        startVideoCoreMiniPlayerTransition(() => {
            setState(draft => {
                draft.miniPlayer = value
            })
        })
    }, [setState])

    const hasPlayback = !!state.playbackInfo && !state.loadingState

    const diagnosticsProperties = React.useMemo(() => new Set([
        "video-params",
        "audio-params",
        "estimated-vf-fps",
        "display-fps",
        "video-bitrate",
        "audio-bitrate",
        "file-format",
        "hwdec-current",
        "avsync",
        "vo-passes",
    ]), [])

    return (
        <>
            {state.active && !state.miniPlayer && <RemoveScrollBar />}

            <MpvCorePreferencesModal
                fullscreen={isFullscreen}
                containerElement={containerElement}
                onTerminate={terminate}
            />

            <MpvCoreScreenshotDirPrompt
                isFullscreen={isFullscreen}
                containerElement={containerElement}
            />



            <MediaCoreDrawer
                open={state.active}
                onOpenChange={open => {
                    if (!open) {
                        if (!state.miniPlayer) {
                            setMiniPlayer(true)
                        } else {
                            terminate("drawer closed")
                        }
                    }
                }}
                borderToBorder
                miniPlayer={state.miniPlayer}
                size={state.miniPlayer ? "md" : "full"}
                side={state.miniPlayer ? "right" : "bottom"}
                contentClass={cn(
                    "p-0 m-0 bg-black border-0 overflow-hidden",
                    !state.miniPlayer && "h-full",
                )}
                allowOutsideInteraction
                overlayClass={cn(state.miniPlayer && "hidden")}
                hideCloseButton
                closeClass={cn(
                    "z-[99]",
                    __isDesktop__ && !state.miniPlayer && "top-8",
                    state.miniPlayer && "left-4",
                )}
                data-native-player-drawer
                onMiniPlayerClick={() => player?.setPaused(!pausedRef.current)}
                onEscapeKeyDown={event => {
                    event.preventDefault()
                    event.stopPropagation()
                    if (isFullscreen) {
                        toggleFullscreen(false)
                    } else if (state.miniPlayer) {
                        if (Date.now() - miniPlayerEnteredAtRef.current < 400 || event.repeat) return
                        setTerminateConfirmOpen(true)
                    } else {
                        setMiniPlayer(true)
                    }
                }}
            >
                <div
                    data-vc-element="container"
                    ref={setContainerElement}
                    className={cn(
                        "relative w-full h-full bg-black overflow-clip flex items-center justify-center text-white select-none outline-none focus:outline-none",
                        (!busy && !state.miniPlayer) && "cursor-none",
                    )}
                    onDrop={handleDrop}
                    onDragOver={event => event.preventDefault()}
                    onPaste={handlePaste}
                    onPointerMove={handleContainerPointerMove}
                    tabIndex={0}
                >
                    <MpvPrismVideo
                        player={player}
                        className="absolute inset-0 h-full w-full"
                        fit="contain"
                        frameTransport="auto"
                        presentationMode="canvas"
                        lowLatency
                        overlayStyle={{ pointerEvents: "auto", zIndex: "auto" }}
                    >


                        <MediaCoreErrorOverlay
                            playbackError={state.playbackError}
                            isMiniPlayer={state.miniPlayer}
                            onClose={() => terminate("error closed")}
                        />

                        {hasPlayback ? (
                            <>
                                <div
                                    data-vc-element="inner-container"
                                    className="absolute inset-0 z-[1]"
                                    onClick={handlePlayerSurfaceClick}
                                    onContextMenu={event => event.preventDefault()}
                                />

                                {showStats && !state.miniPlayer && (
                                    <MpvCoreStats
                                        info={state.playbackInfo}
                                        tracks={tracks}
                                        cache={cacheState}
                                        frameDrops={frameDrops}
                                        diagnostics={diagnostics}
                                        currentTime={currentTime}
                                        duration={duration}
                                        buffered={buffered}
                                        speed={speed}
                                        buffering={buffering}
                                        containerElement={containerElement}
                                        shaderMode={shaderSettings.mode}
                                        anime4kMode={shaderSettings.anime4kMode}
                                        anime4kQuality={shaderSettings.anime4kQuality}
                                        customShadersCount={shaderSettings.customShaders?.length || 0}
                                    />
                                )}

                                {overlayFeedback && (
                                    <MediaCoreFeedbackOverlay
                                        feedback={overlayFeedback}
                                        isMiniPlayer={state.miniPlayer}
                                    />
                                )}

                                <MediaCoreBufferingOverlay buffering={buffering && !state.playbackError} />

                                <MediaCoreTopSectionView
                                    inline={false}
                                    fullscreen={isFullscreen}
                                    isMiniPlayer={state.miniPlayer}
                                    showTopSection={busy || paused || hoveringControlBar}
                                    paused={paused}
                                >
                                    <MpvCoreTopPlaybackInfo
                                        playbackInfo={state.playbackInfo}
                                        isMiniPlayer={state.miniPlayer}
                                        paused={paused}
                                        hoveringControlBar={hoveringControlBar}
                                        toggleFullscreen={toggleFullscreen}
                                        setMiniPlayer={setMiniPlayer}
                                    />
                                    <div
                                        data-vc-element="floating-buttons-container"
                                        className={cn(
                                            "opacity-0 transition-opacity duration-200 ease-in-out",
                                            (busy || paused) && "opacity-100",
                                        )}
                                    >
                                        <MpvCoreFloatingButtons
                                            part="video"
                                            fullscreen={isFullscreen}
                                            miniPlayer={state.miniPlayer}
                                            onEnterMiniPlayer={() => setMiniPlayer(true)}
                                            onExpand={() => setMiniPlayer(false)}
                                            onTerminate={() => terminate("user terminated player")}
                                            onExitFullscreen={() => toggleFullscreen(false)}
                                        />
                                    </div>
                                </MediaCoreTopSectionView>

                                {isPip && (
                                    <div
                                        data-vc-element="pip-overlay"
                                        className="absolute top-0 left-0 w-full h-full z-[100] bg-black flex items-center justify-center"
                                    >
                                        <Button intent="gray-outline" size="xl" onClick={() => togglePip(false)}>
                                            Exit PiP
                                        </Button>
                                    </div>
                                )}

                                <MediaCoreControlBarView
                                    paused={paused}
                                    isMiniPlayer={state.miniPlayer}
                                    cursorBusy={cursorBusy}
                                    hoveringControlBar={hoveringControlBar}
                                    onHoveringControlBarChange={setHoveringControlBar}
                                    containerElement={containerElement}
                                    isMobile={false}
                                    timeRange={
                                        <MpvCoreTimeRange
                                            player={player}
                                            currentTime={currentTime}
                                            duration={duration}
                                            buffered={buffered}
                                            showChapterMarkers={showChapterMarkers}
                                            highlightOPEDChapters={highlightOPEDChapters}
                                            paused={paused}
                                            chapters={chapterCues}
                                            streamUrl={state.playbackInfo?.streamUrl}
                                            playbackType={state.playbackInfo?.playbackType}
                                            isMiniPlayer={state.miniPlayer}
                                        />
                                    }
                                >
                                    <MediaCorePlayButton
                                        paused={paused}
                                        onTogglePlay={() => player?.setPaused(!paused)}
                                        isMobile={false}
                                        isMiniPlayer={state.miniPlayer}
                                    />
                                    {hasPreviousEpisode && (
                                        <MediaCorePreviousButton
                                            onClick={() => playEpisode("previous")}
                                            isMobile={false}
                                            isMiniPlayer={state.miniPlayer}
                                        />
                                    )}
                                    {hasNextEpisode && (
                                        <MediaCoreNextButton
                                            onClick={() => playEpisode("next")}
                                            isMobile={false}
                                            isMiniPlayer={state.miniPlayer}
                                        />
                                    )}
                                    <MediaCoreVolumeButton
                                        volume={volume}
                                        muted={muted}
                                        onVolumeChange={onVolumeChange}
                                        onMuteToggle={onMuteToggle}
                                        isMobile={false}
                                        isMiniPlayer={state.miniPlayer}
                                    />
                                    <MediaCoreTimestamp
                                        currentTime={currentTime}
                                        duration={duration}
                                        timestampMode={timestampMode}
                                        onTimestampModeToggle={() => setTimestampMode(timestampMode === "elapsed" ? "remaining" : "elapsed")}
                                        isMobile={false}
                                    />

                                    <div className="flex flex-1" data-vc-element="control-bar-separator" />

                                    <PlaybackPlayPill
                                        isNativePlayerComponent="control-bar"
                                        show={!state.miniPlayer}
                                    />

                                    {!state.miniPlayer && (
                                        <>
                                            <MpvCoreWatchPartyChat
                                                isMiniPlayer={state.miniPlayer}
                                                isFullscreen={isFullscreen}
                                                containerElement={containerElement}
                                                openMenu={openMenu}
                                                setOpenMenu={setOpenMenu}
                                                showMessage={showMessage}
                                            />

                                            <MpvCoreSettingsMenu
                                                openMenu={openMenu}
                                                openSection={openSection}
                                                openSubSection={openSubSection}
                                                setOpenMenu={setOpenMenu}
                                                setOpenSection={setOpenSection}
                                                setOpenSubSection={setOpenSubSection}
                                                isFullscreen={isFullscreen}
                                                containerElement={containerElement}
                                                speed={speed}
                                                changeSpeed={changeSpeed}
                                                autoPlay={autoPlay}
                                                setAutoPlay={setAutoPlay}
                                                autoNext={autoNext}
                                                setAutoNext={setAutoNext}
                                                autoSkip={autoSkip}
                                                setAutoSkip={setAutoSkip}
                                                subtitleDelay={mpvSettings.subtitleDelay}
                                                setSubtitleDelay={value => {
                                                    setMpvSettings(current => ({ ...current, subtitleDelay: value }))
                                                    player?.setProperty("sub-delay", value)
                                                }}
                                                showChapterMarkers={showChapterMarkers}
                                                setChapterMarkers={setChapterMarkers}
                                                highlightOPEDChapters={highlightOPEDChapters}
                                                setHighlightOPEDChapters={setHighlightOPEDChapters}
                                                showStats={showStats}
                                                setShowStats={setShowStats}
                                                mpvSettings={mpvSettings}
                                                setMpvSettings={setMpvSettings}
                                                shaderSettings={shaderSettings}
                                                setShaderSettings={setShaderSettings}
                                                anime4kDirectory={anime4kDirectory}
                                                anime4kError={anime4kError}
                                                onRefreshAnime4K={() => refreshAnime4KDirectory(shaderSettings.directory || undefined)}
                                                onOpenPreferences={() => setPreferencesOpen(true)}
                                            />

                                            {!!state.playbackInfo?.videoSources?.length && (
                                                <MediaCoreMenu
                                                    name="video"
                                                    openMenu={openMenu}
                                                    onOpenMenuChange={setOpenMenu}
                                                    onOpenSectionChange={setOpenSection}
                                                    onOpenSubSectionChange={setOpenSubSection}
                                                    isFullscreen={isFullscreen}
                                                    containerElement={containerElement}
                                                    trigger={
                                                        <MediaCoreControlButtonIcon
                                                            icons={[["default", LuFilm]]}
                                                            state="default"
                                                            className="text-xl lg:text-2xl"
                                                            onClick={() => { }}
                                                            isMobile={false}
                                                            isMiniPlayer={false}
                                                        />
                                                    }
                                                >
                                                    <MediaCoreMenuTitle>Quality</MediaCoreMenuTitle>
                                                    <MediaCoreMenuBody>
                                                        <MediaCoreSettingSelect
                                                            options={state.playbackInfo.videoSources.toReversed().map(source => ({
                                                                label: source.resolution,
                                                                value: source.index,
                                                                moreInfo: source.label,
                                                            }))}
                                                            value={state.playbackInfo.selectedVideoSource}
                                                            onValueChange={value => {
                                                                const source = state.playbackInfo?.videoSources?.find(item => item.index === Number(
                                                                    value))
                                                                if (source?.url) {
                                                                    suppressEndRef.current = true
                                                                    player?.load(mc_resolveSource(source.url))
                                                                }
                                                            }}
                                                            isFullscreen={isFullscreen}
                                                            containerElement={containerElement}
                                                        />
                                                    </MediaCoreMenuBody>
                                                </MediaCoreMenu>
                                            )}

                                            {!!subtitleTracks.length && (
                                                <MediaCoreMenu
                                                    name="subtitle"
                                                    openMenu={openMenu}
                                                    onOpenMenuChange={setOpenMenu}
                                                    onOpenSectionChange={setOpenSection}
                                                    onOpenSubSectionChange={setOpenSubSection}
                                                    isFullscreen={isFullscreen}
                                                    containerElement={containerElement}
                                                    trigger={
                                                        <MediaCoreControlButtonIcon
                                                            icons={[["default", LuCaptions]]}
                                                            state="default"
                                                            onClick={() => { }}
                                                            isMobile={false}
                                                            isMiniPlayer={false}
                                                        />
                                                    }
                                                >
                                                    <MediaCoreMenuTitle>Subtitles
                                                        <IconButton
                                                            intent="gray-link" size="xs"
                                                            onClick={() => {
                                                                setOpenMenu("settings")
                                                                React.startTransition(() => {
                                                                    setOpenSection("Subtitle Styles")
                                                                })
                                                            }}
                                                            icon={<LuPaintbrush />}
                                                            className="absolute right-2 top-[calc(50%-1rem)]"
                                                        />
                                                    </MediaCoreMenuTitle>
                                                    <MediaCoreMenuBody>
                                                        <MediaCoreSettingSelect
                                                            options={[
                                                                { label: "Off", value: "no" },
                                                                ...subtitleTracks.map(mc_formatSubtitleTrack),
                                                            ]}
                                                            value={selectedSubtitle}
                                                            onValueChange={value => player?.selectTrack("subtitle", value)}
                                                            isFullscreen={isFullscreen}
                                                            containerElement={containerElement}
                                                        />
                                                    </MediaCoreMenuBody>
                                                </MediaCoreMenu>
                                            )}

                                            {audioTracks.length > 1 && (
                                                <MediaCoreMenu
                                                    name="audio"
                                                    openMenu={openMenu}
                                                    onOpenMenuChange={setOpenMenu}
                                                    onOpenSectionChange={setOpenSection}
                                                    onOpenSubSectionChange={setOpenSubSection}
                                                    isFullscreen={isFullscreen}
                                                    containerElement={containerElement}
                                                    trigger={
                                                        <MediaCoreControlButtonIcon
                                                            icons={[["default", LuHeadphones]]}
                                                            state="default"
                                                            className="text-2xl"
                                                            onClick={() => { }}
                                                            isMobile={false}
                                                            isMiniPlayer={false}
                                                        />
                                                    }
                                                >
                                                    <MediaCoreMenuTitle>Audio</MediaCoreMenuTitle>
                                                    <MediaCoreMenuBody>
                                                        <MediaCoreSettingSelect
                                                            options={audioTracks.map(mc_formatAudioTrack)}
                                                            value={selectedAudio}
                                                            onValueChange={value => player?.selectTrack("audio", value)}
                                                            isFullscreen={isFullscreen}
                                                            containerElement={containerElement}
                                                        />
                                                    </MediaCoreMenuBody>
                                                </MediaCoreMenu>
                                            )}

                                            <MpvCoreCastButton
                                                info={state.playbackInfo}
                                                paused={paused}
                                                onCastingStart={() => {
                                                    castWasPausedRef.current = pausedRef.current
                                                    player?.pause()
                                                }}
                                                onCastingEnd={() => {
                                                    if (!castWasPausedRef.current) player?.play()
                                                }}
                                            />

                                        </>
                                    )}

                                    <MediaCorePipButton
                                        isPip={isPip}
                                        onTogglePip={() => togglePip()}
                                        isMobile={false}
                                        isMiniPlayer={state.miniPlayer}
                                    />
                                    <MediaCoreFullscreenButton
                                        isFullscreen={isFullscreen}
                                        onToggleFullscreen={() => {
                                            setMiniPlayer(false)
                                            toggleFullscreen()
                                        }}
                                        isMobile={false}
                                        isMiniPlayer={state.miniPlayer}
                                    />
                                </MediaCoreControlBarView>
                            </>
                        ) : (
                            <MediaCoreLoadingOverlay
                                loadingState={state.loadingState}
                                isMiniPlayer={state.miniPlayer}
                                inline={false}
                                fullscreen={isFullscreen}
                                terminateButton={
                                    <MpvCoreFloatingButtons
                                        part="loading"
                                        fullscreen={isFullscreen}
                                        miniPlayer={state.miniPlayer}
                                        onEnterMiniPlayer={() => setMiniPlayer(true)}
                                        onExpand={() => setMiniPlayer(false)}
                                        onTerminate={() => terminate("user terminated player")}
                                        onExitFullscreen={() => toggleFullscreen(false)}
                                    />
                                }
                            />
                        )}
                    </MpvPrismVideo>
                    {!state.miniPlayer && <VideoCoreInSight />}
                </div>
            </MediaCoreDrawer>

            <Modal
                title="Terminate stream?"
                description="Press Esc again or choose terminate to stop playback."
                titleClass="text-center"
                open={isTerminateConfirmOpen && state.miniPlayer}
                onOpenChange={open => {
                    if (!open) setTerminateConfirmOpen(false)
                }}
                onEscapeKeyDown={event => {
                    event.preventDefault();
                    terminate("user terminated player")
                }}
            >
                <div className="flex gap-2 justify-center items-center">
                    <Button intent="warning-subtle" onClick={() => terminate("user terminated player")}>
                        Terminate stream
                    </Button>
                    <Button intent="white" onClick={() => setTerminateConfirmOpen(false)}>
                        Keep playing
                    </Button>
                </div>
            </Modal>
        </>
    )
}
