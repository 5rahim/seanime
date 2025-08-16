import { useServerMutation, useServerQuery } from "@/api/client/requests"
import { getServerBaseUrl } from "@/api/client/server-url"
import { DeleteLogs_Variables, GetAnnouncements_Variables } from "@/api/generated/endpoint.types"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { MemoryStatsResponse, Status, Updater_Announcement } from "@/api/generated/types"
import { serverAuthTokenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { copyToClipboard } from "@/lib/helpers/browser"
import { __isDesktop__ } from "@/types/constants"
import { useQueryClient } from "@tanstack/react-query"
import { useAtomValue } from "jotai"
import { toast } from "sonner"

export function useGetStatus() {
    return useServerQuery<Status>({
        endpoint: API_ENDPOINTS.STATUS.GetStatus.endpoint,
        method: API_ENDPOINTS.STATUS.GetStatus.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetStatus.key],
        enabled: true,
        retryDelay: 1000,
        // Fixes macOS desktop app startup issue
        retry: 6,
        // Mute error if the platform is desktop
        muteError: __isDesktop__,
    })
}

export function useGetLogFilenames() {
    return useServerQuery<Array<string>>({
        endpoint: API_ENDPOINTS.STATUS.GetLogFilenames.endpoint,
        method: API_ENDPOINTS.STATUS.GetLogFilenames.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetLogFilenames.key],
        enabled: true,
    })
}

export function useDeleteLogs() {
    const qc = useQueryClient()
    return useServerMutation<boolean, DeleteLogs_Variables>({
        endpoint: API_ENDPOINTS.STATUS.DeleteLogs.endpoint,
        method: API_ENDPOINTS.STATUS.DeleteLogs.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.DeleteLogs.key],
        onSuccess: async () => {
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetLogFilenames.key] })
            toast.success("Logs deleted")
        },
    })
}

export function useGetLatestLogContent() {
    const qc = useQueryClient()
    return useServerMutation<string>({
        endpoint: API_ENDPOINTS.STATUS.GetLatestLogContent.endpoint,
        method: API_ENDPOINTS.STATUS.GetLatestLogContent.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetLatestLogContent.key],
        onSuccess: async data => {
            if (!data) return toast.error("Couldn't fetch logs")
            try {
                await copyToClipboard(data)
                toast.success("Copied to clipboard")
            }
            catch (err: any) {
                console.error("Clipboard write error:", err)
                toast.error("Failed to copy logs: " + err.message)
            }
        },
    })
}

export function useGetAnnouncements() {
    return useServerMutation<Array<Updater_Announcement>, GetAnnouncements_Variables>({
        endpoint: API_ENDPOINTS.STATUS.GetAnnouncements.endpoint,
        method: API_ENDPOINTS.STATUS.GetAnnouncements.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetAnnouncements.key],
    })
}

// Memory profiling hooks

export function useGetMemoryStats() {
    return useServerQuery<MemoryStatsResponse>({
        endpoint: API_ENDPOINTS.STATUS.GetMemoryStats.endpoint,
        method: API_ENDPOINTS.STATUS.GetMemoryStats.methods[0],
        queryKey: [API_ENDPOINTS.STATUS.GetMemoryStats.key],
        enabled: false, // Manual trigger only
        refetchInterval: false,
    })
}

export function useForceGC() {
    const qc = useQueryClient()
    return useServerMutation<MemoryStatsResponse>({
        endpoint: API_ENDPOINTS.STATUS.ForceGC.endpoint,
        method: API_ENDPOINTS.STATUS.ForceGC.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.ForceGC.key],
        onSuccess: async () => {
            // Invalidate and refetch memory stats after GC
            await qc.invalidateQueries({ queryKey: [API_ENDPOINTS.STATUS.GetMemoryStats.key] })
            toast.success("Garbage collection completed")
        },
    })
}

