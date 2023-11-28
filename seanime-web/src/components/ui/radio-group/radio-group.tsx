"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import * as radio from "@zag-js/radio-group"
import { normalizeProps, useMachine } from "@zag-js/react"
import { cva, VariantProps } from "class-variance-authority"
import React, { useEffect, useId } from "react"
import { BasicField, BasicFieldAnatomy, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const RadioGroupAnatomy = defineStyleAnatomy({
    stack: cva("UI-RadioGroup__stack w-full space-y-1"),
    radioControl: cva([
        "UI-RadioGroup__radioControl",
        "inline-flex flex-none justify-center items-center border border-gray-300 rounded-full text-white bg-white cursor-pointer transition duration-10 relative",
        "data-[focus]:outline-none data-[focus]:ring-2 ring-offset-1 ring-[--ring]",
        "data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50 data-[disabled=true]:cursor-not-allowed",
        "bg-white border-gray-300 hover:bg-gray-100 hover:text-brand-100",
        "data-[checked=true]:bg-brand-500 data-[checked=true]:dark:bg-brand-500 data-[checked=true]:border-brand-500",
        "dark:bg-gray-700 dark:border-gray-700 dark:hover:bg-gray-700",
        "data-[error=true]:border-red-500",
        "peer-[.is-focused]:ring-2",
    ], {
        variants: {
            size: {
                md: "h-5 w-5 text-xs",
                lg: "h-6 w-6 text-sm",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    radioLabel: cva([
        "UI-RadioGroup__radioLabel font-normal flex-none",
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
    radioHelp: cva([
        "UI-RadioGroup__radioHelp",
    ]),
    radioContainer: cva([
        "UI-RadioGroup__radioContainer",
        "inline-flex w-full gap-2 items-center relative",
    ]),
    radioIcon: cva([
        "UI-RadioGroup__radioIcon",
        "data-[disabled=true]:opacity-50 data-[disabled=true]:cursor-not-allowed",
        "data-[checked=true]:text-white data-[checked=false]:text-transparent",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * RadioGroup
 * -----------------------------------------------------------------------------------------------*/

export interface RadioGroupProps extends BasicFieldOptions, ComponentWithAnatomy<typeof RadioGroupAnatomy>,
    VariantProps<typeof RadioGroupAnatomy.radioLabel> {
    value?: string
    defaultValue?: string
    options: { value: string, label?: React.ReactNode, help?: React.ReactNode }[]
    onChange?: (value: string | null) => void
    checkedIcon?: React.ReactNode
}

export const RadioGroup = React.forwardRef<HTMLDivElement, RadioGroupProps>((props, ref) => {

    const [{
        size = "md",
        value,
        defaultValue,
        onChange,
        options,
        stackClassName,
        radioLabelClassName,
        radioControlClassName,
        radioContainerClassName,
        radioHelpClassName,
        radioIconClassName,
        checkedIcon,
    }, basicFieldProps] = extractBasicFieldProps<RadioGroupProps>(props, useId())

    const [state, send] = useMachine(radio.machine({
        id: basicFieldProps.id,
        value: value,
        name: basicFieldProps.name,
        disabled: basicFieldProps.isDisabled,
        readOnly: basicFieldProps.isReadOnly,
        onChange(details) {
            onChange && onChange(details.value)
        },
    }))

    const api = radio.connect(state, send, normalizeProps)

    // Set default value
    useEffect(() => {
        if (!value && defaultValue) {
            api.setValue(defaultValue)
        }
    }, [])

    // Control the state
    useEffect(() => {
        (value && value.length > 0) && api.setValue(value)
    }, [value || "", options])

    return (
        <>
            <BasicField
                {...basicFieldProps}
            >
                <div className={cn(RadioGroupAnatomy.stack(), stackClassName)} {...api.rootProps} ref={ref}>

                    {options.map((opt) => (
                        <label
                            key={opt.value}
                            {...api.getRadioProps({ value: opt.value })}
                            className={cn(
                                RadioGroupAnatomy.radioContainer(),
                                radioContainerClassName,
                            )}
                            data-checked={api.value === opt.value}
                            tabIndex={-1}
                        >

                            <input {...api.getRadioInputProps({ value: opt.value })} />

                            <div
                                className={cn(RadioGroupAnatomy.radioControl({ size }), radioControlClassName)}
                                {...api.getRadioControlProps({ value: opt.value })}
                                data-error={!!basicFieldProps.error}
                                data-disabled={!!basicFieldProps.isDisabled}
                                data-checked={api.value === opt.value}
                            >
                                {checkedIcon ? checkedIcon :
                                    <span
                                        className={cn(RadioGroupAnatomy.radioIcon(), radioIconClassName)}
                                        data-disabled={!!basicFieldProps.isDisabled}
                                        data-checked={api.value === opt.value}
                                    >
                                        <svg
                                            xmlns="http://www.w3.org/2000/svg"
                                            viewBox="0 0 16 16"
                                            width="16"
                                            height="16"
                                            stroke="currentColor"
                                            fill="currentColor"
                                        >
                                            <path d="M8 4a4 4 0 1 1 0 8 4 4 0 0 1 0-8Z"></path>
                                        </svg>
                                    </span>}
                            </div>

                            <div
                                className={cn(
                                    RadioGroupAnatomy.radioLabel({ size }),
                                    radioLabelClassName,
                                )}
                                {...api.getRadioLabelProps({ value: opt.value })}
                                data-disabled={!!basicFieldProps.isDisabled}
                                data-checked={api.value === opt.value}
                            >
                                {opt.label ?? opt.value}
                            </div>

                            {!!opt.help && <div
                                className={cn(
                                    BasicFieldAnatomy.fieldHelpText(),
                                    radioHelpClassName,
                                )}
                                {...api.getRadioLabelProps({ value: opt.value })}
                                data-checked={api.value === opt.value}
                            >
                                {opt.help}
                            </div>}


                        </label>
                    ))}

                </div>
            </BasicField>
        </>
    )

})

RadioGroup.displayName = "RadioGroup"
