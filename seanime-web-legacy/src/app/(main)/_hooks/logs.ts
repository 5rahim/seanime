import { useGetLatestLogContent } from "@/api/hooks/status.hooks"

export function useHandleCopyLatestLogs() {

    const { mutate: fetchLogs, data, isPending } = useGetLatestLogContent()

    function handleCopyLatestLogs() {
        fetchLogs()
    }

    return {
        handleCopyLatestLogs,
    }
}
