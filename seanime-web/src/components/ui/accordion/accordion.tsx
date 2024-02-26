"use client"

import * as AccordionPrimitive from "@radix-ui/react-accordion"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AccordionAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Accordion__root",
    ]),
    header: cva([
        "UI-Accordion__header",
        "flex text-lg",
    ]),
    trigger: cva([
        "UI-Accordion__trigger",
        "flex flex-1 items-center justify-between px-4 py-2 font-medium transition-all hover:bg-[--subtle] [&[data-state=open]>svg]:rotate-180",
    ]),
    triggerIcon: cva([
        "UI-Accordion__triggerIcon",
        "h-4 w-4 shrink-0 transition-transform duration-200",
    ]),
    item: cva([
        "UI-Accordion__item",
        "",
    ]),
    contentContainer: cva([
        "UI-Accordion__contentContainer",
        "overflow-hidden transition-all data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down",
    ]),
    content: cva([
        "UI-Accordion__content",
        "p-4",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Accordion
 * -----------------------------------------------------------------------------------------------*/

const __AccordionAnatomyContext = React.createContext<ComponentAnatomy<typeof AccordionAnatomy>>({})

export type AccordionProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Root> & ComponentAnatomy<typeof AccordionAnatomy>

export const Accordion = React.forwardRef<HTMLDivElement, AccordionProps>((props, ref) => {

    const {
        className,
        headerClass,
        triggerClass,
        triggerIconClass,
        contentContainerClass,
        contentClass,
        itemClass,
        ...rest
    } = props

    return (
        <__AccordionAnatomyContext.Provider
            value={{
                itemClass,
                headerClass,
                triggerClass,
                triggerIconClass,
                contentContainerClass,
                contentClass,
            }}
        >
            <AccordionPrimitive.Root
                ref={ref}
                className={cn(AccordionAnatomy.root(), className)}
                {...rest}
            />
        </__AccordionAnatomyContext.Provider>
    )

})

Accordion.displayName = "Accordion"

/* -------------------------------------------------------------------------------------------------
 * AccordionItem
 * -----------------------------------------------------------------------------------------------*/

export type AccordionItemProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Item>

export const AccordionItem = React.forwardRef<HTMLDivElement, AccordionItemProps>((props, ref) => {

    const { className, ...rest } = props

    const { itemClass } = React.useContext(__AccordionAnatomyContext)

    return (
        <AccordionPrimitive.Item
            ref={ref}
            className={cn(AccordionAnatomy.item(), itemClass, className)}
            {...rest}
        />
    )

})

AccordionItem.displayName = "AccordionItem"

/* -------------------------------------------------------------------------------------------------
 * AccordionTrigger
 * -----------------------------------------------------------------------------------------------*/

export type AccordionTriggerProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Trigger> &
    Pick<ComponentAnatomy<typeof AccordionAnatomy>, "headerClass" | "triggerIconClass">

export const AccordionTrigger = React.forwardRef<HTMLButtonElement, AccordionTriggerProps>((props, ref) => {

    const {
        className,
        headerClass,
        triggerIconClass,
        children,
        ...rest
    } = props

    const {
        headerClass: _headerClass,
        triggerClass: _triggerClass,
        triggerIconClass: _triggerIconClass,
    } = React.useContext(__AccordionAnatomyContext)

    return (
        <AccordionPrimitive.Header className={cn(AccordionAnatomy.header(), _headerClass, headerClass)}>
            <AccordionPrimitive.Trigger
                ref={ref}
                className={cn(
                    AccordionAnatomy.trigger(),
                    _triggerClass,
                    className,
                )}
                {...rest}
            >
                {children}
                <svg
                    className={cn(AccordionAnatomy.triggerIcon(), _triggerIconClass, triggerIconClass)}
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                >
                    <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth="2"
                        d="M19 9l-7 7-7-7"
                    />
                </svg>
            </AccordionPrimitive.Trigger>
        </AccordionPrimitive.Header>
    )

})

AccordionTrigger.displayName = "AccordionTrigger"

/* -------------------------------------------------------------------------------------------------
 * AccordionContent
 * -----------------------------------------------------------------------------------------------*/

export type AccordionContentProps = React.ComponentPropsWithoutRef<typeof AccordionPrimitive.Content> &
    Pick<ComponentAnatomy<typeof AccordionAnatomy>, "contentContainerClass">

export const AccordionContent = React.forwardRef<HTMLDivElement, AccordionContentProps>((props, ref) => {

    const {
        className,
        contentContainerClass,
        children,
        ...rest
    } = props

    const {
        contentContainerClass: _contentContainerClass,
        contentClass: _contentClass,
    } = React.useContext(__AccordionAnatomyContext)

    return (
        <AccordionPrimitive.Content
            ref={ref}
            className={cn(AccordionAnatomy.contentContainer(), _contentContainerClass, contentContainerClass)}
            {...rest}
        >
            <div className={cn(AccordionAnatomy.content(), _contentClass, className)}>
                {children}
            </div>
        </AccordionPrimitive.Content>
    )
})

AccordionContent.displayName = "AccordionContent"

