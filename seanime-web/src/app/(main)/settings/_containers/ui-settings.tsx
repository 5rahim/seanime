import { useUpdateTheme } from "@/api/hooks/theme.hooks"
import { useCustomCSS } from "@/components/shared/custom-css-provider"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { RadioGroup } from "@/components/ui/radio-group"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ANIME_COLLECTION_SORTING_OPTIONS, CONTINUE_WATCHING_SORTING_OPTIONS, MANGA_COLLECTION_SORTING_OPTIONS } from "@/lib/helpers/filtering"
import { __navigationPreloadModeAtom, NavigationPreloadMode } from "@/lib/navigation-preload-settings"
import { THEME_COLOR_BANK } from "@/lib/theme/theme-bank"
import {
    THEME_DEFAULT_VALUES,
    ThemeLibraryScreenBannerType,
    ThemeMediaPageBannerSizeOptions,
    ThemeMediaPageBannerType,
    ThemeMediaPageBannerTypeOptions,
    useThemeSettings,
} from "@/lib/theme/theme-hooks.ts"
import { __isDesktop__ } from "@/types/constants"
import { colord } from "colord"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React, { useState } from "react"
import { useFormContext, UseFormReturn, useWatch } from "react-hook-form"
import { LuChevronRight } from "react-icons/lu"
import { toast } from "sonner"
import { z } from "zod"
import { useIsSimulatedUser } from "../../_hooks/use-server-status"
import { useServerStatus } from "../../_hooks/use-server-status"
import { SettingsCard } from "../_components/settings-card"
import { SettingsIsDirty } from "../_components/settings-submit-button"

const themeSchema = defineSchema(({ z }) => z.object({
    animeEntryScreenLayout: z.string().min(0).default(THEME_DEFAULT_VALUES.animeEntryScreenLayout),
    smallerEpisodeCarouselSize: z.boolean().default(THEME_DEFAULT_VALUES.smallerEpisodeCarouselSize),
    expandSidebarOnHover: z.boolean().default(THEME_DEFAULT_VALUES.expandSidebarOnHover),
    enableColorSettings: z.boolean().default(false),
    backgroundColor: z.string().min(0).default(THEME_DEFAULT_VALUES.backgroundColor).transform(n => n.trim()),
    accentColor: z.string().min(0).default(THEME_DEFAULT_VALUES.accentColor).transform(n => n.trim()),
    sidebarBackgroundColor: z.string().min(0).default(THEME_DEFAULT_VALUES.sidebarBackgroundColor),
    hideTopNavbar: z.boolean().default(THEME_DEFAULT_VALUES.hideTopNavbar),
    enableMediaCardBlurredBackground: z.boolean().default(THEME_DEFAULT_VALUES.enableMediaCardBlurredBackground),

    libraryScreenBannerType: z.string().default(THEME_DEFAULT_VALUES.libraryScreenBannerType),
    libraryScreenCustomBannerImage: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerImage),
    libraryScreenCustomBannerPosition: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerPosition),
    libraryScreenCustomBannerOpacity: z.number().transform(v => v === 0 ? 100 : v).default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerOpacity),
    libraryScreenCustomBackgroundImage: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundImage),
    libraryScreenCustomBackgroundOpacity: z.number()
        .transform(v => v === 0 ? 100 : v)
        .default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundOpacity),
    libraryScreenCustomBackgroundBlur: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundBlur),
    enableMediaPageBlurredBackground: z.boolean().default(THEME_DEFAULT_VALUES.enableMediaPageBlurredBackground),
    disableSidebarTransparency: z.boolean().default(THEME_DEFAULT_VALUES.disableSidebarTransparency),
    disableLibraryScreenGenreSelector: z.boolean().default(false),
    useLegacyEpisodeCard: z.boolean().default(THEME_DEFAULT_VALUES.useLegacyEpisodeCard),
    disableCarouselAutoScroll: z.boolean().default(THEME_DEFAULT_VALUES.disableCarouselAutoScroll),
    mediaPageBannerType: z.string().default(THEME_DEFAULT_VALUES.mediaPageBannerType),
    mediaPageBannerSize: z.string().default(THEME_DEFAULT_VALUES.mediaPageBannerSize),
    mediaPageBannerInfoBoxSize: z.string().default(THEME_DEFAULT_VALUES.mediaPageBannerInfoBoxSize),
    showEpisodeCardAnimeInfo: z.boolean().default(THEME_DEFAULT_VALUES.showEpisodeCardAnimeInfo),
    continueWatchingDefaultSorting: z.string().default(THEME_DEFAULT_VALUES.continueWatchingDefaultSorting),
    animeLibraryCollectionDefaultSorting: z.string().default(THEME_DEFAULT_VALUES.animeLibraryCollectionDefaultSorting),
    mangaLibraryCollectionDefaultSorting: z.string().default(THEME_DEFAULT_VALUES.mangaLibraryCollectionDefaultSorting),
    showAnimeUnwatchedCount: z.boolean().default(THEME_DEFAULT_VALUES.showAnimeUnwatchedCount),
    showMangaUnreadCount: z.boolean().default(THEME_DEFAULT_VALUES.showMangaUnreadCount),
    hideEpisodeCardDescription: z.boolean().default(THEME_DEFAULT_VALUES.hideEpisodeCardDescription),
    hideDownloadedEpisodeCardFilename: z.boolean().default(THEME_DEFAULT_VALUES.hideDownloadedEpisodeCardFilename),
    customCSS: z.string().default(THEME_DEFAULT_VALUES.customCSS),
    mobileCustomCSS: z.string().default(THEME_DEFAULT_VALUES.mobileCustomCSS),
    unpinnedMenuItems: z.array(z.string()).default(THEME_DEFAULT_VALUES.unpinnedMenuItems),
    enableBlurringEffects: z.boolean().default(THEME_DEFAULT_VALUES.enableBlurringEffects),

}))

export const __ui_fixBorderRenderingArtifacts = atomWithStorage("sea-ui-settings-fix-border-rendering-artifacts", false)

const selectUISettingTabAtom = atom("main")

