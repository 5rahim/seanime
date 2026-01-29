import { VideoCore_VideoPlaybackInfo } from "@/app/(main)/_features/video-core/video-core.atoms"
import { logger } from "@/lib/helpers/debug"
import { atom } from "jotai"

const log = logger("VIDEO CORE MEDIA SESSION")

export const vc_mediaSessionManager = atom<VideoCoreMediaSessionManager | null>(null)

const actions = ["play", "pause", "seekforward", "seekbackward", "seekto"] as const

export class VideoCoreMediaSessionManager {
    private video: HTMLVideoElement | null = null
    private playbackInfo: VideoCore_VideoPlaybackInfo | null = null
    private isActive = false
    private controller = new AbortController()
    private eventTarget = new EventTarget()

    constructor() {
        if (!this.checkSupport()) return
    }

    setVideo(video: HTMLVideoElement | null) {
        this.video = video
        if (this.video && this.isActive) {
            this.setupVideoListeners()
            this.updatePlaybackState()
        }
    }

    setPlaybackInfo(playbackInfo: VideoCore_VideoPlaybackInfo | null) {
        this.playbackInfo = playbackInfo
        if (this.isActive) {
            this.updateMetadata()
        }
    }

    activate() {
        if (!this.checkSupport() || this.isActive) return

        this.isActive = true
        this.setupActionHandlers()
        this.updateMetadata()

        if (this.video) {
            this.setupVideoListeners()
            this.updatePlaybackState()
        }
    }

    deactivate() {
        if (!this.isActive) return

        this.isActive = false

        if ("mediaSession" in navigator) {
            navigator.mediaSession.playbackState = "none"
        }

        this.onDisconnect()
        this.controller.abort()
        this.controller = new AbortController()
    }

    on(type: string, handler: EventListener) {
        this.eventTarget.addEventListener(type, handler)
    }

    off(type: string, handler: EventListener) {
        this.eventTarget.removeEventListener(type, handler)
    }

    destroy() {
        this.deactivate()
        this.video = null
        this.playbackInfo = null
    }

    private checkSupport(): boolean {
        return "mediaSession" in navigator
    }

    private dispatch(type: string, detail?: any) {
        this.eventTarget.dispatchEvent(new CustomEvent(type, { detail }))
    }

    private onDisconnect() {
        if (!("mediaSession" in navigator)) return

        // console.warn("Media session disconnected")

        for (const action of actions) {
            navigator.mediaSession.setActionHandler(action, null)
        }

        navigator.mediaSession.metadata = new MediaMetadata({
            title: "",
            artist: "",
            artwork: [],
        })

        if ("setPositionState" in navigator.mediaSession) {
            navigator.mediaSession.setPositionState()
        }
    }

    private updateMetadata() {
        if (!this.isActive || !("mediaSession" in navigator)) return

        const metadata = this.createMetadata()
        navigator.mediaSession.metadata = metadata
    }

    private updatePlaybackState() {
        if (!this.isActive || !("mediaSession" in navigator) || !this.video) return

        const state = this.video.paused ? "paused" : "playing"
        navigator.mediaSession.playbackState = state
    }

    private createMetadata(): MediaMetadata | null {
        if (!this.playbackInfo) return null

        const episode = this.playbackInfo.episode
        const anime = episode?.baseAnime

        const title = episode?.displayTitle || "Seanime"
        const artist = anime?.title?.userPreferred || anime?.title?.romaji || anime?.title?.english || "Video Player"

        const artwork: MediaImage[] = []
        const imageUrl = episode?.episodeMetadata?.image || anime?.coverImage?.large || anime?.coverImage?.medium

        if (imageUrl) {
            artwork.push({ src: imageUrl, sizes: "512x512", type: "image/webp" })
        }

        return new MediaMetadata({ title, artist, artwork })
    }

    private setupActionHandlers() {
        if (!("mediaSession" in navigator)) return

        const handleAction = this.handleAction.bind(this)
        for (const action of actions) {
            navigator.mediaSession.setActionHandler(action, handleAction)
        }
    }

    private handleAction(details: MediaSessionActionDetails) {
        switch (details.action) {
            case "play":
                this.dispatch("media-play-request")
                break
            case "pause":
                this.dispatch("media-pause-request")
                break
            case "seekto":
            case "seekforward":
            case "seekbackward":
                let seekTime: number
                if (details.seekTime !== undefined) {
                    seekTime = details.seekTime
                } else {
                    const currentTime = this.video?.currentTime || 0
                    const offset = details.seekOffset || (details.action === "seekforward" ? 10 : -10)
                    seekTime = currentTime + offset
                }
                this.dispatch("media-seek-request", { seekTime })
                break
        }
    }

    private setupVideoListeners() {
        if (!this.video) return

        const options = { signal: this.controller.signal }

        this.video.addEventListener("play", this.updatePlaybackState.bind(this), options)
        this.video.addEventListener("pause", this.updatePlaybackState.bind(this), options)
        this.video.addEventListener("timeupdate", this.handleTimeUpdate.bind(this), options)
        this.video.addEventListener("loadedmetadata", this.updatePlaybackState.bind(this), options)
    }

    private handleTimeUpdate() {
        if (!this.isActive || !("mediaSession" in navigator) || !this.video) return

        if ("setPositionState" in navigator.mediaSession) {
            navigator.mediaSession.setPositionState({
                duration: this.video.duration || 0,
                playbackRate: this.video.playbackRate || 1,
                position: this.video.currentTime || 0,
            })
        }
    }
}
