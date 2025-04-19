import { Extension_Extension } from "@/api/generated/types"
import {
    useGetAllExtensions,
    useGetMarketplaceExtensions,
    useInstallExternalExtension,
    useReloadExternalExtension,
} from "@/api/hooks/extensions.hooks"
import { DEFAULT_MARKETPLACE_URL, marketplaceUrlAtom } from "@/app/(main)/extensions/_lib/marketplace.atoms"
import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
import { LuffyError } from "@/components/shared/luffy-error"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Badge } from "@/components/ui/badge"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useAtom } from "jotai/react"
import { orderBy } from "lodash"
import capitalize from "lodash/capitalize"
import Image from "next/image"
import React, { useMemo } from "react"
import { BiSearch } from "react-icons/bi"
import { CgMediaPodcast } from "react-icons/cg"
import { LuBlocks, LuCheck, LuDownload, LuSettings } from "react-icons/lu"
import { PiBookFill } from "react-icons/pi"
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
            {/* URL Change Modal */}
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

            {/* Search and filter */}
            <div className="flex flex-wrap gap-4">

                <div className="flex flex-col lg:flex-row w-full gap-2">
                    <Select
                        value={filterType}
                        onValueChange={setFilterType}
                        options={[
                            { value: "all", label: "All Types" },
                            { value: "plugin", label: "Plugins" },
                            { value: "anime-torrent-provider", label: "Anime Torrents" },
                            { value: "manga-provider", label: "Manga" },
                            { value: "onlinestream-provider", label: "Online Streaming" },
                        ]}
                        fieldClass="lg:max-w-[200px]"
                    />
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
                {/*<SelectTrigger className="w-[180px]">*/}
                {/*    <SelectValue placeholder="Filter by type" />*/}
                {/*</SelectTrigger>*/}
                {/*<SelectContent>*/}
                {/*    <SelectItem value="all">All Types</SelectItem>*/}
                {/*    <SelectItem value="plugin">Plugins</SelectItem>*/}
                {/*    <SelectItem value="anime-torrent-provider">Anime Torrent</SelectItem>*/}
                {/*    <SelectItem value="manga-provider">Manga</SelectItem>*/}
                {/*    <SelectItem value="onlinestream-provider">Online Streaming</SelectItem>*/}
                {/*</SelectContent>*/}
                {/*</Select>*/}
            </div>

            {isLoadingMarketplace && <LoadingSpinner />}

            {(!marketplaceExtensions && !isLoadingMarketplace) && <LuffyError>
                Could not get marketplace extensions.
            </LuffyError>}

            {/* No results message */}
            {(!!marketplaceExtensions && filteredExtensions.length === 0) && (
                <Card className="p-8 text-center">
                    <p className="text-[--muted]">No extensions found matching your criteria.</p>
                </Card>
            )}

            {/* Display extensions by type */}
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
                    <h3 className="flex gap-3 items-center"><PiBookFill />Manga</h3>
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
        </AppLayoutStack>
    )
}

type MarketplaceExtensionCardProps = {
    extension: Extension_Extension
    updateData?: Extension_Extension | undefined
    isInstalled: boolean
}

function MarketplaceExtensionCard(props: MarketplaceExtensionCardProps) {

    const {
        extension,
        updateData,
        isInstalled,
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
                "bg-gray-950 rounded-md p-3",
                !!updateData && "border-[--green]",
            )}
        >
            <div className="absolute top-3 right-3 z-[2]">
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
            </div>

            <div className="z-[1] relative space-y-3">
                <div className="flex gap-3 pr-16">
                    <div className="relative rounded-md size-12 bg-gray-900 overflow-hidden">
                        {!!extension.icon ? (
                            <Image
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
                        <p className="opacity-30 text-xs line-clamp-1 tracking-wide">
                            {extension.id}
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
                        By {extension.author}
                    </Badge>}
                    <Badge className="rounded-md" intent={extension.lang !== "multi" && extension.lang !== "en" ? "blue" : "unstyled"}>
                        {/*{extension.lang.toUpperCase()}*/}
                        {LANGUAGES_LIST[extension.lang?.toLowerCase()]?.name || extension.lang?.toUpperCase() || "Unknown"}
                    </Badge>
                    <Badge className="rounded-md" intent="unstyled">
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
