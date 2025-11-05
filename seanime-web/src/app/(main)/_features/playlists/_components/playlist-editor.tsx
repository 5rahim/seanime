import { Anime_LibraryCollection, Anime_LibraryCollectionEntry, Anime_PlaylistEpisode, Anime_WatchType, Nullish } from "@/api/generated/types"
import { useGetPlaylistEpisodes } from "@/api/hooks/playlist.hooks"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-editor-manager"
import { useHasDebridService, useHasOnlineStreaming, useHasTorrentStreaming } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
import { SeaImage } from "@/components/shared/sea-image"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { getImageUrl } from "@/lib/server/assets"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { arrayMove, SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import React from "react"
import { BiPlus, BiTrash } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { toast } from "sonner"

export function playlist_isSameEpisode(a: Nullish<Anime_PlaylistEpisode>, b: Nullish<Anime_PlaylistEpisode>) {
    if (!a || !b) return false
    return a.episode?.aniDBEpisode === b.episode?.aniDBEpisode && a.episode?.baseAnime?.id === b.episode?.baseAnime?.id
}

export function playlist_getEpisodeKey(a: Anime_PlaylistEpisode) {
    return a.episode?.aniDBEpisode + String(a.episode?.baseAnime?.id) + String(a.episode?.episodeNumber) + String(a.episode?.progressNumber) + String(
        a.episode?.localFile?.path)
}

type PlaylistEditorProps = {
    episodes: Anime_PlaylistEpisode[]
    setEpisodes: (episodes: Anime_PlaylistEpisode[]) => void
    libraryCollection?: Anime_LibraryCollection
}

export function PlaylistEditor(props: PlaylistEditorProps) {

    const {
        episodes: controlledEpisodes,
        setEpisodes: onChange,
        libraryCollection,
        ...rest
    } = props

    /**
     * Fetch the anime library collection
     */

    const firstRender = React.useRef(true)

    const [episodes, setEpisodes] = React.useState(controlledEpisodes)

    const { selectedMedia, setSelectedMedia } = usePlaylistEditorManager()

    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        setEpisodes(controlledEpisodes)
    }, [controlledEpisodes])

    React.useEffect(() => {
        onChange(episodes)
    }, [episodes])

    const handleDragEnd = React.useCallback((event: DragEndEvent) => {
        const { active, over } = event

        if (active.id !== over?.id) {
            setEpisodes((items) => {
                const oldIndex = items.findIndex(item => playlist_getEpisodeKey(item) === active.id)
                const newIndex = items.findIndex(item => playlist_getEpisodeKey(item) === over?.id)

                return arrayMove(items, oldIndex, newIndex)
            })
        }
    }, [])

    const [selectedCategory, setSelectedCategory] = React.useState("CURRENT")
    const [searchInput, setSearchInput] = React.useState("")
    const debouncedSearchInput = useDebounce(searchInput, 500)

    const entries = React.useMemo(() => {
        if (debouncedSearchInput.length !== 0) return (libraryCollection?.lists
            ?.filter(n => n.type === "PLANNING" || n.type === "PAUSED" || n.type === "CURRENT")
            ?.flatMap(n => n.entries)
            ?.filter(Boolean) ?? []).filter(n => n?.media?.title?.english?.toLowerCase()?.includes(debouncedSearchInput.toLowerCase()) ||
            n?.media?.title?.romaji?.toLowerCase()?.includes(debouncedSearchInput.toLowerCase()))

        return libraryCollection?.lists?.filter(n => {
                if (selectedCategory === "-") return n.type === "PLANNING" || n.type === "PAUSED" || n.type === "CURRENT"
                return n.type === selectedCategory
            })
            ?.flatMap(n => n.entries)
            ?.filter(Boolean) ?? []
    }, [libraryCollection, debouncedSearchInput, selectedCategory])


    return (
        <div className="space-y-4">

            <div className="space-y-2">
                <Modal
                    title="Select an anime"
                    contentClass="max-w-4xl"
                    trigger={<Button
                        leftIcon={<BiPlus className="text-2xl" />}
                        intent="white-glass"
                        className="rounded-full"
                        disabled={episodes.length >= 10}
                    >Add episodes</Button>}
                >

                    <div className="grid grid-cols-[150px,1fr] gap-2">
                        <Select
                            value={selectedCategory}
                            onValueChange={v => setSelectedCategory(v)}
                            options={[
                                { label: "Current", value: "CURRENT" },
                                { label: "Paused", value: "PAUSED" },
                                { label: "Planning", value: "PLANNING" },
                                { label: "All", value: "-" },
                            ]}
                            disabled={searchInput.length !== 0}
                        />

                        <TextInput
                            placeholder="Search"
                            value={searchInput}
                            onChange={e => setSearchInput(e.target.value)}
                        />

                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-5 gap-2">
                        {entries?.map(entry => {
                            return (
                                <PlaylistMediaEntryTrigger key={entry.mediaId} entry={entry} episodes={episodes} setEpisodes={setEpisodes} />
                            )
                        })}
                    </div>
                </Modal>
            </div>


            <DndContext
                modifiers={[restrictToVerticalAxis]}
                onDragEnd={handleDragEnd}
            >
                <SortableContext
                    strategy={verticalListSortingStrategy}
                    items={episodes.map(n => playlist_getEpisodeKey(n))}
                >
                    <div className="space-y-2">
                        <ul className="space-y-2">
                            {episodes.filter(Boolean).map((ep, index) => (
                                <SortableItem
                                    key={playlist_getEpisodeKey(ep)}
                                    id={playlist_getEpisodeKey(ep)}
                                    episode={ep}
                                    setEpisodes={setEpisodes}
                                />
                            ))}
                        </ul>
                    </div>
                </SortableContext>
            </DndContext>
        </div>
    )
}

