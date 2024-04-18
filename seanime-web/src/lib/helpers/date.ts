import { format, FormatOptions } from "date-fns"
import { formatDistanceToNow } from "date-fns/formatDistanceToNow"

export function formatDistanceToNowSafe(value: string) {
    try {
        return formatDistanceToNow(value, { addSuffix: true })
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

        let date = new Date()
        date.setFullYear(parseInt(arr[0]))
        date.setMonth(parseInt(arr[1]) - 1)
        date.setDate(parseInt(arr[2]))
        return date
    }
    catch (e) {
        return new Date(value)
    }
}
