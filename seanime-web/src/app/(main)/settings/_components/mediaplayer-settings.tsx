import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Alert } from "@/components/ui/alert"
import { Field } from "@/components/ui/form"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { getDefaultIinaSocket, getDefaultMpvSocket } from "@/lib/server/settings"
import React from "react"
import { useWatch } from "react-hook-form"
import { FcClapperboard, FcVideoCall, FcVlc } from "react-icons/fc"
import { HiPlay } from "react-icons/hi"
import { IoPlayForwardCircleSharp } from "react-icons/io5"
import { LuExternalLink, LuLaptop } from "react-icons/lu"
import { RiSettings3Fill } from "react-icons/ri"

type MediaplayerSettingsProps = {
    isPending: boolean
}

export function MediaplayerSettings(props: MediaplayerSettingsProps) {

    const {
        isPending,
    } = props

    const serverStatus = useServerStatus()
    const selectedPlayer = useWatch({ name: "defaultPlayer" })

    return (
        <>
            <SettingsPageHeader
                title="Desktop Media Player"
                description="Seanime has built-in support for MPV, VLC, IINA, and MPC-HC."
                icon={LuLaptop}
            />

            <SettingsCard>
                <Field.Select
                    name="defaultPlayer"
                    label="Default player"
                    leftIcon={<FcVideoCall />}
                    options={[
                        { label: "MPV", value: "mpv" },
                        { label: "VLC", value: "vlc" },
                        { label: "MPC-HC (Windows)", value: "mpc-hc" },
                        { label: "IINA (macOS)", value: "iina" },
                    ]}
                    help="Player that will be used to open files and track your progress automatically."
                />
                {selectedPlayer === "iina" && <Alert
                    intent="info-basic"
                    description={<p>For IINA to work correctly with Seanime, make sure <strong>Quit after all windows are closed</strong> is <span
                        className="underline"
                    >checked</span> and <strong>Keep window open after playback finishes</strong> is <span className="underline">unchecked</span> in
                                    your IINA general settings.</p>}
                />}
            </SettingsCard>

            <SettingsCard title="Playback">
                <Field.Switch
                    side="right"
                    name="autoPlayNextEpisode"
                    label="Automatically play next episode"
                    help="If enabled, Seanime will play the next episode after a delay when the current episode is completed."
                />
            </SettingsCard>

            <SettingsCard title="Configuration">


                <Field.Text
                    name="mediaPlayerHost"
                    label="Host"
                    help="VLC/MPC-HC"
                />

                <Accordion
                    type="single"
                    className=""
                    triggerClass="text-[--muted] dark:data-[state=open]:text-white px-0 dark:hover:bg-transparent hover:bg-transparent dark:hover:text-white hover:text-black"
                    itemClass=""
                    contentClass="p-4 border rounded-[--radius-md]"
                    collapsible
                    defaultValue={serverStatus?.settings?.mediaPlayer?.defaultPlayer}
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
                                    placeholder={`Default: '${getDefaultMpvSocket(serverStatus?.os ?? "")}'`}
                                />
                                <Field.Text
                                    name="mpvPath"
                                    label="Application path"
                                    placeholder={serverStatus?.os === "windows" ? "e.g. C:/Program Files/mpv/mpv.exe" : serverStatus?.os === "darwin"
                                        ? "e.g. /Applications/mpv.app/Contents/MacOS/mpv"
                                        : "Defaults to CLI"}
                                    help="Leave empty to use the CLI."
                                />
                            </div>
                            <div>
                                <Field.Text
                                    name="mpvArgs"
                                    label="Options"
                                    placeholder="e.g. --no-config --mute=yes"
                                />
                            </div>
                        </AccordionContent>
                    </AccordionItem>

                    <AccordionItem value="iina">
                        <AccordionTrigger>
                            <h4 className="flex gap-2 items-center"><IoPlayForwardCircleSharp className="mr-1 text-purple-100" /> IINA</h4>
                        </AccordionTrigger>
                        <AccordionContent>
                            <div className="flex gap-4">
                                <Field.Text
                                    name="iinaSocket"
                                    label="Socket"
                                    placeholder={`Default: '${getDefaultIinaSocket(serverStatus?.os ?? "")}'`}
                                />
                                <Field.Text
                                    name="iinaPath"
                                    label="CLI path"
                                    placeholder={"Path to the IINA CLI"}
                                    help="Leave empty to use the CLI."
                                />
                            </div>
                            <div>
                                <Field.Text
                                    name="iinaArgs"
                                    label="Options"
                                    placeholder="e.g. --mpv-mute=yes"
                                />
                            </div>
                        </AccordionContent>
                    </AccordionItem>
                </Accordion>
            </SettingsCard>

            <SettingsSubmitButton isPending={isPending} />

        </>
    )
}

export function ExternalPlayerLinkSettings() {

    const { externalPlayerLink, setExternalPlayerLink, encodePath, setEncodePath } = useExternalPlayerLink()

    return (
        <>
            <SettingsPageHeader
                title="External player link"
                description="Send streams to an external player on this device."
                icon={LuExternalLink}
            />

            <Alert
                intent="info" description={<>
                Only applies to this device.
            </>}
            />

            <SettingsCard>
                <TextInput
                    label="Custom scheme"
                    placeholder="Example: outplayer://{url}"
                    value={externalPlayerLink}
                    onValueChange={setExternalPlayerLink}
                />
            </SettingsCard>

            <SettingsCard>
                <Switch
                    side="right"
                    name="encodePath"
                    label="Encode file path in URL (library only)"
                    help="If enabled, the file path will be base64 encoded in the URL to avoid issues with special characters."
                    value={encodePath}
                    onValueChange={setEncodePath}
                />
            </SettingsCard>

            <div className="flex items-center gap-2 text-sm text-gray-500 bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-200 dark:border-gray-800 border-dashed">
                <RiSettings3Fill className="text-base" />
                <span>Settings are saved automatically</span>
            </div>
        </>
    )
}
