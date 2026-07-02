import { Anime_Entry, Anime_Episode, Anime_EpisodeCollection } from "@/api/generated/types"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { TorrentStreamEpisodeSection } from "@/app/(main)/entry/_containers/torrent-stream/_components/torrent-stream-episode-section"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { WSEvents } from "@/lib/server/ws-events"
import { atom } from "jotai"
import { useAtomValue } from "jotai/react"
import { useAtom } from "jotai/react"
import React, { startTransition } from "react"
import { BiExtension } from "react-icons/bi"
import {
    usePluginListenAnimeEntryEpisodeTabEpisodeCollectionEvent,
    usePluginListenAnimeEntryEpisodeTabsUpdatedEvent,
    usePluginSendAnimeEntryEpisodeTabOpenEvent,
    usePluginSendAnimeEntryEpisodeTabSelectEpisodeEvent,
    usePluginSendAnimeEntryEpisodeTabsRenderEvent,
    usePluginSendAnimeEntryEpisodeTabStateChangedEvent,
} from "./generated/plugin-events"

function sortTabs(tabs: PluginAnimeEntryEpisodeTab[]) {
    return tabs.sort((a, b) => {
        const extCmp = a.extensionId.localeCompare(b.extensionId, undefined, { numeric: true })
        if (extCmp !== 0) return extCmp
        return a.name.localeCompare(b.name, undefined, { numeric: true })
    })
}

export function getPluginEpisodeTabViewId(extensionId: string) {
    return `episodeTab:${extensionId}`
}

export function getPluginEpisodeTabExtensionId(viewId: string) {
    return viewId.startsWith("episodeTab:") ? viewId.slice("episodeTab:".length) : ""
}

const __plugin_episodeTabsAtom = atom<PluginAnimeEntryEpisodeTab[]>([])
const __plugin_episodeTabCollectionsAtom = atom<Record<string, Anime_EpisodeCollection | undefined>>({})
const __plugin_episodeTabRenderedExtensionIdsAtom = atom<string[]>([])

export type PluginAnimeEntryEpisodeTab = {
    extensionId: string
    name: string
    icon?: string
    viewId: string
}

export function usePluginAnimeEntryEpisodeTabsListener(props: {
    mediaId: number
    currentView: string
    setView: (view: string) => void
}) {
    const { mediaId, currentView, setView } = props

    const [tabs, setTabs] = useAtom(__plugin_episodeTabsAtom)
    const [, setCollections] = useAtom(__plugin_episodeTabCollectionsAtom)
    const [renderedExtensionIds, setRenderedExtensionIds] = useAtom(__plugin_episodeTabRenderedExtensionIdsAtom)

    const { sendAnimeEntryEpisodeTabsRenderEvent } = usePluginSendAnimeEntryEpisodeTabsRenderEvent()
    const { sendAnimeEntryEpisodeTabOpenEvent } = usePluginSendAnimeEntryEpisodeTabOpenEvent()
    const { sendAnimeEntryEpisodeTabStateChangedEvent } = usePluginSendAnimeEntryEpisodeTabStateChangedEvent()

    const renderTabs = React.useEffectEvent(() => {
        if (!mediaId) return
        setRenderedExtensionIds([])
        sendAnimeEntryEpisodeTabsRenderEvent({ mediaId }, "")
    })

    React.useEffect(() => {
        setCollections({})
        renderTabs()
    }, [mediaId])

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_LOADED,
        onMessage: () => {
            renderTabs()
        },
    })

    useWebsocketMessageListener({
        type: WSEvents.PLUGIN_UNLOADED,
        onMessage: (extensionId: string) => {
            startTransition(() => {
                setTabs(prev => prev.filter(tab => tab.extensionId !== extensionId))
                setRenderedExtensionIds(prev => prev.filter(id => id !== extensionId))
                setCollections(prev => {
                    const next = { ...prev }
                    Object.keys(next).forEach(key => {
                        if (key === getPluginEpisodeTabViewId(extensionId)) {
                            delete next[key]
                        }
                    })
                    return next
                })
                if (currentView === getPluginEpisodeTabViewId(extensionId)) {
                    setView("library")
                }
            })
        },
    })

    usePluginListenAnimeEntryEpisodeTabsUpdatedEvent((event, extensionId) => {
        startTransition(() => {
            setRenderedExtensionIds(prev => prev.includes(extensionId) ? prev : [...prev, extensionId])
            setTabs(prev => {
                const otherTabs = prev.filter(tab => tab.extensionId !== extensionId)
                const extensionTabs = (event.tabs ?? []).map((tab: Record<string, any>) => ({
                    ...tab,
                    extensionId,
                    viewId: getPluginEpisodeTabViewId(extensionId),
                } as PluginAnimeEntryEpisodeTab))
                return sortTabs([...otherTabs, ...extensionTabs])
            })
        })
    }, "")

    usePluginListenAnimeEntryEpisodeTabEpisodeCollectionEvent((event, extensionId) => {
        const viewId = getPluginEpisodeTabViewId(extensionId)
        startTransition(() => {
            setCollections(prev => ({
                ...prev,
                [viewId]: event.episodeCollection as Anime_EpisodeCollection,
            }))
        })
    }, "")

    const selectedTab = tabs.find(tab => tab.viewId === currentView)

    React.useEffect(() => {
        tabs.forEach(tab => {
            sendAnimeEntryEpisodeTabStateChangedEvent({
                isOpen: tab.viewId === currentView,
            }, tab.extensionId)
        })
    }, [currentView, tabs])

    React.useEffect(() => {
        if (!selectedTab || !mediaId) return
        sendAnimeEntryEpisodeTabOpenEvent({
            mediaId,
        }, selectedTab.extensionId)
    }, [mediaId, selectedTab, sendAnimeEntryEpisodeTabOpenEvent])

    return {
        tabs,
        renderedExtensionIds,
    }
}

