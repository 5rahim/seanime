"use client"
import { useAnilistListAnime } from "@/api/hooks/anilist.hooks"
import { useAnilistListManga } from "@/api/hooks/manga.hooks"
import { SeaLink } from "@/components/shared/sea-link"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Select } from "@/components/ui/select"
import { useDebounce } from "@/hooks/use-debounce"
import { Combobox, Dialog, Transition } from "@headlessui/react"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import { useRouter } from "next/navigation"
import React, { Fragment, useEffect, useRef } from "react"
import { BiChevronRight } from "react-icons/bi"
import { FiSearch } from "react-icons/fi"

export const __globalSearch_isOpenAtom = atom(false)

export function GlobalSearch() {

    const [inputValue, setInputValue] = React.useState("")
    const debouncedQuery = useDebounce(inputValue, 500)
    const inputRef = useRef<HTMLInputElement>(null)

    const [type, setType] = React.useState<string>("anime")

    const router = useRouter()

    const [open, setOpen] = useAtom(__globalSearch_isOpenAtom)

    useEffect(() => {
        if(open) {
            setTimeout(() => {
                console.log("open", open, inputRef.current)
                console.log("focusing")
                inputRef.current?.focus()
            }, 300)
        }
    }, [open])

    const { data: animeData, isLoading: animeIsLoading, isFetching: animeIsFetching } = useAnilistListAnime({
        search: debouncedQuery,
        page: 1,
        perPage: 10,
        status: ["FINISHED", "CANCELLED", "NOT_YET_RELEASED", "RELEASING"],
        sort: ["SEARCH_MATCH"],
    }, debouncedQuery.length > 0 && type === "anime")

    const { data: mangaData, isLoading: mangaIsLoading, isFetching: mangaIsFetching } = useAnilistListManga({
        search: debouncedQuery,
        page: 1,
        perPage: 10,
        status: ["FINISHED", "CANCELLED", "NOT_YET_RELEASED", "RELEASING"],
        sort: ["SEARCH_MATCH"],
    }, debouncedQuery.length > 0 && type === "manga")

    const isLoading = type === "anime" ? animeIsLoading : mangaIsLoading
    const isFetching = type === "anime" ? animeIsFetching : mangaIsFetching

    const media = React.useMemo(() => type === "anime" ? animeData?.Page?.media?.filter(Boolean) : mangaData?.Page?.media?.filter(Boolean),
        [animeData, mangaData, type])

    return (
        <>
            <Transition.Root show={open} as={Fragment} afterLeave={() => setInputValue("")} appear>
                <Dialog as="div" className="relative z-50" onClose={setOpen}>
                    <Transition.Child
                        as={Fragment}
                        enter="ease-out duration-300"
                        enterFrom="opacity-0"
                        enterTo="opacity-100"
                        leave="ease-in duration-200"
                        leaveFrom="opacity-100"
                        leaveTo="opacity-0"
                    >
                        <div className="fixed inset-0 bg-black bg-opacity-70 transition-opacity backdrop-blur-sm" />
                    </Transition.Child>

                    <div className="fixed inset-0 z-50 overflow-y-auto p-4 sm:p-6 md:p-20">
                        <Transition.Child
                            as={Fragment}
                            enter="ease-out duration-300"
                            enterFrom="opacity-0 scale-95"
                            enterTo="opacity-100 scale-100"
                            leave="ease-in duration-200"
                            leaveFrom="opacity-100 scale-100"
                            leaveTo="opacity-0 scale-95"
                        >
                            <Dialog.Panel
                                className="mx-auto max-w-3xl transform space-y-4 transition-all"
                            >
                                <div className="absolute right-2 -top-7 z-10">
                                    <SeaLink
                                        href="/search"
                                        className="text-[--muted] hover:text-[--foreground] font-bold"
                                        onClick={() => setOpen(false)}
                                    >
                                        Advanced search &rarr;
                                    </SeaLink>
                                </div>
                                <Combobox>
                                    {({ activeOption }: any) => (
                                        <>
                                            <div
                                                className="relative border bg-gray-950 shadow-2xl ring-1 ring-black ring-opacity-5 w-full rounded-lg "
                                            >
                                                <FiSearch
                                                    className="pointer-events-none absolute top-4 left-4 h-6 w-6 text-[--muted]"
                                                    aria-hidden="true"
                                                />
                                                <Combobox.Input
                                                    ref={inputRef}
                                                    className="h-14 w-full border-0 bg-transparent pl-14 pr-4 text-white placeholder-[--muted] focus:ring-0 sm:text-md"
                                                    placeholder="Search..."
                                                    onChange={(event) => setInputValue(event.target.value)}
                                                />
                                                <div className="block fixed lg:absolute top-2 right-2 z-1">
                                                    <Select
                                                        fieldClass="w-fit"
                                                        value={type}
                                                        onValueChange={(value) => setType(value)}
                                                        options={[
                                                            { value: "anime", label: "Anime" },
                                                            { value: "manga", label: "Manga" },
                                                        ]}
                                                    />
                                                </div>
                                            </div>

                                            {(!!media && media.length > 0) && (
                                                <Combobox.Options
                                                    as="div" static hold
                                                    className="flex divide-[--border] bg-gray-950 shadow-2xl ring-1 ring-black ring-opacity-5 rounded-lg border "
                                                >
                                                    <div
                                                        className={cn(
                                                            "max-h-96 min-w-0 flex-auto scroll-py-2 overflow-y-auto px-6 py-2 my-2",
                                                            { "sm:h-96": activeOption },
                                                        )}
                                                    >
                                                        <div className="-mx-2 text-sm text-[--foreground]">
                                                            {(media).map((item: any) => (
                                                                <Combobox.Option
                                                                    as="div"
                                                                    key={item.id}
                                                                    value={item}
                                                                    onClick={() => {
                                                                        if (type === "anime") {
                                                                            router.push(`/entry?id=${item.id}`)
                                                                        } else {
                                                                            router.push(`/manga/entry?id=${item.id}`)
                                                                        }
                                                                        setOpen(false)
                                                                    }}
                                                                    className={({ active }) =>
                                                                        cn(
                                                                            "flex select-none items-center rounded-[--radius-md] p-2 text-[--muted] cursor-pointer",
                                                                            active && "bg-gray-800 text-white",
                                                                        )
                                                                    }
                                                                >
                                                                    {({ active }) => (
                                                                        <>
                                                                            <div
                                                                                className="h-10 w-10 flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                                                                            >
                                                                                {item.coverImage?.medium && <Image
                                                                                    src={item.coverImage?.medium}
                                                                                    alt={""}
                                                                                    fill
                                                                                    quality={50}
                                                                                    priority
                                                                                    sizes="10rem"
                                                                                    className="object-cover object-center"
                                                                                />}
                                                                            </div>
                                                                            <span
                                                                                className="ml-3 flex-auto truncate"
                                                                            >{item.title?.userPreferred}</span>
                                                                            {active && (
                                                                                <BiChevronRight
                                                                                    className="ml-3 h-7 w-7 flex-none text-gray-400"
                                                                                    aria-hidden="true"
                                                                                />
                                                                            )}
                                                                        </>
                                                                    )}
                                                                </Combobox.Option>
                                                            ))}
                                                        </div>
                                                    </div>

                                                    {activeOption && (
                                                        <div
                                                            className="hidden min-h-96 w-1/2 flex-none flex-col overflow-y-auto sm:flex p-4"
                                                        >
                                                            <div className="flex-none p-6 text-center">
                                                                <div
                                                                    className="h-40 w-32 mx-auto flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                                                                >
                                                                    {activeOption.coverImage?.large && <Image
                                                                        src={activeOption.coverImage?.large}
                                                                        alt={""}
                                                                        fill
                                                                        quality={100}
                                                                        priority
                                                                        sizes="10rem"
                                                                        className="object-cover object-center"
                                                                    />}
                                                                </div>
                                                                <h4 className="mt-3 font-semibold text-[--foreground] line-clamp-3">{activeOption.title?.userPreferred}</h4>
                                                                <p className="text-sm leading-6 text-[--muted]">
                                                                    {activeOption.format}{activeOption.season
                                                                        ? ` - ${capitalize(activeOption.season)} `
                                                                        : " - "}{activeOption.seasonYear
                                                                            ? activeOption.seasonYear
                                                                            : "-"}
                                                                </p>
                                                            </div>
                                                            <SeaLink
                                                                href={type === "anime"
                                                                    ? `/entry?id=${activeOption.id}`
                                                                    : `/manga/entry?id=${activeOption.id}`}
                                                                onClick={() => setOpen(false)}
                                                            >
                                                                <Button
                                                                    type="button"
                                                                    className="w-full"
                                                                    intent="gray-subtle"
                                                                >
                                                                    Open
                                                                </Button>
                                                            </SeaLink>
                                                        </div>
                                                    )}
                                                </Combobox.Options>
                                            )}

                                            {(debouncedQuery !== "" && (!media || media.length === 0) && (isLoading || isFetching)) && (
                                                <LoadingSpinner />
                                            )}

                                            {debouncedQuery !== "" && !isLoading && !isFetching && (!media || media.length === 0) && (
                                                <div className="py-14 px-6 text-center text-sm sm:px-14">
                                                    {<div
                                                        className="h-[10rem] w-[10rem] mx-auto flex-none rounded-[--radius-md] object-cover object-center relative overflow-hidden"
                                                    >
                                                        <Image
                                                            src="/luffy-01.png"
                                                            alt={""}
                                                            fill
                                                            quality={100}
                                                            priority
                                                            sizes="10rem"
                                                            className="object-contain object-top"
                                                        />
                                                    </div>}
                                                    <h5 className="mt-4 font-semibold text-[--foreground]">Nothing
                                                        found</h5>
                                                    <p className="mt-2 text-[--muted]">
                                                        We couldn't find anything with that name. Please try again.
                                                    </p>
                                                </div>
                                            )}
                                        </>
                                    )}
                                </Combobox>
                            </Dialog.Panel>
                        </Transition.Child>
                    </div>
                </Dialog>
            </Transition.Root>
        </>
    )

}
