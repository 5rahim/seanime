import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn } from "../core/styling"
import { extractInputPartProps, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * NativeSelect
 * -----------------------------------------------------------------------------------------------*/

export type NativeSelectProps = Omit<React.ComponentPropsWithRef<"select">, "size"> &
    InputStyling &
    BasicFieldOptions & {
    /**
     * The options to display
     */
    options: { value: string | number, label?: string }[] | undefined
    /**
     * The placeholder text
     */
    placeholder?: string
}

export const NativeSelect = React.forwardRef<HTMLSelectElement, NativeSelectProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<NativeSelectProps>(props, React.useId())

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        className,
        placeholder,
        options,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<NativeSelectProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    return (
        <BasicField{...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <select
                    id={basicFieldProps.id}
                    name={basicFieldProps.name}
                    className={cn(
                        "form-select",
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
                    {...rest}
                    ref={ref}
                >
                    {placeholder && <option value="">{placeholder}</option>}
                    {options?.map(opt => (
                        <option key={opt.value} value={opt.value}>{opt.label ?? opt.value}</option>
                    ))}
                </select>

                <InputAddon {...rightAddonProps} />
                <InputIcon
                    {...rightIconProps}
                    className={cn(
                        rightIconProps.className,
                        !rightAddon ? "mr-8" : null,
                    )}
                />
            </InputContainer>
        </BasicField>
    )

})

NativeSelect.displayName = "NativeSelect"
