import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const AlertAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Alert__root",
        "py-3 px-4 flex justify-between rounded-[--radius]",
    ], {
        variants: {
            intent: {
                "info": "bg-blue-50 text-blue-500 dark:bg-opacity-10 dark:text-blue-200",
                "success": "bg-green-50 text-green-500 dark:bg-opacity-10 dark:text-green-200",
                "warning": "bg-orange-50 text-orange-500 dark:bg-opacity-10 dark:text-orange-200",
                "alert": "bg-red-50 text-red-500 dark:bg-opacity-10 dark:text-red-200",
                "info-basic": "bg-white text-gray-800 border dark:bg-gray-800 dark:text-gray-200",
                "success-basic": "bg-white text-gray-800 border dark:bg-gray-800 dark:text-gray-200",
                "warning-basic": "bg-white text-gray-800 border dark:bg-gray-800 dark:text-gray-200",
                "alert-basic": "bg-white text-gray-800 border dark:bg-gray-800 dark:text-gray-200",
            },
        },
        defaultVariants: {
            intent: "info",
        },
    }),
    detailsContainer: cva([
        "UI-Alert__detailsContainer",
        "flex",
    ]),
    textContainer: cva([
        "UI-Alert__textContainer",
        "flex flex-col self-center ml-3 gap-.5",
    ]),
    title: cva([
        "UI-Alert__title",
        "font-bold",
    ]),
    description: cva([
        "UI-Alert__description",
    ]),
    icon: cva([
        "UI-Alert__icon",
        "text-2xl content-evenly",
    ], {
        variants: {
            intent: {
                "info-basic": "text-blue-500",
                "success-basic": "text-green-500",
                "warning-basic": "text-orange-500",
                "alert-basic": "text-red-500",
                "info": "text-blue-500",
                "success": "text-green-500",
                "warning": "text-orange-500",
                "alert": "text-red-500",
            },
        },
        defaultVariants: {
            intent: "info-basic",
        },
    }),
    closeButton: cva([
        "UI-Alert__closeButton",
        "flex-none self-start text-2xl hover:opacity-50 transition ease-in cursor-pointer h-5 w-5",
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
            width="24"
            height="24"
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
            width="24"
            height="24"
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
            width="24"
            height="24"
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
            width="24"
            height="24"
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
                    <span className={cn(AlertAnatomy.title(), titleClass)}>
                        {title}
                    </span>
                    {!!(description || children) && <div className={cn(AlertAnatomy.description(), descriptionClass)}>
                        {description || children}
                    </div>}
                </div>
            </div>
            {onClose && <button className={cn(AlertAnatomy.closeButton(), closeButtonClass)} onClick={onClose}>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
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
