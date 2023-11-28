"use client"

import React from "react"
import { Disclosure, Transition } from "@headlessui/react"
import { cn, ComponentWithAnatomy, createPolymorphicComponent, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AccordionAnatomy = defineStyleAnatomy({
    container: cva([
        "UI-Accordion__container",
        "space-y-2"
    ]),
    trigger: cva([
        "UI-Accordion__trigger",
        "bg-gray-50 dark:bg-gray-800 hover:bg-gray-100 dark:hover:bg-gray-700",
        "flex w-full justify-between rounded-lg px-4 py-3 text-left font-medium transition",
        "focus:outline-none focus-visible:ring-2 focus-visible:ring-[--ring] ring-offset-1",
    ]),
    panel: cva([
        "UI-Accordion__panel",
        "py-2"
    ]),
    item: cva(["UI-Accordion__item"]),
})

/* -------------------------------------------------------------------------------------------------
 * Accordion
 * -----------------------------------------------------------------------------------------------*/

export interface AccordionProps extends React.ComponentPropsWithoutRef<"div">, ComponentWithAnatomy<typeof AccordionAnatomy> {
}

const _Accordion = (props: AccordionProps) => {

    const {
        children,
        triggerClassName,
        containerClassName,
        panelClassName,
        itemClassName,
        className,
        ...rest
    } = props

    const itemsWithProps = React.Children.map(children, (child) => {
        // Checking isValidElement is the safe way and avoids a typescript error too.
        if (React.isValidElement(child)) {
            return React.cloneElement(child, { triggerClassName, panelClassName, itemClassName } as any)
        }
        return child
    })

    return (
        <div className={cn(AccordionAnatomy.container(), containerClassName, className)}>
            {itemsWithProps}
        </div>
    )

}

/* -------------------------------------------------------------------------------------------------
 * Accordion.Item
 * -----------------------------------------------------------------------------------------------*/

interface AccordionItemProps extends React.ComponentPropsWithoutRef<"div">,
    Omit<ComponentWithAnatomy<typeof AccordionAnatomy>, "containerClassName"> {
    title: string,
    defaultOpen?: boolean
}

export const AccordionItem: React.FC<AccordionItemProps> = (
    {
        children,
        title,
        triggerClassName,
        panelClassName,
        itemClassName,
        className,
        defaultOpen,
        ...rest
    }) => {

    return (
        <Disclosure defaultOpen={defaultOpen}>
            {({ open }) => (
                <div className={cn(AccordionAnatomy.item(), itemClassName, className)} {...rest}>
                    <Disclosure.Button className={cn(AccordionAnatomy.trigger(), triggerClassName)}>
                        {title}
                        <svg
                            className={cn(
                                "ml-1.5 h-5 w-5 flex-shrink-0 transition duration-300",
                                {
                                    "-rotate-180": open,
                                },
                            )}
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
                    </Disclosure.Button>
                    <Transition
                        show={open}
                        enter="transition duration-300 ease-out"
                        enterFrom="transform h-0 opacity-0"
                        enterTo="transform h-full opacity-100"
                        leave="transition duration-75 ease-out"
                        leaveFrom="transform h-full opacity-100"
                        leaveTo="transform h-0 opacity-0"
                    >
                        <Disclosure.Panel static className={cn(AccordionAnatomy.panel(), panelClassName)}>
                            {children}
                        </Disclosure.Panel>
                    </Transition>
                </div>
            )}
        </Disclosure>
    )
}

AccordionItem.displayName = "AccordionItem"

/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_Accordion.Item = AccordionItem

export const Accordion = createPolymorphicComponent<"div", AccordionProps, {
    Item: typeof AccordionItem
}>(_Accordion)

Accordion.displayName = "Accordion"
