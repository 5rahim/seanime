"use client"

import { cva } from "class-variance-authority"
import { Command as CommandPrimitive } from "cmdk"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { InputAnatomy } from "../input"
import { Modal, ModalProps } from "../modal"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CommandAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Command__root",
        "flex h-full w-full flex-col overflow-hidden rounded-[--radius-md] bg-[--paper] text-[--foreground]",
    ]),
    inputContainer: cva([
        "UI-Command__input",
        "flex items-center px-3 py-2",
        "cmdk-input-wrapper",
    ]),
    inputIcon: cva([
        "UI-Command__inputIcon",
        "mr-2 h-5 w-5 shrink-0 opacity-50",
    ]),
    list: cva([
        "UI-Command__list",
        "max-h-[300px] overflow-y-auto overflow-x-hidden",
    ]),
    empty: cva([
        "UI-Command__empty",
        "py-6 text-center text-base text-[--muted]",
    ]),
    group: cva([
        "UI-Command__group",
        "overflow-hidden p-1 text-[--foreground]",
        "[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-sm [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-[--muted]",
    ]),
    separator: cva([
        "UI-Command__separator",
        "-mx-1 h-px bg-[--border]",
    ]),
    item: cva([
        "UI-Command__item",
        "relative flex cursor-default select-none items-center rounded-[--radius] px-2 py-1.5 text-base outline-none",
        "aria-selected:bg-[--subtle] data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50",
        "[&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
    ]),
    itemIconContainer: cva([
        "UI-Command__itemIconContainer",
        "mr-2 text-base shrink-0 w-4",
    ]),
    shortcut: cva([
        "UI-Command__shortcut",
        "ml-auto text-xs tracking-widest text-[--muted]",
    ]),
})

export const CommandDialogAnatomy = defineStyleAnatomy({
    content: cva([
        "UI-CommandDialog__content",
        "overflow-hidden p-0",
    ]),
    command: cva([
        "UI-CommandDialog__command",
        "[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group-heading]]:text-[--muted]",
        "[&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pb-2 [&_[cmdk-group]]:px-2 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5",
        "[&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-2 [&_[cmdk-item]_svg]:h-4 [&_[cmdk-item]_svg]:w-5",
    ]),
})


/* -------------------------------------------------------------------------------------------------
 * Command
 * -----------------------------------------------------------------------------------------------*/

const __CommandAnatomyContext = React.createContext<ComponentAnatomy<typeof CommandAnatomy>>({})

export type CommandProps = React.ComponentPropsWithoutRef<typeof CommandPrimitive> & ComponentAnatomy<typeof CommandAnatomy>

export const Command = React.forwardRef<HTMLDivElement, CommandProps>((props, ref) => {
    const {
        className,
        inputContainerClass,
        inputIconClass,
        listClass,
        emptyClass,
        groupClass,
        separatorClass,
        itemClass,
        itemIconContainerClass,
        shortcutClass,
        loop = true,
        ...rest
    } = props

    return (
        <__CommandAnatomyContext.Provider
            value={{
                inputContainerClass,
                inputIconClass,
                listClass,
                emptyClass,
                groupClass,
                separatorClass,
                itemClass,
                itemIconContainerClass,
                shortcutClass,
            }}
        >
            <CommandPrimitive
                ref={ref}
                className={cn(CommandAnatomy.root(), className)}
                loop={loop}
                {...rest}
            />
        </__CommandAnatomyContext.Provider>
    )
})
Command.displayName = CommandPrimitive.displayName

/* -------------------------------------------------------------------------------------------------
 * CommandInput
 * -----------------------------------------------------------------------------------------------*/

export type CommandInputProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Input>
    & Pick<ComponentAnatomy<typeof CommandAnatomy>, "inputContainerClass" | "inputIconClass">

export const CommandInput = React.forwardRef<HTMLInputElement, CommandInputProps>((props, ref) => {
    const {
        className,
        inputContainerClass,
        inputIconClass,
        ...rest
    } = props

    const {
        inputContainerClass: _inputContainerClass,
        inputIconClass: _inputIconClass,
    } = React.useContext(__CommandAnatomyContext)

    return (
        <div className={cn(CommandAnatomy.inputContainer(), _inputContainerClass, inputContainerClass)} cmdk-input-wrapper="">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                className={cn(CommandAnatomy.inputIcon(), _inputIconClass, inputIconClass)}
            >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.3-4.3" />
            </svg>
            <CommandPrimitive.Input
                ref={ref}
                className={cn(InputAnatomy.root({
                    intent: "unstyled",
                    size: "sm",
                    isDisabled: rest.disabled,
                }), className)}
                {...rest}
            />
        </div>
    )
})
CommandInput.displayName = "CommandInput"

/* -------------------------------------------------------------------------------------------------
 * CommandList
 * -----------------------------------------------------------------------------------------------*/

export type CommandListProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.List>

export const CommandList = React.forwardRef<HTMLDivElement, CommandListProps>((props, ref) => {
    const { className, ...rest } = props

    const { listClass } = React.useContext(__CommandAnatomyContext)

    return (
        <CommandPrimitive.List
            ref={ref}
            className={cn(CommandAnatomy.list(), listClass, className)}
            {...rest}
        />
    )
})
CommandList.displayName = "CommandList"

