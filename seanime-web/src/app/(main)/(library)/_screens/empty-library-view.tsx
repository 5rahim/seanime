import { __scanner_modalIsOpen } from "@/app/(main)/(library)/_containers/scanner-modal"
import { __mainLibrary_paramsAtom, __mainLibrary_paramsInputAtom } from "@/app/(main)/(library)/_lib/handle-library-collection"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { DiscoverPageHeader } from "@/app/(main)/discover/_components/discover-page-header"
import { DiscoverTrending } from "@/app/(main)/discover/_containers/discover-trending"
import { LuffyError } from "@/components/shared/luffy-error"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { HorizontalDraggableScroll } from "@/components/ui/horizontal-draggable-scroll"
import { StaticTabs } from "@/components/ui/tabs"
import { useDebounce } from "@/hooks/use-debounce"
import { useSetAtom } from "jotai/index"
import { useAtom } from "jotai/react"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { LuCog } from "react-icons/lu"

type EmptyLibraryViewProps = {
    isLoading: boolean
    hasEntries: boolean
}

export function EmptyLibraryView(props: EmptyLibraryViewProps) {

    const {
        isLoading,
        hasEntries,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const setScannerModalOpen = useSetAtom(__scanner_modalIsOpen)

    if (hasEntries || isLoading) return null

    /**
     * Show empty library message and trending if library is empty
     */
    return (
        <>
            <DiscoverPageHeader />
            <PageWrapper className="p-4 sm:p-8 pt-0 space-y-8 relative z-[4]" data-empty-library-view-container>
                <div className="text-center space-y-4">
                    <div className="w-fit mx-auto space-y-4">
                        {!!serverStatus?.settings?.library?.libraryPath ? <>
                            <h2>Empty library</h2>
                            <Button
                                intent="primary-outline"
                                leftIcon={<FiSearch />}
                                size="xl"
                                rounded
                                onClick={() => setScannerModalOpen(true)}
                            >
                                Scan your library
                            </Button>
                        </> : (
                            <LuffyError
                                title="Your library is empty"
                                className=""
                            >
                                <div className="text-center space-y-4">
                                    <SeaLink href="/settings?tab=library">
                                        <Button intent="primary-subtle" leftIcon={<LuCog className="text-xl" />}>
                                            Set the path to your local library and scan it
                                        </Button>
                                    </SeaLink>
                                    {serverStatus?.settings?.library?.enableOnlinestream && <p>
                                        <SeaLink href="/settings?tab=onlinestream">
                                            <Button intent="primary-subtle" leftIcon={<LuCog className="text-xl" />}>
                                                Include online streaming in your library
                                            </Button>
                                        </SeaLink>
                                    </p>}
                                    {serverStatus?.torrentstreamSettings?.enabled && <p>
                                        <SeaLink href="/settings?tab=torrentstream">
                                            <Button intent="primary-subtle" leftIcon={<LuCog className="text-xl" />}>
                                                Include torrent streaming in your library
                                            </Button>
                                        </SeaLink>
                                    </p>}
                                    {serverStatus?.debridSettings?.enabled && <p>
                                        <SeaLink href="/settings?tab=debrid">
                                            <Button intent="primary-subtle" leftIcon={<LuCog className="text-xl" />}>
                                                Include debrid streaming in your library
                                            </Button>
                                        </SeaLink>
                                    </p>}
                                </div>
                            </LuffyError>
                        )}
                    </div>
                </div>
                <div className="">
                    <h3>Trending this season</h3>
                    <DiscoverTrending />
                </div>
            </PageWrapper>
        </>
    )
}

function GenreSelector({
    genres,
}: { genres: string[] }) {
    const [params, setParams] = useAtom(__mainLibrary_paramsInputAtom)
    const setActualParams = useSetAtom(__mainLibrary_paramsAtom)
    const debouncedParams = useDebounce(params, 500)

    React.useEffect(() => {
        setActualParams(params)
    }, [debouncedParams])

    if (!genres.length) return null

    return (
        <HorizontalDraggableScroll className="scroll-pb-1 pt-4 flex">
            <div className="flex flex-1"></div>
            <StaticTabs
                className="px-2 overflow-visible gap-2 py-4 w-fit"
                triggerClass="text-base rounded-[--radius-md] ring-2 ring-transparent data-[current=true]:ring-brand-500 data-[current=true]:text-brand-300"
                items={[
                    // {
                    //     name: "All",
                    //     isCurrent: !params!.genre?.length,
                    //     onClick: () => setParams(draft => {
                    //         draft.genre = []
                    //         return
                    //     }),
                    // },
                    ...genres.map(genre => ({
                        name: genre,
                        isCurrent: params!.genre?.includes(genre) ?? false,
                        onClick: () => setParams(draft => {
                            if (draft.genre?.includes(genre)) {
                                draft.genre = draft.genre?.filter(g => g !== genre)
                            } else {
                                draft.genre = [...(draft.genre || []), genre]
                            }
                            return
                        }),
                    })),
                ]}
            />
            <div className="flex flex-1"></div>
        </HorizontalDraggableScroll>
    )
}
