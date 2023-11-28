"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DividerWithLabelAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DividerWithLabel__root",
        "relative"
    ]),
    dividerContainer: cva([
        "DividerWithLabel__dividerContainer",
        "absolute inset-0 flex items-center"
    ]),
    divider: cva([
        "DividerWithLabel__divider",
        "w-full border-t border-gray-300 border-[--border]"
    ]),
    labelContainer: cva([
        "DividerWithLabel__labelContainer",
        "relative flex justify-center"
    ]),
    label: cva([
        "DividerWithLabel__label",
        "bg-[--background-color] px-2 text-sm text-[--muted]"
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DividerWithLabel
 * -----------------------------------------------------------------------------------------------*/

export interface DividerWithLabelProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof DividerWithLabelAnatomy> {
    children?: React.ReactNode
}

export const DividerWithLabel: React.FC<DividerWithLabelProps> = React.forwardRef<HTMLDivElement, DividerWithLabelProps>((props, ref) => {

    const {
        children,
        rootClassName,
        dividerClassName,
        dividerContainerClassName,
        labelClassName,
        labelContainerClassName,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(DividerWithLabelAnatomy.root(), rootClassName, className)}
            {...rest}
            ref={ref}
        >
            <div className={cn(DividerWithLabelAnatomy.dividerContainer(), dividerContainerClassName)}
                 aria-hidden="true">
                <div className={cn(DividerWithLabelAnatomy.divider(), dividerClassName)}/>
            </div>
            <div className={cn(DividerWithLabelAnatomy.labelContainer(), labelContainerClassName)}>
                <span className={cn(DividerWithLabelAnatomy.label(), labelClassName)}>{children}</span>
            </div>
        </div>
    )

})

DividerWithLabel.displayName = "DividerWithLabel"
