import { DirectorySelector } from "@/components/shared/directory-selector"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import React from "react"
import { BiFolder } from "react-icons/bi"

export interface ScreenshotDirModalProps {
    open: boolean
    onClose: () => void
    onSave: (path: string) => void
    portalContainer?: HTMLElement | null
}

export function ScreenshotDirModal({ open, onClose, onSave, portalContainer }: ScreenshotDirModalProps) {
    const [path, setPath] = React.useState("")

    const handleSave = () => {
        if (path) {
            onSave(path)
            onClose()
        }
    }

    return (
        <Modal
            open={open}
            onOpenChange={v => {
                if (!v) onClose()
            }}
            title="Screenshot Folder"
            contentClass="max-w-md space-y-4"
            portalContainer={portalContainer || undefined}
        >
            <p className="text-sm text-[--muted]">
                Select the folder where you would like to save your video screenshots.
            </p>

            <DirectorySelector
                value={path}
                onSelect={setPath}
                label="Screenshot Folder"
                leftIcon={<BiFolder className="text-[--indigo]" />}
            />

            <div className="flex justify-end gap-2 mt-4">
                <Button intent="gray-basic" onClick={onClose}>
                    Cancel
                </Button>
                <Button intent="white" onClick={handleSave} disabled={!path}>
                    Save
                </Button>
            </div>
        </Modal>
    )
}
