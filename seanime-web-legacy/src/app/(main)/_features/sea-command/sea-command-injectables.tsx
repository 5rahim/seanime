import { SeaCommandInjectableItem, useSeaCommandInjectables } from "@/app/(main)/_features/sea-command/use-inject"
import { CommandGroup, CommandItem } from "@/components/ui/command"
import React from "react"
import { useSeaCommandContext } from "./sea-command"

export function SeaCommandInjectables() {
    const ctx = useSeaCommandContext()
    const { input, select, scrollToTop } = ctx
    const injectables = useSeaCommandInjectables()

    // Group items by heading and sort by priority
    const groupedItems = React.useMemo(() => {
        const groups: Record<string, SeaCommandInjectableItem[]> = {}

        Object.values(injectables).forEach(injectable => {
            if (injectable.shouldShow?.({ ctx }) === false) return
            if (!injectable.isCommand && input.startsWith("/")) return

            // const items = injectable.items.filter(item =>
            //     injectable.filter?.(item, input) ?? //Apply custom filter if provided
            //     item.value.toLowerCase().includes(input.toLowerCase()),
            // ).filter(item => // If the item should be rendered based on the input


            const items = injectable.items
                .filter(item =>
                    item.shouldShow?.({ ctx }) ?? true, // Apply custom filter if provided, otherwise don't filter
                ).filter(item =>
                    injectable.filter?.({ item, input }) ?? true, // Apply custom filter if provided, otherwise don't filter
                ).filter(item => // If the item should be rendered based on the input
                    item.showBasedOnInput === "includes" ?
                        item.value.toLowerCase().includes(input.toLowerCase()) :
                        item.showBasedOnInput === "startsWith" ?
                            item.value.toLowerCase().startsWith(input.toLowerCase()) :
                            true,
                ).filter(item => // If the group of items should be rendered based on the input
                    injectable.showBasedOnInput === "includes" ?
                        item.value.toLowerCase().includes(input.toLowerCase()) :
                        item.showBasedOnInput === "startsWith" ?
                            item.value.toLowerCase().startsWith(input.toLowerCase()) :
                            true,
            )

            items.forEach(item => {
                const heading = item.heading || "Custom"
                if (!groups[heading]) groups[heading] = []
                groups[heading].push(item)
            })
        })

        // Sort items in each group by priority
        Object.keys(groups).forEach(heading => {
            groups[heading].sort((a, b) => (b.priority || 0) - (a.priority || 0))
        })

        // Scroll to top when items are rendered
        scrollToTop()

        return groups
    }, [injectables, input])

    if (Object.keys(groupedItems).length === 0) return null

    return (
        <>
            {Object.entries(groupedItems).map(([heading, items]) => (
                <CommandGroup key={heading} heading={heading}>
                    {items.map(item => (
                        <CommandItem
                            key={item.id}
                            value={item.id}
                            onSelect={() => {
                                select(() => item.onSelect({ ctx }))
                            }}
                            className="gap-3"
                        >
                            {item.render()}
                        </CommandItem>
                    ))}
                </CommandGroup>
            ))}
        </>
    )
}
