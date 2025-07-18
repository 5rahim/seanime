import { useLocalSyncSimulatedDataToAnilist } from "@/api/hooks/local.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { LuCloudUpload, LuDatabase, LuUserCog } from "react-icons/lu"

type Props = {
    isPending: boolean
    children?: React.ReactNode
}

export function LocalSettings(props: Props) {

    const {
        isPending,
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const { mutate: upload, isPending: isUploading } = useLocalSyncSimulatedDataToAnilist()

    const confirmDialog = useConfirmationDialog({
        title: "Upload to AniList",
        description: "This will upload your local Seanime collection to your AniList account. Are you sure you want to proceed?",
        actionText: "Upload",
        actionIntent: "primary",
        onConfirm: async () => {
            upload()
        },
    })

    return (
        <div className="space-y-4">

            <SettingsPageHeader
                title="Local Account"
                description="Local anime and manga list managed by Seanime"
                icon={LuUserCog}
            />

            <SettingsCard
                title="AniList"
                // description="You can upload your local Seanime collection to your AniList account."
            >
                <div className={cn(serverStatus?.user?.isSimulated && "opacity-50 pointer-events-none")}>
                    <Field.Switch
                        side="right"
                        name="autoSyncToLocalAccount"
                        label="Auto sync from AniList"
                        help="Periodically update your local collection by using your AniList data."
                    />
                </div>
                <Separator />
                <Button
                    size="sm"
                    intent="primary-subtle"
                    loading={isUploading}
                    leftIcon={<LuCloudUpload className="size-4" />}
                    onClick={() => {
                        confirmDialog.open()
                    }}
                    disabled={serverStatus?.user?.isSimulated}
                >
                    Upload to AniList
                </Button>
            </SettingsCard>

            <SettingsSubmitButton isPending={isPending} />

            <ConfirmationDialog {...confirmDialog} />

        </div>
    )
}
