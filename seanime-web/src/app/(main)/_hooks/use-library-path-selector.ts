import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { upath } from "@/lib/helpers/upath"
import React from "react"

const CUSTOM_VALUE = "__custom__"

type UseLibraryPathSelectorOptions = {
    destination: string
    setDestination: (path: string) => void
    animeFolderName?: string
}

export function useLibraryPathSelector(options: UseLibraryPathSelectorOptions) {
    const { destination, setDestination, animeFolderName } = options

    const serverStatus = useServerStatus()
    const libraryPath = serverStatus?.settings?.library?.libraryPath
    const additionalLibraryPaths = serverStatus?.settings?.library?.libraryPaths

    const allLibraryPaths = React.useMemo(() => {
        const paths: string[] = []
        if (libraryPath) paths.push(libraryPath)
        if (additionalLibraryPaths?.length) {
            paths.push(...additionalLibraryPaths.filter(p => p && p !== libraryPath))
        }
        return paths
    }, [libraryPath, additionalLibraryPaths])

    const selectedLibrary = React.useMemo(() => {
        const sortedPaths = [...allLibraryPaths].sort((a, b) => b.length - a.length)
        const matchingPath = sortedPaths.find(p => destination.startsWith(p))
        return matchingPath ?? CUSTOM_VALUE
    }, [allLibraryPaths, destination])

    const libraryOptions = React.useMemo(() => {
        const options = allLibraryPaths.map(p => ({
            label: p,
            value: p,
        }))
        options.push({
            label: "Custom",
            value: CUSTOM_VALUE,
        })
        return options
    }, [allLibraryPaths])

    const handleLibraryPathSelect = React.useCallback((value: string) => {
        if (value === CUSTOM_VALUE) return

        if (animeFolderName) {
            setDestination(upath.join(value, animeFolderName))
        } else {
            setDestination(value)
        }
    }, [animeFolderName, setDestination])

    return {
        allLibraryPaths,
        libraryPath,
        selectedLibrary,
        libraryOptions,
        handleLibraryPathSelect,
        showLibrarySelector: allLibraryPaths.length > 1,
    }
}
