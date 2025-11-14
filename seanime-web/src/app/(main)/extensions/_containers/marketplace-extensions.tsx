import { Extension_Extension } from "@/api/generated/types"
import {
    useGetAllExtensions,
    useGetMarketplaceExtensions,
    useInstallExternalExtension,
    useReloadExternalExtension,
} from "@/api/hooks/extensions.hooks"
import { EXTENSION_TYPE } from "@/app/(main)/extensions/_containers/extension-list"
import { DEFAULT_MARKETPLACE_URL, marketplaceUrlAtom } from "@/app/(main)/extensions/_lib/marketplace.atoms"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { LuffyError } from "@/components/shared/luffy-error"
import { SeaImage } from "@/components/shared/sea-image"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { StaticTabs } from "@/components/ui/tabs"
import { TextInput } from "@/components/ui/text-input"
import { useAtom } from "jotai/react"
import { orderBy } from "lodash"
import capitalize from "lodash/capitalize"
import { useSearchParams } from "next/navigation"
import React, { useMemo } from "react"
import { AiOutlineExclamationCircle } from "react-icons/ai"
import { BiSearch } from "react-icons/bi"
import { CgMediaPodcast } from "react-icons/cg"
import { LuBlocks, LuBookOpen, LuCheck, LuDownload, LuSettings } from "react-icons/lu"
import { MdDataSaverOn } from "react-icons/md"
import { RiFolderDownloadFill } from "react-icons/ri"
import { toast } from "sonner"

type MarketplaceExtensionsProps = {
    children?: React.ReactNode
}

