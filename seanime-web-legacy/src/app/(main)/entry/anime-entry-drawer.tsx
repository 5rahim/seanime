import { atom, useSetAtom } from "jotai"
import React from "react"

export const __animeDrawer_entryIdAtom = atom<number | null>(null)

type AnimeEntryDrawerProps = {
    children?: React.ReactNode
}

export function AnimeEntryDrawer(props: AnimeEntryDrawerProps) {

    const {
        children,
        ...rest
    } = props

    return (
        <>

        </>
    )
}

export function useSetAnimeDrawerEntryId(entryId: number) {
    const setEntryId = useSetAtom(__animeDrawer_entryIdAtom)
    setEntryId(entryId)
}
