import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AlertAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Alert__root",
        "py-3 px-4 flex justify-between rounded-xl border text-sm transition-colors duration-200",
    ], {
        variants: {
            intent: {
                "info": "bg-blue-500/10 border-blue-500/20 text-gray-900 dark:bg-blue-900/70 dark:border-[--border] dark:text-gray-200",
                "success": "bg-green-500/10 border-green-500/20 text-gray-900 dark:bg-green-900/70 dark:border-[--border] dark:text-gray-200",
                "warning": "bg-orange-500/10 border-orange-500/20 text-gray-900 dark:bg-orange-900/70 dark:border-[--border] dark:text-gray-200",
                "alert": "bg-red-500/10 border-red-500/20 text-gray-900 dark:bg-red-900/70 dark:border-[--border] dark:text-gray-200",

                "info-basic": "bg-white text-gray-800 border-gray-200 dark:bg-gray-900/50 dark:border-[--border] dark:text-gray-200",
                "success-basic": "bg-white text-gray-800 border-gray-200 dark:bg-gray-900/50 dark:border-[--border] dark:text-gray-200",
                "warning-basic": "bg-white text-gray-800 border-gray-200 dark:bg-gray-900/50 dark:border-[--border] dark:text-gray-200",
                "alert-basic": "bg-white text-gray-800 border-gray-200 dark:bg-gray-900/50 dark:border-[--border] dark:text-gray-200",
            },
        },
        defaultVariants: {
            intent: "info",
        },
    }),
    detailsContainer: cva([
        "UI-Alert__detailsContainer",
        "flex w-full items-start",
    ]),
    textContainer: cva([
        "UI-Alert__textContainer",
        "flex flex-col self-start ml-3 gap-0.5",
    ]),
    title: cva([
        "UI-Alert__title",
        "font-semibold text-gray-900 dark:text-[--foreground] mb-0.5",
    ]),
    description: cva([
        "UI-Alert__description",
        "text-xs md:text-sm text-[--muted] dark:text-gray-200",
    ]),
    icon: cva([
        "UI-Alert__icon",
        "text-xl content-evenly flex-none self-start mt-0.5",
    ], {
        variants: {
            intent: {
                "info-basic": "text-blue-500 dark:text-blue-400",
                "success-basic": "text-green-500 dark:text-green-400",
                "warning-basic": "text-orange-500 dark:text-orange-400",
                "alert-basic": "text-red-500 dark:text-red-400",
                "info": "text-blue-600 dark:text-blue-400",
                "success": "text-green-600 dark:text-green-400",
                "warning": "text-orange-600 dark:text-orange-400",
                "alert": "text-red-600 dark:text-red-400",
            },
        },
        defaultVariants: {
            intent: "info-basic",
        },
    }),
    closeButton: cva([
        "UI-Alert__closeButton",
        "flex-none self-start text-lg hover:opacity-80 active:opacity-100 transition-opacity duration-150 cursor-pointer h-5 w-5 ml-4 opacity-50 dark:text-gray-400 dark:hover:text-gray-200",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Alert
 * -----------------------------------------------------------------------------------------------*/

export type AlertProps = React.ComponentPropsWithRef<"div"> &
    VariantProps<typeof AlertAnatomy.root> &
    ComponentAnatomy<typeof AlertAnatomy> & {
    /**
     * The title of the alert
     */
    title?: string,
    /**
     * The description text or content of the alert
     */
    description?: React.ReactNode
    /**
     * Replace the default icon with a custom icon
     *
     * - `iconClass` does not apply to custom icons
     */
    icon?: React.ReactNode
    /**
     * If true, a close button will be rendered
     */
    isClosable?: boolean
    /**
     * Callback invoked when the close button is clicked
     */
    onClose?: () => void
}

export const Alert = React.forwardRef<HTMLDivElement, AlertProps>((props, ref) => {

    const {
        children,
        className,
        title,
        description,
        isClosable,
        onClose,
        intent = "info-basic",
        iconClass,
        detailsContainerClass,
        textContainerClass,
        titleClass,
        descriptionClass,
        closeButtonClass,
        icon,
        ...rest
    } = props

    let Icon: any = null

    if (intent === "info-basic" || intent === "info") {
        Icon = <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <circle cx="12" cy="12" r="10"></circle>
            <path d="M12 16v-4"></path>
            <path d="M12 8h.01"></path>
        </svg>
    } else if (intent === "alert-basic" || intent === "alert") {
        Icon = <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" x2="12" y1="8" y2="12"></line>
            <line x1="12" x2="12.01" y1="16" y2="16"></line>
        </svg>
    } else if (intent === "warning-basic" || intent === "warning") {
        Icon = <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"></path>
            <line x1="12" x2="12" y1="9" y2="13"></line>
            <line x1="12" x2="12.01" y1="17" y2="17"></line>
        </svg>
    } else if (intent === "success-basic" || intent === "success") {
        Icon = <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-5 w-5"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
        >
            <path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2 2 6.477 2 12s4.477 10 10 10z"></path>
            <path d="m9 12 2 2 4-4"></path>
        </svg>
    }

    return (
        <div
            className={cn(
                AlertAnatomy.root({ intent }),
                className,
            )}
            {...rest}
            ref={ref}
        >
            <div className={cn(AlertAnatomy.detailsContainer(), detailsContainerClass)}>
                {icon ? icon : <div className={cn(AlertAnatomy.icon({ intent: intent }), iconClass)}>
                    {Icon && Icon}
                </div>}
                <div className={cn(AlertAnatomy.textContainer(), textContainerClass)}>
                    {!!title && <span className={cn(AlertAnatomy.title(), titleClass)}>
                        {title}
                    </span>}
                    {!!(description || children) && <div className={cn(AlertAnatomy.description(), descriptionClass)}>
                        {description || children}
                    </div>}
                </div>
            </div>
            {onClose && <button className={cn(AlertAnatomy.closeButton(), closeButtonClass)} onClick={onClose}>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-4 w-4"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                >
                    <line x1="18" x2="6" y1="6" y2="18"></line>
                    <line x1="6" x2="18" y1="6" y2="18"></line>
                </svg>
            </button>}
        </div>
    )

})

Alert.displayName = "Alert"
