"use client"
import { PAGE_TRANSITION } from "@/components/shared/page-transition"
import { cn } from "@/components/ui/core/styling"
import { motion } from "motion/react"
import React from "react"

type PageWrapperProps = {
    children?: React.ReactNode
    ref?: React.RefObject<HTMLDivElement>
} & React.ComponentPropsWithoutRef<"div">

export function PageWrapper(props: PageWrapperProps) {

    const {
        children,
        className,
        ref,
        ...rest
    } = props

    return (
        <div data-page-wrapper-container ref={ref}>
            <motion.div
                data-page-wrapper
                {...PAGE_TRANSITION}
                {...rest as any}
                className={cn("z-[5] relative", className)}
            >
                {children}
            </motion.div>
        </div>
    )
}
