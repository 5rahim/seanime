import React, { memo, useMemo } from "react"
import { SearchTorrent } from "@/lib/server/types"
import { createDataGridColumns, DataGrid } from "@/components/ui/datagrid"
import { Tooltip } from "@/components/ui/tooltip"
import { IconButton } from "@/components/ui/button"
import { BiLinkExternal } from "@react-icons/all-files/bi/BiLinkExternal"
import { cn } from "@/components/ui/core"
import {
    TorrentResolutionBadge,
    TorrentSeedersBadge,
} from "@/app/(main)/entry/_containers/torrent-search/_components/torrent-item-badges"
import formatDistanceToNow from "date-fns/formatDistanceToNow"

type TorrentTable = {
    torrents: SearchTorrent[]
    selectedTorrents: SearchTorrent[]
    globalFilter: string,
    setGlobalFilter: React.Dispatch<React.SetStateAction<string>>
    quickSearch: boolean
    isLoading: boolean
    isFetching: boolean
    onToggleTorrent: (t: SearchTorrent) => void
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

    const columns = useMemo(() => createDataGridColumns<SearchTorrent>(() => [
        {
            accessorKey: "name",
            header: "Name",
            cell: info => <div className={"flex items-center gap-2"}>
                <Tooltip trigger={<IconButton
                    icon={<BiLinkExternal/>}
                    intent={"primary-basic"}
                    size={"sm"}
                    onClick={() => window.open(info.row.original.guid, "_blank")}
                />}>View on NYAA</Tooltip>
                <Tooltip
                    trigger={
                        <div
                            className={cn(
                                "text-[.95rem] truncate text-ellipsis cursor-pointer max-w-[90%] overflow-hidden",
                                {
                                    "text-brand-300 font-semibold": selectedTorrents.some(torrent => torrent.guid === info.row.original.guid),
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
            cell: info => <TorrentResolutionBadge resolution={info.getValue<string>()}/>,
            size: 2,
        },
        {
            accessorKey: "seeders",
            header: "Seeders",
            cell: info => <TorrentSeedersBadge seeders={info.getValue<string>()}/>,
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
        <DataGrid<SearchTorrent>
            columns={columns}
            data={torrents?.slice(0, 20)}
            rowCount={torrents?.length ?? 0}
            initialState={{
                pagination: {
                    pageSize: 20,
                    pageIndex: 0,
                },
            }}
            tdClassName={"py-4 data-[row-selected=true]:bg-gray-900"}
            tableBodyClassName={"bg-transparent"}
            footerClassName={"hidden"}
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