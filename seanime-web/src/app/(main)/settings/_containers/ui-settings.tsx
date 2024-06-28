"use client"
import { useUpdateTheme } from "@/api/hooks/theme.hooks"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { THEME_DEFAULT_VALUES, ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { THEME_COLOR_BANK } from "@/lib/theme/theme-bank"
import { colord } from "colord"
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

    libraryScreenBannerType: z.string().default(THEME_DEFAULT_VALUES.libraryScreenBannerType),
    libraryScreenCustomBannerImage: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerImage),
    libraryScreenCustomBannerPosition: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerPosition),
    libraryScreenCustomBannerOpacity: z.number().transform(v => v === 0 ? 100 : v).default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerOpacity),
    libraryScreenCustomBackgroundImage: z.string().default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundImage),
    libraryScreenCustomBackgroundOpacity: z.number()
        .transform(v => v === 0 ? 100 : v)
        .default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundOpacity),
    disableLibraryScreenGenreSelector: z.boolean().default(false),

}))


export function UISettings() {
    const themeSettings = useThemeSettings()

    const { mutate, isPending } = useUpdateTheme()

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
                    libraryScreenBannerType: themeSettings?.libraryScreenBannerType,
                    libraryScreenCustomBannerImage: themeSettings?.libraryScreenCustomBannerImage,
                    libraryScreenCustomBannerPosition: themeSettings?.libraryScreenCustomBannerPosition,
                    libraryScreenCustomBannerOpacity: themeSettings?.libraryScreenCustomBannerOpacity,
                    libraryScreenCustomBackgroundImage: themeSettings?.libraryScreenCustomBackgroundImage,
                    libraryScreenCustomBackgroundOpacity: themeSettings?.libraryScreenCustomBackgroundOpacity,
                    disableLibraryScreenGenreSelector: themeSettings?.disableLibraryScreenGenreSelector,
                }}
                stackClass="space-y-4"
            >
                {(f) => (
                    <>
                        <h3>Main</h3>

                        <Field.Switch
                            label="Enable color settings"
                            name="enableColorSettings"
                        />
                        <div className="flex flex-col md:flex-row gap-3">
                            <Field.ColorPicker
                                name="backgroundColor"
                                label="Background color"
                                help="Default: #0c0c0c"
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

                        <div className="flex flex-col md:flex-row gap-3">
                            <Field.Text
                                label="Background image path"
                                name="libraryScreenCustomBackgroundImage"
                                placeholder="e.g., image.png"
                                help="This will be used as the background image for the library page."
                            />

                            <Field.Number
                                label="Background image opacity"
                                name="libraryScreenCustomBackgroundOpacity"
                                placeholder="Default: 10"
                                min={1}
                                max={100}
                            />
                        </div>
                        <h3>Sidebar</h3>

                        <Field.Switch
                            label="Expand sidebar on hover"
                            name="expandSidebarOnHover"
                        />

                        <Separator />

                        <h3>Library</h3>

                        <Field.Switch
                            label="Smaller episode carousel size"
                            name="smallerEpisodeCarouselSize"
                        />

                        <Field.Switch
                            label="Disable genre selector"
                            name="disableLibraryScreenGenreSelector"
                        />

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
                            )}
                        />

                        {f.watch("libraryScreenBannerType") === ThemeLibraryScreenBannerType.Custom && (
                            <div className="flex flex-col md:flex-row gap-3">
                                <Field.Text
                                    label="Custom banner image path"
                                    name="libraryScreenCustomBannerImage"
                                    placeholder="e.g., image.gif"
                                />
                                <Field.Text
                                    label="Custom banner position"
                                    name="libraryScreenCustomBannerPosition"
                                    placeholder="Default: 50% 50%"
                                />
                                <Field.Number
                                    label="Custom banner Opacity"
                                    name="libraryScreenCustomBannerOpacity"
                                    placeholder="Default: 10"
                                    min={1}
                                    max={100}
                                />
                            </div>
                        )}

                        <div className="mt-4">
                            <Field.Submit role="save" intent="white" rounded loading={isPending}>Save</Field.Submit>
                        </div>
                    </>
                )}
            </Form>
    )

}
