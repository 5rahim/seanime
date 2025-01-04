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
import { toast } from "sonner"

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

    return (
        <Form
            schema={themeSchema}
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
            stackClass="space-y-4"
        >
            {(f) => (
                <>

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

                            <h3>Color scheme</h3>

                            <Field.Switch
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

                            <br />

                            <h3>
                                Background image
                            </h3>

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

                            <br />

                            <h3>Banner image</h3>

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

                        </TabsContent>

                        <TabsContent value="navigation" className="space-y-4">

                            <h3>Sidebar</h3>

                            <Field.Switch
                                label="Expand sidebar on hover"
                                name="expandSidebarOnHover"
                            />

                            <Field.Switch
                                label="Disable transparency"
                                name="disableSidebarTransparency"
                            />

                            <br />

                            <h3>Navbar</h3>

                            <Field.Switch
                                label={process.env.NEXT_PUBLIC_PLATFORM === "desktop" ? "Hide top navbar (Web interface)" : "Hide top navbar"}
                                name="hideTopNavbar"
                                help="Switches to sidebar-only mode."
                            />

                        </TabsContent>

                        <TabsContent value="media" className="space-y-4">

                            <div>
                                <h3>My Library and Manga screens</h3>
                            </div>

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
                                itemContainerClass={cn(
                                    "items-start w-fit cursor-pointer transition border-transparent rounded-[--radius] p-4",
                                    "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                                    "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                    "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                    "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                                    "py-2",
                                )}
                                help={f.watch("libraryScreenBannerType") === ThemeLibraryScreenBannerType.Custom && "Use the banner image on all library screens."}
                            />

                            <Field.Switch
                                label="No genre selector"
                                name="disableLibraryScreenGenreSelector"
                            />

                            <br />

                            <div>
                                <h3>Media page</h3>
                            </div>

                            <Field.RadioCards
                                label="AniList banner image"
                                name="mediaPageBannerType"
                                options={ThemeMediaPageBannerTypeOptions.map(n => ({ value: n.value, label: n.label }))}
                                itemContainerClass={cn(
                                    "items-start w-fit cursor-pointer transition border-transparent rounded-[--radius] p-4",
                                    "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                                    "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                    "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                    "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                                    "py-2",
                                )}
                                help={ThemeMediaPageBannerTypeOptions.find(n => n.value === f.watch("mediaPageBannerType"))?.description}
                            />

                            <Field.RadioCards
                                label="Banner size"
                                name="mediaPageBannerSize"
                                options={ThemeMediaPageBannerSizeOptions.map(n => ({ value: n.value, label: n.label }))}
                                itemContainerClass={cn(
                                    "items-start w-fit cursor-pointer transition border-transparent rounded-[--radius] p-4",
                                    "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                                    "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                    "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                    "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                                    "py-2",
                                )}
                                help={ThemeMediaPageBannerSizeOptions.find(n => n.value === f.watch("mediaPageBannerSize"))?.description}
                            />

                            <Field.RadioCards
                                label="Banner info layout"
                                name="mediaPageBannerInfoBoxSize"
                                options={ThemeMediaPageInfoBoxSizeOptions.map(n => ({ value: n.value, label: n.label }))}
                                itemContainerClass={cn(
                                    "items-start w-fit cursor-pointer transition border-transparent rounded-[--radius] p-4",
                                    "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                                    "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                    "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                    "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                                    "py-2",
                                )}
                                // help={ThemeMediaPageInfoBoxSizeOptions.find(n => n.value === f.watch("mediaPageBannerInfoBoxSize"))?.description}
                            />

                            <Field.Switch
                                label="Blurred gradient background"
                                name="enableMediaPageBlurredBackground"
                                help="Can cause performance issues."
                            />

                            <br />

                            <h3>Media card</h3>

                            <Field.Switch
                                label="Glassy background"
                                name="enableMediaCardBlurredBackground"
                            />

                            <br />

                            <h3>Episode card</h3>

                            <Field.Switch
                                label="Legacy episode cards"
                                name="useLegacyEpisodeCard"
                            />


                            <br />

                            <h3>Carousel</h3>

                            <Field.Switch
                                label="Disable auto-scroll"
                                name="disableCarouselAutoScroll"
                            />

                            <Field.Switch
                                label="Smaller episode cards"
                                name="smallerEpisodeCarouselSize"
                            />

                        </TabsContent>

                        <TabsContent value="browser-client" className="space-y-4">

                            <Switch
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
