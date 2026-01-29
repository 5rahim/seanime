import { AL_BaseAnime } from "@/api/generated/types"
import { atom } from "jotai"
import { ScopeProvider } from "jotai-scope"
import React, { createContext, useContext } from "react"

type MediaPlayerProviderProps = {
    media: AL_BaseAnime | null
    progress: {
        currentProgress: number | null
        currentEpisodeNumber: number | null
        currentEpisodeTitle: string | null
    }
}
type MediaPlayerState = {} & MediaPlayerProviderProps

type ProgressItem = {
    episodeNumber: number
}

export const __seaMediaPlayer_scopedProgressItemAtom = atom<ProgressItem | null>(null)
export const __seaMediaPlayer_scopedCurrentProgressAtom = atom<number>(0)

const MediaPlayerContext = createContext<MediaPlayerState>({
    media: null,
    progress: { currentProgress: null, currentEpisodeNumber: null, currentEpisodeTitle: null },
})

export function SeaMediaPlayerProvider({ children, ...providerProps }: { children?: React.ReactNode } & MediaPlayerProviderProps) {

    return (
        <MediaPlayerContext.Provider
            value={{
                ...providerProps,
            }}
        >
            <ScopeProvider atoms={[__seaMediaPlayer_scopedProgressItemAtom, __seaMediaPlayer_scopedCurrentProgressAtom]}>
                {children}
            </ScopeProvider>
        </MediaPlayerContext.Provider>
    )
}


export function useSeaMediaPlayer() {
    return useContext(MediaPlayerContext)
}


