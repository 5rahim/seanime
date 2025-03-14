import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BadgeAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Badge__root",
        "inline-flex flex-none text-base w-fit overflow-hidden justify-center items-center gap-2",
        "group/badge",
    ], {
        variants: {
            intent: {
                "gray": "text-gray-800 bg-gray-100 border border-gray-500 border-opacity-40 dark:text-gray-300 dark:bg-opacity-10",
                "primary": "text-indigo bg-indigo-50 border border-indigo-500 border-opacity-40 dark:text-indigo-300 dark:bg-opacity-10",
                "success": "text-green bg-green-50 border border-green-500 border-opacity-40 dark:text-green-300 dark:bg-opacity-10",
                "warning": "text-orange bg-orange-50 border border-orange-500 border-opacity-40 dark:text-orange-300 dark:bg-opacity-10",
                "alert": "text-red bg-red-50 border border-red-500 border-opacity-40 dark:text-red-300 dark:bg-opacity-10",
                "blue": "text-blue bg-blue-50 border border-blue-500 border-opacity-40 dark:text-blue-300 dark:bg-opacity-10",
                "info": "text-blue bg-blue-50 border border-blue-500 border-opacity-40 dark:text-blue-300 dark:bg-opacity-10",
                "white": "text-white bg-gray-800 border border-gray-500 border-opacity-40 dark:text-white dark:bg-opacity-10",
                "basic": "text-gray-900 bg-transparent",
                "primary-solid": "text-white bg-indigo-500",
                "success-solid": "text-white bg-green-500",
                "warning-solid": "text-white bg-orange-500",
                "info-solid": "text-white bg-blue-500",
                "alert-solid": "text-white bg-red-500",
                "blue-solid": "text-white bg-blue-500",
                "gray-solid": "text-white bg-gray-500",
                "zinc-solid": "text-white bg-zinc-500",
                "white-solid": "text-gray-900 bg-white",
                "unstyled": "border text-gray-300",
            },
            size: {
                sm: "h-[1.2rem] px-1.5 text-xs",
                md: "h-6 px-2 text-xs",
                lg: "h-7 px-3 text-md",
                xl: "h-8 px-4 text-lg",
            },
            tag: {
                false: "font-semibold tracking-wide rounded-full",
                true: "font-semibold border-none rounded-[--radius]",
            },
        },
        defaultVariants: {
            intent: "gray",
            size: "md",
            tag: false,
        },
    }),
    closeButton: cva([
        "UI-Badge__close-button",
        "appearance-none outline-none text-lg -mr-1 cursor-pointer transition ease-in hover:opacity-60",
        "focus-visible:ring-2 focus-visible:ring-[--ring]",
    ]),
    icon: cva([
        "UI-Badge__icon",
        "inline-flex self-center flex-shrink-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Badge
 * -----------------------------------------------------------------------------------------------*/

export type BadgeProps = React.ComponentPropsWithRef<"span"> &
    VariantProps<typeof BadgeAnatomy.root> &
    ComponentAnatomy<typeof BadgeAnatomy> & {
    /**
     * If true, a close button will be rendered.
     */
    isClosable?: boolean,
    /**
     * Callback invoked when the close button is clicked.
     */
    onClose?: () => void,
    /**
     * The left icon element.
     */
    leftIcon?: React.ReactElement
    /**
     * The right icon element.
     */
    rightIcon?: React.ReactElement
    /**
     * The spacing between the icon and the badge content.
     */
    iconSpacing?: React.CSSProperties["marginRight"]
}

export const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>((props, ref) => {

    const {
        children,
        className,
        size,
        intent,
        tag = false,
        isClosable,
        onClose,
        leftIcon,
        rightIcon,
        iconSpacing = "0",
        closeButtonClass,
        iconClass,
        ...rest
    } = props

    return (
        <span
            ref={ref}
            className={cn(BadgeAnatomy.root({ size, intent, tag }), className)}
            {...rest}
        >
            {leftIcon && <span className={cn(BadgeAnatomy.icon(), iconClass)} style={{ marginRight: iconSpacing }}>{leftIcon}</span>}

            {children}

            {rightIcon && <span className={cn(BadgeAnatomy.icon(), iconClass)} style={{ marginLeft: iconSpacing }}>{rightIcon}</span>}

            {isClosable && <button className={cn(BadgeAnatomy.closeButton(), closeButtonClass)} onClick={onClose}>
                <svg
                    xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"
                    fill="currentColor"
                >
                    <path
                        d="M3.72 3.72a.75.75 0 0 1 1.06 0L8 6.94l3.22-3.22a.749.749 0 0 1 1.275.326.749.749 0 0 1-.215.734L9.06 8l3.22 3.22a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215L8 9.06l-3.22 3.22a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L6.94 8 3.72 4.78a.75.75 0 0 1 0-1.06Z"
                    ></path>
                </svg>
            </button>}
        </span>
    )

})

Badge.displayName = "Badge"
