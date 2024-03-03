/**
 * This file contains bottom gradients for items
 * They change responsively based on the UI settings
 */

import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function AnimeListItemBottomGradient() {

    const ts = useThemeSettings()

    if (!!ts.libraryScreenCustomBackgroundImage || ts.hasCustomBackgroundColor) {
        return (
            <div
                className="z-[5] absolute bottom-0 w-full h-[20%] opacity-80 bg-gradient-to-t from-[--background] to-transparent"
            />
        )
    }

    return (
        <div
            className="z-[5] absolute bottom-0 w-full h-[50%] bg-gradient-to-t from-[--background] to-transparent"
        />
    )
}


export function EpisodeItemBottomGradient() {

    const ts = useThemeSettings()

    if (!!ts.libraryScreenCustomBackgroundImage || ts.hasCustomBackgroundColor) {
        return (
            <div
                className="z-[1] absolute bottom-0 w-full h-full opacity-70 md:h-[60%] bg-gradient-to-t from-[--background] to-transparent"
            />
        )
    }

    return (
        <div
            className="z-[1] absolute bottom-0 w-full h-full md:h-[80%] bg-gradient-to-t from-[--background] to-transparent"
        />
    )
}
