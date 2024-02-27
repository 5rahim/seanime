import * as React from "react"
import { IconButton, IconButtonProps } from "."
import { cn } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * CloseButton
 * -----------------------------------------------------------------------------------------------*/

export type CloseButtonProps = Omit<IconButtonProps, "icon"> & {
    icon?: React.ReactNode
}

export const CloseButton = React.forwardRef<HTMLButtonElement, CloseButtonProps>((props, ref) => {

    const {
        className,
        icon,
        ...rest
    } = props

    return (
        <IconButton
            type="button"
            intent="gray-basic"
            size="sm"
            className={cn(
                "rounded-full text-2xl flex-none",
                className,
            )}
            icon={!icon ? <span>
                <svg
                    xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" width="16" height="16"
                    fill="currentColor"
                >
                    <path
                        d="M3.72 3.72a.75.75 0 0 1 1.06 0L8 6.94l3.22-3.22a.749.749 0 0 1 1.275.326.749.749 0 0 1-.215.734L9.06 8l3.22 3.22a.749.749 0 0 1-.326 1.275.749.749 0 0 1-.734-.215L8 9.06l-3.22 3.22a.751.751 0 0 1-1.042-.018.751.751 0 0 1-.018-1.042L6.94 8 3.72 4.78a.75.75 0 0 1 0-1.06Z"
                    ></path>
                </svg>
            </span> : icon}
            {...rest}
            ref={ref}
        />
    )

})

CloseButton.displayName = "CloseButton"
