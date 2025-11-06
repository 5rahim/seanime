import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { useStoredMangaProviders } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { Field } from "@/components/ui/form"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { useFormContext } from "react-hook-form"
import { LuBookOpen } from "react-icons/lu"
import { toast } from "sonner"

type MangaSettingsProps = {
    isPending: boolean
}

const __manga_storedProvidersHistoryAtom = atom<Record<string, string> | null>(null)

export function MangaSettings(props: MangaSettingsProps) {

    const {
        isPending,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const f = useFormContext()

    const { data: extensions } = useListMangaProviderExtensions()

    const { storedProviders, overwriteStoredProviders, overwriteStoredProvidersWith } = useStoredMangaProviders(extensions)
    const [storedProvidersHistory, setStoredProvidersHistory] = useAtom(__manga_storedProvidersHistoryAtom)

    const options = React.useMemo(() => {
        return [
            { label: "Auto", value: "-" },
            ...(extensions?.map(provider => ({
                label: provider.name,
                value: provider.id,
            })) ?? []).sort((a, b) => a.label.localeCompare(b.label)),
        ]
    }, [extensions])

    const defaultProviderExt = extensions?.find(e => e.id === serverStatus?.settings?.manga?.defaultMangaProvider)

    const confirmDialog = useConfirmationDialog({
        title: "Overwrite all sources",
        description: "This will overwrite the selected source of all manga series you've opened with the default provider. Are you sure you want to proceed?",
        actionText: "Overwrite",
        actionIntent: "warning",
        onConfirm: async () => {
            if (!defaultProviderExt) return
            const oldProviders = structuredClone(storedProviders)
            overwriteStoredProvidersWith(defaultProviderExt.id)
            toast.success("All source selections have been overwritten.")
            setTimeout(() => {
                setStoredProvidersHistory(oldProviders)
            }, 500)
        },
    })

    return (
        <>
            <SettingsPageHeader
                title="Manga"
                description="Manage your manga library"
                icon={LuBookOpen}
            />

            <SettingsCard>
                <Field.Switch
                    side="right"
                    name="enableManga"
                    label={<span className="flex gap-1 items-center">Enable</span>}
                    help="Read manga series, download chapters and track your progress."
                />
            </SettingsCard>

            <SettingsCard>
                <Field.Select
                    name="defaultMangaProvider"
                    label="Default Provider"
                    help="Provider selected by default when opening a new manga series."
                    options={options}
                />
                {(!!defaultProviderExt && f.watch("defaultMangaProvider") === serverStatus?.settings?.manga?.defaultMangaProvider) && (
                    <div className="flex w-full space-x-4 flex-wrap">
                        <Button className="px-0 py-1" intent="warning-link" onClick={() => confirmDialog.open()}>
                            Overwrite all manga sources with {defaultProviderExt.name}
                        </Button>
                        {!!storedProvidersHistory && (
                            <Button
                                className="px-0 py-1" intent="gray-link" onClick={() => {
                                overwriteStoredProviders(storedProvidersHistory)
                                toast.success("Previous source selections have been restored.")
                                setStoredProvidersHistory(null)
                            }}
                            >
                                Undo
                            </Button>
                        )}
                    </div>
                )}
                <Field.Switch
                    side="right"
                    name="mangaAutoUpdateProgress"
                    label="Automatically update progress"
                    help="If enabled, your progress will be automatically updated when you reach the end of a chapter."
                />
            </SettingsCard>

            <SettingsCard title="Local Provider" description="Read manga series from your local directory.">

                <Field.DirectorySelector
                    name="mangaLocalSourceDirectory"
                    label="Local Source Directory"
                    help="Directory where your manga is stored. This is only used by the local manga provider."
                />
            </SettingsCard>

            <ConfirmationDialog {...confirmDialog} />

            <SettingsSubmitButton isPending={isPending} />
        </>
    )
}
