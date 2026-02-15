import { vc_anime4kManager } from "@/app/(main)/_features/video-core/video-core"
import { VideoCore_VideoPlaybackInfo } from "@/app/(main)/_features/video-core/video-core.atoms"
import { vc_showStatsForNerdsAtom } from "@/app/(main)/_features/video-core/video-core.atoms"
import { cn } from "@/components/ui/core/styling"
import { useAtomValue } from "jotai"
import React, { useEffect, useMemo, useState } from "react"
import { vc_miniPlayer } from "./video-core-atoms"

interface VideoCoreStatsForNerdsProps {
    playbackInfo: VideoCore_VideoPlaybackInfo | null
    videoRef: React.RefObject<HTMLVideoElement>
}

interface PerformanceData {
    currentFps: number
    totalFrames: number
    droppedFrames: number
    decodedFrames: number
    corruptedFrames: number
    renderTime: number
    displaySize: { width: number; height: number }
    streamSize: { width: number; height: number }
    availableBuffer: number
    rate: number
    networkState: string
    readyState: string
    currentTime: number
    duration: number
    captureTime: number
    expectedDisplayTime: number
}

export function VideoCoreStatsForNerds({ playbackInfo, videoRef }: VideoCoreStatsForNerdsProps) {
    const videoElement = videoRef.current
    const isMiniPlayer = useAtomValue(vc_miniPlayer)
    const anime4kManager = useAtomValue(vc_anime4kManager)
    const showStats = useAtomValue(vc_showStatsForNerdsAtom)

    const [performance, setPerformance] = useState<PerformanceData | null>(null)
    const [anime4kStats, setAnime4kStats] = useState<{ currentOption: string; totalFrameDrops: number; currentFps: number } | null>(null)
    const [fpsHistory, setFpsHistory] = useState<number[]>([])
    const [a4kFpsHistory, setA4kFpsHistory] = useState<number[]>([])

    useEffect(() => {
        const video = videoRef.current
        if (!video || !video.requestVideoFrameCallback) return

        let frameCallbackId: number
        let lastFrameData: VideoFrameCallbackMetadata | null = null

        const collectFrameStats = (timestamp: number, frameData: VideoFrameCallbackMetadata) => {
            if (!showStats) {
                return
            }
            if (lastFrameData) {
                const timeDelta = frameData.mediaTime - lastFrameData.mediaTime
                const frameDelta = frameData.presentedFrames - lastFrameData.presentedFrames
                const frameTime = timeDelta / frameDelta
                const derivedFps = frameTime > 0 ? 1 / frameTime : 0
                const quality = video.getVideoPlaybackQuality?.()
                const bufferAhead = calculateBufferAhead(video, frameData.mediaTime)

                const networkStates = ["NETWORK_EMPTY", "NETWORK_IDLE", "NETWORK_LOADING", "NETWORK_NO_SOURCE"]
                const readyStates = ["HAVE_NOTHING", "HAVE_METADATA", "HAVE_CURRENT_DATA", "HAVE_FUTURE_DATA", "HAVE_ENOUGH_DATA"]

                setPerformance({
                    currentFps: Math.round(derivedFps * 100) / 100,
                    totalFrames: frameData.presentedFrames,
                    droppedFrames: quality?.droppedVideoFrames ?? 0,
                    decodedFrames: quality?.totalVideoFrames ?? 0,
                    corruptedFrames: quality?.corruptedVideoFrames ?? 0,
                    renderTime: Math.round((frameData.processingDuration || 0) * 100) / 100,
                    displaySize: { width: video.clientWidth, height: video.clientHeight },
                    streamSize: { width: video.videoWidth, height: video.videoHeight },
                    availableBuffer: bufferAhead,
                    rate: video.playbackRate,
                    networkState: networkStates[video.networkState] || "UNKNOWN",
                    readyState: readyStates[video.readyState] || "UNKNOWN",
                    currentTime: Math.round(video.currentTime * 100) / 100,
                    duration: Math.round(video.duration * 100) / 100,
                    captureTime: Math.round((frameData.captureTime || 0) * 100) / 100,
                    expectedDisplayTime: Math.round((frameData.expectedDisplayTime || 0) * 100) / 100,
                })

                setFpsHistory(prev => {
                    const updated = [...prev, derivedFps]
                    return updated.slice(-30)
                })
            }

            lastFrameData = frameData

            setTimeout(() => {
                if (videoRef.current) {
                    frameCallbackId = videoRef.current.requestVideoFrameCallback(collectFrameStats)
                }
            }, 500)
        }

        frameCallbackId = video.requestVideoFrameCallback(collectFrameStats)

        return () => {
            video.cancelVideoFrameCallback(frameCallbackId)
        }
    }, [videoRef, showStats])

    useEffect(() => {
        if (!anime4kManager || !showStats) return

        const updateAnime4kStats = () => {
            const stats = anime4kManager.getStats()
            setAnime4kStats({
                currentOption: stats.currentOption,
                totalFrameDrops: stats.totalFrameDrops,
                currentFps: stats.currentFps,
            })

            setA4kFpsHistory(prev => {
                const updated = [...prev, stats.currentFps]
                return updated.slice(-30)
            })
        }

        updateAnime4kStats()
        const interval = setInterval(updateAnime4kStats, 1000)

        return () => clearInterval(interval)
    }, [anime4kManager, showStats])

    const mediaInfo = useMemo(() => {
        if (!playbackInfo) return null

        const details: { label: string; value: string }[] = []

        if (playbackInfo.localFile) {
            details.push({ label: "File", value: playbackInfo.localFile.name })
        }
        if (playbackInfo.streamPath) {
            details.push({ label: "Path", value: playbackInfo.streamPath })
        }
        if (playbackInfo.streamUrl) {
            details.push({ label: "Stream", value: playbackInfo.streamUrl?.replace("{{SERVER_URL}}", "") })
        }
        if (playbackInfo.media) {
            details.push({
                label: "Title",
                value: playbackInfo.media.title?.userPreferred || playbackInfo.media.title?.romaji || "",
            })
            if (playbackInfo.episode) {
                const epNum = playbackInfo.episode.episodeNumber
                const aniDbNum = playbackInfo.episode.aniDBEpisode
                details.push({
                    label: "Episode",
                    value: `Ep ${epNum}${aniDbNum ? ` (AniDB ${aniDbNum})` : ""}`,
                })
            }
        }

        if (playbackInfo.mkvMetadata) {
            if (playbackInfo.mkvMetadata.mimeCodec) {
                details.push({ label: "MIME Type", value: playbackInfo.mkvMetadata.mimeCodec })
            }
            const videoTrack = playbackInfo.mkvMetadata.videoTracks?.[0]
            if (videoTrack?.codecID) {
                details.push({ label: "Video Codec", value: videoTrack.codecID })
            }
            if (playbackInfo.mkvMetadata.audioTracks?.length) {
                details.push({
                    label: "Audio Codecs",
                    value: playbackInfo.mkvMetadata.audioTracks.map(t => t.codecID).join(", "),
                })
            }
        }

        return details
    }, [playbackInfo])

    if (!mediaInfo || mediaInfo.length === 0 || isMiniPlayer) return null

    return (
        <div
            data-vc-element="stats-for-nerds"
            className="absolute top-24 left-4 z-[100] bg-black/80 text-white p-4 rounded-md font-mono text-xs pointer-events-none select-none max-w-lg"
        >
            <p className="font-bold mb-2">Stats for Nerds</p>
            <div className="space-y-1">
                {mediaInfo.map((item, idx) => (
                    <div key={idx} className="flex gap-2">
                        <span className="font-semibold text-gray-400 flex-none">{item.label}:</span>
                        <span className="break-all line-clamp-3 flex-1 select-all">{item.value}</span>
                    </div>
                ))}

                {performance && (
                    <>
                        <div className="border-t border-gray-700 my-2 pt-2"></div>
                        <StatLine
                            label="Display / Video"
                            value={`${performance.displaySize.width}x${performance.displaySize.height} / ${performance.streamSize.width}x${performance.streamSize.height}`}
                        />
                        <StatLine
                            label="Framerate"
                            value={`${performance.currentFps.toFixed(2)} fps`}
                        />
                        {fpsHistory.length > 0 && <div className="mt-2"><FpsGraph history={fpsHistory} /></div>}

                        <StatLine
                            label="Frames (Total / Dropped)"
                            value={`${performance.totalFrames} / ${performance.droppedFrames}`}
                        />
                        <StatLine
                            label="Render Time"
                            value={`${performance.renderTime.toFixed(2)} ms`}
                        />
                        <StatLine
                            label="Buffer Ahead"
                            value={`${performance.availableBuffer.toFixed(2)} s`}
                        />
                        <StatLine
                            label="Playback Rate"
                            value={`${performance.rate}x`}
                        />
                        <StatLine
                            label="Decoded / Corrupted"
                            value={`${performance.decodedFrames} / ${performance.corruptedFrames}`}
                        />
                        <StatLine
                            label="Network State"
                            value={performance.networkState}
                        />
                        <StatLine
                            label="Ready State"
                            value={performance.readyState}
                        />
                        <StatLine
                            label="Time / Duration"
                            value={`${performance.currentTime.toFixed(2)}s / ${performance.duration.toFixed(2)}s`}
                        />
                        {/*<StatLine*/}
                        {/*    label="Capture / Display"*/}
                        {/*    value={`${performance.captureTime.toFixed(2)}ms / ${performance.expectedDisplayTime.toFixed(2)}ms`}*/}
                        {/*/>*/}

                        {(anime4kStats && anime4kStats.currentOption !== "off") && (
                            <>
                                <div className="border-t border-gray-700 my-2 pt-2"></div>
                                <StatLine
                                    label="Anime4K Mode"
                                    value={anime4kStats.currentOption}
                                />
                                <StatLine
                                    label="A4K Framerate"
                                    value={`${anime4kStats.currentFps.toFixed(2)} fps`}
                                />
                                <StatLine
                                    label="A4K Frame Drops"
                                    value={anime4kStats.totalFrameDrops.toString()}
                                />
                                {a4kFpsHistory.length > 0 && (
                                    <div className="mt-2">
                                        <FpsGraph history={a4kFpsHistory} />
                                    </div>
                                )}
                            </>
                        )}
                    </>
                )}
            </div>
        </div>
    )
}

