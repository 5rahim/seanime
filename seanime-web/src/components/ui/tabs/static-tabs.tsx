import { SeaLink } from "@/components/shared/sea-link"
import { cva } from "class-variance-authority"
import { motion, useReducedMotion } from "motion/react"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

function getActiveBgClass(classes?: string): string {
    if (!classes) return "bg-[--subtle]"
    const matches = classes.match(/data-\[current=true\]:(bg-\S+)/g)
    if (!matches) return "bg-[--subtle]"
    return matches.map(m => m.replace("data-[current=true]:", "")).join(" ")
}

function getActiveBorderClass(classes?: string): string {
    if (!classes) return "border-gray-700"
    const match = classes.match(/data-\[current=true\]:(border-\S+)/)
    return match ? match[1] : "border-gray-700"
}

function getRoundedClass(classes?: string): string {
    if (!classes) return "rounded-xl"
    const match = classes.match(/(rounded-\S+|rounded-full|rounded-none)/)
    return match ? match[1] : "rounded-xl"
}

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const StaticTabsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-StaticTabs__root",
        "flex w-full overflow-hidden overflow-x-auto gap-2 py-1",
    ]),
    trigger: cva([
        "UI-StaticTabs__trigger",
        "group/staticTabs__trigger inline-flex flex-none shrink-0 basis-auto items-center font-medium text-sm transition outline-none min-w-0 justify-center",
        "text-[--muted] hover:text-[--foreground]",
        "h-10 px-4 rounded-xl border border-transparent",
        "data-[current=true]:bg-[--subtle] data-[current=true]:font-semibold data-[current=true]:text-white data-[current=true]:border-gray-700",
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
    icon?: React.ReactNode,
    iconType?: React.ElementType,
    onClick?: () => void,
    isCurrent: boolean
    addon?: React.ReactNode,
}

export type StaticTabsProps = React.ComponentPropsWithRef<"nav"> &
    ComponentAnatomy<typeof StaticTabsAnatomy> & {
    items: StaticTabsItem[]
    pillClass?: string
}

export const StaticTabs = React.forwardRef<HTMLElement, StaticTabsProps>((props, ref) => {

    const {
        children,
        className,
        triggerClass,
        iconClass,
        items,
        pillClass,
        ...rest
    } = props

    const isReducedMotion = useReducedMotion()
    const isAnimated = !isReducedMotion

    const uniqueId = React.useId()
    const layoutId = React.useMemo(() => `static-tab-indicator-${uniqueId.replace(/:/g, "")}`, [uniqueId])

    const overrideClass = isAnimated ? "data-[current=true]:bg-transparent data-[current=true]:border-transparent" : ""

    const mergedTriggerClass = cn(StaticTabsAnatomy.trigger(), triggerClass)
    const activeBg = React.useMemo(() => getActiveBgClass(mergedTriggerClass), [mergedTriggerClass])
    const activeBorder = React.useMemo(() => getActiveBorderClass(mergedTriggerClass), [mergedTriggerClass])
    const roundedClass = React.useMemo(() => getRoundedClass(mergedTriggerClass), [mergedTriggerClass])

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
                        overrideClass,
                        triggerClass,
                        isAnimated && "relative z-0",
                    )}
                    aria-current={tab.isCurrent ? "page" : undefined}
                    data-current={tab.isCurrent}
                >
                    {isAnimated ? (
                        <>
                            <span className="relative z-10 inline-flex items-center">
                                {tab.icon ?? (tab.iconType && <tab.iconType
                                    className={cn(
                                        StaticTabsAnatomy.icon(),
                                        iconClass,
                                    )}
                                    aria-hidden="true"
                                    data-current={tab.isCurrent}
                                />)}
                                <span>{tab.name}</span>
                                {tab.addon}
                            </span>
                            {tab.isCurrent && (
                                <motion.span
                                    layoutId={layoutId}
                                    className={cn("absolute inset-0 -z-10 border", activeBg, activeBorder, roundedClass, pillClass)}
                                    transition={{ type: "spring", stiffness: 500, damping: 38 }}
                                />
                            )}
                        </>
                    ) : (
                        <>
                            {tab.icon ?? (tab.iconType && <tab.iconType
                                className={cn(
                                    StaticTabsAnatomy.icon(),
                                    iconClass,
                                )}
                                aria-hidden="true"
                                data-current={tab.isCurrent}
                            />)}
                            <span>{tab.name}</span>
                            {tab.addon}
                        </>
                    )}
                </SeaLink>
            ) : (
                <div
                    key={tab.name}
                    className={cn(
                        StaticTabsAnatomy.trigger(),
                        "cursor-pointer",
                        overrideClass,
                        triggerClass,
                        isAnimated && "relative z-0",
                    )}
                    aria-current={tab.isCurrent ? "page" : undefined}
                    data-current={tab.isCurrent}
                    onClick={tab.onClick}
                >
                    {isAnimated ? (
                        <>
                            <span className="relative z-10 inline-flex items-center">
                                {tab.icon ?? (tab.iconType && <tab.iconType
                                    className={cn(
                                        StaticTabsAnatomy.icon(),
                                        iconClass,
                                    )}
                                    aria-hidden="true"
                                    data-current={tab.isCurrent}
                                />)}
                                <span>{tab.name}</span>
                                {tab.addon}
                            </span>
                            {tab.isCurrent && (
                                <motion.span
                                    layoutId={layoutId}
                                    className={cn("absolute inset-0 -z-10 border", activeBg, activeBorder, roundedClass, pillClass)}
                                    transition={{ type: "spring", stiffness: 500, damping: 38 }}
                                />
                            )}
                        </>
                    ) : (
                        <>
                            {tab.icon ?? (tab.iconType && <tab.iconType
                                className={cn(
                                    StaticTabsAnatomy.icon(),
                                    iconClass,
                                )}
                                aria-hidden="true"
                                data-current={tab.isCurrent}
                            />)}
                            <span>{tab.name}</span>
                            {tab.addon}
                        </>
                    )}
                </div>
            ))}
        </nav>
    )

})

StaticTabs.displayName = "StaticTabs"
