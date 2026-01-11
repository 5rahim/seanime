import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { upath } from "@/lib/helpers/upath"
import React from "react"

const CUSTOM_VALUE = "__custom__"

type UseLibraryPathSelectorOptions = {
    destination: string
    setDestination: (path: string) => void
    animeFolderName?: string
}

export function useLibraryPathSelection(options: UseLibraryPathSelectorOptions) {
    const { destination, setDestination, animeFolderName } = options

    const serverStatus = useServerStatus()

    const allLibraryPaths = React.useMemo(() => {
        const libraryPath = serverStatus?.settings?.library?.libraryPath
        const additionalLibraryPaths = serverStatus?.settings?.library?.libraryPaths
        const paths: string[] = []
        if (libraryPath) paths.push(libraryPath)
        if (additionalLibraryPaths?.length) {
            paths.push(...additionalLibraryPaths.filter(p => p && p !== libraryPath))
        }
        return paths
    }, [serverStatus?.settings?.library])

    const selectedLibrary = React.useMemo(() => {
        const sortedPaths = [...allLibraryPaths].sort((a, b) => b.length - a.length)
        const matchingPath = sortedPaths.find(p => destination.startsWith(p))
        return matchingPath ?? CUSTOM_VALUE
    }, [allLibraryPaths, destination])

    const libraryOptions = React.useMemo(() => {
        return allLibraryPaths.map(p => ({
            label: p,
            value: p,
        }))
    }, [allLibraryPaths])

    const handleLibraryPathSelect = React.useCallback((value: string) => {
        if (value === CUSTOM_VALUE) return
        setDestination(!!animeFolderName ? upath.join(value, animeFolderName) : value)
    }, [animeFolderName, setDestination])

    return {
        allLibraryPaths,
        selectedLibrary,
        libraryOptions,
        handleLibraryPathSelect,
        showLibrarySelector: allLibraryPaths.length > 1,
    }
}

export type LibraryPathSelectionProps = ReturnType<typeof useLibraryPathSelection>
