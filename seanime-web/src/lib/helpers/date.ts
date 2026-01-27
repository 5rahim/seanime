import { format, FormatDistanceToNowOptions, FormatOptions } from "date-fns"
import { formatDistanceToNow } from "date-fns/formatDistanceToNow"

export function formatDistanceToNowSafe(value: string, options: FormatDistanceToNowOptions | undefined = { addSuffix: true }) {
    try {
        return formatDistanceToNow(value, options)
    }
    catch (e) {
        return "N/A"
    }
}

export function newDateSafe(value: string) {
    try {
        return new Date(value)
    }
    catch (e) {
        return new Date()
    }
}

export function formatSafe(value: Date, formatString: string, options?: FormatOptions | undefined) {
    try {
        return format(value, formatString, options)
    }
    catch (e) {
        let v = new Date()
        return format(v, formatString, options)
    }
}

export function normalizeDate(value: string) {
    try {
        let arr = value.split(/[\-\+ :T]/)
        let year = parseInt(arr[0])
        let month = parseInt(arr[1]) - 1
        let day = parseInt(arr[2])

        return new Date(`${year}-${String(month + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}T00:00:00`)
    }
    catch (e) {
        return new Date(value)
    }
}
