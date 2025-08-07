import { __seaMediaPlayer_mutedAtom, __seaMediaPlayer_volumeAtom } from "@/app/(main)/_features/sea-media-player/sea-media-player.atoms"
import {
    vc_audioManager,
    vc_doAction,
    vc_isFullscreen,
    vc_isMuted,
    vc_pip,
    vc_subtitleManager,
    vc_volume,
    VideoCoreChapterCue,
} from "@/app/(main)/_features/video-core/video-core"
import { vc_defaultKeybindings, vc_keybindingsAtom, VideoCoreKeybindings } from "@/app/(main)/_features/video-core/video-core.atoms"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { NumberInput } from "@/components/ui/number-input"
import { logger } from "@/lib/helpers/debug"
import { atom, useAtom, useAtomValue } from "jotai"
import { atomWithImmer } from "jotai-immer"
import { useSetAtom } from "jotai/react"
import React, { useCallback, useEffect, useRef, useState } from "react"

export const videoCoreKeybindingsModalAtom = atom(false)

// Flash notification system
type FlashNotification = {
    id: string
    message: string
    timestamp: number
}

const flashNotificationAtom = atomWithImmer<FlashNotification | null>(null)

export function FlashNotificationDisplay() {
    const [notification] = useAtom(flashNotificationAtom)

    if (!notification) return null

    return (
        <div className="absolute top-16 left-1/2 transform -translate-x-1/2 z-50 pointer-events-none">
            <div className="text-white px-4 py-2 !text-xl font-bold" style={{ textShadow: "0 1px 10px rgba(0, 0, 0, 0.8)" }}>
                {notification.message}
            </div>
        </div>
    )
}

export function useFlashNotification() {
    const [, setNotification] = useAtom(flashNotificationAtom)
    const timeoutRef = useRef<NodeJS.Timeout | null>(null)

    const showFlash = useCallback((message: string) => {
        const id = Date.now().toString()
        setNotification({ id, message, timestamp: Date.now() })
        if (timeoutRef.current) {
            clearTimeout(timeoutRef.current)
        }

        timeoutRef.current = setTimeout(() => {
            setNotification(null)
            timeoutRef.current = null
        }, 1000)
    }, [])

    return { showFlash }
}

