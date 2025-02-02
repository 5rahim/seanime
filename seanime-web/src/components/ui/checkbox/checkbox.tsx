"use client"

import * as CheckboxPrimitive from "@radix-ui/react-checkbox"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { __CheckboxGroupContext } from "../checkbox"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { mergeRefs } from "../core/utils"
import { hiddenInputStyles } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CheckboxAnatomy = defineStyleAnatomy({
    container: cva("UI-Checkbox__container inline-flex gap-2 items-center"),
    root: cva([
        "UI-Checkbox__root",
        "appearance-none peer block relative overflow-hidden transition h-5 w-5 shrink-0 text-white rounded-[--radius-md] ring-offset-1 border ring-offset-[--background]",
        "border-gray-300 dark:border-gray-700",
        "outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] disabled:cursor-not-allowed data-[disabled=true]:opacity-50",
        "data-[state=unchecked]:bg-white dark:data-[state=unchecked]:bg-gray-700", // Unchecked
        "data-[state=unchecked]:hover:bg-gray-100 dark:data-[state=unchecked]:hover:bg-gray-600", // Unchecked hover
        "data-[state=checked]:bg-brand dark:data-[state=checked]:bg-brand data-[state=checked]:border-brand", // Checked
        "data-[state=indeterminate]:bg-[--muted] dark:data-[state=indeterminate]:bg-gray-700 data-[state=indeterminate]:text-white data-[state=indeterminate]:border-transparent", // Checked
        "data-[error=true]:border-red-500 data-[error=true]:dark:border-red-500 data-[error=true]:data-[state=checked]:border-red-500 data-[error=true]:dark:data-[state=checked]:border-red-500", // Error
    ], {
        variants: {
            size: {
                sm: "h-4 w-4",
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
                sm: "text-sm",
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
        "flex h-full w-full items-center justify-center relative",
    ]),
    checkIcon: cva("UI-Checkbox__checkIcon absolute", {
        variants: {
            size: {
                sm: "h-3 w-3",
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

export type CheckboxProps =
    BasicFieldOptions &
    VariantProps<typeof CheckboxAnatomy.label> &
    ComponentAnatomy<typeof CheckboxAnatomy> &
    Omit<React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root>,
        "value" | "checked" | "disabled" | "required" | "onCheckedChange" | "defaultValue"> & {
    /**
     * If true, no error message will be shown when the field is invalid.
     */
    hideError?: boolean
    /**
     * The size of the checkbox.
     */
    value?: boolean | "indeterminate"
    /**
     * Default value when uncontrolled
     */
    defaultValue?: boolean | "indeterminate"
    /**
     * Callback fired when the value changes
     */
    onValueChange?: (value: boolean | "indeterminate") => void
    /**
     * Ref to the input element
     */
    inputRef?: React.Ref<HTMLInputElement>,
}

export const Checkbox = React.forwardRef<HTMLButtonElement, CheckboxProps>((props, ref) => {

    const [{
        className,
        hideError,
        containerClass,
        checkIconClass,
        labelClass,
        indicatorClass,
        onValueChange,
        defaultValue,
        value: controlledValue,
        size,
        inputRef,
        ...rest
    }, { label, ...basicFieldProps }] = extractBasicFieldProps<CheckboxProps>(props, React.useId())

    const groupContext = React.useContext(__CheckboxGroupContext)

    const _size = groupContext?.group_size ?? size

    const isFirst = React.useRef(true)

    const buttonRef = React.useRef<HTMLButtonElement>(null)

    const [_value, _setValue] = React.useState<boolean | "indeterminate">(controlledValue ?? defaultValue ?? false)

    const handleOnValueChange = React.useCallback((value: boolean) => {
        _setValue(value)
        onValueChange?.(value)
    }, [])

    React.useEffect(() => {
        if (!defaultValue || !isFirst.current) {
            _setValue(controlledValue ?? false)
        }
        isFirst.current = false
    }, [controlledValue])

    return (
        <BasicField
            fieldClass="flex gap-2"
            {...basicFieldProps}
            error={hideError ? undefined : basicFieldProps.error} // The error message hidden when `hideError` is true
        >
            <label
                className={cn(
                    CheckboxAnatomy.container(),
                    containerClass,
                )}
                htmlFor={basicFieldProps.id}
            >
                <CheckboxPrimitive.Root
                    ref={mergeRefs([buttonRef, ref])}
                    id={basicFieldProps.id}
                    className={cn(CheckboxAnatomy.root({ size: _size }), className)}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-error={!!basicFieldProps.error}
                    data-disabled={basicFieldProps.disabled}
                    aria-readonly={basicFieldProps.readonly}
                    data-readonly={basicFieldProps.readonly}
                    checked={_value}
                    onCheckedChange={handleOnValueChange}
                    {...rest}
                >
                    <CheckboxPrimitive.CheckboxIndicator className={cn(CheckboxAnatomy.indicator(), indicatorClass)}>
                        {(_value !== "indeterminate") && <svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 16 16"
                            stroke="currentColor"
                            fill="currentColor"
                            className={cn(CheckboxAnatomy.checkIcon({ size: _size }), checkIconClass)}
                        >
                            <path
                                d="M13.78 4.22a.75.75 0 0 1 0 1.06l-7.25 7.25a.75.75 0 0 1-1.06 0L2.22 9.28a.751.751 0 0 1 .018-1.042.751.751 0 0 1 1.042-.018L6 10.94l6.72-6.72a.75.75 0 0 1 1.06 0Z"
                            />
                        </svg>}

                        {_value === "indeterminate" && <svg
                            xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="3"
                            className={cn(CheckboxAnatomy.checkIcon({ size: _size }), checkIconClass)}
                        >
                            <line x1="5" x2="19" y1="12" y2="12" />
                        </svg>}
                    </CheckboxPrimitive.CheckboxIndicator>
                </CheckboxPrimitive.Root>
                {!!label &&
                    <label
                        className={cn(CheckboxAnatomy.label({ size: _size }), labelClass)}
                        htmlFor={basicFieldProps.id}
                        data-disabled={basicFieldProps.disabled}
                        data-checked={_value === true}
                    >
                        {label}
                    </label>
                }

                <input
                    ref={inputRef}
                    type="checkbox"
                    name={basicFieldProps.name}
                    className={hiddenInputStyles}
                    value={_value === "indeterminate" ? "indeterminate" : _value ? "on" : "off"}
                    checked={basicFieldProps.required ? _value === true : true}
                    aria-hidden="true"
                    required={controlledValue === undefined && basicFieldProps.required}
                    tabIndex={-1}
                    onChange={() => {}}
                    onFocusCapture={() => buttonRef.current?.focus()}
                />
            </label>
        </BasicField>
    )

})

Checkbox.displayName = "Checkbox"
