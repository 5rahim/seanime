import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { useBoolean } from "@/hooks/use-disclosure"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile } from "@/lib/server/types"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React, { useEffect } from "react"
import { FiSearch } from "react-icons/fi"
import { HiOutlineSparkles } from "react-icons/hi"
import { toast } from "sonner"

export const _scannerModalIsOpen = atom(false)
export const _scannerIsScanningAtom = atom(false)

type ScanLibraryProps = {
    enhanced: boolean,
    skipLockedFiles: boolean,
    skipIgnoredFiles: boolean
}

export function ScannerModal() {
    const qc = useQueryClient()

    const [isOpen, setOpen] = useAtom(_scannerModalIsOpen)
    const [, setScannerIsScanning] = useAtom(_scannerIsScanningAtom)
    const enhanced = useBoolean(false)
    const skipLockedFiles = useBoolean(true)
    const skipIgnoredFiles = useBoolean(true)

    // Return data is ignored
    const { mutate: scanLibrary, isPending: isScanning } = useSeaMutation<LocalFile[], ScanLibraryProps>({
        endpoint: SeaEndpoints.SCAN_LIBRARY,
        mutationKey: ["scan-library"],
        onSuccess: async () => {
            toast.success("Library scanned")
            await qc.refetchQueries({ queryKey: ["get-library-collection"] })
            await qc.refetchQueries({ queryKey: ["get-missing-episodes"] })
            await qc.refetchQueries({ queryKey: ["auto-downloader-items"] })
            setOpen(false)
        },
    })

    useEffect(() => {
        setScannerIsScanning(isScanning)
    }, [isScanning])

    function handleScan() {
        scanLibrary({
            enhanced: enhanced.active,
            skipLockedFiles: skipLockedFiles.active,
            skipIgnoredFiles: skipIgnoredFiles.active,
        })
    }

    return (
        <>
            <Modal
                open={isOpen}
                onOpenChange={o => {
                    if (!isScanning) {
                        setOpen(o)
                    }
                }}
                title="Scanner"
                titleClass="text-center"
                contentClass="space-y-4 max-w-2xl overflow-hidden"
            >

                <div
                    className="bg-[url(/pattern-2.svg)] z-[-1] w-full h-[10rem] absolute opacity-70 top-[-5rem] left-0 bg-no-repeat bg-right bg-contain"
                >
                    <div
                        className="w-full absolute bottom-0 h-[10rem] bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>

                <div className="space-y-4 mt-6">

                    <div>
                        <Checkbox
                            label={<span className="flex items-center">Enable enhanced scanning
                                <HiOutlineSparkles className="ml-2 text-amber-500" /></span>}
                            value={enhanced.active}
                            onValueChange={v => enhanced.set(v as boolean)}
                            className="data-[state=checked]:bg-amber-700 dark:data-[state=checked]:bg-amber-700"
                            size="lg"
                        />

                        {enhanced.active && <ul className="list-disc pl-14">
                            <li>Your Anilist anime list data is <strong>not needed</strong></li>
                            <li>Scanning will slow down considerably due to rate limits</li>
                        </ul>}
                    </div>

                    <Separator />

                    <div className="space-y-2">
                        <Checkbox
                            label="Skip locked files"
                            value={skipLockedFiles.active}
                            onValueChange={v => skipLockedFiles.set(v as boolean)}
                            // size="lg"
                        />
                        {/*<Checkbox*/}
                        {/*    label="Skip ignored files"*/}
                        {/*    checked={skipIgnoredFiles.active}*/}
                        {/*    onChange={skipIgnoredFiles.toggle}*/}
                        {/*    // size="lg"*/}
                        {/*/>*/}
                    </div>
                </div>
                <Button
                    onClick={handleScan}
                    intent="primary"
                    leftIcon={<FiSearch />}
                    loading={isScanning}
                    className="w-full"
                >
                    Scan
                </Button>
            </Modal>
        </>
    )

}