const tabsRootClass = cn("w-full contents space-y-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)

const tabsListClass = cn(
    "w-full flex flex-row lg:flex-row flex-wrap h-fit",
)

const tabContentClass = cn(
    "space-y-8 animate-in fade-in-0 duration-300",
)

// compact thumbnail for radio card option labels
function SelThumb({ children }: { children?: React.ReactNode }) {
    return (
        <div className="relative w-11 h-8 shrink-0 rounded overflow-hidden border border-white/[0.07] bg-gray-950">
            {children}
        </div>
    )
}

// banner behavior thumbnail, shows a mini abstract visual of what each banner setting does
function BannerBehaviorThumb({ type }: { type: string }) {
    const isConditional = type.endsWith("-when-unavailable")
    const baseType = isConditional ? type.replace("-when-unavailable", "") : type

    const fullBand = (
        <div className="absolute top-0 left-0 right-0 h-1/2">
            {baseType === "blur" && (
                <div className="absolute inset-0 bg-gradient-to-b from-gray-600/80 to-transparent [filter:blur(2px)] scale-105 origin-top" />
            )}
            {baseType === "dim" && (
                <div className="absolute inset-0 bg-gradient-to-b from-gray-600/25 to-transparent" />
            )}
            {baseType === "hide" && (
                <div className="absolute inset-0 bg-gray-950" />
            )}
            {baseType === "default" && (
                <div className="absolute inset-0 bg-gradient-to-b from-gray-600/80 to-transparent" />
            )}
        </div>
    )

    if (!isConditional) {
        return <SelThumb>{fullBand}</SelThumb>
    }

    // conditional: left half is always clear, right half shows the effect
    const rightHalf =
        baseType === "blur" ? <div className="flex-1 bg-gradient-to-b from-gray-600/80 to-transparent [filter:blur(2px)] scale-110 origin-top" />
            : baseType === "dim" ? <div className="flex-1 bg-gradient-to-b from-gray-600/25 to-transparent" />
                : <div className="flex-1 bg-gray-950" />

    return (
        <SelThumb>
            <div className="absolute top-0 left-0 right-0 h-1/2 flex overflow-hidden">
                <div className="flex-1 bg-gradient-to-b from-gray-600/80 to-transparent" />
                <div className="w-px bg-white/[0.07]" />
                {rightHalf}
            </div>
        </SelThumb>
    )
}

// shared classes for thumbnail-bearing radio cards
const thumbLabelClass = cn(
    "font-medium flex flex-row items-center data-[state=unchecked]:hover:text-[--foreground] data-[state=checked]:text-[--brand] text-[--muted] cursor-pointer")
const thumbContainerClass = "sea-selector-card"

const libraryBannerTypeOptions = [
    {
        value: "dynamic",
        label: (
            <span className="flex items-center gap-3">
                <SelThumb>
                    <div className="absolute inset-x-0 top-0 h-3/5 bg-gradient-to-b from-gray-600/80 to-transparent" />
                    <div className="absolute bottom-3 left-1 flex flex-col gap-0.5">
                        <div className="h-px w-7 rounded-full bg-white/30" />
                        <div className="h-px w-5 rounded-full bg-white/20" />
                    </div>
                </SelThumb>
                <span>Dynamic</span>
            </span>
        ),
    },
    {
        value: "custom",
        label: (
            <span className="flex items-center gap-3">
                <SelThumb>
                    <div className="absolute inset-1 rounded-sm border border-dashed border-white/20 flex items-center justify-center">
                        <svg className="w-3 h-3 text-white/25" viewBox="0 0 12 12" fill="none" stroke="currentColor" strokeWidth="1">
                            <rect x="0.5" y="0.5" width="11" height="11" rx="1.5" />
                            <path d="M1 9l3-3 2 2 3-4 2 5" strokeLinejoin="round" strokeLinecap="round" />
                        </svg>
                    </div>
                </SelThumb>
                <span>Custom</span>
            </span>
        ),
    },
    {
        value: "none",
        label: (
            <span className="flex items-center gap-3">
                <SelThumb>
                    <div className="absolute inset-0 flex items-center justify-center">
                        <div className="h-px w-7 bg-white/20 rotate-12" />
                    </div>
                </SelThumb>
                <span>None</span>
            </span>
        ),
    },
]

const bannerBehaviorOptions = ThemeMediaPageBannerTypeOptions.map(n => ({
    value: n.value,
    label: (
        <span className="flex items-center gap-3">
            <BannerBehaviorThumb type={n.value} />
            <span>{n.label}</span>
        </span>
    ),
}))

const bannerSizeOptions = ThemeMediaPageBannerSizeOptions.map(n => ({
    value: n.value,
    label: (
        <span className="flex items-center gap-3">
            <SelThumb>
                <div
                    className="absolute inset-x-0 top-0 bg-gradient-to-b from-gray-600/80 to-transparent"
                    style={{ height: n.value === "default" ? "65%" : "35%" }}
                />
            </SelThumb>
            <span>{n.label}</span>
        </span>
    ),
}))

function NavigationPreloadThumb({ mode }: { mode: NavigationPreloadMode }) {
    switch (mode) {
        case "disable":
            return (
                <SelThumb>
                    <div className="absolute inset-0 bg-gray-950" />
                    <div className="absolute left-1 top-1 h-1 w-4 rounded-full bg-gray-700/65" />
                    <div className="absolute left-1 top-3 h-1 w-6 rounded-full bg-gray-700/45" />
                    <div className="absolute right-1 top-1.5 h-4 w-3 rounded-sm border border-gray-400/40 bg-gray-500/25" />
                    <div className="absolute top-0 left-0 right-0 bottom-0 flex items-start justify-start pt-1.5 pl-0.5">
                        <div className="h-px w-11 bg-white/25 rotate-[20deg] origin-left" />
                    </div>
                </SelThumb>
            )
        case "faster":
            return (
                <SelThumb>
                    <div className="absolute inset-0 bg-gray-950" />
                    <div className="absolute left-1 top-1 h-1 w-4 rounded-full bg-gray-700/65" />
                    <div className="absolute left-1 top-3 h-1 w-6 rounded-full bg-gray-700/45" />
                    <div className="absolute right-1 top-1.5 h-4 w-3 rounded-sm border border-brand-400/40 bg-brand-500/25" />
                    <div className="absolute right-2 top-2.5 h-px w-4 bg-brand-300/80" />
                </SelThumb>
            )
        case "viewport":
            return (
                <SelThumb>
                    <div className="absolute inset-0 bg-gray-950" />
                    <div className="absolute inset-1 rounded-sm border border-white/[0.07]" />
                    <div className="absolute left-6 top-1.5 h-4 w-3 rounded-sm bg-brand-500/30 border border-brand-400/30" />
                    <div className="absolute left-2 top-1.5 h-4 w-3 rounded-sm bg-brand-500/20 border border-brand-400/20" />
                </SelThumb>
            )
        default:
            return (
                <SelThumb>
                    <div className="absolute inset-0 bg-gray-950" />
                    <div className="absolute left-1 top-1 h-1 w-4 rounded-full bg-gray-700/70" />
                    <div className="absolute left-1 top-3 h-1 w-6 rounded-full bg-gray-700/50" />
                    <div className="absolute right-1 top-1 h-4 w-3 rounded-sm border border-brand-400/35 bg-brand-500/25" />
                    <div className="absolute right-2.5 top-2 h-1.5 w-1.5 rounded-full bg-brand-300/80" />
                </SelThumb>
            )
    }
}

const navigationPreloadOptions: Array<{
    value: NavigationPreloadMode
    title: string
    description?: string
}> = [
    {
        value: "disable",
        title: "Disabled",
        description: "No preloading",
    },
    {
        value: "default",
        title: "Intent",
        description: "Preload on hover",
    },
    {
        value: "faster",
        title: "Faster Intent",
        description: "Preload more aggressively",
    },
    {
        value: "viewport",
        title: "Viewport",
        description: "When visible in the viewport",
    },
]


// smaller thumbnail for Field.Switch label prop
function SwThumb({ children }: { children?: React.ReactNode }) {
    return (
        <div className="relative w-9 h-6 shrink-0 rounded overflow-hidden border border-white/[0.06] bg-gray-950">
            {children}
        </div>
    )
}

// wraps a thumbnail + label text for Field.Switch
function swLabel(thumb: React.ReactNode, text: React.ReactNode) {
    return (
        <span className="flex items-center gap-2.5">
            {thumb}
            <span>{text}</span>
        </span>
    )
}


export function UISettings() {
    const themeSettings = useThemeSettings()
    const serverStatus = useServerStatus()

    const { mutate, isPending } = useUpdateTheme()
    // const [fixBorderRenderingArtifacts, setFixBorerRenderingArtifacts] = useAtom(__ui_fixBorderRenderingArtifacts)
    const [navigationPreloadMode, setNavigationPreloadMode] = useAtom(__navigationPreloadModeAtom)
    const [enableLivePreview, setEnableLivePreview] = useState(false)
    const isSimulatedUser = useIsSimulatedUser()

    const [tab, setTab] = useAtom(selectUISettingTabAtom)

    const formRef = React.useRef<UseFormReturn<any>>(null)

    const { customCSS, setCustomCSS } = useCustomCSS()

    const applyLivePreview = React.useCallback((bgColor: string, accentColor: string) => {
        if (!enableLivePreview) return

        let r = document.querySelector(":root") as any

        // Background color
        r.style.setProperty("--background", bgColor)
        r.style.setProperty("--paper", colord(bgColor).lighten(0.025).toHex())
        r.style.setProperty("--media-card-popup-background", colord(bgColor).lighten(0.025).toHex())
        r.style.setProperty(
            "--hover-from-background-color",
            colord(bgColor).lighten(0.025).desaturate(0.05).toHex(),
        )

        // Gray colors
        r.style.setProperty("--color-gray-400",
            `${colord(bgColor).lighten(0.3).desaturate(0.2).toRgb().r} ${colord(bgColor).lighten(0.3).desaturate(0.2).toRgb().g} ${colord(bgColor)
                .lighten(0.3)
                .desaturate(0.2)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-500",
            `${colord(bgColor).lighten(0.15).desaturate(0.2).toRgb().r} ${colord(bgColor).lighten(0.15).desaturate(0.2).toRgb().g} ${colord(bgColor)
                .lighten(0.15)
                .desaturate(0.2)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-600",
            `${colord(bgColor).lighten(0.1).desaturate(0.2).toRgb().r} ${colord(bgColor).lighten(0.1).desaturate(0.2).toRgb().g} ${colord(bgColor)
                .lighten(0.1)
                .desaturate(0.2)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-700",
            `${colord(bgColor).lighten(0.08).desaturate(0.2).toRgb().r} ${colord(bgColor).lighten(0.08).desaturate(0.2).toRgb().g} ${colord(bgColor)
                .lighten(0.08)
                .desaturate(0.2)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-800",
            `${colord(bgColor).lighten(0.06).desaturate(0.2).toRgb().r} ${colord(bgColor).lighten(0.06).desaturate(0.2).toRgb().g} ${colord(bgColor)
                .lighten(0.06)
                .desaturate(0.2)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-900",
            `${colord(bgColor).lighten(0.04).desaturate(0.05).toRgb().r} ${colord(bgColor).lighten(0.04).desaturate(0.05).toRgb().g} ${colord(bgColor)
                .lighten(0.04)
                .desaturate(0.05)
                .toRgb().b}`)
        r.style.setProperty("--color-gray-950",
            `${colord(bgColor).lighten(0.008).desaturate(0.05).toRgb().r} ${colord(bgColor).lighten(0.008).desaturate(0.05).toRgb().g} ${colord(
                bgColor).lighten(0.008).desaturate(0.05).toRgb().b}`)

        // Accent color
        r.style.setProperty("--color-brand-200",
            `${colord(accentColor).lighten(0.35).desaturate(0.05).toRgb().r} ${colord(accentColor).lighten(0.35).desaturate(0.05).toRgb().g} ${colord(
                accentColor).lighten(0.35).desaturate(0.05).toRgb().b}`)
        r.style.setProperty("--color-brand-300",
            `${colord(accentColor).lighten(0.3).desaturate(0.05).toRgb().r} ${colord(accentColor).lighten(0.3).desaturate(0.05).toRgb().g} ${colord(
                accentColor).lighten(0.3).desaturate(0.05).toRgb().b}`)
        r.style.setProperty("--color-brand-400",
            `${colord(accentColor).lighten(0.1).toRgb().r} ${colord(accentColor).lighten(0.1).toRgb().g} ${colord(accentColor)
                .lighten(0.1)
                .toRgb().b}`)
        r.style.setProperty("--color-brand-500", `${colord(accentColor).toRgb().r} ${colord(accentColor).toRgb().g} ${colord(accentColor).toRgb().b}`)
        r.style.setProperty("--color-brand-600",
            `${colord(accentColor).darken(0.1).toRgb().r} ${colord(accentColor).darken(0.1).toRgb().g} ${colord(accentColor).darken(0.1).toRgb().b}`)
        r.style.setProperty("--color-brand-700",
            `${colord(accentColor).darken(0.15).toRgb().r} ${colord(accentColor).darken(0.15).toRgb().g} ${colord(accentColor)
                .darken(0.15)
                .toRgb().b}`)
        r.style.setProperty("--color-brand-800",
            `${colord(accentColor).darken(0.2).toRgb().r} ${colord(accentColor).darken(0.2).toRgb().g} ${colord(accentColor).darken(0.2).toRgb().b}`)
        r.style.setProperty("--color-brand-900",
            `${colord(accentColor).darken(0.25).toRgb().r} ${colord(accentColor).darken(0.25).toRgb().g} ${colord(accentColor)
                .darken(0.25)
                .toRgb().b}`)
        r.style.setProperty("--color-brand-950",
            `${colord(accentColor).darken(0.3).toRgb().r} ${colord(accentColor).darken(0.3).toRgb().g} ${colord(accentColor).darken(0.3).toRgb().b}`)
        r.style.setProperty("--brand", colord(accentColor).lighten(0.35).desaturate(0.1).toHex())
    }, [enableLivePreview])

    function ObserveColorSettings() {

        const form = useFormContext()

        const accentColor = useWatch({ control: form.control, name: "accentColor" })
        const backgroundColor = useWatch({ control: form.control, name: "backgroundColor" })


        React.useEffect(() => {
            if (!enableLivePreview) return
            applyLivePreview(backgroundColor, accentColor)
        }, [enableLivePreview, backgroundColor, accentColor])

        return null
    }

    function handleSave(data: z.infer<typeof themeSchema>) {
        if (colord(data.backgroundColor).isLight()) {
            toast.error("Seanime does not support light themes")
            return
        }

        const prevEnableColorSettings = themeSettings?.enableColorSettings

        mutate({
            theme: {
                id: 0,
                ...themeSettings,
                ...data,
                libraryScreenCustomBackgroundBlur: data.libraryScreenCustomBackgroundBlur === "-"
                    ? ""
                    : data.libraryScreenCustomBackgroundBlur,
            },
        }, {
            onSuccess() {
                if (data.enableColorSettings !== prevEnableColorSettings && !data.enableColorSettings) {
                    window.location.reload()
                }
                formRef.current?.reset(formRef.current?.getValues())
            },
        })

        setCustomCSS({
            customCSS: data.customCSS,
            mobileCustomCSS: data.mobileCustomCSS,
        })
    }

    return (
        <Form
            schema={themeSchema}
            mRef={formRef}
            onSubmit={handleSave}
            defaultValues={{
                enableColorSettings: themeSettings?.enableColorSettings,
                animeEntryScreenLayout: themeSettings?.animeEntryScreenLayout,
                smallerEpisodeCarouselSize: themeSettings?.smallerEpisodeCarouselSize,
                expandSidebarOnHover: themeSettings?.expandSidebarOnHover,
                backgroundColor: themeSettings?.backgroundColor,
                accentColor: themeSettings?.accentColor,
                sidebarBackgroundColor: themeSettings?.sidebarBackgroundColor,
                hideTopNavbar: themeSettings?.hideTopNavbar,
                enableMediaCardBlurredBackground: themeSettings?.enableMediaCardBlurredBackground,
                libraryScreenBannerType: themeSettings?.libraryScreenBannerType,
                libraryScreenCustomBannerImage: themeSettings?.libraryScreenCustomBannerImage,
                libraryScreenCustomBannerPosition: themeSettings?.libraryScreenCustomBannerPosition,
                libraryScreenCustomBannerOpacity: themeSettings?.libraryScreenCustomBannerOpacity,
                libraryScreenCustomBackgroundImage: themeSettings?.libraryScreenCustomBackgroundImage,
                libraryScreenCustomBackgroundOpacity: themeSettings?.libraryScreenCustomBackgroundOpacity,
                disableLibraryScreenGenreSelector: themeSettings?.disableLibraryScreenGenreSelector,
                libraryScreenCustomBackgroundBlur: themeSettings?.libraryScreenCustomBackgroundBlur || "-",
                enableMediaPageBlurredBackground: themeSettings?.enableMediaPageBlurredBackground,
                disableSidebarTransparency: themeSettings?.disableSidebarTransparency,
                useLegacyEpisodeCard: themeSettings?.useLegacyEpisodeCard,
                disableCarouselAutoScroll: themeSettings?.disableCarouselAutoScroll,
                mediaPageBannerType: themeSettings?.mediaPageBannerType ?? ThemeMediaPageBannerType.Default,
                mediaPageBannerSize: themeSettings?.mediaPageBannerSize ?? ThemeMediaPageBannerType.Default,
                mediaPageBannerInfoBoxSize: themeSettings?.mediaPageBannerInfoBoxSize ?? ThemeMediaPageBannerType.Default,
                showEpisodeCardAnimeInfo: themeSettings?.showEpisodeCardAnimeInfo,
                continueWatchingDefaultSorting: themeSettings?.continueWatchingDefaultSorting,
                animeLibraryCollectionDefaultSorting: themeSettings?.animeLibraryCollectionDefaultSorting,
                mangaLibraryCollectionDefaultSorting: themeSettings?.mangaLibraryCollectionDefaultSorting,
                showAnimeUnwatchedCount: themeSettings?.showAnimeUnwatchedCount,
                showMangaUnreadCount: themeSettings?.showMangaUnreadCount,
                hideEpisodeCardDescription: themeSettings?.hideEpisodeCardDescription,
                hideDownloadedEpisodeCardFilename: themeSettings?.hideDownloadedEpisodeCardFilename,
                customCSS: themeSettings?.customCSS,
                mobileCustomCSS: themeSettings?.mobileCustomCSS,
                unpinnedMenuItems: themeSettings?.unpinnedMenuItems ?? [],
                enableBlurringEffects: themeSettings?.enableBlurringEffects,
            }}
            stackClass="space-y-4 relative"
        >
            {(f) => (
                <>
                    <SettingsIsDirty className="" />
                    <ObserveColorSettings />

                    <Tabs
                        value={tab}
                        onValueChange={setTab}
                        className={tabsRootClass}
                        triggerClass={tabsTriggerClass}
                        listClass={tabsListClass}
                    >
                        <TabsList data-settings-ui-panel-tabs className="flex-wrap max-w-full bg-[--paper] p-2 border rounded-xl">
                            <TabsTrigger value="main">General</TabsTrigger>
                            <TabsTrigger value="css">CSS</TabsTrigger>
                        </TabsList>

                        <TabsContent value="css" className={cn(tabContentClass)} data-settings-ui-panel-css>

                            <SettingsCard>

                                {serverStatus?.themeSettings?.customCSS !== customCSS.customCSS || serverStatus?.themeSettings?.mobileCustomCSS !== customCSS.mobileCustomCSS && (
                                    <Button
                                        intent="white"
                                        disabled={serverStatus?.themeSettings?.customCSS === customCSS.customCSS && serverStatus?.themeSettings?.mobileCustomCSS === customCSS.mobileCustomCSS}
                                        onClick={() => {
                                            setCustomCSS({
                                                customCSS: serverStatus?.themeSettings?.customCSS || "",
                                                mobileCustomCSS: serverStatus?.themeSettings?.mobileCustomCSS || "",
                                            })
                                        }}
                                    >
                                        Apply to this client
                                    </Button>
                                )}

                                <p className="text-[--muted] text-sm">
                                    The custom CSS will be saved on the server and needs to be applied manually to each client.
                                    <br />
                                    In case of an error rendering the UI unusable, you can always remove it from the local storage using the
                                    devtools.
                                </p>

                                <div className="flex flex-col md:flex-row gap-3">

                                    <Field.Textarea
                                        label="Custom CSS"
                                        name="customCSS"
                                        placeholder="Custom CSS"
                                        help="Applied above 1024px screen size."
                                        className="min-h-[500px]"
                                    />

                                    <Field.Textarea
                                        label="Mobile custom CSS"
                                        name="mobileCustomCSS"
                                        placeholder="Custom CSS"
                                        help="Applied below 1024px screen size."
                                        className="min-h-[500px]"
                                    />

                                </div>

                            </SettingsCard>


                        </TabsContent>

                        <TabsContent value="main" className={tabContentClass} data-settings-ui-panel-general>

                            <SettingsCard title="Sorting">

                                {!serverStatus?.settings?.library?.enableWatchContinuity && (
                                    f.watch("continueWatchingDefaultSorting").includes("LAST_WATCHED") ||
                                    f.watch("animeLibraryCollectionDefaultSorting").includes("LAST_WATCHED")
                                ) && (
                                    <Alert
                                        intent="alert"
                                        description="Watch continuity needs to be enabled to use the last watched sorting options."
                                    />
                                )}


                                <Field.Select
                                    label="Continue watching sorting"
                                    name="continueWatchingDefaultSorting"
                                    options={CONTINUE_WATCHING_SORTING_OPTIONS.map(n => ({ value: n.value, label: n.label }))}
                                />

                                <Field.Select
                                    label="Anime library sorting"
                                    name="animeLibraryCollectionDefaultSorting"
                                    options={ANIME_COLLECTION_SORTING_OPTIONS.filter(n => !n.value.includes("END"))
                                        .map(n => ({ value: n.value, label: n.label }))}
                                />

                                <Field.Select
                                    label="Manga library sorting"
                                    name="mangaLibraryCollectionDefaultSorting"
                                    options={MANGA_COLLECTION_SORTING_OPTIONS.filter(n => !n.value.includes("END"))
                                        .map(n => ({ value: n.value, label: n.label }))}
                                />


                            </SettingsCard>

                            <SettingsCard title="Theme">
                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div
                                                className="absolute inset-0"
                                                style={{ background: "conic-gradient(from 0deg, #ef4444, #f97316, #eab308, #22c55e, #3b82f6, #a855f7, #ef4444)" }}
                                            />
                                            <div className="absolute inset-0 bg-black/40" />
                                            <div className="absolute inset-1 rounded-full border border-white/20 bg-gray-950/80" />
                                        </SwThumb>,
                                        "Enable color settings",
                                    )}
                                    name="enableColorSettings"
                                />
                                {f.watch("enableColorSettings") && (
                                    <>
                                        <Switch
                                            side="right"
                                            label="Live preview"
                                            name="enableLivePreview"
                                            help={enableLivePreview && "Disabling will reload the page without applying the changes."}
                                            value={enableLivePreview}
                                            onValueChange={(value) => {
                                                setEnableLivePreview(value)
                                                if (!value) {
                                                    // Reset to saved values if disabling preview
                                                    window.location.reload()
                                                } else {
                                                    // Apply current form values as preview
                                                    applyLivePreview(f.watch("backgroundColor"), f.watch("accentColor"))
                                                }
                                            }}
                                        />
                                        <div className="flex flex-col md:flex-row gap-3">
                                            <Field.ColorPicker
                                                name="backgroundColor"
                                                label="Background color"
                                                help="Default: #070707"
                                            />
                                            <Field.ColorPicker
                                                name="accentColor"
                                                label="Accent color"
                                                help="Default: #6152df"
                                            />
                                        </div>
                                    </>
                                )}

                                {f.watch("enableColorSettings") && (
                                    <div className="flex flex-wrap gap-3 w-full">
                                        {THEME_COLOR_BANK.map((opt) => (
                                            <div
                                                key={opt.name}
                                                className={cn(
                                                    "flex gap-3 items-center w-fit rounded-full border p-1 cursor-pointer",
                                                    themeSettings.backgroundColor === opt.backgroundColor && themeSettings.accentColor === opt.accentColor && "border-[--brand] ring-[--brand] ring-offset-1 ring-offset-[--background]",
                                                )}
                                                onClick={() => {
                                                    f.setValue("backgroundColor", opt.backgroundColor)
                                                    f.setValue("accentColor", opt.accentColor)

                                                    if (enableLivePreview) {
                                                        applyLivePreview(opt.backgroundColor, opt.accentColor)
                                                    } else {
                                                        mutate({
                                                            theme: {
                                                                id: 0,
                                                                ...themeSettings,
                                                                enableColorSettings: true,
                                                                backgroundColor: opt.backgroundColor,
                                                                accentColor: opt.accentColor,
                                                            },
                                                        }, {
                                                            onSuccess() {
                                                                formRef.current?.reset(formRef.current?.getValues())
                                                            },
                                                        })
                                                    }
                                                }}
                                            >
                                                <div
                                                    className="flex gap-1"
                                                >
                                                    <div
                                                        className="w-6 h-6 rounded-full border"
                                                        style={{ backgroundColor: opt.backgroundColor }}
                                                    />
                                                    <div
                                                        className="w-6 h-6 rounded-full border"
                                                        style={{ backgroundColor: opt.accentColor }}
                                                    />
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}


                            </SettingsCard>

                            <SettingsCard title="Banners & Background">
                                <div className="flex flex-col md:flex-row gap-3">
                                    <Field.Text
                                        label="Background image path"
                                        name="libraryScreenCustomBackgroundImage"
                                        placeholder="e.g. image.png"
                                        help="Background image for all pages. Dimmed on non-library screens."
                                    />

                                    <Field.Number
                                        label="Background image opacity"
                                        name="libraryScreenCustomBackgroundOpacity"
                                        placeholder="Default: 10"
                                        min={1}
                                        max={100}
                                    />

                                    {/*<Field.Select*/}
                                    {/*    label="Blur"*/}
                                    {/*    name="libraryScreenCustomBackgroundBlur"*/}
                                    {/*    help="Can cause performance issues."*/}
                                    {/*    options={[*/}
                                    {/*        { label: "None", value: "-" },*/}
                                    {/*        { label: "5px", value: "5px" },*/}
                                    {/*        { label: "10px", value: "10px" },*/}
                                    {/*        { label: "15px", value: "15px" },*/}
                                    {/*    ]}*/}
                                    {/*/>*/}
                                </div>
                                <div className="flex flex-col md:flex-row gap-3">
                                    <Field.Text
                                        label="Banner image path"
                                        name="libraryScreenCustomBannerImage"
                                        placeholder="e.g. image.gif"
                                        help="Banner image for all pages."
                                    />
                                    <Field.Text
                                        label="Banner position"
                                        name="libraryScreenCustomBannerPosition"
                                        placeholder="Default: 50% 50%"
                                    />
                                    <Field.Number
                                        label="Banner opacity"
                                        name="libraryScreenCustomBannerOpacity"
                                        placeholder="Default: 10"
                                        min={1}
                                        max={100}
                                    />
                                </div>

                                <Field.RadioCards
                                    label="Home screen banner type"
                                    name="libraryScreenBannerType"
                                    options={libraryBannerTypeOptions}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    itemLabelClass={thumbLabelClass}
                                    itemContainerClass={thumbContainerClass}
                                    help={f.watch("libraryScreenBannerType") === ThemeLibraryScreenBannerType.Custom && "Use the banner image on all library screens."}
                                />

                                <Field.RadioCards
                                    label="Media screen banner image"
                                    name="mediaPageBannerType"
                                    options={bannerBehaviorOptions}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    itemLabelClass={thumbLabelClass}
                                    itemContainerClass={thumbContainerClass}
                                    radioGroupStackClass="flex-wrap"
                                    help={ThemeMediaPageBannerTypeOptions.find(n => n.value === f.watch("mediaPageBannerType"))?.description}
                                />

                                <Field.RadioCards
                                    label="Media screen banner size"
                                    name="mediaPageBannerSize"
                                    options={bannerSizeOptions}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    itemLabelClass={thumbLabelClass}
                                    itemContainerClass={thumbContainerClass}
                                    radioGroupStackClass="flex-wrap"
                                    help={ThemeMediaPageBannerSizeOptions.find(n => n.value === f.watch("mediaPageBannerSize"))?.description}
                                />
                            </SettingsCard>


                            <SettingsCard title="Tweaks">

                                <RadioGroup
                                    label="Navigation preloading"
                                    value={isSimulatedUser ? "disable" : navigationPreloadMode}
                                    onValueChange={(value) => setNavigationPreloadMode(value as NavigationPreloadMode)}
                                    options={navigationPreloadOptions.map(option => ({
                                        value: option.value,
                                        label: (
                                            <div className="flex items-center gap-3 text-left flex-none">
                                                <NavigationPreloadThumb mode={option.value} />
                                                <span className="flex flex-col gap-0.5">
                                                    <span>{option.title}</span>
                                                    <span className="text-xs leading-4 text-[--muted] data-[state=checked]:text-[--muted]">
                                                        {option.description}
                                                    </span>
                                                </span>
                                            </div>
                                        ),
                                    }))}
                                    fieldClass={cn(
                                        "settings-ui-navigation-preloading",
                                        isSimulatedUser && "pointer-events-none opacity-50",
                                    )}
                                    itemContainerClass={cn(
                                        "cursor-pointer transition border-transparent rounded-[--radius] p-3 w-full md:w-fit",
                                        "bg-transparent dark:hover:bg-gray-900 dark:bg-transparent",
                                        "data-[state=checked]:bg-brand-500/5 dark:data-[state=checked]:bg-gray-900",
                                        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-transparent transition",
                                        "dark:border dark:data-[state=checked]:border-[--border] data-[state=checked]:ring-offset-0",
                                        "items-center",
                                    )}
                                    itemClass={cn(
                                        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                                        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                                        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
                                    )}
                                    itemIndicatorClass="hidden"
                                    itemLabelClass="font-medium justify-center flex flex-col items-center data-[state=unchecked]:hover:text-[--foreground] data-[state=checked]:text-[--brand] text-[--muted] cursor-pointer"
                                    // stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    stackClass={cn("flex flex-col md:flex-row gap-2 space-y-0 flex-wrap")}
                                    help="Applies to media pages on this client. Preloading can cause you to hit rate limits faster."
                                />
                                {isSimulatedUser && (
                                    <p className="text-orange-300/50 text-sm">
                                        Navigation preloading is disabled for use without an AniList account due to rate limits.
                                    </p>
                                )}

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute h-full items-center left-1 flex gap-0.5">
                                                <div className="h-2 w-5 rounded-full bg-gray-700/60 border border-white/[0.07]" />
                                                <div className="h-2 w-4 rounded-full bg-gray-700/60 border border-white/[0.07]" />
                                            </div>
                                            <div className="absolute top-0 left-0 right-0 bottom-0 flex items-start justify-start pt-1.5 pl-0.5">
                                                <div className="h-px w-9 bg-white/25 rotate-[20deg] origin-left" />
                                            </div>
                                        </SwThumb>,
                                        "Remove genre selector",
                                    )}
                                    name="disableLibraryScreenGenreSelector"
                                />


                                {/*<Field.RadioCards*/}
                                {/*    label="Banner info layout"*/}
                                {/*    name="mediaPageBannerInfoBoxSize"*/}
                                {/*    options={ThemeMediaPageInfoBoxSizeOptions.map(n => ({ value: n.value, label: n.label }))}*/}
                                {/*    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"*/}
                                {/*/>*/}

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-600/60 to-gray-900/60" />
                                            <div className="absolute inset-1 rounded-sm bg-white/[0.08] border border-white/10" />
                                        </SwThumb>,
                                        "Enable blurring effects",
                                    )}
                                    help="May impact performance on some devices."
                                    name="enableBlurringEffects"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute inset-x-0 top-0 h-[90%] bg-gradient-to-b from-gray-600/60 to-transparent [filter:blur(3px)] scale-105 origin-top" />
                                            <div className="absolute inset-x-0 top-0 h-1/2 bg-gradient-to-b from-gray-600/20 to-transparent" />
                                        </SwThumb>,
                                        "Media screen blurred background",
                                    )}
                                    name="enableMediaPageBlurredBackground"
                                    help="Can cause performance issues."
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-800/80 to-gray-950" />
                                            <div className="absolute inset-1 rounded-sm bg-gray-900/60 border border-white/[0.06]" />
                                            <div className="absolute top-0.5 right-0.5 w-3 h-3 rounded-full bg-gray-600 flex items-center justify-center">
                                                <span className="text-[5px] text-white leading-none font-bold">3</span>
                                            </div>
                                        </SwThumb>,
                                        "Anime card unwatched count",
                                    )}
                                    name="showAnimeUnwatchedCount"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-800/80 to-gray-950" />
                                            <div className="absolute inset-1 rounded-sm bg-gray-900/60 border border-white/[0.06]" />
                                            <div className="absolute top-0.5 right-0.5 w-3 h-3 rounded-full bg-gray-700 flex items-center justify-center">
                                                <span className="text-[5px] text-white leading-none font-bold">5</span>
                                            </div>
                                        </SwThumb>,
                                        "Manga card unread count",
                                    )}
                                    name="showMangaUnreadCount"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-600/40 to-gray-950" />
                                            <div className="absolute inset-1 rounded-sm bg-white/[0.07] border border-white/10" />
                                        </SwThumb>,
                                        "Media card glassy background",
                                    )}
                                    name="enableMediaCardBlurredBackground"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute inset-0.5 rounded-sm border border-white/10 flex flex-col overflow-hidden">
                                                {/* <div className="w-5 bg-gray-800/60 shrink-0" /> */}
                                                <div className="flex-1 p-0.5 flex flex-col gap-0.5">
                                                    <div className="h-full"></div>
                                                    <div className="h-px w-full bg-white/20 rounded" />
                                                    {/* <div className="h-px w-2/3 bg-white/12 rounded" /> */}
                                                </div>
                                            </div>
                                        </SwThumb>,
                                        "Episode cards: Legacy layout",
                                    )}
                                    name="useLegacyEpisodeCard"
                                />

                                {/*<Field.Switch*/}
                                {/*    side="right"*/}
                                {/*    label="Show anime info"*/}
                                {/*    name="showEpisodeCardAnimeInfo"*/}
                                {/*/>*/}

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute inset-0.5 rounded-sm border border-white/10 flex overflow-hidden">
                                                <div className="flex-1 p-0.5 flex flex-col gap-0.5">
                                                    <div className="h-px w-full bg-white/20 rounded" />
                                                    <div className="h-px w-full bg-white/20 rounded" />
                                                    <div className="h-px w-full bg-white/20 rounded" />
                                                    <div className="h-px w-3/4 bg-white/[0.05] rounded" />
                                                    <div className="h-px w-1/2 bg-white/[0.05] rounded" />
                                                </div>

                                            </div>
                                            <div className="absolute top-0 left-0 right-0 bottom-0 flex items-start justify-start pt-1.5 pl-0.5">
                                                <div className="h-px w-9 bg-white/25 rotate-[20deg] origin-left" />
                                            </div>
                                        </SwThumb>,
                                        "Episode items: Hide summary",
                                    )}
                                    name="hideEpisodeCardDescription"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute inset-0.5 rounded-sm border border-white/10 flex overflow-hidden">
                                                <div className="flex-1 p-0.5 flex flex-col justify-between">
                                                    <div className="h-px w-full bg-white/20 rounded" />
                                                    <div className="h-px w-full bg-white/[0.04] rounded" />
                                                </div>
                                            </div>
                                            <div className="absolute top-0 left-0 right-0 bottom-0 flex items-start justify-start pt-1.5 pl-0.5">
                                                <div className="h-px w-9 bg-white/25 rotate-[20deg] origin-left" />
                                            </div>
                                        </SwThumb>,
                                        "Episode items: Hide filename",
                                    )}
                                    name="hideDownloadedEpisodeCardFilename"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute top-1 left-0.5 right-0.5 bottom-1 flex gap-0.5 overflow-hidden">
                                                <div className="w-5 bg-gray-800/70 rounded-sm shrink-0" />
                                                <div className="w-5 bg-gray-800/50 rounded-sm shrink-0" />
                                                <div className="w-5 bg-gray-800/30 rounded-sm shrink-0" />
                                                <LuChevronRight className="absolute -right-1 top-0 text-gray-500" />
                                            </div>
                                            <div className="absolute top-0 left-0 right-0 bottom-0 flex items-start justify-start pt-1.5 pl-0.5">
                                                <div className="h-px w-9 bg-white/25 rotate-[20deg] origin-left" />
                                            </div>
                                        </SwThumb>,
                                        "Disable carousel auto-scroll",
                                    )}
                                    name="disableCarouselAutoScroll"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute top-1.5 left-0.5 right-0.5 bottom-1.5 flex gap-0.5 overflow-hidden">
                                                <div className="w-4 bg-gray-800/70 rounded-sm shrink-0" />
                                                <div className="w-4 bg-gray-800/60 rounded-sm shrink-0" />
                                                <div className="w-4 bg-gray-800/50 rounded-sm shrink-0" />
                                                <div className="w-4 bg-gray-800/40 rounded-sm shrink-0" />
                                            </div>
                                        </SwThumb>,
                                        "Smaller carousel episode cards",
                                    )}
                                    name="smallerEpisodeCarouselSize"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-600/30 to-gray-950" />
                                            <div className="absolute left-0 top-0 bottom-0 w-6 bg-gray-950 border-r" />
                                            <div className="absolute left-1 top-1.5 flex flex-col gap-0.5">
                                                <div className="h-0.5 w-1.5 bg-white/30 rounded" />
                                                <div className="h-0.5 w-1.5 bg-white/20 rounded" />
                                                <div className="h-0.5 w-1.5 bg-white/20 rounded" />
                                            </div>
                                        </SwThumb>,
                                        "Expand sidebar on hover",
                                    )}
                                    name="expandSidebarOnHover"
                                    help="Causes visual glitches with plugin tray."
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gradient-to-br from-gray-600/30 to-gray-950" />
                                            <div className="absolute left-0 top-0 bottom-0 w-3.5 bg-gray-900" />
                                            <div className="absolute left-1 top-1.5 flex flex-col gap-0.5">
                                                <div className="h-0.5 w-1.5 bg-white/30 rounded" />
                                                <div className="h-0.5 w-1.5 bg-white/20 rounded" />
                                                <div className="h-0.5 w-1.5 bg-white/20 rounded" />
                                            </div>
                                        </SwThumb>,
                                        "Disable sidebar transparency",
                                    )}
                                    name="disableSidebarTransparency"
                                />

                                <Field.Switch
                                    side="right"
                                    label={swLabel(
                                        <SwThumb>
                                            <div className="absolute inset-0 bg-gray-950" />
                                            <div className="absolute left-0 top-0 bottom-0 w-2.5 bg-gray-800" />
                                            <div className="absolute left-3 top-0.5 right-0.5 bottom-0.5 bg-gray-900/50 rounded-sm" />
                                            <div className="absolute left-4 top-1.5 flex flex-col gap-0.5">
                                                <div className="h-px w-10 bg-white/15 rounded" />
                                                <div className="h-px w-7 bg-white/10 rounded" />
                                            </div>
                                        </SwThumb>,
                                        __isDesktop__ ? "Hide top navbar (web interface)" : "Hide top navbar",
                                    )}
                                    name="hideTopNavbar"
                                    help="Switches to sidebar-only mode."
                                />

                                <Field.Combobox
                                    label="Unpinned menu items"
                                    name="unpinnedMenuItems"
                                    emptyMessage="No items selected"
                                    multiple
                                    options={[
                                        {
                                            label: "Schedule",
                                            textValue: "Schedule",
                                            value: "schedule",
                                        },
                                        {
                                            label: "Manga",
                                            textValue: "Manga",
                                            value: "manga",
                                        },
                                        {
                                            label: "Discover",
                                            textValue: "Discover",
                                            value: "discover",
                                        },
                                        {
                                            label: "My lists",
                                            textValue: "My lists",
                                            value: "lists",
                                        },
                                        {
                                            label: "Auto Downloader",
                                            textValue: "Auto Downloader",
                                            value: "auto-downloader",
                                        },
                                        {
                                            label: "Torrent list",
                                            textValue: "Torrent list",
                                            value: "torrent-list",
                                        },
                                        {
                                            label: "Debrid",
                                            textValue: "Debrid",
                                            value: "debrid",
                                        },
                                        {
                                            label: "Scan summaries",
                                            textValue: "Scan summaries",
                                            value: "scan-summaries",
                                        },
                                        {
                                            label: "Search",
                                            textValue: "Search",
                                            value: "search",
                                        },
                                    ]}
                                />

                            </SettingsCard>

                        </TabsContent>

                        {/*<TabsContent value="browser-client" className={tabContentClass}>*/}

                        {/*    <SettingsCard>*/}
                        {/*        <Switch*/}
                        {/*            side="right"*/}
                        {/*            label="Fix border rendering artifacts (client-specific)"*/}
                        {/*            name="enableMediaCardStyleFix"*/}
                        {/*            help="Seanime will try to fix border rendering artifacts. This setting only affects this client/browser."*/}
                        {/*            value={fixBorderRenderingArtifacts}*/}
                        {/*            onValueChange={(v) => {*/}
                        {/*                setFixBorerRenderingArtifacts(v)*/}
                        {/*                if (v) {*/}
                        {/*                    toast.success("Handling border rendering artifacts")*/}
                        {/*                } else {*/}
                        {/*                    toast.success("Border rendering artifacts are no longer handled")*/}
                        {/*                }*/}
                        {/*            }}*/}
                        {/*        />*/}
                        {/*    </SettingsCard>*/}

                        {/*</TabsContent>*/}

                        {tab !== "browser-client" && <div className="mt-4">
                            <Field.Submit role="save" intent="white" rounded loading={isPending}>Save</Field.Submit>
                        </div>}

                    </Tabs>
                </>
            )}
        </Form>
    )

}
