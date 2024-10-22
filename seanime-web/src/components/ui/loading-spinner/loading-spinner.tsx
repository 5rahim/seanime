import { cva } from "class-variance-authority"
import React from "react"
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
        "text-gray-200 dark:text-gray-600 fill-brand-500",
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
        <svg
            aria-hidden="true"
            className={cn(
                LoadingSpinnerAnatomy.icon(),
                className,
            )}
            viewBox="0 0 100 101"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            ref={ref}
            {...rest}
        >
            <path
                d="M100 50.5908C100 78.2051 77.6142 100.591 50 100.591C22.3858 100.591 0 78.2051 0 50.5908C0 22.9766 22.3858 0.59082 50 0.59082C77.6142 0.59082 100 22.9766 100 50.5908ZM9.08144 50.5908C9.08144 73.1895 27.4013 91.5094 50 91.5094C72.5987 91.5094 90.9186 73.1895 90.9186 50.5908C90.9186 27.9921 72.5987 9.67226 50 9.67226C27.4013 9.67226 9.08144 27.9921 9.08144 50.5908Z"
                fill="currentColor"
            />
            <path
                d="M93.9676 39.0409C96.393 38.4038 97.8624 35.9116 97.0079 33.5539C95.2932 28.8227 92.871 24.3692 89.8167 20.348C85.8452 15.1192 80.8826 10.7238 75.2124 7.41289C69.5422 4.10194 63.2754 1.94025 56.7698 1.05124C51.7666 0.367541 46.6976 0.446843 41.7345 1.27873C39.2613 1.69328 37.813 4.19778 38.4501 6.62326C39.0873 9.04874 41.5694 10.4717 44.0505 10.1071C47.8511 9.54855 51.7191 9.52689 55.5402 10.0491C60.8642 10.7766 65.9928 12.5457 70.6331 15.2552C75.2735 17.9648 79.3347 21.5619 82.5849 25.841C84.9175 28.9121 86.7997 32.2913 88.1811 35.8758C89.083 38.2158 91.5421 39.6781 93.9676 39.0409Z"
                fill="currentFill"
            />
        </svg>
    )

})
