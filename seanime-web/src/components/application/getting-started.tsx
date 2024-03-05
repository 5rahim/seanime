import { serverStatusAtom } from "@/atoms/server-status"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Card } from "@/components/ui/card"
import { Field, Form } from "@/components/ui/form"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { getDefaultMpcSocket, settingsSchema, useDefaultSettingsPaths } from "@/lib/server/settings"
import { DEFAULT_TORRENT_PROVIDER, ServerStatus, Settings } from "@/lib/server/types"
import { useSetAtom } from "jotai/react"
import { useRouter } from "next/navigation"
import React, { useEffect, useMemo } from "react"
import { FcClapperboard, FcFolder, FcVideoCall, FcVlc } from "react-icons/fc"
import { HiPlay } from "react-icons/hi"
import { ImDownload } from "react-icons/im"
import { RiFolderDownloadFill } from "react-icons/ri"

export function GettingStarted({ status }: { status: ServerStatus }) {
    const router = useRouter()
    const { getDefaultVlcPath, getDefaultQBittorrentPath, getDefaultTransmissionPath } = useDefaultSettingsPaths()
    const setServerStatus = useSetAtom(serverStatusAtom)

    const { mutate, data, isPending, isSuccess } = useSeaMutation<ServerStatus, Settings>({
        endpoint: SeaEndpoints.SETTINGS,
        mutationKey: ["patch-settings"],
        method: "patch",
    })
    useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
            router.push("/")
        }
    }, [data, isPending])

    useEffect(() => {
        if (isSuccess) {
            router.push("/")
        }
    }, [isSuccess])

    const vlcDefaultPath = useMemo(() => getDefaultVlcPath(status.os), [status.os])
    const qbittorrentDefaultPath = useMemo(() => getDefaultQBittorrentPath(status.os), [status.os])
    const transmissionDefaultPath = useMemo(() => getDefaultTransmissionPath(status.os), [status.os])
    const mpvSocketPath = useMemo(() => getDefaultMpcSocket(status.os), [status.os])

    if (isPending) return <LoadingOverlayWithLogo />

    if (!data) return (
        <div className="container max-w-5xl py-10">
            <div className="mb-4 flex justify-center w-full">
                <img src="/logo.png" alt="logo" className="w-24 h-auto" />
            </div>
            <Card className="relative p-4">
                <AppLayoutStack>
                    <div className="space-y-4">
                        <h3>Getting started</h3>
                        <em className="text-[--muted]">These settings can be modified later.</em>
                        <Form
                            schema={settingsSchema}
                            onSubmit={data => {
                                mutate({
                                    library: {
                                        libraryPath: data.libraryPath,
                                        autoUpdateProgress: false,
                                        disableUpdateCheck: false,
                                        torrentProvider: data.torrentProvider || DEFAULT_TORRENT_PROVIDER,
                                        autoScan: false,
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
                                autoScan: false,
                            }}
                            stackClass="space-y-4"
                        >
                            <Field.DirectorySelector
                                name="libraryPath"
                                label="Library folder"
                                leftIcon={<FcFolder />}
                                shouldExist
                            />

                            <h4 className="text-center">Media Player</h4>

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
                                    <AccordionContent className="p-0 py-4 space-y-4">
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
                                    <AccordionContent className="p-0 py-4 space-y-4">
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
                                    <AccordionContent className="p-0 py-4 space-y-4">
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
                                                name="vlcPath"
                                                label="Executable"
                                            />
                                        </div>
                                    </AccordionContent>
                                </AccordionItem>
                            </Accordion>

                            <h4 className="text-center">Torrent Provider</h4>

                            <Field.Select
                                name="torrentProvider"
                                label="Torrent Provider"
                                leftIcon={<RiFolderDownloadFill className="text-orange-500" />}
                                options={[
                                    { label: "AnimeTosho", value: "animetosho" },
                                    { label: "Nyaa", value: "nyaa" },
                                ]}
                            />

                            <h4 className="text-center">Torrent Client</h4>

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

                            <Field.Submit
                                className="w-full"
                                role="submit"
                                showLoadingOverlayOnSuccess={true}
                                loading={isPending}
                            >Continue</Field.Submit>
                        </Form>
                    </div>
                </AppLayoutStack>
            </Card>
            <p className="text-[--muted] mt-5 text-center">Made by 5rahim</p>
        </div>
    )
}
