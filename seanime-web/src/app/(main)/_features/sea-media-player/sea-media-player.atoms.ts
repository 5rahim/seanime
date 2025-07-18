import { atomWithStorage } from "jotai/utils"

export const __seaMediaPlayer_autoPlayAtom = atomWithStorage("sea-media-player-autoplay", false)

export const __seaMediaPlayer_autoNextAtom = atomWithStorage("sea-media-player-autonext", false)

export const __seaMediaPlayer_autoSkipIntroOutroAtom = atomWithStorage("sea-media-player-autoskip-intro-outro", false)

export const __seaMediaPlayer_discreteControlsAtom = atomWithStorage("sea-media-player-discrete-controls", false)

export const __seaMediaPlayer_volumeAtom = atomWithStorage("sea-media-player-volume", 1)

export const __seaMediaPlayer_mutedAtom = atomWithStorage("sea-media-player-muted", false)
