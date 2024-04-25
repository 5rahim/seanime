import { cn } from "@/components/ui/core/styling"
import React from "react"

type EpisodeListGridProps = {
    children?: React.ReactNode
}

type ContainerSize = "half" | "expanded"

const __GridSizeContext = React.createContext<{ container: ContainerSize }>({ container: "expanded" })

export function EpisodeListGrid(props: EpisodeListGridProps) {

    const {
        children,
        ...rest
    } = props

    const { container } = React.useContext(__GridSizeContext)

    return (
        <div
            className={cn(
                "grid gap-4",
                { "grid grid-cols-1 md:grid-cols-2": container === "half" },
                { "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 min-[2000px]:grid-cols-4": container === "expanded" },
            )}
        >
            {children}
        </div>
    )
}

type EpisodeListGridProviderProps = {
    children?: React.ReactNode
    container: ContainerSize
}

export function EpisodeListGridProvider(props: EpisodeListGridProviderProps) {

    const {
        children,
        container,
    } = props

    return (
        <__GridSizeContext.Provider
            value={{
                container,
            }}
        >
            {children}
        </__GridSizeContext.Provider>
    )
}

