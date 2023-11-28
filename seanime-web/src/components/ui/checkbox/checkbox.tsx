"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React, { useId } from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import type { CheckboxProps as CheckboxPrimitiveProps } from "@radix-ui/react-checkbox"
import * as CheckboxPrimitive from "@radix-ui/react-checkbox"
import { useCheckboxGroupContext } from "../checkbox"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CheckboxAnatomy = defineStyleAnatomy({
    container: cva("UI-Checkbox__container inline-flex gap-2 items-center"),
    control: cva([
        "UI-Checkbox__root",
        "appearance-none peer block relative overflow-hidden transition h-5 w-5 shrink-0 text-white rounded-md ring-offset-1 border ring-offset-background",
        "border-gray-300 dark:border-gray-700",
        "outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] disabled:cursor-not-allowed disabled:opacity-50",
        "data-[state=unchecked]:bg-white dark:data-[state=unchecked]:bg-gray-700", // Unchecked
        "data-[state=unchecked]:hover:bg-gray-100 dark:data-[state=unchecked]:hover:bg-gray-600", // Unchecked hover
        "data-[state=checked]:bg-brand dark:data-[state=checked]:bg-brand data-[state=checked]:border-brand", // Checked
        "data-[state=indeterminate]:bg-[--muted] dark:data-[state=indeterminate]:text-gray-800 data-[state=indeterminate]:border-transparent", // Checked
        "data-[error=true]:border-red-500 data-[error=true]:dark:border-red-500 data-[error=true]:data-[state=checked]:border-red-500 data-[error=true]:dark:data-[state=checked]:border-red-500" // Error
    ], {
        variants: {
            size: {
                md: "h-5 w-5",
                lg: "h-6 w-6",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    label: cva([
        "UI-Checkbox_label",
        "font-normal",
        "data-[disabled=true]:opacity-50",
    ], {
        variants: {
            size: {
                md: "text-md",
                lg: "text-lg",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    indicator: cva([
        "UI-Checkbox__indicator",
        "flex h-full w-full items-center justify-center relative"
    ]),
    icon: cva("UI-Checkbox__icon absolute", {
        variants: {
            size: {
                md: "h-4 w-4",
                lg: "h-5 w-5",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * Checkbox
 * -----------------------------------------------------------------------------------------------*/

export interface CheckboxProps extends Omit<CheckboxPrimitiveProps, "disabled" | "required" | "onCheckedChange" | "onChange">,
    ComponentWithAnatomy<typeof CheckboxAnatomy>,
    VariantProps<typeof CheckboxAnatomy.label>,
    BasicFieldOptions {
    noErrorMessage?: boolean
    onChange?: (value: boolean | "indeterminate") => void
}

export const Checkbox = React.forwardRef<HTMLButtonElement, CheckboxProps>((props, ref) => {

    const [{
        className,
        noErrorMessage,
        containerClassName,
        controlClassName,
        iconClassName,
        labelClassName,
        indicatorClassName,
        onChange,
        value,
        size = "md",
        ...rest
    }, { label, ...basicFieldProps }] = extractBasicFieldProps<CheckboxProps>(props, useId())

    const groupContext = useCheckboxGroupContext()

    const _size = groupContext?.group_size ?? size

    return (
        <BasicField
            fieldClassName={"space-y-.5"}
            {...basicFieldProps}
            error={noErrorMessage ? undefined : basicFieldProps.error} // The error message hidden when `noErrorMessage` is defined
        >
            <label
                className={cn(
                    CheckboxAnatomy.container(),
                    containerClassName
                )}
                htmlFor={basicFieldProps.id}
            >
                <CheckboxPrimitive.Root
                    id={basicFieldProps.id}
                    ref={ref}
                    className={cn(
                        CheckboxAnatomy.control({
                            size: _size,
                        }),
                        controlClassName,
                        className
                    )}
                    disabled={basicFieldProps.isDisabled}
                    required={basicFieldProps.isRequired}
                    data-error={!!basicFieldProps.error}
                    onCheckedChange={(value) => {
                        onChange && onChange(value)
                    }}
                    {...rest}
                >
                    <CheckboxPrimitive.CheckboxIndicator
                        className={cn(CheckboxAnatomy.indicator(), indicatorClassName)}>
                        {(rest.checked !== "indeterminate") && <svg
                            xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" stroke="currentColor"
                            fill="currentColor"
                            className={cn(CheckboxAnatomy.icon({ size: _size }), iconClassName)}
                        >
                            <path
                                fill="#fff"
                                d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"
                            ></path>
                        </svg>}

                        {rest.checked === "indeterminate" &&
                            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                 fill="none" stroke="currentColor"
                                 strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
                                <line x1="5" x2="19" y1="12" y2="12"/>
                            </svg>}
                    </CheckboxPrimitive.CheckboxIndicator>
                </CheckboxPrimitive.Root>
                {(!!label || !!value) &&
                    <label
                        className={cn(
                            CheckboxAnatomy.label({ size: _size }),
                            labelClassName,
                        )}
                        htmlFor={basicFieldProps.id}
                        data-disabled={basicFieldProps.isDisabled}
                    >
                        {label ?? value}
                    </label>
                }
            </label>
        </BasicField>
    )

})

Checkbox.displayName = "Checkbox"
