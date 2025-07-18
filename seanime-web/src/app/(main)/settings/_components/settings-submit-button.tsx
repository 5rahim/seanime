import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { atom, useSetAtom } from "jotai"
import React from "react"
import { useFormContext, useFormState } from "react-hook-form"
import { FiRotateCcw, FiSave } from "react-icons/fi"

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
                    "text-md transition-all group",
                    isDirty && "animate-pulse",
                )}
                intent="white"
                rounded
                loading={isPending}
                leftIcon={<FiSave className="transition-transform duration-200 group-hover:scale-110" />}
            >
                Save
            </Field.Submit>
        </>
    )
}

export function SettingsIsDirty({ className }: { className?: string }) {
    const { isDirty, isLoading, isSubmitting, isValidating } = useFormState()
    const { reset } = useFormContext()
    return isDirty ? <Alert
        intent="info"
        className={cn(
            "absolute -top-4 right-0 p-3 !mt-0 hidden lg:block animate-in slide-in-from-top-2 duration-300",
            className,
        )}
    >
        <div className="flex items-center gap-2">
            <span className="text-sm">You have unsaved changes.</span>
            <Button
                role="save"
                size="md"
                className={cn(
                    "text-md text-[--muted] py-0 h-6 px-2 transition-all duration-200 hover:scale-105 group",
                )}
                intent="white-link"
                onClick={() => reset()}
                leftIcon={<FiRotateCcw className="transition-transform duration-200 group-hover:rotate-180" />}
            >
                Reset
            </Button>
            <Field.Submit
                role="save"
                size="md"
                className={cn(
                    "text-md py-0 h-6 px-2 transition-all duration-200 hover:scale-105 group",
                )}
                intent="white-link"
                disabled={isLoading || isSubmitting || isValidating}
                leftIcon={<FiSave className="transition-transform duration-200 group-hover:scale-110" />}
            >
                Save
            </Field.Submit>
        </div>
    </Alert> : null
}
