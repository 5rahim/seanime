"use client"

import * as AccordionPrimitive from "@radix-ui/react-accordion"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DisclosureAnatomy = defineStyleAnatomy({
    item: cva([
        "UI-Disclosure__item",
    ]),
    contentContainer: cva([
        "UI-Disclosure__contentContainer",
        "overflow-hidden transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down",
    ]),
    content: cva([
        "UI-Disclosure__content",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Disclosure
 * -----------------------------------------------------------------------------------------------*/

const __DisclosureAnatomyContext = React.createContext<ComponentAnatomy<typeof DisclosureAnatomy>>({})

export type DisclosureProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Root> &
    ComponentAnatomy<typeof DisclosureAnatomy>

export const Disclosure = React.forwardRef<HTMLDivElement, DisclosureProps>((props, ref) => {

    const {
        contentContainerClass,
        contentClass,
        itemClass,
        ...rest
    } = props

    return (
        <__DisclosureAnatomyContext.Provider
            value={{
                itemClass,
                contentContainerClass,
                contentClass,
            }}
        >
            <AccordionPrimitive.Root
                ref={ref}
                {...rest}
            />
        </__DisclosureAnatomyContext.Provider>
    )

})
Disclosure.displayName = "Disclosure"

/* -------------------------------------------------------------------------------------------------
 * DisclosureItem
 * -----------------------------------------------------------------------------------------------*/

export type DisclosureItemProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Item> &
    ComponentAnatomy<typeof DisclosureAnatomy>

export const DisclosureItem = React.forwardRef<HTMLDivElement, DisclosureItemProps>((props, ref) => {

    const { className, ...rest } = props

    const { itemClass } = React.useContext(__DisclosureAnatomyContext)

    return (
        <AccordionPrimitive.Item
            ref={ref}
            className={cn(DisclosureAnatomy.item(), itemClass, className)}
            {...rest}
        />
    )

})
DisclosureItem.displayName = "DisclosureItem"

/* -------------------------------------------------------------------------------------------------
 * DisclosureTrigger
 * -----------------------------------------------------------------------------------------------*/

export type DisclosureTriggerProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Trigger>

export const DisclosureTrigger = React.forwardRef<HTMLButtonElement, DisclosureTriggerProps>((props, ref) => {
    return (
        <AccordionPrimitive.Header asChild>
            <AccordionPrimitive.Trigger ref={ref} asChild {...props} />
        </AccordionPrimitive.Header>
    )
})
DisclosureTrigger.displayName = "DisclosureTrigger"

/* -------------------------------------------------------------------------------------------------
 * DisclosureContent
 * -----------------------------------------------------------------------------------------------*/

export type DisclosureContentProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Content>
    & Omit<ComponentAnatomy<typeof DisclosureAnatomy>, "contentClass">

export const DisclosureContent = React.forwardRef<HTMLDivElement, DisclosureContentProps>((props, ref) => {

    const {
        className,
        contentContainerClass,
        children,
        ...rest
    } = props

    const {
        contentContainerClass: _contentContainerClass,
        contentClass: _contentClass,
    } = React.useContext(__DisclosureAnatomyContext)

    return (
        <AccordionPrimitive.Content
            ref={ref}
            className={cn(DisclosureAnatomy.contentContainer(), _contentContainerClass, contentContainerClass)}
            {...rest}
        >
            <div className={cn(DisclosureAnatomy.content(), _contentClass, className)}>
                {children}
            </div>
        </AccordionPrimitive.Content>
    )
})
DisclosureContent.displayName = "DisclosureContent"

