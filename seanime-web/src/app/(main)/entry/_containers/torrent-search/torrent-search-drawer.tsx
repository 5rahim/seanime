import { Anime_Entry, Anime_EntryDownloadEpisode } from "@/api/generated/types"
import { useTorrentSearchSelection } from "@/app/(main)/entry/_containers/torrent-search/_lib/handle-torrent-selection"
import { TorrentConfirmationContinueButton } from "@/app/(main)/entry/_containers/torrent-search/torrent-confirmation-modal"
import { TorrentSearchContainer } from "@/app/(main)/entry/_containers/torrent-search/torrent-search-container"
import { GlowingEffect } from "@/components/shared/glowing-effect"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Modal } from "@/components/ui/modal"
import { useThemeSettings } from "@/lib/theme/hooks"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { usePathname, useRouter, useSearchParams } from "next/navigation"
import React, { useEffect } from "react"

export const __torrentSearch_selectionAtom = atom<TorrentSelectionType | undefined>(undefined)
export const __torrentSearch_selectionEpisodeAtom = atom<number | undefined>(undefined)

export type TorrentSelectionType =
    "torrentstream-select"
    | "torrentstream-select-file"
    | "debridstream-select"
    | "debridstream-select-file"
    | "download"

export function TorrentSearchDrawer(props: { entry: Anime_Entry }) {

    const { entry } = props
    const ts = useThemeSettings()

    const [selectionType, setSelection] = useAtom(__torrentSearch_selectionAtom)
    const searchParams = useSearchParams()
    const router = useRouter()
    const pathname = usePathname()
    const mId = searchParams.get("id")
    const downloadParam = searchParams.get("download")

    useEffect(() => {
        if (!!downloadParam) {
            setSelection("download")
            router.replace(pathname + `?id=${mId}`)
        }
    }, [downloadParam])

    const { onTorrentValidated } = useTorrentSearchSelection({ entry, type: selectionType })

    return (
        <Modal
            open={selectionType !== undefined}
            onOpenChange={() => setSelection(undefined)}
            // size="xl"
            contentClass="max-w-5xl bg-gray-950 bg-opacity-75 firefox:bg-opacity-100 sm:rounded-xl"
            title={`${entry?.media?.title?.userPreferred || "Anime"}`}
            titleClass="max-w-[500px] text-ellipsis truncate"
            data-torrent-search-drawer
            overlayClass="bg-gray-950/70 backdrop-blur-sm"
        >

            <GlowingEffect
                spread={40}
                // blur={1}
                glow={true}
                disabled={false}
                proximity={100}
                inactiveZone={0.01}
                className="opacity-30"
            />

            {/*{(ts.enableMediaPageBlurredBackground) && <div*/}
            {/*    data-media-page-header-blurred-background*/}
            {/*    className={cn(*/}
            {/*        "absolute top-0 left-0 w-full h-full z-[0] bg-[--background] rounded-xl overflow-hidden",*/}
            {/*        "opacity-20",*/}
            {/*    )}*/}
            {/*>*/}
            {/*    <Image*/}
            {/*        data-media-page-header-blurred-background-image*/}
            {/*        src={getImageUrl(entry.media?.bannerImage || "")}*/}
            {/*        alt={""}*/}
            {/*        fill*/}
            {/*        quality={100}*/}
            {/*        sizes="20rem"*/}
            {/*        className={cn(*/}
            {/*            "object-cover object-bottom transition opacity-10",*/}
            {/*            ts.mediaPageBannerSize === ThemeMediaPageBannerSize.Small && "object-left",*/}
            {/*        )}*/}
            {/*    />*/}

            {/*    <div*/}
            {/*        data-media-page-header-blurred-background-blur*/}
            {/*        className="absolute top-0 w-full h-full backdrop-blur-2xl z-[2]"*/}
            {/*    ></div>*/}
            {/*</div>}*/}

            {/*{entry?.media?.bannerImage && <div*/}
            {/*    data-torrent-search-drawer-banner-image-container*/}
            {/*    className="Sea-TorrentSearchDrawer__bannerImage h-36 w-full flex-none object-cover object-center overflow-hidden rounded-t-xl absolute left-0 top-0 z-[-1]"*/}
            {/*>*/}
            {/*    <Image*/}
            {/*        data-torrent-search-drawer-banner-image*/}
            {/*        src={getImageUrl(entry?.media?.bannerImage!)}*/}
            {/*        alt="banner"*/}
            {/*        fill*/}
            {/*        quality={80}*/}
            {/*        priority*/}
            {/*        sizes="20rem"*/}
            {/*        className="object-cover object-center opacity-10"*/}
            {/*    />*/}
            {/*    <div*/}
            {/*        data-torrent-search-drawer-banner-image-bottom-gradient*/}
            {/*        className="Sea-TorrentSearchDrawer__bannerImage-bottomGradient z-[5] absolute bottom-0 w-full h-[70%] bg-gradient-to-t from-[--background] to-transparent"*/}
            {/*    />*/}
            {/*</div>}*/}

            <AppLayoutStack className="relative z-[1]" data-torrent-search-drawer-content>
                {selectionType === "download" && <EpisodeList episodes={entry.downloadInfo?.episodesToDownload} />}
                {!!selectionType && <TorrentSearchContainer type={selectionType} entry={entry} />}
            </AppLayoutStack>

            <TorrentConfirmationContinueButton type={selectionType || "download"} onTorrentValidated={onTorrentValidated} />
        </Modal>
    )

}


function EpisodeList({ episodes }: { episodes: Anime_EntryDownloadEpisode[] | undefined }) {

    if (!episodes || !episodes.length) return null

    const missingEpisodes = episodes.sort((a, b) => a.episodeNumber - b.episodeNumber)

    return (
        <div className="space-y-2" data-torrent-search-drawer-episode-list>
            <p><span className="font-semibold">Missing episode{missingEpisodes.length > 1 ? "s" : ""}</span>: {missingEpisodes.slice(0, 5)
                .map(n => n.episodeNumber)
                .join(", ")}{missingEpisodes.length > 5
                ? `, ..., ${missingEpisodes[missingEpisodes.length - 1].episodeNumber}`
                : ""}
            </p>
        </div>
    )

}
