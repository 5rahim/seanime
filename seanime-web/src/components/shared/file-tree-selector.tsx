import { cn } from "@/components/ui/core/styling"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import React from "react"
import { FcFolder } from "react-icons/fc"
import { FiChevronDown, FiChevronRight, FiFile, FiSearch } from "react-icons/fi"
import { MdVerified } from "react-icons/md"

const filterFilePreviews = (filePreviews: any[], searchTerm: string): any[] => {
    if (!searchTerm.trim()) {
        return filePreviews
    }

    const lowerSearchTerm = searchTerm.toLowerCase()
    return filePreviews.filter(filePreview => {
        const searchableText = [
            filePreview.displayTitle,
            filePreview.displayPath,
            filePreview.path,
        ].join(" ").toLowerCase()

        return searchableText.includes(lowerSearchTerm)
    })
}

export type FileTreeNode = {
    name: string
    type: "file" | "directory"
    path: string
    filePreview?: {
        displayTitle: string
        displayPath: string
        path: string
        isLikely: boolean
    }
    children?: FileTreeNode[]
}

export const buildFileTree = (filePreviews: any[]): FileTreeNode => {
    const root: FileTreeNode = {
        name: "root",
        type: "directory",
        path: "",
        children: [],
    }

    const sortedPreviews = filePreviews.toSorted((a, b) => a.path.localeCompare(b.path))

    sortedPreviews.forEach(filePreview => {
        const pathParts = filePreview.path.split("/").filter((part: string) => part !== "")
        let currentNode = root

        pathParts.forEach((part: string, index: number) => {
            const isFile = index === pathParts.length - 1
            const currentPath = pathParts.slice(0, index + 1).join("/")

            let existingNode = currentNode.children?.find(child => child.name === part)

            if (!existingNode) {
                existingNode = {
                    name: part,
                    type: isFile ? "file" : "directory",
                    path: currentPath,
                    filePreview: isFile ? filePreview : undefined,
                    children: isFile ? undefined : [],
                }
                currentNode.children?.push(existingNode)
            }

            if (!isFile) {
                currentNode = existingNode
            }
        })
    })

    return root
}

export type FileTreeSelectorProps = {
    filePreviews: any[]
    selectedValue: string | number
    onFileSelect: (value: string | number) => void
    getFileValue: (filePreview: any) => string | number
    hasLikelyMatch: boolean
    hasOneLikelyMatch: boolean
    likelyMatchRef: React.RefObject<HTMLDivElement>
}

type FileTreeNodeProps = {
    node: FileTreeNode
    selectedValue: string | number
    onFileSelect: (value: string | number) => void
    getFileValue: (filePreview: any) => string | number
    hasLikelyMatch: boolean
    hasOneLikelyMatch: boolean
    likelyMatchRef: React.RefObject<HTMLDivElement>
    level?: number
}

