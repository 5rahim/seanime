import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { cva, VariantProps } from "class-variance-authority"
import * as React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const PageHeaderAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-PageHeader__root",
        "md:flex md:items-center md:justify-between space-y-2 md:space-y-0 md:space-x-5",
    ]),
    title: cva([
        "UI-PageHeader__title",
        "font-bold text-gray-900 dark:text-gray-200",
    ], {
        variants: {
            size: {
                sm: "text-lg sm:text-xl",
                md: "text-2xl sm:text-3xl",
                lg: "text-3xl sm:text-4xl",
                xl: "text-4xl sm:text-5xl",
            },
        },
        defaultVariants: {
            size: "md",
        },
    }),
    actionContainer: cva([
        "UI-PageHeader__actionContainer",
        "justify-stretch flex flex-col-reverse space-y-4 space-y-reverse sm:flex-row-reverse sm:justify-end",
        "sm:space-y-0 sm:space-x-3 sm:space-x-reverse md:mt-0 md:flex-row md:space-x-3",
    ]),
    textContainer: cva([
        "UI-PageHeader__textContainer",
        "space-y-1",
    ]),
    description: cva([
        "UI-PageHeader__description",
        "text-sm font-medium text-gray-500 dark:text-gray-400",
    ]),
    detailsContainer: cva([
        "UI-PageHeader__detailsContainer",
        "block sm:flex items-start sm:space-x-5",
    ], {
        variants: {
            _withImage: {
                true: "flex-col sm:flex-row",
                false: null,
            },
        },
    }),
})

/* -------------------------------------------------------------------------------------------------
 * PageHeader
 * -----------------------------------------------------------------------------------------------*/

export type PageHeaderProps = React.ComponentPropsWithRef<"header"> &
    ComponentAnatomy<typeof PageHeaderAnatomy> &
    VariantProps<typeof PageHeaderAnatomy.title> & {
    /**
     * Page title.
     */
    title?: string
    /**
     * Page description.
     */
    description?: string
    /**
     * Elements rendered in the action container.
     */
    action?: React.ReactNode
    /**
     * Image elements rendered next to the title and description.
     */
    image?: React.ReactNode
}

export const PageHeader = React.forwardRef<HTMLDivElement, PageHeaderProps>((props, ref) => {

    const {
        children,
        className,
        size = "md",
        title,
        description,
        action,
        image,
        titleClass,
        actionContainerClass,
        descriptionClass,
        detailsContainerClass,
        textContainerClass,
        ...rest
    } = props

    return (
        <header
            ref={ref}
            aria-label={title}
            className={cn(
                PageHeaderAnatomy.root(),
                className,
            )}
            {...rest}
        >
            <div className={cn(PageHeaderAnatomy.detailsContainer({ _withImage: !!image }), detailsContainerClass)}>
                {image && <div className="flex-shrink-0">
                    <div className="relative">
                        {image}
                    </div>
                </div>}
                <div className={cn(PageHeaderAnatomy.textContainer(), textContainerClass)}>
                    <h1 className={cn(PageHeaderAnatomy.title({ size }), titleClass)}>{title}</h1>
                    {description && <p className={cn(PageHeaderAnatomy.description(), descriptionClass)}>
                        {description}
                    </p>}
                </div>
            </div>
            {!!action && <div className={cn(PageHeaderAnatomy.actionContainer(), actionContainerClass)}>
                {action}
            </div>}
        </header>
    )

})

PageHeader.displayName = "PageHeader"