type PlaylistMediaEntryTriggerProps = {
    entry: Anime_LibraryCollectionEntry
    episodes: Anime_PlaylistEpisode[]
    setEpisodes: React.Dispatch<React.SetStateAction<Anime_PlaylistEpisode[]>>
}

function PlaylistMediaEntryTrigger(props: PlaylistMediaEntryTriggerProps) {
    const { entry, episodes, setEpisodes } = props

    const { selectedMedia, setSelectedMedia } = usePlaylistEditorManager()

    const added = episodes.filter(n => n.episode?.baseAnime?.id === entry.mediaId)?.length ?? 0

    return (
        <div
            key={entry.mediaId}
            className="col-span-1 aspect-[7/7] rounded-md border overflow-hidden relative transition cursor-pointer bg-[--background] md:opacity-60 md:hover:opacity-100"
            onClick={() => setSelectedMedia(entry.mediaId)}
        >
            {entry.libraryData && <div data-media-entry-card-body-library-badge className="absolute z-[1] left-0 top-0">
                <Badge
                    size="lg" intent="warning-solid"
                    className="rounded-md rounded-bl-none rounded-tr-none text-orange-900 opacity-80"
                ><IoLibrarySharp /></Badge>
            </div>}
            {added > 0 && <div data-media-entry-card-body-library-badge className="absolute z-[1] right-1 top-1">
                <Badge
                    size="lg" intent="warning-solid"
                    className="rounded-full bg-black text-white opacity-80 size-6"
                >{added}</Badge>
            </div>}

            <SeaImage
                src={entry.media?.coverImage?.large || entry.media?.bannerImage || ""}
                placeholder={imageShimmer(700, 475)}
                sizes="10rem"
                fill
                alt=""
                className="object-center object-cover"
            />
            <p className="line-clamp-2 text-sm absolute m-2 bottom-0 font-semibold z-[10]">
                {entry.media?.title?.userPreferred || entry.media?.title?.romaji}
            </p>
            <div
                className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
            />


        </div>
    )
}

type PlaylistMediaEntryProps = {
    entry: Anime_LibraryCollectionEntry
    episodes: Anime_PlaylistEpisode[]
    setEpisodes: React.Dispatch<React.SetStateAction<Anime_PlaylistEpisode[]>>
}

export function PlaylistMediaEntry(props: PlaylistMediaEntryProps) {
    const { entry, episodes, setEpisodes } = props

    const { selectedMedia, setSelectedMedia } = usePlaylistEditorManager()

    return <Modal
        open={selectedMedia === entry.mediaId}
        onOpenChange={v => {
            if (!v) {
                setSelectedMedia(null)
            }
        }}
        title={entry.media?.title?.userPreferred || entry.media?.title?.romaji || ""}
    >
        <EntryEpisodeList
            selectedEpisodes={episodes}
            setSelectedEpisodes={setEpisodes}
            entry={entry}
        />
    </Modal>
}