/* -------------------------------------------------------------------------------------------------
 * CommandEmpty
 * -----------------------------------------------------------------------------------------------*/

export type CommandEmptyProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Empty>

export const CommandEmpty = React.forwardRef<HTMLDivElement, CommandEmptyProps>((props, ref) => {
    const { className, ...rest } = props

    const { emptyClass } = React.useContext(__CommandAnatomyContext)

    return (
        <CommandPrimitive.Empty
            ref={ref}
            className={cn(CommandAnatomy.empty(), emptyClass, className)}
            {...rest}
        />
    )
})
CommandEmpty.displayName = "CommandEmpty"

/* -------------------------------------------------------------------------------------------------
 * CommandGroup
 * -----------------------------------------------------------------------------------------------*/

export type CommandGroupProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Group>

export const CommandGroup = React.forwardRef<HTMLDivElement, CommandGroupProps>((props, ref) => {
    const { className, ...rest } = props

    const { groupClass } = React.useContext(__CommandAnatomyContext)

    return (
        <CommandPrimitive.Group
            ref={ref}
            className={cn(CommandAnatomy.group(), groupClass, className)}
            {...rest}
        />
    )
})
CommandGroup.displayName = "CommandGroup"

/* -------------------------------------------------------------------------------------------------
 * CommandSeparator
 * -----------------------------------------------------------------------------------------------*/

export type CommandSeparatorProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Separator>

export const CommandSeparator = React.forwardRef<HTMLDivElement, CommandSeparatorProps>((props, ref) => {
    const { className, ...rest } = props

    const { separatorClass } = React.useContext(__CommandAnatomyContext)

    return (
        <CommandPrimitive.Separator
            ref={ref}
            className={cn(CommandAnatomy.separator(), separatorClass, className)}
            {...rest}
        />
    )
})
CommandSeparator.displayName = "CommandSeparator"

/* -------------------------------------------------------------------------------------------------
 * CommandItem
 * -----------------------------------------------------------------------------------------------*/

export type CommandItemProps =
    React.ComponentPropsWithoutRef<typeof CommandPrimitive.Item>
    & Pick<ComponentAnatomy<typeof CommandAnatomy>, "itemIconContainerClass">
    & { leftIcon?: React.ReactNode }

export const CommandItem = React.forwardRef<HTMLDivElement, CommandItemProps>((props, ref) => {
    const { className, itemIconContainerClass, leftIcon, children, ...rest } = props

    const {
        itemClass,
        itemIconContainerClass: _itemIconContainerClass,
    } = React.useContext(__CommandAnatomyContext)

    const itemRef = React.useRef<HTMLDivElement | null>(null)

    React.useEffect(() => {
        const element = itemRef.current
        if (!element) return

        const observer = new MutationObserver((mutations) => {
            mutations.forEach((mutation) => {
                if (mutation.attributeName === "aria-selected" && element.getAttribute("aria-selected") === "true") {
                    element.scrollIntoView({ block: "nearest" })
                }
            })
        })

        observer.observe(element, { attributes: true })
        return () => observer.disconnect()
    }, [])

    const setRefs = React.useCallback(
        (node: HTMLDivElement | null) => {
            itemRef.current = node

            if (ref) {
                if (typeof ref === "function") {
                    ref(node)
                }
            }
        },
        [ref],
    )

    return (
        <CommandPrimitive.Item
            ref={setRefs}
            className={cn(CommandAnatomy.item(), itemClass, className)}
            {...rest}
            data-cmdkvalue={rest.id}
        >
            {leftIcon && (
                <span className={cn(CommandAnatomy.itemIconContainer(), _itemIconContainerClass, itemIconContainerClass)}>
                    {leftIcon}
                </span>
            )}
            {children}
        </CommandPrimitive.Item>
    )
})
CommandItem.displayName = "CommandItem"

/* -------------------------------------------------------------------------------------------------
 * CommandShortcut
 * -----------------------------------------------------------------------------------------------*/

export type CommandShortcutProps = React.ComponentPropsWithoutRef<"span">

export const CommandShortcut = React.forwardRef<HTMLSpanElement, CommandShortcutProps>((props, ref) => {
    const { className, ...rest } = props

    const { shortcutClass } = React.useContext(__CommandAnatomyContext)

    return (
        <span
            ref={ref}
            className={cn(CommandAnatomy.shortcut(), shortcutClass, className)}
            {...rest}
        />
    )
})
CommandShortcut.displayName = "CommandShortcut"

/* -------------------------------------------------------------------------------------------------
 * CommandDialog
 * -----------------------------------------------------------------------------------------------*/

export type CommandDialogProps = ModalProps & ComponentAnatomy<typeof CommandDialogAnatomy> & {
    commandProps: React.ComponentPropsWithoutRef<typeof CommandPrimitive>
}

export const CommandDialog = (props: CommandDialogProps) => {
    const { children, commandClass, contentClass, commandProps, ...rest } = props
    return (
        <Modal
            {...rest}
            contentClass={cn(CommandDialogAnatomy.content(), contentClass)}
        >
            <Command shouldFilter={false} className={cn(CommandDialogAnatomy.command(), commandClass)} {...commandProps}>
                {children}
            </Command>
        </Modal>
    )
}

CommandDialog.displayName = "CommandDialog"
