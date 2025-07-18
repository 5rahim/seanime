"use client"

import * as SelectPrimitive from "@radix-ui/react-select"
import { SelectItem, SelectItemIndicator, SelectItemText } from "@radix-ui/react-select"
import { cva } from "class-variance-authority"
import * as React from "react"
import { BasicField, BasicFieldOptions, extractBasicFieldProps } from "../basic-field"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { mergeRefs } from "../core/utils"
import { extractInputPartProps, hiddenInputStyles, InputAddon, InputAnatomy, InputContainer, InputIcon, InputStyling } from "../input"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SelectAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Select__root",
        "inline-flex items-center justify-between relative whitespace-nowrap truncate",
    ]),
    chevronIcon: cva([
        "UI-Combobox__chevronIcon",
        "ml-2 h-4 w-4 shrink-0 opacity-50",
    ]),
    scrollButton: cva([
        "UI-Select__scrollButton",
        "flex items-center justify-center h-[25px] bg-[--paper] text-base cursor-default",
    ]),
    content: cva([
        "UI-Select__content",
        "w-full overflow-hidden rounded-[--radius] shadow-md bg-[--paper] border leading-none z-50",
    ]),
    viewport: cva([
        "UI-Select__viewport",
        "p-1 z-10",
    ]),
    item: cva([
        "UI-Select__item",
        "text-base leading-none rounded-[--radius] flex items-center h-8 pr-2 pl-8 relative",
        "select-none disabled:opacity-50 disabled:pointer-events-none",
        "data-highlighted:outline-none data-highlighted:bg-[--subtle]",
        "data-[disabled=true]:opacity-50 data-[disabled=true]:pointer-events-none",
    ]),
    checkIcon: cva([
        "UI-Select__checkIcon",
        "absolute left-2 w-4 inline-flex items-center justify-center",
    ]),
})


/* -------------------------------------------------------------------------------------------------
 * Select
 * -----------------------------------------------------------------------------------------------*/

export type SelectOption = { value: string, label?: string, disabled?: boolean }

export type SelectProps = InputStyling &
    BasicFieldOptions &
    Omit<React.ComponentPropsWithoutRef<"button">, "value" | "defaultValue"> &
    ComponentAnatomy<typeof SelectAnatomy> & {
    /**
     * The options to display in the dropdown
     */
    options: SelectOption[] | undefined
    /**
     * The placeholder text
     */
    placeholder?: string
    /**
     * Direction of the text
     */
    dir?: "ltr" | "rtl"
    /**
     * The selected value
     */
    value?: string | undefined
    /**
     * Callback fired when the selected value changes
     */
    onValueChange?: (value: string) => void
    /**
     * Callback fired when the dropdown opens or closes
     */
    onOpenChange?: (open: boolean) => void
    /**
     * Default selected value when uncontrolled
     */
    defaultValue?: string
    /**
     * Ref to the input element
     */
    inputRef?: React.Ref<HTMLSelectElement>
}

