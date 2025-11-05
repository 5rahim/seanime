"use client"

import * as ContextMenuPrimitive from "@radix-ui/react-context-menu"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ContextMenuAnatomy = defineStyleAnatomy({
    subTrigger: cva([
        "UI-ContextMenu__subTrigger",
        "focus:bg-[--subtle] data-[state=open]:bg-[--subtle]",
    ]),
    subContent: cva([
        "UI-ContextMenu__subContent",
        "z-50 min-w-[12rem] overflow-hidden rounded-xl border bg-[--background] p-2 text-[--foreground] shadow-sm",
        "data-[state=open]:animate-in",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
    trigger: cva([
        "UI-ContextMenu__trigger",
    ]),
    content: cva([
        "UI-ContextMenu__content",
    ]),
    root: cva([
        "UI-ContextMenu__root",
        "z-50 min-w-[15rem] overflow-hidden rounded-xl border bg-[--background] p-2 text-[--foreground] shadow-sm",
        "data-[state=open]:animate-in",
        "data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
        "data-[state=closed]:zoom-out-100 data-[state=open]:zoom-in-95",
        "data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2",
        "data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2",
    ]),
    item: cva([
        "UI-ContextMenu__item",
        "relative flex cursor-default select-none items-center rounded-[--radius] cursor-pointer px-2 py-2 text-sm outline-none transition-colors",
        "focus:bg-[--subtle] data-[disabled]:pointer-events-none",
        "data-[disabled]:opacity-50",
        "[&>svg]:mr-2 [&>svg]:text-lg",
    ]),
    group: cva([
        "UI-ContextMenu__group",
    ]),
    label: cva([
        "UI-ContextMenu__label",
        "px-2 py-1.5 text-sm font-semibold text-[--muted]",
    ]),
    separator: cva([
        "UI-ContextMenu__separator",
        "-mx-1 my-1 h-px bg-[--border]",
    ]),
    shortcut: cva([
        "UI-ContextMenu__shortcut",
        "ml-auto text-xs tracking-widest opacity-60",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * ContextMenu
 * -----------------------------------------------------------------------------------------------*/

const __ContextMenuAnatomyContext = React.createContext<ComponentAnatomy<typeof ContextMenuAnatomy> & { className?: string }>({})

export type ContextMenuProps =
    ComponentAnatomy<typeof ContextMenuAnatomy> &
    React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Root> &
    React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Content> & {
    /**
     * Interaction with outside elements will be enabled and other elements will be visible to screen readers.
     */
    allowOutsideInteraction?: boolean
    /**
     * The trigger element that is always visible and is used to open the menu.
     */
    trigger?: React.ReactNode
}

export const ContextMenu = React.forwardRef<HTMLDivElement, ContextMenuProps>((props, ref) => {
    const {
        children,
        trigger,
        // Root
        onOpenChange,
        dir,
        allowOutsideInteraction,
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
        <__ContextMenuAnatomyContext.Provider
            value={{
                className,
                subContentClass,
                subTriggerClass,
                shortcutClass,
                itemClass,
                labelClass,
                separatorClass,
                groupClass,
            }}
        >
            <ContextMenuPrimitive.Root
                dir={dir}
                modal={!allowOutsideInteraction}
                {...rest}
            >

                {children}

            </ContextMenuPrimitive.Root>
        </__ContextMenuAnatomyContext.Provider>
    )
})

ContextMenu.displayName = "ContextMenu"

/* -------------------------------------------------------------------------------------------------
 * ContextMenuTrigger
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuTriggerProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Trigger>

export const ContextMenuTrigger = React.forwardRef<HTMLDivElement, ContextMenuTriggerProps>((props, ref) => {
    const { className, ...rest } = props

    const { triggerClass } = React.useContext(__ContextMenuAnatomyContext)

    return <ContextMenuPrimitive.Trigger ref={ref} className={cn(triggerClass, className)} {...rest} />
})


/* -------------------------------------------------------------------------------------------------
 * ContextMenuContent
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuContentProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Content>

export const ContextMenuContent = React.forwardRef<HTMLDivElement, ContextMenuContentProps>((props, ref) => {
    const { className, ...rest } = props

    const { className: rootClass, contentClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Portal>
            <ContextMenuPrimitive.Content
                ref={ref}
                className={cn(ContextMenuAnatomy.root(), rootClass, contentClass, className)}
                {...rest}
            />
        </ContextMenuPrimitive.Portal>
    )
})

/* -------------------------------------------------------------------------------------------------
 * ContextMenuGroup
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuGroupProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Group>

export const ContextMenuGroup = React.forwardRef<HTMLDivElement, ContextMenuGroupProps>((props, ref) => {
    const { className, ...rest } = props

    const { groupClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Group
            ref={ref}
            className={cn(ContextMenuAnatomy.group(), groupClass, className)}
            {...rest}
        />
    )
})

ContextMenuGroup.displayName = "ContextMenuGroup"

/* -------------------------------------------------------------------------------------------------
 * ContextMenuSub
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuSubProps =
    Pick<ComponentAnatomy<typeof ContextMenuAnatomy>, "subTriggerClass"> &
    Pick<React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Sub>, "defaultOpen" | "open" | "onOpenChange"> &
    React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.SubContent> & {
    /**
     * The content of the default trigger element that will open the sub menu.
     *
     * By default, the trigger will be an item with a right chevron icon.
     */
    triggerContent?: React.ReactNode
    /**
     * Props to pass to the default trigger element that will open the sub menu.
     */
    triggerProps?: React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.SubTrigger>
    triggerInset?: boolean
}

export const ContextMenuSub = React.forwardRef<HTMLDivElement, ContextMenuSubProps>((props, ref) => {
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

    const { subTriggerClass: _subTriggerClass, subContentClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Sub
            {...rest}
        >
            <ContextMenuPrimitive.SubTrigger
                className={cn(
                    ContextMenuAnatomy.item(),
                    ContextMenuAnatomy.subTrigger(),
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
                        ContextMenuAnatomy.shortcut(),
                        "w-4 h-4 ml-auto",
                    )}
                >
                    <path d="m9 18 6-6-6-6" />
                </svg>
            </ContextMenuPrimitive.SubTrigger>

            <ContextMenuPrimitive.Portal>
                <ContextMenuPrimitive.SubContent
                    ref={ref}
                    sideOffset={sideOffset}
                    className={cn(
                        ContextMenuAnatomy.subContent(),
                        subContentClass,
                        className,
                    )}
                    {...rest}
                >
                    {children}
                </ContextMenuPrimitive.SubContent>
            </ContextMenuPrimitive.Portal>
        </ContextMenuPrimitive.Sub>
    )
})

ContextMenuSub.displayName = "ContextMenuSub"


/* -------------------------------------------------------------------------------------------------
 * ContextMenuItem
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuItemProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Item> & {
    inset?: boolean
}

export const ContextMenuItem = React.forwardRef<HTMLDivElement, ContextMenuItemProps>((props, ref) => {
    const { className, inset, ...rest } = props

    const { itemClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Item
            ref={ref}
            className={cn(
                ContextMenuAnatomy.item(),
                inset && "pl-8",
                itemClass,
                className,
            )}
            {...rest}
        />
    )
})
ContextMenuItem.displayName = "ContextMenuItem"

/* -------------------------------------------------------------------------------------------------
 * ContextMenuLabel
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuLabelProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Label> & {
    inset?: boolean
}

export const ContextMenuLabel = React.forwardRef<HTMLDivElement, ContextMenuLabelProps>((props, ref) => {
    const { className, inset, ...rest } = props

    const { labelClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Label
            ref={ref}
            className={cn(
                ContextMenuAnatomy.label(),
                inset && "pl-8",
                labelClass,
                className,
            )}
            {...rest}
        />
    )
})

ContextMenuLabel.displayName = "ContextMenuLabel"

/* -------------------------------------------------------------------------------------------------
 * ContextMenuSeparator
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuSeparatorProps = React.ComponentPropsWithoutRef<typeof ContextMenuPrimitive.Separator>

export const ContextMenuSeparator = React.forwardRef<HTMLDivElement, ContextMenuSeparatorProps>((props, ref) => {
    const { className, ...rest } = props

    const { separatorClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <ContextMenuPrimitive.Separator
            ref={ref}
            className={cn(ContextMenuAnatomy.separator(), separatorClass, className)}
            {...rest}
        />
    )
})

ContextMenuSeparator.displayName = "ContextMenuSeparator"

/* -------------------------------------------------------------------------------------------------
 * ContextMenuShortcut
 * -----------------------------------------------------------------------------------------------*/

export type ContextMenuShortcutProps = React.HTMLAttributes<HTMLSpanElement>

export const ContextMenuShortcut = React.forwardRef<HTMLSpanElement, ContextMenuShortcutProps>((props, ref) => {
    const { className, ...rest } = props

    const { shortcutClass } = React.useContext(__ContextMenuAnatomyContext)

    return (
        <span
            ref={ref}
            className={cn(ContextMenuAnatomy.shortcut(), shortcutClass, className)}
            {...rest}
        />
    )
})

ContextMenuShortcut.displayName = "ContextMenuShortcut"

