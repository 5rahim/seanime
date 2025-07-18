import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SettingsCard, SettingsPageHeader } from "@/app/(main)/settings/_components/settings-card"
import { SettingsSubmitButton } from "@/app/(main)/settings/_components/settings-submit-button"
import { Alert } from "@/components/ui/alert"
import { cn } from "@/components/ui/core/styling"
import { Field } from "@/components/ui/form"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { useWatch } from "react-hook-form"
import { MdOutlineConnectWithoutContact } from "react-icons/md"
import { RiInformation2Line } from "react-icons/ri"

type Props = {
    isPending: boolean
    children?: React.ReactNode
}
const tabsRootClass = cn("w-full contents space-y-4")

const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white",
    "h-10 lg:justify-center px-3 flex-1",
)

const tabsListClass = cn(
    "w-full flex flex-row lg:flex-row flex-wrap h-fit mt-4",
)

export function NakamaSettings(props: Props) {

    const {
        isPending,
        children,
        ...rest
    } = props

    const serverStatus = useServerStatus()
    const nakamaIsHost = useWatch({ name: "nakamaIsHost" })


    return (
        <div className="space-y-4">

            <SettingsPageHeader
                title="Nakama"
                description="Communicate with other Seanime instances"
                icon={MdOutlineConnectWithoutContact}
            />

            <SettingsCard>
                <Field.Switch
                    side="right"
                    name="nakamaEnabled"
                    label="Enable Nakama"
                />

                <Field.Text
                    label="Username"
                    name="nakamaUsername"
                    placeholder="Username"
                    help="The username to identify this server to other instances. If empty a random ID will be assigned."
                />
            </SettingsCard>

            <SettingsCard title="Connect to a host">
                {serverStatus?.settings?.nakama?.isHost && <Alert intent="info" description="Cannot connect to a host while in host mode." />}

                <div
                    className={cn(
                        "space-y-4",
                        serverStatus?.settings?.nakama?.isHost ? "hidden" : "",
                    )}
                >
                    {!serverStatus?.settings?.nakama?.isHost &&
                        <div className="flex items-center gap-2 text-sm bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-700 border-dashed text-blue-100">
                            <RiInformation2Line className="text-base" />
                            <span>The server you're connecting to must be accessible over the internet.
                            </span>
                        </div>}

                    <Field.Text
                        label="Nakama Server URL"
                        name="nakamaRemoteServerURL"
                        placeholder="http://{address}"
                        help="The URL of the Nakama host to connect to."
                    />

                    <Field.Text
                        label="Nakama Passcode"
                        name="nakamaRemoteServerPassword"
                        placeholder="Passcode"
                        help="The passcode to connect to the Nakama host."
                    />

                    <Separator className="!my-6" />

                    <h3>Library</h3>

                    <Field.Switch
                        side="right"
                        name="includeNakamaAnimeLibrary"
                        label="Use Nakama's anime library"
                        help="If enabled, the Nakama's anime library will be used as your library if it is being shared."
                    />
                </div>
            </SettingsCard>

            <SettingsCard title="Host">
                <div className="flex items-center gap-2 text-sm bg-gray-50 dark:bg-gray-900/30 rounded-lg p-3 border border-gray-700 border-dashed text-blue-100">
                    <RiInformation2Line className="text-base" />
                    <span>Enabling host mode does not automatically set up remote access; you must manually expose your server using your
                          preferred method.</span>
                </div>

                {!serverStatus?.serverHasPassword &&
                    <Alert intent="warning" description="Your server is not password protected. Add a password to your config file." />}

                <Field.Switch
                    side="right"
                    name="nakamaIsHost"
                    label="Enable host mode"
                    // moreHelp="Password must be set in the config file"
                    help="If enabled, this server will act as a host for other clients. This requires a host password to be set."
                />

                <Field.Text
                    label="Host Passcode"
                    name="nakamaHostPassword"
                    placeholder="Passcode"
                    help="Set a passcode to secure your host mode. This passcode should be different than your server password."
                />

                {/*<Field.Switch*/}
                {/*    side="right"*/}
                {/*    name="nakamaHostEnablePortForwarding"*/}
                {/*    label="Enable port forwarding"*/}
                {/*    moreHelp="This might not work for all networks."*/}
                {/*    help="If enabled, this server will expose its port to the internet. This might be required for other clients to connect to this server."*/}
                {/*/>*/}

                {nakamaIsHost && <>
                    <Separator className="!my-6" />

                    <h3>Host settings</h3>

                    <Field.Switch
                        side="right"
                        name="nakamaHostShareLocalAnimeLibrary"
                        label="Share local anime library"
                        help="If enabled, this server will share its local anime library to other clients."
                    />

                    <Field.MediaExclusionSelector
                        name="nakamaHostUnsharedAnimeIds"
                        label="Exclude anime from sharing"
                        help="Select anime that you don't want to share with other clients."
                    />
                </>}
            </SettingsCard>


            <SettingsSubmitButton isPending={isPending} />

        </div>
    )
}
