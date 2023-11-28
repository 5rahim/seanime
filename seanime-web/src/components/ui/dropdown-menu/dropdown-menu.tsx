"use client"

import React, { Fragment } from "react"
import {
    cn,
    ComponentWithAnatomy,
    createPolymorphicComponent,
    defineStyleAnatomy,
    getChildDisplayName,
    useMediaQuery,
} from "../core"
import { cva, VariantProps } from "class-variance-authority"
import { Menu, Transition } from "@headlessui/react"
import { Divider, DividerProps } from "../divider"
import { Modal, ModalProps } from "../modal"
import { useDropdownOutOfBounds } from "./use-dropdown-out-of-bounds"
import Link from "next/link"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DropdownMenuAnatomy = defineStyleAnatomy({
    menu: cva([
        "UI-DropdownMenu__menu",
        "relative inline-block text-left",
    ]),
    dropdown: cva([
        "UI-DropdownMenu__dropdown",
        "bg-[--paper] border border-[--border] p-1",
        "absolute z-[100] mt-2 w-56 rounded-[--radius] shadow-md focus:outline-none space-y-1",
    ], {
        variants: {
            top: { true: "", right: "" },
            bottom: { true: "", right: "" },
            left: { true: "", right: "" },
            right: { true: "", right: "" },
        },
        compoundVariants: [
            { bottom: false, className: "origin-top-right right-0" },
            { bottom: true, className: "origin-bottom-right" },
            { left: true, className: "left-0" },
            { bottom: true, left: true, right: false, className: "origin-bottom-right left-0 bottom-0" },
            { right: true, bottom: true, left: false, className: "origin-bottom-right right-0 bottom-0" },
        ]
    }),
    mobileDropdown: cva([
        "DropdownMenu__mobileDropdown",
        "mt-2 space-y-1"
    ]),
    mobilePanel: cva([
        "DropdownMenu__mobilePanel",
        "pt-2 pb-2 pl-4 pr-12"
    ])
})

export const DropdownMenuItemAnatomy = defineStyleAnatomy({
    item: cva(["UI-DropdownMenu__item transition",
        "text-gray-800 dark:text-gray-200 hover:text-black dark:hover:text-white",
        "font-medium group flex w-full items-center rounded-[--radius] px-2 py-2 text-sm gap-2"
    ], {
        variants: {
            active: {
                true: "bg-[--highlight]",
                false: null
            }
        },
        defaultVariants: { active: false }
    })
})

export const DropdownMenuGroupAnatomy = defineStyleAnatomy({
    group: cva(["UI-DropdownMenu__group group",
        "text-gray-800 dark:text-gray-200",
    ]),
    title: cva(["UI-DropdownMenu_title text-[--muted] text-sm font-medium px-2 py-1"]),
    content: cva(["UI-DropdownMenu_content"])
})

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu
 * -----------------------------------------------------------------------------------------------*/

export interface DropdownMenuProps
    extends React.ComponentPropsWithRef<"div">,
        ComponentWithAnatomy<typeof DropdownMenuAnatomy>,
        ComponentWithAnatomy<typeof DropdownMenuItemAnatomy>,
        VariantProps<typeof DropdownMenuAnatomy.dropdown> {
    trigger: React.ReactElement,
    mobilePlacement?: ModalProps["mobilePlacement"]
}

const _DropdownMenu = (props: DropdownMenuProps) => {

    const {
        children,
        trigger,
        menuClassName,
        dropdownClassName,
        mobileDropdownClassName,
        mobilePanelClassName,
        itemClassName,
        className,
        mobilePlacement = "bottom",
        ...rest
    } = props

    const isMobile = useMediaQuery("(max-width: 768px)")

    const [triggerRef, _, triggerSize] = useDropdownOutOfBounds()
    const [componentRef, outOfBounds] = useDropdownOutOfBounds()

    // Pass `itemClassName` to every child
    const itemsWithProps = React.useMemo(() => React.Children.map(children, (child) => {
        if (React.isValidElement(child) && (
            getChildDisplayName(child) === "DropdownMenuItem" ||
            getChildDisplayName(child) === "DropdownMenuGroup" ||
            getChildDisplayName(child) === "DropdownMenuLink")
        ) {
            return React.cloneElement(child, { itemClassName } as any)
        }
        return child
    }), [children])

    const _trigger = React.cloneElement(trigger, { ref: triggerRef })

    return (
        <Menu
            as="div"
            className={cn(
                DropdownMenuAnatomy.menu(),
                menuClassName,
                className
            )}
            {...rest}
        >
            {({ open, close }) => (
                <>
                    <Menu.Button as={Fragment}>
                        {_trigger}
                    </Menu.Button>
                    {/*Desktop*/}
                    {!isMobile && <Transition
                        as={Fragment}
                        enter="transition ease-out duration-100"
                        enterFrom="transform opacity-0 scale-95"
                        enterTo="transform opacity-100 scale-100"
                        leave="transition ease-in duration-75"
                        leaveFrom="transform opacity-100 scale-100"
                        leaveTo="transform opacity-0 scale-95"
                    >
                        <Menu.Items
                            ref={componentRef}
                            className={cn(
                                DropdownMenuAnatomy.dropdown({
                                    top: outOfBounds.top > 0,
                                    bottom: outOfBounds.bottom > 0,
                                    left: outOfBounds.left > 0,
                                    right: outOfBounds.right > 0
                                }),
                                dropdownClassName,
                            )}
                            style={{
                                bottom: outOfBounds.bottom > 0 ? `${triggerSize.height + 8}px` : undefined
                            }}
                        >
                            {itemsWithProps}
                        </Menu.Items>
                    </Transition>}
                    {/*Mobile*/}
                    {isMobile && <Modal
                        isOpen={open}
                        onClose={close}
                        isClosable
                        className="block md:hidden"
                        panelClassName={cn(DropdownMenuAnatomy.mobilePanel(), mobilePanelClassName)}
                        mobilePlacement={mobilePlacement}
                    >
                        <Menu.Items className={cn(DropdownMenuAnatomy.mobileDropdown(), mobileDropdownClassName)}>
                            {itemsWithProps}
                        </Menu.Items>
                    </Modal>}
                </>
            )}
        </Menu>
    )

}

