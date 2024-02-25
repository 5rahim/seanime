"use client"
import {
    ADVANCED_SEARCH_FORMATS,
    ADVANCED_SEARCH_MEDIA_GENRES,
    ADVANCED_SEARCH_SEASONS,
    ADVANCED_SEARCH_SORTING,
    ADVANCED_SEARCH_STATUS,
} from "@/app/(main)/discover/_containers/advanced-search/_lib/constants"
import { __advancedSearch_paramsAtom } from "@/app/(main)/discover/_containers/advanced-search/_lib/parameters"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { Combobox } from "@/components/ui/combobox"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { getYear } from "date-fns"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { BiTrash } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"
import { useUpdateEffect } from "react-use"

export function AdvancedSearchOptions() {

    const [params, setParams] = useAtom(__advancedSearch_paramsAtom)

    return (
        <AppLayoutStack className="px-4 xl:px-0">
            <div className="flex flex-row xl:flex-col gap-4">
                <TitleInput/>
                <Select
                    // label="Sorting"
                    className="w-full"
                    options={ADVANCED_SEARCH_SORTING}
                    value={params.sorting?.[0] || "SCORE_DESC"}
                    onValueChange={v => setParams(draft => {
                        draft.sorting = [v] as any
                        return
                    })}
                    disabled={!!params.title && params.title.length > 0}
                />
            </div>
            <div className="flex flex-row xl:flex-col gap-4 items-end xl:items-start">
                <Combobox
                    multiple
                    emptyMessage="No option found"
                    label="Genre" placeholder="All genres" className="w-full"
                    options={ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({ value: genre, label: genre, textValue: genre }))}
                    value={params.genre ? params.genre : undefined}
                    onValueChange={v => setParams(draft => {
                        draft.genre = v
                        return
                    })}
                />
                <Select
                    label="Format" placeholder="All formats" className="w-full"
                    options={ADVANCED_SEARCH_FORMATS}
                    value={params.format || ""}
                    onValueChange={v => setParams(draft => {
                        draft.format = v as any
                        return
                    })}
                />
                <Select
                    label="Season" placeholder="All seasons" className="w-full"
                    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}
                    value={params.season || ""}
                    onValueChange={v => setParams(draft => {
                        draft.season = v as any
                        return
                    })}
                />
                <Select
                    label="Year" placeholder="Timeless" className="w-full"
                    options={[...Array(70)].map((v, idx) => getYear(new Date()) - idx).map(year => ({
                        value: String(year),
                        label: String(year),
                    }))}
                    value={params.year || ""}
                    onValueChange={v => setParams(draft => {
                        draft.year = v as any
                        return
                    })}
                />
                <Select
                    label="Status" placeholder="All" className="w-full"
                    options={ADVANCED_SEARCH_STATUS}
                    value={params.status?.[0] || ""}
                    onValueChange={v => setParams(draft => {
                        draft.status = [v] as any
                        return
                    })}
                />
                <IconButton
                    icon={<BiTrash />} intent="gray-subtle" className="flex-none" onClick={() => {
                    setParams({
                        active: true,
                        title: null,
                        sorting: null,
                        status: null,
                        genre: null,
                        format: null,
                        season: null,
                        year: null,
                        minScore: null,
                    })
                }}/>
            </div>
            {/*<Select*/}
            {/*    label="Minimum score" placeholder="No preference" className="w-full"*/}
            {/*    options={[...Array(9)].map((v, idx) => 9 - idx).map(score => ({*/}
            {/*        value: String(score),*/}
            {/*        label: String(score),*/}
            {/*    }))}*/}
            {/*    value={params.minScore || ""}*/}
            {/*    onValueChange={v => setParams(draft => {*/}
            {/*        draft.minScore = v as any*/}
            {/*        return*/}
            {/*    })}*/}
            {/*/>*/}

        </AppLayoutStack>
    )
}

function TitleInput() {
    const [inputValue, setInputValue] = useState("")
    const debouncedTitle = useDebounce(inputValue, 500)
    const setParams = useSetAtom(__advancedSearch_paramsAtom)

    useUpdateEffect(() => {
        setParams(draft => {
            draft.title = debouncedTitle
            return
        })
    }, [debouncedTitle])

    return (
        <TextInput
            leftIcon={<FiSearch />} placeholder="Title" className="w-full"
            value={inputValue}
            onValueChange={v => setInputValue(v)}
        />
    )
}
