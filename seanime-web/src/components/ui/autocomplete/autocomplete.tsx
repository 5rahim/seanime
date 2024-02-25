"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList, CommandProps } from "../command"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { mergeRefs } from "../core/utils"
import { extractInputPartProps, hiddenInputStyles, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"
import { Popover } from "../popover"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AutocompleteAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Autocomplete__root",
    ]),
    popover: cva([
        "UI-Autocomplete__popover",
        "w-[--radix-popover-trigger-width] p-0",
    ]),
    checkIcon: cva([
        "UI-Autocomplete__checkIcon",
        "h-4 w-4",
        "data-[selected=true]:opacity-100 data-[selected=false]:opacity-0",
    ]),
    container: cva([
        "UI-Autocomplete__container",
        "relative w-full",
    ]),
    command: cva([
        "UI-Autocomplete__command",
        "focus-within:ring-2 ring-[--ring] transition",
    ]),
})


/* -------------------------------------------------------------------------------------------------
 * Autocomplete
 * -----------------------------------------------------------------------------------------------*/

export type AutocompleteOption = { value: string | null, label: string }

export type AutocompleteProps = Omit<React.ComponentPropsWithRef<"input">, "size" | "value" | "defaultValue"> &
    BasicFieldOptions &
    InputStyling &
    ComponentAnatomy<typeof AutocompleteAnatomy> & {
    /**
     * The selected option
     */
    value?: AutocompleteOption | undefined
    /**
     * Callback invoked when the value changes.
     */
    onValueChange?: (value: { value: string | null, label: string } | undefined) => void
    /**
     * Callback invoked when the input text changes.
     */
    onTextChange?: (value: string) => void
    /**
     * The autocompletion options.
     */
    options: AutocompleteOption[]
    /**
     * The message to display when there are no options.
     *
     * If not provided, the options list will be hidden when there are no options.
     */
    emptyMessage?: React.ReactNode
    /**
     * The placeholder of the input.
     */
    placeholder?: string
    /**
     * Additional props to pass to the command component.
     */
    commandProps?: CommandProps
    /**
     * Default value of the input when uncontrolled.
     */
    defaultValue?: AutocompleteOption
    /**
     * If true, the options list will be filtered based on the input value.
     * Set this to false if you want to filter the options yourself by listening to the `onTextChange` event.
     *
     * @default true
     */
    autoFilter?: boolean
    /**
     * If true, a loading indicator will be displayed.
     */
    isFetching?: boolean
    /**
     * The type of the autocomplete.
     *
     * - `custom`: Arbitrary values are allowed
     * - `options`: Only values from the options list are allowed. Falls back to last valid option if the input value is not in the options list.
     *
     * @default "custom"
     */
    type?: "custom" | "options"
}

