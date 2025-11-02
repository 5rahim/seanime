"use client"

import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DropdownMenuAnatomy = defineStyleAnatomy({
    subTrigger: cva([
        "UI-DropdownMenu__subTrigger",
        "focus:bg-[--subtle] data-[state=open]:bg-[--subtle]",
    ]),
    subContent: cva([
        "UI-DropdownMenu__subContent",
        "z-50 min-w-[12rem] overflow-hidden rounded-xl border bg-[--background] p-2 text-[--foreground] shadow-sm",
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
    root: cva([
        "UI-DropdownMenu__root",
        "z-50 min-w-[15rem] overflow-hidden rounded-xl border bg-[--background] p-2 text-[--foreground] shadow-sm",
        "data-[state=open]:animate-in data-[state=closed]:animate-out",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
    item: cva([
        "UI-DropdownMenu__item",
        "relative flex cursor-default select-none items-center rounded-xl cursor-pointer px-2 py-2 text-sm outline-none transition-colors",
        "focus:bg-[--subtle] data-[disabled]:pointer-events-none",
        "data-[disabled]:opacity-50",
        "[&>svg]:mr-2 [&>svg]:text-lg",
    ]),
    group: cva([
        "UI-DropdownMenu__group",
    ]),
    label: cva([
        "UI-DropdownMenu__label",
        "px-2 py-1.5 text-sm font-semibold text-[--muted]",
    ]),
    separator: cva([
        "UI-DropdownMenu__separator",
        "-mx-1 my-2 h-px bg-[--border]",
    ]),
    shortcut: cva([
        "UI-DropdownMenu__shortcut",
        "ml-auto text-xs tracking-widest opacity-60",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu
 * -----------------------------------------------------------------------------------------------*/

const __DropdownMenuAnatomyContext = React.createContext<ComponentAnatomy<typeof DropdownMenuAnatomy>>({})

export type DropdownMenuProps =
    ComponentAnatomy<typeof DropdownMenuAnatomy> &
    Pick<React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Root>, "defaultOpen" | "open" | "onOpenChange" | "dir"> &
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Content> & {
    /**
     * Interaction with outside elements will be enabled and other elements will be visible to screen readers.
     */
    allowOutsideInteraction?: boolean
    /**
     * The trigger element that is always visible and is used to open the menu.
     */
    trigger?: React.ReactNode
}

export const DropdownMenu = React.forwardRef<HTMLDivElement, DropdownMenuProps>((props, ref) => {
    const {
        children,
        trigger,
        // Root
        defaultOpen,
        open,
        onOpenChange,
        dir,
        allowOutsideInteraction,
        // Content
        sideOffset = 4,
        className,
        subContentClass,
        subTriggerClass,
        shortcutClass,
        itemClass,
        labelClass,
        separatorClass,
        groupClass,
        ...rest
    } = props

    return (
        <__DropdownMenuAnatomyContext.Provider
            value={{
                subContentClass,
                subTriggerClass,
                shortcutClass,
                itemClass,
                labelClass,
                separatorClass,
                groupClass,
            }}
        >
            <DropdownMenuPrimitive.Root
                defaultOpen={defaultOpen}
                open={open}
                onOpenChange={onOpenChange}
                dir={dir}
                modal={!allowOutsideInteraction}
                {...rest}
            >
                <DropdownMenuPrimitive.Trigger asChild>
                    {trigger}
                </DropdownMenuPrimitive.Trigger>

                <DropdownMenuPrimitive.Portal>
                    <DropdownMenuPrimitive.Content
                        ref={ref}
                        sideOffset={sideOffset}
                        className={cn(DropdownMenuAnatomy.root(), className)}
                        {...rest}
                    >
                        {children}
                    </DropdownMenuPrimitive.Content>
                </DropdownMenuPrimitive.Portal>
            </DropdownMenuPrimitive.Root>
        </__DropdownMenuAnatomyContext.Provider>
    )
})

DropdownMenu.displayName = "DropdownMenu"


/* -------------------------------------------------------------------------------------------------
 * DropdownMenuGroup
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuGroupProps = React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Group>

export const DropdownMenuGroup = React.forwardRef<HTMLDivElement, DropdownMenuGroupProps>((props, ref) => {
    const { className, ...rest } = props

    const { groupClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <DropdownMenuPrimitive.Group
            ref={ref}
            className={cn(DropdownMenuAnatomy.group(), groupClass, className)}
            {...rest}
        />
    )
})

DropdownMenuGroup.displayName = "DropdownMenuGroup"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenuSub
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuSubProps =
    Pick<ComponentAnatomy<typeof DropdownMenuAnatomy>, "subTriggerClass"> &
    Pick<React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Sub>, "defaultOpen" | "open" | "onOpenChange"> &
    React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.SubContent> & {
    /**
     * The content of the default trigger element that will open the sub menu.
     *
     * By default, the trigger will be an item with a right chevron icon.
     */
    triggerContent?: React.ReactNode
    /**
     * Props to pass to the default trigger element that will open the sub menu.
     */
    triggerProps?: React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.SubTrigger>
    triggerInset?: boolean
}

export const DropdownMenuSub = React.forwardRef<HTMLDivElement, DropdownMenuSubProps>((props, ref) => {
    const {
        children,
        triggerContent,
        triggerProps,
        triggerInset,
        // Sub
        defaultOpen,
        open,
        onOpenChange,
        // SubContent
        sideOffset = 8,
        className,
        subTriggerClass,
        ...rest
    } = props

    const { subTriggerClass: _subTriggerClass, subContentClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <DropdownMenuPrimitive.Sub
            {...rest}
        >
            <DropdownMenuPrimitive.SubTrigger
                className={cn(
                    DropdownMenuAnatomy.item(),
                    DropdownMenuAnatomy.subTrigger(),
                    triggerInset && "pl-8",
                    _subTriggerClass,
                    subTriggerClass,
                    className,
                )}
                {...triggerProps}
            >
                {triggerContent}
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    className={cn(
                        DropdownMenuAnatomy.shortcut(),
                        "w-4 h-4 ml-auto",
                    )}
                >
                    <path d="m9 18 6-6-6-6" />
                </svg>
            </DropdownMenuPrimitive.SubTrigger>

            <DropdownMenuPrimitive.Portal>
                <DropdownMenuPrimitive.SubContent
                    ref={ref}
                    sideOffset={sideOffset}
                    className={cn(
                        DropdownMenuAnatomy.subContent(),
                        subContentClass,
                        className,
                    )}
                    {...rest}
                >
                    {children}
                </DropdownMenuPrimitive.SubContent>
            </DropdownMenuPrimitive.Portal>
        </DropdownMenuPrimitive.Sub>
    )
})

DropdownMenuSub.displayName = "DropdownMenuSub"


/* -------------------------------------------------------------------------------------------------
 * DropdownMenuItem
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuItemProps = React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Item> & {
    inset?: boolean
}

export const DropdownMenuItem = React.forwardRef<HTMLDivElement, DropdownMenuItemProps>((props, ref) => {
    const { className, inset, ...rest } = props

    const { itemClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <DropdownMenuPrimitive.Item
            ref={ref}
            className={cn(
                DropdownMenuAnatomy.item(),
                inset && "pl-8",
                itemClass,
                className,
            )}
            {...rest}
        />
    )
})
DropdownMenuItem.displayName = "DropdownMenuItem"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenuLabel
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuLabelProps = React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Label> & {
    inset?: boolean
}

export const DropdownMenuLabel = React.forwardRef<HTMLDivElement, DropdownMenuLabelProps>((props, ref) => {
    const { className, inset, ...rest } = props

    const { labelClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <DropdownMenuPrimitive.Label
            ref={ref}
            className={cn(
                DropdownMenuAnatomy.label(),
                inset && "pl-8",
                labelClass,
                className,
            )}
            {...rest}
        />
    )
})

DropdownMenuLabel.displayName = "DropdownMenuLabel"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenuSeparator
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuSeparatorProps = React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Separator>

export const DropdownMenuSeparator = React.forwardRef<HTMLDivElement, DropdownMenuSeparatorProps>((props, ref) => {
    const { className, ...rest } = props

    const { separatorClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <DropdownMenuPrimitive.Separator
            ref={ref}
            className={cn(DropdownMenuAnatomy.separator(), separatorClass, className)}
            {...rest}
        />
    )
})

DropdownMenuSeparator.displayName = "DropdownMenuSeparator"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenuShortcut
 * -----------------------------------------------------------------------------------------------*/

export type DropdownMenuShortcutProps = React.HTMLAttributes<HTMLSpanElement>

export const DropdownMenuShortcut = React.forwardRef<HTMLSpanElement, DropdownMenuShortcutProps>((props, ref) => {
    const { className, ...rest } = props

    const { shortcutClass } = React.useContext(__DropdownMenuAnatomyContext)

    return (
        <span
            ref={ref}
            className={cn(DropdownMenuAnatomy.shortcut(), shortcutClass, className)}
            {...rest}
        />
    )
})

DropdownMenuShortcut.displayName = "DropdownMenuShortcut"

