import { NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { atomWithImmer } from "jotai-immer"

export type NativePlayerState = {
    active: boolean
    playbackInfo: NativePlayer_PlaybackInfo | null
    playbackError: string | null
    loadingState: string | null
}

export const nativePlayer_initialState: NativePlayerState = {
    active: false,
    playbackInfo: null,
    playbackError: null,
    loadingState: null,
}

export const nativePlayer_stateAtom = atomWithImmer<NativePlayerState>(nativePlayer_initialState)
