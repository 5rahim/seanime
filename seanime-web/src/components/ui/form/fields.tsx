"use client"

import { DirectorySelector, DirectorySelectorProps } from "@/components/shared/directory-selector"
import { MediaExclusionSelector, MediaExclusionSelectorProps } from "@/components/shared/media-exclusion-selector"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { useDebounce } from "@/hooks/use-debounce"
import { colord } from "colord"
import React, { forwardRef, useMemo } from "react"
import { HexColorPicker } from "react-colorful"
import { Controller, FormState, get, useController, useFormContext } from "react-hook-form"
import { BiPlus, BiTrash } from "react-icons/bi"
import { useUpdateEffect } from "react-use"
import { Autocomplete, AutocompleteProps } from "../autocomplete"
import { BasicFieldOptions } from "../basic-field"
import { Checkbox, CheckboxGroup, CheckboxGroupProps, CheckboxProps } from "../checkbox"
import { Combobox, ComboboxProps } from "../combobox"
import { CurrencyInput, CurrencyInputProps } from "../currency-input"
import { DatePicker, DatePickerProps, DateRangePicker, DateRangePickerProps } from "../date-picker"
import { NativeSelect, NativeSelectProps } from "../native-select"
import { NumberInput, NumberInputProps } from "../number-input"
import { Popover } from "../popover"
import { RadioGroup, RadioGroupProps } from "../radio-group"
import { Select, SelectProps } from "../select"
import { SimpleDropzone, SimpleDropzoneProps } from "../simple-dropzone"
import { Switch, SwitchProps } from "../switch"
import { TextInput, TextInputProps } from "../text-input"
import { Textarea, TextareaProps } from "../textarea"
import { useFormSchema } from "./form"
import { createPolymorphicComponent } from "./polymorphic-component"
import { SubmitField } from "./submit-field"


/**
 * Add the BasicField types to any Field
 */
export type FieldBaseProps = Omit<BasicFieldOptions, "name"> & {
    name: string
    onChange?: any
    onBlur?: any
    required?: boolean
}

export type FieldComponent<T> = T & FieldBaseProps

export type FieldProps = React.ComponentPropsWithRef<"div">

/**
 * @description This wrapper makes it easier to work with custom form components by controlling their state.
 * @example
 * // Props order
 * <Controller>
 *    <InputComponent
 *       defaultValue={}   // Can be overridden
 *       onChange={}       // Can be overridden
 *       onBlur={}         // Can be overridden
 *       {...props}        // <FieldComponent {...} /> -> <Field.Component {...} />
 *       error={}          // Cannot be overridden
 *    />
 * </Controller>
 * @param InputComponent
 */
export function withControlledInput<T extends FieldBaseProps>(InputComponent: React.FC<T>) {
    return forwardRef<FieldProps, T>(
        (inputProps, ref) => {
            const { control, formState, ...context } = useFormContext()
            const { shape } = useFormSchema()

            /* Get the `required` status from the Schema */
            const required = useMemo(() => {
                return !!get(shape, inputProps.name) &&
                    !get(shape, inputProps.name)?.isOptional() &&
                    !get(shape, inputProps.name)?.isNullable()
            }, [shape])

            return (
                <Controller
                    name={inputProps.name}
                    control={control}
                    rules={{ required: inputProps.required }}
                    render={({ field: { ref: _ref, ...field } }) => (
                        /**
                         * We pass "value, onChange, onBlur, error, required" to all components that will be defined using the wrapper.
                         * For other components like "Switch" and "Checkbox" which do not use the "value" prop, you need to deconstruct it to avoid it
                         * being passed.
                         */
                        <InputComponent
                            value={field.value} // Default prop, can be overridden in Field component definition
                            onChange={callAllHandlers(inputProps.onChange, field.onChange)} // Default prop, can be overridden in Field component
                            onBlur={callAllHandlers(inputProps.onBlur, field.onBlur)} // Default prop, can be overridden in Field component
                            // required={required}
                            {...inputProps as any} // Props passed in <FieldComponent /> then props passed in <Field.Component />
                            // The props below will not be overridden.
                            // e.g: <Field.ComponentField error="Error" /> will not work
                            error={getFormError(field.name, formState)?.message}
                            ref={useMergeRefs(ref, _ref)}
                        />
                    )}
                />
            )
        },
    )
}

