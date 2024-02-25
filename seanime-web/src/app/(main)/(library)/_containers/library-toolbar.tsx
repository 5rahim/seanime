"use client"
import { bulkActionModalAtomIsOpen } from "@/app/(main)/(library)/_containers/bulk-actions/bulk-action-modal"
import { _scannerModalIsOpen } from "@/app/(main)/(library)/_containers/scanner/scanner-modal"
import { _unknownMediaManagerIsOpen } from "@/app/(main)/(library)/_containers/unknown-media/unknown-media-manager"
import { _unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_containers/unmatched-files/unmatched-file-manager"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useOpenInExplorer } from "@/lib/server/hooks"
import { LibraryCollectionList, LocalFile, UnknownGroup } from "@/lib/server/types"
import { useAtomValue, useSetAtom } from "jotai/react"
import Link from "next/link"
import React from "react"
import { BiCollection, BiDotsVerticalRounded, BiFileFind, BiFolder } from "react-icons/bi"
import { FiDatabase, FiSearch } from "react-icons/fi"
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

    const hasScanned = collectionList.some(n => n.entries.length > 0)

    const { openInExplorer } = useOpenInExplorer()

    return (
        <div className="flex w-full justify-between p-4">
            <div className="flex gap-2">
                {!!status?.settings?.library?.libraryPath && hasScanned && <Button
                    intent={hasScanned ? "primary-subtle" : "primary"}
                    leftIcon={<FiSearch />}
                    onClick={() => setScannerModalOpen(true)}
                >
                    {hasScanned ? "Refresh entries" : "Scan your library"}
                </Button>}
                {(unmatchedLocalFiles.length > 0) && <Button
                    intent="alert-outline"
                    leftIcon={<FiDatabase />}
                    className="animate-pulse"
                    onClick={() => setUnmatchedFileManagerOpen(true)}
                >
                    Resolve unmatched ({unmatchedLocalFiles.length})
                </Button>}
                {(unknownGroups.length > 0) && <Button
                    intent="warning-outline"
                    leftIcon={<BiFileFind />}
                    className="animate-pulse"
                    onClick={() => setUnknownMediaManagerOpen(true)}
                >
                    Resolve hidden media ({unknownGroups.length})
                </Button>}
            </div>
            <div className="flex gap-2">
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
        </div>
    )

}