export function useDownloadMemoryProfile() {
    const password = useAtomValue(serverAuthTokenAtom)

    return useServerMutation<string, { profileType: "heap" | "allocs" }>({
        endpoint: API_ENDPOINTS.STATUS.GetMemoryProfile.endpoint,
        method: API_ENDPOINTS.STATUS.GetMemoryProfile.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetMemoryProfile.key],
        onMutate: async (variables) => {
            const profileType = variables.profileType || "heap"
            toast.info(`Generating ${profileType} profile...`)

            let downloadUrl = getServerBaseUrl() + API_ENDPOINTS.STATUS.GetMemoryProfile.endpoint
            if (profileType === "heap") {
                downloadUrl += "?heap=true"
            } else if (profileType === "allocs") {
                downloadUrl += "?allocs=true"
            }

            try {
                const headers: Record<string, string> = {}
                if (password) {
                    headers["X-Seanime-Token"] = password
                }

                const response = await fetch(downloadUrl, {
                    method: "GET",
                    headers,
                })

                if (!response.ok) {
                    throw new Error(`HTTP error: status: ${response.status}`)
                }

                const blob = await response.blob()
                const timestamp = new Date().toISOString().replace(/[:.]/g, "-").split("T")[0] + "_" +
                    new Date().toISOString().replace(/[:.]/g, "-").split("T")[1].split(".")[0]
                const filename = `seanime-${profileType}-profile-${timestamp}.pprof`

                const url = window.URL.createObjectURL(blob)
                const link = document.createElement("a")
                link.href = url
                link.setAttribute("download", filename)
                link.style.display = "none"
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
                window.URL.revokeObjectURL(url)

                toast.success(`Profile "${profileType}" downloaded`)
            }
            catch (error) {
                console.error("Download error:", error)
                toast.error(`Failed to download ${profileType} profile`)
            }

            throw new Error("Download handled in onMutate")
        },
        onError: (error) => {
            if (error.message !== "Download handled in onMutate") {
                toast.error("Failed to download memory profile")
            }
        },
    })
}

export function useDownloadGoRoutineProfile() {
    const password = useAtomValue(serverAuthTokenAtom)

    return useServerMutation<string>({
        endpoint: API_ENDPOINTS.STATUS.GetGoRoutineProfile.endpoint,
        method: API_ENDPOINTS.STATUS.GetGoRoutineProfile.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetGoRoutineProfile.key],
        onMutate: async () => {
            toast.info("Generating goroutine profile...")

            const downloadUrl = getServerBaseUrl() + API_ENDPOINTS.STATUS.GetGoRoutineProfile.endpoint

            try {
                const headers: Record<string, string> = {}
                if (password) {
                    headers["X-Seanime-Token"] = password
                }

                const response = await fetch(downloadUrl, {
                    method: "GET",
                    headers,
                })

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`)
                }

                const blob = await response.blob()
                const timestamp = new Date().toISOString().replace(/[:.]/g, "-").split("T")[0] + "_" +
                    new Date().toISOString().replace(/[:.]/g, "-").split("T")[1].split(".")[0]
                const filename = `seanime-goroutine-profile-${timestamp}.pprof`

                const url = window.URL.createObjectURL(blob)
                const link = document.createElement("a")
                link.href = url
                link.setAttribute("download", filename)
                link.style.display = "none"
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
                window.URL.revokeObjectURL(url)

                toast.success("Goroutine profile downloaded")
            }
            catch (error) {
                console.error("Download error:", error)
                toast.error("Failed to download goroutine profile")
            }

            throw new Error("Download handled in onMutate")
        },
        onError: (error) => {
            if (error.message !== "Download handled in onMutate") {
                toast.error("Failed to download goroutine profile")
            }
        },
    })
}

export function useDownloadCPUProfile() {
    const password = useAtomValue(serverAuthTokenAtom)

    return useServerMutation<string, { duration?: number }>({
        endpoint: API_ENDPOINTS.STATUS.GetCPUProfile.endpoint,
        method: API_ENDPOINTS.STATUS.GetCPUProfile.methods[0],
        mutationKey: [API_ENDPOINTS.STATUS.GetCPUProfile.key],
        onMutate: async (variables) => {
            const duration = variables?.duration || 30
            toast.info(`Generating CPU profile for ${duration} seconds...`)

            const downloadUrl = `${getServerBaseUrl()}${API_ENDPOINTS.STATUS.GetCPUProfile.endpoint}?duration=${duration}`

            try {
                const headers: Record<string, string> = {}
                if (password) {
                    headers["X-Seanime-Token"] = password
                }

                const response = await fetch(downloadUrl, {
                    method: "GET",
                    headers,
                })

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`)
                }

                const blob = await response.blob()
                const timestamp = new Date().toISOString().replace(/[:.]/g, "-").split("T")[0] + "_" +
                    new Date().toISOString().replace(/[:.]/g, "-").split("T")[1].split(".")[0]
                const filename = `seanime-cpu-profile-${timestamp}.pprof`

                const url = window.URL.createObjectURL(blob)
                const link = document.createElement("a")
                link.href = url
                link.setAttribute("download", filename)
                link.style.display = "none"
                document.body.appendChild(link)
                link.click()
                document.body.removeChild(link)
                window.URL.revokeObjectURL(url)

                toast.success(`CPU profile (${duration}s) downloaded`)
            }
            catch (error) {
                console.error("Download error:", error)
                toast.error(`Failed to download CPU profile`)
            }

            throw new Error("Download handled in onMutate")
        },
        onError: (error) => {
            if (error.message !== "Download handled in onMutate") {
                toast.error("Failed to download CPU profile")
            }
        },
    })
}
