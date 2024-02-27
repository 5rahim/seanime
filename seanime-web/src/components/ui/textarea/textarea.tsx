import { cva } from "class-variance-authority"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn, defineStyleAnatomy } from "../core/styling"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TextareaAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Textarea__root",
        "w-full p-2",
    ], {
        variants: {
            size: {
                sm: "h-20",
                md: "h-32",
                lg: "h-64",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * Textarea
 * -----------------------------------------------------------------------------------------------*/

export type TextareaProps = Omit<React.ComponentPropsWithRef<"textarea">, "size"> &
    InputStyling &
    BasicFieldOptions & {
    /**
     * Callback invoked when the value changes. Returns the string value.
     */
    onValueChange?: (value: string) => void
}

export const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<TextareaProps>(props, React.useId())

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        className,
        onValueChange,
        onChange,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<TextareaProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const handleOnChange = React.useCallback((e: React.ChangeEvent<HTMLTextAreaElement>) => {
        onValueChange?.(e.target.value)
        onChange?.(e)
    }, [])

    return (
        <BasicField {...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <textarea
                    id={basicFieldProps.id}
                    name={basicFieldProps.name}
                    className={cn(
                        "form-textarea",
                        InputAnatomy.root({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            isDisabled: !!basicFieldProps.disabled,
                            isReadonly: !!basicFieldProps.readonly,
                            hasRightAddon: !!rightAddon,
                            hasRightIcon: !!rightIcon,
                            hasLeftAddon: !!leftAddon,
                            hasLeftIcon: !!leftIcon,
                        }),
                        TextareaAnatomy.root({ size }),
                        className,
                    )}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-disabled={basicFieldProps.disabled}
                    onChange={handleOnChange}
                    {...rest}
                    ref={ref}
                />

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )

})

Textarea.displayName = "Textarea"
