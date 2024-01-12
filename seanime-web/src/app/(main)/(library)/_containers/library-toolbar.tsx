import { bulkActionModalAtomIsOpen } from "@/app/(main)/(library)/_components/bulk-action-modal"
import { _scannerModalIsOpen } from "@/app/(main)/(library)/_components/scanner-modal"
import { _unknownMediaManagerIsOpen } from "@/app/(main)/(library)/_components/unknown-media-manager"
import { _unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_components/unmatched-file-manager"
import { serverStatusAtom } from "@/atoms/server-status"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core"
import { DropdownMenu } from "@/components/ui/dropdown-menu"
import { useOpenInExplorer } from "@/lib/server/hooks/settings"
import { LibraryCollectionList, LocalFile, UnknownGroup } from "@/lib/server/types"
import { BiCollection } from "@react-icons/all-files/bi/BiCollection"
import { BiDotsVerticalRounded } from "@react-icons/all-files/bi/BiDotsVerticalRounded"
import { BiFileFind } from "@react-icons/all-files/bi/BiFileFind"
import { BiFolder } from "@react-icons/all-files/bi/BiFolder"
import { FiDatabase } from "@react-icons/all-files/fi/FiDatabase"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { useAtomValue, useSetAtom } from "jotai/react"

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
                    leftIcon={<FiSearch/>}
                    onClick={() => setScannerModalOpen(true)}
                >
                    {hasScanned ? "Refresh entries" : "Scan your library"}
                </Button>}
                {(unmatchedLocalFiles.length > 0) && <Button
                    intent="alert-outline"
                    leftIcon={<FiDatabase/>}
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
                <DropdownMenu trigger={<IconButton icon={<BiDotsVerticalRounded/>} intent={"gray-basic"}/>}>
                    <DropdownMenu.Item
                        disabled={!status?.settings?.library?.libraryPath}
                        className={cn({ "!text-[--muted]": !status?.settings?.library?.libraryPath })}
                        onClick={() => {
                            openInExplorer(status?.settings?.library?.libraryPath ?? "")
                        }}
                    >
                        <BiFolder/>
                        <span>Open folder</span>
                    </DropdownMenu.Item>

                    {/*<DropdownMenu.Item*/}
                    {/*    onClick={() => {*/}
                    {/*    }}*/}
                    {/*    disabled={ignoredLocalFiles.length === 0}*/}
                    {/*    className={cn({ "!text-[--muted]": ignoredLocalFiles.length === 0 })}*/}
                    {/*>*/}
                    {/*    <GoDiffIgnored/>*/}
                    {/*    <span>Manage ignored files</span>*/}
                    {/*</DropdownMenu.Item>*/}

                    <DropdownMenu.Item
                        onClick={() => setBulkActionIsOpen(true)}
                        disabled={!hasScanned}
                        className={cn({ "!text-[--muted]": !hasScanned })}
                    >
                        <BiCollection/>
                        <span>Bulk actions</span>
                    </DropdownMenu.Item>
                </DropdownMenu>

            </div>
        </div>
    )

}
