import { Anime_LibraryCollection, Anime_LibraryCollectionEntry, Manga_Collection, Manga_CollectionEntry } from "@/api/generated/types"
import { useLocalAddTrackedMedia, useLocalRemoveTrackedMedia } from "@/api/hooks/local.hooks"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { animeLibraryCollectionAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { imageShimmer } from "@/components/shared/image-helpers"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { useAtomValue } from "jotai/react"
import Image from "next/image"
import React from "react"
import { MdOutlineDownloadForOffline } from "react-icons/md"

type SyncAddMediaModalProps = {
    savedMediaIds: number[]
}

export function SyncAddMediaModal(props: SyncAddMediaModalProps) {

    const {
        savedMediaIds,
        ...rest
    } = props

    const [selectedMedia, setSelectedMedia] = React.useState<{ mediaId: number, type: "manga" | "anime" }[]>([])

    const { mutate: addMedia, isPending: isAdding } = useLocalAddTrackedMedia()

    function handleSave() {
        addMedia({
            media: selectedMedia,
        }, {
            onSuccess: () => {
                setSelectedMedia([])
            },
        })
    }

    return (
        <Modal
            title="Saved media"
            contentClass="max-w-4xl"
            trigger={<Button
                intent="success-subtle"
                rounded
                leftIcon={<MdOutlineDownloadForOffline className="text-2xl" />}
                loading={isAdding}
            >
                Save media
            </Button>}
        >

            <p className="text-[--muted]">
                Select the media you want to save locally. Click on already saved media to remove it from local storage.
            </p>

            <MediaSelector
                selectedMedia={selectedMedia}
                setSelectedMedia={setSelectedMedia}
                savedMediaIds={savedMediaIds}
                onSave={handleSave}
            />
        </Modal>
    )
}

type MediaSelectorProps = {
    children?: React.ReactNode
    savedMediaIds: number[]
    selectedMedia: { mediaId: number, type: "manga" | "anime" }[]
    setSelectedMedia: React.Dispatch<React.SetStateAction<{ mediaId: number, type: "manga" | "anime" }[]>>
    onSave: () => void
}

function MediaSelector(props: MediaSelectorProps) {

    const {
        savedMediaIds,
        selectedMedia,
        setSelectedMedia,
        onSave,
        ...rest
    } = props

    const animeLibraryCollection = useAtomValue(animeLibraryCollectionAtom)

    const { data: mangaLibraryCollection } = useGetMangaCollection()

    const { mutate: removeMedia, isPending: isRemoving } = useLocalRemoveTrackedMedia()

    function handleToggleAnime(mediaId: number) {
        setSelectedMedia(prev => {
            if (prev.find(n => n.mediaId === mediaId)) {
                return prev.filter(n => n.mediaId !== mediaId)
            } else {
                return [...prev, { mediaId, type: "anime" as const }]
            }
        })
    }

    function handleToggleManga(mediaId: number) {
        setSelectedMedia(prev => {
            if (prev.find(n => n.mediaId === mediaId)) {
                return prev.filter(n => n.mediaId !== mediaId)
            } else {
                return [...prev, { mediaId, type: "manga" as const }]
            }
        })
    }

    return (
        <div className="space-y-4">

            <div className="flex items-center">
                <div className="flex flex-1"></div>

                <Button
                    intent="white"
                    onClick={onSave}
                    disabled={selectedMedia.length === 0}
                    rounded
                    leftIcon={<MdOutlineDownloadForOffline className="text-2xl" />}
                >
                    Save locally
                </Button>
            </div>

            {animeLibraryCollection && <>
                <h2 className="text-center">Anime</h2>
                <MediaList
                    collection={animeLibraryCollection}
                    entry={entry => (
                        <MediaItem
                            entry={entry}
                            onClick={() => handleToggleAnime(entry.mediaId)}
                            isSelected={!!selectedMedia.find(n => n.mediaId === entry.mediaId)}
                            isSaved={savedMediaIds.includes(entry.mediaId)}
                            onUntrack={() => {
                                removeMedia({ mediaId: entry.mediaId, type: "anime" })
                            }}
                            isPending={isRemoving}
                        />
                    )}
                />
            </>}
            {mangaLibraryCollection && <>
                <h2 className="text-center">Manga</h2>
                <MediaList
                    collection={mangaLibraryCollection}
                    entry={entry => (
                        <MediaItem
                            entry={entry}
                            onClick={() => handleToggleManga(entry.mediaId)}
                            isSelected={!!selectedMedia.find(n => n.mediaId === entry.mediaId)}
                            isSaved={savedMediaIds.includes(entry.mediaId)}
                            onUntrack={() => {
                                removeMedia({ mediaId: entry.mediaId, type: "manga" })
                            }}
                            isPending={isRemoving}
                        />
                    )}
                />
            </>}
        </div>
    )
}

function MediaList(props: {
    collection: Anime_LibraryCollection | Manga_Collection,
    entry: (entry: Anime_LibraryCollectionEntry | Manga_CollectionEntry) => React.ReactElement,
}) {
    const { collection, entry } = props

    const lists = React.useMemo(() => {
        return {
            CURRENT: collection.lists?.find(n => n.type === "CURRENT")
                ?.entries
                ?.filter(Boolean)
                ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
            PLANNING: collection.lists?.find(n => n.type === "PLANNING")
                ?.entries
                ?.filter(Boolean)
                ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
            COMPLETED: collection.lists?.find(n => n.type === "COMPLETED")
                ?.entries
                ?.filter(Boolean)
                ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
            PAUSED: collection.lists?.find(n => n.type === "PAUSED")
                ?.entries
                ?.filter(Boolean)
                ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
            DROPPED: collection.lists?.find(n => n.type === "DROPPED")
                ?.entries
                ?.filter(Boolean)
                ?.toSorted((a, b) => a.media!.title!.userPreferred!.localeCompare(b.media!.title!.userPreferred!)) ?? [],
        }
    }, [collection])

    return (
        <>
            {!!lists.CURRENT.length && <>
                <h4 className="border-b pb-1 mb-1">Current</h4>
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                    {lists.CURRENT.map(n => {
                        return <React.Fragment key={n.mediaId}>
                            {entry(n)}
                        </React.Fragment>
                    })}
                </div>
            </>}
            {!!lists.PAUSED.length && <>
                <h4 className="border-b pb-1 mb-1">Paused</h4>
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                    {lists.PAUSED.map(n => {
                        return <React.Fragment key={n.mediaId}>
                            {entry(n)}
                        </React.Fragment>
                    })}
                </div>
            </>}
            {!!lists.PLANNING.length && <>
                <h4 className="border-b pb-1 mb-1">Planning</h4>
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                    {lists.PLANNING.map(n => {
                        return <React.Fragment key={n.mediaId}>
                            {entry(n)}
                        </React.Fragment>
                    })}
                </div>
            </>}
            {!!lists.COMPLETED.length && <>
                <h4 className="border-b pb-1 mb-1">Completed</h4>
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                    {lists.COMPLETED.map(n => {
                        return <React.Fragment key={n.mediaId}>
                            {entry(n)}
                        </React.Fragment>
                    })}
                </div>
            </>}
            {!!lists.DROPPED.length && <>
                <h4 className="border-b pb-1 mb-1">Dropped</h4>
                <div className="grid grid-cols-3 md:grid-cols-6 gap-2">
                    {lists.DROPPED.map(n => {
                        return <React.Fragment key={n.mediaId}>
                            {entry(n)}
                        </React.Fragment>
                    })}
                </div>
            </>}
        </>
    )
}

function MediaItem(props: {
    entry: Anime_LibraryCollectionEntry | Manga_CollectionEntry,
    onClick: () => void,
    onUntrack: () => void,
    isSelected: boolean,
    isSaved: boolean
    isPending: boolean
}) {
    const { entry, onClick, isSelected, isSaved, onUntrack, isPending } = props

    const confirmUntrack = useConfirmationDialog({
        title: "Remove offline data",
        description: "This action will remove the offline data for this media entry. Are you sure you want to proceed?",
        onConfirm: () => {
            onUntrack()
        },
    })

    return (
        <>
            <div
                key={entry.mediaId}
                className={cn(
                    "col-span-1 aspect-[6/7] rounded-[--radius-md] overflow-hidden relative bg-[var(--background)] cursor-pointer transition-opacity select-none",
                    isSaved && "",
                    isPending && "pointer-events-none",
                )}
                onClick={() => {
                    if (isPending) return
                    if (!isSaved) {
                        onClick()
                    } else {
                        confirmUntrack.open()
                    }
                }}
            >
                <Image
                    src={entry.media?.coverImage?.large || entry.media?.bannerImage || ""}
                    placeholder={imageShimmer(700, 475)}
                    sizes="10rem"
                    fill
                    alt=""
                    className={cn(
                        "object-center object-cover rounded-[--radius-md] transition-opacity",
                        isSelected ? "opacity-100" : "opacity-60",
                    )}
                />
                <p
                    className={cn(
                        "line-clamp-2 text-sm absolute m-2 bottom-0 font-semibold z-[10]",
                        isSaved && "text-[--green]",
                    )}
                >
                    {entry.media?.title?.userPreferred || entry.media?.title?.romaji}
                </p>
                <div
                    className="z-[5] absolute bottom-0 w-full h-[80%] bg-gradient-to-t from-[--background] to-transparent"
                />
                <div
                    className={cn(
                        "z-[5] absolute top-0 w-full h-[80%] bg-gradient-to-b from-[--background] to-transparent transition-opacity",
                        isSelected ? "opacity-0" : "opacity-100 hover:opacity-80",
                    )}
                />
            </div>

            <ConfirmationDialog {...confirmUntrack} />
        </>
    )
}
