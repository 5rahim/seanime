import { cn, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const InputAnatomy = defineStyleAnatomy({
    input: cva([
        "UI-Input__input",
        "w-full rounded-[--radius]",
        "bg-[--paper] border-[--border] placeholder-gray-400 dark:placeholder-gray-600",
        "disabled:shadow-none disabled:pointer-events-none disabled:opacity-50 disabled:cursor-not-allowed",
        "focus:border-brand-500 focus:ring-1 focus:ring-[--ring]",
        "outline-none focus:outline-none",
        "transition duration-150",
        "shadow-sm",
    ], {
        variants: {
            size: {
                sm: "px-2 py-1.5 text-sm",
                md: "",
                lg: "px-4 py-3 text-md",
            },
            intent: {
                basic: "hover:border-gray-300 dark:hover:border-gray-600",
                filled: "bg-gray-100 dark:bg-gray-800 border-transparent focus:bg-white",
                unstyled: "bg-transparent hover:bg-transparent border-0 shadow-none focus:ring-0 rounded-none p-0 text-base",
            },
            hasError: {
                false: null,
                true: "border-red-500 hover:border-red-200 dark:border-red-500",
            },
            untouchable: {
                false: null,
                true: "shadow-none pointer-events-none opacity-50 cursor-not-allowed bg-gray-50 dark:bg-gray-800",
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
            untouchable: false,
            hasLeftIcon: false,
            hasRightIcon: false,
            hasLeftAddon: false,
            hasRightAddon: false,
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * InputStyling
 * -----------------------------------------------------------------------------------------------*/

export interface InputStyling
    extends Omit<VariantProps<typeof InputAnatomy.input>, "untouchable" | "hasError" | "hasLeftAddon" | "hasRightAddon" | "hasLeftIcon" | "hasRightIcon"> {
    leftAddon?: string
    leftIcon?: React.ReactNode
    rightAddon?: string
    rightIcon?: React.ReactNode
}

/**
 * @description "flex relative"
 */
export const inputContainerStyle = () => cn("UI-Input__inputContainer flex relative")


/* -------------------------------------------------------------------------------------------------
 * Addons Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const InputAddonsAnatomy = defineStyleAnatomy({
    icon: cva([
        "UI-Input__addons--icon pointer-events-none absolute inset-y-0 grid place-content-center text-gray-500 z-[1]",
        "dark:text-gray-300",
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
        "UI-Input__addons--addon bg-gray-50 inline-flex items-center flex-none px-3 border border-gray-300 text-gray-800 shadow-sm text-sm sm:text-md",
        "dark:bg-gray-700 dark:border-gray-700 dark:text-gray-300",
    ], {
        variants: {
            size: { sm: "text-sm", md: "text-md", lg: "text-lg" },
            isLeftAddon: { true: "rounded-l-md", false: null },
            isRightAddon: { true: "rounded-r-md", false: null },
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

export const InputIcon = (
    { icon, size = "md", side, props }: {
        icon: InputStyling["rightIcon"] | undefined,
        size: InputStyling["size"],
        side: "right" | "left",
        props?: Omit<React.ComponentPropsWithoutRef<"span">, "className">,
    },
) => {

    if (!!icon) return <span
        className={cn(InputAddonsAnatomy.icon({ isRightIcon: side === "right", isLeftIcon: side === "left", size }))}
        {...props}
    >
        {icon}
    </span>

    return null
}

/* -------------------------------------------------------------------------------------------------
 * InputAddon
 * -----------------------------------------------------------------------------------------------*/

export const InputAddon = (
    { addon, leftIcon, rightIcon, size = "md", side, props }: {
        addon: InputStyling["rightAddon"] | InputStyling["leftAddon"] | undefined,
        rightIcon: InputStyling["leftIcon"] | undefined,
        leftIcon: InputStyling["rightIcon"] | undefined,
        size: InputStyling["size"],
        side: "right" | "left",
        props?: Omit<React.ComponentPropsWithoutRef<"span">, "className">,
    },
) => {

    if (!!addon) return (
        <span
            className={cn(InputAddonsAnatomy.addon({
                isRightAddon: side === "right",
                isLeftAddon: side === "left",
                hasRightIcon: !!rightIcon,
                hasLeftIcon: !!leftIcon,
                size,
            }))}
            {...props}
        >
            {addon}
        </span>
    )

    return null

}
