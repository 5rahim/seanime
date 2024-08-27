import { useScanLocalFiles } from "@/api/hooks/scan.hooks"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { useBoolean } from "@/hooks/use-disclosure"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { FiSearch } from "react-icons/fi"
import { HiOutlineSparkles } from "react-icons/hi"

export const __scanner_modalIsOpen = atom(false)
export const __scanner_isScanningAtom = atom(false)


export function ScannerModal() {
    const [isOpen, setOpen] = useAtom(__scanner_modalIsOpen)
    const [, setScannerIsScanning] = useAtom(__scanner_isScanningAtom)
    const enhanced = useBoolean(false)
    const skipLockedFiles = useBoolean(true)
    const skipIgnoredFiles = useBoolean(true)

    const { mutate: scanLibrary, isPending: isScanning } = useScanLocalFiles(() => {
        setOpen(false)
    })

    React.useEffect(() => {
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
                title="Library scanner"
                titleClass="text-center"
                contentClass="space-y-4 max-w-2xl"
            >

                <div
                    className="!mt-0 bg-[url(/pattern-2.svg)] z-[-1] w-full h-[4rem] absolute opacity-60 top-0 left-0 bg-no-repeat bg-right bg-cover"
                >
                    <div
                        className="w-full absolute top-0 h-full bg-gradient-to-t from-[--background] to-transparent z-[-2]"
                    />
                </div>

                <div className="space-y-4 mt-6">

                    <AppLayoutStack className="space-y-2">
                        <h5>Matching</h5>
                        <Checkbox
                            label={<span className="flex items-center">Enhanced scanning
                                <HiOutlineSparkles className="ml-2 text-amber-500" /></span>}
                            // label="Enhanced scanning"
                            value={enhanced.active}
                            onValueChange={v => enhanced.set(v as boolean)}
                            className="data-[state=checked]:bg-amber-700 dark:data-[state=checked]:bg-amber-700"
                            // size="lg"
                            help={enhanced.active ? "On: Use API requests, accurate but slower" : "Off: Use AniList account data only, faster"}
                        />
                    </AppLayoutStack>

                    <Separator />

                    <AppLayoutStack className="space-y-2">
                        <h5>Local files</h5>
                        <Checkbox
                            label="Skip locked files"
                            value={skipLockedFiles.active}
                            onValueChange={v => skipLockedFiles.set(v as boolean)}
                            // size="lg"
                        />
                        <Checkbox
                            label="Skip ignored files"
                            value={skipIgnoredFiles.active}
                            onValueChange={v => skipIgnoredFiles.set(v as boolean)}
                            // size="lg"
                        />
                    </AppLayoutStack>
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
