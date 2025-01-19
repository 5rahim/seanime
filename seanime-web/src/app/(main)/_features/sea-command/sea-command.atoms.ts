import { Anime_Entry, Anime_Episode } from "@/api/generated/types"
import { Nullish } from "@/types/common"
import { atom, useAtom, useSetAtom } from "jotai"
import React from "react"

export type CommandLibraryPageParams = {
    episodes: Anime_Episode[]
}

export type CommandAnimePageParams = {
    entry: Nullish<Anime_Entry>
}

export type SeaCommandPage = "anime-library" | "anime-entry" | "other"

export type SeaCommandParams<T extends SeaCommandPage> = {
    page: T
    pageParams?: T extends "anime-library" ?
        CommandLibraryPageParams : T extends "anime-entry" ?
            CommandAnimePageParams : never
}

const paramsAtom = atom<SeaCommandParams<SeaCommandPage>>({ page: "other" })

export function useSeaCommandParams() {
    const [params, setParams] = useAtom(paramsAtom)
    return {
        params,
        setParams,
    }
}

export function useSetSeaCommandParams<T extends SeaCommandPage>(params: SeaCommandParams<T>) {
    const setParams = useSetAtom(paramsAtom)
    React.useEffect(() => {
        setParams(params)
    }, [params])
}

export type SeaCommandInjectableItem = {
    id: string // Unique identifier for the item
    value: string // Value used for filtering/searching
    heading?: string // Optional group heading
    priority?: number // Optional priority (higher = shown first)
    render: (props: { onSelect: () => void }) => React.ReactNode // Render function for the item
    onSelect: () => void // What happens when item is selected
}

export type SeaCommandInjectable = {
    items: SeaCommandInjectableItem[]
    filter?: (item: SeaCommandInjectableItem, input: string) => boolean // Custom filter function
    shouldShow?: (input: string) => boolean // When to show these items
    isCommand?: boolean
}

const injectablesAtom = atom<Record<string, SeaCommandInjectable>>({})

export function useSeaCommandInject() {
    const [injectables, setInjectables] = useAtom(injectablesAtom)

    const inject = React.useCallback((key: string, injectable: SeaCommandInjectable) => {
        setInjectables(prev => ({
            ...prev,
            [key]: injectable,
        }))

    }, [])

    const remove = React.useCallback((key: string) => {
        setInjectables(prev => {
            const next = { ...prev }
            delete next[key]
            return next
        })
    }, [])

    return {
        inject,
        remove,
        injectables,
    }
}