export function VideoCoreKeybindingsModal() {
    const [open, setOpen] = useAtom(videoCoreKeybindingsModalAtom)
    const [keybindings, setKeybindings] = useAtom(vc_keybindingsAtom)
    const [editedKeybindings, setEditedKeybindings] = useState<VideoCoreKeybindings>(keybindings)
    const [recordingKey, setRecordingKey] = useState<string | null>(null)

    // Reset edited keybindings when modal opens
    useEffect(() => {
        if (open) {
            setEditedKeybindings(keybindings)
        }
    }, [open, keybindings])

    const handleKeyRecord = (actionKey: keyof VideoCoreKeybindings) => {
        setRecordingKey(actionKey)

        const handleKeyDown = (e: KeyboardEvent) => {
            e.preventDefault()
            e.stopPropagation()

            setEditedKeybindings(prev => ({
                ...prev,
                [actionKey]: {
                    ...prev[actionKey],
                    key: e.code,
                },
            }))

            setRecordingKey(null)
            document.removeEventListener("keydown", handleKeyDown)
        }

        document.addEventListener("keydown", handleKeyDown)
    }

    const handleSave = () => {
        setKeybindings(editedKeybindings)
        setOpen(false)
    }

    const handleReset = () => {
        setEditedKeybindings(vc_defaultKeybindings)
    }

    const formatKeyDisplay = (keyCode: string) => {
        const keyMap: Record<string, string> = {
            "KeyA": "A", "KeyB": "B", "KeyC": "C", "KeyD": "D", "KeyE": "E", "KeyF": "F",
            "KeyG": "G", "KeyH": "H", "KeyI": "I", "KeyJ": "J", "KeyK": "K", "KeyL": "L",
            "KeyM": "M", "KeyN": "N", "KeyO": "O", "KeyP": "P", "KeyQ": "Q", "KeyR": "R",
            "KeyS": "S", "KeyT": "T", "KeyU": "U", "KeyV": "V", "KeyW": "W", "KeyX": "X",
            "KeyY": "Y", "KeyZ": "Z",
            "ArrowUp": "↑", "ArrowDown": "↓", "ArrowLeft": "←", "ArrowRight": "→",
            "BracketLeft": "[", "BracketRight": "]",
            "Space": "⎵",
        }
        return keyMap[keyCode] || keyCode
    }

    const KeybindingRow = ({
        action,
        description,
        actionKey,
        hasValue = false,
        valueLabel = "",
    }: {
        action: string
        description: string
        actionKey: keyof VideoCoreKeybindings
        hasValue?: boolean
        valueLabel?: string
    }) => (
        <div className="flex items-center justify-between py-3 border-b border-border/50 last:border-b-0">
            <div className="flex-1">
                <div className="font-medium text-sm">{action}</div>
                {hasValue && (
                    <div className="flex items-center gap-2 mt-1">
                        <span className="text-xs text-muted-foreground">{valueLabel}:</span>
                        <NumberInput
                            value={("value" in editedKeybindings[actionKey]) ? (editedKeybindings[actionKey] as any).value : 0}
                            onChange={(value) => setEditedKeybindings(prev => ({
                                ...prev,
                                [actionKey]: { ...prev[actionKey], value: value || 0 },
                            }))}
                            size="sm"
                            fieldClass="w-16"
                            hideControls
                            min={0}
                            step={actionKey.includes("Speed") ? 0.25 : 1}
                        />
                    </div>
                )}
            </div>
            <div className="flex items-center gap-2">
                <Button
                    intent={recordingKey === actionKey ? "white-subtle" : "gray-outline"}
                    size="sm"
                    onClick={() => handleKeyRecord(actionKey)}
                    className={cn(
                        "h-8 px-3 text-lg font-mono",
                        recordingKey === actionKey && "!text-xs text-white",
                    )}
                >
                    {recordingKey === actionKey ? "Press key..." : formatKeyDisplay(editedKeybindings?.[actionKey]?.key ?? "")}
                </Button>
            </div>
        </div>
    )

    return (
        <Modal
            title="Keyboard Shortcuts"
            description="Customize the keyboard shortcuts for the player"
            open={open}
            onOpenChange={setOpen}
            contentClass="max-w-5xl focus:outline-none focus-visible:outline-none outline-none"
        >
            <div className="grid grid-cols-3 gap-8">
                {/* Playback Column */}
                <div>
                    <h3 className="text-lg font-semibold mb-4 text-white">Playback</h3>
                    <div className="space-y-0">
                        <KeybindingRow
                            action="Seek Forward"
                            description="Seek forward"
                            actionKey="seekForward"
                            hasValue={true}
                            valueLabel="Seconds"
                        />
                        <KeybindingRow
                            action="Seek Backward"
                            description="Seek backward"
                            actionKey="seekBackward"
                            hasValue={true}
                            valueLabel="Seconds"
                        />
                        <KeybindingRow
                            action="Seek Forward (Fine)"
                            description="Seek forward (fine)"
                            actionKey="seekForwardFine"
                            hasValue={true}
                            valueLabel="Seconds"
                        />
                        <KeybindingRow
                            action="Seek Backward (Fine)"
                            description="Seek backward (fine)"
                            actionKey="seekBackwardFine"
                            hasValue={true}
                            valueLabel="Seconds"
                        />
                        <KeybindingRow
                            action="Increase Speed"
                            description="Increase playback speed"
                            actionKey="increaseSpeed"
                            hasValue={true}
                            valueLabel="increment"
                        />
                        <KeybindingRow
                            action="Decrease Speed"
                            description="Decrease playback speed"
                            actionKey="decreaseSpeed"
                            hasValue={true}
                            valueLabel="increment"
                        />
                    </div>
                </div>

                {/* Navigation Column */}
                <div>
                    <h3 className="text-lg font-semibold mb-4 text-white">Navigation</h3>
                    <div className="space-y-0">
                        <KeybindingRow
                            action="Next Chapter"
                            description="Skip to next chapter"
                            actionKey="nextChapter"
                        />
                        <KeybindingRow
                            action="Previous Chapter"
                            description="Skip to previous chapter"
                            actionKey="previousChapter"
                        />
                        <KeybindingRow
                            action="Next Episode"
                            description="Play next episode"
                            actionKey="nextEpisode"
                        />
                        <KeybindingRow
                            action="Previous Episode"
                            description="Play previous episode"
                            actionKey="previousEpisode"
                        />
                        <KeybindingRow
                            action="Cycle Subtitles"
                            description="Cycle through subtitle tracks"
                            actionKey="cycleSubtitles"
                        />
                        <KeybindingRow
                            action="Fullscreen"
                            description="Toggle fullscreen"
                            actionKey="fullscreen"
                        />
                        <KeybindingRow
                            action="Picture in Picture"
                            description="Toggle picture in picture"
                            actionKey="pictureInPicture"
                        />
                    </div>
                </div>

                {/* Controls Column */}
                <div>
                    <h3 className="text-lg font-semibold mb-4 text-white">Audio</h3>
                    <div className="space-y-0">
                        <KeybindingRow
                            action="Volume Up"
                            description="Increase volume"
                            actionKey="volumeUp"
                            hasValue={true}
                            valueLabel="Percent"
                        />
                        <KeybindingRow
                            action="Volume Down"
                            description="Decrease volume"
                            actionKey="volumeDown"
                            hasValue={true}
                            valueLabel="Percent"
                        />
                        <KeybindingRow
                            action="Mute"
                            description="Toggle mute"
                            actionKey="mute"
                        />
                        <KeybindingRow
                            action="Cycle Audio"
                            description="Cycle through audio tracks"
                            actionKey="cycleAudio"
                        />
                    </div>
                </div>
            </div>

            <div className="flex items-center justify-between pt-6 mt-6 border-t border-border">
                <Button
                    intent="gray-outline"
                    onClick={handleReset}
                >
                    Reset to Defaults
                </Button>
                <div className="flex gap-2">
                    <Button
                        intent="gray-outline"
                        onClick={() => setOpen(false)}
                    >
                        Cancel
                    </Button>
                    <Button
                        intent="primary"
                        onClick={handleSave}
                    >
                        Save Changes
                    </Button>
                </div>
            </div>
        </Modal>
    )
}

