import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import React from "react"
import { useFormContext } from "react-hook-form"

type DiscordRichPresenceSettingsProps = {
    children?: React.ReactNode
}

export function DiscordRichPresenceSettings(props: DiscordRichPresenceSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const { watch } = useFormContext()

    const enableRichPresence = watch("enableRichPresence")

    return (
        <>
            <Field.Switch
                name="enableRichPresence"
                label={<span className="flex gap-1 items-center">Discord Rich Presence</span>}
            />
            <div
                className={cn(
                    "flex gap-4 items-center flex-col md:flex-row !mt-3",
                    enableRichPresence ? "opacity-100" : "opacity-50 pointer-events-none",
                )}
            >
                <Field.Checkbox
                    name="enableAnimeRichPresence"
                    label="Anime"
                    help="Show what you are watching in Discord."
                    fieldClass="w-fit"
                />
                <Field.Checkbox
                    name="enableMangaRichPresence"
                    label="Manga"
                    help="Show what you are reading in Discord."
                    fieldClass="w-fit"
                />
            </div>

            <Field.Switch
                name="richPresenceHideSeanimeRepositoryButton"
                label="Hide Seanime Repository Button"
            />

            <Field.Switch
                name="richPresenceShowAniListMediaButton"
                label="Show AniList Media Button"
                help="Show a button to open the media page on AniList."
            />

            <Field.Switch
                name="richPresenceShowAniListProfileButton"
                label="Show AniList Profile Button"
                help="Show a button to open your profile page on AniList."
            />
        </>
    )
}
