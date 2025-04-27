"use client"

import { DirectorySelector } from "@/components/shared/directory-selector"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { upath } from "@/lib/helpers/upath"
import { MediaPlayerInstance } from "@vidstack/react"
import React, { useRef } from "react"

export default function TestPage() {

    const playerRef = useRef<MediaPlayerInstance>(null)

    React.useEffect(() => {
        const path = upath.join("C:", "Users", "test", "Downloads", "test.txt")
        console.log(path)
    }, [])

    return <AppLayoutStack>

        <DirectorySelector
            onSelect={() => { }}
            value={""}
        />

        {/* <SeaMediaPlayerLayout
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
        /> */}
    </AppLayoutStack>
}

