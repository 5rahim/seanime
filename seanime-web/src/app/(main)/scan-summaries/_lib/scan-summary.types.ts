/**
 * Scan Summary
 */
import { LocalFile } from "@/app/(main)/(library)/_lib/anime-library.types"

export type ScanSummary = {
    createdAt: string
    id: string
    groups: ScanSummaryGroup[] | undefined
    unmatchedFiles: ScanSummaryFile[] | undefined
}
export type ScanSummaryFile = {
    id: string
    localFile: LocalFile
    logs: ScanSummaryLog[]
}
export type ScanSummaryGroup = {
    id: string
    files: ScanSummaryFile[]
    mediaId: number
    mediaTitle: string
    mediaImage: string
    mediaIsInCollection: boolean
}
export type ScanSummaryLog = {
    id: string
    filePath: string
    message: string
    level: "info" | "warning" | "error"
}
