import { TorrentResolutionBadge, TorrentSeedersBadge } from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { DataGrid, defineDataGridColumns } from "@/components/ui/datagrid"
import { Tooltip } from "@/components/ui/tooltip"
import { AnimeTorrent } from "@/lib/server/types"
import { formatDistanceToNow } from "date-fns/formatDistanceToNow"
import React, { memo, useMemo } from "react"
import { BiLinkExternal } from "react-icons/bi"

type TorrentTable = {
    torrents: AnimeTorrent[]
    selectedTorrents: AnimeTorrent[]
    globalFilter: string,
    setGlobalFilter: React.Dispatch<React.SetStateAction<string>>
    quickSearch: boolean
    isLoading: boolean
    isFetching: boolean
    onToggleTorrent: (t: AnimeTorrent) => void
}

export const TorrentTable = memo((
    {
        torrents,
        selectedTorrents,
        globalFilter,
        setGlobalFilter,
        quickSearch,
        isFetching,
        isLoading,
        onToggleTorrent,
    }: TorrentTable) => {

    const columns = useMemo(() => defineDataGridColumns<AnimeTorrent>(() => [
        {
            accessorKey: "name",
            header: "Name",
            cell: info => <div className="flex items-center gap-2">
                <Tooltip
                    trigger={<IconButton
                        icon={<BiLinkExternal />}
                        intent="primary-basic"
                        size="sm"
                        onClick={() => window.open(info.row.original.link, "_blank")}
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
            size: 80,
        },
        {
            accessorKey: "resolution",
            header: "Resolution",
            cell: info => <TorrentResolutionBadge resolution={info.getValue<string>()} />,
            size: 2,
        },
        {
            accessorKey: "seeders",
            header: "Seeders",
            cell: info => <TorrentSeedersBadge seeders={info.getValue<number>()} />,
            size: 20,
        },
        {
            accessorKey: "date",
            header: "Date",
            cell: info => formatDistanceToNow(new Date(info.getValue<string>()), { addSuffix: true }),
            size: 10,
        },
    ]), [torrents, selectedTorrents])

    return (
        <DataGrid<AnimeTorrent>
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
            hideGlobalSearchInput={quickSearch}
        />
    )

})
