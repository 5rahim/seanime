"use client"
import { useCustomSourceListAnime, useCustomSourceListManga } from "@/api/hooks/custom_source.hooks"
import { useListCustomSourceExtensions } from "@/api/hooks/extensions.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { MediaCardLazyGrid } from "@/app/(main)/_features/media/_components/media-card-grid"
import { MediaEntryCard } from "@/app/(main)/_features/media/_components/media-entry-card"
import { __customSources_paramsAtom } from "@/app/(main)/custom-sources/custom-sources.atom"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Pagination, PaginationEllipsis, PaginationItem, PaginationTrigger } from "@/components/ui/pagination"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useAtom } from "jotai/react"
import { useSearchParams } from "next/navigation"
import React from "react"
import { AiOutlineArrowLeft } from "react-icons/ai"
import { FiSearch } from "react-icons/fi"
import { MdDataSaverOn } from "react-icons/md"

export default function Page() {

    const urlParams = useSearchParams()
    const providerUrlParam = urlParams.get("provider")

    const { data: customSources } = useListCustomSourceExtensions()

    const [params, setParams] = useAtom(__customSources_paramsAtom)
    const [searchValue, setSearchValue] = React.useState(params.search)
    const [provider, setProvider] = React.useState<string | null>(null)

    const shouldFetch = !!provider

    const customSource = customSources?.find(s => s.id === provider)
    const supportsAnime = customSource?.settings?.supportsAnime
    const supportsManga = customSource?.settings?.supportsManga

    const animeQuery = useCustomSourceListAnime({
        provider: provider || "",
        search: params.search,
        page: params.page,
        perPage: params.perPage,
    }, {
        enabled: shouldFetch && params.type === "anime" && !!customSource?.settings?.supportsAnime,
    })

    const mangaQuery = useCustomSourceListManga({
        provider: provider || "",
        search: params.search,
        page: params.page,
        perPage: params.perPage,
    }, {
        enabled: shouldFetch && params.type === "manga" && !!customSource?.settings?.supportsManga,
    })

    const currentQuery = params.type === "anime" ? animeQuery : mangaQuery
    const { data, isLoading, error } = currentQuery


    React.useEffect(() => {
        if (customSources) {
            setProvider(providerUrlParam ? (customSources.find(s => s.id === providerUrlParam)?.id ?? customSources[0].id) : customSources[0].id)
        }
    }, [customSources, providerUrlParam])

    // Handle search input changes
    const handleSearch = React.useCallback(() => {
        setParams(draft => {
            draft.search = searchValue
            draft.page = 1 // Reset to first page on new search
            return
        })
    }, [searchValue, setParams])

    // Handle search on Enter key
    const handleSearchKeyDown = React.useCallback((e: React.KeyboardEvent) => {
        if (e.key === "Enter") {
            handleSearch()
        }
    }, [handleSearch])

    // Update search value when params change
    React.useEffect(() => {
        setSearchValue(params.search)
    }, [params.search])

    React.useLayoutEffect(() => {
        if (!customSource) return
        if (params.type === "anime" && !supportsAnime && supportsManga) {
            setParams(draft => {
                draft.type = "manga"
                return
            })
        } else if (params.type === "manga" && !supportsManga && supportsAnime) {
            setParams(draft => {
                draft.type = "anime"
                return
            })
        }
    }, [params, customSource])

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper data-search-page-container className="space-y-6 px-4 md:p-8 pt-0 pb-10">
                <div className="flex items-center gap-4">
                    <SeaLink href={`/discover`}>
                        <Button leftIcon={<AiOutlineArrowLeft />} rounded intent="gray-outline" size="md">
                            Discover
                        </Button>
                    </SeaLink>
                    {/*<h3>Discover</h3>*/}
                </div>
                <AppLayoutStack>
                    <h3 data-search-page-title className="text-center xl:text-left">
                        Custom source{provider ? `: ${customSource?.name ?? ""}` : "s"}
                    </h3>

                    <div className="flex gap-2">
                        <Select
                            leftAddon={<MdDataSaverOn className={cn("text-indigo-300 font-bold text-xl")} />}
                            placeholder="Select a source" className="w-full"
                            options={customSources?.map(s => ({
                                label: s.name,
                                value: s.id,
                            }))}
                            value={provider ?? ""}
                            onValueChange={v => {
                                setProvider(v)
                                setParams(draft => {
                                    draft.page = 1 // Reset page when changing provider
                                    return
                                })
                            }}
                            fieldClass="w-[400px]"
                        />
                        <Select
                            className="w-full"
                            options={[
                                ...((supportsAnime || !supportsManga) ? [{ label: "Anime", value: "anime" }] : []),
                                ...((supportsManga || !supportsAnime) ? [{ label: "Manga", value: "manga" }] : []),
                            ]}
                            value={params.type}
                            onValueChange={v => setParams(draft => {
                                draft.type = v as any
                                draft.page = 1 // Reset page when changing type
                                return
                            })}
                            fieldClass="w-[240px]"
                        />
                        <TextInput
                            leftIcon={<FiSearch />}
                            placeholder="Search titles..."
                            className="w-full"
                            value={searchValue}
                            onValueChange={setSearchValue}
                            onKeyDown={handleSearchKeyDown}
                        />
                        <Button
                            leftIcon={<FiSearch />}
                            intent="gray-outline"
                            size="md"
                            onClick={handleSearch}
                            loading={isLoading}
                        >
                            Search
                        </Button>
                    </div>

                    {!provider && <div className="text-center py-8 text-[--muted]">
                        Select a source to view its content
                    </div>}

                    {provider && <CustomSourceResults
                        provider={customSource?.name ?? ""}
                        data={data}
                        isLoading={isLoading}
                        error={error}
                        params={params}
                        setParams={setParams}
                    />}

                </AppLayoutStack>
            </PageWrapper>
        </>
    )
}

