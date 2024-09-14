import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { FcFolder } from "react-icons/fc"

type LibrarySettingsProps = {
    isPending: boolean
}

export function LibrarySettings(props: LibrarySettingsProps) {

    const {
        isPending,
        ...rest
    } = props


    return (
        <div className="space-y-4">

            <Field.DirectorySelector
                name="libraryPath"
                label="Library directory"
                leftIcon={<FcFolder />}
                help="Directory where your media is located. (Keep the casing consistent)"
                shouldExist
            />

            <Field.MultiDirectorySelector
                name="libraryPaths"
                label="Additional library directories"
                leftIcon={<FcFolder />}
                help="Include additional directories if your library is spread across multiple locations."
                shouldExist
            />

            <Separator />

            <Field.Switch
                name="autoScan"
                label="Automatically refresh library"
                help={<div>
                    <p>If enabled, your library will be refreshed in the background when new files are added/deleted. Make sure to
                       lock your files regularly.</p>
                    <p>
                        <em>Note:</em> This works best when single files are added/deleted. If you are adding a batch, not all
                                       files
                                       are guaranteed to be picked up.
                    </p>
                </div>}
            />

            <Field.Switch
                name="refreshLibraryOnStart"
                label="Refresh library on startup"
                help={<div>
                    <p>If enabled, your library will be refreshed in the background when the server starts. Make sure to
                       lock your files regularly.</p>
                    <p>
                        <em>Note:</em> Visit the scan summary page to see the results.
                    </p>
                </div>}
            />

            <Field.Switch
                name="enableWatchContinuity"
                label="Enable watch continuity"
                help="If enabled, Seanime will remember your watch progress and resume from where you left off."
            />

            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}
