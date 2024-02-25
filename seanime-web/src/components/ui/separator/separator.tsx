"use client"

import { cn } from "../core/styling"
import * as SeparatorPrimitive from "@radix-ui/react-separator"
import { cva } from "class-variance-authority"
import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SeparatorAnatomy = {
    root: cva([
        "UI-Separator__root",
        "shrink-0 bg-[--border]",
    ], {
        variants: {
            orientation: {
                horizontal: "w-full h-[1px]",
                vertical: "h-full w-[1px]",
            },
        },
    }),
}

/* -------------------------------------------------------------------------------------------------
 * Separator
 * -----------------------------------------------------------------------------------------------*/

export type SeparatorProps = React.ComponentPropsWithoutRef<typeof SeparatorPrimitive.Root>

export const Separator = React.forwardRef<HTMLDivElement, SeparatorProps>((props, ref) => {
    const {
        className,
        orientation = "horizontal",
        decorative = true,
        ...rest
    } = props

    return (
        <SeparatorPrimitive.Root
            ref={ref}
            decorative={decorative}
            orientation={orientation}
            className={cn(
                SeparatorAnatomy.root({ orientation }),
                className,
            )}
            {...rest}
        />
    )
})

Separator.displayName = "Separator"
