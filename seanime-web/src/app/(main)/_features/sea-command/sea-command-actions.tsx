import { useTorrentstreamDropTorrent } from "@/api/hooks/torrentstream.hooks"
import { __issueReport_overlayOpenAtom, __issueReport_recordingAtom } from "@/app/(main)/_features/issue-report/issue-report"
import { useHandleCopyLatestLogs } from "@/app/(main)/_hooks/logs"
import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import { useSetAtom } from "jotai/react"
import React from "react"
import { useSeaCommandContext } from "./sea-command"

export function SeaCommandActions() {

    const { input, select, command: { isCommand, command, args }, scrollToTop, close } = useSeaCommandContext()

    const setIssueRecorderOpen = useSetAtom(__issueReport_overlayOpenAtom)
    const setIssueRecorderIsRecording = useSetAtom(__issueReport_recordingAtom)

    const { handleCopyLatestLogs } = useHandleCopyLatestLogs()
    const { mutate: dropTorrent, isPending: droppingTorrent } = useTorrentstreamDropTorrent()

    const reloadPage = () => {
        window.location.reload()
    }

    return (
        <>
            {command === "logs" && (
                <CommandGroup heading="Actions">
                    <CommandItem
                        value="Logs"
                        onSelect={() => {
                            select(() => {
                                handleCopyLatestLogs()
                            })
                        }}
                    >
                        Copy current server logs
                        <CommandShortcut>Enter</CommandShortcut>
                    </CommandItem>
                </CommandGroup>
            )}
            {command === "issue" && (
                <CommandGroup heading="Actions">
                    <CommandItem
                        value="Issue"
                        onSelect={() => {
                            select(() => {
                                close()
                                React.startTransition(() => {
                                    setIssueRecorderOpen(true)
                                    setTimeout(() => {
                                        setIssueRecorderIsRecording(true)
                                    }, 500)
                                })
                            })
                        }}
                    >
                        Record an issue
                        <CommandShortcut>Enter</CommandShortcut>
                    </CommandItem>
                </CommandGroup>
            )}
            {command === "droptorrent" && (
                <CommandGroup heading="Actions">
                    <CommandItem
                        value="Drop Torrent"
                        onSelect={() => {
                            close()
                            select(() => {
                                dropTorrent(undefined, {
                                    onSuccess: () => {
                                    },
                                })
                            })
                        }}
                    >
                        Drop all torrents from the torrent streaming client
                        <CommandShortcut>{droppingTorrent ? "Dropping..." : "Enter"}</CommandShortcut>
                    </CommandItem>
                </CommandGroup>
            )}
            {command === "reload" && (
                <CommandGroup heading="Actions">
                    <CommandItem
                        value="Reload Page"
                        onSelect={() => {
                            close()
                            select(() => {
                                reloadPage()
                            })
                        }}
                    >
                        Reload the page
                        <CommandShortcut>Enter</CommandShortcut>
                    </CommandItem>
                </CommandGroup>
            )}
        </>
    )
}
