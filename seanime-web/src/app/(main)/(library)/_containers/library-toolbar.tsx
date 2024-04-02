"use client"
import { bulkActionModalAtomIsOpen } from "@/app/(main)/(library)/_containers/bulk-actions/bulk-action-modal"
import { __playlists_modalOpenAtom } from "@/app/(main)/(library)/_containers/playlists/playlists-modal"
import { _scannerModalIsOpen } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { _unknownMediaManagerIsOpen } from "@/app/(main)/(library)/_containers/unknown-media/unknown-media-manager"
import { _unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_containers/unmatched-files/unmatched-file-manager"
import { LibraryCollectionList, LocalFile, UnknownGroup } from "@/app/(main)/(library)/_lib/anime-library.types"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useOpenInExplorer } from "@/lib/server/hooks"
import { useAtomValue, useSetAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { BiCollection, BiDotsVerticalRounded, BiFolder } from "react-icons/bi"
import { FiPlayCircle, FiSearch } from "react-icons/fi"
import { IoLibrarySharp } from "react-icons/io5"
import { PiClockCounterClockwiseFill } from "react-icons/pi"

export type LibraryToolbarProps = {
    collectionList: LibraryCollectionList[]
    ignoredLocalFiles: LocalFile[]
    unmatchedLocalFiles: LocalFile[]
    unknownGroups: UnknownGroup[]
    isLoading: boolean
}

export function LibraryToolbar(props: LibraryToolbarProps) {

    const { collectionList, ignoredLocalFiles, unmatchedLocalFiles, unknownGroups } = props

    const setBulkActionIsOpen = useSetAtom(bulkActionModalAtomIsOpen)

    const status = useAtomValue(serverStatusAtom)
    const setScannerModalOpen = useSetAtom(_scannerModalIsOpen)
    const setUnmatchedFileManagerOpen = useSetAtom(_unmatchedFileManagerIsOpen)
    const setUnknownMediaManagerOpen = useSetAtom(_unknownMediaManagerIsOpen)
    const setPlaylistsModalOpen = useSetAtom(__playlists_modalOpenAtom)

    const hasScanned = collectionList.some(n => n.entries.length > 0)

    const { openInExplorer } = useOpenInExplorer()

    return (
        <div className="flex flex-wrap w-full justify-end gap-2 p-4 relative z-[4]">
            <div className="flex flex-1"></div>
            {(!!status?.settings?.library?.libraryPath && hasScanned) && <Button
                intent={"white-subtle"}
                leftIcon={<FiPlayCircle className="text-2xl" />}
                onClick={() => setPlaylistsModalOpen(true)}
            >
                Playlists
            </Button>}
            {!!status?.settings?.library?.libraryPath && hasScanned && <Button
                intent={hasScanned ? "primary-subtle" : "primary"}
                leftIcon={<FiSearch className="text-xl" />}
                onClick={() => setScannerModalOpen(true)}
            >
                {hasScanned ? "Refresh entries" : "Scan your library"}
            </Button>}
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
                <DropdownMenuItem
                    disabled={!status?.settings?.library?.libraryPath}
                    className={cn("cursor-pointer", { "!text-[--muted]": !status?.settings?.library?.libraryPath })}
                    onClick={() => {
                        openInExplorer(status?.settings?.library?.libraryPath ?? "")
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
