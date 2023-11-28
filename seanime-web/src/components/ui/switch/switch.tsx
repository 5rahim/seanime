"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import React, { useId } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import type { SwitchProps as SwitchPrimitiveProps } from "@radix-ui/react-switch"
import * as SwitchPrimitive from "@radix-ui/react-switch"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SwitchAnatomy = defineStyleAnatomy({
    container: cva([
        "UI-Checkbox__rootLabel inline-flex gap-2 items-center"
    ]),
    control: cva([
        "peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border border-transparent transition-colors disabled:cursor-not-allowed disabled:opacity-50",
        "outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] focus-visible:ring-offset-1",
        "data-[state=unchecked]:bg-gray-200 dark:data-[state=unchecked]:bg-gray-700", // Unchecked
        "data-[state=unchecked]:hover:bg-gray-300 dark:data-[state=unchecked]:hover:bg-gray-600", // Unchecked hover
        "data-[state=checked]:bg-brand", // Checked
        "data-[error=true]:border-red-500", // Checked
    ]),
    thumb: cva([
        "pointer-events-none block h-4 w-4 rounded-full bg-white shadow-lg ring-0 transition-transform data-[state=checked]:translate-x-[1.4rem] data-[state=unchecked]:translate-x-1"
    ]),
    label: cva([
        "UI-Switch__label",
        "relative font-normal",
        "data-disabled:text-gray-300",
    ])
})

/* -------------------------------------------------------------------------------------------------
 * Switch
 * -----------------------------------------------------------------------------------------------*/

export interface SwitchProps
    extends Omit<SwitchPrimitiveProps, "disabled" | "required" | "onCheckedChange" | "onChange">,
        ComponentWithAnatomy<typeof SwitchAnatomy>,
        BasicFieldOptions {
    onChange?: (value: boolean) => void
}

export const Switch = React.forwardRef<HTMLButtonElement, SwitchProps>(({ className, ...props }, ref) => {

    const [{
        value,
        onChange,
        controlClassName,
        labelClassName,
        containerClassName,
        thumbClassName,
        ...rest
    }, { label, ...basicFieldProps }] = extractBasicFieldProps(props, useId())

    return (
        <BasicField
            {...basicFieldProps} // We do not include the label
            id={basicFieldProps.id}
        >
            <div
                className={cn(
                    SwitchAnatomy.container(),
                    containerClassName,
                )}
            >
                <SwitchPrimitive.Root
                    id={basicFieldProps.id}
                    ref={ref}
                    className={cn(
                        SwitchAnatomy.control(),
                        controlClassName,
                        className
                    )}
                    disabled={basicFieldProps.isDisabled}
                    required={basicFieldProps.isRequired}
                    data-error={!!basicFieldProps.error}
                    onCheckedChange={(checked) => {
                        onChange && onChange(checked)
                    }}
                    {...rest}
                >
                    <SwitchPrimitive.Thumb
                        className={cn(
                            SwitchAnatomy.thumb(),
                            thumbClassName
                        )}
                    />
                </SwitchPrimitive.Root>
                {(!!label || !!value) &&
                    <label
                        className={cn(
                            SwitchAnatomy.label(),
                            labelClassName,
                        )}
                        htmlFor={basicFieldProps.id}
                    >
                        {label ?? value}
                    </label>
                }
            </div>
        </BasicField>
    )

})

Switch.displayName = "Switch"
