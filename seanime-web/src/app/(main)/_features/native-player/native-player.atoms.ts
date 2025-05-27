import { NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { atomWithImmer } from "jotai-immer"
import { atomWithStorage } from "jotai/utils"

export type NativePlayerState = {
    active: boolean
    miniPlayer: boolean
    playbackInfo: NativePlayer_PlaybackInfo | null
    playbackError: string | null
    loadingState: string | null
}

export const nativePlayer_initialState: NativePlayerState = {
    active: false,
    miniPlayer: false,
    playbackInfo: null,
    playbackError: null,
    loadingState: null,
}

export const nativePlayer_stateAtom = atomWithImmer<NativePlayerState>(nativePlayer_initialState)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////

export type NativePlayerSettings = {
    preferredSubtitleLanguage: string
    preferredAudioLanguage: string
    // Video enhancement settings
    videoEnhancement: {
        enabled: boolean
        contrast: number      // 0.8 - 1.2 (1.0 = default)
        saturation: number    // 0.8 - 1.3 (1.0 = default)
        brightness: number    // 0.9 - 1.1 (1.0 = default)
    }
}

export const nativePlayer_initialSettings: NativePlayerSettings = {
    preferredSubtitleLanguage: "eng",
    preferredAudioLanguage: "jpn",
    videoEnhancement: {
        enabled: true,
        contrast: 1.05,
        saturation: 1.1,
        brightness: 1.02,
    },
}

export const nativePlayer_settingsAtom = atomWithStorage<NativePlayerSettings>("sea-native-player-settings",
    nativePlayer_initialSettings,
    undefined,
    { getOnInit: true })
