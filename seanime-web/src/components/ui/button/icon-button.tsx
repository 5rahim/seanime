import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React from "react"
import { Button, ButtonProps } from "."

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const IconButtonAnatomy = defineStyleAnatomy({
    iconButton: cva("UI-IconButton__iconButton p-0", {
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

export interface IconButtonProps extends Omit<ButtonProps, "leftIcon" | "rightIcon" | "iconSpacing" | "isUppercase">,
    VariantProps<typeof IconButtonAnatomy.iconButton>, ComponentWithAnatomy<typeof IconButtonAnatomy> {
    icon?: React.ReactElement<any, string | React.JSXElementConstructor<any>>
}

export const IconButton = React.forwardRef<HTMLButtonElement, IconButtonProps>((props, ref) => {

    const {
        children,
        className,
        icon,
        size,
        iconButtonClassName,
        ...rest
    } = props

    return (
        <>
            <Button
                className={cn(
                    IconButtonAnatomy.iconButton({ size }),
                    iconButtonClassName,
                    className,
                )}
                {...rest}
                ref={ref}
            >
                {icon}
            </Button>
        </>
    )

})

IconButton.displayName = "IconButton"
