import { OnlinestreamManagerOpts } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import React from "react"

//@ts-ignore
const __LegacyOnlinestreamManagerContext = React.createContext<OnlinestreamManagerOpts["opts"]>({})

export function useLegacyOnlinestreamManagerContext() {
    return React.useContext(__LegacyOnlinestreamManagerContext)
}

export function LegacyOnlinestreamManagerProvider(props: { children?: React.ReactNode, opts: OnlinestreamManagerOpts["opts"] }) {
    return (
        <__LegacyOnlinestreamManagerContext.Provider
            value={props.opts}
        >
            {props.children}
        </__LegacyOnlinestreamManagerContext.Provider>
    )
}
