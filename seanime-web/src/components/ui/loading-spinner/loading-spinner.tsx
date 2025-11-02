import { cva } from "class-variance-authority"
import React from "react"
import { ImSpinner2 } from "react-icons/im"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const LoadingSpinnerAnatomy = defineStyleAnatomy({
    container: cva([
        "UI-LoadingSpinner__container",
        "flex flex-col w-full items-center h-24 justify-center",
    ]),
    icon: cva([
        "UI-LoadingSpinner__icon",
        "inline w-10 h-10 mr-2 animate-spin",
    ]),
    title: cva([
        "UI-LoadingSpinner__title",
        "text-base font-medium text-[--foreground] py-2",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * LoadingSpinner
 * -----------------------------------------------------------------------------------------------*/

export type LoadingSpinnerProps = React.ComponentPropsWithRef<"div"> & ComponentAnatomy<typeof LoadingSpinnerAnatomy> & {
    spinner?: React.ReactNode
}

export const LoadingSpinner = React.forwardRef<HTMLDivElement, LoadingSpinnerProps>((props, ref) => {

    const {
        children,
        className,
        containerClass,
        iconClass,
        spinner,
        title,
        ...rest
    } = props

    return (
        <div
            className={cn(
                LoadingSpinnerAnatomy.container(),
                containerClass,
            )}
            {...rest}
            ref={ref}
        >
            {spinner ? spinner : <Spinner className={iconClass} />}
            {title && <p className={LoadingSpinnerAnatomy.title()}>{title}</p>}
        </div>
    )

})

LoadingSpinner.displayName = "LoadingSpinner"


/* -------------------------------------------------------------------------------------------------
 * Spinner
 * -----------------------------------------------------------------------------------------------*/

interface SpinnerProps extends React.ComponentPropsWithRef<"svg"> {
    children?: React.ReactNode
}

export const Spinner = React.forwardRef<SVGSVGElement, SpinnerProps>((props, ref) => {

    const { children, className, ...rest } = props

    return (
        <ImSpinner2
            className={cn(
                LoadingSpinnerAnatomy.icon(),
                className,
            )}
        />
    )

})
