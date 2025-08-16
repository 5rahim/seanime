"use client"
import { useOpenInExplorer } from "@/api/hooks/explorer.hooks"
import { useAnimeListTorrentProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useSaveSettings } from "@/api/hooks/settings.hooks"
import { useGetTorrentstreamSettings } from "@/api/hooks/torrentstream.hooks"
import { CustomLibraryBanner } from "@/app/(main)/(library)/_containers/custom-library-banner"
import { __issueReport_overlayOpenAtom } from "@/app/(main)/_features/issue-report/issue-report"
import { useServerStatus, useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ExternalPlayerLinkSettings, MediaplayerSettings } from "@/app/(main)/settings/_components/mediaplayer-settings"
import { PlaybackSettings } from "@/app/(main)/settings/_components/playback-settings"
import { __settings_tabAtom } from "@/app/(main)/settings/_components/settings-page.atoms"
import { SettingsIsDirty, SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { DebridSettings } from "@/app/(main)/settings/_containers/debrid-settings"
import { FilecacheSettings } from "@/app/(main)/settings/_containers/filecache-settings"
import { LibrarySettings } from "@/app/(main)/settings/_containers/library-settings"
import { LogsSettings } from "@/app/(main)/settings/_containers/logs-settings"
import { MangaSettings } from "@/app/(main)/settings/_containers/manga-settings"
import { MediastreamSettings } from "@/app/(main)/settings/_containers/mediastream-settings"
import { ServerSettings } from "@/app/(main)/settings/_containers/server-settings"
import { TorrentstreamSettings } from "@/app/(main)/settings/_containers/torrentstream-settings"
import { UISettings } from "@/app/(main)/settings/_containers/ui-settings"
import { PageWrapper } from "@/components/shared/page-wrapper"
import { SeaLink } from "@/components/shared/sea-link"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { DEFAULT_TORRENT_CLIENT, DEFAULT_TORRENT_PROVIDER, settingsSchema, TORRENT_PROVIDER } from "@/lib/server/settings"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { useSetAtom } from "jotai"
import { useAtom } from "jotai/react"
import capitalize from "lodash/capitalize"
import { useRouter, useSearchParams } from "next/navigation"
import React from "react"
import { UseFormReturn } from "react-hook-form"
import { BiDonateHeart } from "react-icons/bi"
import { CgMediaPodcast, CgPlayListSearch } from "react-icons/cg"
import { FaBookReader, FaDiscord } from "react-icons/fa"
import { GrTest } from "react-icons/gr"
import { HiOutlineServerStack } from "react-icons/hi2"
import { ImDownload } from "react-icons/im"
import { IoLibrary, IoPlayBackCircleSharp } from "react-icons/io5"
import { LuBookKey, LuExternalLink, LuLaptop, LuLibrary, LuPalette, LuWandSparkles } from "react-icons/lu"
import { MdOutlineBroadcastOnHome, MdOutlineConnectWithoutContact, MdOutlineDownloading, MdOutlinePalette } from "react-icons/md"
import { RiFolderDownloadFill } from "react-icons/ri"
import { SiBittorrent } from "react-icons/si"
import { TbDatabaseExclamation } from "react-icons/tb"
import { VscDebugAlt } from "react-icons/vsc"
import { SettingsCard, SettingsNavCard, SettingsPageHeader } from "./_components/settings-card"
import { DiscordRichPresenceSettings } from "./_containers/discord-rich-presence-settings"
import { LocalSettings } from "./_containers/local-settings"
import { NakamaSettings } from "./_containers/nakama-settings"

const tabsRootClass = cn("w-full grid grid-cols-1 lg:grid lg:grid-cols-[300px,1fr] gap-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit lg:w-full border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-9 lg:justify-start px-3 transition-all duration-200 hover:bg-[--subtle]/50 hover:transform",
)

const tabsListClass = cn(
    "w-full flex flex-wrap lg:flex-nowrap h-fit xl:h-10",
    "lg:block p-2 lg:p-0",
)

const tabContentClass = cn(
    "space-y-4 animate-in fade-in-0 slide-in-from-right-2 duration-300",
)

export const dynamic = "force-static"

export default function Page() {
    const status = useServerStatus()
    const setServerStatus = useSetServerStatus()
    const router = useRouter()

    const searchParams = useSearchParams()

    const { mutate, data, isPending } = useSaveSettings()

    const [tab, setTab] = useAtom(__settings_tabAtom)
    const formRef = React.useRef<UseFormReturn<any>>(null)

    const { data: torrentProviderExtensions } = useAnimeListTorrentProviderExtensions()

    const { data: torrentstreamSettings } = useGetTorrentstreamSettings()

    const { mutate: openInExplorer, isPending: isOpening } = useOpenInExplorer()

    React.useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
        }
    }, [data, isPending])

    const setIssueRecorderOpen = useSetAtom(__issueReport_overlayOpenAtom)

    function handleOpenIssueRecorder() {
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

    return (
        <>
            <CustomLibraryBanner discrete />
            <PageWrapper data-settings-page-container className="p-4 sm:p-8 space-y-4">
                {/*<Separator/>*/}


                {/*<Card className="p-0 overflow-hidden">*/}
                <Tabs
                    value={tab}
                    onValueChange={setTab}
                    className={tabsRootClass}
                    triggerClass={tabsTriggerClass}
                    listClass={tabsListClass}
                    data-settings-page-tabs
                >
                    <TabsList className="flex-wrap max-w-full lg:space-y-2">
                        <SettingsNavCard>
                            <div className="flex flex-col gap-4 md:flex-row justify-between items-center">
                                <div className="space-y-2 my-3 px-2">
                                    <h4 className="text-center md:text-left text-xl font-bold">Settings</h4>
                                    <div className="space-y-1">
                                        <p className="text-[--muted] text-sm text-center md:text-left flex items-center gap-2">
                                            {status?.version} {status?.versionName} - {capitalize(status?.os)}{__isTauriDesktop__ &&
                                            <span className="font-medium"> - Tauri</span>}{__isElectronDesktop__ &&
                                            <span className="font-medium"> - Denshi</span>}
                                        </p>
                                        {/* <p className="text-[--muted] text-sm text-center md:text-left">OS: {capitalize(status?.os)} {__isTauriDesktop__ &&
                                            <span className="font-medium">- Tauri</span>}{__isElectronDesktop__ &&
                                         <span className="font-medium">- Denshi</span>}</p> */}
                                    </div>
                                </div>
                                <div>

                                </div>
                            </div>
                            <div className="overflow-x-none lg:overflow-y-hidden overflow-y-scroll h-40 lg:h-auto rounded-[--radius-md] border lg:border-none space-y-1 lg:space-y-0">
                                <TabsTrigger
                                    value="seanime"
                                    className="group"
                                ><LuWandSparkles className="text-lg mr-3 transition-transform duration-200" /> App</TabsTrigger>
                                {/* <TabsTrigger
                                    value="local"
                                    className="group"
                                 ><LuUserCog className="text-lg mr-3 transition-transform duration-200" /> Local Account</TabsTrigger> */}
                                <TabsTrigger
                                    value="library"
                                    className="group"
                                ><IoLibrary className="text-lg mr-3 transition-transform duration-200" /> Anime Library</TabsTrigger>

                                <div className="text-xs lg:text-[--muted] text-center py-1.5 uppercase px-3 border-gray-800 tracking-wide font-medium">
                                    Anime playback
                                </div>

                                <TabsTrigger
                                    value="playback"
                                    className="group"
                                ><IoPlayBackCircleSharp className="text-lg mr-3 transition-transform duration-200" /> Video Playback</TabsTrigger>
                                <TabsTrigger
                                    value="media-player"
                                    className="group"
                                ><LuLaptop className="text-lg mr-3 transition-transform duration-200" /> Desktop Media Player</TabsTrigger>
                                <TabsTrigger
                                    value="external-player-link"
                                    className="group"
                                ><LuExternalLink className="text-lg mr-3 transition-transform duration-200" /> External Player Link</TabsTrigger>
                                <TabsTrigger
                                    value="mediastream"
                                    className="relative group"
                                ><MdOutlineBroadcastOnHome className="text-lg mr-3 transition-transform duration-200" /> Transcoding / Direct
                                                                                                                         play</TabsTrigger>

                                <div className="text-xs lg:text-[--muted] text-center py-1.5 uppercase px-3 border-gray-800 tracking-wide font-medium">
                                    Torrenting
                                </div>

                                <TabsTrigger
                                    value="torrent"
                                    className="group"
                                ><CgPlayListSearch className="text-lg mr-3 transition-transform duration-200" /> Torrent Provider</TabsTrigger>
                                <TabsTrigger
                                    value="torrent-client"
                                    className="group"
                                ><MdOutlineDownloading className="text-lg mr-3 transition-transform duration-200" /> Torrent Client</TabsTrigger>
                                <TabsTrigger
                                    value="torrentstream"
                                    className="relative group"
                                ><SiBittorrent className="text-lg mr-3 transition-transform duration-200" /> Torrent Streaming</TabsTrigger>
                                <TabsTrigger
                                    value="debrid"
                                    className="group"
                                ><HiOutlineServerStack className="text-lg mr-3 transition-transform duration-200" /> Debrid Service</TabsTrigger>

                                <div className="text-xs lg:text-[--muted] text-center py-1.5 uppercase px-3 border-gray-800 tracking-wide font-medium">
                                    Other features
                                </div>

                                <TabsTrigger
                                    value="onlinestream"
                                    className="group"
                                ><CgMediaPodcast className="text-lg mr-3 transition-transform duration-200" /> Online Streaming</TabsTrigger>

                                <TabsTrigger
                                    value="manga"
                                    className="group"
                                ><FaBookReader className="text-lg mr-3 transition-transform duration-200" /> Manga</TabsTrigger>
                                <TabsTrigger
                                    value="nakama"
                                    className="group relative"
                                ><MdOutlineConnectWithoutContact className="text-lg mr-3 transition-transform duration-200" /> Nakama <GrTest
                                    className="text-md text-orange-300/40 absolute right-2 lg:block hidden"
                                /></TabsTrigger>
                                <TabsTrigger
                                    value="discord"
                                    className="group"
                                ><FaDiscord className="text-lg mr-3 transition-transform duration-200" /> Discord</TabsTrigger>

                                <div className="text-xs lg:text-[--muted] text-center py-1.5 uppercase px-3 border-gray-800 tracking-wide font-medium">
                                    Server & Interface
                                </div>

                                <TabsTrigger
                                    value="ui"
                                    className="group"
                                ><MdOutlinePalette className="text-lg mr-3 transition-transform duration-200" /> User Interface</TabsTrigger>
                                {/* <TabsTrigger
                                    value="cache"
                                    className="group"
                                 ><TbDatabaseExclamation className="text-lg mr-3 transition-transform duration-200" /> Cache</TabsTrigger> */}
                                <TabsTrigger
                                    value="logs"
                                    className="group"
                                ><LuBookKey className="text-lg mr-3 transition-transform duration-200" /> Logs & Cache</TabsTrigger>
                            </div>
                        </SettingsNavCard>

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
                    </TabsList>

                    <div className="">
                        <Form
                            schema={settingsSchema}
                            mRef={formRef}
                            onSubmit={data => {
                                mutate({
                                    library: {
                                        libraryPath: data.libraryPath,
                                        autoUpdateProgress: data.autoUpdateProgress,
                                        disableUpdateCheck: data.disableUpdateCheck,
                                        torrentProvider: data.torrentProvider,
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
                                    },
                                    torrent: {
                                        defaultTorrentClient: data.defaultTorrentClient,
                                        qbittorrentPath: data.qbittorrentPath,
                                        qbittorrentHost: data.qbittorrentHost,
                                        qbittorrentPort: data.qbittorrentPort,
                                        qbittorrentPassword: data.qbittorrentPassword,
                                        qbittorrentUsername: data.qbittorrentUsername,
                                        qbittorrentTags: data.qbittorrentTags,
                                        transmissionPath: data.transmissionPath,
                                        transmissionHost: data.transmissionHost,
                                        transmissionPort: data.transmissionPort,
                                        transmissionUsername: data.transmissionUsername,
                                        transmissionPassword: data.transmissionPassword,
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
                                    },
                                    notifications: {
                                        disableNotifications: data?.disableNotifications ?? false,
                                        disableAutoDownloaderNotifications: data?.disableAutoDownloaderNotifications ?? false,
                                        disableAutoScannerNotifications: data?.disableAutoScannerNotifications ?? false,
                                    },
                                }, {
                                    onSuccess: () => {
                                        formRef.current?.reset(formRef.current.getValues())
                                    },
                                })
                            }}
                            defaultValues={{
                                libraryPath: status?.settings?.library?.libraryPath,
                                mediaPlayerHost: status?.settings?.mediaPlayer?.host,
                                torrentProvider: status?.settings?.library?.torrentProvider || DEFAULT_TORRENT_PROVIDER, // (Backwards compatibility)
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
                                transmissionPath: status?.settings?.torrent?.transmissionPath,
                                transmissionHost: status?.settings?.torrent?.transmissionHost,
                                transmissionPort: status?.settings?.torrent?.transmissionPort,
                                transmissionUsername: status?.settings?.torrent?.transmissionUsername,
                                transmissionPassword: status?.settings?.torrent?.transmissionPassword,
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
                            }}
                            stackClass="space-y-0 relative"
                        >
                            {(f) => {
                                return <>
                                    <SettingsIsDirty />
                                    <TabsContent value="seanime" className={tabContentClass}>

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
                                            >
                                                Record an issue
                                            </Button>
                                        </div>

                                        <ServerSettings isPending={isPending} />

                                    </TabsContent>

                                    <TabsContent value="library" className={tabContentClass}>

                                        <SettingsPageHeader
                                            title="Anime Library"
                                            description="Manage your local anime library"
                                            icon={LuLibrary}
                                        />

                                        <LibrarySettings isPending={isPending} />

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
                                            <Field.Switch
                                                side="right"
                                                name="enableOnlinestream"
                                                label="Enable"
                                                help="Watch anime episodes from online sources."
                                            />
                                        </SettingsCard>

                                        <SettingsCard title="My library">
                                            <Field.Switch
                                                side="right"
                                                name="includeOnlineStreamingInLibrary"
                                                label="Include in library"
                                                help="Add non-downloaded shows that are in your currently watching list to 'My library' for streaming"
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
                                            icon={CgPlayListSearch}
                                        />

                                        <SettingsCard>
                                            <Field.Select
                                                name="torrentProvider"
                                                // label="Torrent Provider"
                                                help="Used by the search engine and auto downloader. AnimeTosho is recommended for better results. Select 'None' if you don't need torrent support."
                                                leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                                                options={[
                                                    ...(torrentProviderExtensions?.filter(ext => ext?.settings?.type === "main")?.map(ext => ({
                                                        label: ext.name,
                                                        value: ext.id,
                                                    })) ?? []).sort((a, b) => a?.label?.localeCompare(b?.label) ?? 0),
                                                    { label: "None", value: TORRENT_PROVIDER.NONE },
                                                ]}
                                            />
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
                                                label="Default torrent client"
                                                options={[
                                                    { label: "qBittorrent", value: "qbittorrent" },
                                                    { label: "Transmission", value: "transmission" },
                                                    { label: "None", value: "none" },
                                                ]}
                                            />
                                        </SettingsCard>

                                        <SettingsCard>
                                            <Accordion
                                                type="single"
                                                className=""
                                                triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black transition-all duration-200 hover:translate-x-1"
                                                itemClass="border-b border-[--border] rounded-[--radius] transition-all duration-200 hover:border-[--brand]/30"
                                                contentClass="pb-8 animate-in slide-in-from-top-2 duration-300"
                                                collapsible
                                                defaultValue={status?.settings?.torrent?.defaultTorrentClient}
                                            >
                                                <AccordionItem value="qbittorrent">
                                                    <AccordionTrigger>
                                                        <h4 className="flex gap-2 items-center"><ImDownload className="text-blue-400" /> qBittorrent
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
                                                    </AccordionContent>
                                                </AccordionItem>
                                                <AccordionItem value="transmission">
                                                    <AccordionTrigger>
                                                        <h4 className="flex gap-2 items-center">
                                                            <ImDownload className="text-orange-200" /> Transmission</h4>
                                                    </AccordionTrigger>
                                                    <AccordionContent className="p-0 py-4 space-y-4">
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
                                            </Accordion>
                                        </SettingsCard>

                                        <SettingsCard title="User Interface">
                                            <Field.Switch
                                                side="right"
                                                name="hideTorrentList"
                                                label="Hide torrent list navigation icon"
                                            />
                                            <Field.Switch
                                                side="right"
                                                name="showActiveTorrentCount"
                                                label="Show active torrent count"
                                                help="Show the number of active torrents in the sidebar. (Memory intensive)"
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

                            <SettingsPageHeader
                                title="Transcoding / Direct play"
                                description="Manage transcoding and direct play settings"
                                icon={MdOutlineBroadcastOnHome}
                            />

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


                        {/*<TabsContent value="data" className="space-y-4">*/}

                        {/*    <DataSettings />*/}

                        {/*</TabsContent>*/}

                        <TabsContent value="debrid" className={tabContentClass}>

                            <SettingsPageHeader
                                title="Debrid Service"
                                description="Configure your debrid service integration"
                                icon={HiOutlineServerStack}
                            />

                            <DebridSettings />

                        </TabsContent>
                    </div>
                </Tabs>
                {/*</Card>*/}

            </PageWrapper>
        </>
    )

}
