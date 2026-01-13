import { useDirectstreamConvertSubs } from "@/api/hooks/directstream.hooks"
import { vc_isFullscreen, vc_mediaCaptionsManager, vc_subtitleManager } from "@/app/(main)/_features/video-core/video-core"
import { logger } from "@/lib/helpers/debug"
import { isApple } from "@/lib/utils/browser-detection"
import { useAtomValue } from "jotai"
import { useEffect, useRef, useState } from "react"

const log = logger("VIDEO CORE iOS FULLSCREEN SUBTITLES")

type UseIOSFullscreenSubtitlesProps = {
    videoElement: HTMLVideoElement | null
}

/**
 * iOS doesn't support custom overlays in fullscreen mode, so we need to use
 * native video text tracks. Hook detects when entering fullscreen,
 * converts the currently selected subtitle file to VTT format, adds it
 * as a native track to the video element.
 */
export function useVideoCoreIOSFullscreenSubtitles({
    videoElement,
}: UseIOSFullscreenSubtitlesProps) {
    const nativeTrackRef = useRef<HTMLTrackElement | null>(null)
    const isIOSDevice = isApple()

    const mediaCaptionsManager = useAtomValue(vc_mediaCaptionsManager)
    const subtitleManager = useAtomValue(vc_subtitleManager)
    const isFullscreen = useAtomValue(vc_isFullscreen)

    const [flag, setFlag] = useState<boolean>(false)

    // TODO still broken
    useEffect(() => {
        function reload() {
            setFlag(p => !p)
        }

        subtitleManager?.addEventListener("trackadded", reload)
        mediaCaptionsManager?.addEventListener("trackadded", reload)
        return () => {
            subtitleManager?.removeEventListener("trackadded", reload)
            mediaCaptionsManager?.removeEventListener("trackadded", reload)
        }
    }, [subtitleManager, mediaCaptionsManager])

    const { mutateAsync: convertSubs } = useDirectstreamConvertSubs()

    useEffect(() => {
        // Only run on iOS devices
        if (!isIOSDevice || !videoElement) return

        const handleFullscreenChange = async () => {
            const isNowFullscreen = (videoElement as any).webkitDisplayingFullscreen

            if (isNowFullscreen) {
                log.info("Entered iOS fullscreen, setting up native subtitles")
                await setupNativeSubtitles()
            } else {
                log.info("Exited iOS fullscreen, cleaning up native subtitles")
                cleanupNativeSubtitles()
            }
        }

        const setupNativeSubtitles = async () => {
            try {
                let subtitleSrc: string | null = null
                let subtitleContent: string | null = null
                let subtitleLabel = "Subtitles"
                let subtitleLanguage = "en"
                let subtitleType = "vtt"

                // Get the currently selected subtitle track from either manager
                if (mediaCaptionsManager) {
                    const selectedTrack = mediaCaptionsManager.getSelectedTrack()
                    if (selectedTrack) {
                        subtitleSrc = selectedTrack.src ?? null
                        subtitleContent = selectedTrack.content ?? null
                        subtitleLabel = selectedTrack.label
                        subtitleLanguage = selectedTrack.language
                        subtitleType = selectedTrack.type ?? "vtt"
                        log.info("Using MediaCaptionsManager track", selectedTrack)
                    }
                } else if (subtitleManager) {
                    const selectedTrackNumber = subtitleManager.getSelectedTrackNumberOrNull()
                    if (selectedTrackNumber !== null) {
                        const selectedTrack = subtitleManager.getTrack(selectedTrackNumber)

                        // For file tracks, we need to get the URL
                        const fileTrack = subtitleManager.getFileTrack(selectedTrackNumber)
                        if (fileTrack?.info?.src) {
                            subtitleSrc = fileTrack.info.src
                            subtitleContent = fileTrack.info.content ?? null
                            subtitleLabel = fileTrack.info.label || selectedTrack?.label || "Subtitles"
                            subtitleLanguage = fileTrack.info.language || selectedTrack?.language || "en"
                            subtitleType = fileTrack.info.type || "vtt"
                            log.info("Using SubtitleManager file track", fileTrack)
                        } else {
                            log.warning("Selected track is event-based, cannot use for iOS native subtitles")
                            return
                        }
                    }
                }

                if (!subtitleSrc && !subtitleContent) {
                    log.info("No subtitle track selected")
                    return
                }

                // Parse the subtitle file
                log.info("Parsing subtitle file:", subtitleSrc)
                const convertedContent = subtitleType === "vtt" && !!subtitleContent
                    ? subtitleContent
                    : await convertSubs({ url: subtitleSrc || "", content: subtitleContent || "", to: "vtt" })
                if (!convertedContent) {
                    log.error("Failed to convert subtitle file")
                    return
                }
                const blob = new Blob([convertedContent], { type: "text/vtt" })
                const blobUrl = URL.createObjectURL(blob)

                // Remove any existing native tracks
                cleanupNativeSubtitles()

                // Create and add a native track element
                const track = document.createElement("track")
                track.kind = "subtitles"
                track.label = subtitleLabel
                track.srclang = subtitleLanguage
                track.src = blobUrl
                track.default = true

                videoElement.appendChild(track)
                nativeTrackRef.current = track

                // Enable the track
                track.track.mode = "showing"

                log.info("Native subtitle track added successfully")
            }
            catch (error) {
                log.error("Failed to setup native subtitles:", error)
            }
        }

        const cleanupNativeSubtitles = () => {
            if (nativeTrackRef.current) {
                // Revoke the blob URL to free memory
                if (nativeTrackRef.current.src.startsWith("blob:")) {
                    URL.revokeObjectURL(nativeTrackRef.current.src)
                }

                // Remove the track element
                nativeTrackRef.current.remove()
                nativeTrackRef.current = null
            }
        }

        videoElement.addEventListener("webkitbeginfullscreen", handleFullscreenChange)
        videoElement.addEventListener("webkitendfullscreen", handleFullscreenChange)

        return () => {
            videoElement.removeEventListener("webkitbeginfullscreen", handleFullscreenChange)
            videoElement.removeEventListener("webkitendfullscreen", handleFullscreenChange)
            cleanupNativeSubtitles()
        }
    }, [videoElement, subtitleManager, mediaCaptionsManager, isIOSDevice, isFullscreen, flag])
}
