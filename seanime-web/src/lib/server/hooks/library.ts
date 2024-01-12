import { libraryCollectionAtom } from "@/atoms/collection"
import { missingEpisodesAtom } from "@/atoms/missing-episodes"
import { AnimeCollectionQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation, useSeaQuery } from "@/lib/server/queries/utils"
import { LibraryCollection, LocalFile, LocalFileMetadata, MediaEntryEpisode } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { useAtom, useSetAtom } from "jotai/react"
import { useEffect, useMemo } from "react"
import toast from "react-hot-toast"

export type ScanLibraryProps = {
    enhanced: boolean,
    skipLockedFiles: boolean,
    skipIgnoredFiles: boolean
}

export function useScanLibrary({ onSuccess }: { onSuccess: () => void }) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], ScanLibraryProps>({
        endpoint: SeaEndpoints.SCAN_LIBRARY,
        mutationKey: ["scan-library"],
        onSuccess: async () => {
            toast.success("Library scanned")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
            onSuccess()
        },
    })

    return {
        scanLibrary: mutate,
        isScanning: isPending,
    }
}

//----------------------------------------------------------------------------------------------------------------------

export function useLibraryCollection() {

    const [prev, setLibraryCollectionAtom] = useAtom(libraryCollectionAtom)

    const { data, isLoading, refetch } = useSeaQuery<LibraryCollection>({
        endpoint: SeaEndpoints.LIBRARY_COLLECTION,
        queryKey: ["get-library-collection"],
        placeholderData: prev,
    })

    useEffect(() => {
        if (!!data) {
            setLibraryCollectionAtom(data)
        }
    }, [data])

    const sortedCollection = useMemo(() => {
        if (!data) return []
        return [
            data.lists.find(n => n.type === "current"),
            data.lists.find(n => n.type === "paused"),
            data.lists.find(n => n.type === "planned"),
            data.lists.find(n => n.type === "completed"),
            data.lists.find(n => n.type === "dropped"),
        ].filter(Boolean)
    }, [data])

    return {
        libraryCollectionList: sortedCollection,
        continueWatchingList: data?.continueWatchingList ?? [],
        isLoading: isLoading,
        unmatchedLocalFiles: data?.unmatchedLocalFiles ?? [],
        ignoredLocalFiles: data?.ignoredLocalFiles ?? [],
        unmatchedGroups: data?.unmatchedGroups ?? [],
        unknownGroups: data?.unknownGroups ?? [],
    }

}

//----------------------------------------------------------------------------------------------------------------------

export function useMissingEpisodes() {

    const setAtom = useSetAtom(missingEpisodesAtom)

    const { data, isLoading, status } = useSeaQuery<MediaEntryEpisode[]>({
        endpoint: SeaEndpoints.MISSING_EPISODES,
        queryKey: ["get-missing-episodes"],
    })

    useEffect(() => {
        if (status === "success") {
            setAtom(data ?? [])
        }
    }, [data])

    return {
        missingEpisodes: data ?? [],
        isLoading: isLoading,
    }

}

//----------------------------------------------------------------------------------------------------------------------

export type MediaEntryBulkAction = "unmatch" | "toggle-lock"

export function useMediaEntryBulkAction(mId?: number) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], { mediaId: number, action: MediaEntryBulkAction }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_BULK_ACTION,
        mutationKey: ["media-entry-bulk-action"],
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            if (mId) {
                await qc.refetchQueries({ queryKey: ["get-media-entry", mId] })
            }
        },
    })

    return {
        toggleLock: (mId: number) => mutate({
            mediaId: mId,
            action: "toggle-lock",
        }),
        unmatchAll: (mId: number) => mutate({
            mediaId: mId,
            action: "unmatch",
        }, {
            onSuccess: () => {
                toast.success("Files unmatched")
            },
        }),
        isPending,
    }

}


//----------------------------------------------------------------------------------------------------------------------

export type LocalFileBulkAction = "lock" | "unlock"

export function useLocalFileBulkAction() {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], { action: LocalFileBulkAction }>({
        endpoint: SeaEndpoints.LOCAL_FILES,
        mutationKey: ["local-file-bulk-action"],
        method: "post",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
        },
    })

    return {
        lockFiles: () => mutate({
            action: "lock",
        }, {
            onSuccess: () => {
                toast.success("Files locked")
            },
        }),
        unlockFiles: () => mutate({
            action: "unlock",
        }, {
            onSuccess: () => {
                toast.success("Files unlocked")
            },
        }),
        isPending,
    }

}

export function useRemoveEmptyDirectories() {

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<null>({
        endpoint: SeaEndpoints.EMPTY_DIRECTORIES,
        mutationKey: ["remove-empty-directories"],
        method: "delete",
        onSuccess: async () => {
            toast.success("Empty directories removed")
        },
    })

    return {
        removeEmptyDirectories: () => mutate(),
        isPending,
    }

}

//----------------------------------------------------------------------------------------------------------------------

type UpdateLocalFileVariables = {
    path: string
    metadata?: LocalFileMetadata
    locked: boolean
    ignored: boolean
    mediaId: number
}

export function getDefaultLocalFileVariables(lf: LocalFile): UpdateLocalFileVariables {
    return {
        path: lf.path,
        metadata: lf.metadata,
        locked: lf.locked,
        ignored: lf.ignored,
        mediaId: lf.mediaId,
    }
}

export function useUpdateLocalFile(mId?: number) {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<LocalFile[], UpdateLocalFileVariables>({
        endpoint: SeaEndpoints.LOCAL_FILE,
        mutationKey: ["patch-local-file"],
        method: "patch",
        onSuccess: async () => {
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            if (mId) {
                await qc.refetchQueries({ queryKey: ["get-media-entry", mId] })
            }
        },
    })

    return {
        updateLocalFile: (lf: LocalFile, variables: Partial<UpdateLocalFileVariables>, onSuccess?: () => void) => {
            mutate({
                ...getDefaultLocalFileVariables(lf),
                ...variables,
            }, {
                onSuccess: () => {
                    onSuccess && onSuccess()
                },
            })
        },
        isPending,
    }

}


//------------------------------------------------------------------------------------------------------------------------------

export function useAddUnknownMedia() {

    const qc = useQueryClient()

    // Return data is ignored
    const { mutate, isPending } = useSeaMutation<AnimeCollectionQuery, { mediaIds: number[] }>({
        endpoint: SeaEndpoints.MEDIA_ENTRY_UNKNOWN_MEDIA,
        mutationKey: ["add-unknown-media"],
        onSuccess: async () => {
            // Refetch library collection
            toast.success("AniList is up-to-date")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
        },
    })

    return {
        addUnknownMedia: mutate,
        isPending,
    }

}
