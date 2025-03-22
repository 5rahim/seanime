"use client"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { cn } from "@/components/ui/core/styling"
import { getAssetUrl } from "@/lib/server/assets"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { motion } from "framer-motion"
import React, { useEffect } from "react"
import { useWindowScroll } from "react-use"

type CustomLibraryBannerProps = {
    discrete?: boolean
    isLibraryScreen?: boolean // Anime library or manga library
}

export function CustomLibraryBanner(props: CustomLibraryBannerProps) {
    /**
     * Library screens: Shows the custom banner IF theme settings are set to use a custom banner
     * Other pages: Shows the custom banner
     */
    const { discrete, isLibraryScreen } = props
    const ts = useThemeSettings()
    const image = React.useMemo(() => ts.libraryScreenCustomBannerImage ? getAssetUrl(ts.libraryScreenCustomBannerImage) : "",
        [ts.libraryScreenCustomBannerImage])
    const [dimmed, setDimmed] = React.useState(false)

    const { y } = useWindowScroll()

    useEffect(() => {
        if (y > 100)
            setDimmed(true)
        else
            setDimmed(false)
    }, [(y > 100)])

    if (isLibraryScreen && ts.libraryScreenBannerType !== ThemeLibraryScreenBannerType.Custom) return null
    if (discrete && !!ts.libraryScreenCustomBackgroundImage) return null
    if (!image) return null

    return (
        <>
            {!discrete && <div
                data-custom-library-banner-top-spacer
                className={cn(
                    "py-20",
                    ts.hideTopNavbar && "py-28",
                )}
            ></div>}
            <div
                data-custom-library-banner-container
                className={cn(
                    "__header h-[30rem] z-[1] top-0 w-full fixed group/library-header transition-opacity duration-1000",
                    discrete && "opacity-20",
                    !!ts.libraryScreenCustomBackgroundImage && "absolute", // If there's a background image, make the banner absolute
                    (!ts.libraryScreenCustomBackgroundImage && dimmed) && "opacity-5", // If the user has scrolled down, dim the banner
                    !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    "scroll-locked-offset",
                )}
            >
                {(!ts.disableSidebarTransparency && !discrete) && <div
                    data-custom-library-banner-top-gradient
                    className="hidden lg:block h-full absolute z-[2] w-[20rem] opacity-70 left-0 top-0 bg-gradient bg-gradient-to-r from-[var(--background)] to-transparent"
                />}

                <div
                    data-custom-library-banner-bottom-gradient
                    className="w-full z-[3] absolute bottom-[-5rem] h-[5rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent"
                />
                <motion.div
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 1, delay: 0.2 }}
                    className={cn(
                        "h-[30rem] z-[0] w-full flex-none absolute top-0 overflow-hidden",
                        "scroll-locked-offset",
                    )}
                    data-custom-library-banner-inner-container
                >
                    <div
                        data-custom-library-banner-top-gradient
                        className={cn(
                            "CUSTOM_LIB_BANNER_TOP_FADE w-full absolute z-[2] top-0 h-[5rem] opacity-40 bg-gradient-to-b from-[--background] to-transparent via",
                            discrete && "opacity-70",
                        )}
                    />
                    <div
                        data-custom-library-banner-image
                        className={cn(
                            "CUSTOM_LIB_BANNER_IMG z-[1] absolute inset-0 w-full h-full bg-cover bg-no-repeat transition-opacity duration-1000",
                            "scroll-locked-offset",
                        )}
                        style={{
                            backgroundImage: `url(${image})`,
                            backgroundPosition: ts.libraryScreenCustomBannerPosition || "50% 50%",
                            opacity: (ts.libraryScreenCustomBannerOpacity || 100) / 100,
                            backgroundRepeat: "no-repeat",
                            backgroundSize: "cover",
                        }}
                    />
                    <div
                        data-custom-library-banner-bottom-gradient
                        className={cn(
                            "CUSTOM_LIB_BANNER_BOTTOM_FADE w-full z-[2] absolute bottom-0 h-[20rem] bg-gradient-to-t from-[--background] via-opacity-50 via-10% to-transparent via",
                            discrete && "via-50% via-opacity-100 h-[40rem]",
                        )}
                    />
                </motion.div>
            </div>
        </>
    )

}
