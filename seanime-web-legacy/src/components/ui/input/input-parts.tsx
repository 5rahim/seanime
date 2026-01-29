import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const InputAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Input__root",
        "flex items-center",
        "w-full rounded-xl",
        "bg-[--paper] border border-[--border] placeholder-gray-400 dark:placeholder-gray-500",
        "disabled:cursor-not-allowed",
        "data-[disable=true]:shadow-none data-[disable=true]:opacity-50",
        "focus:dark:border-gray-600 focus:ring-[0.5px] focus:ring-inset focus:dark:ring-gray-700",
        "outline-0",
        "transition duration-150",
        "shadow-sm",
    ], {
        variants: {
            size: {
                sm: "h-8 px-2 py-1 text-sm",
                md: "h-10 px-3",
                lg: "h-12 px-4 py-3 text-md",
            },
            intent: {
                basic: "hover:border-gray-300 dark:hover:border-gray-600",
                filled: "bg-gray-100 hover:bg-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700 border-transparent focus:bg-white dark:focus:bg-gray-900 shadow-none",
                unstyled: "bg-transparent hover:bg-transparent border-0 shadow-none focus:ring-0 rounded-none p-0 text-base",
            },
            hasError: {
                false: null,
                true: "border-red-500 hover:border-red-200 dark:border-red-500",
            },
            isDisabled: {
                false: null,
                true: "shadow-none pointer-events-none opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800",
            },
            isReadonly: {
                false: null,
                true: "pointer-events-none cursor-not-allowed shadow-sm",
            },
            hasLeftAddon: { true: null, false: null },
            hasRightAddon: { true: null, false: null },
            hasLeftIcon: { true: null, false: null },
            hasRightIcon: { true: null, false: null },
        },
        compoundVariants: [
            { hasLeftAddon: true, className: "border-l-transparent hover:border-l-transparent rounded-l-none" },
            /**/
            { hasRightAddon: true, className: "border-r-transparent hover:border-r-transparent rounded-r-none" },
            /**/
            { hasLeftAddon: false, hasLeftIcon: true, size: "sm", className: "pl-10" },
            { hasLeftAddon: false, hasLeftIcon: true, size: "md", className: "pl-10" },
            { hasLeftAddon: false, hasLeftIcon: true, size: "lg", className: "pl-12" },
            /**/
            { hasRightAddon: false, hasRightIcon: true, size: "sm", className: "pr-10" },
            { hasRightAddon: false, hasRightIcon: true, size: "md", className: "pr-10" },
            { hasRightAddon: false, hasRightIcon: true, size: "lg", className: "pr-12" },
        ],
        defaultVariants: {
            size: "md",
            intent: "basic",
            hasError: false,
            isDisabled: false,
            hasLeftIcon: false,
            hasRightIcon: false,
            hasLeftAddon: false,
            hasRightAddon: false,
        },
    }),
})

export const hiddenInputStyles = cn(
    "appearance-none absolute bottom-0 border-0 w-px h-px p-0 -m-px overflow-hidden whitespace-nowrap [clip:rect(0px,0px,0px,0px)] [overflow-wrap:normal]")

/* -------------------------------------------------------------------------------------------------
 * InputContainer
 * -----------------------------------------------------------------------------------------------*/

export const InputContainerAnatomy = defineStyleAnatomy({
    inputContainer: cva([
        "UI-Input__inputContainer",
        "flex relative",
    ]),
})

export type InputContainerProps = {
    className: React.HTMLAttributes<HTMLDivElement>["className"],
    children?: React.ReactNode
}

export const InputContainer = ({ className, children }: InputContainerProps) => {

    return (
        <div className={cn("UI-Input__inputContainer flex relative", className)}>
            {children}
        </div>
    )
}

/* -------------------------------------------------------------------------------------------------
 * InputStyling
 * -----------------------------------------------------------------------------------------------*/

export type InputStyling = Omit<VariantProps<typeof InputAnatomy.root>,
    "isDisabled" | "hasError" | "hasLeftAddon" | "hasRightAddon" | "hasLeftIcon" | "hasRightIcon"> &
    ComponentAnatomy<typeof InputAddonsAnatomy> &
    ComponentAnatomy<typeof InputContainerAnatomy> & {
    leftAddon?: React.ReactNode
    leftIcon?: React.ReactNode
    rightAddon?: React.ReactNode
    rightIcon?: React.ReactNode
}