const StatLine = ({ label, value }: { label: string, value: string }) => (
    <div className="flex gap-2">
        <span className="font-semibold text-gray-400 flex-none">{label}:</span>
        <span className="break-all line-clamp-3 flex-1">{value}</span>
    </div>
)

function calculateBufferAhead(video: HTMLVideoElement, playheadTime: number): number {
    const bufferedRanges = video.buffered
    if (!bufferedRanges || bufferedRanges.length === 0) return 0
    for (let i = 0; i < bufferedRanges.length; i++) {
        const rangeStart = bufferedRanges.start(i)
        const rangeEnd = bufferedRanges.end(i)

        if (playheadTime >= rangeStart && playheadTime <= rangeEnd) {
            return rangeEnd - playheadTime
        }
    }

    return 0
}

const FpsGraph = ({ history }: { history: number[] }) => {
    const maxFps = 26

    const getBarColor = (fps: number) => {
        if (fps >= 22) return "bg-green-500"
        if (fps >= 18) return "bg-orange-500"
        return "bg-red-500"
    }

    return (
        <div className="flex items-end gap-[2px] h-7 w-full">
            {history.map((fps, idx) => {
                const heightPercent = Math.min((fps / maxFps) * 100, 100)
                return (
                    <div
                        key={idx}
                        className={cn(`flex-1`, getBarColor(fps))}
                        style={{ height: `${heightPercent}%` }}
                        title={fps.toFixed(1) + " fps"}
                    />
                )
            })}
        </div>
    )
}
