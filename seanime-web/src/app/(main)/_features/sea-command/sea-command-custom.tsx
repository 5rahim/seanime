import { CommandGroup, CommandItem } from "@/components/ui/command"
import React from "react"
import { useSeaCommandContext } from "./sea-command"
import { SeaCommandInjectableItem, useSeaCommandInject } from "./sea-command.atoms"

export function SeaCommandCustom() {
    const { input, select, scrollToTop } = useSeaCommandContext<"other">()
    const { injectables } = useSeaCommandInject()

    // Group items by heading and sort by priority
    const groupedItems = React.useMemo(() => {
        const groups: Record<string, SeaCommandInjectableItem[]> = {}

        Object.values(injectables).forEach(injectable => {
            if (injectable.shouldShow?.(input) === false) return
            if (!injectable.isCommand && input.startsWith("/")) return

            const items = injectable.items.filter(item =>
                injectable.filter?.(item, input) ??
                item.value.toLowerCase().includes(input.toLowerCase()),
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
                                select(item.onSelect)
                            }}
                        >
                            {item.render({ onSelect: () => select(item.onSelect) })}
                        </CommandItem>
                    ))}
                </CommandGroup>
            ))}
        </>
    )
}
