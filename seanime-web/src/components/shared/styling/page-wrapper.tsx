"use client"
import { PAGE_TRANSITION } from "@/components/shared/styling/page-transition"
import { motion } from "framer-motion"
import React from "react"

type PageWrapperProps = {
    children?: React.ReactNode
} & React.ComponentPropsWithoutRef<"div">

export function PageWrapper(props: PageWrapperProps) {

    const {
        children,
        ...rest
    } = props

    return (
        <motion.div
            {...PAGE_TRANSITION}
            {...rest as any}
        >
            {children}
        </motion.div>
    )
}
