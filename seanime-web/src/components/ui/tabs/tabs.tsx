"use client"

import * as TabsPrimitive from "@radix-ui/react-tabs"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TabsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Tabs__root",
    ]),
    list: cva([
        "UI-Tabs__list",
        "inline-flex h-12 items-center justify-center w-full",
    ]),
    trigger: cva([
        "UI-Tabs__trigger appearance-none shadow-none",
        "inline-flex h-full items-center justify-center whitespace-nowrap px-3 py-1.5 text-sm text-[--muted] font-medium ring-offset-[--background]",
        "transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
        "disabled:pointer-events-none disabled:opacity-50",
        "border-transparent border-b-2 -mb-px",
        "data-[state=active]:border-[--brand] data-[state=active]:text-[--foreground]",
    ]),
    content: cva([
        "UI-Tabs__content",
        "ring-offset-[--background]",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] focus-visible:ring-offset-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Tabs
 * -----------------------------------------------------------------------------------------------*/

const __TabsAnatomyContext = React.createContext<ComponentAnatomy<typeof TabsAnatomy>>({})

export type TabsProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Root> & ComponentAnatomy<typeof TabsAnatomy>

export const Tabs = React.forwardRef<HTMLDivElement, TabsProps>((props, ref) => {
    const {
        className,
        listClass,
        triggerClass,
        contentClass,
        ...rest
    } = props

    return (
        <__TabsAnatomyContext.Provider
            value={{
                listClass,
                triggerClass,
                contentClass,
            }}
        >
            <TabsPrimitive.Root
                ref={ref}
                className={cn(TabsAnatomy.root(), className)}
                {...rest}
            />
        </__TabsAnatomyContext.Provider>
    )
})

Tabs.displayName = "Tabs"

/* -------------------------------------------------------------------------------------------------
 * TabsList
 * -----------------------------------------------------------------------------------------------*/

export type TabsListProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>

export const TabsList = React.forwardRef<HTMLDivElement, TabsListProps>((props, ref) => {
    const { className, ...rest } = props

    const { listClass } = React.useContext(__TabsAnatomyContext)

    return (
        <TabsPrimitive.List
            ref={ref}
            className={cn(TabsAnatomy.list(), listClass, className)}
            {...rest}
        />
    )
})

TabsList.displayName = "TabsList"


/* -------------------------------------------------------------------------------------------------
 * TabsTrigger
 * -----------------------------------------------------------------------------------------------*/

export type TabsTriggerProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>

export const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>((props, ref) => {
    const { className, ...rest } = props

    const { triggerClass } = React.useContext(__TabsAnatomyContext)

    return (
        <TabsPrimitive.Trigger
            ref={ref}
            className={cn(TabsAnatomy.trigger(), triggerClass, className)}
            {...rest}
        />
    )
})

TabsTrigger.displayName = "TabsTrigger"

/* -------------------------------------------------------------------------------------------------
 * TabsContent
 * -----------------------------------------------------------------------------------------------*/

export type TabsContentProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>

export const TabsContent = React.forwardRef<HTMLDivElement, TabsContentProps>((props, ref) => {
    const { className, ...rest } = props

    const { contentClass } = React.useContext(__TabsAnatomyContext)

    return (
        <TabsPrimitive.Content
            ref={ref}
            className={cn(TabsAnatomy.content(), contentClass, className)}
            {...rest}
        />
    )
})

TabsContent.displayName = "TabsContent"

