"use client"

import { Updater_Announcement, Updater_AnnouncementAction, Updater_AnnouncementSeverity } from "@/api/generated/types"
import { useGetAnnouncements } from "@/api/hooks/status.hooks"
import { useWebsocketMessageListener } from "@/app/(main)/_hooks/handle-websockets"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { useUpdateEffect } from "@/components/ui/core/hooks"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { logger } from "@/lib/helpers/debug"
import { WSEvents } from "@/lib/server/ws-events"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { useAtom } from "jotai"
import { atomWithStorage } from "jotai/utils"
import React from "react"
import { FiAlertTriangle, FiInfo } from "react-icons/fi"
import { useEffectOnce } from "react-use"
import { toast } from "sonner"

const dismissedAnnouncementsAtom = atomWithStorage<string[]>("sea-dismissed-announcements", [])

export function Announcements() {
    const { data: announcements, mutate: getAnnouncements } = useGetAnnouncements()
    const [dismissedAnnouncements, setDismissedAnnouncements] = useAtom(dismissedAnnouncementsAtom)
    const [hasShownToasts, setHasShownToasts] = React.useState<string[]>([])

    function handleCheckForAnnouncements() {
        getAnnouncements({
            platform: __isElectronDesktop__ ? "denshi" : __isTauriDesktop__ ? "tauri" : "web",
        })
    }

    useWebsocketMessageListener({
        type: WSEvents.CHECK_FOR_ANNOUNCEMENTS,
        onMessage: () => {
            handleCheckForAnnouncements()
        },
    })

    useEffectOnce(() => {
        handleCheckForAnnouncements()
    })

    useUpdateEffect(() => {
        if (announcements) {
            logger("Announcement").info("Fetched announcements", announcements)
            // Clean up dismissed announcements that are no longer in the announcements
            setDismissedAnnouncements(prev => prev.filter(id => announcements.some(a => a.id === id)))
        }
    }, [announcements])

    const filteredAnnouncements = React.useMemo(() => {
        if (!announcements) return []

        return announcements
            .filter(announcement => {
                if (announcement.notDismissible) return true
                if (dismissedAnnouncements.includes(announcement.id)) return false
                return true
            })
            .sort((a, b) => b.priority - a.priority)
    }, [announcements, dismissedAnnouncements])

    const bannerAnnouncements = filteredAnnouncements.filter(a => a.type === "banner")
    const dialogAnnouncements = filteredAnnouncements.filter(a => a.type === "dialog")
    const toastAnnouncements = filteredAnnouncements.filter(a => a.type === "toast")

    const dismissAnnouncement = (id: string) => {
        setDismissedAnnouncements(prev => [...prev, id])
    }

    const getSeverityIcon = (severity: Updater_AnnouncementSeverity) => {
        switch (severity) {
            case "info":
                return <FiInfo className="size-5 mt-1" />
            case "warning":
                return <FiAlertTriangle className="size-5 mt-1" />
            case "error":
                return <FiAlertTriangle className="size-5 mt-1" />
            case "critical":
                return <FiAlertTriangle className="size-5 mt-1" />
            default:
                return <FiInfo className="size-5 mt-1" />
        }
    }

    const getSeverityIntent = (severity: Updater_AnnouncementSeverity) => {
        switch (severity) {
            case "info":
                return "info-basic"
            case "warning":
                return "warning-basic"
            case "error":
                return "alert-basic"
            case "critical":
                return "alert-basic"
            default:
                return "info-basic"
        }
    }

    const getSeverityBadgeIntent = (severity: Updater_AnnouncementSeverity) => {
        switch (severity) {
            case "info":
                return "blue"
            case "warning":
                return "warning"
            case "error":
                return "alert"
            case "critical":
                return "alert-solid"
            default:
                return "blue"
        }
    }

    React.useEffect(() => {
        toastAnnouncements.forEach(announcement => {
            if (!hasShownToasts.includes(announcement.id)) {
                const toastFunction = announcement.severity === "error" || announcement.severity === "critical"
                    ? toast.error
                    : announcement.severity === "warning"
                        ? toast.warning
                        : toast.info

                toastFunction(announcement.message, {
                    position: "top-right",
                    id: announcement.id,
                    duration: Infinity,
                    action: !announcement.notDismissible ? {
                        label: "OK",
                        onClick: () => dismissAnnouncement(announcement.id),
                    } : undefined,
                    onDismiss: !announcement.notDismissible ? () => dismissAnnouncement(announcement.id) : undefined,
                    onAutoClose: !announcement.notDismissible ? () => dismissAnnouncement(announcement.id) : undefined,
                })

                setHasShownToasts(prev => [...prev, announcement.id])
            }
        })
    }, [toastAnnouncements, hasShownToasts])

    const handleDialogClose = (announcement: Updater_Announcement) => {
        if (!announcement.notDismissible) {
            dismissAnnouncement(announcement.id)
        }
    }

    const handleActionClick = (action: Updater_AnnouncementAction) => {
        if (action.type === "link" && action.url) {
            window.open(action.url, "_blank")
        }
    }

    return (
        <>
            {bannerAnnouncements.map((announcement, index) => (
                <div
                    key={announcement.id + "" + String(index)} className={cn(
                    "fixed bottom-0 left-0 right-0 z-[999] bg-[--background] border-b border-[--border] shadow-lg bg-gradient-to-br",
                )}
                >
                    <Alert
                        intent={getSeverityIntent(announcement.severity) as any}
                        title={announcement.title}

                        description={<div className="space-y-2">
                            <p>
                                {announcement.message}
                            </p>
                            {announcement.actions && announcement.actions.length > 0 && (
                                <div className="flex gap-2">
                                    {announcement.actions.map((action, index) => (
                                        <Button
                                            key={index}
                                            size="sm"
                                            intent="white-outline"
                                            onClick={() => handleActionClick(action)}
                                        >
                                            {action.label}
                                        </Button>
                                    ))}
                                </div>
                            )}
                        </div>}
                        icon={getSeverityIcon(announcement.severity)}
                        onClose={!announcement.notDismissible ? () => dismissAnnouncement(announcement.id) : undefined}
                        className={cn(
                            "rounded-none border-0 border-t shadow-[0_0_10px_0_rgba(0,0,0,0.05)] bg-gradient-to-br",
                            announcement.severity === "critical" && "from-red-950/95 to-red-900/60 dark:text-red-100",
                            announcement.severity === "error" && "from-red-950/95 to-red-900/60 dark:text-red-100",
                            announcement.severity === "warning" && "from-amber-950/95 to-amber-900/60 dark:text-amber-100",
                            announcement.severity === "info" && "from-blue-950/95 to-blue-900/60 dark:text-blue-100",
                        )}
                    />
                </div>
            ))}

            {dialogAnnouncements.map((announcement, index) => (
                <Modal
                    key={announcement.id + "" + String(index)}
                    open={true}
                    onOpenChange={(open) => {
                        if (!open) {
                            handleDialogClose(announcement)
                        }
                    }}
                    hideCloseButton={announcement.notDismissible}
                    title={
                        <div className="flex items-center gap-2">
                            <span
                                className={cn(
                                    announcement.severity === "info" && "text-blue-300",
                                    announcement.severity === "warning" && "text-amber-300",
                                    announcement.severity === "error" && "text-red-300",
                                    announcement.severity === "critical" && "text-red-300",
                                )}
                            >
                                {getSeverityIcon(announcement.severity)}
                            </span>
                            {announcement.title || "Announcement"}
                            {/* <Badge
                             intent={getSeverityBadgeIntent(announcement.severity) as any}
                             size="sm"
                             >
                             {announcement.severity.toUpperCase()}
                             </Badge> */}
                        </div>
                    }
                    overlayClass="bg-gray-950/10"
                >
                    <div className="space-y-4">
                        <p className="text-[--muted]">
                            {announcement.message}
                        </p>

                        <div className="flex gap-2 pt-2">
                            {announcement.actions && announcement.actions.length > 0 && (
                                <div className="flex gap-2 flex-wrap">
                                    {announcement.actions.map((action, index) => (
                                        <Button
                                            key={index}
                                            intent="gray-outline"
                                            onClick={() => handleActionClick(action)}
                                        >
                                            {action.label}
                                        </Button>
                                    ))}
                                </div>
                            )}
                            <div className="flex-1" />
                            <Button
                                intent="white"
                                onClick={() => handleDialogClose(announcement)}
                            >
                                OK
                            </Button>
                        </div>
                    </div>
                </Modal>
            ))}
        </>
    )
}