export function usePluginAnimeEntryEpisodeTabs(props: {
    mediaId: number
    currentView: string
    setView: (view: string) => void
}) {
    const { mediaId, currentView, setView } = props

    const tabs = useAtomValue(__plugin_episodeTabsAtom)
    const collections = useAtomValue(__plugin_episodeTabCollectionsAtom)
    const renderedExtensionIds = useAtomValue(__plugin_episodeTabRenderedExtensionIdsAtom)

    const selectedTab = tabs.find(tab => tab.viewId === currentView)

    const { sendAnimeEntryEpisodeTabSelectEpisodeEvent } = usePluginSendAnimeEntryEpisodeTabSelectEpisodeEvent()

    const selectEpisode = React.useCallback((episode: Anime_Episode) => {
        if (!selectedTab || !mediaId) return

        sendAnimeEntryEpisodeTabSelectEpisodeEvent({
            mediaId,
            episodeNumber: episode.episodeNumber,
            aniDbEpisode: episode.aniDBEpisode ?? "",
            episode,
        }, selectedTab.extensionId)
    }, [mediaId, selectedTab])

    return {
        tabs,
        selectedTab,
        selectedEpisodeCollection: selectedTab ? collections[selectedTab.viewId] : undefined,
        renderedExtensionIds,
        selectEpisode,
    }
}

export function PluginAnimeEntryEpisodeTabContent(props: {
    entry: Anime_Entry
    tab: PluginAnimeEntryEpisodeTab
    episodeCollection: Anime_EpisodeCollection | undefined
    bottomSection?: React.ReactNode
    onSelectEpisode: (episode: Anime_Episode) => void
}) {
    const { entry, tab, episodeCollection, bottomSection, onSelectEpisode } = props

    if (!episodeCollection) {
        return <LoadingSpinner />
    }

    return <>
        <div className="h-10" />

        {/* {episodeCollection.hasMappingError && (
         <div data-plugin-anime-entry-episode-tab-no-metadata-message-container>
         <p className="text-[--red] opacity-50">
         No metadata info available for this anime. Episode mapping may be incomplete.
         </p>
         </div>
         )} */}

        <TorrentStreamEpisodeSection
            contextType={`episodeTab:${tab.extensionId}`}
            episodeCollection={episodeCollection}
            entry={entry}
            onEpisodeClick={onSelectEpisode}
            onPlayNextEpisodeOnMount={onSelectEpisode}
            bottomSection={bottomSection}
        />
    </>
}

export function PluginAnimeEntryTabIcon(props: { icon?: string, className?: string }) {
    const { icon, className, ...rest } = props

    if (!icon) {
        return <BiExtension className={className} aria-hidden="true" {...rest} />
    }

    if (icon.startsWith("http://") || icon.startsWith("https://") || icon.startsWith("data:image/") || icon.startsWith("/")) {
        return <img
            src={icon}
            alt=""
            className={cn("inline-block size-4 rounded-sm object-contain", className)}
            aria-hidden="true"
            {...rest}
        />
    }

    return <span
        {...props}
        className={cn("", className)}
        dangerouslySetInnerHTML={{ __html: icon }}
    />
}
