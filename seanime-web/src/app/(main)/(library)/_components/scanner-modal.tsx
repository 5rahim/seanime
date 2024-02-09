import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Divider } from "@/components/ui/divider"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import { useScanLibrary } from "@/lib/server/hooks/library"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { HiOutlineSparkles } from "@react-icons/all-files/hi/HiOutlineSparkles"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useEffect } from "react"

export const _scannerModalIsOpen = atom(false)
export const _scannerIsScanningAtom = atom(false)


export function ScannerModal() {

    const [isOpen, setOpen] = useAtom(_scannerModalIsOpen)
    const [, setScannerIsScanning] = useAtom(_scannerIsScanningAtom)
    const enhanced = useBoolean(false)
    const skipLockedFiles = useBoolean(true)
    const skipIgnoredFiles = useBoolean(true)

    const { scanLibrary, isScanning } = useScanLibrary({
        onSuccess: () => {
            setOpen(false)
        }
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
                isOpen={isOpen}
                onClose={() => setOpen(false)}
                isClosable={!isScanning}
                title={<h3>Scan library</h3>}
                titleClassName={"text-center"}
                bodyClassName={"space-y-4"}
                size={"xl"}
            >

                <div className={"space-y-4 mt-6"}>

                    <div>
                        <Checkbox
                            label={<span className={"flex items-center"}>Enable enhanced scanning
                                <HiOutlineSparkles className={"ml-2 text-amber-500"}/></span>}
                            checked={enhanced.active}
                            onChange={enhanced.toggle}
                            controlClassName={"data-[state=checked]:bg-amber-700 dark:data-[state=checked]:bg-amber-700"}
                            size={"lg"}
                        />

                        {enhanced.active && <ul className={"list-disc pl-14"}>
                            <li>Your Anilist anime list data is <strong>not needed</strong></li>
                            <li>Scanning will slow down considerably due to rate limits</li>
                        </ul>}
                    </div>

                    <Divider/>

                    <div className={"space-y-2"}>
                        <Checkbox
                            label={"Skip locked files"}
                            checked={skipLockedFiles.active}
                            onChange={skipLockedFiles.toggle}
                            // size={"lg"}
                        />
                        {/*<Checkbox*/}
                        {/*    label={"Skip ignored files"}*/}
                        {/*    checked={skipIgnoredFiles.active}*/}
                        {/*    onChange={skipIgnoredFiles.toggle}*/}
                        {/*    // size={"lg"}*/}
                        {/*/>*/}
                    </div>
                </div>
                <Button
                    onClick={handleScan}
                    intent={"primary"}
                    leftIcon={<FiSearch/>}
                    isLoading={isScanning}
                    className={"w-full"}
                >
                    Scan
                </Button>
            </Modal>
        </>
    )

}
