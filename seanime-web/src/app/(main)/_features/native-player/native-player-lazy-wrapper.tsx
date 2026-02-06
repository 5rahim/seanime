import { VideoCoreProvider } from "@/app/(main)/_features/video-core/video-core"
import React from "react"
import { NativePlayer } from "./native-player"

export default function NativePlayerLazyWrapper() {
    return (
        <VideoCoreProvider key="native-player" id="native-player">
            <NativePlayer />
        </VideoCoreProvider>
    )
}
