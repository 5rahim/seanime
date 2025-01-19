import { SeaCommandContextProps, useSeaCommandContext } from "./sea-command"
import { SeaCommandPage } from "./sea-command.atoms"

export type SeaCommandHandlerProps<T extends SeaCommandPage> = {
    type: T
    shouldShow: (props: SeaCommandContextProps<T>) => boolean
    render: (props: SeaCommandContextProps<T>) => React.ReactNode
}

export function SeaCommandHandler<T extends SeaCommandPage>(props: SeaCommandHandlerProps<T>) {
    const { type, shouldShow, render } = props

    const ctx = useSeaCommandContext<T>()

    if (!shouldShow(ctx)) return null

    return render(ctx)
}
