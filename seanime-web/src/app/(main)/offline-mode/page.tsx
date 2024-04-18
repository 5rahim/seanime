"use client"
import { SnapshotAnimeSelector } from "@/app/(main)/offline-mode/_components/snapshot-anime-selector"
import { useOfflineSnapshotEntry } from "@/app/(main)/offline-mode/_lib/offline-mode.hooks"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Button } from "@/components/ui/button"
import { Drawer } from "@/components/ui/drawer"
import { Separator } from "@/components/ui/separator"
import { format } from "date-fns"
import React from "react"
import { IoCloudOfflineOutline } from "react-icons/io5"
import { toast } from "sonner"

export const dynamic = "force-static"

export default function Page() {
    const [animeMediaIds, setAnimeMediaIds] = React.useState<number[]>([])
    const [open, setOpen] = React.useState(false)

    const {
        createOfflineSnapshot,
        snapshot,
        isLoading,
        isCreating,
        sync,
        isSyncing,
    } = useOfflineSnapshotEntry()

    return (
        <PageWrapper
            className="p-4 sm:p-8 pt-4 relative space-y-4"
        >
            <div>
                <h2 className="text-center lg:text-left">Offline Mode</h2>
            </div>

            {!isLoading && <>
                <div className="text-gray-300">
                    <p className="">
                        Offline mode allows you to use the most important features of Seanime without an internet connection.
                    </p>
                    <p className="">
                        Create a snapshot of your library then enable offline mode in your <code>config.toml</code> file.
                        Your progress will be synced once you're back online.
                    </p>
                    <p>
                        Refer to the documentation for more information.
                    </p>
                </div>

                {!!snapshot?.id && <ul className="text-lg">
                    <li className="flex gap-2 items-center">
                        <IoCloudOfflineOutline className="text-green-300 text-xl" /> Snapshot <span className="text-[--muted]">
                        ({format(snapshot.createdAt, "P HH:mm")})
                    </span>

                        <Button
                            rounded
                            intent="white"
                            size="sm"
                            disabled={!snapshot.used}
                            loading={isSyncing}
                            onClick={() => sync()}
                        >
                            Sync and delete
                        </Button>
                    </li>
                </ul>}

                <Drawer
                    open={open}
                    onOpenChange={v => setOpen(v)}
                    title="Select media"
                    trigger={<Button loading={isCreating} role="save" intent="success-outline" rounded>Create new snapshot</Button>}
                    size="xl"
                >
                    <div className="space-y-4 py-6">
                        <h3>Anime</h3>
                        <SnapshotAnimeSelector
                            animeMediaIds={animeMediaIds}
                            setAnimeMediaIds={setAnimeMediaIds}
                        />

                        <Separator />

                        <h3>Manga</h3>
                        <p className="text-[--muted]">
                            Manga entries will automatically be included based on downloaded chapters.
                        </p>

                        <Separator />

                        <div className="flex">
                            <Button
                                role="save"
                                intent="white"
                                loading={isCreating}
                                onClick={() => {
                                    if (animeMediaIds.length) {
                                        createOfflineSnapshot({ animeMediaIds })
                                        setOpen(false)
                                    } else {
                                        toast.error("Select at least one anime")
                                    }
                                }}
                            >Create snapshot</Button>
                        </div>
                    </div>
                </Drawer>
            </>}
        </PageWrapper>
    )
}
