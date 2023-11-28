"use client"

import { Transition } from "@headlessui/react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import _filter from "lodash/filter"
import _find from "lodash/find"
import React, { Fragment, startTransition, useCallback, useEffect, useId, useMemo, useRef, useState } from "react"
import { Badge } from "../badge"
import { BasicField, extractBasicFieldProps } from "../basic-field"
import { InputAddon, InputAnatomy, inputContainerStyle, InputIcon, InputStyling } from "../input"
import type { TextInputProps } from "../text-input"
import { useUpdateEffect } from "react-use"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const MultiSelectAnatomy = defineStyleAnatomy({
    input: cva([
        "UI-MultiSelect__input",
        "relative flex flex-wrap gap-2 cursor-text p-2",
        "focus-within:ring-1 ring-[--ring] focus-within:border-brand-500 focus-within:hover:border-brand-500",
    ], {
        variants: {
            isOpen: {
                true: "",
                false: null,
            },
        },
    }),
    menuContainer: cva([
        "UI-MultiSelect__menuContainer",
        "absolute z-10 -bottom-2",
        "left-0 translate-y-full max-h-56 w-full overflow-auto rounded-[--radius] p-1 text-base shadow-lg sm:text-sm",
        "ring-1 ring-black ring-opacity-5 focus:outline-none",
        "bg-[--paper] border border-[--border]",
    ]),
    menuItem: cva([
        "UI-MultiSelect__menuItem",
        "relative cursor-pointer py-2 pl-3 pr-9 rounded-[--radius]",
    ], {
        variants: {
            highlighted: {
                true: "bg-[--highlight]",
                false: null,
            },
        },
    }),
    menuItemImage: cva([
        "UI-MultiSelect__menuItemImage",
        "flex-none justify-center w-8 h-8 mr-3 rounded-full overflow-hidden relative bg-slate-200",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * MultiSelect
 * -----------------------------------------------------------------------------------------------*/

export type MultiSelectOption = { value: string, label?: string, description?: string, image?: React.ReactNode }


export interface MultiSelectProps extends Omit<TextInputProps, "defaultValue" | "onChange">, InputStyling,
    ComponentWithAnatomy<typeof MultiSelectAnatomy> {
    options: MultiSelectOption[]
    value?: MultiSelectOption["value"][]
    defaultValue?: MultiSelectOption["value"][]
    onChange?: (values: MultiSelectOption["value"][]) => void
    isLoading?: boolean
    discrete?: boolean
    max?: number
    loadingContent?: React.ReactNode
}

export const MultiSelect = React.forwardRef<HTMLInputElement, MultiSelectProps>((props, ref) => {

    const [{
        children,
        className,
        size = "md",
        intent = "basic",
        isLoading,
        leftAddon,
        leftIcon,
        rightIcon,
        rightAddon,
        options,
        defaultValue,
        placeholder,
        value,
        onChange,
        max,
        discrete = false,
        menuContainerClassName,
        menuItemClassName,
        menuItemImageClassName,
        loadingContent = <p className="w-full text-center">...</p>,
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<MultiSelectProps>(props, useId())

    const inputRef = useRef<HTMLInputElement>(null)
    const ulRef = useRef<HTMLUListElement>(null)

    const [values, setValues] = useState<MultiSelectOption["value"][]>((value ?? defaultValue) ?? [])
    const [tagInputValue, setTagInputValue] = useState("")

    const inputFocus = useDisclosure(false)
    const listDisclosure = useDisclosure(false)
    const [highlightedOptionIndex, setHighlightedOptionIndex] = useState(0)

    // Options that are displayed in the list
    const selectOptions = useMemo(() => {
        // Filter by similar labels or values based on input
        // Show options which are not selected if input is empty
        const filtered = _filter(options, o => !values.includes(o.value))
        return tagInputValue.length > 0 ? _filter(filtered, o => (o.label
            ? o.label.toLowerCase().includes(tagInputValue.toLowerCase())
            : o.value.toLowerCase().includes(tagInputValue.toLowerCase()))) : filtered
    }, [options, values, tagInputValue])

    // Control value
    useUpdateEffect(() => {
        startTransition(() => {
            if (value) setValues(value)
        })
    }, [value])

    useUpdateEffect(() => {
        onChange && onChange(values)
    }, [values])


    const handleAddValue = useCallback((value: string) => {
        startTransition(() => {
            setValues(v => {
                if (!!max) {
                    // Simply add new value
                    if (max !== 1 && v.length < max) {
                        return [...v, value]
                    }
                    // Truncate and add new value
                    if (max !== 1 && v.length >= max) {
                        return [...v.slice(0, v.length - 1), value]
                    }
                    // Replace current value
                    if (max === 1) {
                        return [value]
                    }
                } else {
                    return [...v, value]
                }
                return v
            })
        })
    }, [max])

    const handlePopValue = useCallback(() => {
        setValues(v => {
            return v.slice(0, v.length - 1)
        })
    }, [])

    const handleRemoveValue = useCallback((value: string) => {
        setValues(v => {
            return v.filter(a => a !== value)
        })
    }, [])

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Enter" && inputRef.current) {
            // Add the only option
            if (selectOptions.length === 1 && !!selectOptions[0] && tagInputValue.length > 0) {
                handleAddValue(selectOptions[0].value) // Add value
                setTagInputValue("") // Reset input
                listDisclosure.close() // Close list
            }
            // Add the highlighted option
            if (tagInputValue.length === 0 && selectOptions[highlightedOptionIndex]) {
                handleAddValue(selectOptions[highlightedOptionIndex]!.value)
                setHighlightedOptionIndex(0)
            }
        }
        if ((event.key === "Backspace" || event.key === "Delete") && tagInputValue.length === 0) {
            handlePopValue()
        }
    }

    // Handle key navigation
    const handleKeyUp = useCallback((e: KeyboardEvent) => {
        if (e.key === "ArrowDown") {
            setHighlightedOptionIndex(i => {
                const newI = (i + 1) <= (selectOptions.length - 1) ? (i + 1) : 0
                scrollToHighlighted(newI)
                return newI
            })
        }
        if (e.key === "ArrowUp") {
            setHighlightedOptionIndex(i => {
                const newI = (i - 1) >= 0 ? (i - 1) : (selectOptions.length - 1)
                scrollToHighlighted(newI)
                return newI
            })
        }
    }, [selectOptions])

    useEffect(() => {
        window.addEventListener("keyup", handleKeyUp)
        if (inputRef.current) {
            inputRef.current.addEventListener("keydown", handleKeyDown)
        }

        return () => {
            if (inputRef.current) {
                inputRef.current.removeEventListener("keydown", handleKeyDown)
            }
            window.removeEventListener("keyup", handleKeyUp)
        }
    }, [inputRef, listDisclosure, selectOptions, highlightedOptionIndex])

    const itemsRef = useRef<any>(null)

    function getMap() {
        if (!itemsRef.current) {
            itemsRef.current = new Map()
        }
        return itemsRef.current
    }

    function scrollToHighlighted(index: number) {
        const map = getMap()
        const node = map.get(index)
        if (index === selectOptions.length - 1) {
            ulRef.current?.scrollTo({ top: ulRef.current?.scrollHeight, behavior: "smooth" })
        } else if (index === 0) {
            ulRef.current?.scrollTo({ top: 0, behavior: "smooth" })
        } else {
            node?.scrollIntoView({
                block: "nearest",
                inline: "end",
            })
        }
    }

    return (
        <>
            <BasicField
                {...basicFieldProps}
            >
                <div className={cn(inputContainerStyle())}>

                    <InputAddon addon={leftAddon} rightIcon={rightIcon} leftIcon={leftIcon} size={size} side={"left"}/>
                    <InputIcon icon={leftIcon} size={size} side={"left"}/>

                    <div
                        className={cn(
                            "form-input",
                            InputAnatomy.input({
                                size,
                                intent,
                                hasError: !!basicFieldProps.error,
                                untouchable: !!basicFieldProps.isDisabled,
                                hasRightAddon: !!rightAddon,
                                hasRightIcon: !!rightIcon,
                                hasLeftAddon: !!leftAddon,
                                hasLeftIcon: !!leftIcon,
                            }),
                            MultiSelectAnatomy.input({ isOpen: inputFocus.isOpen }),
                        )}
                        onClick={() => {
                            if (!inputFocus.isOpen && !isLoading) {
                                inputRef.current?.focus()
                            }
                        }}
                    >

                        {isLoading ? (
                            <div>{loadingContent}</div>
                        ) : <>

                            {/*List selected values*/}
                            {values.map((value, index) => (
                                <span key={index}>
                                    <Badge
                                        tag
                                        size="lg"
                                        intent={discrete ? "basic" : "gray"}
                                        isClosable={!basicFieldProps.isDisabled && !discrete}
                                        onClose={() => handleRemoveValue(value)}
                                        className={cn({ "px-1": discrete })}
                                    >
                                        <span>{_find(options, ["value", value])?.label ?? _find(options, ["value", value])?.value}</span>
                                    </Badge>
                                </span>
                            ))}

                            <input
                                id={basicFieldProps.id}
                                value={tagInputValue}
                                onChange={e => {
                                    inputFocus.open()
                                    setTagInputValue(e.target.value ?? "")
                                    if (selectOptions.length > 0) {
                                        listDisclosure.open()
                                    }
                                }}
                                onFocus={() => {
                                    inputFocus.open()
                                    listDisclosure.open()
                                }}
                                onClick={() => {
                                    inputFocus.open()
                                    listDisclosure.open()
                                }}
                                onBlur={() => {
                                    setTimeout(() => {
                                        inputFocus.close()
                                        listDisclosure.close()
                                    }, 200)
                                }}
                                onKeyUp={e => {
                                    e.key === "Enter" && e.preventDefault()
                                }}
                                disabled={basicFieldProps.isDisabled || isLoading}
                                className={cn("p-0 border-none outline-none focus:outline-none focus-visible:outline-none focus-visible:ring-0 focus-visible:ring-offset-0 !bg-transparent", { "w-1": !inputFocus.isOpen })}
                                ref={inputRef}
                            />

                            {/*List*/}
                            <Transition
                                show={listDisclosure.isOpen && selectOptions.length > 0}
                                as={Fragment}
                                leave="transition ease-in duration-100"
                                leaveFrom="opacity-100"
                                leaveTo="opacity-0"
                            >
                                <ul
                                    className={cn(MultiSelectAnatomy.menuContainer(), menuContainerClassName)}
                                    ref={ulRef}
                                >
                                    {selectOptions.map((o, index) => {

                                        const imageComponent = o.image ?
                                            <div
                                                className={cn(MultiSelectAnatomy.menuItemImage(), menuItemImageClassName)}>
                                                {o.image}
                                            </div> : <></>

                                        return (
                                            <li
                                                key={o.value}
                                                className={cn(MultiSelectAnatomy.menuItem({ highlighted: highlightedOptionIndex === index }), menuItemClassName)}
                                                onClick={() => {
                                                    handleAddValue(o.value)
                                                    setTagInputValue("")
                                                    startTransition(() => {
                                                        listDisclosure.close()
                                                        inputRef.current?.focus()
                                                        setTimeout(() => {
                                                            inputFocus.open()
                                                        }, 200)
                                                    })
                                                }}
                                                onMouseEnter={() => {
                                                    setHighlightedOptionIndex(index)
                                                }}
                                                ref={(node) => {
                                                    const map = getMap()
                                                    if (node) {
                                                        map.set(index, node)
                                                    } else {
                                                        map.delete(index)
                                                    }
                                                }}
                                            >
                                                <div className="flex w-full items-center">
                                                    {imageComponent}
                                                    <div>
                                                        <div
                                                            className={cn("text-base block truncate")}
                                                        >
                                                            {o.label ?? o.value}
                                                        </div>
                                                        {o.description && <div
                                                            className={cn("text-sm opacity-70")}>{o.description}</div>}
                                                    </div>
                                                </div>
                                            </li>
                                        )
                                    })}
                                </ul>
                            </Transition>

                        </>}
                    </div>

                    <InputAddon addon={rightAddon} rightIcon={rightIcon} leftIcon={leftAddon} size={size}
                                side={"right"}/>
                    <InputIcon icon={rightIcon} size={size} side={"right"}/>

                </div>
            </BasicField>
        </>
    )

})

MultiSelect.displayName = "MultiSelect"

/* -------------------------------------------------------------------------------------------------
 * Utils
 * -----------------------------------------------------------------------------------------------*/

function useDisclosure(
    initialState: boolean,
    callbacks?: { onOpen?(): void; onClose?(): void },
) {
    const [opened, setOpened] = useState(initialState)

    const open = () => {
        if (!opened) {
            setOpened(true)
            callbacks?.onOpen?.()
        }
    }

    const close = () => {
        if (opened) {
            setOpened(false)
            callbacks?.onClose?.()
        }
    }

    const toggle = () => {
        opened ? close() : open()
    }

    return { isOpen: opened, open, close, toggle } as const
}
