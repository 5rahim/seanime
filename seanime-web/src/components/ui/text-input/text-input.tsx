import { cn } from "../core/styling"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * TextInput
 * -----------------------------------------------------------------------------------------------*/

export type TextInputProps = Omit<React.ComponentPropsWithRef<"input">, "size"> &
    InputStyling &
    BasicFieldOptions & {
    /**
     * Callback invoked when the value changes. Returns the string value.
     */
    onValueChange?: (value: string) => void
}

export const TextInput = React.forwardRef<HTMLInputElement, TextInputProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<TextInputProps>(props, React.useId())

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
    }] = extractInputPartProps<TextInputProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const handleOnChange = React.useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        onValueChange?.(e.target.value)
        onChange?.(e)
    }, [])

    return (
        <BasicField{...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <input
                    id={basicFieldProps.id}
                    name={basicFieldProps.name}
                    className={cn(
                        "form-input",
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
                        className,
                    )}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-disabled={basicFieldProps.disabled}
                    data-readonly={basicFieldProps.readonly}
                    aria-readonly={basicFieldProps.readonly}
                    required={basicFieldProps.required}
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

TextInput.displayName = "TextInput"
