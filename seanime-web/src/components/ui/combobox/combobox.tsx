"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { BiX } from "react-icons/bi"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandProps } from "../command"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { mergeRefs } from "../core/utils"
import { extractInputPartProps, hiddenInputStyles, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"
import { Popover } from "../popover"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ComboboxAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Combobox__root",
        "justify-between h-auto",
    ], {
        variants: {
            size: {
                sm: "min-h-8 px-2 py-1 text-sm",
                md: "min-h-10 px-3 py-2 ",
                lg: "min-h-12 px-4 py-3 text-md",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    popover: cva([
        "UI-Combobox__popover",
        "w-[--radix-popover-trigger-width] p-0",
    ]),
    checkIcon: cva([
        "UI-Combobox__checkIcon",
        "h-4 w-4",
        "data-[selected=true]:opacity-100 data-[selected=false]:opacity-0",
    ]),
    item: cva([
        "UI-Combobox__item",
        "flex gap-1 items-center flex-none truncate bg-gray-100 dark:bg-gray-800 px-2 pr-1 rounded-[--radius] max-w-96",
    ]),
    placeholder: cva([
        "UI-Combobox__placeholder",
        "text-[--muted] truncate",
    ]),
    inputValuesContainer: cva([
        "UI-Combobox__inputValuesContainer",
        "grow flex overflow-hidden gap-2 flex-wrap",
    ]),
    chevronIcon: cva([
        "UI-Combobox__chevronIcon",
        "ml-2 h-4 w-4 shrink-0 opacity-50",
    ]),
    removeItemButton: cva([
        "UI-Badge__removeItemButton",
        "text-lg cursor-pointer transition ease-in hover:opacity-60",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Combobox
 * -----------------------------------------------------------------------------------------------*/

export type ComboboxOption = { value: string, textValue?: string, label: React.ReactNode }

export type ComboboxProps = Omit<React.ComponentPropsWithRef<"button">, "size" | "value"> &
    BasicFieldOptions &
    InputStyling &
    ComponentAnatomy<typeof ComboboxAnatomy> & {
    /**
     * The selected values
     */
    value?: string[]
    /**
     * Callback fired when the selected values change
     */
    onValueChange?: (value: string[]) => void
    /**
     * Callback fired when the search input changes
     */
    onTextChange?: (value: string) => void
    /**
     * Additional props for the command component
     */
    commandProps?: CommandProps
    /**
     * The options to display in the dropdown
     */
    options: ComboboxOption[]
    /**
     * The message to display when there are no options
     */
    emptyMessage: React.ReactNode
    /**
     * The placeholder text
     */
    placeholder?: string
    /**
     * Allow multiple values to be selected
     */
    multiple?: boolean
    /**
     * Default value when uncontrolled
     */
    defaultValue?: string[]
    /**
     * Ref to the input element
     */
    inputRef?: React.Ref<HTMLInputElement>
}

export const Combobox = React.forwardRef<HTMLButtonElement, ComboboxProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<ComboboxProps>(props, React.useId())

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        className,
        popoverClass,
        checkIconClass,
        itemClass,
        placeholderClass,
        inputValuesContainerClass,
        chevronIconClass,
        removeItemButtonClass,
        /**/
        commandProps,
        options,
        emptyMessage,
        placeholder,
        value: controlledValue,
        onValueChange,
        onTextChange,
        multiple = false,
        defaultValue,
        inputRef,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<ComboboxProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const isControlled = controlledValue !== undefined
    const [internalValue, setInternalValue] = React.useState(defaultValue ?? [])
    const value = isControlled ? controlledValue! : internalValue

    const updateValue = React.useCallback((next: string[]) => {
        if (!isControlled) setInternalValue(next)
        onValueChange?.(next)
    }, [isControlled, onValueChange])

    const buttonRef = React.useRef<HTMLButtonElement>(null)
    const [open, setOpen] = React.useState(false)

    const selectedOptions = options.filter((option) => value.includes(option.value))

    const selectedValues = (
        (!!value.length && !!selectedOptions.length) ?
            multiple ? selectedOptions.map((option) => (
                <div key={option.value} className={cn(ComboboxAnatomy.item(), itemClass)}>
                    <span className="truncate">{option.textValue || option.value}</span>
                    <span
                        className={cn(ComboboxAnatomy.removeItemButton(), "rounded-full", removeItemButtonClass)}
                        onClick={(e) => {
                            e.preventDefault()
                            updateValue(value.filter((v) => v !== option.value))
                            if (!multiple) setOpen(false)
                        }}
                    >
                        <BiX />
                    </span>
                </div>
            )) : <span className="truncate">{selectedOptions[0].label}</span>
            : <span className={cn(ComboboxAnatomy.placeholder(), placeholderClass)}>{placeholder}</span>
    )

    return (
        <BasicField {...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <Popover
                    open={open}
                    onOpenChange={setOpen}
                    className={cn(ComboboxAnatomy.popover(), popoverClass)}
                    trigger={
                        <button
                            ref={mergeRefs([buttonRef, ref])}
                            id={basicFieldProps.id}
                            role="combobox"
                            aria-expanded={open}
                            className={cn(
                                InputAnatomy.root({
                                    size,
                                    intent,
                                    hasError: !!basicFieldProps.error,
                                    isDisabled: !!basicFieldProps.disabled,
                                    isReadonly: !!basicFieldProps.readonly,
                                    hasRightAddon: !!rightAddon,
                                    hasRightIcon: !!rightIcon,
                                    hasLeftAddon: !!leftAddon,
                                    hasLeftIcon: !!leftIcon,
                                }),
                                ComboboxAnatomy.root({ size }),
                                className,
                            )}
                            {...rest}
                        >
                            <div className={cn(ComboboxAnatomy.inputValuesContainer(), inputValuesContainerClass)}>
                                {selectedValues}
                            </div>

                            <div className="flex items-center">
                                {(!!value.length && !!selectedOptions.length && !multiple) && (
                                    <span
                                        className={cn(ComboboxAnatomy.removeItemButton(), removeItemButtonClass)}
                                        onClick={(e) => {
                                            e.preventDefault()
                                            updateValue([])
                                            setOpen(false)
                                        }}
                                    >
                                        ✕
                                    </span>
                                )}

                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 24 24"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    className={cn(ComboboxAnatomy.chevronIcon(), chevronIconClass)}
                                >
                                    <path d="m7 15 5 5 5-5" />
                                    <path d="m7 9 5-5 5 5" />
                                </svg>
                            </div>
                        </button>
                    }
                >
                    <Command inputContainerClass="py-1" {...commandProps}>
                        <CommandInput
                            placeholder={placeholder}
                            onValueChange={onTextChange}
                        />
                        <CommandList>
                            <CommandEmpty>{emptyMessage}</CommandEmpty>
                            <CommandGroup>
                                {options.map((option) => (
                                    <CommandItem
                                        key={option.value}
                                        value={option.textValue || option.value}
                                        onSelect={(currentValue) => {
                                            const _option = options.find(
                                                o => (o.textValue || o.value).toLowerCase() === currentValue.toLowerCase(),
                                            )
                                            if (_option) {
                                                const next = multiple
                                                    ? (!value.includes(_option.value)
                                                        ? [...value, _option.value]
                                                        : value.filter((v) => v !== _option.value))
                                                    : (value.includes(_option.value) ? [] : [_option.value])

                                                updateValue(next)
                                            }

                                            if (!multiple) setOpen(false)
                                        }}
                                        leftIcon={
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                viewBox="0 0 24 24"
                                                stroke="currentColor"
                                                strokeWidth="2"
                                                className={cn(ComboboxAnatomy.checkIcon(), checkIconClass)}
                                                data-selected={value.includes(option.value)}
                                            >
                                                <path d="M20 6 9 17l-5-5" />
                                            </svg>
                                        }
                                    >
                                        {option.label}
                                    </CommandItem>
                                ))}
                            </CommandGroup>
                        </CommandList>
                    </Command>
                </Popover>

                <input
                    ref={inputRef}
                    type="text"
                    name={basicFieldProps.name}
                    className={hiddenInputStyles}
                    value={basicFieldProps.required ? (!!value.length ? JSON.stringify(value) : "") : JSON.stringify(value)}
                    aria-hidden="true"
                    required={basicFieldProps.required}
                    tabIndex={-1}
                    onChange={() => {}}
                    onFocusCapture={() => buttonRef.current?.focus()}
                />

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )
})

Combobox.displayName = "Combobox"
