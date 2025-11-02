import { LibraryExplorer_FileTreeNodeJSON, Nullish } from "@/api/generated/types"

/**
 * Recursively collect all file paths from a node
 */
export function libraryExplorer_collectFilePaths(node: LibraryExplorer_FileTreeNodeJSON): string[] {
    const filePaths: string[] = []

    if (node.kind === "file") {
        filePaths.push(node.path)
    } else if (node.kind === "directory" && node.children) {
        for (const child of node.children) {
            filePaths.push(...libraryExplorer_collectFilePaths(child))
        }
    }

    return filePaths
}

/**
 * Recursively collect all local files from a node
 */
export function libraryExplorer_collectLocalFileNodes(node: Nullish<LibraryExplorer_FileTreeNodeJSON>): LibraryExplorer_FileTreeNodeJSON[] {
    if (!node) return []
    const ret: LibraryExplorer_FileTreeNodeJSON[] = []

    if (node.kind === "file") {
        ret.push(node)
    } else if (node.kind === "directory" && node.children) {
        for (const child of node.children) {
            ret.push(...libraryExplorer_collectLocalFileNodes(child))
        }
    }

    return ret
}


/**
 * Calculate checkbox state for a node based on selected paths
 */
export function libraryExplorer_getCheckboxState(node: LibraryExplorer_FileTreeNodeJSON, selectedPaths: Set<string>): boolean | "indeterminate" {
    if (node.kind === "file") {
        return selectedPaths.has(node.path)
    }

    // For directories, check all file children
    const allFilePaths = libraryExplorer_collectFilePaths(node)
    if (allFilePaths.length === 0) return false

    const selectedCount = allFilePaths.filter(path => selectedPaths.has(path)).length

    if (selectedCount === 0) return false
    if (selectedCount === allFilePaths.length) return true
    return "indeterminate"
}
