"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DividerAnatomy = defineStyleAnatomy({
    divider: cva([
        "UI-Divider__divider",
        "w-full border-gray-200 dark:border-gray-700",
    ])
})

/* -------------------------------------------------------------------------------------------------
 * Divider
 * -----------------------------------------------------------------------------------------------*/

export interface DividerProps extends React.ComponentPropsWithRef<"hr">, ComponentWithAnatomy<typeof DividerAnatomy> {
}

export const Divider: React.FC<DividerProps> = React.forwardRef<HTMLHRElement, DividerProps>((props, ref) => {

    const {
        children,
        dividerClassName,
        className,
        ...rest
    } = props

    return (
        <hr className={cn(DividerAnatomy.divider(), dividerClassName, className)} {...rest} ref={ref}/>
    )

})

Divider.displayName = "Divider"
