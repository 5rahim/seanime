import { Debrid_TorrentItemInstantAvailability, HibikeTorrent_AnimeTorrent } from "@/api/generated/types"
import {
    TorrentDebridInstantAvailabilityBadge,
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { Tooltip } from "@/components/ui/tooltip"
import { openTab } from "@/lib/helpers/browser"
import { formatDistanceToNowSafe } from "@/lib/helpers/date"
import React, { memo, useMemo } from "react"
import { BiLinkExternal } from "react-icons/bi"

type TorrentTable = {
    torrents: HibikeTorrent_AnimeTorrent[]
    selectedTorrents: HibikeTorrent_AnimeTorrent[]
    globalFilter: string,
    setGlobalFilter: React.Dispatch<React.SetStateAction<string>>
    smartSearch: boolean
    isLoading: boolean
    isFetching: boolean
    onToggleTorrent: (t: HibikeTorrent_AnimeTorrent) => void
    debridInstantAvailability: Record<string, Debrid_TorrentItemInstantAvailability>
}

export const TorrentTable = memo((
    {
        torrents,
        selectedTorrents,
        globalFilter,
        setGlobalFilter,
        smartSearch,
        isFetching,
        isLoading,
        onToggleTorrent,
        debridInstantAvailability,
    }: TorrentTable) => {

    const columns = useMemo(() => defineDataGridColumns<HibikeTorrent_AnimeTorrent>(() => [
        {
            accessorKey: "name",
            header: "Name",
            cell: info => <div className="flex items-center gap-2">
                <Tooltip
                    trigger={<IconButton
                        icon={<BiLinkExternal />}
                        intent="primary-basic"
                        size="sm"
                        onClick={() => openTab(info.row.original.link)}
                    />}
                >Open in browser</Tooltip>
                <Tooltip
                    trigger={
                        <div
                            className={cn(
                                "text-[.95rem] line-clamp-1 cursor-pointer max-w-[90%] overflow-hidden",
                                {
                                    "text-brand-300 font-semibold": selectedTorrents.some(torrent => torrent.link === info.row.original.link),
                                },
                            )}
                            onClick={() => onToggleTorrent(info.row.original)}
                        >
                            {info.getValue<string>()}
                        </div>}
                >
                    {info.getValue<string>()}
                </Tooltip>
            </div>,
            size: 350,
        },
        {
            accessorKey: "resolution",
            header: "Resolution",
            cell: info => <TorrentResolutionBadge resolution={info.getValue<string>()} />,
            size: 70,
        },
        {
            accessorKey: "seeders",
            header: "Seeders",
            cell: info => (
                <div className="flex items-center gap-2 ">
                    <TorrentSeedersBadge seeders={info.getValue<number>()} />
                    {(!!info.row.original.infoHash && debridInstantAvailability[info.row.original.infoHash]) && (
                        <TorrentDebridInstantAvailabilityBadge />
                    )}
                </div>
            ),
            // size: 20,
        },
        {
            accessorKey: "date",
            header: "Date",
            cell: info => formatDistanceToNowSafe(info.getValue<string>()),
            size: 80,
        },
    ]), [torrents, selectedTorrents, debridInstantAvailability])

    return (
        <DataGrid<HibikeTorrent_AnimeTorrent>
            columns={columns}
            data={torrents?.slice(0, 20)}
            rowCount={torrents?.length ?? 0}
            initialState={{
                pagination: {
                    pageSize: 20,
                    pageIndex: 0,
                },
            }}
            tdClass="py-4 data-[row-selected=true]:bg-gray-900"
            tableBodyClass="bg-transparent"
            footerClass="hidden"
            state={{
                globalFilter,
            }}
            enableManualFiltering={true}
            onGlobalFilterChange={setGlobalFilter}
            isLoading={isLoading || isFetching}
            isDataMutating={isFetching}
            hideGlobalSearchInput={smartSearch}
        />
    )

})

