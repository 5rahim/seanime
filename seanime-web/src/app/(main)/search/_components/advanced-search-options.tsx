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
import { cn } from "@/components/ui/core/styling"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { getYear } from "date-fns"
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

    const highlightTrash = React.useMemo(() => {
        return !(!params.title?.length &&
            (params.sorting === null || params.sorting?.[0] === "SCORE_DESC") &&
            (params.genre === null || !params.genre.length) &&
            (params.status === null || !params.status.length) &&
            params.format === null && params.season === null && params.year === null && params.isAdult === false && params.minScore === null &&
            (params.countryOfOrigin === null || params.type === "anime"))
    }, [params])

    return (
        <AppLayoutStack data-advanced-search-options-container className="px-4 xl:px-0">
            <div data-advanced-search-options-header className="flex flex-col md:flex-row xl:flex-col gap-4">
                <TitleInput />
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
                    leftAddon={
                        <FaSortAmountDown className={cn((params.sorting !== null && params.sorting?.[0] !== "SCORE_DESC") && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<TbSwords className={cn((params.genre !== null && !!params.genre.length) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<MdPersonalVideo className={cn((params.format !== null && !!params.format) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={
                        <BiWorld className={cn((params.countryOfOrigin !== null && !!params.countryOfOrigin) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<MdOutlineBook className={cn((params.format !== null && !!params.format) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<LuLeaf className={cn((params.season !== null && !!params.season) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<LuCalendar className={cn((params.year !== null && !!params.year) && "text-indigo-300 font-bold text-xl")} />}
                    label="Year" placeholder="Timeless" className="w-full"
                    options={[...Array(70)].map((v, idx) => getYear(new Date()) - idx + 2).map(year => ({
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
                    leftAddon={
                        <RiSignalTowerLine className={cn((params.status !== null && !!params.status.length) && "text-indigo-300 font-bold text-xl")} />}
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
                    leftAddon={<FaRegStar className={cn((params.minScore !== null && !!params.minScore) && "text-indigo-300 font-bold text-xl")} />}
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
                    icon={<BiTrash />} intent={highlightTrash ? "alert" : "gray-subtle"} className="flex-none" onClick={() => {
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
                        countryOfOrigin: null,
                        // isAdult: false,
                    }))
                }}
                    disabled={!highlightTrash}
                />
            </div>

        </AppLayoutStack>
    )
}

function TitleInput() {
    const [inputValue, setInputValue] = useState("")
    const debouncedTitle = useDebounce(inputValue, 500)
    const [params, setParams] = useAtom(__advancedSearch_paramsAtom)

    useUpdateEffect(() => {
        setParams(draft => {
            draft.title = debouncedTitle
            return
        })
    }, [debouncedTitle])

    useUpdateEffect(() => {
        setInputValue(params.title || "")
    }, [params.title])

    return (
        <TextInput
            leftIcon={<FiSearch />} placeholder="Title" className="w-full"
            value={inputValue}
            onValueChange={v => setInputValue(v)}
        />
    )
}
