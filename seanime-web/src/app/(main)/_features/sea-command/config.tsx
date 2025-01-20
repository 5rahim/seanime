import React from "react"
import { SeaCommandContextProps, useSeaCommandContext } from "./sea-command"

export type SeaCommandHandlerProps = {
    shouldShow: (props: SeaCommandContextProps) => boolean
    render: (props: SeaCommandContextProps) => React.ReactNode
}

export function SeaCommandHandler(props: SeaCommandHandlerProps) {
    const { shouldShow, render } = props

    const ctx = useSeaCommandContext()

    if (!shouldShow(ctx)) return null

    return render(ctx)
}
