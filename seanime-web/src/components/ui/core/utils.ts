import type * as React from "react"

export function mergeRefs<T = any>(
    refs: Array<React.MutableRefObject<T> | React.LegacyRef<T> | undefined | null>,
): React.RefCallback<T> {
    return (value) => {
        refs.forEach((ref) => {
            if (typeof ref === "function") {
                ref(value)
            } else if (ref != null) {
                (ref as React.MutableRefObject<T | null>).current = value
            }
        })
    }
}

export const isEmpty = (obj: any) => [Object, Array].includes((obj || {}).constructor) && !Object.entries((obj || {})).length

