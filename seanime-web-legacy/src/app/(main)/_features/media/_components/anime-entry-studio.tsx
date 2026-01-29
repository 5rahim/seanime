import { useGetAnilistStudioDetails } from "@/api/hooks/anilist.hooks"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { Badge } from "@/components/ui/badge"
import { Drawer } from "@/components/ui/drawer"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import React from "react"

type AnimeEntryStudioProps = {
    studios?: { nodes?: Array<{ name: string, id: number } | null> | null } | null | undefined
}

export function AnimeEntryStudio(props: AnimeEntryStudioProps) {

    const {
        studios,
        ...rest
    } = props

    if (!studios?.nodes) return null

    return (
        <AnimeEntryStudioDetailsModal studios={studios}>
            <Badge
                size="lg"
                intent="gray"
                className="rounded-full px-0 border-transparent bg-transparent cursor-pointer transition-all hover:bg-transparent hover:text-white hover:-translate-y-0.5"
                data-anime-entry-studio-badge
            >
                {studios?.nodes?.[0]?.name}
            </Badge>
        </AnimeEntryStudioDetailsModal>
    )
}

function AnimeEntryStudioDetailsModal(props: AnimeEntryStudioProps & { children: React.ReactElement }) {

    const {
        studios,
        children,
        ...rest
    } = props

    const studio = studios?.nodes?.[0]

    if (!studio?.name) return null

    return (
        <>
            <Drawer
                trigger={children}
                size="xl"
                title={studio.name}
                data-anime-entry-studio-details-modal
            >
                <div data-anime-entry-studio-details-modal-top-padding className="py-4"></div>
                <AnimeEntryStudioDetailsModalContent studios={studios} />
            </Drawer>
        </>
    )
}

function AnimeEntryStudioDetailsModalContent(props: AnimeEntryStudioProps) {

    const {
        studios,
        ...rest
    } = props

    const { data, isLoading } = useGetAnilistStudioDetails(studios?.nodes?.[0]?.id!)

    if (isLoading) return <LoadingSpinner />

    return (
        <div
            data-anime-entry-studio-details-modal-content
            className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-4 xl:grid-cols-4 2xl:grid-cols-4 gap-4"
        >
            {data?.Studio?.media?.nodes?.map(media => {
                return <div key={media?.id!} className="col-span-1" data-anime-entry-studio-details-modal-content-media-entry-card>
                    <MediaEntryCard
                        media={media}
                        type="anime"
                        showLibraryBadge
                    />
                </div>
            })}
        </div>
    )
}
