import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Divider } from "@/components/ui/divider"
import { Modal } from "@/components/ui/modal"
import { useBoolean } from "@/hooks/use-disclosure"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { LocalFile } from "@/lib/server/types"
import { FiSearch } from "@react-icons/all-files/fi/FiSearch"
import { HiOutlineSparkles } from "@react-icons/all-files/hi/HiOutlineSparkles"
import { useQueryClient } from "@tanstack/react-query"
import { atom } from "jotai"
import { useAtom } from "jotai/react"
import { useEffect } from "react"
import toast from "react-hot-toast"

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
                                <HiOutlineSparkles className={"ml-2 text-amber-500"} /></span>}
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

                    <Divider />

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
                    leftIcon={<FiSearch />}
                    isLoading={isScanning}
                    className={"w-full"}
                >
                    Scan
                </Button>
            </Modal>
        </>
    )

}
