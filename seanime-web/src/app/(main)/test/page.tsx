"use client"

import { VideoCore, VideoCoreProvider } from "@/app/(main)/_features/video-core/video-core"
import { AppLayoutStack } from "@/components/ui/app-layout"
import React from "react"

export default function TestPage() {

    return <AppLayoutStack className="h-full w-full relative">


        <VideoCoreProvider id="test">
            <div className="w-full max-w-7xl aspect-video mx-auto border rounded-lg overflow-hidden">
                <VideoCore
                    id="test"
                    state={{
                        active: true,
                        playbackInfo: {
                            id: "Test",
                            playbackType: "onlinestream",
                            streamUrl: "https://devstreaming-cdn.apple.com/videos/streaming/examples/bipbop_adv_example_hevc/master.m3u8",
                            // streamUrl: "https://stream.mux.com/VZtzUzGRv02OhRnZCxcNg49OilvolTqdnFLEqBsTwaxU/low.mp4",
                            streamType: "stream",
                            subtitleTracks: [
                                {
                                    index: 0,
                                    src: "http://127.0.0.1:43210/english.vtt",
                                    label: "English",
                                    language: "en",
                                    type: "vtt",
                                    default: true,
                                    // useLibassRenderer: true
                                },
                            ],
                        },
                        playbackError: null,
                        loadingState: null,
                    }}
                    inline
                    onTerminateStream={() => {}}
                    onFileUploaded={() => {}}
                />
            </div>
        </VideoCoreProvider>

    </AppLayoutStack>
}

