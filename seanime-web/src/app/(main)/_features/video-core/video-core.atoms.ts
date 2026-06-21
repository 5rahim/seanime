import { VideoCore_PlaybackType, VideoCore_VideoPlaybackInfo, VideoCore_VideoSource, VideoCore_VideoSubtitleTrack } from "@/api/generated/types"
import { atom } from "jotai"
import { atomWithStorage } from "jotai/utils"
import { mediaCorePreferencesAtom } from "@/app/(main)/_features/media-core/media-core-preferences"

export type VideoCoreLifecycleState = {
    active: boolean
    playbackInfo: VideoCore_VideoPlaybackInfo | null
    playbackError: string | null
    loadingState: string | null
}

export type {
    VideoCore_VideoSubtitleTrack, VideoCore_PlaybackType, VideoCore_VideoSource, VideoCore_VideoPlaybackInfo,
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type VideoCoreSettings = {
    preferredSubtitleLanguage: string
    preferredSubtitleBlacklist: string
    preferredAudioLanguage: string
    subtitleDelay: number // in seconds
    // Video enhancement settings
    videoEnhancement: {
        enabled: boolean
        contrast: number      // 0.8 - 1.2 (1.0 = default)
        saturation: number    // 0.8 - 1.3 (1.0 = default)
        brightness: number    // 0.9 - 1.1 (1.0 = default)
    }
    // Subtitle customization settings (ASS)
    subtitleCustomization: {
        enabled: boolean
        fontSize?: number
        fontName?: string
        primaryColor?: string
        outlineColor?: string
        backColor?: string
        backColorOpacity?: number
        outline?: number
        shadow?: number
    }
    // Caption customization settings (non-ASS)
    captionCustomization: {
        fontSize?: number
        textColor?: string
        backgroundColor?: string
        backgroundOpacity?: number
        textShadow?: number
        textShadowColor?: string
    }
}

export const vc_initialSettings: VideoCoreSettings = {
    preferredSubtitleLanguage: "en,eng,english",
    preferredSubtitleBlacklist: "",
    preferredAudioLanguage: "jpn,jp,jap,japanese",
    subtitleDelay: 0,
    videoEnhancement: {
        enabled: true,
        contrast: 1.05,
        saturation: 1.1,
        brightness: 1.02,
    },
    subtitleCustomization: {
        enabled: false,
    },
    captionCustomization: {},
}

// Wrapped atom for backward compatibility
export const vc_settingsRaw = atomWithStorage<Partial<VideoCoreSettings>>("sea-video-core-settings",
    vc_initialSettings,
    undefined,
    { getOnInit: true })

export const vc_settings = atom(
    (get) => {
        const settings = get(vc_settingsRaw)
        return {
            ...vc_initialSettings,
            ...settings,
            subtitleCustomization: {
                ...vc_initialSettings.subtitleCustomization,
                ...(settings.subtitleCustomization || {}),
            },
            captionCustomization: {
                ...vc_initialSettings.captionCustomization,
                ...(settings.captionCustomization || {}),
            },
            videoEnhancement: {
                ...vc_initialSettings.videoEnhancement,
                ...(settings.videoEnhancement || {}),
            },
        } as VideoCoreSettings
    },
    (get, set, update: VideoCoreSettings) => {
        set(vc_settingsRaw, update)
    },
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export interface VideoCoreKeybindings {
    seekForward: { key: string; value: number }
    seekBackward: { key: string; value: number }
    seekForwardFine: { key: string; value: number }
    seekBackwardFine: { key: string; value: number }
    nextChapter: { key: string }
    previousChapter: { key: string }
    volumeUp: { key: string; value: number }
    volumeDown: { key: string; value: number }
    mute: { key: string }
    cycleSubtitles: { key: string }
    cycleAudio: { key: string }
    nextEpisode: { key: string }
    previousEpisode: { key: string }
    fullscreen: { key: string }
    pictureInPicture: { key: string }
    increaseSpeed: { key: string; value: number }
    decreaseSpeed: { key: string; value: number }
    takeScreenshot: { key: string }
    openInSight: { key: string }
    statsForNerds: { key: string }
}

export const vc_defaultKeybindings: VideoCoreKeybindings = {
    seekForward: { key: "KeyD", value: 30 },
    seekBackward: { key: "KeyA", value: 30 },
    seekForwardFine: { key: "ArrowRight", value: 2 },
    seekBackwardFine: { key: "ArrowLeft", value: 2 },
    nextChapter: { key: "KeyE" },
    previousChapter: { key: "KeyQ" },
    volumeUp: { key: "ArrowUp", value: 5 },
    volumeDown: { key: "ArrowDown", value: 5 },
    mute: { key: "KeyM" },
    cycleSubtitles: { key: "KeyJ" },
    cycleAudio: { key: "KeyK" },
    nextEpisode: { key: "KeyN" },
    previousEpisode: { key: "KeyB" },
    fullscreen: { key: "KeyF" },
    pictureInPicture: { key: "KeyP" },
    increaseSpeed: { key: "BracketRight", value: 0.1 },
    decreaseSpeed: { key: "BracketLeft", value: 0.1 },
    takeScreenshot: { key: "KeyI" },
    openInSight: { key: "KeyH" },
    statsForNerds: { key: "KeyZ" },
}

const vc_keybindingsRaw = atomWithStorage<Partial<VideoCoreKeybindings>>("sea-video-core-keybindings",
    vc_defaultKeybindings,
    undefined,
    { getOnInit: true })

export const vc_keybindingsAtom = atom(
    (get) => {
        const stored = get(vc_keybindingsRaw)
        // Merge stored with defaults
        return {
            ...vc_defaultKeybindings,
            ...stored,
        } as VideoCoreKeybindings
    },
    (get, set, update: VideoCoreKeybindings) => {
        set(vc_keybindingsRaw, update)
    },
)

export const vc_useLibassRendererAtom = atomWithStorage("sea-video-core-use-libass-renderer", true, undefined, { getOnInit: true })

export const vc_showChapterMarkersAtom = atom(
    (get) => get(mediaCorePreferencesAtom).chapterMarkers,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.chapterMarkers) : newValue
        set(mediaCorePreferencesAtom, { ...current, chapterMarkers: next })
    }
)
export const vc_highlightOPEDChaptersAtom = atomWithStorage("sea-video-core-highlight-op-ed-chapters", true, undefined, { getOnInit: true })
export const vc_beautifyImageAtom = atomWithStorage("sea-video-core-increase-saturation", false, undefined, { getOnInit: true })
export const vc_autoNextAtom = atom(
    (get) => get(mediaCorePreferencesAtom).autoNext,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.autoNext) : newValue
        set(mediaCorePreferencesAtom, { ...current, autoNext: next })
    }
)
export const vc_autoPlayVideoAtom = atom(
    (get) => get(mediaCorePreferencesAtom).autoPlay,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.autoPlay) : newValue
        set(mediaCorePreferencesAtom, { ...current, autoPlay: next })
    }
)
export const vc_autoSkipOPEDAtom = atom(
    (get) => get(mediaCorePreferencesAtom).autoSkip,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.autoSkip) : newValue
        set(mediaCorePreferencesAtom, { ...current, autoSkip: next })
    }
)
export const vc_storedVolumeAtom = atom(
    (get) => get(mediaCorePreferencesAtom).volume,
    (get, set, newValue: number | ((prev: number) => number)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.volume) : newValue
        set(mediaCorePreferencesAtom, { ...current, volume: next })
    }
)
export const vc_storedMutedAtom = atom(
    (get) => get(mediaCorePreferencesAtom).muted,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.muted) : newValue
        set(mediaCorePreferencesAtom, { ...current, muted: next })
    }
)
export const vc_storedPlaybackRateAtom = atom(
    (get) => get(mediaCorePreferencesAtom).playbackRate,
    (get, set, newValue: number | ((prev: number) => number)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.playbackRate) : newValue
        set(mediaCorePreferencesAtom, { ...current, playbackRate: next })
    }
)
export const vc_showStatsForNerdsAtom = atom(
    (get) => get(mediaCorePreferencesAtom).showStats,
    (get, set, newValue: boolean | ((prev: boolean) => boolean)) => {
        const current = get(mediaCorePreferencesAtom)
        const next = typeof newValue === "function" ? newValue(current.showStats) : newValue
        set(mediaCorePreferencesAtom, { ...current, showStats: next })
    }
)
