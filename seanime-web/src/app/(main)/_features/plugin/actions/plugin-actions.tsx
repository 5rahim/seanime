import { AL_BaseAnime, AL_BaseManga, Anime_Episode, Onlinestream_Episode } from "@/api/generated/types"
import { Button, ButtonProps, IconButton } from "@/components/ui/button"
import { ContextMenuItem, ContextMenuSeparator } from "@/components/ui/context-menu"
import { DropdownMenu, DropdownMenuItem, DropdownMenuSeparator } from "@/components/ui/dropdown-menu"
import React, { useEffect, useState } from "react"
import { BiDotsHorizontal } from "react-icons/bi"
import {
    usePluginListenActionRenderAnimeLibraryDropdownItemsEvent,
    usePluginListenActionRenderAnimePageButtonsEvent,
    usePluginListenActionRenderAnimePageDropdownItemsEvent,
    usePluginListenActionRenderEpisodeCardContextMenuItemsEvent,
    usePluginListenActionRenderEpisodeGridItemMenuItemsEvent,
    usePluginListenActionRenderMangaPageButtonsEvent,
    usePluginListenActionRenderMediaCardContextMenuItemsEvent,
    usePluginSendActionClickedEvent,
    usePluginSendActionRenderAnimeLibraryDropdownItemsEvent,
    usePluginSendActionRenderAnimePageButtonsEvent,
    usePluginSendActionRenderAnimePageDropdownItemsEvent,
    usePluginSendActionRenderEpisodeCardContextMenuItemsEvent,
    usePluginSendActionRenderEpisodeGridItemMenuItemsEvent,
    usePluginSendActionRenderMangaPageButtonsEvent,
    usePluginSendActionRenderMediaCardContextMenuItemsEvent,
} from "../generated/plugin-events"

function sortItems<T extends { label: string }>(items: T[]) {
    return items.sort((a, b) => a.label.localeCompare(b.label, undefined, { numeric: true }))
}

type PluginAnimePageButton = {
    extensionId: string
    intent: string
    onClick: string
    label: string
    style: React.CSSProperties
    id: string
}

export function PluginAnimePageButtons(props: { media: AL_BaseAnime }) {
    const [buttons, setButtons] = useState<PluginAnimePageButton[]>([])

    const { sendActionRenderAnimePageButtonsEvent } = usePluginSendActionRenderAnimePageButtonsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderAnimePageButtonsEvent({}, "")
    }, [])

    // Listen for the action to render the anime page buttons
    usePluginListenActionRenderAnimePageButtonsEvent((event, extensionId) => {
        setButtons(p => {
            const otherButtons = p.filter(b => b.extensionId !== extensionId)
            const extButtons = event.buttons.map((b: Record<string, any>) => ({ ...b, extensionId } as PluginAnimePageButton))
            return sortItems([...otherButtons, ...extButtons])
        })
    }, "")

    // Send
    function handleClick(button: PluginAnimePageButton) {
        sendActionClickedEvent({
            actionId: button.id,
            event: {
                media: props.media,
            },
        }, button.extensionId)
    }

    if (buttons.length === 0) return null

    return <>
        {buttons.map(b => (
            <Button
                key={b.id}
                intent={b.intent as ButtonProps["intent"] || "white-subtle"}
                onClick={() => handleClick(b)}
                style={b.style}
            >{b.label || "???"}</Button>
        ))}
    </>
}


/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginMangaPageButton = {
    extensionId: string
    intent: string
    onClick: string
    label: string
    style: React.CSSProperties
    id: string
}

export function PluginMangaPageButtons(props: { media: AL_BaseManga }) {
    const [buttons, setButtons] = useState<PluginMangaPageButton[]>([])

    const { sendActionRenderMangaPageButtonsEvent } = usePluginSendActionRenderMangaPageButtonsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderMangaPageButtonsEvent({}, "")
    }, [])

    // Listen for the action to render the manga page buttons
    usePluginListenActionRenderMangaPageButtonsEvent((event, extensionId) => {
        setButtons(p => {
            const otherButtons = p.filter(b => b.extensionId !== extensionId)
            const extButtons = event.buttons.map((b: Record<string, any>) => ({ ...b, extensionId } as PluginMangaPageButton))
            return sortItems([...otherButtons, ...extButtons])
        })
    }, "")

    // Send
    function handleClick(button: PluginMangaPageButton) {
        sendActionClickedEvent({
            actionId: button.id,
            event: {
                media: props.media,
            },
        }, button.extensionId)
    }

    if (buttons.length === 0) return null

    return <>
        {buttons.map(b => (
            <Button
                key={b.id}
                intent={b.intent as ButtonProps["intent"] || "white-subtle"}
                onClick={() => handleClick(b)}
                style={b.style}
            >{b.label || "???"}</Button>
        ))}
    </>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginMediaCardContextMenuItem = {
    extensionId: string
    onClick: string
    label: string
    style: React.CSSProperties
    id: string
    for: "anime" | "manga" | "both"
}

