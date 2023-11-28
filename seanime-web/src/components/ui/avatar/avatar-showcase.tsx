import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import React from "react"
import { AvatarProps } from "."

export const AvatarShowcaseAnatomy = defineStyleAnatomy({
    container: cva("UI-AvatarShowcase__container group/container flex items-center"),
    name: cva("UI-AvatarShowcase__name font-medium text-base text-[--text-color] tracking-tight"),
    description: cva("UI-AvatarShowcase__description block text-sm text-[--muted]"),
    detailsContainer: cva("UI-AvatarShowcase__detailsContainer ml-3"),
})

export interface AvatarShowcaseProps extends React.ComponentPropsWithRef<"div">, ComponentWithAnatomy<typeof AvatarShowcaseAnatomy> {
    avatar: React.ReactElement<AvatarProps, string | React.JSXElementConstructor<AvatarProps>> | undefined,
    name: string
    description?: string
}

export const AvatarShowcase = React.forwardRef<HTMLDivElement, AvatarShowcaseProps>((props, ref) => {

    const {
        children,
        className,
        avatar,
        name,
        description,
        nameClassName,
        descriptionClassName,
        detailsContainerClassName,
        containerClassName,
        ...rest
    } = props

    return (
        <>
            <div
                className={cn(
                    AvatarShowcaseAnatomy.container(),
                    containerClassName,
                    className,
                )}
                {...rest}
                ref={ref}
            >
                {avatar}
                <div className={cn(AvatarShowcaseAnatomy.detailsContainer(), detailsContainerClassName)}>
                    <p className={cn(AvatarShowcaseAnatomy.name(), nameClassName)}>{name}</p>
                    {!!description && <span
                        className={cn(AvatarShowcaseAnatomy.description(), descriptionClassName)}>{description}</span>}
                    {children}
                </div>
            </div>
        </>
    )

})
