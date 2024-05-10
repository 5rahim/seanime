import { Field } from "@/components/ui/form"
import React from "react"

export function SettingsSubmitButton({ isPending }: { isPending: boolean }) {
    return (
        <div className="mt-8">
            <Field.Submit role="save" intent="white" rounded loading={isPending}>Save</Field.Submit>
        </div>
    )
}
