import { BetaBadge } from "@/components/application/beta-badge"
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
                label={<span className="flex gap-1 items-center">Discord Rich Presence <BetaBadge /></span>}
            />
            <div
                className={cn(
                    "flex gap-4 items-center flex-col md:flex-row md:pl-4 !mt-3",
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
        </>
    )
}
