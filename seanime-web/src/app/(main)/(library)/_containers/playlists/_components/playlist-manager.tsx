import { AL_BaseAnime, Anime_LibraryCollectionEntry, Anime_LocalFile } from "@/api/generated/types"
import { useGetLocalFiles } from "@/api/hooks/localfiles.hooks"
import { useGetPlaylistEpisodes } from "@/api/hooks/playlist.hooks"
import { animeLibraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { DndContext, DragEndEvent } from "@dnd-kit/core"
import { restrictToVerticalAxis } from "@dnd-kit/modifiers"
import { arrayMove, SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable"
import { CSS } from "@dnd-kit/utilities"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React from "react"
import { BiPlus, BiTrash } from "react-icons/bi"
import { toast } from "sonner"


type PlaylistManagerProps = {
    paths: string[]
    setPaths: (paths: string[]) => void
}

export function PlaylistManager(props: PlaylistManagerProps) {

    const {
        paths: controlledPaths,
        setPaths: onChange,
        ...rest
    } = props

    const libraryCollection = useAtomValue(animeLibraryCollectionAtom)

    const { data: localFiles } = useGetLocalFiles()

    const firstRender = React.useRef(true)

    const [paths, setPaths] = React.useState(controlledPaths)

    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        setPaths(controlledPaths)
    }, [controlledPaths])

    React.useEffect(() => {
        onChange(paths)
    }, [paths])

    const handleDragEnd = React.useCallback((event: DragEndEvent) => {
        const { active, over } = event

        if (active.id !== over?.id) {
            setPaths((items) => {
                const oldIndex = items.indexOf(active.id as any)
                const newIndex = items.indexOf(over?.id as any)

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
                        disabled={paths.length >= 10}
                    >Add an episode</Button>}
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
                                <PlaylistMediaEntry key={entry.mediaId} entry={entry} paths={paths} setPaths={setPaths} />
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
                    items={paths}
                >
                    <div className="space-y-2">
                        <ul className="space-y-2">
                            {paths.map(path => localFiles?.find(n => n.path === path))?.filter(Boolean).map((lf, index) => (
                                <SortableItem
                                    key={lf.path}
                                    id={lf.path}
                                    localFile={lf}
                                    media={libraryCollection?.lists?.flatMap(n => n.entries)
                                        ?.filter(Boolean)
                                        ?.find(n => lf?.mediaId === n.mediaId)?.media}
                                    setPaths={setPaths}
                                />
                            ))}
                        </ul>
                    </div>
                </SortableContext>
            </DndContext>
        </div>
    )
}

type PlaylistMediaEntryProps = {
    entry: Anime_LibraryCollectionEntry
    paths: string[]
    setPaths: React.Dispatch<React.SetStateAction<string[]>>
}

function PlaylistMediaEntry(props: PlaylistMediaEntryProps) {
    const { entry, paths, setPaths } = props
    return <Modal
        title={entry.media?.title?.userPreferred || entry.media?.title?.romaji || ""}
        trigger={(
            <div
                key={entry.mediaId}
                className="col-span-1 aspect-[6/7] rounded-[--radius-md] border overflow-hidden relative transition cursor-pointer bg-[var(--background)] md:opacity-60 md:hover:opacity-100"
            >
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
        )}
    >
        <EntryEpisodeList
            selectedPaths={paths}
            setSelectedPaths={setPaths}
            entry={entry}
        />
    </Modal>
}

function SortableItem({ localFile, id, media, setPaths }: {
    id: string,
    localFile: Anime_LocalFile | undefined,
    media: AL_BaseAnime | undefined,
    setPaths: any
}) {
    const {
        attributes,
        listeners,
        setNodeRef,
        transform,
        transition,
    } = useSortable({ id: id })

    const style = {
        transform: CSS.Transform.toString(transform),
        transition,
    }

    if (!localFile) return null

    if (!media) return (
        <li ref={setNodeRef} style={style}>
            <div
                className="px-2.5 py-2 bg-[var(--background)] border-[--red] rounded-[--radius-md] border flex gap-3 relative"

            >
                <IconButton
                    className="absolute top-2 right-2 rounded-full"
                    icon={<BiTrash />}
                    intent="alert-subtle"
                    size="sm"
                    onClick={() => setPaths((prev: string[]) => prev.filter(n => n !== id))}

                />
                <div

                    className="rounded-full w-4 h-auto bg-[--muted] md:bg-[--subtle] md:hover:bg-[--subtle-highlight] cursor-move"
                    {...attributes} {...listeners}
                />
                <div>
                    <p className="text-lg text-white font-semibold">
                        <span>
                            ???
                        </span>
                        <span className="text-gray-400 font-medium max-w-lg truncate">
                        </span>
                    </p>
                    <p className="text-sm text-[--muted] font-normal italic line-clamp-1">{localFile.name}</p>
                </div>
            </div>
        </li>
    )

    return (
        <li ref={setNodeRef} style={style}>
            <div
                className="px-2.5 py-2 bg-[var(--background)] rounded-[--radius-md] border flex gap-3 relative"

            >
                <IconButton
                    className="absolute top-2 right-2 rounded-full"
                    icon={<BiTrash />}
                    intent="alert-subtle"
                    size="sm"
                    onClick={() => setPaths((prev: string[]) => prev.filter(n => n !== id))}

                />
                <div
                    className="rounded-full w-4 h-auto bg-[--muted] md:bg-[--subtle] md:hover:bg-[--subtle-highlight] cursor-move"
                    {...attributes} {...listeners}
                />
                <div

                    className="w-16 aspect-square rounded-[--radius-md] border overflow-hidden relative transition bg-[var(--background)]"
                >
                    <Image
                        src={media?.coverImage?.large || media?.bannerImage || ""}
                        placeholder={imageShimmer(700, 475)}
                        sizes="10rem"
                        fill
                        alt=""
                        className="object-center object-cover"

                    />
                </div>
                <div>
                    <div className="text-lg text-white font-semibold flex gap-1">
                        {localFile.metadata && <p>
                            {media?.format !== "MOVIE" ? `Episode ${localFile.metadata?.episode}` : "Movie"}
                        </p>}
                        <p className="max-w-full truncate text-gray-400 font-medium max-w-lg truncate">
                            {" - "}{media?.title?.userPreferred || media?.title?.romaji}
                        </p>
                    </div>
                    <p className="text-sm text-[--muted] font-normal italic line-clamp-1">{localFile.name}</p>
                </div>
            </div>
        </li>
    )
}


type EntryEpisodeListProps = {
    entry: Anime_LibraryCollectionEntry
    selectedPaths: string[]
    setSelectedPaths: React.Dispatch<React.SetStateAction<string[]>>
}

function EntryEpisodeList(props: EntryEpisodeListProps) {

    const {
        entry,
        selectedPaths,
        setSelectedPaths,
        ...rest
    } = props

    const { data } = useGetPlaylistEpisodes(entry.mediaId, entry.listData?.progress || 0)

    const handleSelect = (value: string) => {
        if (selectedPaths.length <= 10) {
            setSelectedPaths(prev => {
                if (prev.includes(value)) {
                    return prev.filter(n => n !== value)
                }
                return [...prev, value]
            })
        } else {
            toast.error("You can't add more than 10 episodes to a playlist")
        }
    }

    return (
        <div className="flex flex-col gap-2 overflow-auto p-1">
            {data?.filter(n => !!n.metadata)?.sort((a, b) => a.metadata!.episode - b.metadata!.episode)?.map(lf => {
                return (
                    <div
                        key={lf.path}
                        className={cn(
                            "px-2.5 py-2 bg-[var(--background)] rounded-[--radius-md] border cursor-pointer opacity-80 max-w-full",
                            selectedPaths.includes(lf.path) ? "bg-gray-800 opacity-100 text-white ring-1 ring-[--zinc]" : "hover:bg-[--subtle]",
                            "transition",
                        )}
                        onClick={() => handleSelect(lf.path)}
                    >
                        <p className="">{entry.media?.format !== "MOVIE" ? `Episode ${lf.metadata!.episode}` : "Movie"}</p>
                        <p className="text-sm text-[--muted] font-normal italic max-w-lg line-clamp-1">{lf.name}</p>
                    </div>
                )
            })}
        </div>
    )
}