function SortableItem({ id, episode, setEpisodes }: {
    id: string,
    episode: Anime_PlaylistEpisode
    setEpisodes: React.Dispatch<React.SetStateAction<Anime_PlaylistEpisode[]>>
}) {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
    } = useSortable({ id: id })

    // const { hasTorrentStreaming, hasDebridStreaming } = useHasTorrentOrDebridInclusion()
    const { hasDebridService } = useHasDebridService()
    const { hasTorrentStreaming } = useHasTorrentStreaming()
    const { hasOnlineStreaming } = useHasOnlineStreaming()

    const style = {
        transform: CSS.Transform.toString(transform ? { ...transform, scaleY: 1 } : null),
        transition,
    }

    const streamOptions = React.useMemo(() => {
        let options: { label: string, value: string }[] = []
        if (hasTorrentStreaming) {
            options.push({ label: "Torrent streaming", value: "torrent" })
        }
        if (hasDebridService) {
            options.push({ label: "Debrid streaming", value: "debrid" })
        }
        if (hasOnlineStreaming) {
            options.push({ label: "Online streaming", value: "online" })
        }
        return options
    }, [hasDebridService, hasOnlineStreaming, hasTorrentStreaming])

    // Set default watch type
    const t = React.useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (episode && !episode.watchType && (hasTorrentStreaming || hasDebridService)) {
            t.current = setTimeout(() => {
                setEpisodes(prev => {
                    const foundEp = prev.find(n => playlist_isSameEpisode(n, episode))
                    if (!foundEp) return prev
                    return prev.map(n => {
                        if (playlist_isSameEpisode(n, episode)) {
                            return {
                                ...n,
                                watchType: (hasTorrentStreaming ? "torrent" : hasDebridService ? "debrid" : hasOnlineStreaming
                                    ? "online"
                                    : "") as Anime_WatchType,
                            }
                        }
                        return n
                    })
                })
            }, 300)
        }
        return () => {
            if (t.current) clearTimeout(t.current)
        }
    }, [episode, hasDebridService, hasOnlineStreaming, hasTorrentStreaming])

    if (!episode) return null

    return (
        <li ref={setNodeRef} style={style}>
            <div
                className="px-2.5 py-2 bg-gray-900 hover:bg-gray-900/80 rounded-xl border flex gap-3 relative cursor-move"
                {...attributes} {...listeners}
            >
                <IconButton
                    className="absolute top-2 right-2 rounded-full cursor-pointer"
                    icon={<BiTrash />}
                    intent="alert-subtle"
                    size="sm"
                    onClick={(e) => {
                        setEpisodes((prev: Anime_PlaylistEpisode[]) => prev.filter(n => !playlist_isSameEpisode(n, episode)))
                    }}
                    onPointerDown={(e) => e.stopPropagation()}
                />
                <div className="size-24 aspect-square flex-none rounded-md overflow-hidden relative transition bg-[--background]">
                    {episode.episode!.episodeMetadata?.image && <SeaImage
                        data-episode-card-image
                        src={getImageUrl(episode.episode!.episodeMetadata?.image)}
                        alt={""}
                        fill
                        quality={100}
                        placeholder={imageShimmer(700, 475)}
                        sizes="20rem"
                        className={cn(
                            "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                            episode.isCompleted && "opacity-20",
                        )}
                    />}
                </div>
                <div className="max-w-full space-y-1">
                    <p className="text-sm text-[--muted] font-medium">{episode.episode?.baseAnime?.title?.userPreferred}</p>
                    <p className="">{episode.episode?.baseAnime?.format !== "MOVIE"
                        ? `Episode ${episode.episode!.episodeNumber}`
                        : "Movie"}{episode.isCompleted ? ` (Watched)` : ""}</p>

                    {(!episode.episode?.localFile && !episode.isNakama) && <div className="flex gap-1 flex-wrap">
                        {streamOptions.map(option => {
                            return <div
                                key={option.value}
                                className={cn(
                                    "text-sm flex w-fit py-1 px-2 rounded-xl hover:bg-[--subtle] text-[--muted] hover:text-[--foreground] transition border border-transparent cursor-pointer",
                                    option.value === episode.watchType && "border-white/20 bg-[--subtle] text-white hover:text-white",
                                )}
                                onPointerDown={e => e.stopPropagation()}
                                onClick={e => {
                                    e.stopPropagation()
                                    setEpisodes(prev => {
                                        const foundEp = prev.find(n => playlist_isSameEpisode(n, episode))
                                        if (!foundEp) return prev
                                        return prev.map(n => {
                                            if (playlist_isSameEpisode(n, episode)) {
                                                return {
                                                    ...n,
                                                    watchType: (option.value === n.watchType ? "" : option.value) as Anime_WatchType,
                                                }
                                            }
                                            return n
                                        })
                                    })
                                }}
                            >
                                {option.label}
                            </div>
                        })}
                    </div>}

                    {!!episode.episode?.localFile && <div>
                        <div className="text-sm text-[--muted] line-clamp-1 tracking-wide">
                            {episode.episode?.localFile?.name}
                        </div>
                    </div>}
                </div>
            </div>
        </li>
    )
}