/* -------------------------------------------------------------------------------------------------
 * Addons Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const InputAddonsAnatomy = defineStyleAnatomy({
    icon: cva([
        "UI-Input__addons--icon",
        "pointer-events-none absolute inset-y-0 grid place-content-center text-gray-500",
        "dark:text-gray-300 !z-[1]",
    ], {
        variants: {
            size: { sm: "w-10 text-md", md: "w-12 text-lg", lg: "w-14 text-2xl" },
            isLeftIcon: { true: "left-0", false: null },
            isRightIcon: { true: "right-0", false: null },
        },
        defaultVariants: {
            size: "md",
            isLeftIcon: false, isRightIcon: false,
        },
    }),
    addon: cva([
        "UI-Input__addons--addon",
        "bg-gray-50 inline-flex items-center flex-none px-3 border border-gray-300 text-gray-800 shadow-sm text-sm sm:text-md",
        "dark:bg-[--paper] dark:border-[--border] dark:text-gray-300",
    ], {
        variants: {
            size: { sm: "text-sm", md: "text-md", lg: "text-lg" },
            isLeftAddon: { true: "rounded-l-xl border-r-0", false: null },
            isRightAddon: { true: "rounded-r-xl border-l-0", false: null },
            hasLeftIcon: { true: null, false: null },
            hasRightIcon: { true: null, false: null },
        },
        compoundVariants: [
            { size: "sm", hasLeftIcon: true, isLeftAddon: true, className: "pl-10" },
            { size: "sm", hasRightIcon: true, isRightAddon: true, className: "pr-10" },
            { size: "md", hasLeftIcon: true, isLeftAddon: true, className: "pl-10" },
            { size: "md", hasRightIcon: true, isRightAddon: true, className: "pr-10" },
            { size: "lg", hasLeftIcon: true, isLeftAddon: true, className: "pl-12" },
            { size: "lg", hasRightIcon: true, isRightAddon: true, className: "pr-12" },
        ],
        defaultVariants: {
            size: "md",
            isLeftAddon: false, isRightAddon: false, hasLeftIcon: false, hasRightIcon: false,
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * InputIcon
 * -----------------------------------------------------------------------------------------------*/

export type InputIconProps = {
    icon: InputStyling["leftIcon"] | undefined,
    size: InputStyling["size"],
    side: "right" | "left",
    props?: Omit<React.ComponentPropsWithoutRef<"span">, "className">,
    className?: string,
}

export const InputIcon = ({ icon, size = "md", side, props, className }: InputIconProps) => {

    if (!!icon) return <span
        className={cn(InputAddonsAnatomy.icon({ isRightIcon: side === "right", isLeftIcon: side === "left", size }), className)}
        {...props}
    >
        {icon}
    </span>

    return null
}

/* -------------------------------------------------------------------------------------------------
 * InputAddon
 * -----------------------------------------------------------------------------------------------*/

export type InputAddonProps = {
    addon: InputStyling["rightAddon"] | InputStyling["leftAddon"] | undefined,
    rightIcon: InputStyling["leftIcon"] | undefined,
    leftIcon: InputStyling["rightIcon"] | undefined,
    size: InputStyling["size"],
    side: "right" | "left",
    props?: Omit<React.ComponentPropsWithoutRef<"span">, "className">,
    className?: string,
}

export const InputAddon = ({ addon, leftIcon, rightIcon, size = "md", side, props, className }: InputAddonProps) => {

    if (!!addon) return (
        <span
            className={cn(InputAddonsAnatomy.addon({
                isRightAddon: side === "right",
                isLeftAddon: side === "left",
                hasRightIcon: !!rightIcon,
                hasLeftIcon: !!leftIcon,
                size,
            }), className)}
            {...props}
        >
            {addon}
        </span>
    )

    return null

}

/* -------------------------------------------------------------------------------------------------
 * Utils
 * -----------------------------------------------------------------------------------------------*/

export function extractInputPartProps<T extends InputStyling>(props: T) {
    const {
        size,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        inputContainerClass, // class
        iconClass, // class
        addonClass, // class
        ...rest
    } = props

    return [{
        size,
        leftAddon,
        leftIcon,
        rightAddon,
        rightIcon,
        ...rest,
    }, {
        inputContainerProps: {
            className: inputContainerClass,
        },
        leftAddonProps: {
            addon: leftAddon,
            leftIcon,
            rightIcon,
            size,
            side: "left",
            className: addonClass,
        },
        rightAddonProps: {
            addon: rightAddon,
            leftIcon,
            rightIcon,
            size,
            side: "right",
            className: addonClass,
        },
        leftIconProps: {
            icon: leftIcon,
            size,
            side: "left",
            className: iconClass,
        },
        rightIconProps: {
            icon: rightIcon,
            size,
            side: "right",
            className: iconClass,
        },
    }] as [
        Omit<T, "iconClass" | "addonClass" | "inputContainerClass">,
        {
            inputContainerProps: InputContainerProps,
            leftAddonProps: InputAddonProps,
            rightAddonProps: InputAddonProps,
            leftIconProps: InputIconProps,
            rightIconProps: InputIconProps
        }
    ]
}
