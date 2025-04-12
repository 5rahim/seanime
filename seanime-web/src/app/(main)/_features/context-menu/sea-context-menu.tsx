import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { ContextMenuContent } from "@/components/ui/context-menu"
import { ContextMenu } from "@radix-ui/react-context-menu"

export type SeaContextMenuProps = {
    content: React.ReactNode
    children?: React.ReactNode
    availableWhenOffline?: boolean
    hideMenuIf?: boolean
}

export function SeaContextMenu(props: SeaContextMenuProps) {

    const {
        content,
        children,
        availableWhenOffline = true,
        hideMenuIf,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    return (
        <ContextMenu data-sea-context-menu {...rest}>
            {children}

            {(((serverStatus?.isOffline && availableWhenOffline) || !serverStatus?.isOffline) && !hideMenuIf) &&
                <ContextMenuContent className="max-w-xs" data-sea-context-menu-content>
                    {content}
                </ContextMenuContent>}
        </ContextMenu>
    )
}
