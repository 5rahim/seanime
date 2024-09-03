import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { cn } from "@/components/ui/core/styling"
import { useThemeSettings } from "@/lib/theme/hooks"
import Image from "next/image"
import React from "react"

type MediaEntryPageSmallBannerProps = {
    bannerImage?: string
}

export function MediaEntryPageSmallBanner(props: MediaEntryPageSmallBannerProps) {

    const {
        bannerImage,
        ...rest
    } = props

    const ts = useThemeSettings()

    return (
        <>
            <div
                className={cn(
                    "h-[30rem] w-full flex-none object-cover object-center absolute -top-[5rem] overflow-hidden bg-[--background]",
                    (ts.hideTopNavbar || process.env.NEXT_PUBLIC_PLATFORM === "desktop") && "h-[27rem]",
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                )}
            >
                <div
                    className="w-full absolute z-[2] top-0 h-[8rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via"
                />
                <div className="absolute w-full h-full">
                    {(!!bannerImage) && <Image
                        src={bannerImage || ""}
                        alt="banner image"
                        fill
                        quality={100}
                        priority
                        sizes="100vw"
                        className="object-cover object-center z-[1]"
                    />}
                </div>
                <div
                    className="w-full z-[3] absolute bottom-0 h-[32rem] bg-gradient-to-t from-[--background] via-[--background] via-50% to-transparent"
                />

            </div>
        </>
    )
}
