import { PluginProvider, registry, RenderPluginComponents } from "@/app/(main)/_features/plugin/components/registry"
import { cn } from "@/components/ui/core/styling"
import { PopoverAnatomy } from "@/components/ui/popover"
import { Tooltip } from "@/components/ui/tooltip"
import * as PopoverPrimitive from "@radix-ui/react-popover"
import React from "react"
import {
    usePluginListenTrayUpdatedEvent,
    usePluginSendTrayClosedEvent,
    usePluginSendTrayOpenedEvent,
    usePluginSendTrayRenderEvent,
} from "../generated/plugin-events"

type TrayPluginProps = {
    extensionID: string
}

export const TrayPluginContext = React.createContext<TrayPluginProps>({
    extensionID: "",
})

function TrayPluginProvider(props: { children: React.ReactNode, props: TrayPluginProps }) {
    return <TrayPluginContext.Provider value={props.props}>
        {props.children}
    </TrayPluginContext.Provider>
}

export function useTrayPlugin() {
    const context = React.useContext(TrayPluginContext)
    if (!context) {
        throw new Error("useTrayPlugin must be used within a TrayPluginProvider")
    }
    return context
}

export function TrayPlugin(props: TrayPluginProps) {

    const {
        ...rest
    } = props

    const [open, setOpen] = React.useState(false)

    const { sendTrayOpenedEvent } = usePluginSendTrayOpenedEvent()
    const { sendTrayClosedEvent } = usePluginSendTrayClosedEvent()

    const firstRender = React.useRef(true)
    React.useEffect(() => {
        if (firstRender.current) {
            firstRender.current = false
            return
        }
        if (open) {
            sendTrayOpenedEvent({}, props.extensionID)
        } else {
            sendTrayClosedEvent({}, props.extensionID)
        }
    }, [open])

    return (
        <>
            <PopoverPrimitive.Root
                open={open}
                onOpenChange={setOpen}
                modal={false}
            >
                <PopoverPrimitive.Trigger
                    asChild
                >
                    <div>
                        <Tooltip
                            side="right"
                            trigger={
                                <div className="w-8 h-8 rounded-full flex items-center justify-center overflow-hidden hover:bg-gray-800 cursor-pointer transition-all">
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
                    </div>
                </PopoverPrimitive.Trigger>
                <PopoverPrimitive.Portal>
                    <PopoverPrimitive.Content
                        sideOffset={10}
                        side="right"
                        className={cn(PopoverAnatomy.root(), "w-[30rem] bg-gray-950 min-h-80 p-0 shadow-xl rounded-xl")}
                        onOpenAutoFocus={(e) => e.preventDefault()}
                    >
                        <TrayPluginProvider props={props}>
                            <TrayPluginContent
                                open={open}
                                setOpen={setOpen}
                                {...rest}
                            />
                        </TrayPluginProvider>
                    </PopoverPrimitive.Content>
                </PopoverPrimitive.Portal>
            </PopoverPrimitive.Root>
            {/*<Popover*/}
            {/*    open={open}*/}
            {/*    onOpenChange={setOpen}*/}
            {/*    side="right"*/}
            {/*    trigger={}*/}
            {/*    className="w-[30rem] bg-gray-950/95 min-h-80 p-4 ml-2 shadow-xl rounded-2xl"*/}
            {/*>*/}
            {/*    <TrayPluginContent*/}
            {/*        open={open}*/}
            {/*        setOpen={setOpen}*/}
            {/*        {...rest}*/}
            {/*    />*/}
            {/*</Popover>*/}
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

    const { sendTrayRenderEvent } = usePluginSendTrayRenderEvent()

    React.useEffect(() => {
        if (open) {
            sendTrayRenderEvent({}, extensionID)
        }
    }, [open])

    const [data, setData] = React.useState<any>(null)

    usePluginListenTrayUpdatedEvent((data) => {
        // console.log("tray:updated", extensionID, data)
        setData(data)
    }, extensionID)

    return (
        <div>
            {/*<p className="font-bold">*/}
            {/*    Extension name*/}
            {/*</p>*/}
            {/*<Separator className="my-2" />*/}

            <div
                className={cn(
                    "max-h-[35rem] overflow-y-auto p-4",
                )}
            >

                <PluginProvider registry={registry}>
                    {!!data && <RenderPluginComponents data={data.components} />}
                </PluginProvider>

            </div>
        </div>
    )
}