export const Autocomplete = React.forwardRef<HTMLInputElement, AutocompleteProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<AutocompleteProps>(props, React.useId())

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
        containerClass,
        commandClass,
        /**/
        commandProps,
        options,
        emptyMessage,
        placeholder,
        value: controlledValue,
        onValueChange,
        onTextChange,
        onChange,
        defaultValue,
        autoFilter = true,
        isFetching,
        type,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<AutocompleteProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const isFirst = React.useRef(true)
    const isUpdating = React.useRef(false)

    const inputValueRef = React.useRef<string>(controlledValue?.label || defaultValue?.label || "")
    const [inputValue, setInputValue] = React.useState<string>(controlledValue?.label || defaultValue?.label || "")
    const deferredInputValue = React.useDeferredValue(inputValue)
    inputValueRef.current = inputValue

    const optionsTypeValueRef = React.useRef<AutocompleteOption | undefined>(controlledValue || defaultValue || undefined)
    const [value, setValue] = React.useState<AutocompleteOption | undefined>(controlledValue || defaultValue || undefined)

    const [open, setOpen] = React.useState(false)

    const filteredOptions = React.useMemo(() => {
        if (autoFilter) {
            return options.filter(option => option.label.toLowerCase().includes(deferredInputValue.toLowerCase()))
        }
        return options
    }, [autoFilter, options, deferredInputValue])

    // The options list should open when there are options or when there is an empty message
    const _optionListShouldOpen = !!emptyMessage || (options.length > 0 && filteredOptions.length > 0)

    // Function used to compare two labels
    const by = React.useCallback((a: string, b: string) => a.toLowerCase() === b.toLowerCase(), [])

    const inputRef = React.useRef<HTMLInputElement>(null)
    const commandInputRef = React.useRef<HTMLInputElement>(null)

    // Update the input value when the controlled value changes
    // Only when the default value is empty or when it is an updated value
    React.useEffect(() => {
        if (isUpdating.current) return
        if (!defaultValue || !isFirst.current) {
            setInputValue(controlledValue?.label ?? "")
            setValue(controlledValue)
            _updateOptionsTypeValueRef(controlledValue)
        }
        isFirst.current = false
    }, [controlledValue])

    const handleOnOpenChange = React.useCallback((opening: boolean) => {
        // If the input is disabled or readonly, do not open the popover
        if (basicFieldProps.disabled || basicFieldProps.readonly) return
        // If there are no options and the popover is opening, do not open it
        if (options.length === 0 && opening) return
        // If the input value has not and there are no filtered options, do not open the popover
        // This is to avoid a visual glitch when the popover opens but is empty
        if (inputValueRef.current === inputValue && opening && filteredOptions.length === 0) return

        setOpen(opening)
        if (!opening) {
            React.startTransition(() => {
                inputRef.current?.focus()
            })
        }
    }, [options, inputValue, basicFieldProps.disabled, basicFieldProps.readonly])

    const handleOnTextInputChange = React.useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        isUpdating.current = true
        onChange?.(e) // Emit the change event
        setInputValue(e.target.value) // Update the input value

        // Open the popover if there are filtered options
        if (autoFilter && filteredOptions.length > 0) {
            setOpen(true)
        }
    }, [filteredOptions])

    React.useEffect(() => {
        const v = deferredInputValue

        const _option = options.find(n => by(n.label, v))
        if (_option) {
            handleUpdateValue(_option)
        } else if (v.length > 0) {
            handleUpdateValue({ value: null, label: v })
        } else if (v.length === 0) {
            handleUpdateValue(undefined)
        }

        isUpdating.current = false
    }, [deferredInputValue, autoFilter])

    // Called when an option is selected either by clicking on it or entering a valid value
    const handleUpdateValue = React.useCallback((value: AutocompleteOption | undefined) => {
        setValue(value)
        onValueChange?.(value)
        onTextChange?.(value?.label ?? "")
        _updateOptionsTypeValueRef(value)
    }, [])

    // Focus the command input when arrow down is pressed
    const handleKeyDown = React.useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
        if (!open) {
            setOpen(true)
        }
        if (e.key === "ArrowDown") {
            e.preventDefault()
            commandInputRef.current?.focus()
        }
    }, [open])

    // Conditionally update the options type value ref when it is valid
    const _updateOptionsTypeValueRef = React.useCallback((value: AutocompleteOption | undefined) => {
        if (!!value?.value || value === undefined) {
            optionsTypeValueRef.current = value
        }
    }, [])

    // If the type is `options`, make sure the value is always a valid option
    // If the value entered doesn't match any option, fallback to the last valid option
    const handleOptionsTypeOnBlur = React.useCallback(() => {
        if (type === "options") {
            React.startTransition(() => {
                if (optionsTypeValueRef.current) {
                    setInputValue(optionsTypeValueRef.current.label)
                } else {
                    setInputValue("")
                }
            })
        }
    }, [])

    return (
        <BasicField {...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <Popover
                    open={open && _optionListShouldOpen}
                    onOpenChange={handleOnOpenChange}
                    className={cn(
                        AutocompleteAnatomy.popover(),
                        popoverClass,
                    )}
                    onOpenAutoFocus={e => e.preventDefault()}
                    trigger={
                        <div className={cn(AutocompleteAnatomy.container(), containerClass)}>
                            <input
                                ref={mergeRefs([inputRef, ref])}
                                id={basicFieldProps.id}
                                name={basicFieldProps.name}
                                value={inputValue}
                                onChange={handleOnTextInputChange}
                                onBlur={handleOptionsTypeOnBlur}
                                placeholder={placeholder}
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
                                    AutocompleteAnatomy.root(),
                                )}
                                disabled={basicFieldProps.disabled || basicFieldProps.readonly}
                                data-disabled={basicFieldProps.disabled}
                                data-error={!!basicFieldProps.error}
                                aria-readonly={basicFieldProps.readonly}
                                data-readonly={basicFieldProps.readonly}
                                onKeyDown={handleKeyDown}
                                required={basicFieldProps.required}
                                {...rest}
                            />
                        </div>
                    }
                >
                    <Command
                        className={cn(AutocompleteAnatomy.command(), commandClass)}
                        inputContainerClass="py-1"
                        shouldFilter={autoFilter}
                        {...commandProps}
                    >
                        {isFetching && inputValue.length > 0 && <div className="w-full absolute top-0 left-0 px-1">
                            <div className="h-1 w-full bg-[--subtle] overflow-hidden relative rounded-full">
                                <div className="animate-indeterminate-progress absolute left-0 w-full h-full bg-brand origin-left-right"></div>
                            </div>
                        </div>}
                        <CommandInput
                            value={inputValue}
                            onValueChange={setInputValue}
                            inputContainerClass={hiddenInputStyles}
                            aria-hidden="true"
                            ref={commandInputRef}
                        />
                        <CommandList>
                            {!!emptyMessage && (
                                <CommandEmpty>{emptyMessage}</CommandEmpty>
                            )}
                            <CommandGroup>
                                {options.map(option => (
                                    <CommandItem
                                        key={option.value}
                                        value={option.label}
                                        onSelect={(currentValue) => {
                                            const _option = options.find(n => by(n.label, currentValue))
                                            if (_option) {
                                                if (value?.value === _option.value) {
                                                    handleUpdateValue(undefined)
                                                    setInputValue("")
                                                } else {
                                                    handleUpdateValue(_option)
                                                    setInputValue(_option.label)
                                                }
                                            }
                                            React.startTransition(() => {
                                                inputRef.current?.focus()
                                            })
                                        }}
                                        leftIcon={
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                strokeWidth="2"
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                className={cn(
                                                    AutocompleteAnatomy.checkIcon(),
                                                    checkIconClass,
                                                )}
                                                data-selected={by(option.label, inputValue)}
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

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )
})

Autocomplete.displayName = "Autocomplete"
