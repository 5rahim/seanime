import { useServerQuery } from "@/api/client/requests"
import { API_ENDPOINTS } from "@/api/generated/endpoints"
import { DB_ScanSummaryItem } from "@/api/generated/types"

export function useGetScanSummaries() {
    return useServerQuery<Array<DB_ScanSummaryItem>>({
        endpoint: API_ENDPOINTS.SCAN_SUMMARY.GetScanSummaries.endpoint,
        method: API_ENDPOINTS.SCAN_SUMMARY.GetScanSummaries.methods[0],
        queryKey: [API_ENDPOINTS.SCAN_SUMMARY.GetScanSummaries.key],
        enabled: true,
    })
}

