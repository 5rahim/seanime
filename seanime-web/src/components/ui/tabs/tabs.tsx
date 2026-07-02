import * as TabsPrimitive from "@radix-ui/react-tabs"
import { cva } from "class-variance-authority"
import { motion, useReducedMotion } from "motion/react"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const TabsAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Tabs__root",
    ]),
    list: cva([
        "UI-Tabs__list",
        "inline-flex items-center justify-center w-full",
    ], {
        variants: {
            variant: {
                underline: "h-12 border-b border-[--border]",
                pill: "w-full flex flex-wrap md:flex-nowrap bg-[--paper] p-1 border rounded-xl h-fit",
                none: "",
            },
        },
        defaultVariants: {
            variant: "underline",
        },
    }),
    trigger: cva([
        "UI-Tabs__trigger appearance-none shadow-none",
        "inline-flex h-full items-center justify-center whitespace-nowrap text-sm font-medium ring-offset-[--background]",
        "transition-all focus-visible:outline-none focus-visible:ring-1 ring-offset-1 ring-offset-[--background] focus-visible:ring-white/40",
        "disabled:pointer-events-none disabled:opacity-50",
    ], {
        variants: {
            variant: {
                underline: [
                    "px-3 py-1.5 border-transparent border-b-2 -mb-px text-[--muted]",
                    "data-[state=active]:border-[--brand] data-[state=active]:text-[--foreground]",
                ],
                pill: [
                    "text-base px-6 h-auto py-2 rounded-[--radius-md] w-fit md:w-full border-none text-[--muted]",
                    "data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
                ],
                none: [
                    "px-3 py-1.5 border-transparent border-b-2 -mb-px text-[--muted]",
                    "data-[state=active]:border-[--brand] data-[state=active]:text-[--foreground]",
                ],
            },
        },
        defaultVariants: {
            variant: "underline",
        },
    }),
    content: cva([
        "UI-Tabs__content",
        "focus-visible:outline-none",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Helpers
 * -----------------------------------------------------------------------------------------------*/

function getActiveBgClass(classes?: string): string {
    if (!classes) return "bg-[--subtle]"
    const matches = classes.match(/data-\[state=active\]:(bg-\S+)/g)
    if (!matches) return "bg-[--subtle]"
    return matches.map(m => m.replace("data-\[state=active\]:", "")).join(" ")
}

function getRoundedClass(classes?: string): string {
    if (!classes) return "rounded-[--radius-md]"
    const match = classes.match(/(rounded\S*)/)
    return match ? match[1] : "rounded-[--radius-md]"
}

function getJustifyClass(classes?: string): string {
    if (!classes) return "justify-center"
    const match = classes.match(/((?:lg:|md:|sm:)?justify-\S+)/g)
    return match ? match.join(" ") : "justify-center"
}

/* -------------------------------------------------------------------------------------------------
 * Tabs
 * -----------------------------------------------------------------------------------------------*/

interface TabsContextValue extends ComponentAnatomy<typeof TabsAnatomy> {
    activeTab?: string
    layoutId?: string
    variant?: "underline" | "pill" | "none"
}

const __TabsAnatomyContext = React.createContext<TabsContextValue>({})

export type TabsProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Root> & ComponentAnatomy<typeof TabsAnatomy> & {
    variant?: "underline" | "pill" | "none"
}

export const Tabs = React.forwardRef<HTMLDivElement, TabsProps>((props, ref) => {
    const {
        className,
        listClass,
        triggerClass,
        contentClass,
        variant = "underline",
        value: valueProp,
        defaultValue,
        onValueChange,
        ...rest
    } = props

    const [activeTab, setActiveTab] = React.useState(valueProp ?? defaultValue)

    React.useEffect(() => {
        if (valueProp !== undefined) {
            setActiveTab(valueProp)
        }
    }, [valueProp])

    const handleValueChange = React.useCallback((val: string) => {
        if (valueProp === undefined) {
            setActiveTab(val)
        }
        onValueChange?.(val)
    }, [onValueChange, valueProp])

    const uniqueId = React.useId()
    const layoutId = React.useMemo(() => `tab-indicator-${uniqueId.replace(/:/g, "")}`, [uniqueId])

    return (
        <__TabsAnatomyContext.Provider
            value={{
                listClass,
                triggerClass,
                contentClass,
                activeTab,
                layoutId,
                variant,
            }}
        >
            <TabsPrimitive.Root
                ref={ref}
                value={valueProp}
                defaultValue={defaultValue}
                onValueChange={handleValueChange}
                className={cn(TabsAnatomy.root(), className)}
                {...rest}
            />
        </__TabsAnatomyContext.Provider>
    )
})

Tabs.displayName = "Tabs"

/* -------------------------------------------------------------------------------------------------
 * TabsList
 * -----------------------------------------------------------------------------------------------*/

export type TabsListProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.List> & {
    variant?: "underline" | "pill" | "none"
}

export const TabsList = React.forwardRef<HTMLDivElement, TabsListProps>((props, ref) => {
    const { className, variant: variantProp, ...rest } = props

    const { listClass, variant: contextVariant } = React.useContext(__TabsAnatomyContext)
    const variant = variantProp ?? contextVariant

    return (
        <TabsPrimitive.List
            ref={ref}
            className={cn(TabsAnatomy.list({ variant }), listClass, className)}
            {...rest}
        />
    )
})

TabsList.displayName = "TabsList"


/* -------------------------------------------------------------------------------------------------
 * TabsTrigger
 * -----------------------------------------------------------------------------------------------*/

export type TabsTriggerProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger> & {
    variant?: "underline" | "pill" | "none"
}

export const TabsTrigger = React.forwardRef<HTMLButtonElement, TabsTriggerProps>((props, ref) => {
    const { className, variant: variantProp, ...rest } = props

    const { triggerClass, activeTab, layoutId, variant: contextVariant } = React.useContext(__TabsAnatomyContext)
    const variant = variantProp ?? contextVariant
    const isReducedMotion = useReducedMotion()

    const isActive = activeTab === rest.value
    const isAnimated = !isReducedMotion && variant !== "none"

    const mergedClasses = cn(TabsAnatomy.trigger({ variant }), triggerClass, className)

    // Override active background / border when layout animations are enabled
    const overrideClass = isAnimated
        ? variant === "underline"
            ? "data-[state=active]:border-transparent"
            : variant === "pill"
                ? "data-[state=active]:bg-transparent"
                : ""
        : ""

    const activeBgClass = React.useMemo(() => getActiveBgClass(mergedClasses), [mergedClasses])
    const roundedClass = React.useMemo(() => getRoundedClass(mergedClasses), [mergedClasses])
    const justifyClass = React.useMemo(() => getJustifyClass(mergedClasses), [mergedClasses])

    return (
        <TabsPrimitive.Trigger
            ref={ref}
            data-tab-trigger={rest.value}
            className={cn(
                TabsAnatomy.trigger({ variant }),
                triggerClass,
                overrideClass,
                className,
                isAnimated && "relative z-0",
            )}
            {...rest}
        >
            {isAnimated ? (
                <>
                    <span className={cn("relative z-10 w-full h-full inline-flex items-center", justifyClass)}>
                        {props.children}
                    </span>
                    {isActive && variant === "underline" && (
                        <motion.span
                            layoutId={layoutId}
                            className="absolute bottom-0 left-0 right-0 h-[2px] bg-[--brand] z-10"
                            transition={{ type: "spring", stiffness: 500, damping: 38 }}
                        />
                    )}
                    {isActive && variant === "pill" && (
                        <motion.span
                            layoutId={layoutId}
                            className={cn("absolute inset-0 -z-10", activeBgClass, roundedClass)}
                            transition={{ type: "spring", stiffness: 500, damping: 38 }}
                        />
                    )}
                </>
            ) : (
                props.children
            )}
        </TabsPrimitive.Trigger>
    )
})

TabsTrigger.displayName = "TabsTrigger"

/* -------------------------------------------------------------------------------------------------
 * TabsContent
 * -----------------------------------------------------------------------------------------------*/

export type TabsContentProps = React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>

export const TabsContent = React.forwardRef<HTMLDivElement, TabsContentProps>((props, ref) => {
    const { className, ...rest } = props

    const { contentClass } = React.useContext(__TabsAnatomyContext)

    return (
        <TabsPrimitive.Content
            ref={ref}
            className={cn(TabsAnatomy.content(), contentClass, className)}
            {...rest}
        />
    )
})

TabsContent.displayName = "TabsContent"

