import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ContextMenuContent } from "@/components/ui/context-menu"
import { ContextMenu } from "@radix-ui/react-context-menu"

export type SeaContextMenuProps = {
    content: React.ReactNode
    children?: React.ReactNode
    availableWhenOffline?: boolean
}

export function SeaContextMenu(props: SeaContextMenuProps) {

    const {
        content,
        children,
        availableWhenOffline = false,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    return (<ContextMenu {...rest}>
        {children}

        {((serverStatus?.isOffline && availableWhenOffline) || !serverStatus?.isOffline) && <ContextMenuContent className="max-w-xs">
            {content}
        </ContextMenuContent>}
    </ContextMenu>)
}
