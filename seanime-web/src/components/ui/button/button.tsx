import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const ButtonAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Button_root",
        "whitespace-nowrap font-medium rounded-lg",
        "inline-flex items-center text-white text-center justify-center",
        "focus-visible:outline-none focus-visible:ring-1 ring-offset-1 ring-offset-[--background] focus-visible:ring-white/40",
        "disabled:opacity-50 disabled:pointer-events-none disabled:transform-none",
    ], {
        variants: {
            intent: {
                "primary": "text-white border bg-brand-500 border-brand-400/20 active:bg-opacity-100 dark:bg-opacity-70 dark:hover:bg-opacity-90",
                "primary-outline": "text-[--brand] border border-brand-500/40 bg-transparent hover:bg-brand-500/10 active:bg-brand-500/20 dark:border-brand-400/30 dark:hover:bg-brand-500/15 dark:active:bg-brand-500/25",
                "primary-subtle": "text-[--brand] border bg-brand-50 border-transparent --border-brand-300/10 hover:bg-brand-100 active:bg-brand-200 dark:bg-opacity-10 dark:hover:bg-opacity-20",
                "primary-link": "text-[--brand] border border-transparent bg-transparent hover:underline active:text-brand-700 dark:active:text-brand-300",
                "primary-basic": "text-[--brand] border border-transparent bg-transparent hover:bg-brand-100 active:bg-brand-200 dark:hover:bg-opacity-10 dark:active:text-brand-300",

                "warning": "text-white border bg-orange-500 border-orange-400/20 active:bg-opacity-100 dark:bg-opacity-85 dark:hover:bg-opacity-90",
                "warning-outline": "text-[--orange] border border-orange-500/40 bg-transparent hover:bg-orange-500/10 active:bg-orange-500/20 dark:border-orange-400/30 dark:hover:bg-orange-500/15 dark:active:bg-orange-500/25",
                "warning-subtle": "text-[--orange] border bg-orange-50 border-transparent --border-orange-300/10 hover:bg-orange-100 active:bg-orange-200 dark:bg-opacity-10 dark:hover:bg-opacity-20",
                "warning-link": "text-[--orange] border border-transparent bg-transparent hover:underline active:text-orange-700 dark:active:text-orange-300",
                "warning-basic": "text-[--orange] border border-transparent bg-transparent hover:bg-orange-100 active:bg-orange-200 dark:hover:bg-opacity-10 dark:active:text-orange-300",

                "success": "text-white border bg-green-500 border-green-400/20 active:bg-opacity-100 dark:bg-opacity-85 dark:hover:bg-opacity-90",
                "success-outline": "text-[--green] border border-green-500/40 bg-transparent hover:bg-green-500/10 active:bg-green-500/20 dark:border-green-400/30 dark:hover:bg-green-500/15 dark:active:bg-green-500/25",
                "success-subtle": "text-[--green] border bg-green-50 border-transparent --border-green-300/10 hover:bg-green-100 active:bg-green-200 dark:bg-opacity-10 dark:hover:bg-opacity-20",
                "success-link": "text-[--green] border border-transparent bg-transparent hover:underline active:text-green-700 dark:active:text-green-300",
                "success-basic": "text-[--green] border border-transparent bg-transparent hover:bg-green-100 active:bg-green-200 dark:hover:bg-opacity-10 dark:active:text-green-300",

                "alert": "text-white border bg-red-500 border-red-400/20 active:bg-opacity-100 dark:bg-opacity-85 dark:hover:bg-opacity-90",
                "alert-outline": "text-[--red] border border-red-500/40 bg-transparent hover:bg-red-500/10 active:bg-red-500/20 dark:border-red-400/30 dark:hover:bg-red-500/15 dark:active:bg-red-500/25",
                "alert-subtle": "text-[--red] border bg-red-50 border-transparent --border-red-300/10 hover:bg-red-100 active:bg-red-200 dark:bg-opacity-10 dark:hover:bg-opacity-20",
                "alert-link": "text-[--red] border border-transparent bg-transparent hover:underline active:text-red-700 dark:active:text-red-300",
                "alert-basic": "text-[--red] border border-transparent bg-transparent hover:bg-red-100 active:bg-red-200 dark:hover:bg-opacity-10 dark:active:text-red-300",

                "gray": "bg-gray-500 hover:bg-gray-600 active:bg-gray-700 border border-transparent",
                "gray-outline": "text-gray-600 border border-gray-500/30 bg-transparent hover:bg-gray-500/10 active:bg-gray-500/20 dark:text-gray-300 dark:border-gray-400/20 dark:hover:bg-gray-500/15 dark:active:bg-gray-500/25",
                "gray-subtle": "text-[--gray] border bg-gray-100 border-transparent --border-gray-500/20 hover:bg-gray-200 active:bg-gray-300 dark:text-gray-100 dark:bg-opacity-10 dark:hover:bg-opacity-20",
                "gray-link": "text-[--gray] border border-transparent bg-transparent hover:underline active:text-gray-700 dark:text-gray-300 dark:active:text-gray-200",
                "gray-basic": "text-[--gray] border border-transparent bg-transparent hover:bg-gray-100 active:bg-gray-200 dark:active:bg-opacity-20 dark:text-gray-200 dark:hover:bg-opacity-10 dark:active:text-gray-200",

                "white": "text-[#000] bg-white hover:bg-gray-200 active:bg-gray-300 border border-transparent",
                "white-outline": "text-white border border-white/30 bg-transparent hover:bg-white/10 active:bg-white/20",
                "white-subtle": "text-white bg-white bg-opacity-15 hover:bg-opacity-20 border border-transparent --border-white/10 active:bg-opacity-25",
                "white-link": "text-white border border-transparent bg-transparent hover:underline active:text-gray-200",
                "white-basic": "text-white border border-transparent bg-transparent hover:bg-white hover:bg-opacity-15 active:bg-opacity-20 active:text-white-300",
            },
            rounded: {
                true: "rounded-full",
                false: null,
            },
            contentWidth: {
                true: "w-fit",
                false: null,
            },
            size: {
                xs: "text-xs h-6 px-2",
                sm: "text-xs h-8 px-2.5",
                md: "text-sm h-9 px-3",
                lg: "text-sm h-10 px-4",
                xl: "text-base h-12 px-6",
            },
        },
        defaultVariants: {
            intent: "primary",
            size: "md",
        },
    }),
    icon: cva([
        "UI-Button__icon",
        "inline-flex self-center flex-shrink-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Button
 * -----------------------------------------------------------------------------------------------*/


export type ButtonProps = React.ComponentPropsWithoutRef<"button"> &
    VariantProps<typeof ButtonAnatomy.root> &
    ComponentAnatomy<typeof ButtonAnatomy> & {
    loading?: boolean,
    leftIcon?: React.ReactNode
    rightIcon?: React.ReactNode
    iconSpacing?: React.CSSProperties["marginInline"]
    hideTextOnSmallScreen?: boolean
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>((props, ref) => {

    const {
        children,
        size,
        className,
        rounded = false,
        contentWidth = false,
        intent,
        leftIcon,
        rightIcon,
        iconSpacing = "0.5rem",
        loading,
        iconClass,
        disabled,
        hideTextOnSmallScreen,
        ...rest
    } = props

    const hasCustomAnimation = className?.includes("animate-")
    const isIconButton = className?.includes("UI-IconButton_root")

    return (
        <button
            type="button"
            className={cn(
                ButtonAnatomy.root({
                    size,
                    intent,
                    rounded,
                    contentWidth,
                }),
                !hasCustomAnimation && (
                    isIconButton
                        ? "transition-all duration-150 ease-[cubic-bezier(0.25,1,0.5,1)] motion-safe:hover:scale-[1.04] motion-safe:active:scale-[0.96]"
                        : "transition-all duration-150 ease-[cubic-bezier(0.25,1,0.5,1)] motion-safe:hover:scale-[1.01] motion-safe:active:scale-[0.98]"
                ),
                className,
            )}
            disabled={disabled || loading}
            aria-disabled={disabled}
            {...rest}
            ref={ref}
        >
            {loading ? (
                <>
                    <svg
                        width="15"
                        height="15"
                        fill="currentColor"
                        className="animate-spin"
                        viewBox="0 0 1792 1792"
                        xmlns="http://www.w3.org/2000/svg"
                        style={{ marginInlineEnd: !hideTextOnSmallScreen ? iconSpacing : 0 }}
                    >
                        <path
                            d="M526 1394q0 53-37.5 90.5t-90.5 37.5q-52 0-90-38t-38-90q0-53 37.5-90.5t90.5-37.5 90.5 37.5 37.5 90.5zm498 206q0 53-37.5 90.5t-90.5 37.5-90.5-37.5-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm-704-704q0 53-37.5 90.5t-90.5 37.5-90.5-37.5-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm1202 498q0 52-38 90t-90 38q-53 0-90.5-37.5t-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm-964-996q0 66-47 113t-113 47-113-47-47-113 47-113 113-47 113 47 47 113zm1170 498q0 53-37.5 90.5t-90.5 37.5-90.5-37.5-37.5-90.5 37.5-90.5 90.5-37.5 90.5 37.5 37.5 90.5zm-640-704q0 80-56 136t-136 56-136-56-56-136 56-136 136-56 136 56 56 136zm530 206q0 93-66 158.5t-158 65.5q-93 0-158.5-65.5t-65.5-158.5q0-92 65.5-158t158.5-66q92 0 158 66t66 158z"
                        >
                        </path>
                    </svg>
                    {children}
                </>
            ) : <>
                {leftIcon &&
                    <span
                        className={cn(ButtonAnatomy.icon(), iconClass)}
                        style={{ marginInlineEnd: !hideTextOnSmallScreen ? iconSpacing : 0 }}
                    >
                        {leftIcon}
                    </span>}
                <span
                    className={cn(
                        hideTextOnSmallScreen && cn(
                            "hidden",
                            leftIcon && "pl-[0.5rem]",
                            rightIcon && "pr-[0.5rem]",
                        ),
                        "md:inline-block",
                    )}
                >
                    {children}
                </span>
                {rightIcon &&
                    <span
                        className={cn(ButtonAnatomy.icon(), iconClass)}
                        style={{ marginInlineStart: !hideTextOnSmallScreen ? iconSpacing : 0 }}
                    >
                        {rightIcon}
                    </span>}
            </>}
        </button>
    )

})

Button.displayName = "Button"
