import { cn } from "../core"
import React, { useId } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * TextInput
 * -----------------------------------------------------------------------------------------------*/

export interface TextInputProps extends Omit<React.ComponentPropsWithRef<"input">, "size">,
    Omit<InputStyling, "hasError" | "isDisabled">,
    BasicFieldOptions {
}

export const TextInput = React.forwardRef<HTMLInputElement, TextInputProps>((props, ref) => {

    const [{
        className,
        size = "md",
        intent = "basic",
        leftAddon = undefined,
        leftIcon = undefined,
        rightAddon = undefined,
        rightIcon = undefined,
        disabled,
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<TextInputProps>(props, useId())

    return (
        <>
            <BasicField
                {...basicFieldProps}
            >
                <div className={cn(inputContainerStyle())}>

                    <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size} side={"left"}/>
                    <InputIcon icon={leftIcon} size={size} side={"left"}/>

                    <input
                        id={basicFieldProps.id}
                        name={basicFieldProps.name}
                        className={cn(
                            "form-input",
                            InputAnatomy.input({
                                size,
                                intent,
                                hasError: !!basicFieldProps.error,
                                untouchable: !!basicFieldProps.isDisabled,
                                hasRightAddon: !!rightAddon,
                                hasRightIcon: !!rightIcon,
                                hasLeftAddon: !!leftAddon,
                                hasLeftIcon: !!leftIcon,
                            }),
                            className,
                        )}
                        spellCheck="false"
                        disabled={basicFieldProps.isDisabled || disabled}
                        {...rest}
                        ref={ref}
                    />

                    <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size}
                                side={"right"}/>
                    <InputIcon icon={rightIcon} size={size} side={"right"}/>

                </div>
            </BasicField>
        </>
    )

})

TextInput.displayName = "TextInput"