function CustomSourceResults({
    provider,
    data,
    isLoading,
    error,
    params,
    setParams,
}: {
    provider: string
    data: any
    isLoading: boolean
    error: any
    params: {
        search: string
        page: number
        perPage: number
        type: "anime" | "manga"
    }
    setParams: (updater: (draft: {
        search: string
        page: number
        perPage: number
        type: "anime" | "manga"
    }) => void) => void
}) {
    const media = data?.media || []
    const totalPages = data?.totalPages || 1
    const currentPage = params.page

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

            if (currentPage <= 3) {
                // Show pages 1,2,3,4,5 with ellipsis at end
                for (let i = 2; i <= 4; i++) {
                    pages.push(i)
                }
                if (totalPages > 4) {
                    pages.push("ellipsis")
                    pages.push(totalPages)
                }
            } else if (currentPage >= totalPages - 2) {
                // Show ellipsis at start and last few pages
                pages.push("ellipsis")
                for (let i = totalPages - 3; i <= totalPages; i++) {
                    if (i > 1) pages.push(i)
                }
            } else {
                // Show ellipsis on both sides
                pages.push("ellipsis")
                pages.push(currentPage - 1)
                pages.push(currentPage)
                pages.push(currentPage + 1)
                pages.push("ellipsis")
                pages.push(totalPages)
            }
        }

        return pages
    }

    const handlePageChange = (newPage: number) => {
        if (newPage >= 1 && newPage <= totalPages) {
            setParams(draft => {
                draft.page = newPage
                return
            })
        }
    }

    if (error) {
        return (
            <LuffyError title="Failed to load content">
                <p>Error loading content from {provider}</p>
            </LuffyError>
        )
    }

    if (isLoading) {
        return (
            <div className="flex justify-center py-8">
                <LoadingSpinner />
            </div>
        )
    }

    if (!media?.length) {
        return (
            <LuffyError title="No results found">
                <p>No {params.type} found for your search criteria</p>
            </LuffyError>
        )
    }

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <h4 className="text-lg font-medium">
                    {media.length} result{media.length === 1 ? "" : "s"}
                </h4>
                <div className="flex items-center gap-4">
                    <Select
                        value={String(params.perPage)}
                        onValueChange={v => setParams(draft => {
                            draft.perPage = Number(v)
                            draft.page = 1 // Reset to first page
                            return
                        })}
                        options={[20, 50, 100].map(size => ({
                            value: String(size),
                            label: `${size} per page`,
                        }))}
                        fieldClass="w-auto"
                        className="w-auto"
                        size="sm"
                    />
                    {totalPages > 1 && (
                        <div className="text-sm text-[--muted]">
                            Page {currentPage} of {totalPages}
                        </div>
                    )}
                </div>
            </div>

            <MediaCardLazyGrid itemCount={media.length}>
                {media.map((item: any, index: number) => (
                    <MediaEntryCard
                        key={`${item.id}-${index}`}
                        media={item}
                        type={params.type}
                        showLibraryBadge={true}
                    />
                ))}
            </MediaCardLazyGrid>

            {totalPages > 1 && (
                <div className="flex justify-center">
                    <Pagination>
                        <PaginationTrigger
                            direction="previous"
                            onClick={() => handlePageChange(currentPage - 1)}
                            disabled={currentPage <= 1 || isLoading}
                        />
                        {getVisiblePages().map((page, index) => (
                            page === "ellipsis" ? (
                                <PaginationEllipsis key={`ellipsis-${index}`} />
                            ) : (
                                <PaginationItem
                                    key={page}
                                    value={page}
                                    onClick={() => handlePageChange(page as number)}
                                    disabled={isLoading}
                                    className={page === currentPage ? "bg-[--brand] text-white" : ""}
                                />
                            )
                        ))}
                        <PaginationTrigger
                            direction="next"
                            onClick={() => handlePageChange(currentPage + 1)}
                            disabled={currentPage >= totalPages || isLoading}
                        />
                    </Pagination>
                </div>
            )}
        </div>
    )
}
