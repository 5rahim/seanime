import React, { useMemo } from "react"
import { useFormContext } from "react-hook-form"
import locales from "./locales.json"
import { useUILocaleConfig } from "../core"
import { LoadingOverlay } from "../loading-spinner"
import { Button, ButtonProps } from "../button"

/* -------------------------------------------------------------------------------------------------
 * SubmitField
 * -----------------------------------------------------------------------------------------------*/

export interface SubmitFieldProps extends Omit<ButtonProps, "type"> {
    uploadHandler?: any
    role?: "submit" | "save" | "create" | "add" | "search" | "update"
    disableOnSuccess?: boolean
    disableIfInvalid?: boolean
    showLoadingOverlayOnSuccess?: boolean
    loadingOverlay?: React.ReactNode
}

export const SubmitField = React.forwardRef<HTMLButtonElement, SubmitFieldProps>((props, ref) => {

    const {
        children,
        isLoading,
        isDisabled,
        uploadHandler,
        role = "save",
        disableOnSuccess = role === "create",
        disableIfInvalid = false,
        showLoadingOverlayOnSuccess = false,
        loadingOverlay,
        ...rest
    } = props

    const { formState } = useFormContext()
    const { locale } = useUILocaleConfig()

    const disableSuccess = useMemo(() => disableOnSuccess ? formState.isSubmitSuccessful : false, [formState.isSubmitSuccessful])
    const disableInvalid = useMemo(() => disableIfInvalid ? !formState.isValid : false, [formState.isValid])

    return (
        <>
            {((role === "create" || showLoadingOverlayOnSuccess) && !!loadingOverlay) ?? (
                <LoadingOverlay show={formState.isSubmitSuccessful}/>
            )}

            <Button
                type="submit"
                isLoading={formState.isSubmitting || isLoading || uploadHandler?.isLoading} // || ml.mutationLoading}
                isDisabled={disableInvalid || isDisabled || disableSuccess}//|| !formState.isDirty}
                ref={ref}
                {...rest}
            >
                {children ? children : locales["form"][role][locale as "fr" | "en"]}
            </Button>
        </>
    )

})