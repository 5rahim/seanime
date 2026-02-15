import { cn } from "../core/styling"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"
import { BiHide, BiShow } from "react-icons/bi"

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
        type,
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

    const [showPassword, setShowPassword] = React.useState(false)
    const isPasswordInput = type === "password"
    const actualType = isPasswordInput ? (showPassword ? "text" : "password") : type

    const handleOnChange = React.useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        onValueChange?.(e.target.value)
        onChange?.(e)
    }, [])

    const togglePasswordVisibility = React.useCallback(() => {
        setShowPassword(prev => !prev)
    }, [])

    const finalRightAddon = isPasswordInput
        ? (showPassword ? <BiHide className="cursor-pointer" onClick={togglePasswordVisibility} /> : <BiShow
            className="cursor-pointer"
            onClick={togglePasswordVisibility}
        />)
        : rightAddon

    return (
        <BasicField{...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <input
                    id={basicFieldProps.id}
                    name={basicFieldProps.name}
                    type={actualType}
                    className={cn(
                        "form-input",
                        InputAnatomy.root({
                            size,
                            intent,
                            hasError: !!basicFieldProps.error,
                            isDisabled: !!basicFieldProps.disabled,
                            isReadonly: !!basicFieldProps.readonly,
                            hasRightAddon: !!rightAddon || isPasswordInput,
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

                <InputAddon {...rightAddonProps} addon={finalRightAddon} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )

})

TextInput.displayName = "TextInput"
