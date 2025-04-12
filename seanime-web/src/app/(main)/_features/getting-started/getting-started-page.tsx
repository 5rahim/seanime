import { Status } from "@/api/generated/types"
import { useGettingStarted } from "@/api/hooks/settings.hooks"
import { useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Card } from "@/components/ui/card"
import { CheckboxProps } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Field, Form } from "@/components/ui/form"
import {
    DEFAULT_DOH_PROVIDER,
    DEFAULT_TORRENT_PROVIDER,
    getDefaultMpcSocket,
    gettingStartedSchema,
    TORRENT_PROVIDER,
    useDefaultSettingsPaths,
} from "@/lib/server/settings"
import { useRouter } from "next/navigation"
import React from "react"
import { FcClapperboard, FcFolder, FcVideoCall, FcVlc } from "react-icons/fc"
import { HiPlay } from "react-icons/hi"
import { HiOutlineServerStack } from "react-icons/hi2"
import { ImDownload } from "react-icons/im"
import { RiFolderDownloadFill } from "react-icons/ri"

/**
 * @description
 * - Page to set up the initial settings for the application
 */
export function GettingStartedPage({ status }: { status: Status }) {
    const router = useRouter()
    const { getDefaultVlcPath, getDefaultQBittorrentPath, getDefaultTransmissionPath } = useDefaultSettingsPaths()
    const setServerStatus = useSetServerStatus()

    const { mutate, data, isPending, isSuccess } = useGettingStarted()

    /**
     * If the settings are returned, redirect to the home page
     */
    React.useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
            router.push("/")
        }
    }, [data, isPending])

    const vlcDefaultPath = React.useMemo(() => getDefaultVlcPath(status.os), [status.os])
    const qbittorrentDefaultPath = React.useMemo(() => getDefaultQBittorrentPath(status.os), [status.os])
    const transmissionDefaultPath = React.useMemo(() => getDefaultTransmissionPath(status.os), [status.os])
    const mpvSocketPath = React.useMemo(() => getDefaultMpcSocket(status.os), [status.os])

    if (isPending) return <LoadingOverlayWithLogo />

    if (!data) return (
        <div className="container max-w-5xl py-10">
            <div className="mb-4 flex justify-center w-full">
                <img src="/logo_2.png" alt="logo" className="w-36 h-auto" />
            </div>
            <Card className="relative p-4 bg-transparent border-none">
                <AppLayoutStack>
                    <div className="space-y-4 p-1">
                        <div>
                            <h3 className="text-center">Getting started</h3>
                            <p className="italic text-[--muted] text-center">These settings can be modified later.</p>
                        </div>
                        <Form
                            schema={gettingStartedSchema}
                            onSubmit={data => {
                                mutate({
                                    library: {
                                        libraryPath: data.libraryPath,
                                        autoUpdateProgress: true,
                                        disableUpdateCheck: false,
                                        torrentProvider: data.torrentProvider || DEFAULT_TORRENT_PROVIDER,
                                        autoScan: false,
                                        disableAnimeCardTrailers: false,
                                        enableManga: data.enableManga,
                                        enableOnlinestream: data.enableOnlinestream,
                                        dohProvider: DEFAULT_DOH_PROVIDER,
                                        openTorrentClientOnStart: false,
                                        openWebURLOnStart: false,
                                        refreshLibraryOnStart: false,
                                        autoPlayNextEpisode: false,
                                        enableWatchContinuity: false,
                                        libraryPaths: [],
                                        autoSyncOfflineLocalData: false,
                                        includeOnlineStreamingInLibrary: false,
                                        scannerMatchingThreshold: 0,
                                        scannerMatchingAlgorithm: "",
                                    },
                                    manga: {
                                        defaultMangaProvider: "",
                                        mangaAutoUpdateProgress: false,
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
                                    },
                                    discord: {
                                        enableRichPresence: data.enableRichPresence,
                                        enableAnimeRichPresence: true,
                                        enableMangaRichPresence: true,
                                        richPresenceHideSeanimeRepositoryButton: false,
                                        richPresenceShowAniListMediaButton: false,
                                        richPresenceShowAniListProfileButton: false,
                                    },
                                    torrent: {
                                        defaultTorrentClient: data.defaultTorrentClient,
                                        qbittorrentPath: data.qbittorrentPath,
                                        qbittorrentHost: data.qbittorrentHost,
                                        qbittorrentPort: data.qbittorrentPort,
                                        qbittorrentPassword: data.qbittorrentPassword,
                                        qbittorrentUsername: data.qbittorrentUsername,
                                        qbittorrentTags: "",
                                        transmissionPath: data.transmissionPath,
                                        transmissionHost: data.transmissionHost,
                                        transmissionPort: data.transmissionPort,
                                        transmissionUsername: data.transmissionUsername,
                                        transmissionPassword: data.transmissionPassword,
                                        showActiveTorrentCount: false,
                                        hideTorrentList: false,
                                    },
                                    anilist: {
                                        hideAudienceScore: false,
                                        enableAdultContent: data.enableAdultContent,
                                        blurAdultContent: false,
                                    },
                                    enableTorrentStreaming: data.enableTorrentStreaming,
                                    enableTranscode: data.enableTranscode,
                                    notifications: {
                                        disableNotifications: false,
                                        disableAutoDownloaderNotifications: false,
                                        disableAutoScannerNotifications: false,
                                    },
                                    debridProvider: data.debridProvider,
                                    debridApiKey: data.debridApiKey,
                                })
                            }}
                            defaultValues={{
                                mediaPlayerHost: "127.0.0.1",
                                vlcPort: 8080,
                                mpcPort: 13579,
                                defaultPlayer: "mpv",
                                vlcPath: vlcDefaultPath,
                                qbittorrentPath: qbittorrentDefaultPath,
                                qbittorrentHost: "127.0.0.1",
                                qbittorrentPort: 8081,
                                transmissionPath: transmissionDefaultPath,
                                transmissionHost: "127.0.0.1",
                                transmissionPort: 9091,
                                mpcPath: "C:/Program Files/MPC-HC/mpc-hc64.exe",
                                torrentProvider: DEFAULT_TORRENT_PROVIDER,
                                mpvSocket: mpvSocketPath,
                                enableRichPresence: false,
                                autoScan: false,
                                enableManga: true,
                                enableOnlinestream: false,
                                enableAdultContent: true,
                                enableTorrentStreaming: false,
                                enableTranscode: false,
                                debridProvider: "none",
                                debridApiKey: "",
                            }}
                            stackClass="space-y-8"
                        >
                            {(f) => (<>

                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Local anime library</h4>
                                    </div>

                                    <Field.DirectorySelector
                                        name="libraryPath"
                                        label="Anime library path"
                                        leftIcon={<FcFolder />}
                                        shouldExist
                                        help="Enter the full path. You can add more folders later."
                                    />
                                </Card>


                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Desktop Media Player</h4>

                                        <p className="text-[--muted] text-center text-sm">
                                            Used to play media on your host computer and track your progress automatically.
                                        </p>
                                        <p className="text-[--muted] text-center text-sm">
                                            Used for <span className="font-bold">anime library</span>, <span className="font-bold">torrent streaming &
                                                                                                                                   debrid
                                                                                                                                   streaming</span>.
                                        </p>
                                    </div>

                                    <div className="space-y-4">
                                        <Field.Select
                                            name="defaultPlayer"
                                            label="Default media player"
                                            required
                                            leftIcon={<FcVideoCall />}
                                            options={[
                                                { label: "MPV", value: "mpv" },
                                                { label: "VLC", value: "vlc" },
                                                { label: "MPC-HC", value: "mpc-hc" },
                                            ]}
                                        />

                                        <Accordion
                                            type="single"
                                            className=""
                                            triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                                            itemClass="border-b"
                                            contentClass="pb-8"
                                            collapsible
                                        >
                                            <AccordionItem value="mpv">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center"><HiPlay /> MPV</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="px-1 py-4 space-y-4">
                                                    <div className="flex gap-4 flex-col md:flex-row">
                                                        <Field.Text
                                                            name="mpvSocket"
                                                            label="Socket / Pipe"
                                                        />
                                                    </div>
                                                </AccordionContent>
                                            </AccordionItem>
                                            <AccordionItem value="vlc">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center"><FcVlc /> VLC</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="px-1 py-4 space-y-4">
                                                    <p className="text-[--muted]">
                                                        Leave these fields if you don't want to use VLC.
                                                    </p>
                                                    <div className="flex gap-4 flex-col md:flex-row">
                                                        <Field.Text
                                                            name="mediaPlayerHost"
                                                            label="Host"
                                                        />
                                                        <Field.Number
                                                            name="vlcPort"
                                                            label="Port"
                                                            formatOptions={{
                                                                useGrouping: false,
                                                            }}
                                                        />
                                                    </div>
                                                    <div className="flex gap-4 flex-col md:flex-row">
                                                        <Field.Text
                                                            name="vlcUsername"
                                                            label="Username"
                                                        />
                                                        <Field.Text
                                                            name="vlcPassword"
                                                            label="Password"
                                                        />
                                                    </div>
                                                    <Field.Text
                                                        name="vlcPath"
                                                        label="Executable"
                                                        help="Path to the VLC executable, this is used to launch the application."
                                                    />
                                                </AccordionContent>
                                            </AccordionItem>
                                            <AccordionItem value="mpc-hc">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center"><FcClapperboard /> MPC-HC</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="px-1 py-4 space-y-4">
                                                    <p className="text-[--muted]">
                                                        Leave these fields if you don't want to use MPC-HC.
                                                    </p>
                                                    <div className="flex gap-4 flex-col md:flex-row">
                                                        <Field.Text
                                                            name="mediaPlayerHost"
                                                            label="Host"
                                                        />
                                                        <Field.Number
                                                            name="mpcPort"
                                                            label="Port"
                                                            formatOptions={{
                                                                useGrouping: false,
                                                            }}
                                                        />
                                                    </div>
                                                    <div className="flex gap-4 flex-col md:flex-row">
                                                        <Field.Text
                                                            name="mpcPath"
                                                            label="Executable"
                                                            help="Path to the MPC-HC executable, this is used to launch the application."
                                                        />
                                                    </div>
                                                </AccordionContent>
                                            </AccordionItem>
                                        </Accordion>
                                    </div>
                                </Card>


                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Torrent Provider Extension</h4>

                                        <p className="text-[--muted] text-center text-sm">
                                            Built-in torrent provider extension used by the search engine and Auto Downloader.
                                        </p>
                                        <p className="text-[--muted] text-center text-sm">
                                            Used for <span className="font-bold">anime library</span>, <span className="font-bold">torrent streaming &
                                                                                                                                   debrid
                                                                                                                                   streaming</span>.
                                        </p>
                                    </div>


                                    <Field.Select
                                        name="torrentProvider"
                                        label="Torrent Provider"
                                        required
                                        leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                                        options={[
                                            { label: "AnimeTosho (recommended)", value: TORRENT_PROVIDER.ANIMETOSHO },
                                            { label: "Nyaa", value: TORRENT_PROVIDER.NYAA },
                                        ]}
                                    />
                                </Card>


                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Torrent Client</h4>

                                        <p className="text-[--muted] text-center">
                                            Torrent client used to download anime.
                                        </p>
                                    </div>

                                    <div className="space-y-4">
                                        <Field.Select
                                            name="defaultTorrentClient"
                                            label="Default torrent client"
                                            leftIcon={<ImDownload className="text-blue-400" />}
                                            options={[
                                                { label: "qBittorrent", value: "qbittorrent" },
                                                { label: "Transmission", value: "transmission" },
                                                { label: "None", value: "none" },
                                            ]}
                                        />

                                        <Accordion
                                            type="single"
                                            className=""
                                            triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                                            itemClass="border-b"
                                            contentClass="pb-8"
                                            collapsible
                                        >
                                            <AccordionItem value="qbittorrent">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center">qBittorrent</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="px-1 py-4 space-y-4">
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
                                                        help="Path to the qBittorrent executable, this is used to launch the application."
                                                    />
                                                </AccordionContent>
                                            </AccordionItem>
                                            <AccordionItem value="transmission">
                                                <AccordionTrigger>
                                                    <h4 className="flex gap-2 items-center">Transmission</h4>
                                                </AccordionTrigger>
                                                <AccordionContent className="px-1 py-4 space-y-4">
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
                                                        help="Path to the Transmission executable, this is used to launch the application."
                                                    />
                                                </AccordionContent>
                                            </AccordionItem>
                                        </Accordion>
                                    </div>
                                </Card>

                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Debrid Service</h4>

                                        <p className="text-[--muted] text-center">
                                            Debrid service used to download or stream anime.
                                        </p>
                                    </div>

                                    <div className="space-y-4">
                                        <Field.Select
                                            name="debridProvider"
                                            label="Debrid service"
                                            leftIcon={<HiOutlineServerStack className="text-yellow-500" />}
                                            options={[
                                                { label: "None", value: "none" },
                                                { label: "TorBox", value: "torbox" },
                                                { label: "Real-Debrid", value: "realdebrid" },
                                            ]}
                                        />

                                        {f.watch("debridProvider") !== "none" && f.watch("debridProvider") !== "" && (
                                            <Field.Text
                                                name="debridApiKey"
                                                label="API Key"
                                                help="API key provided by the debrid service."
                                            />
                                        )}
                                    </div>
                                </Card>

                                <Card className="p-4 space-y-4">
                                    <div>
                                        <h4 className="text-center">Features</h4>

                                        <p className="text-[--muted] text-center">
                                            Select additional features you want to use.
                                        </p>
                                    </div>

                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                        <Field.Checkbox
                                            name="enableManga"
                                            label={<span>Manga</span>}
                                            help="Read and download manga chapters."
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />

                                        <Field.Checkbox
                                            name="enableTranscode"
                                            label={<span>Media streaming / Transcoding</span>}
                                            help="Stream downloaded episodes to other devices using transcoding or direct play. FFmpeg is required."
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />

                                        <Field.Checkbox
                                            name="enableTorrentStreaming"
                                            label={<span>Torrent streaming</span>}
                                            help="Stream torrents directly to your media player without having to wait for the download to complete."
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />

                                        <Field.Checkbox
                                            name="enableOnlinestream"
                                            label={<span>Online streaming</span>}
                                            help="Watch anime episodes from online sources."
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />

                                        <Field.Checkbox
                                            name="enableRichPresence"
                                            label={<span>Discord Rich Presence</span>}
                                            help="Show what you're watching/reading on Discord."
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />


                                        <Field.Checkbox
                                            name="enableAdultContent"
                                            label={<span>NSFW</span>}
                                            help={<div>
                                                <p>Show adult content in your library and search results.</p>
                                            </div>}
                                            size="lg"
                                            {...cardCheckboxStyles}
                                        />
                                    </div>
                                </Card>


                                <Field.Submit
                                    className="w-full"
                                    role="submit"
                                    showLoadingOverlayOnSuccess={true}
                                    loading={isPending}
                                >Continue</Field.Submit>
                            </>)}
                        </Form>
                    </div>
                </AppLayoutStack>
            </Card>
            <p className="text-[--muted] mt-5 text-center">Made by 5rahim</p>
        </div>
    )
}

const cardCheckboxStyles: CheckboxProps = {
    labelClass: cn(
        "block cursor-pointer transition overflow-hidden w-full py-1 rounded-[--radius-md]",
        "bg-gray-50 hover:bg-[--subtle] dark:bg-gray-950 border-dashed",
        "data-[checked=false]:text-[--muted] hover:data-[checked=false]:opacity-50",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950",
        "focus:ring-2 ring-brand-100 dark:ring-brand-900 ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "data-[checked=true]:ring-offset-0",
        "data-[checked=true]:text-[--brand]",
    ),
    containerClass: "flex items-center justify-between",
    className: cn(
        "absolute top-0 right-0 z-10 rounded-tl-none rounded-br-none",
    ),
    fieldClass: "border p-4 rounded-[--radius-md] relative",
    fieldHelpTextClass: "text-pretty",
}
