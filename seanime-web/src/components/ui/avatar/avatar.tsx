"use client"

import * as AvatarPrimitive from "@radix-ui/react-avatar"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"


/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AvatarAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Avatar__root",
        "relative flex shrink-0 overflow-hidden rounded-full",
    ], {
        variants: {
            size: {
                xs: "h-6 w-6",
                sm: "h-8 w-8",
                md: "h-10 w-10",
                lg: "h-14 w-14",
                xl: "h-20 w-20",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    image: cva([
        "UI-Avatar__image",
        "aspect-square h-full w-full",
    ]),
    fallback: cva([
        "UI-Avatar__fallback",
        "flex h-full w-full items-center justify-center rounded-full bg-[--muted] text-white dark:text-gray-800 font-semibold",
    ]),
    fallbackIcon: cva([
        "UI-Avatar__fallback-icon",
        "fill-transparent",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Avatar
 * -----------------------------------------------------------------------------------------------*/

export type AvatarProps =
    React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Root> &
    React.ComponentPropsWithoutRef<typeof AvatarPrimitive.Image> &
    ComponentAnatomy<typeof AvatarAnatomy> &
    VariantProps<typeof AvatarAnatomy.root> & {
    fallback?: React.ReactNode
    imageRef?: React.Ref<HTMLImageElement>
    fallbackRef?: React.Ref<HTMLSpanElement>
}

export const Avatar = React.forwardRef<HTMLImageElement, AvatarProps>((props, ref) => {
    const {
        className,
        children,
        imageRef,
        fallbackRef,
        asChild,
        imageClass,
        fallbackClass,
        fallback,
        fallbackIconClass,
        size,
        ...rest
    } = props
    return (
        <AvatarPrimitive.Root
            ref={ref}
            className={cn(AvatarAnatomy.root({ size }), className)}
        >
            <AvatarPrimitive.Image
                ref={imageRef}
                className={cn(AvatarAnatomy.image(), imageClass)}
                {...rest}
            />
            <AvatarPrimitive.Fallback
                ref={fallbackRef}
                className={cn(AvatarAnatomy.fallback(), fallbackClass)}
            >
                {(!fallback) &&
                    <svg
                        viewBox="0 0 128 128" className={cn(AvatarAnatomy.fallbackIcon(), fallbackIconClass)}
                        role="img" aria-label="avatar"
                    >
                        <path
                            fill="currentColor"
                            d="M103,102.1388 C93.094,111.92 79.3504,118 64.1638,118 C48.8056,118 34.9294,111.768 25,101.7892 L25,95.2 C25,86.8096 31.981,80 40.6,80 L87.4,80 C96.019,80 103,86.8096 103,95.2 L103,102.1388 Z"
                        ></path>
                        <path
                            fill="currentColor"
                            d="M63.9961647,24 C51.2938136,24 41,34.2938136 41,46.9961647 C41,59.7061864 51.2938136,70 63.9961647,70 C76.6985159,70 87,59.7061864 87,46.9961647 C87,34.2938136 76.6985159,24 63.9961647,24"
                        ></path>
                    </svg>}
                {fallback}
            </AvatarPrimitive.Fallback>
        </AvatarPrimitive.Root>
    )
})
Avatar.displayName = "Avatar"
