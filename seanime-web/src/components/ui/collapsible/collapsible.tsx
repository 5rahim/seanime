"use client"

import * as CollapsiblePrimitive from "@radix-ui/react-collapsible"
import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * Collapsible
 * -----------------------------------------------------------------------------------------------*/

export const Collapsible = CollapsiblePrimitive.Root

Collapsible.displayName = "Collapsible"

/* -------------------------------------------------------------------------------------------------
 * CollapsibleTrigger
 * -----------------------------------------------------------------------------------------------*/

export type CollapsibleTriggerProps = React.ComponentPropsWithoutRef<typeof CollapsiblePrimitive.Trigger>

export const CollapsibleTrigger = React.forwardRef<HTMLButtonElement, CollapsibleTriggerProps>((props, ref) => {
    const { children, ...rest } = props

    return (
        <CollapsiblePrimitive.Trigger
            ref={ref}
            asChild
            {...rest}
        >
            {children}
        </CollapsiblePrimitive.Trigger>
    )
})

CollapsibleTrigger.displayName = "CollapsibleTrigger"

/* -------------------------------------------------------------------------------------------------
 * CollapsibleContent
 * -----------------------------------------------------------------------------------------------*/

export type CollapsibleContentProps = React.ComponentPropsWithoutRef<typeof CollapsiblePrimitive.Content>

export const CollapsibleContent = React.forwardRef<HTMLDivElement, CollapsibleContentProps>((props, ref) => {

    return (
        <CollapsiblePrimitive.Content
            ref={ref}
            {...props}
        />
    )
})

CollapsibleContent.displayName = "CollapsibleContent"