_DropdownMenu.displayName = "DropdownMenu"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu.Item
 * -----------------------------------------------------------------------------------------------*/

interface DropdownMenuItemProps extends React.ComponentPropsWithRef<"button">, ComponentWithAnatomy<typeof DropdownMenuItemAnatomy> {
}

export const DropdownMenuItem: React.FC<DropdownMenuItemProps> = React.forwardRef<HTMLButtonElement, DropdownMenuItemProps>((props, ref) => {

    const { children, itemClassName, className, ...rest } = props

    return <Menu.Item as={Fragment}>
        {({ active }) => (
            <button
                className={cn(DropdownMenuItemAnatomy.item({ active }), itemClassName, className)}
                ref={ref}
                {...rest}
            >
                {children}
            </button>
        )}
    </Menu.Item>

})

DropdownMenuItem.displayName = "DropdownMenuItem"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu.Link
 * - You can change the `a` element to a `Link` if you are using Next.js
 * -----------------------------------------------------------------------------------------------*/

interface DropdownMenuLinkProps extends React.ComponentPropsWithRef<"a">, ComponentWithAnatomy<typeof DropdownMenuItemAnatomy> {
    href: string
}

export const DropdownMenuLink: React.FC<DropdownMenuLinkProps> = React.forwardRef<HTMLAnchorElement, DropdownMenuLinkProps>((props, ref) => {

    const { children, className, itemClassName, href, ...rest } = props

    return <Menu.Item as={Fragment}>
        {({ active }) => (
            <Link
                href={href}
                className={cn(DropdownMenuItemAnatomy.item({ active }), itemClassName, className)}
                ref={ref}
                {...rest}
            >
                {children}
            </Link>
        )}
    </Menu.Item>

})

DropdownMenuLink.displayName = "DropdownMenuLink"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu.Group
 * -----------------------------------------------------------------------------------------------*/

interface DropdownMenuGroupProps extends React.ComponentPropsWithRef<"div">,
    ComponentWithAnatomy<typeof DropdownMenuGroupAnatomy>,
    ComponentWithAnatomy<typeof DropdownMenuItemAnatomy> {
    title?: string
}

export const DropdownMenuGroup: React.FC<DropdownMenuGroupProps> = React.forwardRef<HTMLDivElement, DropdownMenuGroupProps>((props, ref) => {

    const {
        children,
        className,
        groupClassName,
        title,
        titleClassName,
        contentClassName,
        itemClassName,
        ...rest
    } = props

    // Pass `itemClassName` to every child
    const itemsWithProps = React.useMemo(() => React.Children.map(children, (child) => {
        if (React.isValidElement(child) && (
            getChildDisplayName(child) === "DropdownMenuItem" ||
            getChildDisplayName(child) === "DropdownMenuGroup" ||
            getChildDisplayName(child) === "DropdownMenuLink")
        ) {
            return React.cloneElement(child, { itemClassName } as any)
        }
        return child
    }), [children])

    return <div
        className={cn(DropdownMenuGroupAnatomy.group(), groupClassName, className)}
        aria-label={title}
        ref={ref}
        {...rest}
    >
        {title && <div className={cn(DropdownMenuGroupAnatomy.title(), titleClassName)} aria-labelledby={title}>
            {title}
        </div>}
        <div className={cn(DropdownMenuGroupAnatomy.content(), contentClassName)}>
            {itemsWithProps}
        </div>
    </div>

})

DropdownMenuGroup.displayName = "DropdownMenuGroup"

/* -------------------------------------------------------------------------------------------------
 * DropdownMenu.Divider
 * -----------------------------------------------------------------------------------------------*/

interface DropdownMenuDivider extends DividerProps {
}

export const DropdownMenuDivider: React.FC<DropdownMenuDivider> = React.forwardRef<HTMLHRElement, DropdownMenuDivider>(
    (props, ref) => {

        return <Divider {...props} ref={ref}/>

    }
)

DropdownMenuDivider.displayName = "DropdownMenuDivider"


/* -------------------------------------------------------------------------------------------------
 * Component
 * -----------------------------------------------------------------------------------------------*/

_DropdownMenu.Item = DropdownMenuItem
_DropdownMenu.Link = DropdownMenuLink
_DropdownMenu.Group = DropdownMenuGroup
_DropdownMenu.Divider = DropdownMenuDivider

export const DropdownMenu = createPolymorphicComponent<"div", DropdownMenuProps, {
    Item: typeof DropdownMenuItem
    Link: typeof DropdownMenuLink
    Group: typeof DropdownMenuGroup
    Divider: typeof DropdownMenuDivider
}>(_DropdownMenu)

DropdownMenu.displayName = "DropdownMenu"
