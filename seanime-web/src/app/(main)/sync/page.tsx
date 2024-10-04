"use client"
import { Sync_QueueState } from "@/api/generated/types"
import { useSyncAnilistData, useSyncGetTrackedMediaItems, useSyncLocalData } from "@/api/hooks/sync.hooks"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Spinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { anilist_getListDataFromEntry } from "@/lib/helpers/media"
import { WSEvents } from "@/lib/server/ws-events"
import React from "react"
import { LuDownloadCloud, LuFolderSync, LuUploadCloud } from "react-icons/lu"

export const dynamic = "force-static"

export default function Page() {

    const [syncModalOpen, setSyncModalOpen] = React.useState(false)

    const { data: trackedMediaItems } = useSyncGetTrackedMediaItems()

    const { mutate: syncLocal, isPending: isSyncingLocal } = useSyncLocalData()

    const { mutate: syncAnilist, isPending: isSyncingAnilist } = useSyncAnilistData()

    const trackedAnimeItems = React.useMemo(() => {
        return trackedMediaItems?.filter(n => n.type === "anime" && !!n.animeEntry?.media) ?? []
    }, [trackedMediaItems])

    const trackedMangaItems = React.useMemo(() => {
        return trackedMediaItems?.filter(n => n.type === "manga" && !!n.mangaEntry?.media) ?? []
    }, [trackedMediaItems])

    const [queueState, setQueueState] = React.useState<Sync_QueueState | null>(null)
    useWebsocketMessageListener<Sync_QueueState>({
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


    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative space-y-8"
        >
            <div className="flex justify-between">
                <div>
                    <h2 className="text-center lg:text-left">Sync for offline</h2>
                    <p className="text-[--muted]">
                        View your tracked media for offline syncing.
                    </p>
                </div>

                <div>
                    <Modal
                        title="Sync"
                        open={syncModalOpen}
                        onOpenChange={v => {
                            if (isSyncingLocal) return
                            return setSyncModalOpen(v)
                        }}
                        trigger={<Button
                            intent="white"
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
                                leftIcon={<LuDownloadCloud className="text-2xl" />}
                                loading={isSyncingLocal}
                                disabled={isSyncingAnilist}
                                onClick={handleSyncLocal}
                            >
                                Update local data
                            </Button>
                            <p className="text-sm">
                                Download the latest data from AniList to your local collection.
                                This will overwrite your local changes.
                                This is done automatically every 30 minutes.
                            </p>
                            <Separator />
                            <Button
                                intent="gray-outline"
                                rounded
                                className="w-full"
                                leftIcon={<LuUploadCloud className="text-2xl" />}
                                disabled={isSyncingLocal}
                                loading={isSyncingAnilist}
                                onClick={handleSyncAnilist}
                            >
                                Update AniList data
                            </Button>
                            <p>
                                Update your AniList data with the latest data from your local collection.
                                This should be done manually.
                            </p>
                        </div>
                    </Modal>
                </div>
            </div>

            {/*{(queueState && (Object.keys(queueState.animeTasks!).length > 0 || Object.keys(queueState.mangaTasks!).length > 0)) &&*/}
            {/*    <div className="border rounded-md p-2">*/}
            {/*        <p className="flex items-center gap-1">*/}
            {/*            <Spinner className="size-6" />*/}
            {/*            <span>Syncing in progress</span>*/}
            {/*        </p>*/}
            {/*    </div>}*/}

            {(!trackedAnimeItems?.length && !trackedMangaItems?.length) && <LuffyError
                title="No tracked media"
            />}

            {!!trackedAnimeItems?.length && <div className="space-y-4">
                <h3>Tracked anime</h3>
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
                <h3>Tracked manga</h3>
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
