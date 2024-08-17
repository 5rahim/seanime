import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Field } from "@/components/ui/form"
import React from "react"

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
            <h3>Manga</h3>

            <Field.Switch
                name="enableManga"
                label={<span className="flex gap-1 items-center">Enable</span>}
                help="Read manga series, download chapters and track your progress."
            />

            <Field.Select
                name="defaultMangaProvider"
                label="Default Provider"
                help="Select the default provider for manga series."
                options={options}
            />

            <SettingsSubmitButton isPending={isPending} />
        </>
    )
}
