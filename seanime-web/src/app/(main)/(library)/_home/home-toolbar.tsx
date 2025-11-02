"use client"
import { Anime_LibraryCollectionList, Anime_LocalFile, Anime_UnknownGroup } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useGetAllExtensions } from "@/api/hooks/extensions.hooks"
import { __bulkAction_modalAtomIsOpen } from "@/app/(main)/(library)/_containers/bulk-action-modal"
import { __ignoredFileManagerIsOpen } from "@/app/(main)/(library)/_containers/ignored-file-manager"
import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { __unknownMedia_drawerIsOpen } from "@/app/(main)/(library)/_containers/unknown-media-manager"
import { __unmatchedFileManagerIsOpen } from "@/app/(main)/(library)/_containers/unmatched-file-manager"
import { __home_currentView } from "@/app/(main)/(library)/_home/home-screen"
import { HomeSettingsButton } from "@/app/(main)/(library)/_home/home-settings-button"
import { libraryExplorer_drawerOpenAtom } from "@/app/(main)/_features/library-explorer/library-explorer.atoms"
import { useNakamaStatus } from "@/app/(main)/_features/nakama/nakama-manager"
import { usePlaylistEditorManager } from "@/app/(main)/_features/playlists/lib/playlist-editor-manager"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AddExtensionModal } from "@/app/(main)/extensions/_containers/add-extension-modal"
import { SeaLink } from "@/components/shared/sea-link"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DropdownMenu, DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { Tooltip } from "@/components/ui/tooltip"
import { useThemeSettings } from "@/lib/theme/hooks"
import { useAtom, useSetAtom } from "jotai/react"
import React from "react"
import { BiCollection, BiDotsVerticalRounded, BiFolder } from "react-icons/bi"
import { HiExclamation } from "react-icons/hi"
import { IoHome, IoLibraryOutline, IoLibrarySharp } from "react-icons/io5"
import { LuFolderSearch, LuFolderSync, LuFolderTree } from "react-icons/lu"
import { MdOutlineConnectWithoutContact, MdOutlineVideoLibrary } from "react-icons/md"
import { TbFileSad, TbReportSearch } from "react-icons/tb"
import { PluginAnimeLibraryDropdownItems } from "../../_features/plugin/actions/plugin-actions"

export type HomeToolbarProps = {
    collectionList: Anime_LibraryCollectionList[]
    ignoredLocalFiles: Anime_LocalFile[]
    unmatchedLocalFiles: Anime_LocalFile[]
    unknownGroups: Anime_UnknownGroup[]
    isLoading: boolean
    hasEntries: boolean
    isStreamingOnly: boolean
    isNakamaLibrary: boolean
    className?: string
}

