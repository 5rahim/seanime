import { useServerMutation } from "@/api/client/requests"
import { UpdateTheme_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { Models_Theme } from "@/api/generated/types"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useCheckForUpdates } from "@/api/hooks/releases.hooks"
import { useSaveSettings } from "@/api/hooks/settings.hooks"
import { useGetTorrentstreamSettings } from "@/api/hooks/torrentstream.hooks"
import { electronUpdateModalOpenAtom } from "@/app/(main)/_electron/electron-update-modal"
import { CustomLibraryBanner } from "@/app/(main)/_features/anime-library/_containers/custom-library-banner"
import { __issueReport_overlayOpenAtom } from "@/app/(main)/_features/issue-report/issue-report"
import { updateModalOpenAtom as webUpdateModalOpenAtom } from "@/app/(main)/_features/update/update-modal"
import { useServerDisabledFeatures, useServerStatus, useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ExternalPlayerLinkSettings, MediaplayerSettings } from "@/app/(main)/settings/_components/mediaplayer-settings"
import { PlaybackSettings } from "@/app/(main)/settings/_components/playback-settings"
import { __settings_tabAtom } from "@/app/(main)/settings/_components/settings-page.atoms"
import { SettingsIsDirty, SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { AnimeLibrarySettings } from "@/app/(main)/settings/_containers/anime-library-settings"
import { DebridSettings } from "@/app/(main)/settings/_containers/debrid-settings"
import { FilecacheSettings } from "@/app/(main)/settings/_containers/filecache-settings"
import { LogsSettings } from "@/app/(main)/settings/_containers/logs-settings"
import { MangaSettings } from "@/app/(main)/settings/_containers/manga-settings"
import { MediastreamSettings } from "@/app/(main)/settings/_containers/mediastream-settings"
import { ServerSettings } from "@/app/(main)/settings/_containers/server-settings"
import { TorrentstreamSettings } from "@/app/(main)/settings/_containers/torrentstream-settings"
import { UISettings } from "@/app/(main)/settings/_containers/ui-settings"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Field, Form } from "@/components/ui/form"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { useRouter, useSearchParams } from "@/lib/navigation"
import { DEFAULT_TORRENT_CLIENT, DEFAULT_TORRENT_PROVIDER, settingsSchema, TORRENT_PROVIDER } from "@/lib/server/settings"
import { THEME_DEFAULT_VALUES } from "@/lib/theme/theme-hooks"
import { __isElectronDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { BiDonateHeart } from "react-icons/bi"
import { CgMediaPodcast } from "react-icons/cg"
import { FaDiscord } from "react-icons/fa"
import { HiOutlineServerStack } from "react-icons/hi2"
import {
    LuBookKey,
    LuBookOpen,
    LuCircleArrowOutUpRight,
    LuCirclePlay,
    LuFileSearch,
    LuLibrary,
    LuMonitor,
    LuMonitorPlay,
    LuPalette,
    LuTabletSmartphone,
    LuWandSparkles,
} from "react-icons/lu"
import { LuRefreshCw } from "react-icons/lu"
import { MdOutlineConnectWithoutContact, MdOutlineDownloading, MdOutlinePalette } from "react-icons/md"
import { RiFolderDownloadFill } from "react-icons/ri"
import { SiBittorrent, SiQbittorrent, SiTransmission } from "react-icons/si"
import { TbDatabaseExclamation } from "react-icons/tb"
import { VscDebugAlt } from "react-icons/vsc"
import { toast } from "sonner"
import { SettingsCard, SettingsNavCard, SettingsPageHeader } from "./_components/settings-card"
import { DenshiSettings } from "./_containers/denshi-settings"
import { DiscordRichPresenceSettings } from "./_containers/discord-rich-presence-settings"
import { LocalSettings } from "./_containers/local-settings"
import { NakamaSettings } from "./_containers/nakama-settings"

const tabContentClass = cn(
    "space-y-8 animate-in fade-in-0 duration-400",
)


export default function Page() {
    const status = useServerStatus()
    const { isFeatureDisabled, showFeatureWarning } = useServerDisabledFeatures()
    const setServerStatus = useSetServerStatus()
    const router = useRouter()
    const queryClient = useQueryClient()

    const searchParams = useSearchParams()

    const { mutateAsync: saveSettings, data, isPending } = useSaveSettings()

    const [tab, setTab] = useAtom(__settings_tabAtom)
    const formRef = React.useRef<UseFormReturn<any>>(null)

    const { data: torrentProviderExtensions } = useAnimeListTorrentProviderExtensions()

    const { data: torrentstreamSettings } = useGetTorrentstreamSettings()

    const { mutateAsync: saveThemeSettings } = useServerMutation<Models_Theme, UpdateTheme_Variables>({
        endpoint: API_ENDPOINTS.THEME.UpdateTheme.endpoint,
        method: API_ENDPOINTS.THEME.UpdateTheme.methods[0],
        mutationKey: [API_ENDPOINTS.THEME.UpdateTheme.key, "settings-page"],
        onSuccess: async () => {
            await queryClient.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetStatus.key] })
        },
    })

    const { mutate: openInExplorer, isPending: isOpening } = useOpenInExplorer()

    const { mutate: checkForUpdates, isPending: isCheckingForUpdates } = useCheckForUpdates()
    const setWebUpdateModalOpen = useSetAtom(webUpdateModalOpenAtom)
    const setElectronUpdateModalOpen = useSetAtom(electronUpdateModalOpenAtom)

    React.useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
        }
    }, [data, isPending])

    const setIssueRecorderOpen = useSetAtom(__issueReport_overlayOpenAtom)

    function handleOpenIssueRecorder() {
        if (isFeatureDisabled("UpdateSettings")) return showFeatureWarning()

        setIssueRecorderOpen(true)
        router.push("/")
    }

    const previousTab = React.useRef(tab)
    React.useEffect(() => {
        if (tab !== previousTab.current) {
            previousTab.current = tab
            formRef.current?.reset()
        }
    }, [tab])

    React.useEffect(() => {
        const initialTab = searchParams.get("tab")
        if (initialTab) {
            setTab(initialTab)
            setTimeout(() => {
                // Remove search param
                if (searchParams.has("tab")) {
                    const newParams = new URLSearchParams(searchParams)
                    newParams.delete("tab")
                    router.replace(`?${newParams.toString()}`, { scroll: false })
                }
            }, 500)
        }
    }, [searchParams])

    if (!status?.settings) return <LoadingSpinner />

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper data-settings-page-container className="p-4 sm:p-8 space-y-4 relative">
                {/*<Separator/>*/}


                {/*<Card className="p-0 overflow-hidden">*/}
                <Tabs
                    value={tab}
                    onValueChange={setTab}
                    variant="pill"
                    className={cn("w-full grid grid-cols-1 lg:grid lg:grid-cols-[300px,1fr] gap-4")}
                    triggerClass={cn(
                        "text-base font-medium w-fit lg:w-full border-0 data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white py-0",
                        "h-9 lg:justify-start px-3 transition-all duration-200 hover:bg-[--subtle]/50 hover:transform rounded-lg",
                    )}
                    listClass={cn(
                        "w-full flex flex-wrap lg:flex-nowrap h-fit",
                        "lg:block p-2 lg:p-0",
                    )}
                    data-settings-page-tabs
                >
                    <TabsList variant="none" className="flex-wrap max-w-full lg:space-y-2 lg:sticky lg:top-10">
                        <SettingsNavCard>
                            <div className="overflow-x-none overflow-y-hidden rounded-[--radius-md] space-y-1 lg:space-y-3 lg:block">

                                <Card className="bg-transparent border-transparent">
                                    <div className="space-y-2 p-0 w-full">
                                        <h4 className=" text-xl font-bold text-center">Settings</h4>

                                    </div>
                                </Card>
                                <Card className="block border-0 bg-transparent lg:border lg:bg-[--paper] overflow-clip p-1">
                                    <TabsTrigger
                                        value="seanime"
                                        className="group"
                                    ><LuWandSparkles className="text-base mr-2 transition-transform duration-200" /> App</TabsTrigger>
                                    <TabsTrigger
                                        value="ui"
                                        className="group"
                                    ><MdOutlinePalette className="text-base mr-2 transition-transform duration-200" /> User Interface</TabsTrigger>
                                    {/* <TabsTrigger
                                     value="local"
                                     className="group"
                                     ><LuUserCog className="text-base mr-2 transition-transform duration-200" /> Local Account</TabsTrigger> */}
                                    <TabsTrigger
                                        value="library"
                                        className="group"
                                    ><LuLibrary className="text-base mr-2 transition-transform duration-200" /> Local Anime Library</TabsTrigger>
                                </Card>

                                {/*<div className="text-sm lg:text-[--foreground] py-1.5 px-3 tracking-wide font-medium hidden lg:block">*/}
                                {/*    Anime playback*/}
                                {/*</div>*/}

                                <Card className="contents lg:block border-0 bg-transparent lg:border lg:bg-[--paper] overflow-clip p-1">
                                    <TabsTrigger
                                        value="playback"
                                        className="group"
                                    ><LuCirclePlay className="text-base mr-2 transition-transform duration-200" /> Video Playback</TabsTrigger>

                                    <TabsTrigger
                                        value="media-player"
                                        className="group"
                                    ><LuMonitorPlay className="text-base mr-2 transition-transform duration-200" /> Desktop Media Player</TabsTrigger>
                                    <TabsTrigger
                                        value="external-player-link"
                                        className="group"
                                    ><LuCircleArrowOutUpRight className="text-base mr-2 transition-transform duration-200" /> External Player
                                                                                                                              Link</TabsTrigger>
                                    <TabsTrigger
                                        value="mediastream"
                                        className="relative group"
                                    ><LuTabletSmartphone className="text-base mr-2 transition-transform duration-200" /> Transcoding / Direct
                                                                                                                         Play</TabsTrigger>
                                </Card>

                                {/*<div className="text-sm lg:text-[--foreground] py-1.5 px-3 tracking-wide font-medium hidden lg:block">*/}
                                {/*    Torrenting*/}
                                {/*</div>*/}

                                <Card className="contents lg:block border-0 bg-transparent lg:border lg:bg-[--paper] overflow-clip p-1">
                                    <TabsTrigger
                                        value="torrent"
                                        className="group"
                                    ><LuFileSearch className="text-base mr-2 transition-transform duration-200" /> Torrent Provider</TabsTrigger>
                                    <TabsTrigger
                                        value="torrent-client"
                                        className="group"
                                    ><MdOutlineDownloading className="text-base mr-2 transition-transform duration-200" /> Torrent
                                                                                                                           Client</TabsTrigger>
                                    <TabsTrigger
                                        value="torrentstream"
                                        className="relative group"
                                    ><SiBittorrent className="text-base mr-2 transition-transform duration-200" /> Torrent Streaming</TabsTrigger>
                                    <TabsTrigger
                                        value="debrid"
                                        className="group"
                                    ><HiOutlineServerStack className="text-base mr-2 transition-transform duration-200" /> Debrid
                                                                                                                           Service</TabsTrigger>
                                </Card>

                                {/*<div className="text-sm lg:text-[--foreground] py-1.5 px-3 tracking-wide font-medium hidden lg:block">*/}
                                {/*    Other features*/}
                                {/*</div>*/}

                                <Card className="contents lg:block border-0 bg-transparent lg:border lg:bg-[--paper] overflow-clip p-1">
                                    <TabsTrigger
                                        value="onlinestream"
                                        className="group"
                                    ><CgMediaPodcast className="text-base mr-2 transition-transform duration-200" /> Online Streaming</TabsTrigger>

                                    <TabsTrigger
                                        value="manga"
                                        className="group"
                                    ><LuBookOpen className="text-base mr-2 transition-transform duration-200" /> Manga</TabsTrigger>
                                    <TabsTrigger
                                        value="nakama"
                                        className="group relative"
                                    ><MdOutlineConnectWithoutContact className="text-base mr-2 transition-transform duration-200" /> Nakama</TabsTrigger>
                                    <TabsTrigger
                                        value="discord"
                                        className="group"
                                    ><FaDiscord className="text-base mr-2 transition-transform duration-200" /> Discord</TabsTrigger>
                                </Card>

                                {/*<div className="text-sm lg:text-[--foreground] py-1.5 px-3 tracking-wide font-medium hidden lg:block">*/}
                                {/*    App*/}
                                {/*</div>*/}

                                <Card className="contents lg:block border-0 bg-transparent lg:border lg:bg-[--paper] overflow-clip p-1">
                                    {__isElectronDesktop__ && (
                                        <TabsTrigger
                                            value="denshi"
                                            className="group"
                                        ><LuMonitor className="text-base mr-2 transition-transform duration-200" /> Denshi</TabsTrigger>
                                    )}
                                    {/* <TabsTrigger
                                     value="cache"
                                     className="group"
                                     ><TbDatabaseExclamation className="text-base mr-2 transition-transform duration-200" /> Cache</TabsTrigger> */}
                                    <TabsTrigger
                                        value="logs"
                                        className="group"
                                    ><LuBookKey className="text-base mr-2 transition-transform duration-200" /> Logs & Cache</TabsTrigger>
                                </Card>
                            </div>
                        </SettingsNavCard>

                        <div className="space-y-3">
                            <div className="space-y-1">
                                <p className="text-[--muted] text-xs w-full text-center">
                                    <span className="font-semibold">{status?.version}</span> {status?.versionName} • {capitalize(status?.os)}{__isElectronDesktop__ &&
                                    <span className="font-medium"> • Denshi</span>}
                                </p>
                                <p className="text-[--muted] text-sm w-full">

                                </p>
                            </div>

                            <div className="flex justify-center !mt-0 pb-4">
                                <SeaLink
                                    href="https://github.com/sponsors/5rahim"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    <Button
                                        intent="gray-link"
                                        size="md"
                                        leftIcon={<BiDonateHeart className="text-lg" />}
                                    >
                                        Donate
                                    </Button>
                                </SeaLink>
                            </div>
                        </div>
                    </TabsList>

                    <div className="">
                        <Form
                            key={`${status?.settings?.updatedAt ?? "settings"}:${status?.themeSettings?.updatedAt ?? "theme"}`}
                            schema={settingsSchema}
                            mRef={formRef}
                            onSubmit={async data => {
                                await saveSettings({
                                    library: {
                                        libraryPath: data.libraryPath,
                                        autoUpdateProgress: data.autoUpdateProgress,
                                        disableUpdateCheck: data.disableUpdateCheck,
                                        torrentProvider: data.torrentProvider,
                                        autoSelectTorrentProvider: data.autoSelectTorrentProvider,
                                        autoScan: data.autoScan,
                                        enableOnlinestream: data.enableOnlinestream,
                                        includeOnlineStreamingInLibrary: data.includeOnlineStreamingInLibrary ?? false,
                                        disableAnimeCardTrailers: data.disableAnimeCardTrailers,
                                        enableManga: data.enableManga,
                                        dohProvider: data.dohProvider === "-" ? "" : data.dohProvider,
                                        openTorrentClientOnStart: data.openTorrentClientOnStart,
                                        openWebURLOnStart: data.openWebURLOnStart,
                                        refreshLibraryOnStart: data.refreshLibraryOnStart,
                                        autoPlayNextEpisode: data.autoPlayNextEpisode ?? false,
                                        enableWatchContinuity: data.enableWatchContinuity ?? false,
                                        libraryPaths: data.libraryPaths ?? [],
                                        autoSyncOfflineLocalData: data.autoSyncOfflineLocalData ?? false,
                                        scannerMatchingThreshold: data.scannerMatchingThreshold,
                                        scannerMatchingAlgorithm: data.scannerMatchingAlgorithm === "-" ? "" : data.scannerMatchingAlgorithm,
                                        autoSyncToLocalAccount: data.autoSyncToLocalAccount ?? false,
                                        autoSaveCurrentMediaOffline: data.autoSaveCurrentMediaOffline ?? false,
                                        useFallbackMetadataProvider: data.useFallbackMetadataProvider ?? false,
                                        scannerUseLegacyMatching: data.scannerUseLegacyMatching ?? false,
                                        scannerConfig: data.scannerConfig ?? "",
                                        updateChannel: data.updateChannel || "github",
                                        enableExtensionSecureMode: data.enableExtensionSecureMode ?? false,
                                        defaultPlaybackSource: data.defaultPlaybackSource === "-" ? "" : data.defaultPlaybackSource,
                                        showTorrentAvailability: data.showTorrentAvailability ?? false,
                                    },
                                    nakama: {
                                        enabled: data.nakamaEnabled ?? false,
                                        username: data.nakamaUsername,
                                        isHost: data.nakamaIsHost ?? false,
                                        remoteServerURL: data.nakamaRemoteServerURL,
                                        remoteServerPassword: data.nakamaRemoteServerPassword,
                                        hostShareLocalAnimeLibrary: data.nakamaHostShareLocalAnimeLibrary ?? false,
                                        hostPassword: data.nakamaHostPassword,
                                        includeNakamaAnimeLibrary: data.includeNakamaAnimeLibrary ?? false,
                                        hostUnsharedAnimeIds: data?.nakamaHostUnsharedAnimeIds ?? [],
                                        hostEnablePortForwarding: data.nakamaHostEnablePortForwarding ?? false,
                                    },
                                    manga: {
                                        defaultMangaProvider: data.defaultMangaProvider === "-" ? "" : data.defaultMangaProvider,
                                        mangaAutoUpdateProgress: data.mangaAutoUpdateProgress ?? false,
                                        mangaLocalSourceDirectory: data.mangaLocalSourceDirectory || "",
                                    },
                                    mediaPlayer: {
                                        host: data.mediaPlayerHost,
                                        defaultPlayer: data.defaultPlayer,
                                        vlcPort: data.vlcPort,
                                        vlcUsername: data.vlcUsername || "",
                                        vlcPassword: data.vlcPassword,
                                        vlcPath: data.vlcPath || "",
                                        mpcPort: data.mpcPort,
                                        mpcPath: data.mpcPath || "",
                                        mpvSocket: data.mpvSocket || "",
                                        mpvPath: data.mpvPath || "",
                                        mpvArgs: data.mpvArgs || "",
                                        iinaSocket: data.iinaSocket || "",
                                        iinaPath: data.iinaPath || "",
                                        iinaArgs: data.iinaArgs || "",
                                        vcTranslate: data.vcTranslate ?? false,
                                        vcTranslateApiKey: data.vcTranslateApiKey || "",
                                        vcTranslateProvider: data.vcTranslateProvider || "",
                                        vcTranslateTargetLanguage: data.vcTranslateTargetLanguage || "",
                                        vcTranslateBaseUrl: data.vcTranslateBaseUrl || "",
                                        vcTranslateModel: data.vcTranslateModel || "",
                                        mpvPrismLogging: data.mpvPrismLogging ?? false,
                                        mpvPrismEnabled: data.mpvPrismEnabled ?? false,
                                        screenshotDir: data.screenshotDir || "",
                                    },
                                    torrent: {
                                        defaultTorrentClient: data.defaultTorrentClient,
                                        qbittorrentPath: data.qbittorrentPath,
                                        qbittorrentHost: data.qbittorrentHost,
                                        qbittorrentPort: data.qbittorrentPort,
                                        qbittorrentPassword: data.qbittorrentPassword,
                                        qbittorrentUsername: data.qbittorrentUsername,
                                        qbittorrentTags: data.qbittorrentTags,
                                        qbittorrentCategory: data.qbittorrentCategory,
                                        transmissionPath: data.transmissionPath,
                                        transmissionHost: data.transmissionHost,
                                        transmissionPort: data.transmissionPort,
                                        transmissionUsername: data.transmissionUsername,
                                        transmissionPassword: data.transmissionPassword,
                                        seanimePort: data.seanimePort,
                                        seanimeMaxConnections: data.seanimeMaxConnections,
                                        seanimeDownloadLimit: data.seanimeDownloadLimit,
                                        seanimeUploadLimit: data.seanimeUploadLimit,
                                        seanimeMaxActiveDownloads: data.seanimeMaxActiveDownloads,
                                        showActiveTorrentCount: data.showActiveTorrentCount ?? false,
                                        hideTorrentList: data.hideTorrentList ?? false,
                                    },
                                    discord: {
                                        enableRichPresence: data?.enableRichPresence ?? false,
                                        enableAnimeRichPresence: data?.enableAnimeRichPresence ?? false,
                                        enableMangaRichPresence: data?.enableMangaRichPresence ?? false,
                                        richPresenceHideSeanimeRepositoryButton: data?.richPresenceHideSeanimeRepositoryButton ?? false,
                                        richPresenceShowAniListMediaButton: data?.richPresenceShowAniListMediaButton ?? false,
                                        richPresenceShowAniListProfileButton: data?.richPresenceShowAniListProfileButton ?? false,
                                        richPresenceUseMediaTitleStatus: data?.richPresenceUseMediaTitleStatus ?? false,
                                    },
                                    anilist: {
                                        hideAudienceScore: data.hideAudienceScore,
                                        enableAdultContent: data.enableAdultContent,
                                        blurAdultContent: data.blurAdultContent,
                                        disableCacheLayer: data.disableCacheLayer,
                                    },
                                    notifications: {
                                        disableNotifications: data?.disableNotifications ?? false,
                                        disableAutoDownloaderNotifications: data?.disableAutoDownloaderNotifications ?? false,
                                        disableAutoScannerNotifications: data?.disableAutoScannerNotifications ?? false,
                                    },
                                })

                                const prevTheme = status?.themeSettings ?? { id: 0, ...THEME_DEFAULT_VALUES }
                                const shouldSaveSpoilerSettings =
                                    data.hideAnimeSpoilers !== prevTheme.hideAnimeSpoilers
                                    || data.hideAnimeSpoilerThumbnails !== prevTheme.hideAnimeSpoilerThumbnails
                                    || data.hideAnimeSpoilerTitles !== prevTheme.hideAnimeSpoilerTitles
                                    || data.hideAnimeSpoilerDescriptions !== prevTheme.hideAnimeSpoilerDescriptions
                                    || data.hideAnimeSpoilerSkipNextEpisode !== prevTheme.hideAnimeSpoilerSkipNextEpisode

                                if (shouldSaveSpoilerSettings) {
                                    await saveThemeSettings({
                                        theme: {
                                            ...prevTheme,
                                            hideAnimeSpoilers: data.hideAnimeSpoilers ?? false,
                                            hideAnimeSpoilerThumbnails: data.hideAnimeSpoilerThumbnails ?? true,
                                            hideAnimeSpoilerTitles: data.hideAnimeSpoilerTitles ?? true,
                                            hideAnimeSpoilerDescriptions: data.hideAnimeSpoilerDescriptions ?? true,
                                            hideAnimeSpoilerSkipNextEpisode: data.hideAnimeSpoilerSkipNextEpisode ?? false,
                                        },
                                    })
                                }

                                formRef.current?.reset(formRef.current.getValues())

                                if (__isElectronDesktop__ && window.electron?.denshiSettings) {
                                    const denshiSettings = await window.electron.denshiSettings.get()
                                    await window.electron.denshiSettings.set({
                                        ...denshiSettings,
                                        updateChannel: data.updateChannel || "github",
                                    })
                                }
                            }}
                            defaultValues={{
                                libraryPath: status?.settings?.library?.libraryPath,
                                mediaPlayerHost: status?.settings?.mediaPlayer?.host,
                                torrentProvider: status?.settings?.library?.torrentProvider || DEFAULT_TORRENT_PROVIDER, // (Backwards compatibility)
                                autoSelectTorrentProvider: status?.settings?.library?.autoSelectTorrentProvider || DEFAULT_TORRENT_PROVIDER, // (Backwards
                                // compatibility)
                                autoScan: status?.settings?.library?.autoScan,
                                defaultPlayer: status?.settings?.mediaPlayer?.defaultPlayer,
                                vlcPort: status?.settings?.mediaPlayer?.vlcPort,
                                vlcUsername: status?.settings?.mediaPlayer?.vlcUsername,
                                vlcPassword: status?.settings?.mediaPlayer?.vlcPassword,
                                vlcPath: status?.settings?.mediaPlayer?.vlcPath,
                                mpcPort: status?.settings?.mediaPlayer?.mpcPort,
                                mpcPath: status?.settings?.mediaPlayer?.mpcPath,
                                mpvSocket: status?.settings?.mediaPlayer?.mpvSocket,
                                mpvPath: status?.settings?.mediaPlayer?.mpvPath,
                                mpvArgs: status?.settings?.mediaPlayer?.mpvArgs,
                                iinaSocket: status?.settings?.mediaPlayer?.iinaSocket,
                                iinaPath: status?.settings?.mediaPlayer?.iinaPath,
                                iinaArgs: status?.settings?.mediaPlayer?.iinaArgs,
                                defaultTorrentClient: status?.settings?.torrent?.defaultTorrentClient || DEFAULT_TORRENT_CLIENT, // (Backwards
                                // compatibility)
                                hideTorrentList: status?.settings?.torrent?.hideTorrentList ?? false,
                                qbittorrentPath: status?.settings?.torrent?.qbittorrentPath,
                                qbittorrentHost: status?.settings?.torrent?.qbittorrentHost,
                                qbittorrentPort: status?.settings?.torrent?.qbittorrentPort,
                                qbittorrentPassword: status?.settings?.torrent?.qbittorrentPassword,
                                qbittorrentUsername: status?.settings?.torrent?.qbittorrentUsername,
                                qbittorrentTags: status?.settings?.torrent?.qbittorrentTags,
                                qbittorrentCategory: status?.settings?.torrent?.qbittorrentCategory,
                                transmissionPath: status?.settings?.torrent?.transmissionPath,
                                transmissionHost: status?.settings?.torrent?.transmissionHost,
                                transmissionPort: status?.settings?.torrent?.transmissionPort,
                                transmissionUsername: status?.settings?.torrent?.transmissionUsername,
                                transmissionPassword: status?.settings?.torrent?.transmissionPassword,
                                seanimePort: status?.settings?.torrent?.seanimePort || 50007,
                                seanimeMaxConnections: status?.settings?.torrent?.seanimeMaxConnections || 50,
                                seanimeDownloadLimit: status?.settings?.torrent?.seanimeDownloadLimit ?? 0,
                                seanimeUploadLimit: status?.settings?.torrent?.seanimeUploadLimit ?? 0,
                                seanimeMaxActiveDownloads: status?.settings?.torrent?.seanimeMaxActiveDownloads || 3,
                                hideAudienceScore: status?.settings?.anilist?.hideAudienceScore ?? false,
                                autoUpdateProgress: status?.settings?.library?.autoUpdateProgress ?? false,
                                disableUpdateCheck: status?.settings?.library?.disableUpdateCheck ?? false,
                                enableOnlinestream: status?.settings?.library?.enableOnlinestream ?? false,
                                includeOnlineStreamingInLibrary: status?.settings?.library?.includeOnlineStreamingInLibrary ?? false,
                                disableAnimeCardTrailers: status?.settings?.library?.disableAnimeCardTrailers ?? false,
                                enableManga: status?.settings?.library?.enableManga ?? false,
                                enableRichPresence: status?.settings?.discord?.enableRichPresence ?? false,
                                enableAnimeRichPresence: status?.settings?.discord?.enableAnimeRichPresence ?? false,
                                enableMangaRichPresence: status?.settings?.discord?.enableMangaRichPresence ?? false,
                                enableAdultContent: status?.settings?.anilist?.enableAdultContent ?? false,
                                blurAdultContent: status?.settings?.anilist?.blurAdultContent ?? false,
                                dohProvider: status?.settings?.library?.dohProvider || "-",
                                openTorrentClientOnStart: status?.settings?.library?.openTorrentClientOnStart ?? false,
                                openWebURLOnStart: status?.settings?.library?.openWebURLOnStart ?? false,
                                refreshLibraryOnStart: status?.settings?.library?.refreshLibraryOnStart ?? false,
                                richPresenceHideSeanimeRepositoryButton: status?.settings?.discord?.richPresenceHideSeanimeRepositoryButton ?? false,
                                richPresenceShowAniListMediaButton: status?.settings?.discord?.richPresenceShowAniListMediaButton ?? false,
                                richPresenceShowAniListProfileButton: status?.settings?.discord?.richPresenceShowAniListProfileButton ?? false,
                                richPresenceUseMediaTitleStatus: status?.settings?.discord?.richPresenceUseMediaTitleStatus ?? false,
                                disableNotifications: status?.settings?.notifications?.disableNotifications ?? false,
                                disableAutoDownloaderNotifications: status?.settings?.notifications?.disableAutoDownloaderNotifications ?? false,
                                disableAutoScannerNotifications: status?.settings?.notifications?.disableAutoScannerNotifications ?? false,
                                defaultMangaProvider: status?.settings?.manga?.defaultMangaProvider || "-",
                                mangaAutoUpdateProgress: status?.settings?.manga?.mangaAutoUpdateProgress ?? false,
                                showActiveTorrentCount: status?.settings?.torrent?.showActiveTorrentCount ?? false,
                                autoPlayNextEpisode: status?.settings?.library?.autoPlayNextEpisode ?? false,
                                enableWatchContinuity: status?.settings?.library?.enableWatchContinuity ?? false,
                                libraryPaths: status?.settings?.library?.libraryPaths ?? [],
                                autoSyncOfflineLocalData: status?.settings?.library?.autoSyncOfflineLocalData ?? false,
                                scannerMatchingThreshold: status?.settings?.library?.scannerMatchingThreshold ?? 0.5,
                                scannerMatchingAlgorithm: status?.settings?.library?.scannerMatchingAlgorithm || "-",
                                mangaLocalSourceDirectory: status?.settings?.manga?.mangaLocalSourceDirectory || "",
                                autoSyncToLocalAccount: status?.settings?.library?.autoSyncToLocalAccount ?? false,
                                nakamaEnabled: status?.settings?.nakama?.enabled ?? false,
                                nakamaUsername: status?.settings?.nakama?.username ?? "",
                                nakamaIsHost: status?.settings?.nakama?.isHost ?? false,
                                nakamaRemoteServerURL: status?.settings?.nakama?.remoteServerURL ?? "",
                                nakamaRemoteServerPassword: status?.settings?.nakama?.remoteServerPassword ?? "",
                                nakamaHostShareLocalAnimeLibrary: status?.settings?.nakama?.hostShareLocalAnimeLibrary ?? false,
                                nakamaHostPassword: status?.settings?.nakama?.hostPassword ?? "",
                                includeNakamaAnimeLibrary: status?.settings?.nakama?.includeNakamaAnimeLibrary ?? false,
                                nakamaHostUnsharedAnimeIds: status?.settings?.nakama?.hostUnsharedAnimeIds ?? [],
                                autoSaveCurrentMediaOffline: status?.settings?.library?.autoSaveCurrentMediaOffline ?? false,
                                useFallbackMetadataProvider: status?.settings?.library?.useFallbackMetadataProvider ?? false,
                                vcTranslate: status?.settings?.mediaPlayer?.vcTranslate ?? false,
                                vcTranslateApiKey: status?.settings?.mediaPlayer?.vcTranslateApiKey ?? "",
                                vcTranslateProvider: status?.settings?.mediaPlayer?.vcTranslateProvider ?? "",
                                vcTranslateTargetLanguage: status?.settings?.mediaPlayer?.vcTranslateTargetLanguage ?? "",
                                vcTranslateBaseUrl: status?.settings?.mediaPlayer?.vcTranslateBaseUrl ?? "",
                                vcTranslateModel: status?.settings?.mediaPlayer?.vcTranslateModel ?? "",
                                mpvPrismLogging: status?.settings?.mediaPlayer?.mpvPrismLogging ?? false,
                                mpvPrismEnabled: status?.settings?.mediaPlayer?.mpvPrismEnabled ?? false,
                                screenshotDir: status?.settings?.mediaPlayer?.screenshotDir ?? "",
                                scannerUseLegacyMatching: status?.settings?.library?.scannerUseLegacyMatching ?? false,
                                scannerConfig: status?.settings?.library?.scannerConfig ?? "",
                                updateChannel: status?.settings?.library?.updateChannel || "github",
                                enableExtensionSecureMode: status?.settings?.library?.enableExtensionSecureMode ?? false,
                                defaultPlaybackSource: status?.settings?.library?.defaultPlaybackSource || "-",
                                hideAnimeSpoilers: status?.themeSettings?.hideAnimeSpoilers ?? THEME_DEFAULT_VALUES.hideAnimeSpoilers,
                                hideAnimeSpoilerThumbnails: status?.themeSettings?.hideAnimeSpoilerThumbnails ?? THEME_DEFAULT_VALUES.hideAnimeSpoilerThumbnails,
                                hideAnimeSpoilerTitles: status?.themeSettings?.hideAnimeSpoilerTitles ?? THEME_DEFAULT_VALUES.hideAnimeSpoilerTitles,
                                hideAnimeSpoilerDescriptions: status?.themeSettings?.hideAnimeSpoilerDescriptions ?? THEME_DEFAULT_VALUES.hideAnimeSpoilerDescriptions,
                                hideAnimeSpoilerSkipNextEpisode: status?.themeSettings?.hideAnimeSpoilerSkipNextEpisode ?? THEME_DEFAULT_VALUES.hideAnimeSpoilerSkipNextEpisode,
                                showTorrentAvailability: status?.settings?.library?.showTorrentAvailability ?? false,
                            }}
                            stackClass="space-y-0 relative"
                        >
                            {(f) => {
                                const selectedTorrentProvider = torrentProviderExtensions?.find(ext => ext.id === f.watch("torrentProvider"))
                                const torrentProviderMissing = !!torrentProviderExtensions && !selectedTorrentProvider

                                return <>
                                    <SettingsIsDirty />
                                    <TabsContent value="seanime" className={tabContentClass}>

                                        <div className="space-y-3">
                                            <SettingsPageHeader
                                                title="App"
                                                description="General app settings"
                                                icon={LuWandSparkles}
                                            />

                                            <div className="flex flex-wrap gap-2 slide-in-from-bottom duration-500 delay-150">
                                                {!!status?.dataDir && <Button
                                                    size="sm"
                                                    intent="gray-outline"
                                                    onClick={() => openInExplorer({
                                                        path: status?.dataDir,
                                                    })}
                                                    className="transition-all duration-200 hover:scale-105 hover:shadow-md"
                                                    leftIcon={
                                                        <RiFolderDownloadFill className="transition-transform duration-200 group-hover:scale-110" />}
                                                >
                                                    Open Data directory
                                                </Button>}
                                                <Button
                                                    size="sm"
                                                    intent="gray-outline"
                                                    onClick={handleOpenIssueRecorder}
                                                    leftIcon={<VscDebugAlt className="transition-transform duration-200 group-hover:scale-110" />}
                                                    className="transition-all duration-200 hover:scale-105 hover:shadow-md group"
                                                    data-open-issue-recorder-button
                                                >
                                                    Record an issue
                                                </Button>
                                                <Button
                                                    size="sm"
                                                    intent="gray-outline"
                                                    onClick={() => {
                                                        checkForUpdates(undefined, {
                                                            onSuccess: (data) => {
                                                                if (data?.release) {
                                                                    queryClient.setQueryData([API_ENDPOINTS.RELEASES.GetLatestUpdate.key], data)

                                                                    if (__isElectronDesktop__) {
                                                                        // Also trigger Electron update
                                                                        if (window.electron) {
                                                                            window.electron.checkForUpdates().catch(() => { })
                                                                        }
                                                                        setElectronUpdateModalOpen(true)
                                                                    } else {
                                                                        setWebUpdateModalOpen(true)
                                                                    }
                                                                } else {
                                                                    toast.success("You are running the latest version")
                                                                }

                                                            },
                                                        })
                                                    }}
                                                    loading={isCheckingForUpdates}
                                                    leftIcon={<LuRefreshCw className="transition-transform duration-200 group-hover:rotate-180" />}
                                                    className="transition-all duration-200 hover:scale-105 hover:shadow-md group"
                                                    data-check-for-updates-button
                                                >
                                                    Check for updates
                                                </Button>
                                            </div>
                                        </div>

                                        <ServerSettings isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="library" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Local Anime Library"
                                            description="Manage your local anime library"
                                            icon={LuLibrary}
                                        />

                                        <AnimeLibrarySettings isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="local" className={tabContentClass}>

                                        <LocalSettings isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="manga" className={tabContentClass}>

                                        <MangaSettings isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="onlinestream" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Online Streaming"
                                            description="Configure online streaming settings"
                                            icon={CgMediaPodcast}
                                        />

                                        <SettingsCard>
                                            <div data-settings-enable-onlinestream>
                                                <Field.Switch
                                                    side="right"
                                                    name="enableOnlinestream"
                                                    label="Enable"
                                                    help="Watch anime episodes from online sources."
                                                />
                                            </div>
                                        </SettingsCard>

                                        <SettingsCard title="Home Screen">
                                            <Field.Switch
                                                side="right"
                                                name="includeOnlineStreamingInLibrary"
                                                label="Include streaming in anime lists"
                                                help="Show currently watching streaming titles in your anime lists."
                                            />
                                        </SettingsCard>

                                        <SettingsSubmitButton isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="discord" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Discord"
                                            description="Configure Discord rich presence settings"
                                            icon={FaDiscord}
                                        />

                                        <DiscordRichPresenceSettings />

                                        <SettingsSubmitButton isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="torrent" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Torrent Provider"
                                            description="Configure the torrent provider"
                                            icon={LuFileSearch}
                                        />

                                        <SettingsCard>
                                            <Field.Select
                                                name="torrentProvider"
                                                label="Default Provider"
                                                help="Used by the search engine. Select 'None' if you don't need torrent support."
                                                leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                                                options={[
                                                    ...(torrentProviderExtensions?.filter(ext => ext?.settings?.type === "main")?.map(ext => ({
                                                        label: ext.name,
                                                        value: ext.id,
                                                    })) ?? []).sort((a, b) => a?.label?.localeCompare(b?.label) ?? 0),
                                                    { label: "None", value: TORRENT_PROVIDER.NONE },
                                                ]}
                                            />
                                            <Field.Switch
                                                side="right"
                                                name="showTorrentAvailability"
                                                label="Show torrent availability on recent episodes"
                                                help="Adds a badge to recent episodes missing from your library, and to Continue Watching when using torrent or Debrid streaming."
                                                disabled={torrentProviderMissing}
                                            />
                                            {torrentProviderMissing && <Alert
                                                intent="warning"
                                                description="Choose a torrent provider to check episode availability."
                                            />}
                                        </SettingsCard>


                                        {/*<Separator />*/}

                                        {/*<h3>DNS over HTTPS</h3>*/}

                                        {/*<Field.Select*/}
                                        {/*    name="dohProvider"*/}
                                        {/*    // label="Torrent Provider"*/}
                                        {/*    help="Choose a DNS over HTTPS provider to resolve domain names for torrent search."*/}
                                        {/*    leftIcon={<FcFilingCabinet className="-500" />}*/}
                                        {/*    options={[*/}
                                        {/*        { label: "None", value: "-" },*/}
                                        {/*        { label: "Cloudflare", value: "cloudflare" },*/}
                                        {/*        { label: "Quad9", value: "quad9" },*/}
                                        {/*    ]}*/}
                                        {/*/>*/}

                                        <SettingsSubmitButton isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="media-player" className={tabContentClass}>
                                        <MediaplayerSettings isPending={isPending} />
                                    </TabsContent>


                                    <TabsContent value="external-player-link" className={tabContentClass}>
                                        <ExternalPlayerLinkSettings />
                                    </TabsContent>

                                    <TabsContent value="playback" className={tabContentClass}>
                                        <PlaybackSettings />
                                    </TabsContent>

                                    <TabsContent value="torrent-client" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Torrent Client"
                                            description="Configure the torrent client"
                                            icon={MdOutlineDownloading}
                                        />

                                        <SettingsCard>
                                            <Field.Select
                                                name="defaultTorrentClient"
                                                label="Default Torrent Client"
                                                options={[
                                                    { label: "qBittorrent", value: "qbittorrent" },
                                                    { label: "Transmission", value: "transmission" },
                                                    ...(status?.featureFlags?.builtinTorrentClient ? [{ label: "Built-in", value: "seanime" }] : []),
                                                    { label: "None", value: "none" },
                                                ]}
                                            />
                                        </SettingsCard>

                                        {/*<SettingsCard>*/}
                                        <Accordion
                                            type="single"
                                            className="group/settings-card relative bg-[--paper] rounded-xl border overflow-hidden"
                                            triggerClass="px-4 py-3 text-[--muted] dark:data-[state=open]:text-white dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white !font-medium transition-all duration-200 hover:translate-x-1"
                                            itemClass="border-b border-[--border] rounded-none transition-all duration-200 hover:border-[--brand]/30 hover:bg-gray-900"
                                            contentClass="!p-4 animate-in duration-300"
                                            collapsible
                                            defaultValue={status?.settings?.torrent?.defaultTorrentClient}
                                        >
                                            <AccordionItem value="qbittorrent">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center">
                                                        <SiQbittorrent className="text-blue-400" /> qBittorrent
                                                    </h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="p-0 py-4 space-y-4">
                                                    <Field.Text
                                                        name="qbittorrentHost"
                                                        label="Host"
                                                    />
                                                    <div className="flex flex-col md:flex-row gap-4">
                                                        <Field.Text
                                                            name="qbittorrentUsername"
                                                            label="Username"
                                                        />
                                                        <Field.Text
                                                            name="qbittorrentPassword"
                                                            label="Password"
                                                            type="password"
                                                        />
                                                        <Field.Number
                                                            name="qbittorrentPort"
                                                            label="Port"
                                                            formatOptions={{
                                                                useGrouping: false,
                                                            }}
                                                        />
                                                    </div>
                                                    <Field.Text
                                                        name="qbittorrentPath"
                                                        label="Executable"
                                                    />
                                                    <Field.Text
                                                        name="qbittorrentTags"
                                                        label="Tags"
                                                        help="Comma separated tags to apply to downloaded torrents. e.g. seanime,anime"
                                                    />
                                                    <Field.Text
                                                        name="qbittorrentCategory"
                                                        label="Category"
                                                        help="Category to apply to downloaded torrents."
                                                    />
                                                </AccordionContent>
                                            </AccordionItem>
                                            <AccordionItem value="transmission">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center">
                                                        <SiTransmission className="text-orange-200" /> Transmission</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="p-0 py-4 space-y-4 !border-b-0">
                                                    <Field.Text
                                                        name="transmissionHost"
                                                        label="Host"
                                                    />
                                                    <div className="flex flex-col md:flex-row gap-4">
                                                        <Field.Text
                                                            name="transmissionUsername"
                                                            label="Username"
                                                        />
                                                        <Field.Text
                                                            name="transmissionPassword"
                                                            label="Password"
                                                            type="password"
                                                        />
                                                        <Field.Number
                                                            name="transmissionPort"
                                                            label="Port"
                                                            formatOptions={{
                                                                useGrouping: false,
                                                            }}
                                                        />
                                                    </div>
                                                    <Field.Text
                                                        name="transmissionPath"
                                                        label="Executable"
                                                    />
                                                </AccordionContent>
                                            </AccordionItem>
                                            {status?.featureFlags?.builtinTorrentClient && (
                                                <AccordionItem value="seanime">
                                                    <AccordionTrigger>
                                                        <h4 className="flex gap-2 items-center">
                                                            <SiBittorrent className="text-[--brand]" /> Built-in
                                                        </h4>
                                                    </AccordionTrigger>
                                                    <AccordionContent className="p-0 py-4 space-y-4">
                                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                            <Field.Number
                                                                name="seanimePort"
                                                                label="Listening port"
                                                                formatOptions={{ useGrouping: false }}
                                                            />
                                                            <Field.Number name="seanimeMaxConnections" label="Connections per torrent" />
                                                            <Field.Number name="seanimeMaxActiveDownloads" label="Active downloads" />
                                                        </div>
                                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                                            <Field.Number
                                                                name="seanimeDownloadLimit"
                                                                label="Download limit (KB/s)"
                                                                help="Set to 0 for no limit."
                                                            />
                                                            <Field.Number
                                                                name="seanimeUploadLimit"
                                                                label="Upload limit (KB/s)"
                                                                help="Set to 0 for no limit."
                                                            />
                                                        </div>
                                                    </AccordionContent>
                                                </AccordionItem>
                                            )}
                                        </Accordion>
                                        {/*</SettingsCard>*/}

                                        <SettingsCard title="Integration">
                                            {/*<Field.Switch*/}
                                            {/*    side="right"*/}
                                            {/*    name="hideTorrentList"*/}
                                            {/*    label="Hide torrent list navigation icon"*/}
                                            {/*/>*/}
                                            <Field.Switch
                                                side="right"
                                                name="showActiveTorrentCount"
                                                label="Show active torrent count"
                                                help="Show the number of active torrents in the sidebar. (Memory intensive)"
                                            />
                                            <Field.Switch
                                                side="right"
                                                name="openTorrentClientOnStart"
                                                label="Open torrent client on startup"
                                            />
                                        </SettingsCard>

                                        <SettingsSubmitButton isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="nakama" className={tabContentClass}>

                                        <NakamaSettings isPending={isPending} />

                                    </TabsContent>
                                </>
                            }}
                        </Form>

                        {/* <TabsContent value="cache" className={tabContentClass}>

                         <SettingsPageHeader
                         title="Cache"
                         description="Manage the cache"
                         icon={TbDatabaseExclamation}
                         />

                         <FilecacheSettings />

                         </TabsContent> */}

                        <TabsContent value="mediastream" className={tabContentClass}>

                            <MediastreamSettings />

                        </TabsContent>

                        <TabsContent value="ui" className={tabContentClass}>

                            <SettingsPageHeader
                                title="User Interface"
                                description="Customize the user interface"
                                icon={LuPalette}
                            />

                            <UISettings />

                        </TabsContent>

                        <TabsContent value="torrentstream" className={tabContentClass}>

                            <SettingsPageHeader
                                title="Torrent Streaming"
                                description="Configure torrent streaming settings"
                                icon={SiBittorrent}
                            />

                            <TorrentstreamSettings settings={torrentstreamSettings} />

                        </TabsContent>

                        <TabsContent value="logs" className={tabContentClass}>

                            <SettingsPageHeader
                                title="Logs"
                                description="View the logs"
                                icon={LuBookKey}
                            />


                            <LogsSettings />

                            <Separator />

                            <SettingsPageHeader
                                title="Cache"
                                description="Manage the cache"
                                icon={TbDatabaseExclamation}
                            />

                            <FilecacheSettings />

                        </TabsContent>

                        {__isElectronDesktop__ && (
                            <TabsContent value="denshi" className={tabContentClass}>

                                <SettingsPageHeader
                                    title="Denshi"
                                    description="Desktop client settings"
                                    icon={LuMonitor}
                                />

                                <DenshiSettings />

                            </TabsContent>
                        )}


                        {/*<TabsContent value="data" className="space-y-4">*/}

                        {/*    <DataSettings />*/}

                        {/*</TabsContent>*/}

                        <TabsContent value="debrid" className={tabContentClass}>

                            <DebridSettings />

                        </TabsContent>
                    </div>
                </Tabs>
                {/*</Card>*/}

            </PageWrapper>
        </>
    )

}
