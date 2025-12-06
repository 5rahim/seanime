import { vc_isFullscreen, vc_mediaCaptionsManager, vc_subtitleManager } from "@/app/(main)/_features/video-core/video-core"
import { logger } from "@/lib/helpers/debug"
import { isApple } from "@/lib/utils/browser-detection"
import { useAtomValue } from "jotai"
import { parseResponse } from "media-captions"
import { useEffect, useRef } from "react"

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
                let subtitleLabel = "Subtitles"
                let subtitleLanguage = "en"

                // Get the currently selected subtitle track from either manager
                if (mediaCaptionsManager) {
                    const selectedTrack = mediaCaptionsManager.getSelectedTrack()
                    if (selectedTrack) {
                        subtitleSrc = selectedTrack.src
                        subtitleLabel = selectedTrack.label
                        subtitleLanguage = selectedTrack.language
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
                            subtitleLabel = fileTrack.info.label || selectedTrack?.label || "Subtitles"
                            subtitleLanguage = fileTrack.info.language || selectedTrack?.language || "en"
                            log.info("Using SubtitleManager file track", fileTrack)
                        } else {
                            log.warning("Selected track is MKV-based, cannot use for iOS native subtitles")
                            return
                        }
                    }
                }

                if (!subtitleSrc) {
                    log.info("No subtitle track selected")
                    return
                }

                // Parse the subtitle file
                log.info("Parsing subtitle file:", subtitleSrc)
                const result = await parseResponse(fetch(subtitleSrc))

                // Convert to VTT format
                const vttContent = convertToVTT(result.cues)
                const blob = new Blob([vttContent], { type: "text/vtt" })
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
    }, [videoElement, subtitleManager, mediaCaptionsManager, isIOSDevice, isFullscreen])
}

function convertToVTT(cues: any[]): string {
    let vtt = "WEBVTT\n\n"

    for (let i = 0; i < cues.length; i++) {
        const cue = cues[i]

        const startTime = formatVTTTimestamp(cue.startTime)
        const endTime = formatVTTTimestamp(cue.endTime)

        let text = cue.text || ""

        // Add the cue
        vtt += `${i + 1}\n`
        vtt += `${startTime} --> ${endTime}\n`
        vtt += `${text}\n\n`
    }

    return vtt
}

// Format seconds to VTT timestamp format (HH:MM:SS.mmm)
function formatVTTTimestamp(timeInSeconds: number): string {
    const hours = Math.floor(timeInSeconds / 3600)
    const minutes = Math.floor((timeInSeconds % 3600) / 60)
    const seconds = Math.floor(timeInSeconds % 60)
    const milliseconds = Math.floor((timeInSeconds % 1) * 1000)

    return `${pad(hours, 2)}:${pad(minutes, 2)}:${pad(seconds, 2)}.${pad(milliseconds, 3)}`
}

function pad(num: number, size: number): string {
    let str = num.toString()
    while (str.length < size) {
        str = "0" + str
    }
    return str
}

