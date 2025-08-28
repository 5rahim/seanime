import { Anime_LibraryCollection, Anime_LibraryCollectionEntry, Anime_PlaylistEpisode, Anime_WatchType } from "@/api/generated/types"
import { useGetPlaylistEpisodes } from "@/api/hooks/playlist.hooks"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-manager"
import { useHasTorrentOrDebridInclusion } from "@/app/(main)/_hooks/use-server-status"
import { imageShimmer } from "@/components/shared/image-helpers"
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
import Image from "next/image"
import React from "react"
import { BiPlus, BiTrash } from "react-icons/bi"
import { IoLibrarySharp } from "react-icons/io5"
import { toast } from "sonner"

function isSameEpisode(a: Anime_PlaylistEpisode, b: Anime_PlaylistEpisode) {
    return a.episode?.aniDBEpisode === b.episode?.aniDBEpisode && a.episode?.baseAnime?.id === b.episode?.baseAnime?.id
}

function getEpisodeKey(a: Anime_PlaylistEpisode) {
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
                const oldIndex = items.findIndex(item => getEpisodeKey(item) === active.id)
                const newIndex = items.findIndex(item => getEpisodeKey(item) === over?.id)

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
                        intent="white"
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
                    items={episodes.map(n => getEpisodeKey(n))}
                >
                    <div className="space-y-2">
                        <ul className="space-y-2">
                            {episodes.filter(Boolean).map((ep, index) => (
                                <SortableItem
                                    key={getEpisodeKey(ep)}
                                    id={getEpisodeKey(ep)}
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

    return (
        <div
            key={entry.mediaId}
            className="col-span-1 aspect-[6/7] rounded-md border overflow-hidden relative transition cursor-pointer bg-[--background] md:opacity-60 md:hover:opacity-100"
            onClick={() => setSelectedMedia(entry.mediaId)}
        >
            {entry.libraryData && <div data-media-entry-card-body-library-badge className="absolute z-[1] left-0 top-0">
                <Badge
                    size="lg" intent="warning-solid"
                    className="rounded-md rounded-bl-none rounded-tr-none text-orange-900 opacity-80"
                ><IoLibrarySharp /></Badge>
            </div>}

            <Image
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

const radioGroupClasses = {
    itemClass: cn(
        "border-transparent absolute top-2 right-2 bg-transparent dark:bg-transparent dark:data-[state=unchecked]:bg-transparent",
        "data-[state=unchecked]:bg-transparent data-[state=unchecked]:hover:bg-transparent dark:data-[state=unchecked]:hover:bg-transparent",
        "focus-visible:ring-0 focus-visible:ring-offset-0 focus-visible:ring-offset-transparent",
    ),
    stackClass: "space-y-0 flex flex-wrap gap-2",
    itemIndicatorClass: "hidden",
    itemLabelClass: "font-normal text-sm tracking-wide line-clamp-1 truncate flex flex-col items-center data-[state=checked]:text-[--gray] cursor-pointer",
    itemContainerClass: cn(
        "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[state=checked]:bg-white dark:data-[state=checked]:bg-gray-950",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[state=checked]:border-[--gray] data-[state=checked]:ring-offset-0",
        "w-fit",
    ),
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

    const { hasTorrentStreaming, hasDebridStreaming } = useHasTorrentOrDebridInclusion()

    const style = {
        transform: CSS.Transform.toString(transform ? { ...transform, scaleY: 1 } : null),
        transition,
    }

    const streamOptions = React.useMemo(() => {
        let options: { label: string, value: string }[] = []
        if (hasTorrentStreaming) {
            options.push({ label: "Torrent streaming", value: "torrent" })
        }
        if (hasDebridStreaming) {
            options.push({ label: "Debrid streaming", value: "debrid" })
        }
        return options
    }, [hasDebridStreaming, hasTorrentStreaming])

    // Set default watch type
    React.useEffect(() => {
        if (episode && !episode.watchType && (hasTorrentStreaming || hasDebridStreaming)) {
            setEpisodes(prev => {
                const foundEp = prev.find(n => isSameEpisode(n, episode))
                if (!foundEp) return prev
                return prev.map(n => {
                    if (isSameEpisode(n, episode)) {
                        return {
                            ...n,
                            watchType: (hasTorrentStreaming ? "torrent" : hasDebridStreaming ? "debrid" : "") as Anime_WatchType,
                        }
                    }
                    return n
                })
            })
        }
    }, [episode, hasDebridStreaming, hasTorrentStreaming])

    if (!episode) return null

    return (
        <li ref={setNodeRef} style={style}>
            <div
                className="px-2.5 py-2 bg-[--background] rounded-md border flex gap-3 relative cursor-move"
                {...attributes} {...listeners}
            >
                <IconButton
                    className="absolute top-2 right-2 rounded-full cursor-pointer"
                    icon={<BiTrash />}
                    intent="alert-subtle"
                    size="sm"
                    onClick={(e) => {
                        setEpisodes((prev: Anime_PlaylistEpisode[]) => prev.filter(n => !isSameEpisode(n, episode)))
                    }}
                    onPointerDown={(e) => e.stopPropagation()}
                />
                <div className="size-24 aspect-square flex-none rounded-md overflow-hidden relative transition bg-[--background]">
                    {episode.episode!.episodeMetadata?.image && <Image
                        data-episode-card-image
                        src={getImageUrl(episode.episode!.episodeMetadata?.image)}
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
                    <p className="text-sm text-[--muted]">{episode.episode?.baseAnime?.title?.userPreferred}</p>
                    <p className="">{episode.episode?.baseAnime?.format !== "MOVIE" ? `Episode ${episode.episode!.episodeNumber}` : "Movie"}</p>

                    {(!episode.episode?.localFile && !episode.isNakama) && <div className="mt-1">
                        {streamOptions.map(option => {
                            return <div
                                key={option.value}
                                className={cn(
                                    "text-sm flex w-fit py-1 px-1.5 rounded-md bg-[--subtle] border border-transparent cursor-pointer",
                                    option.value === episode.watchType && "border-[--white]",
                                )}
                                onPointerDown={e => e.stopPropagation()}
                                onClick={e => {
                                    e.stopPropagation()
                                    setEpisodes(prev => {
                                        const foundEp = prev.find(n => isSameEpisode(n, episode))
                                        if (!foundEp) return prev
                                        return prev.map(n => {
                                            if (isSameEpisode(n, episode)) {
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

                        {/*<RadioGroup*/}
                        {/*    {...radioGroupClasses}*/}
                        {/*    options={streamOptions}*/}
                        {/*    value={episode.watchType}*/}
                        {/*    onValueChange={(value) => {*/}
                        {/*        setEpisodes(prev => {*/}
                        {/*            const foundEp = prev.find(n => isSameEpisode(n, episode))*/}
                        {/*            console.log(foundEp, value)*/}
                        {/*            if(!foundEp) return prev*/}
                        {/*            return prev.map(n => {*/}
                        {/*                if(isSameEpisode(n, episode)) {*/}
                        {/*                    return {*/}
                        {/*                        ...n,*/}
                        {/*                        watchType: value as Anime_WatchType,*/}
                        {/*                    }*/}
                        {/*                }*/}
                        {/*                return n*/}
                        {/*            })*/}
                        {/*        })*/}
                        {/*    }}*/}
                        {/*/>*/}
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
        if (episodeToAdd && selectedMedia && selectedMedia === entry.mediaId) {
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
    }, [episodeToAdd, selectedMedia, selectedEpisodes])

    const handleSelect = (ep: Anime_PlaylistEpisode) => {
        React.startTransition(() => {
            setSelectedEpisodes(prev => {
                if (prev.find(n => isSameEpisode(n, ep))) {
                    return prev.filter(n => !isSameEpisode(n, ep))
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
                        key={getEpisodeKey(ep)}
                        className={cn(
                            "grid grid-cols-[auto,1fr] px-2.5 py-2 bg-[--background] rounded-md border cursor-pointer overflow-hidden items-center gap-3 opacity-80 max-w-full",
                            selectedEpisodes.find(n => isSameEpisode(ep, n))
                                ? "bg-gray-800 opacity-100 text-white ring-1 ring-[--zinc]"
                                : "hover:bg-[--subtle]",
                            "transition",
                        )}
                        onClick={() => handleSelect(ep)}
                    >
                        <div className="w-16 flex-none aspect-square rounded-md overflow-hidden relative transition bg-[--background]">
                            {ep.episode!.episodeMetadata?.image && <Image
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
