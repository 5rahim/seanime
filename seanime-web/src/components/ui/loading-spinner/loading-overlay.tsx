import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"
import { LoadingSpinner } from "./loading-spinner"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const LoadingOverlayAnatomy = defineStyleAnatomy({
    overlay: cva([
        "UI-LoadingOverlay__overlay overflow-hidden",
        "absolute bg-[--background]/50 w-full h-full z-10 inset-0 pt-4 flex flex-col items-center justify-center backdrop-blur-sm",
        "!mt-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * LoadingOverlay
 * -----------------------------------------------------------------------------------------------*/

export type LoadingOverlayProps = {
    children?: React.ReactNode
    /**
     * Whether to show the loading spinner
     */
    showSpinner?: boolean
    /**
     * If true, the loading overlay will be unmounted
     */
    hide?: boolean
    className?: string
}

export const LoadingOverlay = React.forwardRef<HTMLDivElement, LoadingOverlayProps>((props, ref) => {

    const {
        children,
        hide = false,
        showSpinner = true,
        className,
        ...rest
    } = props

    if (hide) return null

    return (
        <div
            ref={ref}
            className={cn(LoadingOverlayAnatomy.overlay(), className)}
            {...rest}
        >
            {showSpinner && <LoadingSpinner className="justify-auto" />}
            {children}
        </div>
    )

})

LoadingOverlay.displayName = "LoadingOverlay"
