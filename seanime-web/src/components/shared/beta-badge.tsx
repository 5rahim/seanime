import { Badge, BadgeProps } from "@/components/ui/badge"

type Props = BadgeProps

export function BetaBadge(props: Props) {
    return (
        <Badge intent="warning" size="sm" className="align-middle ml-1.5" {...props}>Experimental</Badge>
    )
}

export function AlphaBadge(props: Props) {
    return (
        <Badge intent="warning" size="sm" className="align-middle ml-1.5" {...props}>Alpha</Badge>
    )
}
