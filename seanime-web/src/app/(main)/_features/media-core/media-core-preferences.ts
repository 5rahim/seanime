import { atomWithStorage } from "jotai/utils"

export interface MediaCorePreferences {
    version: number
    autoPlay: boolean
    autoNext: boolean
    volume: number
    muted: boolean
    playbackRate: number
    autoSkip: boolean
    skipPatterns: string
    showStats: boolean
    chapterMarkers: boolean
    timestampMode: "elapsed" | "remaining"
}

export const mediaCoreDefaultPreferences: MediaCorePreferences = {
    version: 2,
    autoPlay: true,
    autoNext: true,
    volume: 1.0,
    muted: false,
    playbackRate: 1.0,
    autoSkip: false,
    skipPatterns: "",
    showStats: false,
    chapterMarkers: true,
    timestampMode: "elapsed",
}

const PREFERENCES_KEY = "sea-media-core-preferences"

function parsePreferences(value: unknown): MediaCorePreferences | null {
    if (!value || typeof value !== "object") return null
    const parsed = value as Partial<MediaCorePreferences>
    if (parsed.version !== 1 && parsed.version !== 2) return null

    return {
        ...mediaCoreDefaultPreferences,
        ...parsed,
        version: 2,
        skipPatterns: typeof parsed.skipPatterns === "string" ? parsed.skipPatterns : "",
    }
}

const customStorage = {
    getItem(key: string, initialValue: MediaCorePreferences): MediaCorePreferences {
        try {
            const raw = localStorage.getItem(key)
            if (raw) {
                const parsed = JSON.parse(raw) as any
                const preferences = parsePreferences(parsed)
                if (preferences) {
                    if (parsed.version !== 2 || typeof parsed.skipPatterns !== "string") {
                        localStorage.setItem(key, JSON.stringify(preferences))
                    }
                    return preferences
                }
            }
        } catch (e) {
            console.error("Failed to parse sea-media-core-preferences", e)
        }

        // Migration precedence: existing unified value -> VideoCore legacy value -> MpvCore legacy value -> default.
        const getLegacyValue = <T>(vcKey: string, mcKey: string, fallback: T): T => {
            try {
                const vcRaw = localStorage.getItem(vcKey)
                if (vcRaw !== null) return JSON.parse(vcRaw) as T
                const mcRaw = localStorage.getItem(mcKey)
                if (mcRaw !== null) return JSON.parse(mcRaw) as T
            } catch (e) {
                console.error(`Failed to migrate legacy keys ${vcKey} / ${mcKey}`, e)
            }
            return fallback
        }

        const migrated: MediaCorePreferences = {
            version: 2,
            autoPlay: getLegacyValue("sea-video-core-auto-play", "sea-mpv-core-auto-play", mediaCoreDefaultPreferences.autoPlay),
            autoNext: getLegacyValue("sea-video-core-auto-next", "sea-mpv-core-auto-next", mediaCoreDefaultPreferences.autoNext),
            volume: getLegacyValue("sea-video-core-volume", "sea-mpv-core-volume", mediaCoreDefaultPreferences.volume),
            muted: getLegacyValue("sea-video-core-muted", "sea-mpv-core-muted", mediaCoreDefaultPreferences.muted),
            playbackRate: getLegacyValue("sea-video-core-playback-rate", "sea-mpv-core-playback-rate", mediaCoreDefaultPreferences.playbackRate),
            autoSkip: getLegacyValue("sea-video-core-auto-skip-op-ed", "sea-mpv-core-auto-skip", mediaCoreDefaultPreferences.autoSkip),
            skipPatterns: mediaCoreDefaultPreferences.skipPatterns,
            showStats: getLegacyValue("sea-video-core-show-stats-for-nerds", "sea-mpv-core-show-stats", mediaCoreDefaultPreferences.showStats),
            chapterMarkers: getLegacyValue("sea-video-core-chapter-markers", "sea-mpv-core-chapter-markers", mediaCoreDefaultPreferences.chapterMarkers),
            timestampMode: getLegacyValue<"elapsed" | "remaining">("sea-video-core-timestamp-type", "dummy-nonexistent-key", mediaCoreDefaultPreferences.timestampMode),
        }

        try {
            localStorage.setItem(key, JSON.stringify(migrated))
        } catch (e) {
            console.error("Failed to write migrated sea-media-core-preferences", e)
        }

        return migrated
    },
    setItem(key: string, value: MediaCorePreferences) {
        localStorage.setItem(key, JSON.stringify(value))
    },
    removeItem(key: string) {
        localStorage.removeItem(key)
    },
    subscribe(key: string, callback: (value: MediaCorePreferences) => void, initialValue: MediaCorePreferences) {
        const handleStorage = (e: StorageEvent) => {
            if (e.key === key) {
                if (e.newValue === null) {
                    callback(initialValue)
                } else {
                    try {
                        callback(parsePreferences(JSON.parse(e.newValue)) ?? initialValue)
                    } catch {
                        callback(initialValue)
                    }
                }
            }
        }
        window.addEventListener("storage", handleStorage)
        return () => window.removeEventListener("storage", handleStorage)
    }
}

export const mediaCorePreferencesAtom = atomWithStorage<MediaCorePreferences>(
    PREFERENCES_KEY,
    mediaCoreDefaultPreferences,
    customStorage,
    { getOnInit: true }
)
