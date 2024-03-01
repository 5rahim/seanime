import { formatDistanceToNow } from "date-fns/formatDistanceToNow"

export function formatDistanceToNowSafe(value: string) {
    try {
        return formatDistanceToNow(value, { addSuffix: true })
    }
    catch (e) {
        return "N/A"
    }
}
