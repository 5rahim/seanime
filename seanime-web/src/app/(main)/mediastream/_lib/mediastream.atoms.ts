import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"

const __mediastream_filePath = atomWithStorage("sea-mediastream-filepath", "")

export function useMediastreamCurrentFile() {
    const [filePath, setFilePath] = useAtom(__mediastream_filePath)

    return {
        filePath,
        setFilePath,
    }
}
