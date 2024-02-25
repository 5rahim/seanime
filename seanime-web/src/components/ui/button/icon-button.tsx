import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { Button, ButtonProps } from "."
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const IconButtonAnatomy = defineStyleAnatomy({
    root: cva("UI-IconButton_root p-0 flex-none", {
        variants: {
            size: {
                xs: "text-xl h-6 w-6",
                sm: "text-xl h-8 w-8",
                md: "text-2xl h-10 w-10",
                lg: "text-3xl h-12 w-12",
                xl: "text-4xl h-14 w-14",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * IconButton
 * -----------------------------------------------------------------------------------------------*/


export type IconButtonProps = Omit<ButtonProps, "leftIcon" | "rightIcon" | "iconSpacing" | "iconClass" | "children"> &
    VariantProps<typeof IconButtonAnatomy.root> & {
    icon?: React.ReactNode
}

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>((props, ref) => {

    const {
        className,
        icon,
        size,
        loading,
        ...rest
    } = props

    return (
        <Button
            className={cn(
                IconButtonAnatomy.root({ size }),
                className,
            )}
            loading={loading}
            iconSpacing="0"
            {...rest}
            ref={ref}
        >
            {!loading && icon}
        </Button>
    )

})

IconButton.displayName = "IconButton"