type PluginMediaCardContextMenuItemsProps = {
    for: "anime" | "manga",
    media: AL_BaseAnime | AL_BaseManga
}

export function PluginMediaCardContextMenuItems(props: PluginMediaCardContextMenuItemsProps) {
    const [items, setItems] = useState<PluginMediaCardContextMenuItem[]>([])

    const { sendActionRenderMediaCardContextMenuItemsEvent } = usePluginSendActionRenderMediaCardContextMenuItemsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderMediaCardContextMenuItemsEvent({}, "")
    }, [])

    // Listen for the action to render the media card context menu items
    usePluginListenActionRenderMediaCardContextMenuItemsEvent((event, extensionId) => {
        setItems(p => {
            const otherItems = p.filter(b => b.extensionId !== extensionId)
            const extItems = event.items
                .filter((i: PluginMediaCardContextMenuItem) => i.for === props.for || i.for === "both")
                .map((b: Record<string, any>) => ({ ...b, extensionId } as PluginMangaPageButton))
            return sortItems([...otherItems, ...extItems])
        })
    }, "")

    // Send
    function handleClick(item: PluginMediaCardContextMenuItem) {
        sendActionClickedEvent({
            actionId: item.id,
            event: {
                media: props.media,
            },
        }, item.extensionId)
    }

    if (items.length === 0) return null

    return <>
        <ContextMenuSeparator className="my-2" />
        {items.map(i => (
            <ContextMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</ContextMenuItem>
        ))}
    </>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginAnimeLibraryDropdownMenuItem = {
    extensionId: string
    onClick: string
    label: string
    id: string
    style: React.CSSProperties
}

export function PluginAnimeLibraryDropdownItems() {
    const [items, setItems] = useState<PluginAnimeLibraryDropdownMenuItem[]>([])

    const { sendActionRenderAnimeLibraryDropdownItemsEvent } = usePluginSendActionRenderAnimeLibraryDropdownItemsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderAnimeLibraryDropdownItemsEvent({}, "")
    }, [])


    // Listen for the action to render the anime library dropdown items
    usePluginListenActionRenderAnimeLibraryDropdownItemsEvent((event, extensionId) => {
        setItems(p => {
            const otherItems = p.filter(i => i.extensionId !== extensionId)
            const extItems = event.items.map((i: Record<string, any>) => ({ ...i, extensionId } as PluginAnimeLibraryDropdownMenuItem))
            return sortItems([...otherItems, ...extItems])
        })
    }, "")

    // Send
    function handleClick(item: PluginAnimeLibraryDropdownMenuItem) {
        sendActionClickedEvent({
            actionId: item.id,
            event: {},
        }, item.extensionId)
    }

    if (items.length === 0) return null

    return <>
        <DropdownMenuSeparator />
        {items.map(i => (
            <DropdownMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</DropdownMenuItem>
        ))}
    </>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginEpisodeCardContextMenuItem = {
    extensionId: string
    onClick: string
    label: string
    id: string
    style: React.CSSProperties
}

