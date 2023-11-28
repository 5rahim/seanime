import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva, VariantProps } from "class-variance-authority"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PageHeaderAnatomy = defineStyleAnatomy({
    body: cva("UI-PageHeader__body md:flex md:items-center md:justify-between space-y-2 md:space-y-0 md:space-x-5"),
    title: cva("UI-PageHeader__title font-bold text-gray-900 dark:text-gray-200", {
        variants: {
            size: {
                sm: "text-lg sm:text-xl",
                md: "text-xl sm:text-2xl",
                lg: "text-2xl sm:text-3xl",
                xl: "text-2xl sm:text-4xl",
            },
        },
        defaultVariants: {
            size: "xl",
        },
    }),
    actionContainer: cva([
        "UI-PageHeader__actionContainer",
        "justify-stretch flex flex-col-reverse space-y-4 space-y-reverse sm:flex-row-reverse sm:justify-end",
        "sm:space-y-0 sm:space-x-3 sm:space-x-reverse md:mt-0 md:flex-row md:space-x-3"
    ]),
    description: cva([
        "UI-PageHeader__description",
        "text-sm font-medium text-gray-500 dark:text-gray-400"
    ]),
    detailsContainer: cva([
        "UI-PageHeader__detailsContainer",
        "block sm:flex items-start sm:space-x-5"
    ], {
        variants: {
            withImage: {
                true: "flex-col gap-2 sm:flex-row sm:gap-6",
                false: null,
            },
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * PageHeader
 * -----------------------------------------------------------------------------------------------*/

export interface PageHeaderProps extends React.ComponentPropsWithRef<"header">,
    ComponentWithAnatomy<typeof PageHeaderAnatomy>,
    VariantProps<typeof PageHeaderAnatomy.title> {
    title?: string
    description?: string
    action?: React.ReactNode
    image?: React.ReactNode
}

export const PageHeader = React.forwardRef<HTMLDivElement, PageHeaderProps>((props, ref) => {

    const {
        children,
        className,
        size = "xl",
        title,
        description,
        action,
        image,
        titleClassName,
        actionContainerClassName,
        descriptionClassName,
        detailsContainerClassName,
        bodyClassName,
        ...rest
    } = props

    return (
        <>
            <header
                aria-label={title}
                className={cn(
                    PageHeaderAnatomy.body(),
                    bodyClassName,
                    className,
                )}
                ref={ref}
                {...rest}
            >
                <div className={cn(PageHeaderAnatomy.detailsContainer(), detailsContainerClassName)}>
                    {image && <div className="flex-shrink-0">
                        <div className="relative">
                            {image}
                        </div>
                    </div>}
                    <div className="">
                        <h1 className={cn(PageHeaderAnatomy.title({ size }), titleClassName)}>{title}</h1>
                        {description && <p className={cn(PageHeaderAnatomy.description(), descriptionClassName)}>
                            {description}
                        </p>}
                    </div>
                </div>
                {!!action && <div className={cn(PageHeaderAnatomy.actionContainer(), actionContainerClassName)}>
                    {action}
                </div>}
            </header>
        </>
    )

})

PageHeader.displayName = "PageHeader"
