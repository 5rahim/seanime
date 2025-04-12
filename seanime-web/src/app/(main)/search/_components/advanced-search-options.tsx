"use client"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import {
    ADVANCED_SEARCH_COUNTRIES_MANGA,
    ADVANCED_SEARCH_FORMATS,
    ADVANCED_SEARCH_FORMATS_MANGA,
    ADVANCED_SEARCH_MEDIA_GENRES,
    ADVANCED_SEARCH_SEASONS,
    ADVANCED_SEARCH_SORTING,
    ADVANCED_SEARCH_SORTING_MANGA,
    ADVANCED_SEARCH_STATUS,
    ADVANCED_SEARCH_TYPE,
} from "@/app/(main)/search/_lib/advanced-search-constants"
import { __advancedSearch_paramsAtom } from "@/app/(main)/search/_lib/advanced-search.atoms"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { IconButton } from "@/components/ui/button"
import { Combobox } from "@/components/ui/combobox"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { getYear } from "date-fns"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useState } from "react"
import { BiTrash, BiWorld } from "react-icons/bi"
import { FaRegStar, FaSortAmountDown } from "react-icons/fa"
import { FiSearch } from "react-icons/fi"
import { LuCalendar, LuLeaf } from "react-icons/lu"
import { MdOutlineBook, MdPersonalVideo } from "react-icons/md"
import { RiSignalTowerLine } from "react-icons/ri"
import { TbSwords } from "react-icons/tb"
import { useUpdateEffect } from "react-use"

export function AdvancedSearchOptions() {

    const serverStatus = useServerStatus()
    const [params, setParams] = useAtom(__advancedSearch_paramsAtom)

    return (
        <AppLayoutStack data-advanced-search-options-container className="px-4 xl:px-0">
            <div data-advanced-search-options-header className="flex flex-col md:flex-row xl:flex-col gap-4">
                <TitleInput/>
                <Select
                    className="w-full"
                    options={ADVANCED_SEARCH_TYPE}
                    value={params.type}
                    onValueChange={v => setParams(draft => {
                        draft.type = v as "anime" | "manga"
                        return
                    })}
                />
                <Select
                    // label="Sorting"
                    leftAddon={<FaSortAmountDown />}
                    className="w-full"
                    options={params.type === "anime" ? ADVANCED_SEARCH_SORTING : ADVANCED_SEARCH_SORTING_MANGA}
                    value={params.sorting?.[0] || "SCORE_DESC"}
                    onValueChange={v => setParams(draft => {
                        draft.sorting = [v] as any
                        return
                    })}
                    disabled={!!params.title && params.title.length > 0}
                />
            </div>
            <div data-advanced-search-options-content className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-1 gap-4 items-end xl:items-start">
                <Combobox
                    multiple
                    leftAddon={<TbSwords />}
                    emptyMessage="No options found"
                    label="Genre" placeholder="All genres" className="w-full"
                    options={ADVANCED_SEARCH_MEDIA_GENRES.map(genre => ({ value: genre, label: genre, textValue: genre }))}
                    value={params.genre ? params.genre : []}
                    onValueChange={v => setParams(draft => {
                        draft.genre = v
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                {params.type === "anime" && <Select
                    leftAddon={<MdPersonalVideo />}
                    label="Format" placeholder="All formats" className="w-full"
                    options={ADVANCED_SEARCH_FORMATS}
                    value={params.format || ""}
                    onValueChange={v => setParams(draft => {
                        draft.format = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                {params.type === "manga" && <Select
                    leftAddon={<BiWorld />}
                    label="Format" placeholder="All countries" className="w-full"
                    options={ADVANCED_SEARCH_COUNTRIES_MANGA}
                    value={params.countryOfOrigin || ""}
                    onValueChange={v => setParams(draft => {
                        draft.countryOfOrigin = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                {params.type === "manga" && <Select
                    leftAddon={<MdOutlineBook />}
                    label="Format" placeholder="All formats" className="w-full"
                    options={ADVANCED_SEARCH_FORMATS_MANGA}
                    value={params.format || ""}
                    onValueChange={v => setParams(draft => {
                        draft.format = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                {params.type === "anime" && <Select
                    leftAddon={<LuLeaf />}
                    placeholder="All seasons" className="w-full"
                    options={ADVANCED_SEARCH_SEASONS.map(season => ({ value: season.toUpperCase(), label: season }))}
                    value={params.season || ""}
                    onValueChange={v => setParams(draft => {
                        draft.season = v as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                <Select
                    leftAddon={<LuCalendar />}
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
                    fieldLabelClass="hidden"
                />
                <Select
                    leftAddon={<RiSignalTowerLine />}
                    label="Status" placeholder="All statuses" className="w-full"
                    options={ADVANCED_SEARCH_STATUS}
                    value={params.status?.[0] || ""}
                    onValueChange={v => setParams(draft => {
                        draft.status = [v] as any
                        return
                    })}
                    fieldLabelClass="hidden"
                />
                <Select
                    leftAddon={<FaRegStar />}
                    placeholder="All scores" className="w-full"
                    options={[...Array(9)].map((v, idx) => 9 - idx).map(score => ({
                        value: String(score),
                        label: String(score),
                    }))}
                    value={params.minScore || ""}
                    onValueChange={v => setParams(draft => {
                        draft.minScore = v as any
                        return
                    })}
                />
                {serverStatus?.settings?.anilist?.enableAdultContent && <Switch
                    label="Adult"
                    value={params.isAdult}
                    onValueChange={v => setParams(draft => {
                        draft.isAdult = v
                        return
                    })}
                    fieldLabelClass="hidden"
                />}
                <IconButton
                    icon={<BiTrash />} intent="gray-subtle" className="flex-none" onClick={() => {
                    setParams(prev => ({
                        ...prev,
                        active: true,
                        title: null,
                        sorting: null,
                        status: null,
                        genre: null,
                        format: null,
                        season: null,
                        year: null,
                        minScore: null,
                    }))
                }}/>
            </div>

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
