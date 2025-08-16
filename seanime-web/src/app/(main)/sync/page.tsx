"use client"
import { AL_BaseAnime, AL_BaseManga, Local_QueueState } from "@/api/generated/types"
import {
    useLocalGetHasLocalChanges,
    useLocalGetLocalStorageSize,
    useLocalGetTrackedMediaItems,
    useLocalSetHasLocalChanges,
    useLocalSyncAnilistData,
    useLocalSyncData,
    useSetOfflineMode,
} from "@/api/hooks/local.hooks"
import { useGetMangaCollection } from "@/api/hooks/manga.hooks"
import { animeLibraryCollectionWithoutStreamsAtom } from "@/app/(main)/_atoms/anime-library-collection.atoms"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SyncAddMediaModal } from "@/app/(main)/sync/_containers/sync-add-media-modal"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Alert } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner, Spinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { anilist_getListDataFromEntry } from "@/lib/helpers/media"
import { WSEvents } from "@/lib/server/ws-events"
import { useAtomValue } from "jotai/react"
import React from "react"
import { LuCloud, LuCloudDownload, LuCloudOff, LuCloudUpload, LuFolderSync } from "react-icons/lu"
import { VscSyncIgnored } from "react-icons/vsc"
import { toast } from "sonner"

export const dynamic = "force-static"

