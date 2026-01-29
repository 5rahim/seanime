import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { ScrollArea } from "@/components/ui/scroll-area"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import React from "react"
import { BiChevronRight, BiFolderOpen } from "react-icons/bi"
import { FaFolder } from "react-icons/fa"
import { FiFile, FiFolder } from "react-icons/fi"

type FileSelectorProps = {
    kind: "file" | "directory" | "both"
    onSelectPath: (path: string) => void
    selectedPath: string
    fileExtensions?: string[]
}

export function FileSelector(props: FileSelectorProps) {

    const {
        kind,
        onSelectPath,
        selectedPath,
        fileExtensions,
    } = props

    const [path, setPath] = React.useState<string>("")
    const debouncedPath = useDebounce(path, 500)
    const [modalOpen, setModalOpen] = React.useState(false)

    const firstRender = React.useRef(true)
    React.useLayoutEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        setPath(props.selectedPath)
    }, [props.selectedPath])

    React.useEffect(() => {
        if (path) {
            props.onSelectPath(path)
        }
    }, [path])

    const handleManualPathSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        const node = findNodeByPath(exampleData, path)
        if (node) {
            setPath(node.path)
            console.log("Manually selected:", node.path)
        } else {
            console.log("Path not found:", path)
            setPath("")
        }
    }

    return (
        <>
            <div className="relative">
                <TextInput
                    leftIcon={<FaFolder />}
                    value={path}
                    onValueChange={setPath}
                    // rightIcon={<div className="flex">
                    //     {isLoading ? null : (data?.exists ?
                    //         <BiCheck className="text-green-500" /> : shouldExist ?
                    //             <BiX className="text-red-500" /> : <BiFolderPlus />)}
                    // </div>}
                    // onBlur={checkDirectoryExists}
                />
                <BiFolderOpen
                    className="text-2xl cursor-pointer absolute z-[1] top-0 right-0"
                    onClick={() => setModalOpen(true)}
                />
            </div>

            <FileSelectorModal
                isOpen={modalOpen}
                onOpenChange={() => setModalOpen(!modalOpen)}
                kind={kind}
                onSelectPath={setPath}
                selectedPath={selectedPath}
                fileExtensions={fileExtensions}
            />
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

const exampleData: TreeNode = {
    name: "project",
    type: "directory",
    path: "/project",
    children: [
        {
            name: "src",
            type: "directory",
            path: "/project/src",
        },
        {
            name: "public",
            type: "directory",
            path: "/project/public",
        },
        { name: "package.json", type: "file", path: "/project/package.json" },
        { name: "README.md", type: "file", path: "/project/README.md" },
    ],
}

const findNodeByPath = (node: TreeNode, path: string): TreeNode | null => {
    if (node.path === path) return node
    if (node.children) {
        for (const child of node.children) {
            const found = findNodeByPath(child, path)
            if (found) return found
        }
    }
    return null
}

function FileSelectorModal(props: FileSelectorProps & { isOpen: boolean, onOpenChange: () => void }) {
    const { isOpen, onOpenChange, kind, selectedPath, onSelectPath, fileExtensions } = props

    return (
        <Modal
            title="Select a file or directory"
            open={isOpen}
            onOpenChange={onOpenChange}
            contentClass="max-w-3xl"
        >
            <div className="space-y-4">
                <TextInput
                    value={selectedPath}
                    onValueChange={onSelectPath}
                />

                <ScrollArea
                    className={cn(
                        "h-60 rounded-[--radius-md] border",
                    )}
                >
                    {exampleData.children?.map(node => (
                        <TreeNode
                            data={node}
                            level={0}
                            kind={kind}
                            onSelect={node => onSelectPath(node.path)}
                            selectedPath={selectedPath}
                            fileExtensions={fileExtensions}
                        />
                    ))}
                </ScrollArea>
            </div>
        </Modal>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type TreeNode = {
    name: string
    type: "file" | "directory"
    path: string
    children?: TreeNode[]
}

type TreeProps = {
    data: TreeNode
    kind: "file" | "directory" | "both"
    onSelect: (node: TreeNode) => void
    selectedPath: string | null
    fileExtensions?: string[]
}


const TreeNode: React.FC<TreeProps & { level: number }> = ({
    data,
    kind,
    onSelect,
    selectedPath,
    level,
    fileExtensions,
}) => {

    React.useEffect(() => {
        if (selectedPath && selectedPath.startsWith(data.path)) {

        }
    }, [selectedPath, data.path])

    const handleSelect = () => {
        if (kind === "both" || kind === data.type) {
            onSelect(data)
        }
    }

    const isSelectable = kind === "both" || kind === data.type
    const isSelected = selectedPath === data.path
    const isVisible = data.type === "directory" || kind !== "directory"
    const hasValidExtension = !fileExtensions ||
        data.type === "directory" ||
        fileExtensions.some(ext => data.name.endsWith(ext))

    if (!isVisible || !hasValidExtension) {
        return null
    }

    return (
        <div>
            <div
                className={cn(
                    "flex items-center",
                    (isSelectable && !isSelected) && "hover:bg-gray-950",
                    isSelected && "bg-gray-800",
                )}
            >
                <div
                    className={cn(
                        "flex items-center gap-2 py-1 px-2 w-full",
                        isSelectable && "cursor-pointer",
                    )}
                    onClick={handleSelect}
                >
                    <div className="flex items-center">
                        {data.type === "directory" ? (
                            <FiFolder className="w-4 h-4 text-[--brand]" />
                        ) : (
                            <FiFile className="w-4 h-4 text-[--muted]" />
                        )}
                    </div>
                    <span
                        className={cn(
                            isSelectable ? "cursor-pointer" : "cursor-default",
                        )}
                    >{data.name}</span>
                </div>

                <div className="flex flex-1"></div>

                {data.type === "directory" && <IconButton
                    intent="white-basic"
                    size="xs"
                    className="mr-2"
                    icon={<BiChevronRight />}
                    onClick={e => {
                        e.preventDefault()
                    }}
                />}

            </div>
        </div>
    )
}
