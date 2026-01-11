import { useDirectorySelector } from "@/api/hooks/directory_selector.hooks"
import { LibraryPathSelectionProps } from "@/app/(main)/_hooks/use-library-path-selection"
import { Button, IconButton } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Select } from "@/components/ui/select"
import { TextInput, TextInputProps } from "@/components/ui/text-input"
import { useBoolean } from "@/hooks/use-disclosure"
import { upath } from "@/lib/helpers/upath"
import React from "react"
import { BiCheck, BiFolderOpen, BiFolderPlus, BiX } from "react-icons/bi"
import { FaFolder } from "react-icons/fa"
import { FiChevronLeft, FiFolder } from "react-icons/fi"
import { HiMiniChevronUpDown } from "react-icons/hi2"
import { useUpdateEffect } from "react-use"
import { useDebounce } from "use-debounce"

export type DirectorySelectorProps = {
    defaultValue?: string
    onSelect: (path: string) => void
    shouldExist?: boolean
    value: string
    libraryPathSelectionProps?: LibraryPathSelectionProps
} & Omit<TextInputProps, "onSelect" | "value">

export const DirectorySelector = React.memo(React.forwardRef<HTMLInputElement, DirectorySelectorProps>(function (props: DirectorySelectorProps, ref) {

    const {
        defaultValue,
        onSelect,
        value,
        shouldExist,
        libraryPathSelectionProps: libraryProps,
        label,
        ...rest
    } = props

    const firstRender = React.useRef(true)

    const sanitizePath = React.useCallback((path: string) => {
        if (!path) return ""
        return upath.normalizeSafe(path.replace(/[<>"]/g, ""))
    }, [])

    const [input, setInputRaw] = React.useState(defaultValue ? sanitizePath(defaultValue) : "")
    const [debouncedInput] = useDebounce(input, 300)
    const selectorState = useBoolean(false)
    const prevState = React.useRef<string>(input)
    const currentState = React.useRef<string>(input)

    const setInput = React.useCallback((newInput: string) => {
        setInputRaw(sanitizePath(newInput))
    }, [sanitizePath])

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
        if (value !== currentState.current) {
            setInput(value)
        }
    }, [value])

    React.useEffect(() => {
        currentState.current = input
        if (input === ".") {
            setInputRaw("")
        }
    }, [input])

    useUpdateEffect(() => {
        const trimmedValue = debouncedInput.trim()
        onSelect(trimmedValue)
        prevState.current = trimmedValue

        // if (!isLoading && data && shouldExist && !data.exists && input.length > 0) {
        //     onSelect("")
        // }
    }, [debouncedInput, data])

    const checkDirectoryExists = React.useCallback(() => {
        if (!isLoading && data && shouldExist && !data.exists && input.length > 0) {
            React.startTransition(() => {
                setInputRaw("")
            })
        }
    }, [isLoading, data, input, shouldExist, prevState.current])

    const [librarySelectionOpen, setLibrarySelectionOpen] = React.useState(false)

    return (
        <>
            <div className="space-y-1">
                <div className="relative">
                    <TextInput
                        leftIcon={<FaFolder />}
                        {...rest}
                        label={<div className="flex items-center gap-1">
                            {label}
                            {libraryProps?.showLibrarySelector && (
                                <Popover
                                    open={librarySelectionOpen}
                                    onOpenChange={setLibrarySelectionOpen}
                                    className="w-[400px] p-2 ml-[30px]"
                                    sideOffset={-4}
                                    trigger={<Button size="sm" intent="gray-link" leftIcon={<HiMiniChevronUpDown />} className="!text-[--muted]">
                                        Change library
                                    </Button>}
                                >
                                    <Select
                                        value={libraryProps.selectedLibrary}
                                        options={libraryProps.libraryOptions}
                                        onValueChange={v => {
                                            libraryProps.handleLibraryPathSelect(v)
                                            setLibrarySelectionOpen(false)
                                        }}
                                    />
                                </Popover>
                            )}
                        </div>}
                        value={input}
                        rightIcon={<div className="flex">
                            {isLoading ? null : (data?.exists ?
                                <BiCheck className="text-green-500" /> : shouldExist ?
                                    input.length > 0 ? <BiX className="text-red-500" /> : null : <BiFolderPlus />)}
                        </div>}
                        onChange={e => {
                            setInput(e.target.value ?? "")
                        }}
                        ref={ref}
                        onBlur={checkDirectoryExists}
                    />

                    <div className="absolute z-[1] top-0 right-0 flex items-center">
                        <BiFolderOpen
                            className="text-2xl cursor-pointer"
                            onClick={selectorState.on}
                        />
                    </div>
                </div>
            </div>
            <Modal
                open={selectorState.active}
                onOpenChange={v => {
                    selectorState.toggle()
                    if (!v) {
                        checkDirectoryExists()
                    }
                }}
                title="Select a directory"
                contentClass="mt-4 space-y-2 max-w-4xl"
            >
                <div className="flex gap-2 items-center">
                    <IconButton
                        onClick={() => data?.basePath && setInput(data?.basePath)}
                        intent="gray-basic"
                        rounded
                        size="sm"
                        icon={<FiChevronLeft />}
                        disabled={(!data?.basePath?.length || data?.basePath?.length === 1)}
                    />
                    <TextInput
                        leftIcon={<FaFolder />}
                        value={input}
                        rightIcon={isLoading ? null : (data?.exists ?
                            <BiCheck className="text-green-500" /> : shouldExist ?
                                <BiX className="text-red-500" /> : <BiFolderPlus />)}
                        onChange={e => {
                            setInput(e.target.value ?? "")
                        }}
                        onClick={() => {
                            if (shouldExist) selectorState.on()
                        }}
                        ref={ref}
                    />
                </div>

                {(!data?.exists && data?.suggestions && data.suggestions.length > 0) &&
                    <div
                        className="w-full flex flex-none flex-nowrap overflow-x-auto gap-2 items-center rounded-[--radius-md]"
                    >
                        <div className="flex-none">Suggestions:</div>
                        {data.suggestions.map(folder => (
                            <div
                                key={folder.fullPath}
                                className="py-1 flex items-center gap-2 text-sm px-3 rounded-[--radius-md] border flex-none cursor-pointer bg-gray-900 hover:bg-gray-800"
                                onClick={() => setInput(folder.fullPath)}
                            >
                                <FiFolder className="w-4 h-4 text-[--brand]" />
                                <span className="break-normal">{folder.folderName}</span>
                            </div>
                        ))}
                    </div>}


                {(data && !!data?.content?.length) &&
                    <ScrollArea
                        className="h-60 rounded-[--radius-md] border !mt-0"
                    >
                        {data.content.map(folder => (
                            <div
                                key={folder.fullPath}
                                className="flex items-center gap-2 py-2 px-3 cursor-pointer hover:bg-gray-800"
                                onClick={() => setInput(folder.fullPath)}
                            >
                                <FiFolder className="w-4 h-4 text-[--brand]" />
                                <span className="break-normal">{folder.folderName}</span>
                            </div>
                        ))}
                    </ScrollArea>}
            </Modal>
        </>
    )

}))
