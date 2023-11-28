import { cn } from "../core"
import React, { useId } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { InputAnatomy, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Textarea
 * -----------------------------------------------------------------------------------------------*/

export interface TextareaProps extends React.ComponentPropsWithRef<"textarea">, InputStyling, BasicFieldOptions {
}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>((props, ref) => {

    const [{
        children,
        className,
        intent = "basic",
        ...rest
    }, basicFieldProps] = extractBasicFieldProps(props, useId())

    return (
        <>
            <BasicField
                className={cn("w-full gap-1")}
                {...basicFieldProps}
            >
            <textarea
                id={basicFieldProps.id}
                className={cn(
                    "form-textarea",
                    InputAnatomy.input({
                        intent,
                        hasError: !!basicFieldProps.error,
                        untouchable: !!basicFieldProps.isDisabled,
                    }),
                    className,
                )}
                {...rest}
                ref={ref}
            />
            </BasicField>
        </>
    )

})

Textarea.displayName = "Textarea"
