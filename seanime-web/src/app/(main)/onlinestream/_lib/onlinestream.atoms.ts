import { atom } from "jotai/index"
import { atomWithStorage } from "jotai/utils"

export const __onlinestream_selectedProviderAtom = atomWithStorage<string | null>("sea-onlinestream-provider", null)

export const __onlinestream_selectedDubbedAtom = atom<boolean>(false)

// Variable used for the episode source query
export const __onlinestream_selectedEpisodeNumberAtom = atom<number | undefined>(undefined)

export const __onlinestream_autoPlayAtom = atomWithStorage("sea-onlinestream-autoplay", false)

export const __onlinestream_autoNextAtom = atomWithStorage("sea-onlinestream-autonext", false)

export const __onlinestream_autoSkipIntroOutroAtom = atomWithStorage("sea-onlinestream-autoskip-intro-outro", false)

export const __onlinestream_selectedServerAtom = atomWithStorage<string | undefined>("sea-onlinestream-server", undefined)

export const __onlinestream_qualityAtom = atomWithStorage<string | undefined>("sea-onlinestream-quality", undefined)

export const __onlinestream_volumeAtom = atomWithStorage<number>("sea-onlinestream-volume", 1)
