"use client"
import { Anime_LibraryCollectionList, Anime_LocalFile, Anime_UnknownGroup } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { __bulkAction_modalAtomIsOpen } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { __ignoredFileManagerIsOpen } from "@/app/(main)/(library)/_containers/ignored-file-manager"
import { PlayRandomEpisodeButton } from "@/app/(main)/(library)/_containers/play-random-episode-button"
import { __playlists_modalOpenAtom } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { __unknownMedia_drawerIsOpen } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { __unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { __library_viewAtom } from "@/app/(main)/(library)/_lib/library-view.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Tooltip } from "@/components/ui/tooltip"
import { ThemeLibraryScreenBannerType, useThemeSettings } from "@/lib/theme/hooks"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { BiCollection, BiDotsVerticalRounded, BiFolder } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"
import { IoLibrary, IoLibrarySharp } from "react-icons/io5"
import { MdOutlineVideoLibrary } from "react-icons/md"
import { PiClockCounterClockwiseFill } from "react-icons/pi"
import { TbFileSad, TbReload } from "react-icons/tb"
import { PluginAnimeLibraryDropdownItems } from "../../_features/plugin/actions/plugin-actions"

export type LibraryToolbarProps = {
    collectionList: Anime_LibraryCollectionList[]
    ignoredLocalFiles: Anime_LocalFile[]
    unmatchedLocalFiles: Anime_LocalFile[]
    unknownGroups: Anime_UnknownGroup[]
    isLoading: boolean
    hasEntries: boolean
    isStreamingOnly: boolean
    isNakamaLibrary: boolean
}

export function LibraryToolbar(props: LibraryToolbarProps) {

    const {
        collectionList,
        ignoredLocalFiles,
        unmatchedLocalFiles,
        unknownGroups,
        hasEntries,
        isStreamingOnly,
        isNakamaLibrary,
    } = props

    const ts = useThemeSettings()
    const setBulkActionIsOpen = useSetAtom(__bulkAction_modalAtomIsOpen)

    const status = useServerStatus()
    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)
    const setUnmatchedFileManagerOpen = useSetAtom(__unmatchedFileManagerIsOpen)
    const setIgnoredFileManagerOpen = useSetAtom(__ignoredFileManagerIsOpen)
    const setUnknownMediaManagerOpen = useSetAtom(__unknownMedia_drawerIsOpen)
    const setPlaylistsModalOpen = useSetAtom(__playlists_modalOpenAtom)

    const [libraryView, setLibraryView] = useAtom(__library_viewAtom)

    const { mutate: openInExplorer } = useOpenInExplorer()

    const hasLibraryPath = !!status?.settings?.library?.libraryPath

    return (
        <>
            {(ts.libraryScreenBannerType === ThemeLibraryScreenBannerType.Dynamic && hasEntries) && <div
                className={cn(
                    "h-28",
                    ts.hideTopNavbar && "h-40",
                )}
                data-library-toolbar-top-padding
            ></div>}
            <div className="flex flex-wrap w-full justify-end gap-2 p-4 relative z-[10]" data-library-toolbar-container>
                <div className="flex flex-1" data-library-toolbar-spacer></div>
                {(hasEntries) && (
                    <>
                        <Tooltip
                            trigger={<IconButton
                                data-library-toolbar-switch-view-button
                                intent={libraryView === "base" ? "white-subtle" : "white"}
                                icon={<IoLibrary className="text-2xl" />}
                                onClick={() => setLibraryView(p => p === "detailed" ? "base" : "detailed")}
                            />}
                        >
                            Switch view
                        </Tooltip>

                        {!(isStreamingOnly || isNakamaLibrary) && <Tooltip
                            trigger={<IconButton
                                data-library-toolbar-playlists-button
                                intent={"white-subtle"}
                                icon={<MdOutlineVideoLibrary className="text-2xl" />}
                                onClick={() => setPlaylistsModalOpen(true)}
                            />}
                        >Playlists</Tooltip>}

                        {!(isStreamingOnly || isNakamaLibrary) && <PlayRandomEpisodeButton />}

                        {!(isStreamingOnly || isNakamaLibrary) && hasLibraryPath && <Button
                            data-library-toolbar-scan-button
                            intent={hasEntries ? "primary-subtle" : "primary"}
                            leftIcon={hasEntries ? <TbReload className="text-xl" /> : <FiSearch className="text-xl" />}
                            onClick={() => setScannerModalOpen(true)}
                            hideTextOnSmallScreen
                        >
                            {hasEntries ? "Refresh library" : "Scan your library"}
                        </Button>}
                    </>
                )}
                {(unmatchedLocalFiles.length > 0) && <Button
                    data-library-toolbar-unmatched-button
                    intent="alert"
                    leftIcon={<IoLibrarySharp />}
                    className="animate-bounce"
                    onClick={() => setUnmatchedFileManagerOpen(true)}
                >
                    Resolve unmatched ({unmatchedLocalFiles.length})
                </Button>}
                {(unknownGroups.length > 0) && <Button
                    data-library-toolbar-unknown-button
                    intent="warning"
                    leftIcon={<IoLibrarySharp />}
                    className="animate-bounce"
                    onClick={() => setUnknownMediaManagerOpen(true)}
                >
                    Resolve hidden media ({unknownGroups.length})
                </Button>}

                {(!isStreamingOnly && !isNakamaLibrary && hasLibraryPath) &&
                    <DropdownMenu
                        trigger={<IconButton
                            data-library-toolbar-dropdown-menu-trigger
                            icon={<BiDotsVerticalRounded />} intent="gray-basic"
                        />}
                    >

                        <DropdownMenuItem
                            data-library-toolbar-open-directory-button
                            disabled={!hasLibraryPath}
                            className={cn("cursor-pointer", { "!text-[--muted]": !hasLibraryPath })}
                            onClick={() => {
                                openInExplorer({ path: status?.settings?.library?.libraryPath ?? "" })
                            }}
                        >
                            <BiFolder />
                            <span>Open directory</span>
                        </DropdownMenuItem>

                        <DropdownMenuItem
                            data-library-toolbar-bulk-actions-button
                            onClick={() => setBulkActionIsOpen(true)}
                            disabled={!hasEntries}
                            className={cn({ "!text-[--muted]": !hasEntries })}
                        >
                            <BiCollection />
                            <span>Bulk actions</span>
                        </DropdownMenuItem>

                        <DropdownMenuItem
                            data-library-toolbar-ignored-files-button
                            onClick={() => setIgnoredFileManagerOpen(true)}
                            // disabled={!hasEntries}
                            className={cn({ "!text-[--muted]": !hasEntries })}
                        >
                            <TbFileSad />
                            <span>Ignored files</span>
                        </DropdownMenuItem>

                        <SeaLink href="/scan-summaries">
                            <DropdownMenuItem
                                data-library-toolbar-scan-summaries-button
                            // className={cn({ "!text-[--muted]": !hasEntries })}
                            >
                                <PiClockCounterClockwiseFill />
                                <span>Scan summaries</span>
                            </DropdownMenuItem>
                        </SeaLink>

                        <PluginAnimeLibraryDropdownItems />
                    </DropdownMenu>}

            </div>
        </>
    )

}
