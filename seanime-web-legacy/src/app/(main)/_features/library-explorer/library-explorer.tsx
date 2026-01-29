import { useGetLibraryExplorerFileTree, useRefreshLibraryExplorerFileTree } from "@/api/generated/library_explorer.hooks"
import { AL_BaseAnime, Anime_LocalFile, Anime_LocalFileType, LibraryExplorer_FileTreeNodeJSON } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useDeleteLocalFiles, useUpdateLocalFileData, useUpdateLocalFiles } from "@/api/hooks/localfiles.hooks"
import { __unknownMedia_drawerIsOpen, UnknownMediaManager } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { __unmatchedFileManagerIsOpen, UnmatchedFileManager } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { __anilist_userAnimeMediaAtom } from "@/app/(main)/_atoms/anilist.atoms"
import { LibraryExplorerSuperUpdate } from "@/app/(main)/_features/library-explorer/library-explorer-super-update"
import { LibraryExplorerSuperUpdateDrawer } from "@/app/(main)/_features/library-explorer/library-explorer-super-update-drawer"
import {
    libraryExplorer_collectFilePaths,
    libraryExplorer_collectLocalFileNodes,
    libraryExplorer_getCheckboxState,
} from "@/app/(main)/_features/library-explorer/library-explorer.utils"
import { FilepathSelector } from "@/app/(main)/_features/media/_components/filepath-selector"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { SeaImage } from "@/components/shared/sea-image"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { ContextMenuItem, ContextMenuLabel, ContextMenuSeparator, ContextMenuSub, ContextMenuTrigger } from "@/components/ui/context-menu"
import { cn } from "@/components/ui/core/styling"
import { Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { Tooltip } from "@/components/ui/tooltip"
import { upath } from "@/lib/helpers/upath"
import { ContextMenuGroup } from "@radix-ui/react-context-menu"
import { useAtom, useAtomValue, useSetAtom } from "jotai"
import { pascalCase } from "pascal-case"
import React, { memo } from "react"
import { BiChevronDown, BiChevronRight, BiFolder, BiListCheck, BiLockOpenAlt, BiSearch } from "react-icons/bi"
import { FaRegEdit } from "react-icons/fa"
import { FiFolder, FiHardDrive } from "react-icons/fi"
import {
    LuClipboardPlus,
    LuClipboardX,
    LuEye,
    LuFilePen,
    LuFileQuestion,
    LuFileVideo2,
    LuFilter,
    LuFilterX,
    LuFolderSync,
    LuPlus,
    LuTrash2,
} from "react-icons/lu"
import { MdOutlineAdd, MdOutlineRemoveDone, MdVideoFile } from "react-icons/md"
import { RiFolderOpenFill } from "react-icons/ri"
import { VscVerified } from "react-icons/vsc"
import { useWindowSize } from "react-use"
import { Virtuoso, VirtuosoHandle } from "react-virtuoso"
import { toast } from "sonner"
import { localFileMetadataSchema } from "../../entry/_containers/episode-list/episode-item"
import { SeaContextMenu } from "../context-menu/sea-context-menu"
import { MediaEntryCard } from "../media/_components/media-entry-card"
import { useMediaPreviewModal } from "../media/_containers/media-preview-modal"
import {
    libraryExplorer_drawerOpenAtom,
    LibraryExplorer_Filter,
    libraryExplorer_isSelectingPathsAtom,
    libraryExplorer_matchLocalFilesAtom,
    libraryExplorer_openDirectoryAtom,
    libraryExplorer_selectedNodeAtom,
    libraryExplorer_selectedPathsAtom,
    libraryExplorer_superUpdateDrawerOpenAtom,
} from "./library-explorer.atoms"

interface FlattenedTreeItem {
    node: LibraryExplorer_FileTreeNodeJSON
    level: number
    index: number
}

function hasMatchingFiles(userMedia: AL_BaseAnime[] | undefined, node: LibraryExplorer_FileTreeNodeJSON, filter: LibraryExplorer_Filter): boolean {
    if (node.kind === "file") {
        switch (filter) {
            case "UNMATCHED":
                return !!node.localFile && !node.localFile?.mediaId && !node.localFile.ignored
            case "UNLOCKED":
                return !!node.localFile?.mediaId && !node.localFile.locked && !node.localFile.ignored
            case "IGNORED":
                return !!node.localFile?.ignored
            case "UNKNOWN_MEDIA":
                return !!node.localFile?.mediaId && !userMedia?.find(m => m.id === node.localFile!.mediaId) && !node.localFile?.ignored
            default:
                return true
        }
    }
    return node.children?.some(child => hasMatchingFiles(userMedia, child, filter)) ?? false
}

function nodeOrDescendantMatches(node: LibraryExplorer_FileTreeNodeJSON, searchTerm: string): boolean {
    if (node.name.toLowerCase().includes(searchTerm)) {
        return true
    }
    if (node.children) {
        return node.children.some(child => nodeOrDescendantMatches(child, searchTerm))
    }
    return false
}

// flatten the tree structure based on expanded nodes and search filter to avoid recursion
function flattenTreeNodes(
    userMedia: AL_BaseAnime[] | undefined,
    node: LibraryExplorer_FileTreeNodeJSON,
    expandedNodes: Set<string>,
    searchTerm: string,
    filter: LibraryExplorer_Filter = undefined,
    level: number = 0,
    result: FlattenedTreeItem[] = [],
): FlattenedTreeItem[] {
    const isDirectory = node.kind === "directory"
    const hasChildren = node.children && node.children.length > 0

    // Filter children based on search term first
    let filteredChildren = searchTerm && node.children
        ? node.children.filter(child => nodeOrDescendantMatches(child, searchTerm))
        : node.children

    // Apply additional filters based on file properties
    if (filter && filteredChildren) {
        // First check if this node should be included based on filter
        if (isDirectory) {
            // For directories, check if they have any valid files in their tree
            if (!hasMatchingFiles(userMedia, node, filter)) {
                return result
            }
        }

        // Filter the children
        filteredChildren = filteredChildren.filter(child => {
            if (child.kind === "directory") {
                return hasMatchingFiles(userMedia, child, filter)
            }

            // For files, apply the filter directly
            switch (filter) {
                case "UNMATCHED":
                    return !child.localFile?.mediaId && !child.localFile?.ignored
                case "UNLOCKED":
                    return !!child.localFile?.mediaId && !child.localFile.locked && !child.localFile?.ignored
                case "IGNORED":
                    return !!child.localFile?.ignored
                case "UNKNOWN_MEDIA":
                    return !!child.localFile?.mediaId && !userMedia?.find(m => m.id === child.localFile!.mediaId) && !child.localFile?.ignored
                default:
                    return true
            }
        })
    }

    // Check if node should be shown
    const shouldShow = !searchTerm ||
        node.name.toLowerCase().includes(searchTerm) ||
        (filteredChildren && filteredChildren.length > 0)

    if (!shouldShow) return result

    // Add current node to result
    result.push({
        node,
        level,
        index: result.length,
    })

    // Add children if directory is expanded
    if (isDirectory && hasChildren && expandedNodes.has(node.path)) {
        const childrenToProcess = filteredChildren || []
        childrenToProcess.forEach(child => {
            flattenTreeNodes(userMedia, child, expandedNodes, searchTerm, filter, level + 1, result)
        })
    }

    return result
}

export function LibraryExplorer() {
    const { data: fileTree, isLoading } = useGetLibraryExplorerFileTree()
    const refreshMutation = useRefreshLibraryExplorerFileTree()
    const [selectedNode, setSelectedNode] = useAtom(libraryExplorer_selectedNodeAtom)
    const [searchTerm, setSearchTerm] = React.useState("")
    const [expandedNodes, setExpandedNodes] = React.useState<Set<string>>(new Set())
    const [open] = useAtom(libraryExplorer_drawerOpenAtom)
    const [isSelectingPaths, setIsSelectingPaths] = useAtom(libraryExplorer_isSelectingPathsAtom)
    const [selectedFilter, setSelectedFilter] = React.useState<LibraryExplorer_Filter>(undefined)

    const userMedia = useAtomValue(__anilist_userAnimeMediaAtom)

    const [matchLocalFiles] = useAtom(libraryExplorer_matchLocalFilesAtom)

    const ref = React.useRef<VirtuosoHandle>(null)

    const { mutate: updateLocalFiles } = useUpdateLocalFiles()
    const { mutate: openInExplorer } = useOpenInExplorer()
    const [, setUnmatchedFileManagerOpen] = useAtom(__unmatchedFileManagerIsOpen)
    const [, setUnknownMediaManagerOpen] = useAtom(__unknownMedia_drawerIsOpen)
    const [, setMatchLocalFiles] = useAtom(libraryExplorer_matchLocalFilesAtom)

    const handleUnmatchFiles = (paths: string[]) => {
        updateLocalFiles({
            paths,
            action: "unmatch",
        })
    }

    const handleLockFiles = (paths: string[]) => {
        updateLocalFiles({
            paths,
            action: "lock",
        })
    }

    const handleUnlockFiles = (paths: string[]) => {
        updateLocalFiles({
            paths,
            action: "unlock",
        })
    }

    const handleIgnoreFiles = (paths: string[]) => {
        updateLocalFiles({
            paths,
            action: "ignore",
        })
    }

    const handleUnignoreFiles = (paths: string[]) => {
        updateLocalFiles({
            paths,
            action: "unignore",
        })
    }

    const handleMatchFiles = (nodes: LibraryExplorer_FileTreeNodeJSON[]) => {
        setMatchLocalFiles(nodes?.filter(n => n.localFile && !n.localFile?.mediaId)?.map(n => n.localFile!) ?? [])
        React.startTransition(() => {
            setUnmatchedFileManagerOpen(true)
        })
    }
    const handleResolveFileMedia = (lfs: Anime_LocalFile[]) => {
        setMatchLocalFiles(lfs ?? [])
        React.startTransition(() => {
            setUnknownMediaManagerOpen(true)
        })
    }

    const handleOpenInExplorer = (path: string) => {
        openInExplorer({ path })
    }

    const handleRefresh = () => {
        refreshMutation.mutate()
    }

    const handleToggleExpand = React.useCallback((nodePath: string) => {
        setExpandedNodes(prev => {
            const newSet = new Set(prev)
            if (newSet.has(nodePath)) {
                newSet.delete(nodePath)
            } else {
                newSet.add(nodePath)
            }
            return newSet
        })
    }, [])

    // collapse all nodes below unexpanded node if they are expanded
    const prevExpandedNodes = React.useRef<Set<string>>(expandedNodes)
    const t = React.useRef<NodeJS.Timeout | null>(null)

    React.useEffect(() => {
        if (t.current) {
            clearTimeout(t.current)
            t.current = null
        }
        t.current = setTimeout(() => {
            if (expandedNodes.size === 0) return
            const currentExpanded = expandedNodes
            const previousExpanded = prevExpandedNodes.current

            // Find nodes that were just collapsed
            const collapsedNodes = Array.from(previousExpanded).filter(path => !currentExpanded.has(path))

            if (collapsedNodes.length > 0) {
                const pathsToCollapse: string[] = []

                collapsedNodes.forEach(collapsedPath => {
                    // Find all expanded nodes that are children of the collapsed node
                    Array.from(currentExpanded).forEach(expandedPath => {
                        if (expandedPath.startsWith(collapsedPath + "/")) {
                            pathsToCollapse.push(expandedPath)
                        }
                    })
                })

                if (pathsToCollapse.length > 0) {
                    setExpandedNodes(prev => {
                        const newSet = new Set(prev)
                        pathsToCollapse.forEach(path => newSet.delete(path))
                        return newSet
                    })
                }
            }

            prevExpandedNodes.current = currentExpanded
        }, 100)

        return () => {
            if (t.current) {
                clearTimeout(t.current)
                t.current = null
            }
        }
    }, [expandedNodes])

    const handleSelectNode = React.useCallback((node: LibraryExplorer_FileTreeNodeJSON) => {
        setSelectedNode(node)
    }, [])

    const handleToggleSelectingPaths = React.useCallback(() => {
        setIsSelectingPaths(p => !p)
    }, [])

    const [selectedPaths, setSelectedPaths] = useAtom(libraryExplorer_selectedPathsAtom)

    React.useEffect(() => {
        if (!isSelectingPaths) {
            setSelectedPaths(new Set())
        }
    }, [isSelectingPaths])

    const handlePathSelection = React.useCallback((node: LibraryExplorer_FileTreeNodeJSON, checked: boolean) => {
        setSelectedPaths(prev => {
            const newPaths = new Set(prev)

            if (node.kind === "file") {
                // For files, simply add or remove the file path
                if (checked) {
                    newPaths.add(node.path)
                } else {
                    newPaths.delete(node.path)
                }
            } else {
                // For directories, add or remove all file paths within it
                const filePaths = libraryExplorer_collectFilePaths(node)
                if (checked) {
                    filePaths.forEach(path => newPaths.add(path))
                } else {
                    filePaths.forEach(path => newPaths.delete(path))
                }
            }

            return newPaths
        })
    }, [])

    const { width } = useWindowSize()

    const fileNodes = libraryExplorer_collectLocalFileNodes(fileTree?.root)
    const unignoredFileNodes = fileNodes?.filter(n => !n.localFile?.ignored)
    const hasUnscannedFiles = unignoredFileNodes?.some(n => !n.localFile)

    // Calculate flattened tree items for virtualization
    const flattenedItems = React.useMemo(() => {
        if (!fileTree?.root) return []
        return flattenTreeNodes(userMedia, fileTree.root, expandedNodes, searchTerm.toLowerCase(), selectedFilter)
    }, [fileTree?.root, expandedNodes, searchTerm, selectedFilter, userMedia])

    // Select directory user wants to open
    const [directoryToOpen, setDirectoryToOpen] = useAtom(libraryExplorer_openDirectoryAtom)

    const findNodeAndParents = (node: LibraryExplorer_FileTreeNodeJSON, path: string[] = []): {
        node: LibraryExplorer_FileTreeNodeJSON;
        parents: string[]
    } | null => {
        if (!directoryToOpen) return null
        if (node.path.toLowerCase() === directoryToOpen.toLowerCase()) {
            return { node, parents: path }
        }

        if (node.children) {
            for (const child of node.children) {
                const result = findNodeAndParents(child, [...path, node.path])
                if (result) return result
            }
        }
        return null
    }

    React.useEffect(() => {
        if (directoryToOpen && open && fileTree?.root) {
            // Find node and collect parent paths in one traversal

            const result = findNodeAndParents(fileTree.root)
            if (result) {
                setSelectedNode(result.node)

                // Expand all parent paths and the target directory
                const pathsToExpand = new Set([...result.parents, directoryToOpen])
                setExpandedNodes(prev => new Set([...prev, ...pathsToExpand]))

                // Scroll to node after expansion
                setTimeout(() => {
                    const updatedItems = flattenTreeNodes(userMedia, fileTree.root!, pathsToExpand, searchTerm.toLowerCase(), selectedFilter)
                    const nodeIndex = updatedItems.findIndex(n => n.node.path.toLowerCase() === directoryToOpen.toLowerCase())
                    if (nodeIndex >= 0) {
                        ref.current?.scrollToIndex({ index: nodeIndex, align: "start" })
                    }
                }, 500)
            }

            setDirectoryToOpen(null)
        }
    }, [directoryToOpen, open, fileTree?.root, findNodeAndParents, searchTerm, selectedFilter, userMedia])

    const hasUnlockedFiles = unignoredFileNodes?.some(n => n.localFile && !!n.localFile.mediaId && !n.localFile.locked)
    const unmatchedFiles = unignoredFileNodes?.filter(n => !!n.localFile && !n.localFile.mediaId)
    const unknownMediaFiles = unignoredFileNodes?.filter(n => !!n.localFile && !!n.localFile.mediaId && userMedia?.findIndex(m => m.id === n.localFile!.mediaId) === -1)

    const handleToggleFilter = (filter?: LibraryExplorer_Filter) => {
        if (!filter) {
            return setSelectedFilter(undefined)
        }

        setSelectedFilter(filter)
    }

    if (isLoading) {
        return (
            <div className="p-4 lg:p-8 flex-1 overflow-y-auto flex items-center justify-center">
                <LoadingSpinner />
            </div>
        )
    }

    return (
        <>
            <div className="hidden lg:flex h-full">
                <div className="flex-1 flex flex-col border-r bg-gray-950">
                    <div className="p-4 border-b space-y-3">
                        <div className="flex items-center gap-2">
                            <div className="flex items-center gap-3 flex-wrap">
                                <h2 className="text-lg font-semibold text-gray-100 2xl:block hidden">Library Explorer</h2>
                                {hasUnlockedFiles && (
                                    <Alert
                                        intent="info"
                                        className="text-sm py-1 px-3 cursor-pointer"
                                        description="Lock all correctly matched files"
                                        onClick={() => {
                                            setSelectedFilter("UNLOCKED")
                                        }}
                                    />
                                )}
                                {!!unmatchedFiles?.length && (
                                    <Alert
                                        intent="warning"
                                        className="text-sm py-1 px-3 cursor-pointer"
                                        description={`${unmatchedFiles.length} unmatched file${unmatchedFiles.length != 1 ? "s" : ""}`}
                                        onClick={() => {
                                            setSelectedFilter("UNMATCHED")
                                        }}
                                    />
                                )}
                                {!!unknownMediaFiles?.length && (
                                    <Alert
                                        intent="warning"
                                        className="text-sm py-1 px-3 cursor-pointer"
                                        description={`${unknownMediaFiles.length} file${unknownMediaFiles.length != 1 ? "s" : ""} with hidden media`}
                                        onClick={() => {
                                            setSelectedFilter("UNKNOWN_MEDIA")
                                        }}
                                    />
                                )}
                            </div>
                            <div className="flex flex-1"></div>
                            <LibraryExplorerBulkActions
                                fileNodes={fileNodes}
                                handleMatchFiles={handleMatchFiles}
                                handleUnmatchFiles={handleUnmatchFiles}
                                handleIgnoreFiles={handleIgnoreFiles}
                                handleUnignoreFiles={handleUnignoreFiles}
                            />
                            <LibraryExplorerSuperUpdate
                                fileNodes={fileNodes}
                            />
                            {!selectedFilter && <Popover
                                trigger={<Button
                                    leftIcon={<LuFilter className="text-xl" />}
                                    size="sm"
                                    intent={!!selectedFilter ? "white" : "gray-subtle"}
                                    onClick={() => {handleToggleFilter()}}
                                    className={cn(
                                        !!selectedFilter && "animate-pulse",
                                    )}
                                >
                                    Filter
                                </Button>}
                            >
                                <Button intent="gray-link" size="sm" className="w-full" onClick={() => handleToggleFilter("UNMATCHED")}>
                                    Unmatched files
                                </Button>
                                <Button intent="gray-link" size="sm" className="w-full" onClick={() => handleToggleFilter("UNLOCKED")}>
                                    Unlocked files
                                </Button>
                                <Button intent="gray-link" size="sm" className="w-full" onClick={() => handleToggleFilter("IGNORED")}>
                                    Ignored files
                                </Button>
                                <Button intent="gray-link" size="sm" className="w-full" onClick={() => handleToggleFilter("UNKNOWN_MEDIA")}>
                                    Unknown media
                                </Button>
                            </Popover>}
                            {!!selectedFilter && (
                                <Button
                                    leftIcon={<LuFilterX className="text-xl" />}
                                    size="sm"
                                    intent={"white"}
                                    onClick={() => {handleToggleFilter()}}
                                    className={cn(
                                        "animate-pulse",
                                    )}
                                >
                                    Filter: {!!selectedFilter ? pascalCase(selectedFilter) : ""}
                                </Button>
                            )}
                            <Button
                                leftIcon={<BiListCheck className="text-xl" />}
                                size="sm"
                                intent={isSelectingPaths ? "white" : "gray-subtle"}
                                onClick={handleToggleSelectingPaths}
                                className={cn(
                                    isSelectingPaths && "animate-pulse",
                                )}
                            >
                                Select{isSelectingPaths ? "ing" : ""}
                            </Button>
                            <IconButton
                                icon={<LuFolderSync />}
                                size="sm"
                                intent="gray-subtle"
                                onClick={handleRefresh}
                                loading={refreshMutation.isPending}
                            />
                        </div>
                        <TextInput
                            placeholder="Search files and folders..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            leftIcon={<BiSearch />}
                            size="sm"
                        />

                        {hasUnscannedFiles && (
                            <Alert
                                intent="warning"
                                description="Some files have not been scanned yet. Please scan the library to be able to perform actions on them."
                            />
                        )}
                    </div>

                    <div className="flex-1 library-explorer-tree-container">
                        <Virtuoso
                            ref={ref}
                            data={flattenedItems}
                            itemContent={(index, item) => (
                                <VirtualizedTreeNode
                                    key={item.node.path}
                                    item={item}
                                    isExpanded={expandedNodes.has(item.node.path)}
                                    onToggleExpand={handleToggleExpand}
                                    onSelect={handleSelectNode}
                                    selectedPath={selectedNode?.path}
                                    localFiles={fileTree?.localFiles}
                                    selectedPaths={selectedPaths}
                                    onPathSelection={handlePathSelection}
                                    windowWidth={width}
                                    onUnmatchFiles={handleUnmatchFiles}
                                    onLockFiles={handleLockFiles}
                                    onUnlockFiles={handleUnlockFiles}
                                    onIgnoreFiles={handleIgnoreFiles}
                                    onUnignoreFiles={handleUnignoreFiles}
                                    onMatchFiles={handleMatchFiles}
                                    onResolveMedia={handleResolveFileMedia}
                                    onOpenInExplorer={handleOpenInExplorer}
                                />
                            )}
                            style={{ height: "100%" }}
                            components={{
                                Footer: () => (
                                    <div className="p-2">
                                        <p className="text-xs text-gray-400 text-center py-2">
                                            End
                                        </p>
                                    </div>
                                ),
                            }}
                        />
                    </div>
                </div>

                <div className="flex flex-col flex-none bg-gray-950/50 w-80">
                    <LibraryInfoPanel localFiles={fileTree?.localFiles} />
                </div>
            </div>

            <UnmatchedFileManager
                unmatchedGroups={[
                    {
                        dir: upath.dirname(matchLocalFiles[0]?.path || ""),
                        localFiles: matchLocalFiles,
                    },
                ]}
            />
            <UnknownMediaManager
                unknownGroups={[
                    {
                        mediaId: matchLocalFiles[0]?.mediaId || 0,
                        localFiles: matchLocalFiles,
                    },
                ]}
                onActionComplete={() => {
                    setUnknownMediaManagerOpen(false)
                    setMatchLocalFiles([])
                }}
            />

            <LibraryExplorerSuperUpdateDrawer
                fileNodes={fileNodes}
            />
        </>
    )
}

type LibraryExplorerBulkActionsProps = {
    fileNodes: LibraryExplorer_FileTreeNodeJSON[]
    handleMatchFiles: (nodes: LibraryExplorer_FileTreeNodeJSON[]) => void
    handleUnmatchFiles: (paths: string[]) => void
    handleIgnoreFiles: (paths: string[]) => void
    handleUnignoreFiles: (paths: string[]) => void
}

export function LibraryExplorerBulkActions(props: LibraryExplorerBulkActionsProps) {
    const {
        fileNodes,
        handleMatchFiles,
        handleUnmatchFiles,
        handleIgnoreFiles,
        handleUnignoreFiles,
    } = props

    const [isSelectingPaths] = useAtom(libraryExplorer_isSelectingPathsAtom)
    const [selectedPaths] = useAtom(libraryExplorer_selectedPathsAtom)
    const [deleteModalOpen, setDeleteModalOpen] = React.useState(false)

    const selectedPathFileNodes = fileNodes?.filter(n => selectedPaths.has(n.path))
    const shouldShowUnmatchFiles = selectedPathFileNodes?.some(n => n.kind === "file" && !!n.localFile && !!n.localFile?.mediaId)
    const shouldShowIgnoreFiles = selectedPathFileNodes?.some(n => n.kind === "file" && !!n.localFile && !n.localFile.ignored)
    const shouldShowUnIgnoreFiles = selectedPathFileNodes?.every(n => n.kind === "file" && !!n.localFile && n.localFile.ignored)

    function handleMatchOrUnmatch() {
        if (shouldShowUnmatchFiles) {
            handleUnmatchFiles(Array.from(selectedPaths))
        } else {
            handleMatchFiles(selectedPathFileNodes)
        }
    }

    function handleToggleIgnore() {
        if (shouldShowIgnoreFiles) {
            handleIgnoreFiles(Array.from(selectedPaths))
        } else if (shouldShowUnIgnoreFiles) {
            handleUnignoreFiles(Array.from(selectedPaths))
        }
    }

    return (
        <>
            {isSelectingPaths && !!selectedPathFileNodes?.length && (
                <>
                    <Button
                        leftIcon={<MdOutlineRemoveDone className="text-xl" />}
                        size="sm"
                        intent={shouldShowUnmatchFiles ? "warning-link" : "success-link"}
                        onClick={handleMatchOrUnmatch}
                    >
                        {shouldShowUnmatchFiles ? "Unmatch" : "Match"} {selectedPathFileNodes.length} file{selectedPathFileNodes.length != 1
                        ? "s"
                        : ""}
                    </Button>
                    {(shouldShowIgnoreFiles || shouldShowUnIgnoreFiles) && <Button
                        leftIcon={<LuClipboardX className="text-xl" />}
                        size="sm"
                        intent={"gray-link"}
                        onClick={handleToggleIgnore}
                    >
                        {shouldShowIgnoreFiles ? "Ignore" : "Un-ignore"} {selectedPathFileNodes.length} file{selectedPathFileNodes.length != 1
                        ? "s"
                        : ""}
                    </Button>}
                    <Button
                        leftIcon={<LuTrash2 className="text-xl" />}
                        size="sm"
                        intent="alert-subtle"
                        onClick={() => setDeleteModalOpen(true)}
                    >
                        Delete {selectedPathFileNodes.length} file{selectedPathFileNodes.length != 1 ? "s" : ""}
                    </Button>
                </>
            )}
            <LibraryExplorerBulkDeleteModal
                open={deleteModalOpen}
                onOpenChange={setDeleteModalOpen}
                selectedPaths={Array.from(selectedPaths)}
                fileNodes={selectedPathFileNodes}
            />
        </>
    )
}

interface VirtualizedTreeNodeProps {
    item: FlattenedTreeItem
    isExpanded: boolean
    onToggleExpand: (path: string) => void
    onSelect: (node: LibraryExplorer_FileTreeNodeJSON) => void
    selectedPath?: string
    localFiles: Record<string, Anime_LocalFile> | undefined
    selectedPaths: Set<string>
    onPathSelection: (node: LibraryExplorer_FileTreeNodeJSON, checked: boolean) => void
    windowWidth: number
    onUnmatchFiles: (paths: string[]) => void
    onLockFiles: (paths: string[]) => void
    onUnlockFiles: (paths: string[]) => void
    onIgnoreFiles: (paths: string[]) => void
    onUnignoreFiles: (paths: string[]) => void
    onMatchFiles: (nodes: LibraryExplorer_FileTreeNodeJSON[]) => void
    onResolveMedia: (lfs: Anime_LocalFile[]) => void
    onOpenInExplorer: (path: string) => void
}

const VirtualizedTreeNode = memo(({
    item,
    isExpanded,
    onToggleExpand,
    onSelect,
    selectedPath,
    localFiles,
    selectedPaths,
    onPathSelection,
    windowWidth,
    onUnmatchFiles,
    onLockFiles,
    onUnlockFiles,
    onIgnoreFiles,
    onUnignoreFiles,
    onMatchFiles,
    onResolveMedia,
    onOpenInExplorer,
}: VirtualizedTreeNodeProps) => {
    const { node, level } = item
    const isDirectory = node.kind === "directory"
    const isSelected = selectedPath === node.path
    const hasDirectoryChildren = node.children && node.children.some(n => n.kind === "directory")
    const hasChildren = node.children && node.children.length > 0

    const userMedia = useAtomValue(__anilist_userAnimeMediaAtom)
    const isSelectingPaths = useAtomValue(libraryExplorer_isSelectingPathsAtom)
    const setSelectedPaths = useSetAtom(libraryExplorer_selectedPathsAtom)
    const setSuperUpdateDrawerOpen = useSetAtom(libraryExplorer_superUpdateDrawerOpenAtom)
    const [isMetadataModalOpen, setMetadataModalOpen] = React.useState(false)
    const { updateLocalFile } = useUpdateLocalFileData()

    const getFileIcon = () => {
        if (isDirectory) {
            if (level === 0) return <FiHardDrive className="size-5 text-brand-400" />
            return isExpanded ?
                <RiFolderOpenFill
                    className={cn(
                        "size-5 text-brand-400/90",
                        !node.mediaIds?.length && "text-[--muted]",
                    )}
                /> :
                <FiFolder
                    className={cn(
                        "size-5 text-brand-400/80",
                        !node.mediaIds?.length && "text-[--muted]",
                    )}
                />
        }

        if (!node.localFile) return <LuFileQuestion className="size-4 text-[--muted]" />

        return <LuFileVideo2
            className={cn(
                "size-4 text-[--muted]",
                node.localFile.metadata?.type === "main" && "text-brand-400/70",
                node.localFile.metadata?.type === "special" && "text-cyan-200/50",
                node.localFile.metadata?.type === "nc" && "text-gray-200/30",
            )}
        />
    }

    const handleClick = () => {
        onSelect(node)
        if (isDirectory && hasChildren) {
            onToggleExpand(node.path)
        }
    }

    const handleStopPropagation = React.useCallback((e: React.MouseEvent) => {
        e.stopPropagation()
    }, [])

    const handleCheckboxStopPropagation = React.useCallback((e: React.MouseEvent) => {
        e.stopPropagation()
    }, [])

    const handleCheckboxValueChange = React.useCallback((checked: boolean | "indeterminate") => {
        const isChecked = checked === true
        onPathSelection(node, isChecked)
    }, [node])

    const paddingLeft = level * 32 + 8

    const media = node.mediaIds?.length === 1 ? userMedia?.find(n => n.id === node.mediaIds?.[0]) : undefined
    const isUnknownMedia = node.mediaIds?.length === 1 && !media

    const fileNodes = libraryExplorer_collectLocalFileNodes(node)
    const nonIgnoredFileNodes = fileNodes?.filter(n => !n.localFile?.ignored)
    const matchedFileNodes = nonIgnoredFileNodes?.filter(n => !!n.localFile?.mediaId)

    const fileCount = nonIgnoredFileNodes?.length ?? 0
    const matchedFileCount = matchedFileNodes?.length ?? 0

    const allFileMatched = nonIgnoredFileNodes?.every(n => !!n.localFile?.mediaId) ?? false
    const allFileIgnored = fileNodes?.every(n => !!n.localFile?.ignored)
    const allFileScanned = nonIgnoredFileNodes?.every(n => !!n.localFile)

    const _isLocked = nonIgnoredFileNodes?.every(n => !!n.localFile?.locked || !n.localFile)
    const [isLocked, setOptimisticIsLocked] = React.useState(_isLocked ?? false)

    React.useEffect(() => {
        setOptimisticIsLocked(_isLocked ?? false)
    }, [_isLocked])

    const maxNameWidth = React.useMemo(() => {
        return windowWidth - paddingLeft - 620
    }, [windowWidth, paddingLeft, media?.title?.userPreferred])

    const { isPending } = useUpdateLocalFiles()

    const handleToggleLocked = () => {
        if (isPending) return
        if (isDirectory) {
            if (isLocked) {
                onUnlockFiles(libraryExplorer_collectFilePaths(node))
            } else {
                onLockFiles(libraryExplorer_collectFilePaths(node))
            }
        } else {
            if (node.localFile?.locked) {
                onUnlockFiles([node.path])
            } else {
                onLockFiles([node.path])
            }
        }
        setOptimisticIsLocked(p => !p)
    }

    const handleToggleLockedClick = (e: React.MouseEvent) => {
        e.stopPropagation()
        handleToggleLocked()
    }

    // Dropdown menu item handlers
    const handleOpenInExplorerClick = () => {
        onOpenInExplorer(node.path)
    }

    const handleUnmatchDirectory = () => {
        onUnmatchFiles(libraryExplorer_collectFilePaths(node))
    }

    const handleMatchDirectory = () => {
        onMatchFiles(libraryExplorer_collectLocalFileNodes(node))
    }

    const handleIgnoreDirectory = () => {
        onIgnoreFiles(libraryExplorer_collectFilePaths(node))
    }

    const handleUnignoreDirectory = () => {
        onUnignoreFiles(libraryExplorer_collectFilePaths(node))
    }

    const handleUnmatchSingleFile = () => {
        onUnmatchFiles([node.path])
    }

    const handleMatchSingleFile = () => {
        onMatchFiles(libraryExplorer_collectLocalFileNodes(node))
    }

    const handleIgnoreSingleFile = () => {
        onIgnoreFiles([node.path])
    }

    const handleUnignoreSingleFile = () => {
        onUnignoreFiles([node.path])
    }

    const [deleteModalOpen, setDeleteModalOpen] = React.useState(false)

    const handleDeleteDirectory = () => {
        setDeleteModalOpen(true)
    }

    const handleDeleteSingleFile = () => {
        setDeleteModalOpen(true)
    }

    const handleResolveMedia = () => {
        const id = media?.id ?? node.mediaIds?.[0] ?? 0
        if (!id || !localFiles) {
            toast.error("No media found")
            return
        }
        onResolveMedia(Object.values(localFiles)?.filter(n => n.mediaId === id) ?? [])
    }

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    function handleOpenMediaPreview() {
        const id = media?.id ?? node.mediaIds?.[0] ?? 0
        if (!id) {
            toast.error("No media found")
            return
        }
        setPreviewModalMediaId(id, "anime")
    }

    function handleOpenSuperUpdate() {
        if (isDirectory) {
            const paths = libraryExplorer_collectFilePaths(node)
            setSelectedPaths(new Set(paths))
            React.startTransition(() => {
                setSuperUpdateDrawerOpen(true)
            })
        } else {
            setSelectedPaths(new Set([node.path]))
            React.startTransition(() => {
                setSuperUpdateDrawerOpen(true)
            })
        }
    }

    const isScannedFile = !!node.localFile

    const [contextMenuOpen, setContextMenuOpen] = React.useState(false)

    const confirmLockDialog = useConfirmationDialog({
        title: "Lock all files",
        description: "This will lock all files in the directory. Are you sure you want to proceed?",
        actionText: "Lock all",
        actionIntent: "primary",
        onConfirm: async () => {
            handleToggleLockedClick(new MouseEvent("click") as any)
        },
    })
    const confirmUnlockDialog = useConfirmationDialog({
        title: "Unlock all files",
        description: "This will unlock all files in the directory. Are you sure you want to proceed?",
        actionText: "Unlock all",
        actionIntent: "primary",
        onConfirm: async () => {
            handleToggleLockedClick(new MouseEvent("click") as any)
        },
    })

    return (
        <div className="px-2">
            <ConfirmationDialog {...confirmLockDialog} />
            <ConfirmationDialog {...confirmUnlockDialog} />
            <SeaContextMenu
                onOpenChange={setContextMenuOpen}
                content={
                    <ContextMenuGroup>
                        <ContextMenuLabel className="text-[--muted] line-clamp-2 py-0 my-2 text-xs tracking-wide">
                            {node.name}
                        </ContextMenuLabel>
                        {node.mediaIds?.length === 1 && <ContextMenuItem
                            onClick={handleOpenMediaPreview}
                        >
                            <LuEye /> Preview anime
                        </ContextMenuItem>}
                        {isUnknownMedia && <ContextMenuItem
                            onClick={handleResolveMedia}
                        >
                            <LuPlus /> Resolve unknown media
                        </ContextMenuItem>}
                        <ContextMenuItem
                            onClick={handleOpenSuperUpdate}
                            // className={cn("text-[--violet]")}
                        >
                            <FaRegEdit /> Super update
                        </ContextMenuItem>
                        {(isDirectory && !!fileCount) && <>
                            {allFileMatched && <ContextMenuItem
                                onClick={handleUnmatchDirectory}
                                className={cn("text-[--orange]", isPending && "opacity-50 pointer-events-none")}
                            >
                                <MdOutlineRemoveDone className="text-lg" /> Unmatch files
                            </ContextMenuItem>}
                            {!allFileMatched && <ContextMenuItem
                                onClick={handleMatchDirectory}
                                className={cn("text-[--green]", isPending && "opacity-50 pointer-events-none")}
                            >
                                <MdOutlineAdd className="text-lg" /> Match files
                            </ContextMenuItem>}
                            {!allFileIgnored && <ContextMenuItem
                                onClick={handleIgnoreDirectory}
                                className={cn("", isPending && "opacity-50 pointer-events-none")}
                            >
                                <LuClipboardX className="text-lg" /> Ignore files
                            </ContextMenuItem>}
                            {allFileIgnored && <ContextMenuItem
                                onClick={handleUnignoreDirectory}
                                className={cn("text-purple-300", isPending && "opacity-50 pointer-events-none")}
                            >
                                <LuClipboardPlus className="text-lg" /> Un-ignore files
                            </ContextMenuItem>}
                        </>}
                        {(!isDirectory && isScannedFile) && <>
                            {!!node.localFile?.mediaId && <ContextMenuItem
                                onClick={() => setMetadataModalOpen(true)}
                                className={cn("text-[--blue]", isPending && "opacity-50 pointer-events-none")}
                            >
                                <LuFilePen className="text-lg" /> Edit metadata
                            </ContextMenuItem>}
                            {!!node.localFile?.mediaId && <ContextMenuItem
                                onClick={handleUnmatchSingleFile}
                                className={cn("text-[--orange]", isPending && "opacity-50 pointer-events-none")}
                            >
                                <MdOutlineRemoveDone className="text-lg" /> Unmatch file
                            </ContextMenuItem>}
                            {!node.localFile?.mediaId && <ContextMenuItem
                                onClick={handleMatchSingleFile}
                                className={cn("text-[--green]", isPending && "opacity-50 pointer-events-none")}
                            >
                                <MdOutlineAdd className="text-lg" /> Match file
                            </ContextMenuItem>}
                            {!node?.localFile?.ignored && <ContextMenuItem
                                onClick={handleIgnoreSingleFile}
                                className={cn("", isPending && "opacity-50 pointer-events-none")}
                            >
                                <LuClipboardX className="text-lg" /> Ignore file
                            </ContextMenuItem>}
                            {node?.localFile?.ignored && <ContextMenuItem
                                onClick={handleUnignoreSingleFile}
                                className={cn("text-purple-300", isPending && "opacity-50 pointer-events-none")}
                            >
                                <LuClipboardPlus className="text-lg" /> Un-ignore file
                            </ContextMenuItem>}
                        </>}
                        <ContextMenuSeparator className="!my-2" />
                        <ContextMenuSub
                            triggerContent="More"
                        >
                            {isDirectory && !!fileCount && <ContextMenuItem
                                onClick={handleDeleteDirectory}
                                className={cn("text-[--red]")}
                            >
                                <LuTrash2 className="text-lg" /> Delete files
                            </ContextMenuItem>}
                            {(!isDirectory && isScannedFile) && <ContextMenuItem
                                onClick={handleDeleteSingleFile}
                                className={cn("text-[--red]")}
                            >
                                <LuTrash2 className="text-lg" /> Delete file
                            </ContextMenuItem>}
                            <ContextMenuItem onClick={handleOpenInExplorerClick}>
                                <BiFolder /> Open in explorer
                            </ContextMenuItem>
                        </ContextMenuSub>

                    </ContextMenuGroup>
                }
            >
                <ContextMenuTrigger>
                    <div
                        className={cn(
                            "flex items-center px-2 h-10 rounded-md cursor-pointer select-none group/tree-node transition-colors",
                            !isSelected && "hover:bg-gray-800/50",
                            isSelected && "bg-brand-500/20 text-brand-100",
                            contextMenuOpen && "bg-gray-800/30",
                        )}
                        style={{ paddingLeft }}
                        onClick={handleClick}
                    >
                        <div className="flex items-center gap-2 flex-1 min-w-0">
                            {isDirectory && hasChildren && (
                                <div className="w-4 h-4 flex items-center justify-center">
                                    {isExpanded ? (
                                        <BiChevronDown className="size-4 text-gray-400" />
                                    ) : (
                                        <BiChevronRight className="size-4 text-gray-400" />
                                    )}
                                </div>
                            )}

                            {(isSelectingPaths && (isDirectory || isScannedFile)) && <div className="flex h-full items-center">
                                <Checkbox
                                    size="md"
                                    fieldClass="flex-shrink-0 flex items-center"
                                    value={libraryExplorer_getCheckboxState(node, selectedPaths)}
                                    onClick={handleCheckboxStopPropagation}
                                    onValueChange={handleCheckboxValueChange}
                                />
                            </div>}

                            <div className="flex-shrink-0">
                                {getFileIcon()}
                            </div>
                            {isDirectory && !!media &&
                                <div className="size-6 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden ml-2">
                                    <SeaImage
                                        src={media.coverImage?.medium || ""}
                                        alt={media.title?.userPreferred || ""}
                                        fill
                                        sizes="1rem"
                                        className="object-cover object-center"
                                    />
                                </div>}
                            {isDirectory && !media && !!node.mediaIds?.length && node.mediaIds?.length > 1 &&
                                <div className="h-6 px-2 flex-none rounded-[--radius-md] flex items-center justify-center bg-gray-800 ml-2 text-white text-xs font-semibold">
                                    {node.mediaIds?.length}
                                </div>}

                            <span
                                className="text-md tracking-wide text-gray-200 break-all group-hover/tree-node:text-gray-100 flex gap-1 items-center w-full overflow-hidden"
                                style={{ maxWidth: maxNameWidth + "px" }}
                            >
                                <span
                                    className={cn(
                                        "truncate min-w-0",
                                        !isDirectory && !!node.localFile && !node.localFile.mediaId && "text-orange-200",
                                        !isDirectory && !isScannedFile && "text-red-200",
                                        !isDirectory && node.localFile?.ignored && "text-[--muted] italic",
                                    )}
                                >{node.name === "root" ? "Anime Libraries" : node.name}</span>
                                {(!!media || isUnknownMedia) && (
                                    <span
                                        className={cn(
                                            "hidden tracking-normal 2xl:flex text-[--muted] text-sm flex-shrink whitespace-nowrap line-clamp-1 items-center gap-1",
                                        )}
                                        style={{ maxWidth: 200 }}
                                    >
                                        <span> - </span>
                                        <span>{!isUnknownMedia ? media?.title?.userPreferred : "(?)"}</span>
                                    </span>
                                )}
                                {isUnknownMedia && <Tooltip trigger={<Badge intent="unstyled">Unknown media</Badge>}>
                                    This media is not in your collection.
                                </Tooltip>}
                                {(allFileIgnored && isDirectory) && (
                                    <span
                                        className={cn(
                                            "hidden 2xl:flex text-[--muted] text-sm flex-shrink whitespace-nowrap line-clamp-1 items-center gap-1",
                                        )}
                                        style={{ maxWidth: 200 }}
                                    >
                                        <span> </span>
                                        <span>(Ignored)</span>
                                    </span>
                                )}
                            </span>
                        </div>

                        {isDirectory && !!fileCount && !allFileIgnored && (
                            <Tooltip
                                trigger={<span
                                    className={cn(
                                        "text-xs bg-green-500/20 text-[--green] px-2 py-0.5 rounded-full",
                                        fileCount !== matchedFileCount && "bg-orange-500/20 text-[--orange]",
                                        !allFileScanned && "bg-red-500/20 text-[--red]",
                                    )}
                                >
                                    {matchedFileCount} / {fileCount}
                                </span>}
                            >
                                Matched files
                            </Tooltip>
                        )}

                        {!isDirectory && isScannedFile && !node.localFile?.mediaId && !node.localFile?.ignored && (
                            <div className="text-xs text-orange-200">
                                Not matched
                            </div>
                        )}
                        {!isDirectory && !isScannedFile && (
                            <div className="text-xs text-red-200">
                                Not scanned
                            </div>
                        )}

                        {((isDirectory && !hasDirectoryChildren && matchedFileNodes?.length > 0) || (isScannedFile && !!node.localFile!.mediaId)) &&
                            <Tooltip
                                trigger={
                                    <IconButton
                                        icon={isLocked ? <VscVerified /> : <BiLockOpenAlt />}
                                        intent={isLocked ? "success-subtle" : "warning-subtle"}
                                        size={"xs"}
                                        className="hover:opacity-60 ml-2"
                                        loading={isPending}
                                        onClick={handleToggleLockedClick}
                                    />
                                }
                            >
                                {isLocked ? (isDirectory ? "Unlock all files" : "Unlock") : (isDirectory ? "Lock all files" : "Lock")}
                            </Tooltip>}
                        {((isDirectory && hasDirectoryChildren && matchedFileNodes?.length > 0)) && <Tooltip
                            trigger={
                                <IconButton
                                    icon={isLocked ? <VscVerified /> : <BiLockOpenAlt />}
                                    intent={isLocked ? "success-subtle" : "warning-subtle"}
                                    size={"xs"}
                                    className="hover:opacity-60 ml-2"
                                    loading={isPending}
                                    onClick={e => {
                                        e.stopPropagation()
                                        if (!isLocked) {
                                            confirmLockDialog.open()
                                        } else {
                                            confirmUnlockDialog.open()
                                        }
                                    }}
                                />
                            }
                        >
                            {isLocked ? "Unlock all files" : "Lock all files"}
                        </Tooltip>}
                    </div>
                </ContextMenuTrigger>
            </SeaContextMenu>

            {node.localFile && <Modal
                open={isMetadataModalOpen}
                onOpenChange={() => setMetadataModalOpen(false)}
                title="File metadata"
                titleClass="text-center"
                contentClass="max-w-xl"
            >
                <p className="w-full line-clamp-2 text-sm px-4 text-center py-2 flex-none">{node.localFile?.name}</p>
                <Form
                    schema={localFileMetadataSchema}
                    onSubmit={(data) => {
                        if (node.localFile) {
                            updateLocalFile(node.localFile, {
                                metadata: {
                                    ...node.localFile?.metadata,
                                    type: data.type as Anime_LocalFileType,
                                    episode: data.episode,
                                    aniDBEpisode: data.aniDBEpisode,
                                },
                            }, () => {
                                setMetadataModalOpen(false)
                                toast.success("Metadata saved")
                            })
                        }
                    }}
                    onError={console.log}
                    //@ts-ignore
                    defaultValues={{ ...node.localFile.metadata }}
                >
                    <Field.Number
                        label="Episode number" name="episode"
                        help="Relative episode number. If movie, episode number = 1"
                        required
                    />
                    <Field.Text
                        label="AniDB episode"
                        name="aniDBEpisode"
                        help="Specials typically contain the letter S"
                    />
                    <Field.Select
                        label="Type"
                        name="type"
                        options={[
                            { label: "Main", value: "main" },
                            { label: "Special", value: "special" },
                            { label: "NC/Other", value: "nc" },
                        ]}
                    />
                    <div className="w-full flex justify-end">
                        <Field.Submit role="save" intent="success" loading={isPending}>Save</Field.Submit>
                    </div>
                </Form>
            </Modal>}

            <LibraryExplorerDeleteFileModal
                open={deleteModalOpen}
                onOpenChange={setDeleteModalOpen}
                selectedPaths={isDirectory ? libraryExplorer_collectFilePaths(node) : [node.path]}
                isDirectory={isDirectory}
            />
        </div>
    )
})


function LibraryInfoPanel({}: { localFiles: Record<string, Anime_LocalFile> | undefined }) {
    const selectedNode = useAtomValue(libraryExplorer_selectedNodeAtom)

    const userMedia = useAtomValue(__anilist_userAnimeMediaAtom)

    const isDirectory = selectedNode?.kind === "directory"
    const fileExtension = isDirectory ? null : selectedNode?.name.split(".").pop()?.toUpperCase()

    const media = selectedNode?.mediaIds?.length === 1 ? userMedia?.find(n => n.id === selectedNode?.mediaIds?.[0]) : undefined
    const directoryCount = selectedNode?.children?.filter(c => c.kind === "directory")?.length ?? 0
    const fileCount = selectedNode?.children?.filter(c => c.kind === "file")?.length ?? 0

    const associatedMedia = userMedia?.filter(n => selectedNode?.mediaIds?.includes(n.id))

    const { setPreviewModalMediaId } = useMediaPreviewModal()

    if (!selectedNode) {
        return (
            <div className="p-4 flex-1 flex items-center justify-center">
                <div className="text-center text-gray-500">
                    <FiFolder className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p className="text-sm">Select a file or folder to view details</p>
                </div>
            </div>
        )
    }

    return (
        <div className="p-4 flex flex-col">
            <div className="flex flex-col items-center text-center mb-6">
                <div className="w-16 h-16 flex items-center justify-center mb-3">
                    {isDirectory ? (
                        <FiFolder
                            className={cn(
                                "w-12 h-12 text-brand-400",
                                !selectedNode.mediaIds?.length && "text-[--muted]",
                            )}
                        />
                    ) : (
                        <MdVideoFile
                            className={cn(
                                "w-12 h-12 text-[--muted]",
                                selectedNode.localFile?.metadata?.type === "main" && "text-brand-400/80",
                                selectedNode.localFile?.metadata?.type === "special" && "text-cyan-200/50",
                                selectedNode.localFile?.metadata?.type === "nc" && "text-gray-200/30",
                            )}
                        />
                    )}
                </div>
                <h3 className="font-semibold text-gray-100 break-all text-sm leading-tight">
                    {selectedNode.name}
                </h3>
                {fileExtension && (
                    <span className="text-xs text-gray-400 mt-1">{fileExtension} File</span>
                )}
            </div>

            <Separator className="mb-4" />

            <div className="space-y-3 text-sm">
                <div>
                    <dt className="text-gray-400 text-sm uppercase tracking-wide mb-1">Type</dt>
                    <dd className="text-gray-200">{isDirectory ? "Folder" : "File"}</dd>
                </div>

                <div>
                    <dt className="text-gray-400 text-sm uppercase tracking-wide mb-1">Path</dt>
                    <dd className="text-gray-200 text-xs font-mono bg-gray-900 p-2 rounded break-all">
                        {selectedNode.path}
                    </dd>
                </div>

                {selectedNode.size && (
                    <div>
                        <dt className="text-gray-400 text-sm uppercase tracking-wide mb-1">Size</dt>
                        <dd className="text-gray-200">{formatFileSize(selectedNode.size)}</dd>
                    </div>
                )}

                {isDirectory && selectedNode.mediaIds && selectedNode.mediaIds.length > 0 && (
                    <div>
                        <dt className="text-gray-400 text-sm uppercase tracking-wide mb-1">
                            Associated Media
                        </dt>
                        <dd className="text-gray-200">
                            {selectedNode.mediaIds.length} anime series
                        </dd>
                        {selectedNode.mediaIds.length > 1 && selectedNode.mediaIds.length <= 5 && associatedMedia?.map(media => (
                            <dd
                                key={media.id}
                                onClick={() => setPreviewModalMediaId(media.id, "anime")}
                                className="text-gray-200 cursor-pointer underline line-clamp-2 hover:opacity-80"
                            >
                                {media.title?.userPreferred}
                            </dd>
                        ))}
                    </div>
                )}

                {selectedNode.localFile && (
                    <div className="space-y-1.5">
                        <dt className="text-gray-400 text-sm uppercase tracking-wide">
                            Library File
                        </dt>
                        {selectedNode.localFile?.mediaId > 0 && <dd className="text-[--green] text-sm">✓ Matched</dd>}
                        {selectedNode.localFile?.mediaId === 0 && <dd className="text-[--orange] text-sm">Not matched</dd>}
                        <dd className="text-gray-200 text-sm">
                            Episode: <span className="font-semibold">{selectedNode.localFile?.metadata?.episode ?? "N/A"}</span>
                        </dd>
                        <dd className="text-gray-200 text-sm">
                            AniDB Episode: <span className="font-semibold">{selectedNode.localFile?.metadata?.aniDBEpisode ?? "N/A"}</span>
                        </dd>
                        <dd className="text-gray-200 text-sm">
                            Type: <span className="font-semibold">{selectedNode.localFile?.metadata?.type?.toUpperCase()}</span>
                        </dd>
                    </div>
                )}

                {isDirectory && selectedNode.children && (
                    <div>
                        <dt className="text-gray-400 text-sm uppercase tracking-wide mb-1">Contents</dt>
                        <dd className="text-gray-200">
                            {directoryCount} folder{directoryCount != 1 ? "s" : ""}, {" "}
                            {fileCount} file{fileCount != 1 ? "s" : ""}
                        </dd>
                    </div>
                )}

                {!!media && (
                    <div className="p-4">
                        <MediaEntryCard
                            media={media}
                            type="anime"
                            onClick={() => {
                                setPreviewModalMediaId(media.id, "anime")
                            }}
                        />
                    </div>
                )}
            </div>
        </div>
    )
}

function formatFileSize(bytes: number): string {
    const units = ["B", "KB", "MB", "GB", "TB"]
    let size = bytes
    let unitIndex = 0

    while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024
        unitIndex++
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`
}

type LibraryExplorerDeleteFileModalProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    selectedPaths: string[]
    isDirectory: boolean
}

function LibraryExplorerDeleteFileModal(props: LibraryExplorerDeleteFileModalProps) {
    const { open, onOpenChange, selectedPaths, isDirectory } = props

    const [filepaths, setFilepaths] = React.useState<string[]>([])

    React.useEffect(() => {
        if (open) {
            setFilepaths(selectedPaths)
        }
    }, [open, selectedPaths])

    const { mutate: deleteFiles, isPending: isDeleting } = useDeleteLocalFiles()

    const refreshMutation = useRefreshLibraryExplorerFileTree()

    const confirmDelete = useConfirmationDialog({
        title: "Delete file",
        description: "This action cannot be undone.",
        actionIntent: "alert",
        actionText: "Delete",
        onConfirm: () => {
            if (filepaths.length === 0) return

            deleteFiles({ paths: filepaths }, {
                onSuccess: () => {
                    onOpenChange(false)
                    refreshMutation.mutate()
                },
            })
        },
    })

    return (
        <>
            <Modal
                open={open}
                onOpenChange={onOpenChange}
                contentClass="max-w-2xl"
                title={<span>Delete file</span>}
                titleClass="text-center"
            >
                <div className="space-y-2 mt-2">
                    {isDirectory ? (
                        <FilepathSelector
                            className="max-h-96"
                            filepaths={filepaths}
                            allFilepaths={selectedPaths}
                            onFilepathSelected={setFilepaths}
                            showFullPath
                        />
                    ) : (
                        <div className="p-4 bg-gray-900 rounded-md">
                            <p className="text-sm text-gray-300 break-all">{selectedPaths[0]}</p>
                        </div>
                    )}

                    <div className="flex justify-end gap-2 mt-2">
                        <Button
                            intent="alert"
                            onClick={() => confirmDelete.open()}
                            loading={isDeleting}
                        >
                            Delete
                        </Button>
                        <Button
                            intent="white"
                            onClick={() => onOpenChange(false)}
                            disabled={isDeleting}
                        >
                            Cancel
                        </Button>
                    </div>
                </div>
            </Modal>
            <ConfirmationDialog {...confirmDelete} />
        </>
    )
}

type LibraryExplorerBulkDeleteModalProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    selectedPaths: string[]
    fileNodes: LibraryExplorer_FileTreeNodeJSON[]
}

function LibraryExplorerBulkDeleteModal(props: LibraryExplorerBulkDeleteModalProps) {
    const { open, onOpenChange, selectedPaths, fileNodes } = props

    const [filepaths, setFilepaths] = React.useState<string[]>([])
    const hasDirectories = fileNodes?.some(n => n.kind === "directory")

    React.useEffect(() => {
        if (open) {
            // Expand directories to their file paths
            const expandedPaths: string[] = []
            fileNodes?.forEach(node => {
                if (node.kind === "directory") {
                    expandedPaths.push(...libraryExplorer_collectFilePaths(node))
                } else {
                    expandedPaths.push(node.path)
                }
            })
            setFilepaths(expandedPaths)
        }
    }, [open, fileNodes])

    const { mutate: deleteFiles, isPending: isDeleting } = useDeleteLocalFiles()

    const refreshMutation = useRefreshLibraryExplorerFileTree()

    const confirmDelete = useConfirmationDialog({
        title: "Delete files",
        description: "This action cannot be undone.",
        actionIntent: "alert",
        actionText: "Delete",
        onConfirm: () => {
            if (filepaths.length === 0) return

            deleteFiles({ paths: filepaths }, {
                onSuccess: () => {
                    onOpenChange(false)
                    refreshMutation.mutate()
                },
            })
        },
    })

    const allFilepaths = React.useMemo(() => {
        const expandedPaths: string[] = []
        fileNodes?.forEach(node => {
            if (node.kind === "directory") {
                expandedPaths.push(...libraryExplorer_collectFilePaths(node))
            } else {
                expandedPaths.push(node.path)
            }
        })
        return expandedPaths
    }, [fileNodes])

    return (
        <>
            <Modal
                open={open}
                onOpenChange={onOpenChange}
                contentClass="max-w-2xl"
                title={<span>Select files to delete</span>}
                titleClass="text-center"
            >
                <div className="space-y-2 mt-2">
                    {hasDirectories ? (
                        <FilepathSelector
                            className="max-h-96"
                            filepaths={filepaths}
                            allFilepaths={allFilepaths}
                            onFilepathSelected={setFilepaths}
                            showFullPath
                        />
                    ) : (
                        <div className="space-y-2 max-h-96 overflow-y-auto">
                            {selectedPaths.map(path => (
                                <div key={path} className="p-2 bg-gray-900 rounded-md">
                                    <p className="text-sm text-gray-300 break-all">{path}</p>
                                </div>
                            ))}
                        </div>
                    )}

                    <div className="flex justify-end gap-2 mt-2">
                        <Button
                            intent="alert"
                            onClick={() => confirmDelete.open()}
                            loading={isDeleting}
                        >
                            Delete
                        </Button>
                        <Button
                            intent="white"
                            onClick={() => onOpenChange(false)}
                            disabled={isDeleting}
                        >
                            Cancel
                        </Button>
                    </div>
                </div>
            </Modal>
            <ConfirmationDialog {...confirmDelete} />
        </>
    )
}

