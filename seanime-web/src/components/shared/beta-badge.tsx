import { Badge, BadgeProps } from "@/components/ui/badge"

type Props = BadgeProps

export function BetaBadge(props: Props) {
    return (
        <Badge intent="warning" size="sm" className="align-middle ml-2 border-transparent" {...props}>Beta</Badge>
    )
}

export function AlphaBadge(props: Props) {
    return (
        <Badge intent="warning" size="sm" className="align-middle ml-2 border-transparent" {...props}>Alpha</Badge>
    )
}