const withUncontrolledInput = <T extends FieldBaseProps>(InputComponent: React.FC<T>) => {
    return forwardRef<HTMLInputElement, T>(
        (props, ref) => {
            const { register, formState } = useFormContext()
            const { ref: _ref, ...field } = register(props.name)

            return (
                <InputComponent
                    {...props as any}
                    onChange={callAllHandlers(props.onChange, field.onChange)}
                    onBlur={callAllHandlers(props.onBlur, field.onBlur)}
                    error={getFormError(props.name, formState)?.message}
                    name={field.name}
                    ref={useMergeRefs(ref, _ref)}
                />
            )
        },
    )
}


const TextInputField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<TextInputProps>>(
    (props, ref) => {
        return <TextInput
            {...props}
            value={props.value ?? ""}
            ref={ref}
        />
    },
)))

const TextareaField = React.memo(withControlledInput(forwardRef<HTMLTextAreaElement, FieldComponent<TextareaProps>>(
    (props, ref) => {
        return <Textarea
            {...props}
            value={props.value ?? ""}
            ref={ref}
        />
    },
)))

const DatePickerField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<DatePickerProps>>((
    { onChange, ...props }, ref) => {

    return <DatePicker
        {...props}
        onValueChange={onChange}
        ref={ref}
    />
})))

const DateRangePickerField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<DateRangePickerProps>>((
    { onChange, ...props }, ref) => {

    return <DateRangePicker
        {...props}
        onValueChange={onChange}
        ref={ref}
    />
})))


const NativeSelectField = React.memo(withControlledInput(forwardRef<HTMLSelectElement, FieldComponent<NativeSelectProps>>(
    (props, ref) => {
        const context = useFormContext()
        const controller = useController({ name: props.name })

        // Set the default value as the first option if no default value is passed and there is no placeholder
        React.useEffect(() => {
            if (!get(context.formState.defaultValues, props.name) && !controller.field.value && !props.placeholder) {
                controller.field.onChange(props.options?.[0]?.value)
            }
        }, [])

        return <NativeSelect
            {...props}
            ref={ref}
        />
    },
)))

const ColorPickerField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<TextInputProps>>(
    (props, ref) => {
        const context = useFormContext()
        const controller = useController({ name: props.name })
        const validColorRef = React.useRef("#000")

        const [value, setValue] = React.useState(get(context.formState.defaultValues, props.name) || "#000")
        const deferredValue = useDebounce(value, 200)

        const valueRef = React.useRef(value)

        React.useEffect(() => {
            controller.field.onChange(deferredValue)
            if (colord(deferredValue).isValid()) {
                validColorRef.current = deferredValue
            }
        }, [deferredValue])

        const handleColorChange = React.useCallback(() => {
            if (!colord(value).isValid()) {
                setValue(validColorRef.current)
            } else {
                setValue(value)
            }
            valueRef.current = value
        }, [validColorRef.current, value])


        useUpdateEffect(() => {
            if (controller.field.value !== valueRef.current) {
                setValue(controller.field.value)
                valueRef.current = controller.field.value
            }
        }, [controller.field.value])

        return <TextInput
                {...props}
                ref={ref}
                value={value}
                onValueChange={setValue}
                onBlur={handleColorChange}
                rightAddon={<Popover
                    className="flex justify-center"
                    trigger={<div className="cursor-pointer size-7 rounded-[--radius-md]" style={{ backgroundColor: value }} />}
                >
                    <HexColorPicker
                        color={value}
                        onChange={color => {
                            setValue(color)
                        }}
                    />
                </Popover>}
                className="w-full"
            />
    },
)))