export function HomeToolbar(props: HomeToolbarProps) {

    const {
        collectionList,
        ignoredLocalFiles,
        unmatchedLocalFiles,
        unknownGroups,
        hasEntries,
        isStreamingOnly,
        isNakamaLibrary,
        className,
    } = props

    const ts = useThemeSettings()
    const setBulkActionIsOpen = useSetAtom(__bulkAction_modalAtomIsOpen)
    const nakamaStatus = useNakamaStatus()

    const status = useServerStatus()
    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)
    const setUnmatchedFileManagerOpen = useSetAtom(__unmatchedFileManagerIsOpen)
    const setIgnoredFileManagerOpen = useSetAtom(__ignoredFileManagerIsOpen)
    const setUnknownMediaManagerOpen = useSetAtom(__unknownMedia_drawerIsOpen)
    const setLibraryExplorerDrawerOpen = useSetAtom(libraryExplorer_drawerOpenAtom)
    const { setModalOpen } = usePlaylistEditorManager()

    const { data: allExtensions, isLoading: isExtensionsLoading } = useGetAllExtensions(false)

    const [homeView, setHomeView] = useAtom(__home_currentView)

    const { mutate: openInExplorer } = useOpenInExplorer()

    const hasLibraryPath = !!status?.settings?.library?.libraryPath

    return (
        <>
            <div className={cn("flex flex-wrap w-full justify-end gap-1 p-4 relative z-[10]", className)} data-home-toolbar-container>
                <div className="flex flex-1 pointer-events-none" data-home-toolbar-spacer></div>
                {isNakamaLibrary && <Tooltip
                    trigger={<div className="flex items-center px-2 h-10">
                        <MdOutlineConnectWithoutContact className="size-8" />
                    </div>}
                >
                    {nakamaStatus?.hostConnectionStatus?.username}'s Library
                </Tooltip>}
                {(!isExtensionsLoading && !allExtensions?.extensions?.some(n => n.type === "anime-torrent-provider")) &&
                    <AddExtensionModal extensions={allExtensions?.extensions}>
                        <span>
                            <Tooltip
                                trigger={<Button
                                    data-home-toolbar-missing-extensions-button
                                    className="animate-bounce"
                                    intent={"warning"}
                                    leftIcon={<HiExclamation className="text-2xl" />}
                                >
                                    Add missing extensions
                                </Button>}
                            >
                                No torrent providers installed.
                            </Tooltip>
                        </span>
                    </AddExtensionModal>}
                {(hasEntries) && (
                    <>
                        {(!isStreamingOnly && !isNakamaLibrary) && <Tooltip
                            trigger={<IconButton
                                data-home-toolbar-switch-view-button
                                intent={homeView === "base" ? "white-subtle" : "white"}
                                icon={homeView === "base" ? <IoLibraryOutline className="text-2xl" /> : <IoHome className="text-2xl" />}
                                onClick={() => setHomeView(p => p === "detailed" ? "base" : "detailed")}
                            />}
                        >
                            {homeView === "base" ? "Local Anime Library" : "Home"}
                        </Tooltip>}

                        {(!isStreamingOnly && !isNakamaLibrary && hasLibraryPath) && <Tooltip
                            trigger={<IconButton
                                data-home-toolbar-switch-view-button
                                intent={"white-subtle"}
                                icon={<LuFolderTree className="text-2xl" />}
                                onClick={() => {
                                    setLibraryExplorerDrawerOpen(true)
                                }}
                                className={cn(unmatchedLocalFiles.length > 0 && "animate-pulse")}
                            />}
                        >
                            Library Explorer
                        </Tooltip>}

                        <Tooltip
                            trigger={<IconButton
                                data-home-toolbar-playlists-button
                                intent={"white-subtle"}
                                icon={<MdOutlineVideoLibrary className="text-2xl" />}
                                onClick={() => setModalOpen(true)}
                            />}
                        >Playlists</Tooltip>
                    </>
                )}
                {/*Shows up even when there's no local entries*/}
                {!isNakamaLibrary && hasLibraryPath && <Tooltip
                    trigger={<div>
                        <Button
                            data-home-toolbar-scan-button
                            intent={hasEntries ? "white-subtle" : "primary"}
                            leftIcon={hasEntries ? <LuFolderSync className="text-xl" /> : <LuFolderSearch className="text-xl" />}
                            onClick={() => setScannerModalOpen(true)}
                            hideTextOnSmallScreen
                        >
                            {hasEntries ? "Refresh" : "Scan"}
                        </Button>
                    </div>}
                >
                    {hasEntries ? "Refresh Anime Library" : "Scan Anime Library"}
                </Tooltip>}
                {(unmatchedLocalFiles.length > 0) && <Button
                    data-home-toolbar-unmatched-button
                    intent="alert"
                    leftIcon={<IoLibrarySharp />}
                    className="animate-bounce"
                    onClick={() => setUnmatchedFileManagerOpen(true)}
                >
                    Resolve unmatched ({unmatchedLocalFiles.length})
                </Button>}
                {(unknownGroups.length > 0) && <Button
                    data-home-toolbar-unknown-button
                    intent="warning"
                    leftIcon={<IoLibrarySharp />}
                    className="animate-bounce"
                    onClick={() => setUnknownMediaManagerOpen(true)}
                >
                    Resolve hidden media ({unknownGroups.length})
                </Button>}

                <HomeSettingsButton type="toolbar" />

                {(!isStreamingOnly && !isNakamaLibrary && hasLibraryPath) &&
                    <DropdownMenu
                        trigger={<IconButton
                            data-home-toolbar-dropdown-menu-trigger
                            icon={<BiDotsVerticalRounded />} intent="gray-basic"
                        />}
                    >

                        {/*<DropdownMenuItem*/}
                        {/*    data-home-toolbar-open-library-explorer-button*/}
                        {/*    disabled={!hasLibraryPath}*/}
                        {/*    className={cn("cursor-pointer", { "!text-[--muted]": !hasLibraryPath })}*/}
                        {/*    onClick={() => {*/}
                        {/*        setLibraryExplorerDrawerOpen(true)*/}
                        {/*    }}*/}
                        {/*>*/}
                        {/*    <LuFolderTree />*/}
                        {/*    <span>Library explorer</span>*/}
                        {/*</DropdownMenuItem>*/}

                        <DropdownMenuItem
                            data-home-toolbar-open-directory-button
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
                            data-home-toolbar-bulk-actions-button
                            onClick={() => setBulkActionIsOpen(true)}
                            disabled={!hasEntries}
                            className={cn({ "!text-[--muted]": !hasEntries })}
                        >
                            <BiCollection />
                            <span>Bulk actions</span>
                        </DropdownMenuItem>

                        <DropdownMenuItem
                            data-home-toolbar-ignored-files-button
                            onClick={() => setIgnoredFileManagerOpen(true)}
                            // disabled={!hasEntries}
                            className={cn({ "!text-[--muted]": !hasEntries })}
                        >
                            <TbFileSad />
                            <span>Ignored files</span>
                        </DropdownMenuItem>

                        <SeaLink href="/scan-summaries">
                            <DropdownMenuItem
                                data-home-toolbar-scan-summaries-button
                                // className={cn({ "!text-[--muted]": !hasEntries })}
                            >
                                <TbReportSearch />
                                <span>Scan summaries</span>
                            </DropdownMenuItem>
                        </SeaLink>

                        {/*{!(isStreamingOnly || isNakamaLibrary) && <PlayRandomEpisodeButton />}*/}

                        <PluginAnimeLibraryDropdownItems />
                    </DropdownMenu>}

            </div>
        </>
    )

}
