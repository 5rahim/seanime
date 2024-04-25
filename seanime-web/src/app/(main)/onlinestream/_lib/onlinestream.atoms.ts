import { OnlinestreamProvider } from "@/app/(main)/onlinestream/_lib/onlinestream.enums"
import { atom } from "jotai/index"
import { atomWithStorage } from "jotai/utils"

export const __onlinestream_selectedProviderAtom = atomWithStorage<string>("sea-onlinestream-provider", OnlinestreamProvider.GOGOANIME)

export const __onlinestream_selectedDubbedAtom = atom<boolean>(false)

export const __onlinestream_selectedEpisodeNumberAtom = atom<number | undefined>(undefined)

export const __onlinestream_autoPlayAtom = atomWithStorage("sea-onlinestream-autoplay", false)

export const __onlinestream_autoNextAtom = atomWithStorage("sea-onlinestream-autonext", false)

export const __onlinestream_selectedServerAtom = atomWithStorage<string | undefined>("sea-onlinestream-server", undefined)

export const __onlinestream_qualityAtom = atomWithStorage<string | undefined>("sea-onlinestream-quality", undefined)
