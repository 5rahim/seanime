"use client"
import { Button } from "@/components/ui/button"
import { AnimeCollectionQuery } from "@/lib/anilist/gql/graphql"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"
import { IoReload } from "react-icons/io5"
import { toast } from "sonner"

interface RefreshAnilistButtonProps {
    children?: React.ReactNode
}

export const RefreshAnilistButton: React.FC<RefreshAnilistButtonProps> = (props) => {

    const { children, ...rest } = props

    const qc = useQueryClient()

    /**
     * @description
     * - Asks the server to fetch an up-to-date version of the user's AniList collection.
     * - When the request succeeds, we refetch queries related to the AniList collection.
     */
    const { mutate, isPending } = useSeaMutation<AnimeCollectionQuery>({ // Ignore return data
        endpoint: SeaEndpoints.ANILIST_COLLECTION, // POST bypasses the cache
        mutationKey: ["refresh-anilist-collection"],
        onSuccess: async () => {
            // Refetch library collection
            toast.success("AniList is up-to-date")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-anilist-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
        },
    })

    return (
        <>
            <Button
                onClick={() => mutate()}
                intent="warning-subtle"
                rightIcon={<IoReload />}
                loading={isPending}
                leftIcon={<svg
                    xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="24" height="24"
                    viewBox="0 0 24 24" role="img"
                >
                    <path
                        d="M6.361 2.943 0 21.056h4.942l1.077-3.133H11.4l1.052 3.133H22.9c.71 0 1.1-.392 1.1-1.101V17.53c0-.71-.39-1.101-1.1-1.101h-6.483V4.045c0-.71-.392-1.102-1.101-1.102h-2.422c-.71 0-1.101.392-1.101 1.102v1.064l-.758-2.166zm2.324 5.948 1.688 5.018H7.144z"
                    />
                </svg>}
                className={""}
            >
                Refresh AniList
            </Button>
        </>
    )

}
