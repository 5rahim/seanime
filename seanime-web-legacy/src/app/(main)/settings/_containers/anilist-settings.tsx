import { useLocalSyncSimulatedDataToAnilist } from "@/api/hooks/local.hooks"
import { SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import React from "react"
import { SiAnilist } from "react-icons/si"

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

    const { mutate: upload, isPending: isUploading } = useLocalSyncSimulatedDataToAnilist()

    const confirmDialog = useConfirmationDialog({
        title: "Upload to AniList",
        description: "This will upload your local Seanime collection to your AniList account. Are you sure you want to proceed?",
        actionText: "Upload",
        actionIntent: "primary",
        onConfirm: async () => {
            if (isUploading) return
            upload()
        },
    })

    return (
        <div className="space-y-4">

            <SettingsPageHeader
                title="AniList"
                description="Manage your AniList account"
                icon={SiAnilist}
            />


            <SettingsSubmitButton isPending={isPending} />

            <ConfirmationDialog {...confirmDialog} />

        </div>
    )
}
