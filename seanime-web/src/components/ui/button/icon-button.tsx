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
                xs: "text-lg h-6 w-6",
                sm: "text-lg h-8 w-8",
                md: "text-xl h-9 w-9",
                lg: "text-2xl h-10 w-10",
                xl: "text-3xl h-12 w-12",
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
