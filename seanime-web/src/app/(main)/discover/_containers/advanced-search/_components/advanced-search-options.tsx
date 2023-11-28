"use client"
import React, { useState } from "react"
import { useDebounce } from "@/hooks/use-debounce"
import { useSetAtom } from "jotai"
import { useUpdateEffect } from "react-use"
import { TextInput } from "@/components/ui/text-input"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { __advancedSearch_paramsAtom } from "@/app/(main)/discover/_containers/advanced-search/_lib/parameters"
import { Select } from "@/components/ui/select"
import { MultiSelect } from "@/components/ui/multi-select"
import { getYear } from "date-fns"
import { IconButton } from "@/components/ui/button"
import { BiTrash } from "@react-icons/all-files/bi/BiTrash"
import { AppLayoutStack } from "@/components/ui/app-layout"
import {
    ADVANCED_SEARCH_FORMATS,
    ADVANCED_SEARCH_MEDIA_GENRES,
    ADVANCED_SEARCH_SEASONS,
    ADVANCED_SEARCH_SORTING,
    ADVANCED_SEARCH_STATUS,
} from "@/app/(main)/discover/_containers/advanced-search/_lib/constants"
import { useAtom } from "jotai/react"

export function AdvancedSearchOptions() {

    const [params, setParams] = useAtom(__advancedSearch_paramsAtom)

    return (
        <AppLayoutStack className={"px-4 xl:px-0"}>
            <div className={"flex flex-row xl:flex-col gap-4"}>
                <TitleInput/>
                <Select
                    // label={"Sorting"}
                    className={"w-full"}
                    options={ADVANCED_SEARCH_SORTING}
                    value={params.sorting || "SCORE_DESC"}
                    onChange={e => setParams(draft => {
                        draft.sorting = [e.target.value] as any
                        return
                    })}
                    isDisabled={!!params.title && params.title.length > 0}
                />
            </div>
            <div className={"flex flex-row xl:flex-col gap-4 items-end xl:items-start"}>
                <MultiSelect
                    label={"Genre"} placeholder={"All genres"} className={"w-full"}
                    options={ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({ value: genre, label: genre }))}
                    value={params.genre ? params.genre : undefined}
                    onChange={e => setParams(draft => {
                        draft.genre = e
                        return
                    })}
                />
                <Select
                    label={"Format"} placeholder={"All formats"} className={"w-full"}
                    options={ADVANCED_SEARCH_FORMATS}
                    value={params.format || ""}
                    onChange={e => setParams(draft => {
                        draft.format = e.target.value as any
                        return
                    })}
                />
                <Select
                    label={"Season"} placeholder={"All seasons"} className={"w-full"}
                    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}
                    value={params.season || ""}
                    onChange={e => setParams(draft => {
                        draft.season = e.target.value as any
                        return
                    })}
                />
                <Select
                    label={"Year"} placeholder={"Timeless"} className={"w-full"}
                    options={[...Array(70)].map((v, idx) => getYear(new Date()) - idx).map(year => ({
                        value: String(year),
                        label: String(year),
                    }))}
                    value={params.year || ""}
                    onChange={e => setParams(draft => {
                        draft.year = e.target.value as any
                        return
                    })}
                />
                <Select
                    label={"Status"} placeholder={"All"} className={"w-full"}
                    options={ADVANCED_SEARCH_STATUS}
                    value={params.status || ""}
                    onChange={e => setParams(draft => {
                        draft.status = [e.target.value] as any
                        return
                    })}
                />
                <IconButton icon={<BiTrash/>} intent={"gray-subtle"} className={"flex-none"} onClick={() => {
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
            {/*    label={"Minimum score"} placeholder={"No preference"} className={"w-full"}*/}
            {/*    options={[...Array(9)].map((v, idx) => 9 - idx).map(score => ({*/}
            {/*        value: String(score),*/}
            {/*        label: String(score),*/}
            {/*    }))}*/}
            {/*    value={params.minScore || ""}*/}
            {/*    onChange={e => setParams(draft => {*/}
            {/*        draft.minScore = e.target.value as any*/}
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
            leftIcon={<FiSearch/>} placeholder={"Title"} className={"w-full"}
            value={inputValue}
            onChange={e => setInputValue(e.target.value)}
        />
    )
}
