import { SeaCommandContextProps } from "@/app/(main)/_features/sea-command/sea-command"
import { useAtomValue } from "jotai"
import { atom, useSetAtom } from "jotai/index"
import React from "react"

export type SeaCommandInjectableItem = {
    id: string // Unique identifier for the item
    value: string // Value used for filtering/searching
    heading?: string // Optional group heading
    priority?: number // Priority of the item (higher = shown first) (used to sort items in a group)
    render: () => React.ReactNode // Render function for the item
    onSelect: (props: { ctx: SeaCommandContextProps }) => void // What happens when item is selected
    shouldShow?: (props: { ctx: SeaCommandContextProps }) => boolean
    showBasedOnInput?: "startsWith" | "includes" // Optional automatic filtered based on the item value and input
    data?: any
}

export type SeaCommandInjectable = {
    items: SeaCommandInjectableItem[]
    filter?: (props: { item: SeaCommandInjectableItem, input: string }) => boolean // Custom filter function
    shouldShow?: (props: { ctx: SeaCommandContextProps }) => boolean // When to show these items
    isCommand?: boolean
    showBasedOnInput?: "startsWith" | "includes" // Optional automatic filtered based on the item value and input
    priority?: number // Priority of the items (used to sort groups)
}

const injectablesAtom = atom<Record<string, SeaCommandInjectable>>({})
// useSeaCommandInject
// Example:
// const { inject, remove } = useSeaCommandInject()
// React.useEffect(() => {
//     inject("continue-watching", {
//         items: episodes.map(episode => ({
//             data: episode,
//             id: `${episode.type}-${episode.localFile?.path || ""}-${episode.episodeNumber}`,
//             value: `${episode.episodeNumber}`,
//             heading: "Continue Watching",
//         })),
//         priority: 100,
//     })
//
//     return () => {
//         remove("continue-watching")
//     }
// }, [episodes])

export function useSeaCommandInjectables() {
    return useAtomValue(injectablesAtom)
}

export function useSeaCommandInject() {
    const setInjectables = useSetAtom(injectablesAtom)

    const inject = (key: string, injectable: SeaCommandInjectable) => {
        // setInjectables(prev => ({
        //     ...prev,
        //     [key]: injectable,
        // }))
        // Add to injectables based on priority
        setInjectables(prev => {
            // Transform into an array of items
            const items = Object.keys(prev).map(key => ({ injectable: prev[key], key }))
            // Add the new injectable to the array
            items.push({ injectable, key })
            // Sort the items by priority
            items.sort((a, b) => (b.injectable.priority || 0) - (a.injectable.priority || 0))
            // Transform back into an object
            const ret = items.reduce((acc, item) => {
                acc[item.key] = item.injectable
                return acc
            }, {} as Record<string, SeaCommandInjectable>)
            return ret
        })
    }

    const remove = (key: string) => {
        setInjectables(prev => {
            const next = { ...prev }
            delete next[key]
            return next
        })
    }

    return {
        inject,
        remove,
    }

}
