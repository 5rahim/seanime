"use client"

import React from "react"
import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PaperAnatomy = defineStyleAnatomy({
    paper: cva([
        "UI-Paper__paper",
        "overflow-hidden rounded-lg bg-white border border-[--border] bg-[--paper]",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Paper
 * -----------------------------------------------------------------------------------------------*/

export interface PaperProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof PaperAnatomy> {
}

export const Paper: React.FC<PaperProps> = React.forwardRef<HTMLDivElement, PaperProps>((props, ref) => {

    const {
        children,
        paperClassName,
        className,
        ...rest
    } = props

    return (
        <div
            className={cn(PaperAnatomy.paper(), paperClassName, className)}
            {...rest}
            ref={ref}
        >
            {children}
        </div>
    )

})

Paper.displayName = "Paper"
