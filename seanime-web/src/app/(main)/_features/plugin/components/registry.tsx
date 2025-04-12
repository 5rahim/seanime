"use client"

import {
    PluginButton,
    PluginCheckbox,
    PluginDiv,
    PluginFlex,
    PluginForm,
    PluginInput,
    PluginRadioGroup,
    PluginSelect,
    PluginStack,
    PluginSwitch,
    PluginText,
} from "@/app/(main)/_features/plugin/components/registry-components"
import type React from "react"
import { createContext, useContext } from "react"
import { ErrorBoundary } from "react-error-boundary"

// Create and initialize the registry
export const registry: ComponentRegistry = new Map([
    ["flex", PluginFlex],
    ["text", PluginText],
    ["div", PluginDiv],
    ["stack", PluginStack],
    ["input", PluginInput],
    ["button", PluginButton],
    ["form", PluginForm],
    ["switch", PluginSwitch],
    ["radio-group", PluginRadioGroup],
    ["checkbox", PluginCheckbox],
    ["select", PluginSelect],
] as any)


////////////////////////////

interface RenderPluginComponentsProps {
    data: PluginComponent | PluginComponent[]
    fallback?: (props: any) => React.ReactNode
}

// Fallback component when type is not found
function DefaultFallback({ type }: { type: string }) {
    return <div className="p-4 text-muted-foreground">Component type &quot;{type}&quot; not found</div>
}

// Error fallback
function ErrorFallbackComponent({ error }: { error: Error }) {
    return (
        <div className="p-4 text-destructive" role="alert">
            <p>Something went wrong:</p>
            <pre className="mt-2 text-sm">{error.message}</pre>
        </div>
    )
}

export function RenderPluginComponents({ data, fallback: Fallback = DefaultFallback }: RenderPluginComponentsProps) {
    const registry = usePluginRegistry()

    // Handle array of components
    if (Array.isArray(data)) {
        return (
            <>
                {data.map((component) => (
                    <RenderPluginComponents key={component.key || component.id} data={component} fallback={Fallback} />
                ))}
            </>
        )
    }

    // Get component from registry
    const Component = registry.get(data.type)

    if (!Component) {
        return <Fallback type={data.type} />
    }

    // Render component with error boundary
    return (
        <ErrorBoundary FallbackComponent={ErrorFallbackComponent}>
            <Component key={data.id} {...data.props} />
        </ErrorBoundary>
    )
}


// Type for component props
export type BaseComponentProps = Record<string, any>

// Type for component definition
export interface PluginComponent<T extends BaseComponentProps = BaseComponentProps> {
    type: string
    props: T
    id: string
    key?: string
}

// Type for nested components
export interface PluginComponentWithChildren<T extends BaseComponentProps = BaseComponentProps>
    extends PluginComponent<T> {
    props: T & {
        items?: PluginComponent[]
    }
}

// Type for component renderer function
export type ComponentRenderer<T extends BaseComponentProps = BaseComponentProps> = React.ComponentType<T>

// Registry type
export type ComponentRegistry = Map<string, ComponentRenderer>

// Create context for the registry
const PluginContext = createContext<ComponentRegistry | null>(null)

export function usePluginRegistry() {
    const context = useContext(PluginContext)
    if (!context) {
        throw new Error("usePluginRegistry must be used within a PluginProvider")
    }
    return context
}

// Provider component
export function PluginProvider({
    children,
    registry,
}: {
    children: React.ReactNode
    registry: ComponentRegistry
}) {
    return <PluginContext.Provider value={registry}>{children}</PluginContext.Provider>
}

