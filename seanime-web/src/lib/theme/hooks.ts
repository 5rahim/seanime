import { serverStatusAtom } from "@/atoms/server-status"
import { ThemeSettings } from "@/lib/server/types"
import { useAtomValue } from "jotai/react"

export const THEME_DEFAULT_VALUES: { [key: string]: any } = {
    animeEntryScreenLayout: "stacked",
    smallerEpisodeCarouselSize: false,
    expandSidebarOnHover: false,
    backgroundColor: "#0c0c0c",
    sidebarBackgroundColor: "#0c0c0c",
    libraryScreenBanner: "episode",
    libraryScreenBannerPosition: "50% 50%",
    libraryScreenCustomBanner: "",
    libraryScreenCustomBannerAutoDim: 1, // not implemented
    libraryScreenShowCustomBackground: false,
    libraryScreenCustomBackground: "",
    libraryScreenCustomBackgroundAutoDim: 1, // not implemented
}

export const enum ThemeLibraryScreenBanner {
    Episode = "episode",
    Custom = "custom",
}

export function useThemeSettings(): ThemeSettings {
    const serverStatus = useAtomValue(serverStatusAtom)
    return {
        animeEntryScreenLayout: getThemeValue("animeEntryScreenLayout", serverStatus?.themeSettings),
        smallerEpisodeCarouselSize: getThemeValue("smallerEpisodeCarouselSize", serverStatus?.themeSettings),
        expandSidebarOnHover: getThemeValue("expandSidebarOnHover", serverStatus?.themeSettings),
        backgroundColor: getThemeValue("backgroundColor", serverStatus?.themeSettings),
        sidebarBackgroundColor: getThemeValue("sidebarBackgroundColor", serverStatus?.themeSettings),
        libraryScreenBanner: getThemeValue("libraryScreenBanner", serverStatus?.themeSettings),
        libraryScreenBannerPosition: getThemeValue("libraryScreenBannerPosition", serverStatus?.themeSettings),
        libraryScreenCustomBanner: getThemeValue("libraryScreenCustomBanner", serverStatus?.themeSettings),
        libraryScreenCustomBannerAutoDim: getThemeValue("libraryScreenCustomBannerAutoDim", serverStatus?.themeSettings),
        libraryScreenShowCustomBackground: getThemeValue("libraryScreenShowCustomBackground", serverStatus?.themeSettings),
        libraryScreenCustomBackground: getThemeValue("libraryScreenCustomBackground", serverStatus?.themeSettings),
        libraryScreenCustomBackgroundAutoDim: getThemeValue("libraryScreenCustomBackgroundAutoDim", serverStatus?.themeSettings),
    }
}

function getThemeValue(key: string, settings: ThemeSettings | undefined | null): any {
    if (!settings) {
        return THEME_DEFAULT_VALUES[key]
    }
    const val = (settings as any)[key]
    if (typeof val === "string" && val === "") {
        return THEME_DEFAULT_VALUES[key]
    } else {
        return val
    }
}
