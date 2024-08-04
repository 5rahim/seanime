import { Models_Theme } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"

export type ThemeSettings = Omit<Models_Theme, "id">
export const THEME_DEFAULT_VALUES: ThemeSettings = {
    enableColorSettings: false,
    animeEntryScreenLayout: "stacked",
    smallerEpisodeCarouselSize: false,
    expandSidebarOnHover: false,
    backgroundColor: "#070707",
    accentColor: "#6152df",
    sidebarBackgroundColor: "#070707",
    hideTopNavbar: false,
    enableMediaCardBlurredBackground: false,
    libraryScreenBannerType: ThemeLibraryScreenBannerType.Dynamic,
    libraryScreenCustomBannerImage: "",
    libraryScreenCustomBannerPosition: "50% 50%",
    libraryScreenCustomBannerOpacity: 100,
    libraryScreenCustomBackgroundImage: "",
    libraryScreenCustomBackgroundOpacity: 10,
    disableLibraryScreenGenreSelector: false,
    libraryScreenCustomBackgroundBlur: "",
    enableMediaPageBlurredBackground: false,
    disableSidebarTransparency: false,
}

export const enum ThemeLibraryScreenBannerType {
    Dynamic = "dynamic",
    Custom = "custom",
}

export type ThemeSettingsHook = {
    hasCustomBackgroundColor: boolean
} & ThemeSettings

/**
 * Get the current theme settings
 * This hook will return the default values if some values are not set
 */
export function useThemeSettings(): ThemeSettingsHook {
    const serverStatus = useServerStatus()
    return {
        enableColorSettings: getThemeValue("enableColorSettings", serverStatus?.themeSettings),
        animeEntryScreenLayout: getThemeValue("animeEntryScreenLayout", serverStatus?.themeSettings),
        smallerEpisodeCarouselSize: getThemeValue("smallerEpisodeCarouselSize", serverStatus?.themeSettings),
        expandSidebarOnHover: getThemeValue("expandSidebarOnHover", serverStatus?.themeSettings),
        backgroundColor: getThemeValue("backgroundColor", serverStatus?.themeSettings),
        accentColor: getThemeValue("accentColor", serverStatus?.themeSettings),
        hideTopNavbar: getThemeValue("hideTopNavbar", serverStatus?.themeSettings),
        enableMediaCardBlurredBackground: getThemeValue("enableMediaCardBlurredBackground", serverStatus?.themeSettings),
        sidebarBackgroundColor: getThemeValue("sidebarBackgroundColor", serverStatus?.themeSettings),
        libraryScreenBannerType: getThemeValue("libraryScreenBannerType", serverStatus?.themeSettings),
        libraryScreenCustomBannerImage: getThemeValue("libraryScreenCustomBannerImage", serverStatus?.themeSettings),
        libraryScreenCustomBannerPosition: getThemeValue("libraryScreenCustomBannerPosition", serverStatus?.themeSettings),
        libraryScreenCustomBannerOpacity: getThemeValue("libraryScreenCustomBannerOpacity", serverStatus?.themeSettings),
        libraryScreenCustomBackgroundImage: getThemeValue("libraryScreenCustomBackgroundImage", serverStatus?.themeSettings),
        libraryScreenCustomBackgroundOpacity: getThemeValue("libraryScreenCustomBackgroundOpacity", serverStatus?.themeSettings),
        disableLibraryScreenGenreSelector: getThemeValue("disableLibraryScreenGenreSelector", serverStatus?.themeSettings),
        libraryScreenCustomBackgroundBlur: getThemeValue("libraryScreenCustomBackgroundBlur", serverStatus?.themeSettings),
        enableMediaPageBlurredBackground: getThemeValue("enableMediaPageBlurredBackground", serverStatus?.themeSettings),
        disableSidebarTransparency: getThemeValue("disableSidebarTransparency", serverStatus?.themeSettings),

        hasCustomBackgroundColor: !!serverStatus?.themeSettings?.backgroundColor && serverStatus?.themeSettings?.backgroundColor !== THEME_DEFAULT_VALUES.backgroundColor,
    }
}

function getThemeValue(key: string, settings: ThemeSettings | undefined | null): any {
    if (!settings) {
        // @ts-ignore
        return THEME_DEFAULT_VALUES[key]
    }
    const val = (settings as any)[key]
    if (typeof val === "string" && val === "") {
        // @ts-ignore
        return THEME_DEFAULT_VALUES[key]
    } else if (typeof val === "number" && val === 0) {
        // @ts-ignore
        return THEME_DEFAULT_VALUES[key]
    } else {
        return val
    }
}
