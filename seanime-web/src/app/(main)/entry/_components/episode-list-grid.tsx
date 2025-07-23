import { cn } from "@/components/ui/core/styling"
import { Pagination, PaginationEllipsis, PaginationItem, PaginationTrigger } from "@/components/ui/pagination"
import React, { useState } from "react"

type EpisodeListGridProps = {
    children?: React.ReactNode
}

export function EpisodeListGrid(props: EpisodeListGridProps) {

    const {
        children,
        ...rest
    } = props


    return (
        <div
            className={cn(
                "grid grid-cols-1 lg:grid-cols-2 2xl:grid-cols-3 min-[2000px]:grid-cols-4",
                "gap-4",
            )}
            {...rest}
            data-episode-list-grid
        >
            {children}
        </div>
    )
}

type EpisodeListPaginatedGridProps = {
    length: number
    renderItem: (index: number) => React.ReactNode
    itemsPerPage?: number
    minLengthBeforePagination?: number
    shouldDefaultToPageWithEpisode?: number // episode number
}

export function EpisodeListPaginatedGrid(props: EpisodeListPaginatedGridProps) {
    const {
        length,
        renderItem,
        itemsPerPage = 24,
        minLengthBeforePagination = 29,
        shouldDefaultToPageWithEpisode,
    } = props

    const [page, setPage] = useState(1)

    // Update page when shouldDefaultToPageWithEpisode changes
    React.useEffect(() => {
        if (shouldDefaultToPageWithEpisode && length >= minLengthBeforePagination) {
            const targetPage = Math.ceil(shouldDefaultToPageWithEpisode / itemsPerPage)
            const maxPage = Math.ceil(length / itemsPerPage)
            const validPage = Math.min(Math.max(1, targetPage), maxPage)
            setPage(validPage)
        } else {
            setPage(1)
        }
    }, [shouldDefaultToPageWithEpisode])

    // Only use pagination if we have enough items
    const shouldPaginate = length >= minLengthBeforePagination

    const totalPages = shouldPaginate ? Math.ceil(length / itemsPerPage) : 1
    const startIndex = shouldPaginate ? (page - 1) * itemsPerPage : 0
    const endIndex = shouldPaginate ? Math.min(startIndex + itemsPerPage, length) : length

    const currentItems = Array.from({ length: endIndex - startIndex }, (_, index) => renderItem(startIndex + index))

    // Calculate which page numbers to show
    const getVisiblePages = () => {
        const pages: (number | "ellipsis")[] = []
        const maxVisiblePages = 5

        if (totalPages <= maxVisiblePages) {
            // Show all pages if total is small
            for (let i = 1; i <= totalPages; i++) {
                pages.push(i)
            }
        } else {
            // Always show first page
            pages.push(1)

            if (page <= 3) {
                // Show pages 1,2,3,4,5 with ellipsis at end
                for (let i = 2; i <= 4; i++) {
                    pages.push(i)
                }
                if (totalPages > 4) {
                    pages.push("ellipsis")
                    pages.push(totalPages)
                }
            } else if (page >= totalPages - 2) {
                // Show ellipsis at start and last few pages
                pages.push("ellipsis")
                for (let i = totalPages - 3; i <= totalPages; i++) {
                    if (i > 1) pages.push(i)
                }
            } else {
                // Show ellipsis on both sides
                pages.push("ellipsis")
                pages.push(page - 1)
                pages.push(page)
                pages.push(page + 1)
                pages.push("ellipsis")
                pages.push(totalPages)
            }
        }

        return pages
    }

    const handlePageChange = (newPage: number) => {
        if (newPage >= 1 && newPage <= totalPages) {
            setPage(newPage)
        }
    }

    if (length === 0) {
        return null
    }

    return (
        <>
            <div
                className={cn(
                    "grid grid-cols-1 lg:grid-cols-2 2xl:grid-cols-3 min-[2000px]:grid-cols-4",
                    "gap-4",
                )}
                data-episode-list-grid
            >
                {currentItems}
            </div>

            {shouldPaginate && totalPages > 1 && (
                <div className="flex justify-center mt-6">
                    <Pagination>
                        <PaginationTrigger
                            direction="previous"
                            data-disabled={page === 1}
                            onClick={() => handlePageChange(page - 1)}
                        />

                        {getVisiblePages().map((pageNum, index) => (
                            pageNum === "ellipsis" ? (
                                <PaginationEllipsis key={`ellipsis-${index}`} />
                            ) : (
                                <PaginationItem
                                    key={pageNum}
                                    value={pageNum}
                                    data-selected={page === pageNum}
                                    onClick={() => handlePageChange(pageNum)}
                                />
                            )
                        ))}

                        <PaginationTrigger
                            direction="next"
                            data-disabled={page === totalPages}
                            onClick={() => handlePageChange(page + 1)}
                        />
                    </Pagination>
                </div>
            )}
        </>
    )
}
