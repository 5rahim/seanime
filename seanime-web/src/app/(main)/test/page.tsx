"use client"

import { MediaPlayerInstance } from "@vidstack/react"
import { useRef } from "react"
import { SeaMediaPlayer } from "../_features/sea-media-player/sea-media-player"
import { SeaMediaPlayerLayout } from "../_features/sea-media-player/sea-media-player-layout"

export default function TestPage() {

    const playerRef = useRef<MediaPlayerInstance>(null)

    return <div>
        <SeaMediaPlayerLayout
            mediaId={130003}
            title={"test"}
            episodes={[]}

            mediaPlayer={
                <SeaMediaPlayer
                    url={{
                        src: "",
                        type: "video/webm",
                    }}
                    isPlaybackError={false}
                    isLoading={false}
                    playerRef={playerRef}
                    poster={""}
                    onProviderChange={() => { }}
                    onProviderSetup={() => { }}
                    onCanPlay={() => { }}
                    onGoToNextEpisode={() => { }}
                    tracks={[]}
                    mediaInfoDuration={0}
                    loadingText={<>
                        <p>Extracting video metadata...</p>
                        <p>This might take a while.</p>
                    </>}
                />
            }
            episodeList={[]}
        />
    </div>
}

