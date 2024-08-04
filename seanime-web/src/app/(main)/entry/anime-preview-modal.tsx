import { useGetAnilistAnimeDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { MediaPageHeaderEntryDetails } from "@/app/(main)/_features/media/_components/media-page-header-components"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_components/relations-recommendations-section"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import Image from "next/image"
import React from "react"

// unused

type AnimePreviewModalProps = {
    children?: React.ReactNode
}

const __anime_previewMediaIdAtom = atom<number | undefined>(undefined)

export function AnimePreviewModal(props: AnimePreviewModalProps) {

    const {
        children,
        ...rest
    } = props

    const [mediaId, setMediaId] = useAtom(__anime_previewMediaIdAtom)

    const { data: entry, isLoading: entryLoading } = useGetAnimeEntry(mediaId)
    const { data: details, isLoading: detailsLoading } = useGetAnilistAnimeDetails(mediaId)

    const media = entry?.media

    return (
        <>
            <Modal
                open={!!mediaId}
                onOpenChange={v => setMediaId(v ? mediaId : undefined)}
                contentClass="max-w-7xl relative"
                {...rest}
            >

                {entryLoading && <LoadingSpinner />}

                {(!entryLoading && entry) && <>

                    <div className="z-[1] relative">
                        <MediaPageHeaderEntryDetails
                            coverImage={entry.media?.coverImage?.extraLarge || entry.media?.coverImage?.large}
                            title={entry.media?.title?.userPreferred}
                            color={entry.media?.coverImage?.color}
                            englishTitle={entry.media?.title?.english}
                            romajiTitle={entry.media?.title?.romaji}
                            startDate={entry.media?.startDate}
                            season={entry.media?.season}
                            progressTotal={entry.media?.episodes}
                            status={entry.media?.status}
                            description={entry.media?.description}
                            listData={entry.listData}
                            media={entry.media!}
                            type="anime"
                        />

                        <RelationsRecommendationsSection entry={entry} details={details} />
                    </div>

                    <div className="absolute top-0 left-0 w-full h-full z-[0] bg-[--background] rounded-xl">
                        <Image
                            src={media?.bannerImage || ""}
                            alt={""}
                            fill
                            quality={100}
                            sizes="20rem"
                            className="object-cover object-center transition opacity-15"
                        />

                        <div
                            className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "
                        ></div>
                    </div>

                </>}


            </Modal>
        </>
    )
}
