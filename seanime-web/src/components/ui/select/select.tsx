"use client"

import { cn } from "../core"
import React, { useId } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Select
 * -----------------------------------------------------------------------------------------------*/

export interface SelectProps extends Omit<React.ComponentPropsWithRef<"select">, "size">, InputStyling, BasicFieldOptions {
    options?: { value: string | number, label?: string }[]
}

export const Select = React.forwardRef<HTMLSelectElement, SelectProps>((props, ref) => {

    const [{
        children,
        className,
        size = "md",
        intent = "basic",
        leftIcon,
        leftAddon,
        rightAddon,
        rightIcon,
        options = [],
        placeholder,
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<SelectProps>(props, useId())

    return (
        <>
            <BasicField
                {...basicFieldProps}
            >
                <div className={cn(inputContainerStyle())}>

                    <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size} side={"left"}/>
                    <InputIcon icon={leftIcon} size={size} side={"left"}/>

                    <select
                        id={basicFieldProps.id}
                        name={basicFieldProps.name}
                        className={cn(
                            "form-select",
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
                        disabled={basicFieldProps.isDisabled}
                        {...rest}
                        ref={ref}
                    >
                        {placeholder && <option value="">{placeholder}</option>}
                        {options.map(opt => (
                            <option key={opt.value} value={opt.value}>{opt.label ?? opt.value}</option>
                        ))}
                    </select>

                    <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size}
                                side={"right"}/>
                    <InputIcon icon={rightIcon} size={size} side={"right"}/>

                </div>
            </BasicField>
        </>
    )

})

Select.displayName = "Select"