export function PluginEpisodeCardContextMenuItems(props: { episode: Anime_Episode | undefined }) {
    const [items, setItems] = useState<PluginEpisodeCardContextMenuItem[]>([])

    const { sendActionRenderEpisodeCardContextMenuItemsEvent } = usePluginSendActionRenderEpisodeCardContextMenuItemsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderEpisodeCardContextMenuItemsEvent({}, "")
    }, [])

    // Listen for the action to render the episode card context menu items
    usePluginListenActionRenderEpisodeCardContextMenuItemsEvent((event, extensionId) => {
        setItems(p => {
            const otherItems = p.filter(i => i.extensionId !== extensionId)
            const extItems = event.items.map((i: Record<string, any>) => ({ ...i, extensionId } as PluginEpisodeCardContextMenuItem))
            return sortItems([...otherItems, ...extItems])
        })
    }, "")

    // Send
    function handleClick(item: PluginEpisodeCardContextMenuItem) {
        sendActionClickedEvent({
            actionId: item.id,
            event: {
                episode: props.episode,
            },
        }, item.extensionId)
    }

    if (items.length === 0) return null

    return <>
        <ContextMenuSeparator />
        {items.map(i => (
            <ContextMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</ContextMenuItem>
        ))}
    </>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginEpisodeGridItemMenuItem = {
    extensionId: string
    onClick: string
    label: string
    id: string
    type: "library" | "torrentstream" | "debridstream" | "onlinestream" | "undownloaded" | "medialinks" | "mediastream"
    style: React.CSSProperties
}

export function PluginEpisodeGridItemMenuItems(props: {
    isDropdownMenu: boolean,
    type: PluginEpisodeGridItemMenuItem["type"],
    episode: Anime_Episode | Onlinestream_Episode | undefined
}) {
    const [items, setItems] = useState<PluginEpisodeGridItemMenuItem[]>([])

    const { sendActionRenderEpisodeGridItemMenuItemsEvent } = usePluginSendActionRenderEpisodeGridItemMenuItemsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderEpisodeGridItemMenuItemsEvent({}, "")
    }, [])

    // Listen for the action to render the episode grid item context menu items
    usePluginListenActionRenderEpisodeGridItemMenuItemsEvent((event, extensionId) => {
        setItems(p => {
            const otherItems = p.filter(i => i.extensionId !== extensionId && i.type === props.type)
            const extItems = event.items.filter((i: PluginEpisodeGridItemMenuItem) => i.type === props.type)
                .map((i: Record<string, any>) => ({ ...i, extensionId } as PluginEpisodeGridItemMenuItem))
            return sortItems([...otherItems, ...extItems])
        })
    }, "")

    // Send
    function handleClick(item: PluginEpisodeGridItemMenuItem) {
        sendActionClickedEvent({
            actionId: item.id,
            event: {
                episode: props.episode,
            },
        }, item.extensionId)
    }

    if (items.length === 0) return null

    if (props.isDropdownMenu) {
        return <DropdownMenu
            trigger={
                <IconButton
                    icon={<BiDotsHorizontal />}
                    intent="gray-basic"
                    size="xs"
                />
            }
        >
            {items.map(i => (
                <DropdownMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</DropdownMenuItem>
            ))}
        </DropdownMenu>
    }

    return <>
        <DropdownMenuSeparator />
        {items.map(i => (
            <DropdownMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</DropdownMenuItem>
        ))}
    </>
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type PluginAnimePageDropdownMenuItem = {
    extensionId: string
    onClick: string
    label: string
    id: string
    style: React.CSSProperties
}

export function PluginAnimePageDropdownItems(props: { media: AL_BaseAnime }) {
    const [items, setItems] = useState<PluginAnimePageDropdownMenuItem[]>([])

    const { sendActionRenderAnimePageDropdownItemsEvent } = usePluginSendActionRenderAnimePageDropdownItemsEvent()
    const { sendActionClickedEvent } = usePluginSendActionClickedEvent()

    useEffect(() => {
        sendActionRenderAnimePageDropdownItemsEvent({}, "")
    }, [])

    // Listen for the action to render the anime page dropdown items
    usePluginListenActionRenderAnimePageDropdownItemsEvent((event, extensionId) => {
        setItems(p => {
            const otherItems = p.filter(i => i.extensionId !== extensionId)
            const extItems = event.items.map((i: Record<string, any>) => ({ ...i, extensionId } as PluginAnimePageDropdownMenuItem))
            return sortItems([...otherItems, ...extItems])
        })
    }, "")

    // Send
    function handleClick(item: PluginAnimePageDropdownMenuItem) {
        sendActionClickedEvent({
            actionId: item.id,
            event: {
                media: props.media,
            },
        }, item.extensionId)
    }

    if (items.length === 0) return null

    return <>
        <DropdownMenuSeparator />
        {items.map(i => (
            <DropdownMenuItem key={i.id} onClick={() => handleClick(i)} style={i.style}>{i.label || "???"}</DropdownMenuItem>
        ))}
    </>

}
