"use client"
import { serverStatusAtom } from "@/atoms/server-status"
import { BetaBadge } from "@/components/application/beta-badge"
import { tabsListClass, tabsTriggerClass } from "@/components/shared/styling/classnames"
import { PageWrapper } from "@/components/shared/styling/page-wrapper"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { Field, Form } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { getDefaultMpcSocket, settingsSchema } from "@/lib/server/settings"
import { DEFAULT_TORRENT_CLIENT, DEFAULT_TORRENT_PROVIDER, ServerStatus, Settings } from "@/lib/server/types"
import { useAtom } from "jotai/react"
import Link from "next/link"
import React, { useEffect } from "react"
import { FcClapperboard, FcFolder, FcVideoCall, FcVlc } from "react-icons/fc"
import { GoArrowRight } from "react-icons/go"
import { HiPlay } from "react-icons/hi"
import { ImDownload } from "react-icons/im"
import { LuLayoutDashboard } from "react-icons/lu"
import { RiFolderDownloadFill } from "react-icons/ri"
import { toast } from "sonner"


export default function Page() {
    const [status, setServerStatus] = useAtom(serverStatusAtom)

    const { mutate, data, isPending } = useSeaMutation<ServerStatus, Settings>({
        endpoint: SeaEndpoints.SETTINGS,
        mutationKey: ["patch-settings"],
        method: "patch",
        onSuccess: () => {
            toast.success("Settings updated")
        },
    })

    useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
        }
    }, [data, isPending])

    return (
        <PageWrapper className="p-4 sm:p-8 space-y-4">
            <div className="flex flex-col gap-4 md:flex-row justify-between items-center">
                <div className="space-y-1">
                    <h2>Settings</h2>
                    <p className="text-[--muted]">App version: {status?.version}-{status?.os}</p>
                </div>
                <div>
                    <Link href="/settings/ui">
                        <Button
                            className="rounded-full"
                            intent="primary-subtle"
                            leftIcon={<LuLayoutDashboard />}
                            rightIcon={<GoArrowRight />}
                        >Customize UI</Button>
                    </Link>
                </div>
            </div>
            {/*<Separator/>*/}
            <Form
                schema={settingsSchema}
                onSubmit={data => {
                    mutate({
                        library: {
                            libraryPath: data.libraryPath,
                            autoUpdateProgress: data.autoUpdateProgress,
                            disableUpdateCheck: data.disableUpdateCheck,
                            torrentProvider: data.torrentProvider,
                            autoScan: data.autoScan,
                            enableOnlinestream: data.enableOnlinestream,
                            disableAnimeCardTrailers: data.disableAnimeCardTrailers,
                            enableManga: data.enableManga,
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
                        torrent: {
                            defaultTorrentClient: data.defaultTorrentClient,
                            qbittorrentPath: data.qbittorrentPath,
                            qbittorrentHost: data.qbittorrentHost,
                            qbittorrentPort: data.qbittorrentPort,
                            qbittorrentPassword: data.qbittorrentPassword,
                            qbittorrentUsername: data.qbittorrentUsername,
                            transmissionPath: data.transmissionPath,
                            transmissionHost: data.transmissionHost,
                            transmissionPort: data.transmissionPort,
                            transmissionUsername: data.transmissionUsername,
                            transmissionPassword: data.transmissionPassword,
                        },
                        anilist: {
                            hideAudienceScore: data.hideAudienceScore,
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
                    defaultTorrentClient: status?.settings?.torrent?.defaultTorrentClient || DEFAULT_TORRENT_CLIENT, // (Backwards compatibility)
                    qbittorrentPath: status?.settings?.torrent?.qbittorrentPath,
                    qbittorrentHost: status?.settings?.torrent?.qbittorrentHost,
                    qbittorrentPort: status?.settings?.torrent?.qbittorrentPort,
                    qbittorrentPassword: status?.settings?.torrent?.qbittorrentPassword,
                    qbittorrentUsername: status?.settings?.torrent?.qbittorrentUsername,
                    transmissionPath: status?.settings?.torrent?.transmissionPath,
                    transmissionHost: status?.settings?.torrent?.transmissionHost,
                    transmissionPort: status?.settings?.torrent?.transmissionPort,
                    transmissionUsername: status?.settings?.torrent?.transmissionUsername,
                    transmissionPassword: status?.settings?.torrent?.transmissionPassword,
                    hideAudienceScore: status?.settings?.anilist?.hideAudienceScore ?? false,
                    autoUpdateProgress: status?.settings?.library?.autoUpdateProgress ?? false,
                    disableUpdateCheck: status?.settings?.library?.disableUpdateCheck ?? false,
                    enableOnlinestream: status?.settings?.library?.enableOnlinestream ?? false,
                    disableAnimeCardTrailers: status?.settings?.library?.disableAnimeCardTrailers ?? false,
                    enableManga: status?.settings?.library?.enableManga ?? false,
                }}
                stackClass="space-y-4"
            >

                {/*<Card className="p-0 overflow-hidden">*/}
                    <Tabs
                        defaultValue="seanime"
                        triggerClass={tabsTriggerClass}
                        listClass={tabsListClass}
                    >
                        <TabsList>
                            <TabsTrigger value="seanime">Seanime</TabsTrigger>
                            <TabsTrigger value="media-player">Media Player</TabsTrigger>
                            <TabsTrigger value="torrent-client">Torrent Client</TabsTrigger>
                        </TabsList>

                        <div className="pt-4">
                            <TabsContent value="seanime" className="space-y-6">

                                <h3>Library</h3>

                                <Field.DirectorySelector
                                    name="libraryPath"
                                    label="Library folder"
                                    leftIcon={<FcFolder />}
                                    help="Folder where your anime library is located. (Keep the casing consistent)"
                                    shouldExist
                                />
                                <Separator />
                                <Field.Switch
                                    name="autoScan"
                                    label="Automatically refresh library"
                                    help={<div>
                                        <p>If enabled, your library will be refreshed in the background when new files are added/deleted. Make sure to
                                           lock your files regularly.</p>
                                        <p>
                                            <em>Note:</em> This works best when single files are added/deleted. If you are adding a batch, not all
                                                           files
                                                           are guaranteed to be picked up.
                                        </p>
                                    </div>}
                                />
                                <Separator />
                                <Field.Switch
                                    name="disableUpdateCheck"
                                    label="Do not check for updates"
                                    help="If enabled, Seanime will not check for new releases."
                                />
                                <Separator />

                                <h3>Watching</h3>

                                <Field.Switch
                                    name="autoUpdateProgress"
                                    label="Automatically update progress"
                                    help="If enabled, your progress will be automatically updated without having to confirm it when you watch 90% of an episode."
                                />
                                <Separator />

                                <h3>AniList</h3>

                                <Field.Switch
                                    name="hideAudienceScore"
                                    label="Hide audience score"
                                    help="If enabled, the audience score will be hidden until you decide to view it."
                                />
                                <Separator />

                                <h3>Features</h3>

                                <Field.Switch
                                    name="enableManga"
                                    label={<span className="flex gap-1 items-center">Manga <BetaBadge /></span>}
                                    help="Read manga chapters and track your progress."
                                />
                                <Field.Switch
                                    name="enableOnlinestream"
                                    label={<span className="flex gap-1 items-center">Online streaming <BetaBadge /></span>}
                                    help="Watch anime episodes from online sources."
                                />
                                <Field.Switch
                                    name="disableAnimeCardTrailers"
                                    label="Disable anime card trailers"
                                    help=""
                                />
                                <Separator />

                                <h3>Torrent Provider</h3>

                                <Field.Select
                                    name="torrentProvider"
                                    // label="Torrent Provider"
                                    help="Used by the search engine and auto downloader. AnimeTosho is recommended for better results."
                                    leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                                    options={[
                                        { label: "AnimeTosho (recommended)", value: "animetosho" },
                                        { label: "Nyaa", value: "nyaa" },
                                    ]}
                                />

                            </TabsContent>

                            <TabsContent value="media-player" className="space-y-4">
                                <Field.Select
                                    name="defaultPlayer"
                                    label="Default player"
                                    leftIcon={<FcVideoCall />}
                                    options={[
                                        { label: "VLC", value: "vlc" },
                                        { label: "MPC-HC (Windows only)", value: "mpc-hc" },
                                        { label: "MPV", value: "mpv" },
                                    ]}
                                    help="Player that will be used to open files and track your progress automatically."
                                />

                                <Field.Text
                                    name="mediaPlayerHost"
                                    label="Host"
                                    help="VLC/MPC-HC"
                                />

                                <Accordion
                                    type="single"
                                    className=""
                                    triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                                    itemClass="border-b"
                                    contentClass="pb-8"
                                    collapsible
                                >
                                    <AccordionItem value="vlc">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><FcVlc /> VLC</h4>
                                        </AccordionTrigger>
                                        <AccordionContent className="space-y-4">
                                            <div className="flex flex-col md:flex-row gap-4">
                                                <Field.Text
                                                    name="vlcUsername"
                                                    label="Username"
                                                />
                                                <Field.Text
                                                    name="vlcPassword"
                                                    label="Password"
                                                />
                                                <Field.Number
                                                    name="vlcPort"
                                                    label="Port"
                                                    formatOptions={{
                                                        useGrouping: false,
                                                    }}
                                                    hideControls
                                                />
                                            </div>
                                            <Field.Text
                                                name="vlcPath"
                                                label="Application path"
                                            />
                                        </AccordionContent>
                                    </AccordionItem>

                                    <AccordionItem value="mpc-hc">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><FcClapperboard /> MPC-HC</h4>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="flex flex-col md:flex-row gap-4">
                                                <Field.Number
                                                    name="mpcPort"
                                                    label="Port"
                                                    formatOptions={{
                                                        useGrouping: false,
                                                    }}
                                                    hideControls
                                                />
                                                <Field.Text
                                                    name="mpcPath"
                                                    label="Application path"
                                                />
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>

                                    <AccordionItem value="mpv">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><HiPlay className="mr-1 text-purple-100" /> MPV</h4>
                                        </AccordionTrigger>
                                        <AccordionContent>
                                            <div className="flex gap-4">
                                                <Field.Text
                                                    name="mpvSocket"
                                                    label="Socket"
                                                    placeholder={`Default: '${getDefaultMpcSocket(status?.os ?? "")}'`}
                                                />
                                                <Field.Text
                                                    name="mpvPath"
                                                    label="Application path"
                                                    placeholder="Defaults to 'mpv' command"
                                                    help="Leave empty to automatically use the 'mpv' command"
                                                />
                                            </div>
                                        </AccordionContent>
                                    </AccordionItem>
                                </Accordion>

                            </TabsContent>

                            <TabsContent value="torrent-client" className="space-y-4">

                                <Field.Select
                                    name="defaultTorrentClient"
                                    label="Default torrent client"
                                    options={[
                                        { label: "qBittorrent", value: "qbittorrent" },
                                        { label: "Transmission", value: "transmission" },
                                    ]}
                                />

                                <Accordion
                                    type="single"
                                    className=""
                                    triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                                    itemClass="border-b"
                                    contentClass="pb-8"
                                    collapsible
                                    defaultValue={status?.settings?.torrent?.defaultTorrentClient}
                                >
                                    <AccordionItem value="qbittorrent">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><ImDownload className="text-blue-400" /> qBittorrent</h4>
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
                                        </AccordionContent>
                                    </AccordionItem>
                                    <AccordionItem value="transmission">
                                        <AccordionTrigger>
                                            <h4 className="flex gap-2 items-center"><ImDownload className="text-orange-200" /> Transmission</h4>
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

                            </TabsContent>

                            <div className="mt-4">
                                <Field.Submit role="save" loading={isPending}>Save</Field.Submit>
                            </div>
                        </div>
                    </Tabs>
                {/*</Card>*/}

            </Form>
        </PageWrapper>
    )

}
