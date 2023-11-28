"use client"

import { cn } from "../core"
import React, { createContext, useContext, useId, useLayoutEffect, useState } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { Checkbox, CheckboxProps } from "."


/* -------------------------------------------------------------------------------------------------
 * Provider
 * -----------------------------------------------------------------------------------------------*/

interface CheckboxGroupContextValue {
    group_size: CheckboxProps["size"]
}

const _CheckboxGroupContext = createContext<CheckboxGroupContextValue | null>(null)
export const CheckboxGroupProvider = _CheckboxGroupContext.Provider
export const useCheckboxGroupContext = () => useContext(_CheckboxGroupContext)

/* -------------------------------------------------------------------------------------------------
 * CheckboxGroup
 * -----------------------------------------------------------------------------------------------*/

export interface CheckboxGroupProps extends BasicFieldOptions {
    value?: string[]
    defaultValue?: string[]
    onChange?: (value: string[]) => void
    size?: CheckboxProps["size"]
    stackClassName?: string
    checkboxContainerClassName?: string
    checkboxLabelClassName?: string
    checkboxControlClassName?: string
    checkboxIconClassName?: string
    options: { value: string, label?: React.ReactNode }[]
}

export const CheckboxGroup = React.forwardRef<HTMLDivElement, CheckboxGroupProps>((props, ref) => {

    const [{
        value,
        defaultValue = [],
        onChange,
        stackClassName,
        checkboxLabelClassName,
        checkboxControlClassName,
        checkboxContainerClassName,
        checkboxIconClassName,
        options,
        size = undefined,
    }, basicFieldProps] = extractBasicFieldProps<CheckboxGroupProps>(props, useId())

    // Keep track of selected values
    const [selectedValues, setSelectedValues] = useState<string[]>(value ?? defaultValue)

    // Control the state
    useLayoutEffect(() => {
        if (value) {
            setSelectedValues(value)
        }
    }, [value])


    return (
        <>
            <CheckboxGroupProvider value={{
                group_size: size
            }}>
                <BasicField
                    {...basicFieldProps}
                    ref={ref}
                >
                    <div className={cn("space-y-1", stackClassName)}>
                        {options.map((opt) => (
                            <Checkbox
                                key={opt.value}
                                label={opt.label}
                                value={opt.value}
                                checked={selectedValues.includes(opt.value)}
                                onChange={checked => {
                                    setSelectedValues(p => {
                                        let newArr = [...p]
                                        if (checked === true) {
                                            if (p.indexOf(opt.value) === -1) newArr.push(opt.value)
                                        } else if (checked === false) {
                                            newArr = newArr.filter(v => v !== opt.value)
                                        }
                                        if (onChange) {
                                            onChange(newArr)
                                        }
                                        return newArr
                                    })
                                }}
                                error={basicFieldProps.error}
                                noErrorMessage
                                labelClassName={checkboxLabelClassName}
                                controlClassName={checkboxControlClassName}
                                containerClassName={checkboxContainerClassName}
                                iconClassName={checkboxIconClassName}
                                isDisabled={basicFieldProps.isDisabled}
                                isReadOnly={basicFieldProps.isReadOnly}
                                tabIndex={0}
                            />
                        ))}
                    </div>
                </BasicField>
            </CheckboxGroupProvider>
        </>
    )

})

CheckboxGroup.displayName = "CheckboxGroup"
