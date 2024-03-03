"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { ThemeSettings } from "@/lib/server/types"
import { THEME_DEFAULT_VALUES, useThemeSettings } from "@/lib/theme/hooks"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { MdVerified } from "react-icons/md"
import { toast } from "sonner"

const themeSchema = defineSchema(({ z }) => z.object({
    animeEntryScreenLayout: z.string().min(0).default(THEME_DEFAULT_VALUES.animeEntryScreenLayout),
    smallerEpisodeCarouselSize: z.boolean().default(THEME_DEFAULT_VALUES.smallerEpisodeCarouselSize),
    expandSidebarOnHover: z.boolean().default(THEME_DEFAULT_VALUES.expandSidebarOnHover),
    backgroundColor: z.string().min(0).default(THEME_DEFAULT_VALUES.backgroundColor),
    sidebarBackgroundColor: z.string().min(0).default(THEME_DEFAULT_VALUES.sidebarBackgroundColor),
    libraryScreenBanner: z.string().default(THEME_DEFAULT_VALUES.libraryScreenBanner),
    libraryScreenBannerPosition: z.string().min(0).default(THEME_DEFAULT_VALUES.libraryScreenBannerPosition),
    libraryScreenCustomBanner: z.string().min(0).default(THEME_DEFAULT_VALUES.libraryScreenCustomBanner),
    libraryScreenCustomBannerAutoDim: z.number().default(THEME_DEFAULT_VALUES.libraryScreenCustomBannerAutoDim),
    libraryScreenShowCustomBackground: z.boolean().default(THEME_DEFAULT_VALUES.libraryScreenShowCustomBackground),
    libraryScreenCustomBackground: z.string().min(0).default(THEME_DEFAULT_VALUES.libraryScreenCustomBackground),
    libraryScreenCustomBackgroundAutoDim: z.number().default(THEME_DEFAULT_VALUES.libraryScreenCustomBackgroundAutoDim),
}))

