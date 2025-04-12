import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { useThemeSettings } from "@/lib/theme/hooks"
import React from "react"

export function MediaEntryPageLoadingDisplay() {
    const ts = useThemeSettings()

    if (!!ts.libraryScreenCustomBackgroundImage) {
        return null
    }

    return (
        <div data-media-entry-page-loading-display className="__header h-[30rem] fixed left-0 top-0 w-full">
            <div
                className={cn(
                    "h-[30rem] w-full flex-none object-cover object-center absolute top-0 overflow-hidden",
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                )}
            >
                <div
                    className="w-full absolute z-[1] top-0 h-[15rem] bg-gradient-to-b from-[--background] to-transparent via"
                />
                <Skeleton className="h-full absolute w-full" />
                <div
                    className="w-full absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-transparent to-transparent"
                />
            </div>
        </div>
    )
}
