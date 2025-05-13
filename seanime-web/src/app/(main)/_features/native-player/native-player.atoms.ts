import { NativePlayer_PlaybackInfo } from "@/api/generated/types"
import { atomWithImmer } from "jotai-immer"

type State = {
    active: boolean
    miniPlayer: boolean
    playbackInfo: NativePlayer_PlaybackInfo | null
    playbackError: string | null
    loadingState: string | null
}

export const nativePlayer_initialState: State = {
    active: false,
    miniPlayer: false,
    playbackInfo: null,
    playbackError: null,
    loadingState: null,
}

export const nativePlayer_stateAtom = atomWithImmer<State>(nativePlayer_initialState)
