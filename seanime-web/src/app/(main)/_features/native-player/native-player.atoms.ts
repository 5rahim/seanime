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
}

export const nativePlayer_initialSettings: NativePlayerSettings = {
    preferredSubtitleLanguage: "eng",
    preferredAudioLanguage: "jpn",
}

export const nativePlayer_settingsAtom = atomWithStorage<NativePlayerSettings>("sea-native-player-settings",
    nativePlayer_initialSettings,
    undefined,
    { getOnInit: true })
