import { CommandDialog, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import mousetrap from "mousetrap"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { PluginProvider, registry, RenderPluginComponents } from "../components/registry"
import {
    usePluginListenCommandPaletteCloseEvent,
    usePluginListenCommandPaletteGetInputEvent,
    usePluginListenCommandPaletteOpenEvent,
    usePluginListenCommandPaletteSetInputEvent,
    usePluginListenCommandPaletteUpdatedEvent,
    usePluginSendCommandPaletteClosedEvent,
    usePluginSendCommandPaletteInputEvent,
    usePluginSendCommandPaletteItemSelectedEvent,
    usePluginSendCommandPaletteOpenedEvent,
    usePluginSendRenderCommandPaletteEvent,
} from "../generated/plugin-events"

export type PluginCommandPaletteInfo = {
    extensionId: string
    placeholder: string
    keyboardShortcut: string
}

type CommandItem = {
    id: string
    value: string
    filterType: string
    heading: string

    // Either the label or the components should be set
    label: string // empty string if components are set
    components?: any
}

export function PluginCommandPalette(props: { extensionId: string, info: PluginCommandPaletteInfo }) {

    const { extensionId, info } = props

    const router = useRouter()
    const pathname = usePathname()

    const [open, setOpen] = React.useState(false)
    const [input, setInput] = React.useState("")
    const [activeItemId, setActiveItemId] = React.useState("")
    const [items, setItems] = React.useState<CommandItem[]>([])
    const [placeholder, setPlaceholder] = React.useState(info.placeholder)

    const { sendRenderCommandPaletteEvent } = usePluginSendRenderCommandPaletteEvent()
    const { sendCommandPaletteInputEvent } = usePluginSendCommandPaletteInputEvent()
    const { sendCommandPaletteOpenedEvent } = usePluginSendCommandPaletteOpenedEvent()
    const { sendCommandPaletteClosedEvent } = usePluginSendCommandPaletteClosedEvent()
    const { sendCommandPaletteItemSelectedEvent } = usePluginSendCommandPaletteItemSelectedEvent()

    // const parsedCommandProps = useSeaCommand_ParseCommand(input)

    // Register the keyboard shortcut
    React.useEffect(() => {
        if (!!info.keyboardShortcut) {
            mousetrap.bind(info.keyboardShortcut, () => {
                setInput("")
                React.startTransition(() => {
                    setOpen(true)
                })
            })

            return () => {
                mousetrap.unbind(info.keyboardShortcut)
            }
        }
    }, [info.keyboardShortcut])

    // Render the command palette
    useUpdateEffect(() => {
        if (!open) {
            setInput("")
            sendCommandPaletteClosedEvent({}, extensionId)
        }

        if (open) {
            sendCommandPaletteOpenedEvent({}, extensionId)
            sendRenderCommandPaletteEvent({}, extensionId)
        }
    }, [open, extensionId])

    // Send the input when the server requests it
    usePluginListenCommandPaletteGetInputEvent((data) => {
        sendCommandPaletteInputEvent({ value: input }, extensionId)
    }, extensionId)

    // Set the input when the server sends it
    usePluginListenCommandPaletteSetInputEvent((data) => {
        setInput(data.value)
    }, extensionId)

    // Open the command palette when the server requests it
    usePluginListenCommandPaletteOpenEvent((data) => {
        setOpen(true)
    }, extensionId)

    // Close the command palette when the server requests it
    usePluginListenCommandPaletteCloseEvent((data) => {
        setOpen(false)
    }, extensionId)

    // Continuously listen to render the command palette
    usePluginListenCommandPaletteUpdatedEvent((data) => {
        setItems(data.items)
        setPlaceholder(data.placeholder)
    }, extensionId)

    const commandListRef = React.useRef<HTMLDivElement>(null)

    function scrollToTop() {
        const list = commandListRef.current
        if (!list) return () => { }

        const t = setTimeout(() => {
            list.scrollTop = 0
            // Find and focus the first command item
            const firstItem = list.querySelector("[cmdk-item]") as HTMLElement
            if (firstItem) {
                const value = firstItem.getAttribute("data-value")
                if (value) {
                    setActiveItemId(value)
                }
            }
        }, 100)

        return () => clearTimeout(t)
    }

    React.useEffect(() => {
        const cl = scrollToTop()
        return () => cl()
    }, [input, pathname])

    // Group items by heading and sort by priority
    const groupedItems = React.useMemo(() => {
        const groups: Record<string, CommandItem[]> = {}

        const _items = items.filter(item =>
            item.filterType === "includes" ?
                item.value.toLowerCase().includes(input.toLowerCase()) :
                item.filterType === "startsWith" ?
                    item.value.toLowerCase().startsWith(input.toLowerCase()) :
                    true)

        _items.forEach(item => {
            const heading = item.heading || ""
            if (!groups[heading]) groups[heading] = []
            groups[heading].push(item)
        })

        // Scroll to top when items are rendered
        scrollToTop()

        return groups
    }, [items, input])

    function handleSelect(item: CommandItem) {
        // setInput("")
        sendCommandPaletteItemSelectedEvent({ itemId: item.id }, extensionId)
    }


    return (
        <CommandDialog
            open={open}
            onOpenChange={setOpen}
            commandProps={{
                value: activeItemId,
                onValueChange: setActiveItemId,
            }}
            overlayClass="bg-black/30"
            contentClass="max-w-2xl"
            commandClass="h-[300px]"
        >
            <CommandInput
                placeholder={placeholder || ""}
                value={input}
                onValueChange={setInput}
            />
            <CommandList className="mb-2" ref={commandListRef}>

                <PluginProvider registry={registry}>
                    {Object.entries(groupedItems).map(([heading, items]) => (
                        <CommandGroup key={heading} heading={heading}>
                            {items.map(item => (
                                <CommandItem
                                    key={item.id}
                                    value={item.value}
                                    onSelect={() => {
                                        handleSelect(item)
                                    }}
                                    className="block"
                                >
                                    {!!item.label ? item.label : <RenderPluginComponents data={item.components} />}
                                </CommandItem>
                            ))}
                        </CommandGroup>
                    ))}
                </PluginProvider>

            </CommandList>
        </CommandDialog>
    )
}