const SelectField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<SelectProps>>(
    ({ onChange, ...props }, ref) => {
        return <Select
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const NumberField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<NumberInputProps>>(
    ({ onChange, ...props }, ref) => {
        return <NumberInput
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))


const ComboboxField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<ComboboxProps>>(
    ({ onChange, ...props }, ref) => {
        return <Combobox
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const SwitchField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<SwitchProps>>(
    ({ onChange, ...props }, ref) => {
        return <Switch
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const CheckboxField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<CheckboxProps>>(
    ({ onChange, ...props }, ref) => {
        return <Checkbox
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const CheckboxGroupField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<CheckboxGroupProps>>(
    ({ onChange, ...props }, ref) => {
        return <CheckboxGroup
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))


const RadioGroupField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<RadioGroupProps>>(
    ({ onChange, ...props }, ref) => {
        return <RadioGroup
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))


const RadioCardsField = React.memo(withControlledInput(forwardRef<HTMLButtonElement, FieldComponent<RadioGroupProps>>(
    ({ onChange, ...props }, ref) => {
        return <RadioGroup
            // itemContainerClass={cn(
            //     "items-start cursor-pointer transition border-transparent rounded-[--radius] p-4 w-full",
            //     "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-900",
            //     "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
            //     "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
            //     "border border-transparent data-[state=checked]:border-[--brand] data-[state=checked]:ring-offset-0",
            // )}
            // itemClass={cn(
            //     "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
            //     "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent
            // dark:data-[state=unchecked]:hover:bg-transparent", "focus-visible:ring-0 focus-visible:ring-offset-0
            // focus-visible:ring-offset-transparent", )} itemIndicatorClass="hidden" itemLabelClass="font-medium flex flex-col items-center
            // data-[state=checked]:text-[--brand] cursor-pointer"
            itemContainerClass={cn(
                "items-start cursor-pointer transition border-transparent rounded-[--radius] p-3 w-full md:w-fit",
                "bg-transparent dark:hover:bg-gray-900 dark:bg-transparent",
                "data-[state=checked]:bg-brand-500/5 dark:data-[state=checked]:bg-gray-900",
                "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-transparent transition",
                "dark:border dark:data-[state=checked]:border-[--border] data-[state=checked]:ring-offset-0",
            )}
            itemClass={cn(
                "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
                "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
                "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
            )}
            itemIndicatorClass="hidden"
            itemLabelClass="font-medium flex flex-col items-center data-[state=unchecked]:hover:text-[--foreground] data-[state=checked]:text-[--brand] text-[--muted] cursor-pointer"
            // stackClass="flex flex-col md:flex-row flex-wrap gap-2 space-y-0"
            {...props}
            onValueChange={onChange}
            stackClass="flex flex-col md:flex-row gap-2 space-y-0"
            ref={ref}
        />
    },
)))


const CurrencyInputField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<CurrencyInputProps>>(
    ({ onChange, ...props }, ref) => {
        return <CurrencyInput
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const AutocompleteField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<AutocompleteProps>>(
    ({ onChange, ...props }, ref) => {
        return <Autocomplete
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

const SimpleDropzoneField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<SimpleDropzoneProps>>(
    ({ onChange, value, ...props }, ref) => {

        const controller = useController({ name: props.name })

        // Set the default value to an empty array
        React.useEffect(() => {
            controller.field.onChange([])
        }, [])

        return <SimpleDropzone
            {...props}
            onValueChange={onChange}
            ref={ref}
        />
    },
)))

type DirectorySelectorFieldProps = Omit<DirectorySelectorProps, "onSelect" | "value"> & { value?: string }

const DirectorySelectorField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<DirectorySelectorFieldProps>>(
    ({ value, onChange, shouldExist, ...props }, ref) => {
        const context = useFormContext()
        const controller = useController({ name: props.name })

        const defaultValue = useMemo(() => get(context.formState.defaultValues, props.name) ?? "", [])

        React.useEffect(() => {
            controller.field.onChange(defaultValue)
        }, [])

        return <DirectorySelector
            shouldExist={shouldExist}
            {...props}
            value={value ?? ""}
            defaultValue={defaultValue}
            onSelect={value => controller.field.onChange(value)}
            ref={ref}
        />
    },
)))

type MultiDirectorySelectorFieldProps = Omit<DirectorySelectorProps, "onSelect" | "value"> & { value?: string[] }

const MultiDirectorySelectorField = React.memo(withControlledInput(forwardRef<HTMLInputElement, FieldComponent<MultiDirectorySelectorFieldProps>>(
    ({ value, onChange, shouldExist, label, help, ...props }, ref) => {
        const context = useFormContext()
        const controller = useController({ name: props.name })

        const [paths, setPaths] = React.useState<string[]>([])

        const defaultValue = useMemo(() => get(context.formState.defaultValues, props.name) ?? [], [])
        React.useEffect(() => {
            setPaths(defaultValue)
        }, [])


        React.useEffect(() => {
            controller.field.onChange(paths.filter(p => p))
        }, [paths])

        return <div className="space-y-2">
            <div>
                {label && <label className="block text-md font-medium">{label}</label>}
                {help && <p className="text-sm text-[--muted]">{help}</p>}
            </div>
            {paths?.map((v, i) => (
                <div className="flex items-center gap-2" key={i}>
                    <div className="w-full">
                        <DirectorySelector
                            shouldExist={shouldExist}
                            {...props}
                            label="Directory"
                            value={v ?? ""}
                            defaultValue={v ?? ""}
                            onSelect={value => {
                                setPaths(prev => {
                                    const newPaths = [...prev]
                                    newPaths[i] = value
                                    return newPaths
                                })
                            }}
                            ref={ref}
                            fieldClass="w-full"
                        />
                    </div>
                    <IconButton
                        rounded
                        size="sm"
                        intent="alert-subtle"
                        icon={<BiTrash />}
                        onClick={() => setPaths(prev => prev.filter((_, index) => index !== i))}
                    />
                </div>
            ))}
            <IconButton
                rounded
                size="sm"
                intent="gray-subtle"
                icon={<BiPlus />}
                onClick={() => setPaths(prev => [...prev, ""])}
            />
        </div>
    },
)))

const MediaExclusionSelectorField = React.memo(withControlledInput(forwardRef<HTMLDivElement, FieldComponent<MediaExclusionSelectorProps>>(
    ({ onChange, ...props }, ref) => {
        return <MediaExclusionSelector
            {...props}
            onChange={onChange}
            ref={ref}
        />
    },
)))

export const Field = createPolymorphicComponent<"div", FieldProps, {
    Text: typeof TextInputField,
    Textarea: typeof TextareaField,
    Select: typeof SelectField,
    NativeSelect: typeof NativeSelectField,
    Switch: typeof SwitchField,
    Checkbox: typeof CheckboxField,
    CheckboxGroup: typeof CheckboxGroupField,
    RadioGroup: typeof RadioGroupField,
    Currency: typeof CurrencyInputField,
    Number: typeof NumberField,
    DatePicker: typeof DatePickerField
    DateRangePicker: typeof DateRangePickerField
    Combobox: typeof ComboboxField
    Autocomplete: typeof AutocompleteField
    SimpleDropzone: typeof SimpleDropzoneField
    DirectorySelector: typeof DirectorySelectorField
    MultiDirectorySelector: typeof MultiDirectorySelectorField
    RadioCards: typeof RadioCardsField
    ColorPicker: typeof ColorPickerField
    MediaExclusionSelector: typeof MediaExclusionSelectorField
    Submit: typeof SubmitField
}>({
    Text: TextInputField,
    Textarea: TextareaField,
    Select: SelectField,
    NativeSelect: NativeSelectField,
    Switch: SwitchField,
    Checkbox: CheckboxField,
    CheckboxGroup: CheckboxGroupField,
    RadioGroup: RadioGroupField,
    Currency: CurrencyInputField,
    Number: NumberField,
    DatePicker: DatePickerField,
    DateRangePicker: DateRangePickerField,
    Combobox: ComboboxField,
    Autocomplete: AutocompleteField,
    SimpleDropzone: SimpleDropzoneField,
    DirectorySelector: DirectorySelectorField,
    MultiDirectorySelector: MultiDirectorySelectorField,
    ColorPicker: ColorPickerField,
    RadioCards: RadioCardsField,
    MediaExclusionSelector: MediaExclusionSelectorField,
    Submit: SubmitField,
})

Field.displayName = "Field"

/* -------------------------------------------------------------------------------------------------
 * Utils
 * -----------------------------------------------------------------------------------------------*/

export const getFormError = (name: string, formState: FormState<{ [x: string]: any }>) => {
    return get(formState.errors, name)
}

export type ReactRef<T> = React.RefCallback<T> | React.MutableRefObject<T>

export function assignRef<T = any>(
    ref: ReactRef<T> | null | undefined,
    value: T,
) {
    if (ref == null) return

    if (typeof ref === "function") {
        ref(value)
        return
    }

    try {
        ref.current = value
    }
    catch (error) {
        throw new Error(`Cannot assign value '${value}' to ref '${ref}'`)
    }
}

export function mergeRefs<T>(...refs: (ReactRef<T> | null | undefined)[]) {
    return (node: T | null) => {
        refs.forEach((ref) => {
            assignRef(ref, node)
        })
    }
}

export function useMergeRefs<T>(...refs: (ReactRef<T> | null | undefined)[]) {
    return useMemo(() => mergeRefs(...refs), refs)
}

type Args<T extends Function> = T extends (...args: infer R) => any ? R : never

function callAllHandlers<T extends (event: any) => void>(
    ...fns: (T | undefined)[]
) {
    return function func(event: Args<T>[0]) {
        fns.some((fn) => {
            fn?.(event)
            return event?.defaultPrevented
        })
    }
}
