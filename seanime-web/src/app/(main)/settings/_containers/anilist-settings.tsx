import { useLocalSyncSimulatedDataToAnilist } from "@/api/hooks/local.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button } from "@/components/ui/button"
import { Field } from "@/components/ui/form"
import React from "react"

type Props = {
    isPending: boolean
    children?: React.ReactNode
}

export function AnilistSettings(props: Props) {

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

            <h3>AniList</h3>

            <SettingsCard>
                <Field.Switch
                    side="right"
                    name="hideAudienceScore"
                    label="Hide audience score"
                    help="If enabled, the audience score will be hidden until you decide to view it."
                />

                <Field.Switch
                    side="right"
                    name="disableAnimeCardTrailers"
                    label="Disable anime card trailers"
                    help=""
                />
            </SettingsCard>

            {!serverStatus?.user?.isSimulated && <SettingsCard
                title="Unauthenticated collection"
                description="You can upload your local Seanime collection to your AniList account."
            >
                <Button
                    size="sm"
                    intent="white-subtle"
                    loading={isUploading}
                    onClick={() => {
                        confirmDialog.open()
                    }}
                >
                    Upload to AniList
                </Button>
            </SettingsCard>}

            <SettingsSubmitButton isPending={isPending} />

            <ConfirmationDialog {...confirmDialog} />

        </div>
    )
}
