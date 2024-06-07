"use client"
import { Anime_LibraryCollectionList, Anime_LocalFile, Anime_UnknownGroup } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { __bulkAction_modalAtomIsOpen } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { PlayRandomEpisodeButton } from "@/app/(main)/(library)/_containers/play-random-episode-button"
import { __playlists_modalOpenAtom } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { __unknownMedia_drawerIsOpen } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { __unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Tooltip } from "@/components/ui/tooltip"
import { useSetAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { BiCollection, BiDotsVerticalRounded, BiFolder } from "react-icons/bi"
import { FiPlayCircle, FiSearch } from "react-icons/fi"
import { IoLibrarySharp } from "react-icons/io5"
import { PiClockCounterClockwiseFill } from "react-icons/pi"

export type LibraryToolbarProps = {
    collectionList: Anime_LibraryCollectionList[]
    ignoredLocalFiles: Anime_LocalFile[]
    unmatchedLocalFiles: Anime_LocalFile[]
    unknownGroups: Anime_UnknownGroup[]
    isLoading: boolean
}

export function LibraryToolbar(props: LibraryToolbarProps) {

    const { collectionList, ignoredLocalFiles, unmatchedLocalFiles, unknownGroups } = props

    const setBulkActionIsOpen = useSetAtom(__bulkAction_modalAtomIsOpen)

    const status = useServerStatus()
    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)
    const setUnmatchedFileManagerOpen = useSetAtom(__unmatchedFileManagerIsOpen)
    const setUnknownMediaManagerOpen = useSetAtom(__unknownMedia_drawerIsOpen)
    const setPlaylistsModalOpen = useSetAtom(__playlists_modalOpenAtom)

    const hasScanned = collectionList.some(n => !!n.entries?.length)

    const { mutate: openInExplorer } = useOpenInExplorer()

    return (
        <div className="flex flex-wrap w-full justify-end gap-2 p-4 relative z-[4]">
            <div className="flex flex-1"></div>
            {(!!status?.settings?.library?.libraryPath && hasScanned) && (
                <>
                    <Tooltip
                        trigger={<IconButton
                            intent={"white-subtle"}
                            icon={<FiPlayCircle className="text-2xl" />}
                            onClick={() => setPlaylistsModalOpen(true)}
                        />}
                    >Playlists</Tooltip>
                    <PlayRandomEpisodeButton />
                    <Button
                        intent={hasScanned ? "primary-subtle" : "primary"}
                        leftIcon={<FiSearch className="text-xl" />}
                        onClick={() => setScannerModalOpen(true)}
                    >
                        {hasScanned ? "Refresh entries" : "Scan your library"}
                    </Button>
                </>
            )}
            {(unmatchedLocalFiles.length > 0) && <Button
                intent="alert"
                leftIcon={<IoLibrarySharp />}
                className=""
                onClick={() => setUnmatchedFileManagerOpen(true)}
            >
                Resolve unmatched ({unmatchedLocalFiles.length})
            </Button>}
            {(unknownGroups.length > 0) && <Button
                intent="warning"
                leftIcon={<IoLibrarySharp />}
                className=""
                onClick={() => setUnknownMediaManagerOpen(true)}
            >
                Resolve hidden media ({unknownGroups.length})
            </Button>}
            <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded />} intent="gray-basic" />}>
                {/*<DropdownMenuItem*/}
                {/*    disabled={!hasScanned}*/}
                {/*    className={cn("cursor-pointer", { "!text-[--muted]": !status?.settings?.library?.libraryPath })}*/}
                {/*    onClick={() => {*/}

                {/*    }}*/}
                {/*>*/}
                {/*    <FaSearch />*/}
                {/*    <span>Find</span>*/}
                {/*</DropdownMenuItem>*/}

                <DropdownMenuItem
                    disabled={!status?.settings?.library?.libraryPath}
                    className={cn("cursor-pointer", { "!text-[--muted]": !status?.settings?.library?.libraryPath })}
                    onClick={() => {
                        openInExplorer({ path: status?.settings?.library?.libraryPath ?? "" })
                    }}
                >
                    <BiFolder />
                    <span>Open folder</span>
                </DropdownMenuItem>

                {/*<DropdownMenu.Item*/}
                {/*    onClick={() => {*/}
                {/*    }}*/}
                {/*    disabled={ignoredLocalFiles.length === 0}*/}
                {/*    className={cn({ "!text-[--muted]": ignoredLocalFiles.length === 0 })}*/}
                {/*>*/}
                {/*    <GoDiffIgnored/>*/}
                {/*    <span>Manage ignored files</span>*/}
                {/*</DropdownMenu.Item>*/}

                <DropdownMenuItem
                    onClick={() => setBulkActionIsOpen(true)}
                    disabled={!hasScanned}
                    className={cn({ "!text-[--muted]": !hasScanned })}
                >
                    <BiCollection />
                    <span>Bulk actions</span>
                </DropdownMenuItem>

                <Link href="/scan-summaries">
                    <DropdownMenuItem

                        className={cn({ "!text-[--muted]": !hasScanned })}
                    >
                        <PiClockCounterClockwiseFill />
                        <span>Scan summaries</span>
                    </DropdownMenuItem>
                </Link>
            </DropdownMenu>

        </div>
    )

}
