import { useScanLocalFiles } from "@/api/hooks/scan.hooks"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { useBoolean } from "@/hooks/use-disclosure"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import React from "react"
import { FiSearch } from "react-icons/fi"

export const __scanner_modalIsOpen = atom(false)
export const __scanner_isScanningAtom = atom(false)


export function ScannerModal() {
    const serverStatus = useServerStatus()
    const [isOpen, setOpen] = useAtom(__scanner_modalIsOpen)
    const [, setScannerIsScanning] = useAtom(__scanner_isScanningAtom)
    const anilistDataOnly = useBoolean(true)
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
            enhanced: !anilistDataOnly.active,
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
                        <h5>Local files</h5>
                        <Switch
                            side="right"
                            label="Skip locked files"
                            value={skipLockedFiles.active}
                            onValueChange={v => skipLockedFiles.set(v as boolean)}
                            // size="lg"
                        />
                        <Switch
                            side="right"
                            label="Skip ignored files"
                            value={skipIgnoredFiles.active}
                            onValueChange={v => skipIgnoredFiles.set(v as boolean)}
                            // size="lg"
                        />
                        <Separator />

                        <AppLayoutStack className="space-y-2">
                            <h5>Matching data</h5>
                            <Switch
                                side="right"
                                label="Use my AniList lists only"
                                moreHelp="Disabling this will cause Seanime to use more requests which may lead to rate limits and slower scanning"
                                // label="Enhanced scanning"
                                value={anilistDataOnly.active}
                                onValueChange={v => anilistDataOnly.set(v as boolean)}
                                // className="data-[state=checked]:bg-amber-700 dark:data-[state=checked]:bg-amber-700"
                                // size="lg"
                                help={!anilistDataOnly.active ? "Caution: Slower for large libraries" : ""}
                            />
                        </AppLayoutStack>

                    </AppLayoutStack>
                </div>
                <Button
                    onClick={handleScan}
                    intent="primary"
                    leftIcon={<FiSearch />}
                    loading={isScanning}
                    className="w-full"
                    disabled={!serverStatus?.settings?.library?.libraryPath}
                >
                    Scan
                </Button>
            </Modal>
        </>
    )

}