const FileTreeNodeComponent: React.FC<FileTreeNodeProps> = ({
    node,
    selectedValue,
    onFileSelect,
    getFileValue,
    hasLikelyMatch,
    hasOneLikelyMatch,
    likelyMatchRef,
    level = 0,
}) => {
    const [isOpen, setIsOpen] = React.useState(level === 0 || level === 1)

    const toggleOpen = (e: React.MouseEvent) => {

    }

    const handleFileSelect = (e: React.MouseEvent) => {
        e.stopPropagation()
        if (node.type === "file" && node.filePreview) {
            onFileSelect(getFileValue(node.filePreview))
        }
        if (node.type === "directory") {
            setIsOpen(!isOpen)
        }
    }

    const isSelected = node.type === "file" && node.filePreview && selectedValue === getFileValue(node.filePreview)
    const isLikelyMatch = node.type === "file" && node.filePreview?.isLikely

    return (
        <div>
            <div
                className={cn(
                    "flex items-center py-1.5 px-2 border border-transparent rounded-[--radius]",
                    node.type === "file" && "cursor-pointer",
                    node.type === "file" && !isSelected && "hover:bg-[--subtle]",
                    isSelected && "bg-white dark:bg-gray-950 border border-gray-400",
                    (hasLikelyMatch && !isSelected && !isLikelyMatch && node.type === "file") && "opacity-60",
                )}
                onClick={handleFileSelect}
                ref={hasOneLikelyMatch && isLikelyMatch ? likelyMatchRef : undefined}
            >
                <div className="flex items-center">
                    {node.type === "directory" && (
                        <span className="mr-1 cursor-pointer">
                            {isOpen ? (
                                <FiChevronDown className="size-5" />
                            ) : (
                                <FiChevronRight className="size-5" />
                            )}
                        </span>
                    )}
                    {node.type === "directory" ? (
                        <FcFolder className="size-5 mr-2 text-[--white] cursor-pointer" />
                    ) : (
                        <FiFile className="size-5 mr-2 text-[--muted]" />
                    )}
                </div>

                <div className="flex flex-col flex-1 min-w-0 cursor-pointer">
                    {node.type === "file" && node.filePreview ? (
                        <>
                            <p className="mb-1 line-clamp-1 font-medium">
                                {node.filePreview.displayTitle}
                            </p>
                            {isLikelyMatch && (
                                <p className="flex items-center">
                                    <MdVerified className="text-[--green] mr-1" />
                                    <span className="text-white text-sm">Likely match</span>
                                </p>
                            )}
                            <p className="font-normal line-clamp-2 text-sm text-[--muted]">{node.filePreview.displayPath}</p>
                        </>
                    ) : (
                        <span
                            className={cn(
                                "font-medium",
                                node.type === "directory" ? "text-[--white]" : "cursor-pointer",
                            )}
                        >
                            {node.name}
                        </span>
                    )}
                </div>
            </div>

            {node.type === "directory" && isOpen && node.children && (
                <div className="ml-4 border-l pl-2">
                    {node.children.map((child, index) => (
                        <FileTreeNodeComponent
                            key={index}
                            node={child}
                            selectedValue={selectedValue}
                            onFileSelect={onFileSelect}
                            getFileValue={getFileValue}
                            hasLikelyMatch={hasLikelyMatch}
                            hasOneLikelyMatch={hasOneLikelyMatch}
                            likelyMatchRef={likelyMatchRef}
                            level={level + 1}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}

export const FileTreeSelector: React.FC<FileTreeSelectorProps> = ({
    filePreviews,
    selectedValue,
    onFileSelect,
    getFileValue,
    hasLikelyMatch,
    hasOneLikelyMatch,
    likelyMatchRef,
}) => {
    const [searchTerm, setSearchTerm] = React.useState("")
    const debouncedSearchTerm = useDebounce(searchTerm, 300)

    if (!filePreviews || filePreviews.length === 0) {
        return null
    }

    const filteredFilePreviews = filterFilePreviews(filePreviews, debouncedSearchTerm)
    const fileTree = buildFileTree(filteredFilePreviews)

    return (
        <div className="flex flex-col gap-3">
            <TextInput
                value={searchTerm}
                onValueChange={setSearchTerm}
                placeholder="Search files..."
                className="focus:ring-0 active:ring-0"
            />

            <div className="flex flex-col gap-1">
                {fileTree.children && fileTree.children.length > 0 ? (
                    fileTree.children.map((child, index) => (
                        <FileTreeNodeComponent
                            key={index}
                            node={child}
                            selectedValue={selectedValue}
                            onFileSelect={onFileSelect}
                            getFileValue={getFileValue}
                            hasLikelyMatch={hasLikelyMatch || false}
                            hasOneLikelyMatch={hasOneLikelyMatch || false}
                            likelyMatchRef={likelyMatchRef}
                            level={0}
                        />
                    ))
                ) : debouncedSearchTerm.trim() ? (
                    <div className="text-center py-8 text-[--muted]">
                        <FiSearch className="mx-auto mb-2 size-8 opacity-50" />
                        <p>No files found matching "{debouncedSearchTerm}"</p>
                    </div>
                ) : null}
            </div>
        </div>
    )
}

export type FileTreeMultiSelectorProps = {
    filePreviews: any[]
    selectedIndices: number[]
    onSelectionChange: (indices: number[]) => void
    getFileValue: (filePreview: any) => number
}

type FileTreeMultiNodeProps = {
    node: FileTreeNode
    selectedIndices: number[]
    onSelectionChange: (indices: number[]) => void
    getFileValue: (filePreview: any) => number
    level?: number
}

const FileTreeMultiNodeComponent: React.FC<FileTreeMultiNodeProps> = ({
    node,
    selectedIndices,
    onSelectionChange,
    getFileValue,
    level = 0,
}) => {
    const [isOpen, setIsOpen] = React.useState(level === 0 || level === 1)

    const toggleOpen = (e: React.MouseEvent) => {
        e.stopPropagation()
        setIsOpen(!isOpen)
    }

    const handleFileSelection = React.useCallback((checked: boolean) => {
        if (node.type === "file" && node.filePreview) {
            const fileValue = getFileValue(node.filePreview)
            let newIndices: number[]

            if (checked) {
                // Add the file if not already selected
                newIndices = selectedIndices.includes(fileValue)
                    ? selectedIndices
                    : [...selectedIndices, fileValue]
            } else {
                // Remove the file
                newIndices = selectedIndices.filter(idx => idx !== fileValue)
            }

            onSelectionChange(newIndices)
        }
    }, [node, selectedIndices, onSelectionChange, getFileValue])

    const handleDirectorySelection = React.useCallback((checked: boolean) => {
        if (node.type === "directory" && node.children) {
            const getAllFileIndicesInDirectory = (n: FileTreeNode): number[] => {
                const indices: number[] = []
                if (n.type === "file" && n.filePreview) {
                    indices.push(getFileValue(n.filePreview))
                } else if (n.children) {
                    n.children.forEach(child => {
                        indices.push(...getAllFileIndicesInDirectory(child))
                    })
                }
                return indices
            }

            const directoryFileIndices = getAllFileIndicesInDirectory(node)

            if (checked) {
                // Add all directory files that aren't already selected
                const newIndices = [...selectedIndices]
                directoryFileIndices.forEach(idx => {
                    if (!newIndices.includes(idx)) {
                        newIndices.push(idx)
                    }
                })
                onSelectionChange(newIndices)
            } else {
                // Remove all directory files
                onSelectionChange(selectedIndices.filter(idx => !directoryFileIndices.includes(idx)))
            }
        }
    }, [node, selectedIndices, onSelectionChange, getFileValue])

    const isFileSelected = node.type === "file" && node.filePreview && selectedIndices.includes(getFileValue(node.filePreview))

    // Check if directory is fully/partially selected
    const getDirectorySelectionState = React.useMemo(() => {
        if (node.type === "file") return { isSelected: isFileSelected, isPartial: false }

        const getAllFileIndicesInDirectory = (n: FileTreeNode): number[] => {
            const indices: number[] = []
            if (n.type === "file" && n.filePreview) {
                indices.push(getFileValue(n.filePreview))
            } else if (n.children) {
                n.children.forEach(child => {
                    indices.push(...getAllFileIndicesInDirectory(child))
                })
            }
            return indices
        }

        const directoryFileIndices = getAllFileIndicesInDirectory(node)
        const selectedCount = directoryFileIndices.filter(idx => selectedIndices.includes(idx)).length

        return {
            isSelected: selectedCount === directoryFileIndices.length && directoryFileIndices.length > 0,
            isPartial: selectedCount > 0 && selectedCount < directoryFileIndices.length,
        }
    }, [node, selectedIndices, getFileValue])

    return (
        <div>
            <div
                className={cn(
                    "flex items-center py-1.5 px-2 border rounded-[--radius] cursor-pointer transition-colors",
                    // File selection styles
                    node.type === "file" && isFileSelected && "border bg-gray-900 text-white",
                    node.type === "file" && !isFileSelected && "border-transparent",
                    // Directory selection styles
                    node.type === "directory" && getDirectorySelectionState.isSelected && "border bg-gray-900",
                    node.type === "directory" && getDirectorySelectionState.isPartial && "bg-gray-900",
                    node.type === "directory" && !getDirectorySelectionState.isSelected && !getDirectorySelectionState.isPartial && "border-transparent",
                )}
                onClick={(e) => {
                    e.stopPropagation()

                    if (node.type === "file") {
                        const currentlySelected = isFileSelected
                        handleFileSelection(!currentlySelected)
                    } else {
                        const currentlySelected = getDirectorySelectionState.isSelected
                        handleDirectorySelection(!currentlySelected)
                    }
                }}
            >
                <div className="flex items-center">
                    {node.type === "directory" && (
                        <span
                            className="mr-1 cursor-pointer" onClick={(e) => {
                            e.stopPropagation()
                            toggleOpen(e)
                        }}
                        >
                            {isOpen ? (
                                <FiChevronDown className="size-5" />
                            ) : (
                                <FiChevronRight className="size-5" />
                            )}
                        </span>
                    )}
                    {node.type === "directory" ? (
                        <FcFolder
                            className="size-5 mr-2 text-[--white] cursor-pointer"
                            onClick={(e) => {
                                e.stopPropagation()
                                toggleOpen(e)
                            }}
                        />
                    ) : (
                        <FiFile className="size-5 mr-2 text-[--muted]" />
                    )}
                </div>

                <div className="flex flex-col flex-1 min-w-0">
                    {node.type === "file" && node.filePreview ? (
                        <>
                            <p className="mb-1 line-clamp-1 font-medium">
                                {node.filePreview.displayTitle}
                            </p>
                            <p className="font-normal line-clamp-2 text-sm text-[--muted]">{node.filePreview.displayPath}</p>
                        </>
                    ) : (
                        <span className="font-medium text-[--white]">
                            {node.name}
                        </span>
                    )}
                </div>

                {/* Selection indicator */}
                <div className="ml-2 flex items-center">
                    {node.type === "file" && isFileSelected && (
                        <div className="w-2 h-2 bg-brand rounded-full" />
                    )}
                    {node.type === "directory" && getDirectorySelectionState.isSelected && (
                        <div className="w-2 h-2 bg-brand rounded-full" />
                    )}
                    {node.type === "directory" && getDirectorySelectionState.isPartial && (
                        <div className="w-2 h-2 bg-brand/60 rounded-full" />
                    )}
                </div>
            </div>

            {node.type === "directory" && isOpen && node.children && (
                <div className="ml-4 border-l pl-2">
                    {node.children.map((child, index) => (
                        <FileTreeMultiNodeComponent
                            key={index}
                            node={child}
                            selectedIndices={selectedIndices}
                            onSelectionChange={onSelectionChange}
                            getFileValue={getFileValue}
                            level={level + 1}
                        />
                    ))}
                </div>
            )}
        </div>
    )
}

export const FileTreeMultiSelector: React.FC<FileTreeMultiSelectorProps> = ({
    filePreviews,
    selectedIndices,
    onSelectionChange,
    getFileValue,
}) => {
    const [searchTerm, setSearchTerm] = React.useState("")
    const debouncedSearchTerm = useDebounce(searchTerm, 300)

    if (!filePreviews || filePreviews.length === 0) {
        return null
    }

    const filteredFilePreviews = filterFilePreviews(filePreviews, debouncedSearchTerm)
    const fileTree = buildFileTree(filteredFilePreviews)

    return (
        <div className="flex flex-col gap-3">
            <TextInput
                value={searchTerm}
                onValueChange={setSearchTerm}
                placeholder="Search files..."
                className="focus:ring-0 active:ring-0"
            />

            <div className="flex flex-col gap-2">
                {fileTree.children && fileTree.children.length > 0 ? (
                    fileTree.children.map((child, index) => (
                        <FileTreeMultiNodeComponent
                            key={index}
                            node={child}
                            selectedIndices={selectedIndices}
                            onSelectionChange={onSelectionChange}
                            getFileValue={getFileValue}
                            level={0}
                        />
                    ))
                ) : debouncedSearchTerm.trim() ? (
                    <div className="text-center py-8 text-[--muted]">
                        <FiSearch className="mx-auto mb-2 size-8 opacity-50" />
                        <p>No files found matching "{debouncedSearchTerm}"</p>
                    </div>
                ) : null}
            </div>
        </div>
    )
}
