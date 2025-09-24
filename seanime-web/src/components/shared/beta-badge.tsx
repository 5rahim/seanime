import { Badge, BadgeProps } from "@/components/ui/badge"
import { cn } from "../ui/core/styling"

type Props = BadgeProps

export function BetaBadge({ className, ...props }: Props) {
    return (
        <Badge intent="warning" size="sm" className={cn("align-middle ml-2 border-transparent", className)} {...props}>Beta</Badge>
    )
}

export function AlphaBadge({ className, ...props }: Props) {
    return (
        <Badge intent="warning" size="sm" className={cn("align-middle ml-2 border-transparent", className)} {...props}>Alpha</Badge>
    )
}
