import { DirectorySelector } from "@/components/shared/directory-selector"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { upath } from "@/lib/helpers/upath"
import React from "react"
import { BiFolder } from "react-icons/bi"

export interface ScreenshotDirModalProps {
    open: boolean
    onClose: () => void
    onSave: (path: string) => Promise<boolean> | void | boolean
    portalContainer?: HTMLElement | null
}

export function ScreenshotDirModal({ open, onClose, onSave, portalContainer }: ScreenshotDirModalProps) {
    const [path, setPath] = React.useState("")
    const [saving, setSaving] = React.useState(false)

    const isAbsolute = React.useMemo(() => {
        if (!path) return true
        return upath.isAbsolute(path)
    }, [path])

    const handleSave = async () => {
        if (path && isAbsolute) {
            setSaving(true)
            try {
                const success = await onSave(path)
                if (success !== false) {
                    onClose()
                }
            }
            catch (e) {
            }
            finally {
                setSaving(false)
            }
        }
    }

    return (
        <Modal
            open={open}
            onOpenChange={v => {
                if (!v && !saving) onClose()
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
                error={!isAbsolute ? "Must be an absolute path" : ""}
            />

            <div className="flex justify-end gap-2 mt-4">
                <Button intent="gray-basic" onClick={onClose} disabled={saving}>
                    Cancel
                </Button>
                <Button intent="white" onClick={handleSave} disabled={!path || !isAbsolute} loading={saving}>
                    Save
                </Button>
            </div>
        </Modal>
    )
}
