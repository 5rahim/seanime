import { useDirectorySelector } from "@/api/hooks/directory_selector.hooks"
import { Modal } from "@/components/ui/modal"
import { TextInput, TextInputProps } from "@/components/ui/text-input"
import { useBoolean } from "@/hooks/use-disclosure"
import React from "react"
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

export const DirectorySelector = React.memo(React.forwardRef<HTMLInputElement, DirectorySelectorProps>(function (props: DirectorySelectorProps, ref) {

    const {
        defaultValue,
        onSelect,
        value,
        shouldExist,
        ...rest
    } = props

    const firstRender = React.useRef(true)

    const [input, setInput] = React.useState(defaultValue ? upath.normalize(defaultValue) : "")
    const [debouncedInput] = useDebounce(input, 300)
    const selectorState = useBoolean(false)
    const prevState = React.useRef<string>(input)

    const { data, isLoading, error } = useDirectorySelector(debouncedInput)

    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (value !== input) {
            setInput(value)
        }
    }, [value])

    React.useEffect(() => {
        if (input === ".") {
            setInput("")
        }
    }, [input])

    useUpdateEffect(() => {
        onSelect(debouncedInput)
        prevState.current = debouncedInput
    }, [debouncedInput, data])

    const checkDirectoryExists = React.useCallback(() => {
        if (!isLoading && data && shouldExist && !data.exists && input.length > 0) {
            React.startTransition(() => {
                setInput(prevState.current)
            })
        }
    }, [isLoading, data, input, shouldExist, prevState.current])

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
                        <div className="flex-none">Subdirectories:</div>
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
                {(data && !!data?.content?.length) &&
                    <div
                        className="w-full flex flex-col flex-none flex-nowrap overflow-x-auto gap-1 max-h-60"
                    >
                        <div className="flex-none">Subdirectories:</div>
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
