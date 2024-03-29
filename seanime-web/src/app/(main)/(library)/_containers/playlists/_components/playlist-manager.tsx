import { libraryCollectionAtom } from "@/app/(main)/_loaders/library-collection"
import { imageShimmer } from "@/components/shared/styling/image-helpers"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { BaseMediaFragment } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaQuery } from "@/lib/server/query"
import { LibraryCollectionEntry, LocalFile } from "@/lib/server/types"
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

    const libraryCollection = useAtomValue(libraryCollectionAtom)

    const { data: localFiles } = useSeaQuery<LocalFile[]>({
        endpoint: SeaEndpoints.LOCAL_FILES,
        queryKey: ["get-local-files"],
    })

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


    return (
        <div className="space-y-4">

            <div className="space-y-2">
                <Modal
                    title="Select an anime"
                    trigger={<Button
                        leftIcon={<BiPlus className="text-2xl" />}
                        intent="white"
                        className="rounded-full"
                        disabled={paths.length >= 10}
                    >Add an episode</Button>}
                >
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                        {libraryCollection?.lists?.filter(n => n.type === "planned" || n.type === "paused" || n.type === "current")
                            ?.flatMap(n => n.entries)
                            ?.map(entry => {
                                return (
                                    <Modal
                                        title={entry.media?.title?.userPreferred || entry.media?.title?.romaji || ""}
                                        trigger={(
                                            <div
                                                key={entry.mediaId}
                                                className="col-span-1 aspect-[6/7] rounded-md border overflow-hidden relative transition cursor-pointer bg-[#0c0c0c] md:opacity-60 md:hover:opacity-100 md:hover:scale-105"
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
                            {paths.map((path, index) => (
                                <SortableItem
                                    key={path}
                                    id={path}
                                    localFile={localFiles?.find(n => n.path === path)}
                                    media={libraryCollection?.lists?.flatMap(n => n.entries)
                                        ?.find(n => localFiles?.find(n => n.path === path)?.mediaId === n.mediaId)?.media}
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

function SortableItem({ localFile, id, media, setPaths }: {
    id: string,
    localFile: LocalFile | undefined,
    media: BaseMediaFragment | undefined,
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
                className="px-2.5 py-2 bg-[#0c0c0c] border-[--red] rounded-md border flex gap-3 relative"

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
                className="px-2.5 py-2 bg-[#0c0c0c] rounded-md border flex gap-3 relative"

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

                    className="w-16 aspect-square rounded-md border overflow-hidden relative transition bg-[#0c0c0c]"
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
                    <p className="text-lg text-white font-semibold">
                        <span>
                            {media?.format !== "MOVIE" ? `Episode ${localFile.metadata.episode}` : "Movie"}
                        </span>
                        <span className="text-gray-400 font-medium max-w-lg truncate">
                            {" - "}{media?.title?.userPreferred || media?.title?.romaji}
                        </span>
                    </p>
                    <p className="text-sm text-[--muted] font-normal italic line-clamp-1">{localFile.name}</p>
                </div>
            </div>
        </li>
    )
}


type EntryEpisodeListProps = {
    entry: LibraryCollectionEntry
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

    const { data } = useSeaQuery<LocalFile[]>({
        endpoint: SeaEndpoints.PLAYLIST_EPISODES.replace("{id}", String(entry.mediaId)).replace("{progress}", String(entry.listData?.progress || 0)),
        queryKey: ["playlist-episodes", entry.mediaId],
    })

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
            {data?.sort((a, b) => a.metadata.episode - b.metadata.episode)?.map(lf => {
                return (
                    <div
                        key={lf.path}
                        className={cn(
                            "px-2.5 py-2 bg-[#0c0c0c] rounded-md border cursor-pointer opacity-80 max-w-full",
                            selectedPaths.includes(lf.path) ? "bg-gray-800 opacity-100 text-white ring-1 ring-[--zinc]" : "hover:bg-[--subtle]",
                            "transition",
                        )}
                        onClick={() => handleSelect(lf.path)}
                    >
                        <p className="">{entry.media?.format !== "MOVIE" ? `Episode ${lf.metadata.episode}` : "Movie"}</p>
                        <p className="text-sm text-[--muted] font-normal italic max-w-lg line-clamp-1">{lf.name}</p>
                    </div>
                )
            })}
        </div>
    )
}
