import { Modal } from "@/components/ui/modal"
import { TextInput, TextInputProps } from "@/components/ui/text-input"
import { useBoolean } from "@/hooks/use-disclosure"
import { useQuery } from "@tanstack/react-query"
import axios from "axios"
import React, { memo, startTransition, useCallback, useEffect, useRef, useState } from "react"
import { BiCheck, BiFolderOpen, BiFolderPlus, BiX } from "react-icons/bi"
import { FaFolder } from "react-icons/fa"
import { useUpdateEffect } from "react-use"
import * as upath from "upath"
import { useDebounce } from "use-debounce"

export type DirectorySelectorProps = {
    defaultValue?: string
    onSelect: (path: string) => void
    shouldExist?: boolean
    value: string
} & Omit<TextInputProps, "onSelect" | "value">

type DirectorySelectorResponse = {
    exists: boolean,
    suggestions: { fullPath: string, folderName: string }[],
    content?: { fullPath: string, folderName: string }[]
}

export const DirectorySelector = memo(React.forwardRef<HTMLInputElement, DirectorySelectorProps>(function (props: DirectorySelectorProps, ref) {

    const {
        defaultValue,
        onSelect,
        value,
        shouldExist,
        ...rest
    } = props

    const firstRender = useRef(true)

    const [input, setInput] = useState(defaultValue ? upath.normalize(defaultValue) : "")
    const [debouncedInput] = useDebounce(input, 500)
    const selectorState = useBoolean(false)
    const prevState = useRef<string>(input)

    const { data, isLoading, error } = useQuery({
        queryKey: ["directory-settings", debouncedInput],
        queryFn: async () => {
            const res = await axios.post<DirectorySelectorResponse>("http://" + (process.env.NODE_ENV === "development"
                ? `${window.location.hostname}:43211`
                : window.location.host) + "/api/v1/directory-selector", {
                input: debouncedInput,
            })
            return res.data
        },
        enabled: debouncedInput.length > 0,
    })

    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (value !== input) {
            setInput(value)
        }
    }, [value])

    useEffect(() => {
        if (input === ".") {
            setInput("")
        }
    }, [input])

    useUpdateEffect(() => {
        if (debouncedInput.length > 0) {
            if (shouldExist && data?.exists) {
                onSelect(debouncedInput)
                prevState.current = debouncedInput
            } else if (!shouldExist) {
                onSelect(debouncedInput)
                prevState.current = debouncedInput
            }
        }
    }, [debouncedInput, data])

    const checkDirectoryExists = useCallback(() => {
        if (!isLoading && data && shouldExist && !data.exists) {
            startTransition(() => {
                setInput(prevState.current)
            })
        }
    }, [isLoading, data, shouldExist, prevState.current])

    return (
        <>
            <div className="space-y-1">
                <div className="relative">
                    <TextInput
                        leftIcon={<FaFolder />}
                        {...rest}
                        value={input}
                        rightIcon={<div className="flex">
                            {isLoading ? null : (data?.exists ?
                                <BiCheck className="text-green-500" /> : shouldExist ?
                                    <BiX className="text-red-500" /> : <BiFolderPlus />)}
                        </div>}
                        onChange={e => {
                            setInput(upath.normalize(e.target.value ?? ""))
                        }}
                        ref={ref}
                        onBlur={checkDirectoryExists}
                    />
                    <BiFolderOpen
                        className="text-2xl cursor-pointer absolute z-[1] top-0 right-0"
                        onClick={selectorState.on}
                    />
                </div>
                {(!data?.exists && data?.suggestions && data.suggestions.length > 0) &&
                    <div
                        className="w-full flex flex-none flex-nowrap overflow-x-auto gap-2 items-center bg-gray-800 rounded-md p-1 px-4"
                    >
                        <div className="flex-none">Sub-folders:</div>
                        {data.suggestions.map(folder => (
                            <div
                                key={folder.fullPath}
                                className="py-1 text-sm px-3 rounded-md border  flex-none cursor-pointer bg-gray-900 hover:bg-gray-800"
                                onClick={() => setInput(upath.normalize(folder.fullPath))}
                            >
                                {folder.folderName}
                            </div>
                        ))}
                    </div>}
            </div>
            <Modal
                open={selectorState.active}
                onOpenChange={selectorState.toggle}
                title="Directory selector"
                contentClass="mt-4 space-y-4"
            >
                <TextInput
                    leftIcon={<FaFolder />}
                    value={input}
                    rightIcon={isLoading ? null : (data?.exists ?
                        <BiCheck className="text-green-500" /> : shouldExist ?
                            <BiX className="text-red-500" /> : <BiFolderPlus />)}
                    onChange={e => {
                        setInput(upath.normalize(e.target.value ?? ""))
                    }}
                    onClick={() => {
                        if (shouldExist) selectorState.on()
                    }}
                    ref={ref}
                />
                {(data && (data?.content && data.content.length > 0)) &&
                    <div
                        className="w-full flex flex-col flex-none flex-nowrap overflow-x-auto gap-1 max-h-60"
                    >
                        <div className="flex-none">Sub-folders:</div>
                        {data.content.map(folder => (
                            <div
                                key={folder.fullPath}
                                className="w-full py-2 text-sm px-3 rounded-md border  flex-none cursor-pointer bg-gray-900 hover:bg-gray-800 truncate"
                                onClick={() => setInput(upath.normalize(folder.fullPath))}
                            >
                                {folder.folderName}
                            </div>
                        ))}
                    </div>}
            </Modal>
        </>
    )

}))