export function MarketplaceExtensions(props: MarketplaceExtensionsProps) {
    const {
        children,
        ...rest
    } = props

    const [searchTerm, setSearchTerm] = React.useState("")
    const [filterType, setFilterType] = React.useState<string>("all")
    const [filterLanguage, setFilterLanguage] = React.useState<string>("all")
    const [marketplaceUrl, setMarketplaceUrl] = useAtom(marketplaceUrlAtom)
    const [isUrlModalOpen, setIsUrlModalOpen] = React.useState(false)
    const [tempUrl, setTempUrl] = React.useState(marketplaceUrl)
    const [urlError, setUrlError] = React.useState("")
    const [isUpdatingUrl, setIsUpdatingUrl] = React.useState(false)

    const { data: marketplaceExtensions, isPending: isLoadingMarketplace, refetch } = useGetMarketplaceExtensions(marketplaceUrl)
    const { data: allExtensions, isPending: isLoadingAllExtensions } = useGetAllExtensions(false)

    const searchParams = useSearchParams()
    React.useEffect(() => {
        const type = searchParams.get("type")
        if (type) {
            setFilterType(type)
        }
    }, [searchParams])

    function orderExtensions(extensions: Extension_Extension[] | undefined) {
        return extensions ?
            orderBy(extensions, ["name", "manifestUri"])
            : []
    }

    function isExtensionInstalled(extensionID: string) {
        return !!allExtensions?.extensions?.find(n => n.id === extensionID) ||
            !!allExtensions?.invalidExtensions?.find(n => n.id === extensionID)
    }

    // Filter extensions based on search term, filter type, and language
    const filteredExtensions = React.useMemo(() => {
        if (!marketplaceExtensions) return []

        let filtered = [...marketplaceExtensions]

        // Filter by type if not "all"
        if (filterType !== "all") {
            filtered = filtered.filter(ext => ext.type === filterType)
        }

        // Filter by language if not "all"
        if (filterLanguage !== "all") {
            filtered = filtered.filter(ext => ext.lang?.toLowerCase() === filterLanguage.toLowerCase())
        }

        // Filter by search term
        if (searchTerm) {
            const term = searchTerm.toLowerCase()
            filtered = filtered.filter(ext =>
                ext.name.toLowerCase().includes(term) ||
                ext.description?.toLowerCase().includes(term) ||
                ext.id.toLowerCase().includes(term),
            )
        }

        return orderExtensions(filtered)
    }, [marketplaceExtensions, searchTerm, filterType, filterLanguage])

    // Get available languages from extensions
    const availableLanguages = useMemo(() => {
        if (!marketplaceExtensions) return []

        // Get unique languages from extensions
        const langSet = new Set<string>()
        marketplaceExtensions.forEach(ext => {
            if (ext.lang) langSet.add(ext.lang.toLowerCase())
        })

        // Convert to array and sort
        return Array.from(langSet).sort()
    }, [marketplaceExtensions])

    // Create language options for dropdown
    const languageOptions = useMemo(() => {
        const options = [{ value: "all", label: "All Languages" }]

        availableLanguages.forEach(langCode => {
            const langInfo = LANGUAGES_LIST[langCode]
            if (langInfo) {
                options.push({
                    value: langCode,
                    label: langInfo.name || langCode.toUpperCase(),
                })
            } else {
                options.push({
                    value: langCode,
                    label: langCode.toUpperCase(),
                })
            }
        })

        return options
    }, [availableLanguages])

    // Group extensions by type
    const pluginExtensions = filteredExtensions.filter(n => n.type === "plugin")
    const animeTorrentExtensions = filteredExtensions.filter(n => n.type === "anime-torrent-provider")
    const mangaExtensions = filteredExtensions.filter(n => n.type === "manga-provider")
    const onlinestreamExtensions = filteredExtensions.filter(n => n.type === "onlinestream-provider")
    const customSources = filteredExtensions.filter(n => n.type === "custom-source")

    // if (isLoadingMarketplace || isLoadingAllExtensions) return <LoadingSpinner />

    // validate URL
    const validateUrl = (url: string): boolean => {
        try {
            new URL(url)
            setUrlError("")
            return true
        }
        catch (e) {
            setUrlError("Please enter a valid URL")
            return false
        }
    }

    // handle URL change
    const handleUrlChange = async () => {
        if (validateUrl(tempUrl)) {
            setIsUpdatingUrl(true)
            try {
                setMarketplaceUrl(tempUrl)
                await refetch()
                setIsUrlModalOpen(false)
                toast.success("Marketplace URL updated")
            }
            catch (error) {
                toast.error("Failed to fetch extensions from the provided URL")
                console.error("Error fetching extensions:", error)
            }
            finally {
                setIsUpdatingUrl(false)
            }
        }
    }

    // reset URL to default
    const resetToDefaultUrl = async () => {
        setTempUrl(DEFAULT_MARKETPLACE_URL)
        setUrlError("")
    }

    // apply default URL immediately
    const applyDefaultUrl = async () => {
        setIsUpdatingUrl(true)
        try {
            setMarketplaceUrl(DEFAULT_MARKETPLACE_URL)
            await refetch()
            setIsUrlModalOpen(false)
            toast.success("Reset to default marketplace URL")
        }
        catch (error) {
            toast.error("Failed to fetch extensions from the default URL")
            console.error("Error fetching extensions:", error)
        }
        finally {
            setIsUpdatingUrl(false)
        }
    }

    return (
        <AppLayoutStack className="gap-6">
            <Modal
                open={isUrlModalOpen}
                onOpenChange={setIsUrlModalOpen}
                title="Repository URL"
            >
                <div className="space-y-4">
                    <p className="text-sm text-[--muted]">
                        Enter the URL of the repository JSON file.
                    </p>

                    <TextInput
                        label="Marketplace URL"
                        value={tempUrl}
                        onValueChange={(value) => {
                            setTempUrl(value)
                            // Validate as user types, but only if there's some input
                            if (value) validateUrl(value)
                        }}
                        error={urlError}
                        placeholder="Enter marketplace URL"
                    />

                    <div className="flex justify-between">
                        <div className="flex gap-2">
                            {/*<Button*/}
                            {/*    intent="gray-outline"*/}
                            {/*    onClick={resetToDefaultUrl}*/}
                            {/*>*/}
                            {/*    Set to Default*/}
                            {/*</Button>*/}
                            <Button
                                intent="primary-subtle"
                                onClick={applyDefaultUrl}
                                loading={isUpdatingUrl}
                                disabled={isUpdatingUrl}
                            >
                                Apply Default
                            </Button>
                        </div>

                        <div className="flex gap-2">
                            <Button
                                intent="gray-outline"
                                onClick={() => setIsUrlModalOpen(false)}
                            >
                                Cancel
                            </Button>

                            <Button
                                intent="primary"
                                onClick={handleUrlChange}
                                disabled={!tempUrl || !!urlError || isUpdatingUrl}
                                loading={isUpdatingUrl}
                            >
                                Save
                            </Button>
                        </div>
                    </div>
                </div>
            </Modal>

            <div className="flex items-center gap-2 flex-wrap">
                <div>
                    <h2>
                        Marketplace
                    </h2>
                    <p className="text-[--muted] text-sm">
                        Browse and install extensions from the repository.
                    </p>
                    <p className="text-[--muted] text-xs mt-1">
                        Source: {marketplaceUrl === DEFAULT_MARKETPLACE_URL ?
                        <span>Official repository</span> :
                        <span>{marketplaceUrl}</span>
                    }
                    </p>
                </div>

                <div className="flex flex-1"></div>

                <div className="flex items-center gap-2">
                    <Button
                        className="rounded-full"
                        intent="gray-outline"
                        onClick={() => {
                            refetch()
                            toast.success("Refreshed", { duration: 1000 })
                        }}
                    >
                        Refresh
                    </Button>
                    <Button
                        className="rounded-full"
                        intent="gray-outline"
                        leftIcon={<LuSettings />}
                        onClick={() => {
                            setTempUrl(marketplaceUrl)
                            setUrlError("")
                            setIsUrlModalOpen(true)
                        }}
                    >
                        Change repository
                    </Button>
                </div>
            </div>

            <div className="flex flex-wrap gap-4">
                <StaticTabs
                    className="h-10 w-fit border rounded-full"
                    triggerClass="px-4 py-1 text-sm"
                    items={[
                        {
                            name: "All Types",
                            isCurrent: filterType === "all",
                            onClick: () => setFilterType("all"),
                            // iconType: IoGrid,
                        },
                        {
                            name: "Plugins",
                            isCurrent: filterType === "plugin",
                            onClick: () => setFilterType("plugin"),
                            // iconType: LuBlocks,
                        },
                        {
                            name: "Anime Torrents",
                            isCurrent: filterType === "anime-torrent-provider",
                            onClick: () => setFilterType("anime-torrent-provider"),
                            // iconType: RiFolderDownloadFill,
                        },
                        {
                            name: "Manga",
                            isCurrent: filterType === "manga-provider",
                            onClick: () => setFilterType("manga-provider"),
                            // iconType: LuBookOpen,
                        },
                        {
                            name: "Online Streaming",
                            isCurrent: filterType === "onlinestream-provider",
                            onClick: () => setFilterType("onlinestream-provider"),
                            // iconType: CgMediaPodcast,
                        },
                        {
                            name: "Custom Sources",
                            isCurrent: filterType === "custom-source",
                            onClick: () => setFilterType("custom-source"),
                            // iconType: CgMediaPodcast,
                        },
                    ]}
                />

                <div className="flex flex-col lg:flex-row w-full gap-2">
                    <Select
                        value={filterLanguage}
                        onValueChange={setFilterLanguage}
                        options={languageOptions}
                        fieldClass="lg:max-w-[200px]"
                    />
                    <TextInput
                        placeholder="Search extensions..."
                        value={searchTerm}
                        onValueChange={(v) => setSearchTerm(v)}
                        className="pl-10"
                        leftIcon={<BiSearch />}
                    />
                </div>
            </div>

            {isLoadingMarketplace && <LoadingSpinner />}

            {(!marketplaceExtensions && !isLoadingMarketplace) && <LuffyError>
                Could not get marketplace extensions.
            </LuffyError>}

            {(!!marketplaceExtensions && filteredExtensions.length === 0) && (
                <Card className="p-8 text-center">
                    <p className="text-[--muted]">No extensions found matching your criteria.</p>
                </Card>
            )}

            {!!pluginExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><LuBlocks /> Plugins</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {pluginExtensions.map(extension => (
                            <MarketplaceExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!animeTorrentExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><RiFolderDownloadFill />Anime torrents</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {animeTorrentExtensions.map(extension => (
                            <MarketplaceExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!mangaExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><LuBookOpen />Manga</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {mangaExtensions.map(extension => (
                            <MarketplaceExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!onlinestreamExtensions?.length && (
                <Card className="p-4 space-y-6">
                    <h3 className="flex gap-3 items-center"><CgMediaPodcast /> Online streaming</h3>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {onlinestreamExtensions.map(extension => (
                            <MarketplaceExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}

            {!!customSources?.length && (
                <Card className="p-4 space-y-6">
                    <div>
                        <h3 className="flex gap-3 items-center"><MdDataSaverOn /> Custom sources <Popover
                            className="text-sm"
                            trigger={
                                <AiOutlineExclamationCircle className="text-[1.2rem] transition-opacity opacity-45 hover:opacity-90 cursor-pointer" />}
                        >
                            Custom sources do not provide any streaming features. Torrent and online streaming providers are needed for this.
                        </Popover></h3>
                        <p className="text-[--muted] text-sm">
                            Custom sources let you browse media beyond what AniList provides.
                        </p>
                    </div>
                    <div className="grid grid-cols-1 lg:grid-cols-3 2xl:grid-cols-4 gap-4">
                        {customSources.map(extension => (
                            <MarketplaceExtensionCard
                                key={extension.id}
                                extension={extension}
                                isInstalled={isExtensionInstalled(extension.id)}
                            />
                        ))}
                    </div>
                </Card>
            )}
        </AppLayoutStack>
    )
}

type MarketplaceExtensionCardProps = {
    extension: Extension_Extension
    updateData?: Extension_Extension | undefined
    isInstalled: boolean
    hideInstallButton?: boolean
    showType?: boolean
}

export function MarketplaceExtensionCard(props: MarketplaceExtensionCardProps) {

    const {
        extension,
        updateData,
        isInstalled,
        hideInstallButton,
        showType,
        ...rest
    } = props

    const { mutate: reloadExternalExtension, isPending: isReloadingExtension } = useReloadExternalExtension()

    const [installModalOpen, setInstallModalOpen] = React.useState(false)

    const {
        mutate: installExtension,
        data: installResponse,
        isPending: isInstalling,
    } = useInstallExternalExtension()

    React.useEffect(() => {
        if (installResponse) {
            toast.success(installResponse.message)
            setInstallModalOpen(false)
        }
    }, [installResponse])

    return (
        <div
            className={cn(
                "group/extension-card border border-[rgb(255_255_255_/_5%)] relative overflow-hidden",
                "bg-gray-900 rounded-xl p-3",
                !!updateData && "border-[--green]",
            )}
        >
            {!hideInstallButton && <div className="absolute top-3 right-3 z-[2]">
                <div className=" flex flex-row gap-1 z-[2] flex-wrap justify-end">
                    {!isInstalled ? <IconButton
                        size="sm"
                        intent="primary-subtle"
                        icon={<LuDownload />}
                        loading={isInstalling}
                        onClick={() => installExtension({ manifestUri: extension.manifestURI })}
                    /> : <IconButton
                        size="sm"
                        disabled
                        intent="success-subtle"
                        icon={<LuCheck />}
                    />
                    }
                </div>
            </div>}

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className={cn("relative rounded-md size-12 bg-gray-950 overflow-hidden", !!extension.icon && "bg-gray-900")}>
                        {!!extension.icon ? (
                            <SeaImage
                                src={extension.icon}
                                alt="extension icon"
                                crossOrigin="anonymous"
                                fill
                                quality={100}
                                priority
                                className="object-cover"
                            />
                        ) : <div className="w-full h-full flex items-center justify-center">
                            <p className="text-2xl font-bold">
                                {(extension.name[0]).toUpperCase()}
                            </p>
                        </div>}
                    </div>

                    <div>
                        <p className="font-semibold line-clamp-1">
                            {extension.name}
                        </p>
                        <p className="text-xs line-clamp-1 tracking-wide">
                            {showType && <span className="opacity-70">{EXTENSION_TYPE[extension.type]} - </span>}
                            <span className="opacity-30">{extension.id}</span>
                        </p>
                    </div>
                </div>

                {extension.description && (
                    <Popover
                        trigger={<p className="text-sm text-[--muted] line-clamp-2 cursor-pointer">
                            {extension.description}
                        </p>}
                    >
                        <p className="text-sm">
                            {extension.description}
                        </p>
                    </Popover>
                )}

                <div className="flex gap-2 flex-wrap">
                    {!!extension.version && <Badge className="rounded-md tracking-wide">
                        {extension.version}
                    </Badge>}
                    {<Badge className="rounded-md" intent="unstyled">
                        {extension.author}
                    </Badge>}
                    {extension.lang?.toUpperCase() !== "MULTI" && <Badge
                        className="border-transparent rounded-md"
                        intent={extension.lang !== "multi" ? "blue" : "unstyled"}
                    >
                        {/*{extension.lang.toUpperCase()}*/}
                        {LANGUAGES_LIST[extension.lang?.toLowerCase()]?.nativeName || extension.lang?.toUpperCase() || "Unknown"}
                    </Badge>}
                    <Badge className="border-transparent rounded-md text-[--muted] px-0" intent="unstyled">
                        {capitalize(extension.language)}
                    </Badge>
                    {!!updateData && <Badge className="rounded-md" intent="success">
                        Update available
                    </Badge>}
                </div>

            </div>
        </div>
    )
}
