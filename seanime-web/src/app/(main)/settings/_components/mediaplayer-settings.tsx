import { useExternalPlayerLink } from "@/app/(main)/_atoms/playback.atoms"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion"
import { Alert } from "@/components/ui/alert"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import { TextInput } from "@/components/ui/text-input"
import { getDefaultMpcSocket } from "@/lib/server/settings"
import React from "react"
import { FcClapperboard, FcVideoCall, FcVlc } from "react-icons/fc"
import { HiPlay } from "react-icons/hi"

type MediaplayerSettingsProps = {
    isPending: boolean
}

export function MediaplayerSettings(props: MediaplayerSettingsProps) {

    const {
        isPending,
    } = props

    const serverStatus = useServerStatus()

    const { externalPlayerLink, setExternalPlayerLink } = useExternalPlayerLink()

    return (
        <>
            <div>
                <h3>External Media Player</h3>

                <p className="text-[--muted]">
                    Manage your external media players.
                </p>
            </div>

            <Separator className="!mt-10" />
            <br />

            <div>
                <h3>Desktop Media Player</h3>

                <p className="text-[--muted]">
                    Seanime has built-in support for MPV, VLC, and MPC-HC.
                </p>
            </div>

            <Field.Select
                name="defaultPlayer"
                label="Default player"
                leftIcon={<FcVideoCall />}
                options={[
                    { label: "MPV", value: "mpv" },
                    { label: "VLC", value: "vlc" },
                    { label: "MPC-HC", value: "mpc-hc" },
                ]}
                help="Player that will be used to open files and track your progress automatically."
            />

            <Separator />

            <h4>Playback</h4>

            <Field.Switch
                name="autoPlayNextEpisode"
                label="Automatically play next episode"
                help="If enabled, Seanime will play the next episode after a delay when the current episode is completed."
            />

            <Separator />


            <h4>
                Configuration
            </h4>

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
                                placeholder={`Default: '${getDefaultMpcSocket(serverStatus?.os ?? "")}'`}
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

            <SettingsSubmitButton isPending={isPending} />

            <Separator className="!mt-10" />
            <br />

            <div>
                <h3>
                    External player link
                </h3>
                <p className="text-[--muted]">
                    Enter a custom scheme format for opening files with an external player on this device.
                    Ensure the player supports HTTP sources.
                </p>
            </div>

            <Alert
                intent="info" description={<>
                This is device-specific.
            </>}
            />

            <TextInput
                label="Custom scheme"
                placeholder="Example: outplayer://{url}"
                value={externalPlayerLink}
                onValueChange={setExternalPlayerLink}
            />

            <p className="italic text-sm text-[--muted]">
                Changes are saved automatically.
            </p>

        </>
    )
}
