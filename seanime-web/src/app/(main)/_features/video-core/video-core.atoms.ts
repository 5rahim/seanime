import { atomWithStorage } from "jotai/utils"

export type VideoCoreSettings = {
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

export const vc_initialSettings: VideoCoreSettings = {
    preferredSubtitleLanguage: "eng",
    preferredAudioLanguage: "jpn",
    videoEnhancement: {
        enabled: true,
        contrast: 1.05,
        saturation: 1.1,
        brightness: 1.02,
    },
}

export const vc_settings = atomWithStorage<VideoCoreSettings>("sea-video-core-settings",
    vc_initialSettings,
    undefined,
    { getOnInit: true })

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
}

export const vc_keybindingsAtom = atomWithStorage("sea-video-core-keybindings", vc_defaultKeybindings, undefined, { getOnInit: true })

export const vc_showChapterMarkersAtom = atomWithStorage("sea-video-core-chapter-markers", true, undefined, { getOnInit: true })
export const vc_highlightOPEDChaptersAtom = atomWithStorage("sea-video-core-highlight-op-ed-chapters", true, undefined, { getOnInit: true })
export const vc_beautifyImageAtom = atomWithStorage("sea-video-core-beautify-image", true, undefined, { getOnInit: true })
