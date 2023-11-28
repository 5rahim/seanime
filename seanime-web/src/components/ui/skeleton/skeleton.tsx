import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const SkeletonAnatomy = defineStyleAnatomy({
    skeleton: cva("UI-Skeleton__skeleton", {
        variants: {
            type: {
                box: "h-14 bg-gray-100 dark:bg-gray-800 w-full rounded-[--radius] animate-pulse",
                text: "h-2 bg-gray-100 dark:bg-gray-800 rounded-[--radius] animate-purple",
            },
        },
        defaultVariants: {},
    }),
})

/* -------------------------------------------------------------------------------------------------
 * Skeleton
 * -----------------------------------------------------------------------------------------------*/

export interface SkeletonProps extends React.ComponentPropsWithRef<"div">, VariantProps<typeof SkeletonAnatomy.skeleton>,
    ComponentWithAnatomy<typeof SkeletonAnatomy> {
}

export const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>((props, ref) => {

    const {
        children,
        className,
        type = "box",
        skeletonClassName,
        ...rest
    } = props

    // Text
    if (type === "text") {
        return (
            <>
                <div className="flex flex-col gap-3">
                    <div
                        className={cn("w-full", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                    <div
                        className={cn("w-full", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                    <div
                        className={cn("w-full", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                    <div
                        className={cn("w-[98%]", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                    <div
                        className={cn("w-[95%]", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                    <div
                        className={cn("w-[90%]", SkeletonAnatomy.skeleton({ type }), skeletonClassName, className)} {...rest}
                        ref={ref}></div>
                </div>
            </>
        )
    }

    // Box
    return (
        <>
            <div
                className={cn(
                    SkeletonAnatomy.skeleton({ type }),
                    skeletonClassName,
                    className,
                )}
                {...rest}
                ref={ref}
            ></div>
        </>
    )

})

Skeleton.displayName = "Skeleton"
