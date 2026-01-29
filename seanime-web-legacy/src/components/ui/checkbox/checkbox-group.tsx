"use client"

import * as React from "react"
import { Checkbox, CheckboxProps } from "."
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn } from "../core/styling"
import { hiddenInputStyles } from "../input"

/* -------------------------------------------------------------------------------------------------
 * CheckboxGroup
 * -----------------------------------------------------------------------------------------------*/

type CheckboxGroupContextValue = {
    group_size: CheckboxProps["size"]
}

export const __CheckboxGroupContext = React.createContext<CheckboxGroupContextValue | null>(null)

export type CheckboxGroupOption = { value: string, label?: React.ReactNode, disabled?: boolean, readonly?: boolean }

export type CheckboxGroupProps = BasicFieldOptions & {
    /**
     * The value of the checkbox group.
     */
    value?: string[]
    /**
     * The default value of the checkbox group when uncontrolled.
     */
    defaultValue?: string[]
    /**
     * Callback invoked when the value of the checkbox group changes.
     */
    onValueChange?: (value: string[]) => void
    /**
     * The size of the checkboxes.
     */
    size?: CheckboxProps["size"]
    /**
     * The options of the checkbox group.
     */
    options: CheckboxGroupOption[]
    /**
     * Class names applied to the container.
     */
    stackClass?: string
    /**
     * Class names applied to each checkbox container.
     */
    itemContainerClass?: string
    /**
     * Class names applied to each checkbox label.
     */
    itemLabelClass?: string
    /**
     * Class names applied to each checkbox button.
     */
    itemClass?: string
    /**
     * Class names applied to each checkbox check icon.
     */
    itemCheckIconClass?: string
}

export const CheckboxGroup = React.forwardRef<HTMLInputElement, CheckboxGroupProps>((props, ref) => {

    const [{
        value: controlledValue,
        defaultValue = [],
        onValueChange,
        stackClass,
        itemLabelClass,
        itemClass,
        itemContainerClass,
        itemCheckIconClass,
        options,
        size = undefined,
    }, basicFieldProps] = extractBasicFieldProps<CheckboxGroupProps>(props, React.useId())

    const [value, setValue] = React.useState<string[]>(controlledValue ?? defaultValue)

    const handleUpdateValue = React.useCallback((v: string) => {
        return (checked: boolean | "indeterminate") => {
            setValue(p => {
                const newArr = checked === true
                    ? [...p, ...(p.includes(v) ? [] : [v])]
                    : checked === false
                        ? p.filter(v1 => v1 !== v)
                        : [...p]
                onValueChange?.(newArr)
                return newArr
            })
        }
    }, [])

    React.useEffect(() => {
        if (controlledValue !== undefined) {
            setValue(controlledValue)
        }
    }, [controlledValue])


    return (
        <__CheckboxGroupContext.Provider
            value={{
                group_size: size,
            }}
        >
            <BasicField {...basicFieldProps}>
                <div className={cn("UI-CheckboxGroup__stack space-y-1", stackClass)}>
                    {options.map((opt) => (
                        <Checkbox
                            key={opt.value}
                            label={opt.label}
                            value={value.includes(opt.value)}
                            onValueChange={handleUpdateValue(opt.value)}
                            hideError
                            error={basicFieldProps.error}
                            className={itemClass}
                            labelClass={itemLabelClass}
                            containerClass={itemContainerClass}
                            checkIconClass={itemCheckIconClass}
                            disabled={basicFieldProps.disabled || opt.disabled}
                            readonly={basicFieldProps.readonly || opt.readonly}
                            tabIndex={0}
                        />
                    ))}
                </div>

                <input
                    ref={ref}
                    type="text"
                    id={basicFieldProps.name}
                    name={basicFieldProps.name}
                    className={hiddenInputStyles}
                    value={basicFieldProps.required
                        ? (!!value.length ? JSON.stringify(value) : "")
                        : JSON.stringify(value)}
                    aria-hidden="true"
                    required={basicFieldProps.required}
                    tabIndex={-1}
                    onChange={() => {}}
                />

            </BasicField>
        </__CheckboxGroupContext.Provider>
    )

})

CheckboxGroup.displayName = "CheckboxGroup"
