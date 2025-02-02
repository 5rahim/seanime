import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SkeletonAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Skeleton__root",
        "animate-pulse rounded-[--radius-md] bg-[--subtle] w-full h-12",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Skeleton
 * -----------------------------------------------------------------------------------------------*/

export type SkeletonProps = React.ComponentPropsWithoutRef<"div">

export const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <div
            ref={ref}
            className={cn(SkeletonAnatomy.root(), className)}
            {...rest}
        />
    )
})

Skeleton.displayName = "Skeleton"
