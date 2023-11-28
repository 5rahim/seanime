"use client"

import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React, { useEffect, useRef, useState } from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AvatarAnatomy = defineStyleAnatomy({
    body: cva([
        "UI-Avatar__body",
        "inline-flex rounded-full justify-center align-center flex-shrink-0 bg-gray-400"
    ], {
        variants: {
            size: {
                xs: "w-6 h-6",
                sm: "w-8 h-8",
                md: "w-12 h-12",
                lg: "w-16 h-16",
                xl: "w-24 h-24",
                "2xl": "w-32 h-32",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    fallbackIcon: cva(["UI-Avatar__fallback-icon fill-gray-500"]),
    image: cva([
        "UI-Avatar__image",
        "w-full h-full object-cover rounded-full"
    ]),
    placeholder: cva([
        "UI-Avatar__placeholder",
        "uppercase flex w-full h-full items-center justify-center",
        "bg-gray-600 text-gray-50 font-semibold rounded-full"
    ], {
        variants: {
            size: {
                xs: "text-xs",
                sm: "text-sm",
                md: "text-md",
                lg: "text-lg",
                xl: "text-xl",
                "2xl": "text-2xl",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * Avatar
 * -----------------------------------------------------------------------------------------------*/

export interface AvatarProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof AvatarAnatomy>,
    VariantProps<typeof AvatarAnatomy.body> {
    src?: string | null
    placeholder?: string
}

export const Avatar = React.forwardRef<HTMLDivElement, AvatarProps>((props, ref) => {

    const {
        children,
        className,
        size,
        src,
        placeholder,
        bodyClassName,
        fallbackIconClassName,
        imageClassName,
        placeholderClassName,
        ...rest
    } = props

    const [displayImage, setDisplayImage] = useState(!!src && src?.length > 0)

    useEffect(() => {
        setDisplayImage(!!src && src?.length > 0)
    }, [src])

    const imgRef = useRef<HTMLImageElement>(null)

    return (
        <>
            <div
                className={cn(
                    AvatarAnatomy.body({ size }),
                    bodyClassName,
                    className,
                )}
                {...rest}
                ref={ref}
            >
                {(!displayImage && !placeholder) &&
                    <svg viewBox="0 0 128 128" className={cn(AvatarAnatomy.fallbackIcon(), fallbackIconClassName)}
                         role="img" aria-label="avatar">
                        <path
                            className="fill-gray-200"
                            d="M103,102.1388 C93.094,111.92 79.3504,118 64.1638,118 C48.8056,118 34.9294,111.768 25,101.7892 L25,95.2 C25,86.8096 31.981,80 40.6,80 L87.4,80 C96.019,80 103,86.8096 103,95.2 L103,102.1388 Z"
                        ></path>
                        <path
                            className="fill-gray-200"
                            d="M63.9961647,24 C51.2938136,24 41,34.2938136 41,46.9961647 C41,59.7061864 51.2938136,70 63.9961647,70 C76.6985159,70 87,59.7061864 87,46.9961647 C87,34.2938136 76.6985159,24 63.9961647,24"
                        ></path>
                    </svg>}
                {(!displayImage && placeholder) &&
                    <span
                        className={cn(AvatarAnatomy.placeholder({ size }), placeholderClassName)}>{placeholder}</span>}
                {displayImage && <img
                    ref={imgRef}
                    src={src ?? ""}
                    className={cn(AvatarAnatomy.image(), imageClassName)}
                    onError={e => {
                        e.currentTarget.style.display = "none"
                        setDisplayImage(false)
                    }}
                    onLoad={e => {
                        e.currentTarget.style.display = "block"
                        setDisplayImage(true)
                    }}
                />}
            </div>
        </>
    )

})