export default function Page() {
    const serverStatus = useServerStatus()

    const [syncModalOpen, setSyncModalOpen] = React.useState(false)

    const { data: trackedMediaItems, isLoading } = useLocalGetTrackedMediaItems()
    const { mutate: syncLocal, isPending: isSyncingLocal } = useLocalSyncData()
    const { mutate: syncAnilist, isPending: isSyncingAnilist } = useLocalSyncAnilistData()
    const { data: hasLocalChanges } = useLocalGetHasLocalChanges()
    const { mutate: syncHasLocalChanges, isPending: isChangingLocalChangeStatus } = useLocalSetHasLocalChanges()
    const { data: localStorageSize } = useLocalGetLocalStorageSize()
    const { mutate: setOfflineMode, isPending: isSettingOfflineMode } = useSetOfflineMode()

    const trackedAnimeItems = React.useMemo(() => {
        return trackedMediaItems?.filter(n => n.type === "anime" && !!n.animeEntry?.media) ?? []
    }, [trackedMediaItems])

    const trackedMangaItems = React.useMemo(() => {
        return trackedMediaItems?.filter(n => n.type === "manga" && !!n.mangaEntry?.media) ?? []
    }, [trackedMediaItems])

    const animeLibraryCollection = useAtomValue(animeLibraryCollectionWithoutStreamsAtom)
    const { data: mangaLibraryCollection } = useGetMangaCollection()

    const unsavedAnime = React.useMemo(() => {
        const trackedIds = new Set(trackedAnimeItems.map(n => n.mediaId))
        const currentList = animeLibraryCollection?.lists?.find(n => n.type === "CURRENT")
        let unsavedAnime: AL_BaseAnime[] = []
        // only include entries that have local files
        for (const entry of currentList?.entries ?? []) {
            if (!trackedIds.has(entry.mediaId)) {
                unsavedAnime.push(entry.media!)
            }
        }
        return unsavedAnime
    }, [animeLibraryCollection?.lists, trackedAnimeItems])

    const unsavedManga = React.useMemo(() => {
        const trackedIds = new Set(trackedMangaItems.map(n => n.mediaId))
        const currentList = mangaLibraryCollection?.lists?.find(n => n.type === "CURRENT")
        let unsavedManga: AL_BaseManga[] = []
        for (const entry of currentList?.entries ?? []) {
            if (!trackedIds.has(entry.mediaId)) {
                unsavedManga.push(entry.media!)
            }
        }
        return unsavedManga
    }, [mangaLibraryCollection?.lists, trackedMangaItems])

    const [queueState, setQueueState] = React.useState<Local_QueueState | null>(null)
    useWebsocketMessageListener<Local_QueueState>({
        type: WSEvents.SYNC_LOCAL_QUEUE_STATE,
        onMessage: data => {
            setQueueState(data)
        },
    })

    function handleSyncLocal() {
        syncLocal(undefined, {
            onSuccess: () => {
                setSyncModalOpen(false)
            },
        })
    }

    function handleSyncAnilist() {
        syncAnilist(undefined, {
            onSuccess: () => {
                setSyncModalOpen(false)
            },
        })
    }

    function handleIgnoreLocalChanges() {
        syncHasLocalChanges({
            updated: false,
        }, {
            onSuccess: () => {
                toast.success("Local changes ignored.")
                handleSyncLocal()
            },
        })
    }

    if (isLoading) return <LoadingSpinner />

    if (serverStatus?.user?.isSimulated) {
        return <LuffyError
            title="Not authenticated"
        >
            This feature is only available for authenticated users.
        </LuffyError>
    }

    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative space-y-8"
        >

            <Button
                intent="gray-subtle"
                rounded
                className=""
                leftIcon={serverStatus?.isOffline ? <LuCloudOff className="text-2xl" /> : <LuCloud className="text-2xl" />}
                loading={isSettingOfflineMode}
                onClick={() => {
                    setOfflineMode({
                        enabled: !serverStatus?.isOffline,
                    })
                }}
            >
                {serverStatus?.isOffline ? "Disable offline mode" : "Enable offline mode"}
            </Button>

            <div className="flex flex-col lg:flex-row gap-2">
                <div>
                    <h2 className="">Offline media</h2>
                    <p className="text-[--muted]">
                        View the media you've saved locally for offline use.
                    </p>
                </div>

                <div className="flex flex-1"></div>

                <div className="contents">
                    <Modal
                        title="Sync"
                        open={syncModalOpen}
                        onOpenChange={v => {
                            if (isSyncingLocal) return
                            return setSyncModalOpen(v)
                        }}
                        trigger={<Button
                            intent="white-subtle"
                            rounded
                            leftIcon={<LuFolderSync className="text-2xl" />}
                            loading={isSyncingLocal}
                        >
                            Sync now
                        </Button>}
                    >
                        <div className="space-y-4">

                            <Button
                                intent="white"
                                rounded
                                className="w-full"
                                leftIcon={<LuCloudDownload className="text-2xl" />}
                                loading={isSyncingLocal}
                                disabled={isSyncingAnilist}
                                onClick={handleSyncLocal}
                            >
                                Update local data
                            </Button>
                            <p className="text-sm">
                                Update your local snapshots with the data from AniList.
                                This will overwrite your offline changes. You can automate this in <kbd>Settings {`>`} Seanime {`>`} Offline</kbd>.
                            </p>
                            <Separator />
                            <Button
                                intent="primary-subtle"
                                rounded
                                className="w-full"
                                leftIcon={<LuCloudUpload className="text-2xl" />}
                                disabled={isSyncingLocal}
                                loading={isSyncingAnilist}
                                onClick={handleSyncAnilist}
                            >
                                Upload local changes to AniList
                            </Button>
                            <p className="text-sm">
                                Update your AniList lists with the data from your local snapshots.
                                This should be done after you've made changes offline.
                            </p>

                            <Alert
                                intent="warning-basic"
                                description="Changes are irreversible."
                            />
                        </div>
                    </Modal>

                    <SyncAddMediaModal
                        savedMediaIds={trackedMediaItems?.map(n => n.mediaId) ?? []}
                    />
                </div>
            </div>

            {(!!unsavedAnime?.length || !!unsavedManga?.length) && (
                <Alert
                    intent="info-basic"
                    className="border-transparent"
                    description={
                        <div className="space-y-2">
                            <p>
                                <span>You have not saved {!!unsavedAnime?.length
                                    ? `${unsavedAnime?.length} anime`
                                    : ""}{(!!unsavedAnime?.length && !!unsavedManga?.length) ? " and " : ""}{!!unsavedManga?.length
                                    ? `${unsavedManga?.length} manga`
                                    : ""} that you're currently {!!unsavedAnime?.length
                                    ? "watching"
                                    : ""}{(!!unsavedAnime.length && !!unsavedManga.length) ? " and " : ""}{!!unsavedManga?.length
                                    ? "reading"
                                    : ""}.</span>
                            </p>
                        </div>
                    }
                />
            )}

            <p className="text-sm">
                <span>Local storage size: </span>
                <span>{localStorageSize}</span>
            </p>

            {hasLocalChanges && <>
                <Alert
                    intent="warning"
                    description={<div className="space-y-2">
                        <p>
                            <span>You have local changes that have not been synced to AniList.</span>
                            {serverStatus?.settings?.library?.autoSyncOfflineLocalData &&
                                <span> Automatic refreshing of offline data will be paused.</span>}
                        </p>
                        <div className="flex items-center gap-2 flex-wrap">
                            <Button
                                intent="white"
                                leftIcon={<LuCloudUpload className="text-2xl" />}
                                onClick={() => {
                                    handleSyncAnilist()
                                    syncHasLocalChanges({
                                        updated: false,
                                    })
                                }}
                                loading={isSyncingAnilist}
                                disabled={isChangingLocalChangeStatus}
                            >
                                Upload local changes
                            </Button>
                            <Button
                                intent="alert"
                                leftIcon={<VscSyncIgnored className="text-2xl" />}
                                onClick={handleIgnoreLocalChanges}
                                loading={isChangingLocalChangeStatus}
                                disabled={isSyncingAnilist}
                            >
                                Delete local changes
                            </Button>
                        </div>
                    </div>}
                />
            </>}

            {/*{(queueState && (Object.keys(queueState.animeTasks!).length > 0 || Object.keys(queueState.mangaTasks!).length > 0)) &&*/}
            {/*    <div className="border rounded-[--radius-md] p-2">*/}
            {/*        <p className="flex items-center gap-1">*/}
            {/*            <Spinner className="size-6" />*/}
            {/*            <span>Syncing in progress</span>*/}
            {/*        </p>*/}
            {/*    </div>}*/}

            {(!trackedAnimeItems?.length && !trackedMangaItems?.length) && <LuffyError
                title="No tracked media"
            />}

            {!!trackedAnimeItems?.length && <div className="space-y-4">
                <h3>Saved anime</h3>
                <MediaCardLazyGrid itemCount={trackedAnimeItems?.length}>
                    {trackedAnimeItems?.map((item) => (
                        <MediaEntryCard
                            key={item.mediaId}
                            type="anime"
                            media={item.animeEntry!.media!}
                            listData={anilist_getListDataFromEntry(item.animeEntry!)}
                            overlay={!!queueState?.animeTasks?.[item.mediaId] && <SyncingBadge />}
                            containerClassName={cn(!!queueState?.animeTasks?.[item.mediaId] && "animate-pulse")}
                        />
                    ))}
                </MediaCardLazyGrid>
            </div>}

            {!!trackedMangaItems?.length && <div className="space-y-4">
                <h3>Saved manga</h3>
                <MediaCardLazyGrid itemCount={trackedMangaItems?.length}>
                    {trackedMangaItems?.map((item) => (
                        <MediaEntryCard
                            key={item.mediaId}
                            type="manga"
                            media={item.mangaEntry!.media!}
                            listData={anilist_getListDataFromEntry(item.mangaEntry!)}
                            overlay={!!queueState?.mangaTasks?.[item.mediaId] && <SyncingBadge />}
                            containerClassName={cn(!!queueState?.mangaTasks?.[item.mediaId] && "animate-pulse")}
                        />
                    ))}
                </MediaCardLazyGrid>
            </div>}
        </PageWrapper>
    )
}

function SyncingBadge() {
    return (
        <Badge
            intent="gray-solid"
            className="rounded-tl-md rounded-bl-none rounded-tr-none rounded-br-md bg-gray-950 border gap-0"
        >
            <Spinner className="size-4 px-0" />
            <span>
                Syncing
            </span>
        </Badge>
    )
}


