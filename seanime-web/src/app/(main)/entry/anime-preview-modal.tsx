import { useGetAnilistAnimeDetails } from "@/api/hooks/anilist.hooks"
import { useGetAnimeEntry } from "@/api/hooks/anime_entries.hooks"
import { TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE } from "@/app/(main)/_features/custom-ui/styles"
import { MediaPageHeaderEntryDetails } from "@/app/(main)/_features/media/_components/media-page-header-components"
import { RelationsRecommendationsSection } from "@/app/(main)/entry/_components/relations-recommendations-section"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { getImageUrl } from "@/lib/server/assets"
import { ThemeMediaPageBannerSize, useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import Image from "next/image"
import React from "react"
import { BiX } from "react-icons/bi"

// unused

type AnimePreviewModalProps = {
    children?: React.ReactNode
}

const __anime_previewMediaIdAtom = atom<number | undefined>(undefined)
const __anime_previewPreviewMediaIdAtom = atom<number | undefined>(undefined)

export function useAnimePreviewModal() {
    const setMediaId = useSetAtom(__anime_previewMediaIdAtom)

    return {
        setPreviewModalMediaId: setMediaId,
    }
}

export function AnimePreviewModal(props: AnimePreviewModalProps) {

    const {
        children,
        ...rest
    } = props

    const ts = useThemeSettings()
    const [mediaId, setMediaId] = useAtom(__anime_previewMediaIdAtom)

    const { data: entry, isLoading: entryLoading } = useGetAnimeEntry(mediaId)
    const { data: details, isLoading: detailsLoading } = useGetAnilistAnimeDetails(mediaId)

    const containerRef = React.useRef(null)

    const media = entry?.media

    const backgroundImage = media?.bannerImage
    const bannerImage = media?.bannerImage || media?.coverImage?.extraLarge

    return (
        <>
            <Modal
                open={!!mediaId}
                onOpenChange={v => setMediaId(v ? mediaId : undefined)}
                contentClass="max-w-7xl relative"
                closeButton={<div className="z-[8] absolute right-1 top-1 lg:-right-5 lg:-top-3">
                    <IconButton intent="alert" className="rounded-full" icon={<BiX />} />
                </div>}
                {...rest}
            >

                <div
                    className={cn(
                        "absolute z-[0] opacity-30",
                        "w-full flex-none object-cover object-center z-[3] bg-[--background] h-[12rem]",
                        ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small ? "lg:h-[23rem]" : "h-[12rem] lg:h-[22rem] 2xl:h-[30rem]",
                        !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                    )}
                >

                    {/*BOTTOM OVERFLOW FADE*/}
                    <div
                        className={cn(
                            "w-full z-[2] absolute bottom-[-5rem] h-[5rem] bg-gradient-to-b from-[--background] via-transparent via-100% to-transparent",
                            !ts.disableSidebarTransparency && TRANSPARENT_SIDEBAR_BANNER_IMG_STYLE,
                        )}
                    />

                    <div
                        className={cn(
                            "absolute top-0 left-0 w-full h-full",
                        )}
                    >
                        {(!!bannerImage) && <Image
                            src={getImageUrl(bannerImage || "")}
                            alt="banner image"
                            fill
                            quality={100}
                            priority
                            sizes="100vw"
                            className={cn(
                                "object-cover object-center z-[1]",
                            )}
                        />}

                        {/*LEFT MASK*/}
                        <div
                            className={cn(
                                "hidden lg:block max-w-[60rem] xl:max-w-[100rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background]  transition-opacity to-transparent",
                                "opacity-85 duration-1000",
                                // y > 300 && "opacity-70",
                            )}
                        />
                        <div
                            className={cn(
                                "hidden lg:block max-w-[60rem] xl:max-w-[80rem] w-full z-[2] h-full absolute left-0 bg-gradient-to-r from-[--background] from-25% transition-opacity to-transparent",
                                "opacity-50 duration-500",
                            )}
                        />
                    </div>

                    {/*BOTTOM FADE*/}
                    <div
                        className={cn(
                            "w-full z-[3] absolute bottom-0 h-[50%] bg-gradient-to-t from-[--background] via-transparent via-100% to-transparent",
                        )}
                    />

                    <div
                        className={cn(
                            "absolute h-full w-full block lg:hidden bg-[--background] opacity-70 z-[2]",
                        )}
                    />

                </div>

                {entryLoading && <LoadingSpinner />}

                {(!entryLoading && entry) && <>

                    <div className="z-[5] relative">
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

                        {detailsLoading ? <LoadingSpinner /> : <div className="space-y-6 pt-6" ref={containerRef}>
                            <RelationsRecommendationsSection entry={entry} details={details} />
                        </div>}
                    </div>

                    {/*<div className="absolute top-0 left-0 w-full h-full z-[0] bg-[--background] rounded-xl">*/}
                    {/*    <Image*/}
                    {/*        src={media?.bannerImage || ""}*/}
                    {/*        alt={""}*/}
                    {/*        fill*/}
                    {/*        quality={100}*/}
                    {/*        sizes="20rem"*/}
                    {/*        className="object-cover object-center transition opacity-15"*/}
                    {/*    />*/}

                    {/*    <div*/}
                    {/*        className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2] "*/}
                    {/*    ></div>*/}
                    {/*</div>*/}

                </>}


            </Modal>
        </>
    )
}
