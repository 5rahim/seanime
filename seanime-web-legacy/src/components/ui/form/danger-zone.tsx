"use client"

import { cva } from "class-variance-authority"
import * as React from "react"
import { Button } from "../button"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"
import { LoadingOverlay } from "../loading-spinner"
import { Modal } from "../modal"
import locales from "./locales.json"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const DangerZoneAnatomy = defineStyleAnatomy({
    root: cva([
        "UI-DangerZone__root",
        "p-4 flex flex-col sm:flex-row gap-2 text-center sm:text-left rounded-[--radius-md] border border-dashed",
    ]),
    icon: cva([
        "UI-DangerZone__icon",
        "place-self-center sm:place-self-start text-[--red] w-4 mt-2",
    ]),
    title: cva([
        "UI-DangerZone__title",
        "text-lg text-[--red] font-semibold",
    ]),
    dialogTitle: cva([
        "UI-DangerZone__dialogTitle",
        "text-lg font-medium leading-6",
    ]),
    dialogBody: cva([
        "UI-DangerZone__dialogBody",
        "mt-2 text-base text-[--muted]",
    ]),
    dialogAction: cva([
        "UI-DangerZone__dialogAction",
        "mt-2 flex gap-2 justify-end",
    ]),
})

/* -------------------------------------------------------------------------------------------------
 * DangerZone
 * -----------------------------------------------------------------------------------------------*/

export type DangerZoneProps = React.ComponentPropsWithRef<"div"> & ComponentAnatomy<typeof DangerZoneAnatomy> & {
    /**
     * Description of the action that will be performed when the delete button is clicked.
     */
    actionText: string
    /**
     * Callback fired when the delete button is clicked.
     */
    onDelete?: () => void
    /**
     * If true, a loading overlay will be shown when the delete button is clicked.
     * @default true
     **/
    showLoadingOverlayOnDelete?: boolean
    locale?: "fr" | "en"
}

export const DangerZone = React.forwardRef<HTMLDivElement, DangerZoneProps>((props, ref) => {

    const {
        children,
        actionText,
        onDelete,
        className,
        locale = "en",
        showLoadingOverlayOnDelete = true,
        titleClass,
        iconClass,
        dialogBodyClass,
        dialogTitleClass,
        dialogActionClass,
        ...rest
    } = props

    const [isOpen, setIsOpen] = React.useState(false)

    const [blockScreen, setBlockScreen] = React.useState<boolean>(false)

    return (
        <>
            <LoadingOverlay hide={!blockScreen} />

            <div ref={ref} className={cn(DangerZoneAnatomy.root(), className)} {...rest}>
                <span className={cn(DangerZoneAnatomy.icon(), iconClass)}>
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor">
                        <path
                            d="M6.457 1.047c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0 1 14.082 15H1.918a1.75 1.75 0 0 1-1.543-2.575Zm1.763.707a.25.25 0 0 0-.44 0L1.698 13.132a.25.25 0 0 0 .22.368h12.164a.25.25 0 0 0 .22-.368Zm.53 3.996v2.5a.75.75 0 0 1-1.5 0v-2.5a.75.75 0 0 1 1.5 0ZM9 11a1 1 0 1 1-2 0 1 1 0 0 1 2 0Z"
                        ></path>
                    </svg>
                </span>
                <div>
                    <h2 className={cn(DangerZoneAnatomy.title(), titleClass)}>{locales["dangerZone"]["name"][locale]}</h2>
                    <p className=""><span
                        className="font-semibold"
                    >{actionText}</span>. {locales["dangerZone"]["irreversible_action"][locale]}
                    </p>
                    <Button
                        size="sm"
                        intent="alert-subtle"
                        className="mt-2"
                        leftIcon={<span className="w-4">
                            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor">
                                <path
                                    d="M11 1.75V3h2.25a.75.75 0 0 1 0 1.5H2.75a.75.75 0 0 1 0-1.5H5V1.75C5 .784 5.784 0 6.75 0h2.5C10.216 0 11 .784 11 1.75ZM4.496 6.675l.66 6.6a.25.25 0 0 0 .249.225h5.19a.25.25 0 0 0 .249-.225l.66-6.6a.75.75 0 0 1 1.492.149l-.66 6.6A1.748 1.748 0 0 1 10.595 15h-5.19a1.75 1.75 0 0 1-1.741-1.575l-.66-6.6a.75.75 0 1 1 1.492-.15ZM6.5 1.75V3h3V1.75a.25.25 0 0 0-.25-.25h-2.5a.25.25 0 0 0-.25.25Z"
                                ></path>
                            </svg>
                        </span>}
                        onClick={() => setIsOpen(true)}
                    >{locales["dangerZone"]["delete"][locale]}</Button>
                </div>
            </div>

            <Modal open={isOpen} onOpenChange={open => setIsOpen(open)} contentClass="gap-2">
                <h3 className={cn(DangerZoneAnatomy.dialogTitle(), dialogTitleClass)}>
                    {locales["dangerZone"]["confirm_delete"][locale]}
                </h3>
                <div className={cn(DangerZoneAnatomy.dialogBody(), dialogBodyClass)}>
                    {locales["dangerZone"]["irreversible_action"][locale]}
                </div>

                <div className={cn(DangerZoneAnatomy.dialogAction(), dialogActionClass)}>
                    <Button
                        intent="alert" size="sm" onClick={() => {
                        setIsOpen(false)
                        showLoadingOverlayOnDelete && setBlockScreen(true)
                        onDelete && onDelete()
                    }}
                    >{locales["dangerZone"]["delete"][locale]}</Button>
                    <Button
                        intent="gray-outline"
                        size="sm"
                        onClick={() => setIsOpen(false)}
                    >{locales["dangerZone"]["cancel"][locale]}</Button>
                </div>
            </Modal>
        </>
    )

})
