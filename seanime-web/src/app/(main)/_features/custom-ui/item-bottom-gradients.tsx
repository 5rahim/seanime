/**
 * This file contains bottom gradients for items
 * They change responsively based on the UI settings
 */

import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function MediaCardBodyBottomGradient() {

    const ts = useThemeSettings()

    if (!!ts.libraryScreenCustomBackgroundImage || ts.hasCustomBackgroundColor) {
        return (
            <div
                data-media-card-body-bottom-gradient
                className="z-[5] absolute inset-x-0 bottom-0 w-full h-[40%] opacity-80 bg-gradient-to-t from-[#0c0c0c] to-transparent"
            />
        )
    }

    return (
        <div
            data-media-card-body-bottom-gradient
            className="z-[5] absolute inset-x-0 bottom-0 w-full opacity-90 to-40% h-[50%] bg-gradient-to-t from-[#0c0c0c] to-transparent"
        />
    )
}


export function EpisodeItemBottomGradient() {

    const ts = useThemeSettings()

    // if (!!ts.libraryScreenCustomBackgroundImage || ts.hasCustomBackgroundColor) {
    //     return (
    //         <div
    //             className="z-[1] absolute inset-x-0 bottom-0 w-full h-full opacity-80 md:h-[60%] bg-gradient-to-t from-[--background] to-transparent"
    //         />
    //     )
    // }

    if (ts.useLegacyEpisodeCard) {
        return <div
            data-episode-item-bottom-gradient
            className="z-[1] absolute inset-x-0 bottom-0 w-full h-full opacity-90 md:h-[80%] bg-gradient-to-t from-[#0c0c0c] to-transparent"
        />
    }

    return <div
        data-episode-item-bottom-gradient
        className="z-[1] absolute inset-x-0 bottom-0 w-full h-full opacity-50 md:h-[70%] bg-gradient-to-t from-[#0c0c0c] to-transparent"
    />
}