export function VideoCoreKeybindingController(props: {
    active: boolean
    videoRef: React.RefObject<HTMLVideoElement>,
    chapterCues: VideoCoreChapterCue[],
    introEndTime: number | undefined,
    introStartTime: number | undefined
}) {
    const {
        active,
        videoRef,
        chapterCues,
        introEndTime,
        introStartTime,
    } = props

    const [keybindings] = useAtom(vc_keybindingsAtom)
    const fullscreen = useAtomValue(vc_isFullscreen)
    const pip = useAtomValue(vc_pip)
    const volume = useAtomValue(vc_volume)
    const setVolume = useSetAtom(__seaMediaPlayer_volumeAtom)
    const muted = useAtomValue(vc_isMuted)
    const setMuted = useSetAtom(__seaMediaPlayer_mutedAtom)
    const { showFlash } = useFlashNotification()

    const action = useSetAtom(vc_doAction)

    const subtitleManager = useAtomValue(vc_subtitleManager)
    const audioManager = useAtomValue(vc_audioManager)

    // Rate limiting for seeking operations
    const lastSeekTime = useRef(0)
    const SEEK_THROTTLE_MS = 100 // Minimum time between seek operations

    function seek(seconds: number) {
        action({ type: "seek", payload: { time: seconds } })
    }

    function seekTo(to: number) {
        action({ type: "seekTo", payload: { time: to } })
    }

    //
    // Keyboard shortcuts
    //

    const handleKeyboardShortcuts = useCallback((e: KeyboardEvent) => {
        // Don't handle shortcuts if we're in an input field or modal is open
        if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
            return
        }

        if (!videoRef.current || !active) {
            return
        }

        const video = videoRef.current

        // Handle escape key to exit fullscreen
        if (e.code === "Escape" && fullscreen) {
            e.preventDefault()
            // mediaStore.dispatch({
            //     type: "mediaexitfullscreenrequest",
            // })
            return
        }

        // Helper function to check if seeking is rate limited
        const canSeek = () => {
            const now = Date.now()
            if (now - lastSeekTime.current < SEEK_THROTTLE_MS) {
                return false
            }
            lastSeekTime.current = now
            return true
        }

        // Check which shortcut was pressed
        if (e.code === keybindings.seekForward.key) {
            e.preventDefault()
            if (!canSeek()) return

            if (props.introEndTime && props.introStartTime && video.currentTime < props.introEndTime && video.currentTime >= props.introStartTime) {
                seekTo(props.introEndTime)
                showFlash("Skipped intro")
                return
            }
            seek(keybindings.seekForward.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekBackward.key) {
            e.preventDefault()
            if (!canSeek()) return
            seek(-keybindings.seekBackward.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekForwardFine.key) {
            e.preventDefault()
            // if (!canSeek()) return
            video.dispatchEvent(new Event("seeking"))
            seek(keybindings.seekForwardFine.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.seekBackwardFine.key) {
            e.preventDefault()
            // if (!canSeek()) return
            video.dispatchEvent(new Event("seeking"))
            seek(-keybindings.seekBackwardFine.value)
            video.dispatchEvent(new Event("seeked"))
        } else if (e.code === keybindings.nextChapter.key) {
            e.preventDefault()
            handleNextChapter()
        } else if (e.code === keybindings.previousChapter.key) {
            e.preventDefault()
            handlePreviousChapter()
        } else if (e.code === keybindings.volumeUp.key) {
            e.preventDefault()
            const newVolume = Math.min(1, volume + keybindings.volumeUp.value / 100)
            setVolume(newVolume)
        } else if (e.code === keybindings.volumeDown.key) {
            e.preventDefault()
            const newVolume = Math.max(0, volume - keybindings.volumeDown.value / 100)
            setVolume(newVolume)
        } else if (e.code === keybindings.mute.key) {
            e.preventDefault()
            setMuted(!muted)
        } else if (e.code === keybindings.cycleSubtitles.key) {
            e.preventDefault()
            handleCycleSubtitles()
        } else if (e.code === keybindings.cycleAudio.key) {
            e.preventDefault()
            handleCycleAudio()
        } else if (e.code === keybindings.nextEpisode.key) {
            e.preventDefault()
            handleNextEpisode()
        } else if (e.code === keybindings.previousEpisode.key) {
            e.preventDefault()
            handlePreviousEpisode()
        } else if (e.code === keybindings.fullscreen.key) {
            e.preventDefault()
            handleToggleFullscreen()
        } else if (e.code === keybindings.pictureInPicture.key) {
            e.preventDefault()
            handleTogglePictureInPicture()
        } else if (e.code === keybindings.increaseSpeed.key) {
            e.preventDefault()
            const newRate = Math.min(8, video.playbackRate + keybindings.increaseSpeed.value)
            video.playbackRate = newRate
            showFlash(`Speed: ${newRate.toFixed(2)}x`)
        } else if (e.code === keybindings.decreaseSpeed.key) {
            e.preventDefault()
            const newRate = Math.max(0.20, video.playbackRate - keybindings.decreaseSpeed.value)
            video.playbackRate = newRate
            showFlash(`Speed: ${newRate.toFixed(2)}x`)
        }
    }, [keybindings, volume, muted, seek, active, fullscreen, pip, showFlash, introEndTime, introStartTime])

    // Keyboard shortcut handlers
    const handleNextChapter = useCallback(() => {
        if (!videoRef.current || !chapterCues) return

        const currentTime = videoRef.current.currentTime

        // Sort chapters by start time to ensure proper order
        const sortedChapters = [...chapterCues].sort((a, b) => a.startTime - b.startTime)

        // Find the next chapter (with a small buffer to avoid edge cases)
        const nextChapter = sortedChapters.find(chapter => chapter.startTime > currentTime + 1)
        if (nextChapter) {
            seekTo(nextChapter.startTime)
            // Try to get chapter name from video track cues
            const chapterName = nextChapter.text
            showFlash(chapterName ? `Chapter: ${chapterName}` : `Chapter ${sortedChapters.indexOf(nextChapter) + 1}`)
        } else {
            // If no next chapter, go to the end
            const lastChapter = sortedChapters[sortedChapters.length - 1]
            if (lastChapter && lastChapter.endTime) {
                seekTo(lastChapter.endTime)
                showFlash("End of chapters")
            }
        }
    }, [chapterCues, seekTo, showFlash])

    const handlePreviousChapter = useCallback(() => {
        if (!videoRef.current || !chapterCues) return

        const currentTime = videoRef.current.currentTime

        // Sort chapters by start time to ensure proper order
        const sortedChapters = [...chapterCues].sort((a, b) => a.startTime - b.startTime)

        // Find the current chapter first
        const currentChapterIndex = sortedChapters.findIndex((chapter, index) => {
            const nextChapter = sortedChapters[index + 1]
            return chapter.startTime <= currentTime && (!nextChapter || currentTime < nextChapter.startTime)
        })

        if (currentChapterIndex > 0) {
            // Go to previous chapter
            const previousChapter = sortedChapters[currentChapterIndex - 1]
            seekTo(previousChapter.startTime)
            const chapterName = previousChapter.text
            showFlash(chapterName ? `Chapter: ${chapterName}` : `Chapter ${currentChapterIndex}`)
        } else if (currentChapterIndex === 0) {
            // Already in first chapter, go to the beginning
            seekTo(0)
            const firstChapter = sortedChapters[0]
            const chapterName = firstChapter.text
            showFlash(chapterName ? `Chapter: ${chapterName}` : "Chapter 1")
        } else {
            // If we can't determine current chapter, just go to the beginning
            seekTo(0)
            showFlash("Beginning")
        }
    }, [chapterCues, seekTo, showFlash])


    const handleCycleSubtitles = useCallback(() => {
        if (!videoRef.current) return

        const textTracks = Array.from(videoRef.current.textTracks).filter(track => track.kind === "subtitles")
        if (textTracks.length === 0) {
            showFlash("No subtitle tracks")
            return
        }

        // Find currently showing track
        let currentTrackIndex = -1
        for (let i = 0; i < textTracks.length; i++) {
            if (textTracks[i].mode === "showing") {
                currentTrackIndex = i
                break
            }
        }

        // Cycle to next track or disable if we're at the end
        const nextIndex = currentTrackIndex + 1

        // Disable all tracks first
        for (let i = 0; i < textTracks.length; i++) {
            textTracks[i].mode = "disabled"
        }

        // Enable next track if available
        if (nextIndex < textTracks.length) {
            textTracks[nextIndex].mode = "showing"
            subtitleManager?.selectTrack(Number(textTracks[nextIndex].id))
            const trackName = textTracks[nextIndex].label || `Track ${nextIndex + 1}`
            showFlash(`Subtitles: ${trackName}`)
        } else {
            // If we've cycled through all, disable subtitles
            subtitleManager?.setNoTrack()
            showFlash("Subtitles: Off")
        }
    }, [subtitleManager])

    const handleCycleAudio = useCallback(() => {
        if (!videoRef.current) return

        const audioTracks = videoRef.current.audioTracks
        if (!audioTracks || audioTracks.length <= 1) {
            showFlash("No additional audio tracks")
            return
        }

        // Find currently enabled track
        let currentTrackIndex = -1
        for (let i = 0; i < audioTracks.length; i++) {
            if (audioTracks[i].enabled) {
                currentTrackIndex = i
                break
            }
        }

        // Cycle to next track
        const nextIndex = (currentTrackIndex + 1) % audioTracks.length

        // Disable all tracks first
        for (let i = 0; i < audioTracks.length; i++) {
            audioTracks[i].enabled = false
        }

        // Enable next track
        audioTracks[nextIndex].enabled = true
        audioManager?.selectTrack(nextIndex)

        const trackName = audioTracks[nextIndex].label || audioTracks[nextIndex].language || `Track ${nextIndex + 1}`
        showFlash(`Audio: ${trackName}`)
    }, [])

    const log = logger("VideoCoreKeybindings")

    const handleNextEpisode = useCallback(() => {
        // Placeholder for next episode functionality
        log.info("Next episode shortcut pressed - not implemented yet")
    }, [])

    const handlePreviousEpisode = useCallback(() => {
        // Placeholder for previous episode functionality
        log.info("Previous episode shortcut pressed - not implemented yet")
    }, [])

    const handleToggleFullscreen = useCallback(() => {
        // mediaStore.dispatch({
        //     type: fullscreen ? "mediaexitfullscreenrequest" : "mediaenterfullscreenrequest",
        // })

        React.startTransition(() => {
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        })
    }, [fullscreen])

    const handleTogglePictureInPicture = useCallback(() => {
        // mediaStore.dispatch({
        //     type: pip ? "mediaexitpiprequest" : "mediaenterpiprequest",
        // })

        React.startTransition(() => {
            setTimeout(() => {
                videoRef.current?.focus()
            }, 100)
        })
    }, [pip])

    // Add keyboard event listeners
    useEffect(() => {
        if (!active) return

        document.addEventListener("keydown", handleKeyboardShortcuts)

        return () => {
            document.removeEventListener("keydown", handleKeyboardShortcuts)
        }
    }, [handleKeyboardShortcuts, active])

    // Handle fullscreen state changes to ensure video gets focused
    useEffect(() => {
        if (!active) return

        const handleFullscreenChange = () => {
            // Small delay to ensure fullscreen transition is complete
            setTimeout(() => {
                if (document.fullscreenElement && videoRef.current) {
                    videoRef.current.focus()
                }
            }, 100)
        }

        document.addEventListener("fullscreenchange", handleFullscreenChange)

        return () => {
            document.removeEventListener("fullscreenchange", handleFullscreenChange)
        }
    }, [active])

    return null
}
