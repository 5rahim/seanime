"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const CardAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-Card__root",
        "rounded-xl border bg-[--paper] shadow-sm",
    ]),
    header: cva([
        "UI-Card__header",
        "flex flex-col space-y-1.5 p-4",
    ]),
    title: cva([
        "UI-Card__title",
        "text-2xl font-semibold leading-none tracking-tight",
    ]),
    description: cva([
        "UI-Card__description",
        "text-sm text-[--muted]",
    ]),
    content: cva([
        "UI-Card__content",
        "p-4 pt-0",
    ]),
    footer: cva([
        "UI-Card__footer",
        "flex items-center p-4 pt-0",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * Card
 * -----------------------------------------------------------------------------------------------*/

export type CardProps = React.ComponentPropsWithoutRef<"div">

export const Card = React.forwardRef<HTMLDivElement, CardProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <div
            ref={ref}
            className={cn(CardAnatomy.root(), className)}
            {...rest}
        />
    )
})
Card.displayName = "Card"

/* -------------------------------------------------------------------------------------------------
 * CardHeader
 * -----------------------------------------------------------------------------------------------*/

export type CardHeaderProps = React.ComponentPropsWithoutRef<"div">

export const CardHeader = React.forwardRef<HTMLDivElement, CardHeaderProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <div
            ref={ref}
            className={cn(CardAnatomy.header(), className)}
            {...rest}
        />
    )
})
CardHeader.displayName = "CardHeader"

/* -------------------------------------------------------------------------------------------------
 * CardTitle
 * -----------------------------------------------------------------------------------------------*/

export type CardTitleProps = React.ComponentPropsWithoutRef<"h3">

export const CardTitle = React.forwardRef<HTMLHeadingElement, CardTitleProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <h3
            ref={ref}
            className={cn(CardAnatomy.title(), className)}
            {...rest}
        />
    )
})
CardTitle.displayName = "CardTitle"

/* -------------------------------------------------------------------------------------------------
 * CardDescription
 * -----------------------------------------------------------------------------------------------*/

export type CardDescriptionProps = React.ComponentPropsWithoutRef<"p">

export const CardDescription = React.forwardRef<HTMLParagraphElement, CardDescriptionProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <p
            ref={ref}
            className={cn(CardAnatomy.description(), className)}
            {...rest}
        />
    )
})
CardDescription.displayName = "CardDescription"

/* -------------------------------------------------------------------------------------------------
 * CardContent
 * -----------------------------------------------------------------------------------------------*/

export type CardContentProps = React.ComponentPropsWithoutRef<"div">

export const CardContent = React.forwardRef<HTMLDivElement, CardContentProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <div
            ref={ref}
            className={cn(CardAnatomy.content(), className)}
            {...rest}
        />
    )
})
CardContent.displayName = "CardContent"

/* -------------------------------------------------------------------------------------------------
 * CardFooter
 * -----------------------------------------------------------------------------------------------*/

export type CardFooterProps = React.ComponentPropsWithoutRef<"div">

export const CardFooter = React.forwardRef<HTMLDivElement, CardFooterProps>((props, ref) => {
    const { className, ...rest } = props
    return (
        <div
            ref={ref}
            className={cn(CardAnatomy.footer(), className)}
            {...rest}
        />
    )
})
CardFooter.displayName = "CardFooter"

