"use client"

import { VideoCore, VideoCoreProvider } from "@/app/(main)/_features/video-core/video-core"
import { AppLayoutStack } from "@/components/ui/app-layout"
import React from "react"

export default function TestPage() {

    return <AppLayoutStack className="h-full w-full relative">


        <VideoCoreProvider id="test">
            <VideoCore
                id="test"
                state={{
                    active: true,
                    playbackInfo: {
                        id: "Test",
                        streamType: "onlinestream",
                        streamUrl: "https://stream.mux.com/fXNzVtmtWuyz00xnSrJg4OJH6PyNo6D02UzmgeKGkP5YQ/high.mp4",
                    },
                    playbackError: null,
                    loadingState: null,
                }}
                inline
                onTerminateStream={() => {}}
                onFileUploaded={() => {}}
            />
        </VideoCoreProvider>

    </AppLayoutStack>
}

