import { PAGE_TRANSITION } from "@/components/shared/page-transition"
import { cn } from "@/components/ui/core/styling"
import { motion } from "motion/react"
import React from "react"

type PageWrapperProps = {
    children?: React.ReactNode
} & React.ComponentPropsWithoutRef<"div">

export const PageWrapper = React.forwardRef<HTMLDivElement, PageWrapperProps>((props, ref) => {

    const {
        children,
        className,
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
})
