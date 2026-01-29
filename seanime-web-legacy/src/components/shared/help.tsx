import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import * as React from "react"
import { AiOutlineExclamationCircle } from "react-icons/ai"

type HelpProps = {
    content?: React.ReactNode
    className?: string
    triggerClass?: string
}

export function Help(props: HelpProps) {
    const { content, className, triggerClass } = props
    return (
        <Popover
            className={cn("text-sm", className)}
            trigger={<AiOutlineExclamationCircle className={cn("transition-opacity opacity-45 hover:opacity-90", triggerClass)} />}
        >
            {content}
        </Popover>
    )
}