type EntryEpisodeListProps = {
    entry: Anime_LibraryCollectionEntry
    selectedEpisodes: Anime_PlaylistEpisode[]
    setSelectedEpisodes: React.Dispatch<React.SetStateAction<Anime_PlaylistEpisode[]>>
}

function EntryEpisodeList(props: EntryEpisodeListProps) {

    const {
        entry,
        selectedEpisodes,
        setSelectedEpisodes,
        ...rest
    } = props

    const { data, isLoading } = useGetPlaylistEpisodes(entry.mediaId)
    const { episodeToAdd, selectedMedia, setSelectedMedia, setEpisodeToAdd } = usePlaylistEditorManager()

    const t = React.useRef<NodeJS.Timeout | null>(null)
    React.useEffect(() => {
        if (data && episodeToAdd && selectedMedia && selectedMedia === entry.mediaId) {
            t.current = setTimeout(() => {
                // Check if already added
                if (selectedEpisodes.find(n => n.episode?.baseAnime?.id === selectedMedia && n.episode.aniDBEpisode === episodeToAdd)) {
                    toast.info("Episode already added")
                    return
                }
                const foundEp = data?.find(n => n.episode?.aniDBEpisode === episodeToAdd)
                if (foundEp) {
                    handleSelect(foundEp)
                    toast.info("Episode added to playlist")
                    setEpisodeToAdd(null)
                }
            }, 400)
        }

        return () => {
            if (t.current) {
                clearTimeout(t.current)
            }
        }
    }, [data, episodeToAdd, selectedMedia, selectedEpisodes])

    const handleSelect = (ep: Anime_PlaylistEpisode) => {
        React.startTransition(() => {
            setSelectedEpisodes(prev => {
                if (prev.find(n => playlist_isSameEpisode(n, ep))) {
                    return prev.filter(n => !playlist_isSameEpisode(n, ep))
                }
                if (prev.length >= 20) {
                    toast.error("You can't add more than 20 episodes to a playlist")
                    return prev
                }
                return [...prev, ep]
            })
        })
    }

    return (
        <div className="flex flex-col gap-2 overflow-auto p-1">
            {isLoading && <LoadingSpinner />}
            {data?.length === 0 && <p className="text-center text-sm text-[--muted]">No episodes found</p>}
            {data?.filter(n => !!n.episode)?.sort((a, b) => a.episode!.progressNumber - b.episode!.progressNumber)?.map(ep => {
                return (
                    <div
                        key={playlist_getEpisodeKey(ep)}
                        className={cn(
                            "grid grid-cols-[auto,1fr] px-2.5 py-2 bg-[--background] rounded-md border cursor-pointer overflow-hidden items-center gap-3 opacity-80 max-w-full",
                            selectedEpisodes.find(n => playlist_isSameEpisode(ep, n))
                                ? "bg-gray-800 opacity-100 text-white ring-1 ring-[--zinc]"
                                : "hover:bg-[--subtle]",
                            "transition",
                        )}
                        onClick={() => handleSelect(ep)}
                    >
                        <div className="w-16 flex-none aspect-square rounded-md overflow-hidden relative transition bg-[--background]">
                            {ep.episode!.episodeMetadata?.image && <SeaImage
                                data-episode-card-image
                                src={getImageUrl(ep.episode!.episodeMetadata?.image)}
                                alt={""}
                                fill
                                quality={100}
                                placeholder={imageShimmer(700, 475)}
                                sizes="20rem"
                                className={cn(
                                    "object-cover rounded-lg object-center transition lg:group-hover/episode-card:scale-105 duration-200",
                                )}
                            />}
                        </div>
                        <div className="max-w-full">
                            <p className="">{entry.media?.format !== "MOVIE" ? `Episode ${ep.episode!.episodeNumber}` : "Movie"}</p>
                            {ep.episode!.localFile &&
                                <p className="text-xs text-[--muted] tracking-wide italic max-w-full line-clamp-2">{ep.episode!.localFile?.name}</p>}

                        </div>
                    </div>
                )
            })}
        </div>
    )
}
