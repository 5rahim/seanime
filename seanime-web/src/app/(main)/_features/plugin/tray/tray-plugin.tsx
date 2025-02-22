import { PluginProvider, registry, RenderPluginComponents } from "@/app/(main)/_features/plugin/tray/registry"
import { useWebsocketPluginMessageListener, useWebsocketSender } from "@/app/(main)/_hooks/handle-websockets"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { Separator } from "@/components/ui/separator"
import { Tooltip } from "@/components/ui/tooltip"
import React from "react"

type TrayPluginProps = {
    extensionID: string
}

export function TrayPlugin(props: TrayPluginProps) {

    const {
        ...rest
    } = props

    const [open, setOpen] = React.useState(false)

    return (
        <>
            <Popover
                open={open}
                onOpenChange={setOpen}
                side="right"
                trigger={<div>
                    <Tooltip
                        side="right"
                        trigger={
                            <div className="w-8 h-8 rounded-full flex items-center justify-center overflow-hidden hover:bg-gray-600 cursor-pointer transition-all">
                                <div className="w-8 h-8 rounded-full flex items-center justify-center overflow-hidden relative">
                                    <span>T</span>
                                    {/*<Image*/}
                                    {/*    src="https://raw.githubusercontent.com/5rahim/hibike/main/icons/seadex.png"*/}
                                    {/*    alt="logo"*/}
                                    {/*    fill*/}
                                    {/*    className="p-1 w-full h-full object-contain"*/}
                                    {/*/>*/}
                                </div>
                            </div>}
                    >
                        Extension name
                    </Tooltip>
                </div>}
                className="w-[30rem] bg-gray-950/95 min-h-80 p-4 ml-2 shadow-xl rounded-2xl"
            >
                <TrayPluginContent
                    open={open}
                    setOpen={setOpen}
                    {...rest}
                />
            </Popover>
        </>
    )
}

type TrayPluginContentProps = {
    open: boolean
    setOpen: (open: boolean) => void
} & TrayPluginProps

function TrayPluginContent(props: TrayPluginContentProps) {
    const {
        extensionID,
        open,
        setOpen,
    } = props

    const { sendPluginMessage } = useWebsocketSender()

    React.useEffect(() => {
        if (open) {
            sendPluginMessage("tray:render", {})
        }
    }, [open])

    const [data, setData] = React.useState<any>(null)

    useWebsocketPluginMessageListener({
        extensionId: extensionID,
        type: "tray:updated",
        onMessage: (data) => {
            console.log("tray:updated", extensionID, data)
            setData(data)
        },
    })

    return (
        <div>
            <p className="font-bold">
                Extension name
            </p>
            <Separator className="my-2" />

            <div
                className={cn(
                    "max-h-[35rem] overflow-y-auto",
                )}
            >

                <PluginProvider registry={registry}>
                    {!!data && <RenderPluginComponents data={data.components} />}
                </PluginProvider>

            </div>
        </div>
    )
}