export default function Page() {
    const [serverStatus, setServerStatus] = useAtom(serverStatusAtom)
    const qc = useQueryClient()

    const themeSettings = useThemeSettings()

    const { mutate, data, isPending } = useSeaMutation<ThemeSettings, ThemeSettings>({
        endpoint: SeaEndpoints.THEME,
        mutationKey: ["patch-theme-settings"],
        method: "patch",
        onSuccess: async () => {
            toast.success("Theme updated")
            await qc.refetchQueries({ queryKey: ["status"] })
        },
    })

    return (
        <PageWrapper className="p-4 sm:p-8 space-y-4">
            <div className="flex gap-4 items-center">
                <Link href={`/settings`}>
                    <IconButton icon={<AiOutlineArrowLeft />} rounded intent="white-outline" size="sm" />
                </Link>
                <div className="space-y-1">
                    <h2>Theme</h2>
                    <p className="text-[--muted]">
                        Change the look and feel of Seanime
                    </p>
                </div>
            </div>

            <p>
                Images should be in a folder called <code>assets</code> at the root of your Seanime installation.
            </p>

            <Separator />
            <Form
                schema={themeSchema}
                onSubmit={data => {
                    mutate({
                        ...themeSettings,
                        ...data,
                    })
                }}
                defaultValues={{
                    animeEntryScreenLayout: themeSettings?.animeEntryScreenLayout,
                    smallerEpisodeCarouselSize: themeSettings?.smallerEpisodeCarouselSize,
                    expandSidebarOnHover: themeSettings?.expandSidebarOnHover,
                    backgroundColor: themeSettings?.backgroundColor,
                    sidebarBackgroundColor: themeSettings?.sidebarBackgroundColor,
                    libraryScreenBanner: themeSettings?.libraryScreenBanner,
                    libraryScreenBannerPosition: themeSettings?.libraryScreenBannerPosition,
                    libraryScreenCustomBanner: themeSettings?.libraryScreenCustomBanner,
                    libraryScreenCustomBannerAutoDim: themeSettings?.libraryScreenCustomBannerAutoDim,
                    libraryScreenShowCustomBackground: themeSettings?.libraryScreenShowCustomBackground,
                    libraryScreenCustomBackground: themeSettings?.libraryScreenCustomBackground,
                    libraryScreenCustomBackgroundAutoDim: themeSettings?.libraryScreenCustomBackgroundAutoDim,
                }}
                stackClass="space-y-4"
            >
                {(f) => (
                    <>
                        <h3>Main</h3>

                        <div className="flex flex-col md:flex-row gap-4 w-full">
                            <Field.ColorPicker
                                name="backgroundColor"
                                label="Background color"
                            />

                            <Field.ColorPicker
                                name="sidebarBackgroundColor"
                                label="Sidebar background color"
                            />
                        </div>

                        <Field.Switch
                            label="Expand sidebar on hover"
                            name="expandSidebarOnHover"
                        />

                        <Separator />

                        <h3>Library Page</h3>

                        <Field.Text
                            label="Background image path"
                            name="libraryScreenCustomBackground"
                            placeholder="e.g., /path/to/image.jpg"
                            help="This will be used as the background image for the library page."
                        />

                        <h5>Continue Watching</h5>

                        <Field.Switch
                            label="Smaller episode carousel size"
                            name="smallerEpisodeCarouselSize"
                        />

                        <h5>Banner</h5>

                        <Field.RadioCards
                            label="Banner type"
                            name="libraryScreenBanner"
                            options={[
                                {
                                    label: "Dynamic Banner",
                                    value: "episode",
                                },
                                {
                                    label: "Custom Banner",
                                    value: "custom",
                                },
                            ]}
                        />

                        {f.watch("libraryScreenBanner") === "custom" && (
                            <>
                                <Field.Text
                                    label="Custom Banner image path"
                                    name="libraryScreenCustomBanner"
                                    placeholder="e.g., /path/to/image.jpg"
                                />
                                <Field.Text label="Custom Banner position" name="libraryScreenBannerPosition" placeholder="50% 50%" />
                            </>
                        )}


                        <Separator />

                        <h3>Anime Entry</h3>

                        <Field.RadioCards
                            label="Layout"
                            name="animeEntryScreenLayout"
                            options={[
                                {
                                    label: <div className="w-full space-y-2">
                                        <p className="mb-1 flex items-center"><MdVerified className="text-lg inline-block mr-2" />New layout</p>
                                        <div className="grid grid-cols-1 gap-2 w-full">
                                            <div className="w-full h-20 rounded-sm bg-gray-600" />
                                            <div className="grid grid-cols-3 gap-2">
                                                <div className="w-full h-12 rounded-sm bg-gray-700" />
                                                <div className="w-full h-12 rounded-sm bg-gray-700" />
                                                <div className="w-full h-12 rounded-sm bg-gray-700" />
                                            </div>
                                            <div className="grid grid-cols-4 gap-2">
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                <div className="w-full h-6 rounded-sm bg-gray-700" />
                                            </div>
                                        </div>
                                    </div>,
                                    value: "stacked",
                                },
                                {
                                    label: <div className="w-full space-y-2">
                                        <p className="mb-1 flex items-center">Old layout</p>
                                        <div className="grid grid-cols-2 gap-2 w-full">
                                            <div className="space-y-2">
                                                <div className="w-full h-20 rounded-sm bg-gray-700" />
                                                <div className="grid grid-cols-3 gap-2">
                                                    <div className="w-full h-8 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-8 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-8 rounded-sm bg-gray-700" />
                                                </div>
                                                <div className="grid grid-cols-2 gap-2">
                                                    <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-6 rounded-sm bg-gray-700" />
                                                </div>
                                            </div>
                                            <div className="space-y-2">
                                                <div className="w-full h-28 rounded-sm bg-gray-600" />
                                                <div className="grid grid-cols-1 gap-2">
                                                    <div className="w-full h-3 rounded-sm bg-gray-700" />
                                                    <div className="w-full h-3 rounded-sm bg-gray-700" />
                                                </div>
                                            </div>
                                        </div>
                                    </div>,
                                    value: "old",
                                },
                            ]}
                            itemContainerClass={cn(
                                "items-start max-w-[300px]  cursor-pointer transition border-transparent rounded-[--radius] p-4 w-full",
                                "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
                                "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
                                "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
                                "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
                            )}
                            itemLabelClass="font-medium flex flex-col items-center data-[state=checked]:text-[--brand] cursor-pointer w-full"
                        />

                        <div className="mt-4">
                            <Field.Submit role="save" loading={isPending}>Save</Field.Submit>
                        </div>
                    </>
                )}
            </Form>
        </PageWrapper>
    )

}
