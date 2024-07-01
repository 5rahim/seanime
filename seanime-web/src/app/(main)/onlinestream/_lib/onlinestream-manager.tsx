import { OnlinestreamManagerOpts } from "@/app/(main)/onlinestream/_lib/handle-onlinestream"
import React from "react"

//@ts-ignore
const __OnlinestreamManagerContext = React.createContext<OnlinestreamManagerOpts["opts"]>({})

export function useOnlinestreamManagerContext() {
    return React.useContext(__OnlinestreamManagerContext)
}

export function OnlinestreamManagerProvider(props: { children?: React.ReactNode, opts: OnlinestreamManagerOpts["opts"] }) {
    return (
        <__OnlinestreamManagerContext.Provider
            value={props.opts}
        >
            {props.children}
        </__OnlinestreamManagerContext.Provider>
    )
}
