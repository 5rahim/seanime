"use client"
import { useUpdateTheme } from "@/api/hooks/theme.hooks"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
    THEME_DEFAULT_VALUES,
    ThemeLibraryScreenBannerType,
    ThemeMediaPageBannerSizeOptions,
    ThemeMediaPageBannerType,
    ThemeMediaPageBannerTypeOptions,
    ThemeMediaPageInfoBoxSizeOptions,
    useThemeSettings,
} from "@/lib/theme/hooks"
import { THEME_COLOR_BANK } from "@/lib/theme/theme-bank"
import { colord } from "colord"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { toast } from "sonner"
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
}))

export const __ui_fixBorderRenderingArtifacts = atomWithStorage("sea-ui-settings-fix-border-rendering-artifacts", false)

const selectUISettingTabAtom = atom("main")

const tabsRootClass = cn("w-full contents space-y-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-md w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)

const tabsListClass = cn(
    "w-full flex flex-row lg:flex-row flex-wrap h-fit",
)

export function UISettings() {
    const themeSettings = useThemeSettings()

    const { mutate, isPending } = useUpdateTheme()
    const [fixBorderRenderingArtifacts, setFixBorerRenderingArtifacts] = useAtom(__ui_fixBorderRenderingArtifacts)

    const [tab, setTab] = useAtom(selectUISettingTabAtom)

    const formRef = React.useRef<UseFormReturn<any>>(null)

    return (
        <Form
            schema={themeSchema}
            mRef={formRef}
            onSubmit={data => {
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
                    },
                })
            }}
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
            }}
            stackClass="space-y-4 relative"
        >
            {(f) => (
                <>
                    <SettingsIsDirty className="-top-14" />

                    <Tabs
                        value={tab}
                        onValueChange={setTab}
                        className={tabsRootClass}
                        triggerClass={tabsTriggerClass}
                        listClass={tabsListClass}
                    >
                        <TabsList className="flex-wrap max-w-full">
                            <TabsTrigger value="main">Theme</TabsTrigger>
                            <TabsTrigger value="media">Media</TabsTrigger>
                            <TabsTrigger value="navigation">Navigation</TabsTrigger>
                            <TabsTrigger value="browser-client">Rendering</TabsTrigger>
                        </TabsList>

                        <TabsContent value="main" className="space-y-4">

                            <SettingsCard title="Color scheme">
                                <Field.Switch
                                    side="right"
                                    label="Enable color settings"
                                    name="enableColorSettings"
                                />
                                <div className="flex flex-col md:flex-row gap-3">
                                    <Field.ColorPicker
                                        name="backgroundColor"
                                        label="Background color"
                                        help="Default: #070707"
                                        disabled={!f.watch("enableColorSettings")}
                                    />
                                    <Field.ColorPicker
                                        name="accentColor"
                                        label="Accent color"
                                        help="Default: #6152df"
                                        disabled={!f.watch("enableColorSettings")}
                                    />
                                </div>

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

                            <SettingsCard title="Background image">

                                <div className="flex flex-col md:flex-row gap-3">
                                    <Field.Text
                                        label="Image path"
                                        name="libraryScreenCustomBackgroundImage"
                                        placeholder="e.g., image.png"
                                        help="Background image for all pages. Dimmed on non-library screens."
                                    />

                                    <Field.Number
                                        label="Opacity"
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

                            </SettingsCard>

                            <SettingsCard title="Banner image">

                                <div className="flex flex-col md:flex-row gap-3">
                                    <Field.Text
                                        label="Image path"
                                        name="libraryScreenCustomBannerImage"
                                        placeholder="e.g., image.gif"
                                        help="Banner image for all pages."
                                    />
                                    <Field.Text
                                        label="Position"
                                        name="libraryScreenCustomBannerPosition"
                                        placeholder="Default: 50% 50%"
                                    />
                                    <Field.Number
                                        label="Opacity"
                                        name="libraryScreenCustomBannerOpacity"
                                        placeholder="Default: 10"
                                        min={1}
                                        max={100}
                                    />
                                </div>

                            </SettingsCard>

                        </TabsContent>

                        <TabsContent value="navigation" className="space-y-4">

                            <SettingsCard title="Sidebar">

                                <Field.Switch
                                    side="right"
                                    label="Expand sidebar on hover"
                                    name="expandSidebarOnHover"
                                />

                                <Field.Switch
                                    side="right"
                                    label="Disable transparency"
                                    name="disableSidebarTransparency"
                                />

                            </SettingsCard>

                            <SettingsCard title="Navbar">

                                <Field.Switch
                                    side="right"
                                    label={process.env.NEXT_PUBLIC_PLATFORM === "desktop" ? "Hide top navbar (Web interface)" : "Hide top navbar"}
                                    name="hideTopNavbar"
                                    help="Switches to sidebar-only mode."
                                />

                            </SettingsCard>

                        </TabsContent>

                        <TabsContent value="media" className="space-y-4">

                            <SettingsCard title="Collection screens">

                                <Field.RadioCards
                                    label="Banner type"
                                    name="libraryScreenBannerType"
                                    options={[
                                        {
                                            label: "Dynamic Banner",
                                            value: "dynamic",
                                        },
                                        {
                                            label: "Custom Banner",
                                            value: "custom",
                                        },
                                        {
                                            label: "None",
                                            value: "none",
                                        },
                                    ]}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    help={f.watch("libraryScreenBannerType") === ThemeLibraryScreenBannerType.Custom && "Use the banner image on all library screens."}
                                />

                                <Field.Switch
                                    side="right"
                                    label="No genre selector"
                                    name="disableLibraryScreenGenreSelector"
                                />

                            </SettingsCard>

                            <SettingsCard title="Media page">

                                <Field.RadioCards
                                    label="AniList banner image"
                                    name="mediaPageBannerType"
                                    options={ThemeMediaPageBannerTypeOptions.map(n => ({ value: n.value, label: n.label }))}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    help={ThemeMediaPageBannerTypeOptions.find(n => n.value === f.watch("mediaPageBannerType"))?.description}
                                />

                                <Field.RadioCards
                                    label="Banner size"
                                    name="mediaPageBannerSize"
                                    options={ThemeMediaPageBannerSizeOptions.map(n => ({ value: n.value, label: n.label }))}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                    help={ThemeMediaPageBannerSizeOptions.find(n => n.value === f.watch("mediaPageBannerSize"))?.description}
                                />

                                <Field.RadioCards
                                    label="Banner info layout"
                                    name="mediaPageBannerInfoBoxSize"
                                    options={ThemeMediaPageInfoBoxSizeOptions.map(n => ({ value: n.value, label: n.label }))}
                                    stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
                                // help={ThemeMediaPageInfoBoxSizeOptions.find(n => n.value === f.watch("mediaPageBannerInfoBoxSize"))?.description}
                                />

                                <Field.Switch
                                    side="right"
                                    label="Blurred gradient background"
                                    name="enableMediaPageBlurredBackground"
                                    help="Can cause performance issues."
                                />

                            </SettingsCard>

                            <SettingsCard title="Media card">

                                <Field.Switch
                                    side="right"
                                    label="Glassy background"
                                    name="enableMediaCardBlurredBackground"
                                />

                            </SettingsCard>

                            <SettingsCard title="Episode card">

                                <Field.Switch
                                    side="right"
                                    label="Legacy episode cards"
                                    name="useLegacyEpisodeCard"
                                />


                            </SettingsCard>

                            <SettingsCard title="Carousel">

                                <Field.Switch
                                    side="right"
                                    label="Disable auto-scroll"
                                    name="disableCarouselAutoScroll"
                                />

                                <Field.Switch
                                    side="right"
                                    label="Smaller episode cards"
                                    name="smallerEpisodeCarouselSize"
                                />

                            </SettingsCard>

                        </TabsContent>

                        <TabsContent value="browser-client" className="space-y-4">

                            <SettingsCard>
                                <Switch
                                    side="right"
                                    label="Fix border rendering artifacts (client-specific)"
                                    name="enableMediaCardStyleFix"
                                    help="Seanime will try to fix border rendering artifacts. This setting only affects this client/browser."
                                    value={fixBorderRenderingArtifacts}
                                    onValueChange={(v) => {
                                        setFixBorerRenderingArtifacts(v)
                                        if (v) {
                                            toast.success("Handling border rendering artifacts")
                                        } else {
                                            toast.success("Border rendering artifacts are no longer handled")
                                        }
                                    }}
                                />
                            </SettingsCard>

                        </TabsContent>

                        {tab !== "browser-client" && <div className="mt-4">
                            <Field.Submit role="save" intent="white" rounded loading={isPending}>Save</Field.Submit>
                        </div>}

                    </Tabs>
                </>
            )}
        </Form>
    )

}
