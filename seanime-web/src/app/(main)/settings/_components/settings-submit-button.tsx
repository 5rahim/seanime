import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { atom, useSetAtom } from "jotai"
import React from "react"
import { useFormContext, useFormState } from "react-hook-form"

export const settingsFormIsDirtyAtom = atom(false)

export function SettingsSubmitButton({ isPending }: { isPending: boolean }) {

    const { isDirty } = useFormState()

    const setSettingsFormIsDirty = useSetAtom(settingsFormIsDirtyAtom)

    React.useEffect(() => {
        setSettingsFormIsDirty(isDirty)
    }, [isDirty])

    return (
        <>
            <Field.Submit
                role="save"
                size="md"
                className={cn(
                    "text-md",
                    isDirty && "animate-pulse",
                )}
                intent="white"
                rounded
                loading={isPending}
            >
                Save
            </Field.Submit>
        </>
    )
}

export function SettingsIsDirty({ className }: { className?: string }) {
    const { isDirty, isLoading, isSubmitting, isValidating } = useFormState()
    const { reset } = useFormContext()
    return isDirty ? <Alert intent="info" className={cn("absolute -top-4 right-0 p-3 !mt-0 hidden lg:block", className)}>
        <div>
            You have unsaved changes. <Button
            role="save"
            size="md"
            className={cn(
                "text-md text-[--muted] py-0 h-5 px-1",
            )}
            intent="white-link"
            onClick={() => reset()}
        >
            Reset
        </Button> <Field.Submit
            role="save"
            size="md"
            className={cn(
                "text-md py-0 h-5 px-1",
            )}
            intent="white-link"
            disabled={isLoading || isSubmitting || isValidating}
        >
            Save
        </Field.Submit>
        </div>
    </Alert> : null
}