export const Select = React.forwardRef<HTMLButtonElement, SelectProps>((props, ref) => {

    const [props1, basicFieldProps] = extractBasicFieldProps<SelectProps>(props, React.useId())

    const [{
        size,
        intent,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        /**/
        className,
        placeholder,
        options,
        chevronIconClass,
        scrollButtonClass,
        contentClass,
        viewportClass,
        checkIconClass,
        itemClass,
        /**/
        dir,
        value: controlledValue,
        onValueChange,
        onOpenChange,
        defaultValue,
        inputRef,
        ...rest
    }, {
        inputContainerProps,
        leftAddonProps,
        leftIconProps,
        rightAddonProps,
        rightIconProps,
    }] = extractInputPartProps<SelectProps>({
        ...props1,
        size: props1.size ?? "md",
        intent: props1.intent ?? "basic",
        leftAddon: props1.leftAddon,
        leftIcon: props1.leftIcon,
        rightAddon: props1.rightAddon,
        rightIcon: props1.rightIcon,
    })

    const isFirst = React.useRef(true)

    const buttonRef = React.useRef<HTMLButtonElement>(null)

    const [_value, _setValue] = React.useState<string | undefined>(controlledValue ?? defaultValue)

    const handleOnValueChange = React.useCallback((value: string) => {
        if (value === "__placeholder__") {
            _setValue("")
            onValueChange?.("")
            return
        }
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
        <BasicField {...basicFieldProps}>
            <InputContainer {...inputContainerProps}>
                <InputAddon {...leftAddonProps} />
                <InputIcon {...leftIconProps} />

                <SelectPrimitive.Root
                    dir={dir}
                    value={_value}
                    onValueChange={handleOnValueChange}
                    onOpenChange={onOpenChange}
                    defaultValue={defaultValue}
                >

                    <SelectPrimitive.Trigger
                        ref={mergeRefs([buttonRef, ref])}
                        id={basicFieldProps.id}
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
                            SelectAnatomy.root(),
                            className,
                        )}
                        aria-label={basicFieldProps.name || "Select"}
                        {...rest}
                    >
                        <SelectPrimitive.Value placeholder={placeholder} />

                        <SelectPrimitive.Icon className={cn(!!rightIcon && "hidden")}>
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                className={cn(SelectAnatomy.chevronIcon(), chevronIconClass)}
                            >
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                        </SelectPrimitive.Icon>

                    </SelectPrimitive.Trigger>

                    <SelectPrimitive.Portal>
                        <SelectPrimitive.Content className={cn(SelectAnatomy.content(), contentClass)}>

                            <SelectPrimitive.ScrollUpButton className={cn(SelectAnatomy.scrollButton(), scrollButtonClass)}>
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    className={cn(SelectAnatomy.chevronIcon(), "rotate-180", chevronIconClass)}
                                >
                                    <path d="m6 9 6 6 6-6" />
                                </svg>
                            </SelectPrimitive.ScrollUpButton>

                            <SelectPrimitive.Viewport className={cn(SelectAnatomy.viewport(), viewportClass)}>

                                {(!!placeholder && !basicFieldProps.required) && (
                                    <SelectItem
                                        className={cn(
                                            SelectAnatomy.item(),
                                            itemClass,
                                        )}
                                        value="__placeholder__"
                                    >
                                        <SelectItemText className="flex-none whitespace-nowrap truncate">{placeholder}</SelectItemText>
                                    </SelectItem>
                                )}

                                {options?.map(option => (
                                    <SelectItem
                                        key={option.value}
                                        className={cn(
                                            SelectAnatomy.item(),
                                            itemClass,
                                        )}
                                        value={option.value}
                                        disabled={option.disabled}
                                        data-disabled={option.disabled}
                                    >
                                        <SelectItemText className="flex-none whitespace-nowrap truncate">{option.label}</SelectItemText>
                                        <SelectItemIndicator asChild>
                                            <svg
                                                xmlns="http://www.w3.org/2000/svg"
                                                viewBox="0 0 24 24"
                                                fill="none"
                                                stroke="currentColor"
                                                strokeWidth="2"
                                                strokeLinecap="round"
                                                strokeLinejoin="round"
                                                className={cn(
                                                    SelectAnatomy.checkIcon(),
                                                    checkIconClass,
                                                )}
                                            >
                                                <path d="M20 6 9 17l-5-5" />
                                            </svg>
                                        </SelectItemIndicator>
                                    </SelectItem>
                                ))}

                            </SelectPrimitive.Viewport>

                            <SelectPrimitive.ScrollDownButton className={cn(SelectAnatomy.scrollButton(), scrollButtonClass)}>
                                <svg
                                    xmlns="http://www.w3.org/2000/svg"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth="2"
                                    strokeLinecap="round"
                                    strokeLinejoin="round"
                                    className={cn(SelectAnatomy.chevronIcon(), chevronIconClass)}
                                >
                                    <path d="m6 9 6 6 6-6" />
                                </svg>
                            </SelectPrimitive.ScrollDownButton>

                        </SelectPrimitive.Content>
                    </SelectPrimitive.Portal>

                </SelectPrimitive.Root>

                <select
                    ref={inputRef}
                    name={basicFieldProps.name}
                    className={hiddenInputStyles}
                    aria-hidden="true"
                    required={basicFieldProps.required}
                    disabled={basicFieldProps.disabled}
                    value={_value}
                    tabIndex={-1}
                    onChange={() => {}}
                    onFocusCapture={() => buttonRef.current?.focus()}
                >
                    <option value="" />
                    {options?.map(option => (
                        <option
                            key={option.value}
                            value={option.value}
                            disabled={option.disabled}
                        />
                    ))}
                </select>

                <InputAddon {...rightAddonProps} />
                <InputIcon {...rightIconProps} />
            </InputContainer>
        </BasicField>
    )

})

Select.displayName = "Select"
