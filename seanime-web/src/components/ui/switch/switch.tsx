"use client"

import { hiddenInputStyles } from "@/components/ui/input"
import { Popover } from "@/components/ui/popover"
import * as SwitchPrimitive from "@radix-ui/react-switch"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { AiOutlineExclamationCircle } from "react-icons/ai"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { mergeRefs } from "../core/utils"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/
export const SwitchAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Switch__root",
        "peer inline-flex shrink-0 cursor-pointer items-center rounded-full border border-transparent transition-colors",
        "disabled:cursor-not-allowed data-[disabled=true]:opacity-50",
        "outline-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] focus-visible:ring-offset-1",
        "data-[state=unchecked]:bg-gray-200 dark:data-[state=unchecked]:bg-gray-700", // Unchecked
        "data-[state=unchecked]:hover:bg-gray-300 dark:data-[state=unchecked]:hover:bg-gray-600", // Unchecked hover
        "data-[state=checked]:bg-brand", // Checked
        "data-[error=true]:border-red-500", // Checked
    ], {
        variants: {
            size: {
                sm: "h-5 w-9",
                md: "h-6 w-11",
                lg: "h-7 w-14",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    container: cva([
        "UI-Switch__container",
        "inline-flex gap-2 items-center",
    ], {
        variants: {
            side: {
                left: "",
                right: "w-full flex-row-reverse",
            },
        },
        defaultVariants: {
            side: "left",
        },
    }),
    thumb: cva([
        "UI-Switch__thumb",
        "pointer-events-none block rounded-full bg-white shadow-lg ring-0 transition-transform",
        "data-[state=unchecked]:translate-x-1",
    ], {
        variants: {
            size: {
                sm: "h-3 w-3 data-[state=checked]:translate-x-[1.1rem]",
                md: "h-4 w-4 data-[state=checked]:translate-x-[1.4rem]",
                lg: "h-5 w-5 data-[state=checked]:translate-x-[1.9rem]",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    label: cva([
        "UI-Switch__label",
        "relative font-normal",
        "data-[disabled=true]:text-gray-300 cursor-pointer user-select-none select-none",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Switch
 * -----------------------------------------------------------------------------------------------*/

export type SwitchProps = BasicFieldOptions &
    ComponentAnatomy<typeof SwitchAnatomy> &
    VariantProps<typeof SwitchAnatomy.root> &
    VariantProps<typeof SwitchAnatomy.container> &
    Omit<React.ComponentPropsWithoutRef<typeof SwitchPrimitive.Root>,
        "value" | "checked" | "disabled" | "required" | "defaultValue" | "defaultChecked" | "onCheckedChange"> & {
    /**
     * Whether the switch is checked
     */
    value?: boolean
    /**
     * Callback fired when the value changes
     */
    onValueChange?: (value: boolean) => void
    /**
     * Default value when uncontrolled
     */
    defaultValue?: boolean
    /**
     * Ref to the input element
     */
    inputRef?: React.Ref<HTMLInputElement>
    className?: string
    moreHelp?: React.ReactNode
}

export const Switch = React.forwardRef<HTMLButtonElement, SwitchProps>((props, ref) => {

    const [{
        size,
        value: controlledValue,
        className,
        onValueChange,
        labelClass,
        containerClass,
        thumbClass,
        defaultValue,
        inputRef,
        side,
        moreHelp,
        ...rest
    }, { label, ...basicFieldProps }] = extractBasicFieldProps(props, React.useId())

    const isFirst = React.useRef(true)

    const buttonRef = React.useRef<HTMLButtonElement>(null)

    const [_value, _setValue] = React.useState<boolean | undefined>(controlledValue ?? defaultValue ?? false)

    const handleOnValueChange = React.useCallback((value: boolean) => {
        _setValue(value)
        onValueChange?.(value)
    }, [])

    React.useEffect(() => {
        if (!defaultValue || !isFirst.current) {
            _setValue(controlledValue)
        }
        isFirst.current = false
    }, [controlledValue])

    return (
        <BasicField
            {...basicFieldProps}
            id={basicFieldProps.id}
            fieldClass={cn(
                "w-fit",
                side === "right" && "w-full group/switch transition-all duration-200 hover:bg-gray-300/5 rounded-[--radius] p-2 w-[calc(100%_+_1rem)] -ml-2 border border-transparent border-dashed hover:border-[--subtle]",
                basicFieldProps.fieldClass,
            )}
            fieldHelpTextClass={cn("")}
        >
            <div className={cn(SwitchAnatomy.container({ side }), containerClass)}>
                <SwitchPrimitive.Root
                    ref={mergeRefs([buttonRef, ref])}
                    id={basicFieldProps.id}
                    className={cn(SwitchAnatomy.root({ size }), className)}
                    disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                    data-disabled={basicFieldProps.disabled}
                    data-readonly={basicFieldProps.readonly}
                    data-error={!!basicFieldProps.error}
                    checked={_value}
                    onCheckedChange={handleOnValueChange}
                    defaultChecked={defaultValue}
                    {...rest}
                >
                    <SwitchPrimitive.Thumb className={cn(SwitchAnatomy.thumb({ size }), thumbClass)} />
                </SwitchPrimitive.Root>
                <div className="flex flex-1"></div>
                {!!label && <div className="flex items-center gap-1">
                    <label
                        className={cn(
                            SwitchAnatomy.label(),
                            labelClass,
                            side === "right" && "font-semibold transition-transform __group-hover/switch:-translate-y-0.5",
                        )}
                        htmlFor={basicFieldProps.id}
                        data-disabled={basicFieldProps.disabled}
                    >
                        {label}
                    </label>
                    {moreHelp && <Popover
                        className="text-sm"
                        trigger={<AiOutlineExclamationCircle className="transition-opacity opacity-45 hover:opacity-90" />}
                    >
                        {moreHelp}
                    </Popover>}
                </div>}

                <input
                    ref={inputRef}
                    type="checkbox"
                    name={basicFieldProps.name}
                    className={hiddenInputStyles}
                    value={_value ? "on" : "off"}
                    checked={basicFieldProps.required ? _value : true}
                    aria-hidden="true"
                    required={controlledValue === undefined && basicFieldProps.required}
                    tabIndex={-1}
                    onChange={() => {}}
                    onFocusCapture={() => buttonRef.current?.focus()}
                />
            </div>
        </BasicField>
    )

})

Switch.displayName = "Switch"
