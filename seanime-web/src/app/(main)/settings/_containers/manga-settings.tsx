import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Field } from "@/components/ui/form"
import React from "react"
import { FaBookReader } from "react-icons/fa"
import { LuBook, LuBookDashed, LuBookKey } from "react-icons/lu"

type MangaSettingsProps = {
    isPending: boolean
}

export function MangaSettings(props: MangaSettingsProps) {

    const {
        isPending,
        ...rest
    } = props

    const { data: extensions } = useListMangaProviderExtensions()

    const options = React.useMemo(() => {
        return [
            { label: "Auto", value: "-" },
            ...(extensions?.map(provider => ({
                label: provider.name,
                value: provider.id,
            })) ?? []).sort((a, b) => a.label.localeCompare(b.label)),
        ]
    }, [extensions])

    return (
        <>
            <SettingsPageHeader
                title="Manga"
                description="Manage your manga library"
                icon={FaBookReader}
            />

            <SettingsCard>
                <Field.Switch
                    side="right"
                    name="enableManga"
                    label={<span className="flex gap-1 items-center">Enable</span>}
                    help="Read manga series, download chapters and track your progress."
                />
                <Field.Switch
                    side="right"
                    name="mangaAutoUpdateProgress"
                    label="Automatically update progress"
                    help="If enabled, your progress will be automatically updated when you reach the end of a chapter."
                />
            </SettingsCard>

            <SettingsCard title="Sources">
                <Field.Select
                    name="defaultMangaProvider"
                    label="Default Provider"
                    help="Select the default provider for manga series."
                    options={options}
                />

                <Field.DirectorySelector
                    name="mangaLocalSourceDirectory"
                    label="Local Source Directory"
                    help="The directory where your manga is stored. This is only used for local manga provider."
                />
            </SettingsCard>

            <SettingsSubmitButton isPending={isPending} />
        </>
    )
}
