import { SeaLink } from "@/components/shared/sea-link"
import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const StaticTabsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-StaticTabs__root",
        "flex w-full overflow-hidden overflow-x-auto",
    ]),
    trigger: cva([
        "UI-StaticTabs__trigger",
        "group/staticTabs__trigger inline-flex flex-none shrink-0 basis-auto items-center font-medium text-sm transition outline-none min-w-0 justify-center",
        "text-[--muted] hover:text-[--foreground]",
        "h-10 px-4 rounded-full",
        "data-[current=true]:bg-[--subtle] data-[current=true]:font-semibold data-[current=true]:text-[--foreground]",
        "focus-visible:bg-[--subtle]",
    ]),
    icon: cva([
        "UI-StaticTabs__icon",
        "-ml-0.5 mr-2 h-4 w-4",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * StaticTabs
 * -----------------------------------------------------------------------------------------------*/

export type StaticTabsItem = {
    name: string,
    href?: string | null | undefined,
    iconType?: React.ElementType,
    onClick?: () => void,
    isCurrent: boolean
    addon?: React.ReactNode,
}

export type StaticTabsProps = React.ComponentPropsWithRef<"nav"> &
    ComponentAnatomy<typeof StaticTabsAnatomy> & {
    items: StaticTabsItem[]
}

export const StaticTabs = React.forwardRef<HTMLElement, StaticTabsProps>((props, ref) => {

    const {
        children,
        className,
        triggerClass,
        iconClass,
        items,
        ...rest
    } = props

    return (
        <nav
            ref={ref}
            className={cn(StaticTabsAnatomy.root(), className)}
            role="navigation"
            {...rest}
        >
            {items.map((tab) => !!tab.href ? (
                <SeaLink
                    key={tab.name}
                    href={tab.href ?? "#"}
                    className={cn(
                        StaticTabsAnatomy.trigger(),
                        triggerClass,
                    )}
                    aria-current={tab.isCurrent ? "page" : undefined}
                    data-current={tab.isCurrent}
                >
                    {tab.iconType && <tab.iconType
                        className={cn(
                            StaticTabsAnatomy.icon(),
                            iconClass,
                        )}
                        aria-hidden="true"
                        data-current={tab.isCurrent}
                    />}
                    <span>{tab.name}</span>
                    {tab.addon}
                </SeaLink>
            ) : (
                <div
                    key={tab.name}
                    className={cn(
                        StaticTabsAnatomy.trigger(),
                        "cursor-pointer",
                        triggerClass,
                    )}
                    aria-current={tab.isCurrent ? "page" : undefined}
                    data-current={tab.isCurrent}
                    onClick={tab.onClick}
                >
                    {tab.iconType && <tab.iconType
                        className={cn(
                            StaticTabsAnatomy.icon(),
                            iconClass,
                        )}
                        aria-hidden="true"
                        data-current={tab.isCurrent}
                    />}
                    <span>{tab.name}</span>
                    {tab.addon}
                </div>
            ))}
        </nav>
    )

})

StaticTabs.displayName = "StaticTabs"
